package storageclient

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
)

type LocalClient struct {
	BasePath string
}

// to test locally
func NewLocalClient(basePath string) *LocalClient {
	return &LocalClient{BasePath: basePath}
}

func (c *LocalClient) CreateObject(ctx context.Context, bucketName string, objectName string, data []byte) (string, error) {
	objectPath := filepath.Join(c.BasePath, bucketName, objectName)

	// Ensure the directory exists
	dirPath := filepath.Dir(objectPath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	// Write the data to the file
	if err := ioutil.WriteFile(objectPath, data, 0644); err != nil {
		return "", err
	}

	return objectPath, nil
}

func (c *LocalClient) GetObject(ctx context.Context, bucketName string, objectName string) ([]byte, error) {
	objectPath := filepath.Join(c.BasePath, bucketName, objectName)
	return ioutil.ReadFile(objectPath)
}
