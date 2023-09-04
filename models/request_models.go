package models

type PostObjectRequestBody struct {
	ObjectName string `form:"objectName" binding:"required"`
	BucketName string `form:"bucketName" binding:"required"`
}
