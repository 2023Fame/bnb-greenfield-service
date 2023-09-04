// storageclient/bnb_client.go

package storageclient

import (
	"bytes"
	"context"
	"errors"
	"file-service/models"
	"file-service/util"
	"fmt"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

type BNBClient struct {
	cli          client.Client
	account      *types.Account
	primarySP    string
	chargedQuota uint64
	visibility   storageTypes.VisibilityType
	opts         types.CreateBucketOptions
}

func NewClient(privateKey string, chainId string, rpcAddr string) *BNBClient {
	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	util.HandleErr(err, "New account from private key error")

	cli, err := client.New(chainId, rpcAddr, client.Option{DefaultAccount: account})
	util.HandleErr(err, "unable to new greenfield client")

	c := &BNBClient{
		cli:     cli,
		account: account,
	}

	// get storage providers list
	ctx := context.Background() // Create a background context, can be replaced with a more relevant context if necessary
	spLists, err := c.cli.ListStorageProviders(ctx, true)
	util.HandleErr(err, "fail to list in service sps")

	// choose the first sp to be the primary SP
	c.primarySP = spLists[0].GetOperatorAddress()
	log.Printf("primarySP:  %v", c.primarySP)

	c.chargedQuota = uint64(100)
	c.visibility = storageTypes.VISIBILITY_TYPE_PUBLIC_READ

	c.opts = types.CreateBucketOptions{Visibility: c.visibility, ChargedQuota: c.chargedQuota}

	return c
}

func (c *BNBClient) CreateBucket(ctx *gin.Context, bucketName string) (string, error) {
	// bucketName : testbucket
	txnBucketHash, err := c.cli.CreateBucket(ctx, bucketName, c.primarySP, c.opts)
	util.HandleErr(err, "CreateBucket failed ------------1-------.")
	log.Printf("Create Bucket: txnHash: %v", txnBucketHash)

	return txnBucketHash, nil
}

func (c *BNBClient) ExistBucket(ctx *gin.Context, bucketName string) bool {
	_, err := c.cli.HeadBucket(ctx, bucketName)
	if err != nil {
		return false
	}
	return true
}

func (c *BNBClient) CreateObject(ctx *gin.Context, bucketName string, objectName string, buffer []byte) (string, error) {
	if !c.ExistBucket(ctx, bucketName) {
		txnBucketHash, _ := c.CreateBucket(ctx, bucketName)
		log.Printf("Created/Checked bucket with txnHash: %v", txnBucketHash)
	}

	log.Printf("Start upload object.")
	// Upload the object
	txnHash, err := c.cli.CreateObject(ctx, bucketName, objectName, bytes.NewReader(buffer), types.CreateObjectOptions{})

	if util.HandleErr(err, "CreateObject failed --------2---------.") {
		return "", err
	}

	log.Printf("Start put object.")
	// Resumable upload object.
	err = c.cli.PutObject(ctx, bucketName, objectName, int64(len(buffer)),
		bytes.NewReader(buffer), types.PutObjectOptions{
			PartSize:         1024 * 1024 * 16,
			DisableResumable: false,
			TxnHash:          txnHash})

	waitObjectSeal(c.cli, bucketName, objectName)

	util.HandleErr(err, "PutObject")

	log.Printf("CreateObject txnHash : %v", txnHash)

	log.Printf("object: %s has been uploaded to SP\n", objectName)
	return txnHash, nil
}

func (c *BNBClient) GetObject(ctx *gin.Context, bucketName string, objectName string) ([]byte, error) {
	// Get the Object
	reader, info, err := c.cli.GetObject(ctx, bucketName, objectName, types.GetObjectOptions{})
	if util.HandleErr(err, "GetObject failed ------------4------------}") {
		log.Printf("object name: %v", objectName)
		return nil, err
	}

	log.Printf("get object %s successfully, size %d.. \n", info.ObjectName, info.Size)
	objectBytes, err := io.ReadAll(reader)
	return objectBytes, nil
}

func (c *BNBClient) GetObjectResumable(ctx *gin.Context, bucketName string, objectName string, userUniqueID string) (string, error) {
	// Path for storing the object - userUniqueID ensures each download has a unique file
	// Construct file path
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
		return "", err
	}

	filePath := filepath.Join(cwd, "tmp", bucketName+"_"+objectName+"_"+userUniqueID)

	baseTempFilePath := filePath + "_last"
	// Initialize rangeStr to "bytes=0-"
	rangeStr := "bytes=0-"

	// Check if baseTempFilePath exists. If it exists, set the range from its current size.
	if _, err := os.Stat(baseTempFilePath); err == nil {
		if fileInfo, err := os.Stat(baseTempFilePath); err == nil {
			rangeStr = fmt.Sprintf("bytes=%d-", fileInfo.Size())
		}
	} else if !os.IsNotExist(err) {
		// Handle other potential errors when checking the file
		log.Printf("Error while checking base temp file: %v with error: %v", baseTempFilePath, err)
		return "", err
	}

	// Ensure the tempFilePath with the opts.Range exists.
	tempFilePath := filePath + "_" + c.account.GetAddress().String() + rangeStr + types.TempFileSuffix
	if _, err := os.Stat(tempFilePath); os.IsNotExist(err) {
		file, err := os.Create(tempFilePath)
		if err != nil {
			util.ReportError(ctx, err, util.BNBClientCreateTempFileError)
			log.Printf("Failed to create temp file with range: %v with error: %v", tempFilePath, err)
			return "", err
		}
		file.Close()
		log.Printf("Temp file with range: %v created successfully.", tempFilePath)
	}

	// Download the Object using FGetObjectResumable
	err = c.cli.FGetObjectResumable(ctx, bucketName, objectName, filePath, types.GetObjectOptions{
		SupportResumable: true,
		Range:            rangeStr,
		PartSize:         1024 * 1024 * 16,
	})

	if err != nil {
		util.ReportError(ctx, err, util.BNBFGetObjectResumableError)
		log.Printf("FGetObjectResumable failed for object name: %v with error: %v", objectName, err)
		return "", err
	}

	if fileExists(baseTempFilePath) {
		err = os.Remove(baseTempFilePath)
		if err != nil {
			util.ReportError(ctx, err, util.BNBClientDownloadError)
			return "", err
		}
	}
	// rename temp file
	err = os.Rename(filePath, baseTempFilePath)
	if err != nil {
		return "", err
	}

	log.Printf("get object %s successfully. \n", objectName)

	return baseTempFilePath, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (c *BNBClient) ListObjects(ctx *gin.Context, bucketName string) (models.ObjectListResponse, error) {
	objects, err := c.cli.ListObjects(ctx, bucketName, types.ListObjectsOptions{
		ShowRemovedObject: false, Delimiter: "", MaxKeys: 100, EndPointOptions: &types.EndPointOptions{
			Endpoint:  "",
			SPAddress: "",
		}})
	if err != nil {
		return models.ObjectListResponse{}, err
	}

	var response models.ObjectListResponse
	for _, obj := range objects.Objects {
		objectBytes, err := c.GetObject(ctx, bucketName, obj.ObjectInfo.ObjectName)
		if err != nil {
			util.HandleErr(err, "")
			continue
		}
		objectInfo := models.ObjectInfo{
			ObjectName: obj.ObjectInfo.ObjectName,
			Data:       objectBytes,
			Type:       obj.ObjectInfo.ContentType,
		}
		response.Objects = append(response.Objects, objectInfo)
	}
	return response, nil
}

func waitObjectSeal(cli client.Client, bucketName, objectName string) {
	ctx := context.Background()
	// wait for the object to be sealed
	timeout := time.After(15 * time.Second)
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-timeout:
			err := errors.New("object not sealed after 15 seconds")
			util.HandleErr(err, "")
		case <-ticker.C:
			objectDetail, err := cli.HeadObject(ctx, bucketName, objectName)
			util.HandleErr(err, "HeadObject")
			if objectDetail.ObjectInfo.GetObjectStatus().String() == "OBJECT_STATUS_SEALED" {
				ticker.Stop()
				fmt.Printf("put object %s successfully \n", objectName)
				return
			}
		}
	}
}

func CleanupOldFiles() {
	// Run this function in a goroutine at regular intervals to clean up old files
	// Define a file age threshold as per your requirements
	thresholdAge := 24 * time.Hour

	tmpDir := "/tmp/"
	files, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		filePath := filepath.Join(tmpDir, file.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Unable to stat file: %s", filePath)
			continue
		}

		if time.Since(fileInfo.ModTime()) > thresholdAge {
			os.Remove(filePath)
		}
	}
}
