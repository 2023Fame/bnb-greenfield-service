package util

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

func UnzipFolder(file *multipart.FileHeader) (string, error) {

	tmpDir, err := ioutil.TempDir("", "upload")

	zipFile, _ := file.Open()
	defer func(zipFile multipart.File) {
		err := zipFile.Close()
		HandleErr(err, "zipFile failed.")
	}(zipFile)

	reader, err := zip.NewReader(zipFile, file.Size)

	for _, f := range reader.File {
		filePath := filepath.Join(tmpDir, f.Name)
		if !strings.HasPrefix(filePath, filepath.Clean(tmpDir)+string(os.PathSeparator)) {
			return "", fmt.Errorf("invalid file path: %s", filePath)
		}
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(filePath, os.ModePerm)
			HandleErr(err, "os.MkdirAll failed.")
			continue
		}
		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return "", err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		_, err = io.Copy(outFile, rc)

		err = outFile.Close()
		HandleErr(err, "outFile failed.")
		if err != nil {
			return "", err
		}
		err = rc.Close()
		HandleErr(err, "rc.Close failed.")
		if err != nil {
			return "", err
		}
	}

	return tmpDir, nil
}

func WalkFolder(root string, fn func(path string, info os.FileInfo)) error {

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fn(path, info)
		}
		return nil
	})
	return err
}
