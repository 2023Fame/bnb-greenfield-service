package v1

import (
	"archive/zip"
	"bytes"
	"file-service/config"
	"file-service/models"
	"file-service/storageclient"
	"file-service/util"
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	client *storageclient.BNBClient
}

func NewController(client *storageclient.BNBClient) *Controller {
	return &Controller{client: client}
}

func (c *Controller) PostObject(ctx *gin.Context) {
	var body models.PostObjectRequestBody
	if err := ctx.ShouldBindWith(&body, binding.Form); err != nil {
		util.ReportError(ctx, err, util.ParseBindError(err))
		return
	}
	objectName := body.ObjectName
	bucketName := body.BucketName

	log.Printf("1-1. UploadFolder start!")
	file, err := ctx.FormFile("folder")
	if err != nil {
		util.ReportError(ctx, err, util.FormDataFileRetrieveError)
		return
	}

	fileBytes, err := file.Open()
	if err != nil {
		util.ReportError(ctx, err, util.FileOpenError)
		return
	}

	defer func(fileBytes multipart.File) {
		err := fileBytes.Close()
		if err != nil {
			util.ReportError(ctx, err, util.FileOpenError)
		}
	}(fileBytes)

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, fileBytes); err != nil {
		util.ReportError(ctx, err, util.FileReadError)
		return
	}

	encryptedBytes, err := util.Encrypt(buffer.Bytes(), config.PrivateAESKey)
	if err != nil {
		util.ReportError(ctx, err, util.EncryptionError)
		return
	}

	_, err = c.client.CreateObject(ctx, bucketName, objectName, encryptedBytes)
	if err != nil {
		util.ReportError(ctx, err, util.BNBClientUploadError)
		return
	}

	log.Printf("1-2. UploadFolder success!")
	ctx.JSON(200, gin.H{
		"msg": "folder uploaded",
	})
}

// parseRangeHeader to parse HTTP Range header
func parseRangeHeader(header string) (start, end int64, err error) {
	_, err = fmt.Sscanf(header, "bytes=%d-%d", &start, &end)
	return
}

func getValidatedParams(ctx *gin.Context) (map[string]string, bool) {
	paramValuesInterface, ok := ctx.Get("ValidatedParams")
	if !ok {
		util.ReportError(ctx, nil, util.ParametersNotValidatedError)
		return nil, false
	}

	paramValues, ok := paramValuesInterface.(map[string]string)
	if !ok {
		util.ReportError(ctx, nil, util.ReadingParametersError)
		return nil, false
	}

	return paramValues, true
}

func (c *Controller) GetObject(ctx *gin.Context) {
	paramValues, ok := getValidatedParams(ctx)
	if !ok {
		util.ReportError(ctx, nil, util.GetValidatedParametersError)
		return
	}

	objectName := paramValues["objectName"]
	bucketName := paramValues["bucketName"]

	zipBytes, err := c.client.GetObject(ctx, bucketName, objectName)
	if err != nil {
		util.ReportError(ctx, err, util.BNBClientUploadError)
		return
	}

	decryptedBytes, err := util.Decrypt(zipBytes, config.PrivateAESKey)
	if err != nil {
		util.ReportError(ctx, err, util.DecryptionError)
		return
	}

	// Set Content-Length header
	ctx.Header("Content-Length", strconv.Itoa(len(decryptedBytes)))

	ctx.Header("Content-Disposition", "attachment; filename="+objectName)

	ctx.Data(200, "application/zip", decryptedBytes)
}

func (c *Controller) GetObjectResumable(ctx *gin.Context) {
	paramValues, ok := getValidatedParams(ctx)
	if !ok {
		util.ReportError(ctx, nil, util.GetValidatedParametersError)
		return
	}

	objectName := paramValues["objectName"]
	bucketName := paramValues["bucketName"]
	userAdress := paramValues["userAdress"]

	filePath, err := c.client.GetObjectResumable(ctx, bucketName, objectName, userAdress)
	if err != nil {
		util.ReportError(ctx, err, util.BNBClientDownloadError)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		util.ReportError(ctx, err, util.FileOpenError)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		util.ReportError(ctx, err, util.FileStatError)
		return
	}

	var buffer []byte
	rangeHeader := ctx.GetHeader("Range")
	if rangeHeader != "" {
		start, end, err := parseRangeHeader(rangeHeader)
		if err != nil {
			util.ReportError(ctx, err, util.InvalidRangeHeader)
			return
		}

		bytesToRead := end - start + 1
		buffer = make([]byte, bytesToRead)
		file.Seek(start, io.SeekStart)
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			util.ReportError(ctx, err, util.FileReadError)
			return
		}

		ctx.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
		ctx.Writer.WriteHeader(http.StatusPartialContent)
	} else {
		buffer, err = ioutil.ReadAll(file)
		if err != nil {
			util.ReportError(ctx, err, util.FileReadError)
			return
		}
	}

	decryptedBytes, err := util.Decrypt(buffer, config.PrivateAESKey)
	if err != nil {
		util.ReportError(ctx, err, util.DecryptionError)
		return
	}

	// Set Content-Length header
	ctx.Header("Content-Length", strconv.Itoa(len(decryptedBytes)))

	ctx.Header("Content-Disposition", "attachment; filename="+objectName)

	ctx.Data(200, "application/zip", decryptedBytes)
}

func (c *Controller) ListObjects(ctx *gin.Context) {
	bucketName := ctx.Query("bucketName")
	if bucketName == "" {
		util.ReportError(ctx, nil, util.GetBucketNameArgumentError)
		return
	}

	folder, err := c.client.ListObjects(ctx, bucketName)
	if err != nil {
		util.ReportError(ctx, err, util.BNBClientDownloadError)
		return
	}

	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)
	for _, obj := range folder.Objects {
		decryptedByte, err := util.Decrypt(obj.Data, config.PrivateAESKey)
		if err != nil {
			util.ReportError(ctx, err, util.DecryptionError)
			return
		}
		zipFile, err := zipWriter.Create(obj.ObjectName)
		if err != nil {
			util.ReportError(ctx, err, util.ZipEntryCreationError)
			return
		}
		_, err = zipFile.Write(decryptedByte)
		if err != nil {
			util.ReportError(ctx, err, util.ZipWriteError)
			return
		}
	}
	err = zipWriter.Close()
	if err != nil {
		util.ReportError(ctx, err, util.ZipFinalizeError)
		return
	}

	// Set Content-Length header
	ctx.Header("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	ctx.Header("Content-Disposition", "attachment; filename=aggregated.zip")
	ctx.Data(200, "application/zip", buffer.Bytes())
}

func (c *Controller) HelloWorld(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"msg": "Hello world!",
	})
}
