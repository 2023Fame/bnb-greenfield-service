// router/router.go
package router

import (
	v1 "file-service/api/v1"
	"file-service/middleware"
	"file-service/storageclient"
	"file-service/util"
	"github.com/gin-gonic/gin"
)

func SetupRouter(client *storageclient.BNBClient) *gin.Engine {

	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(middleware.ErrorHandlingMiddleware())
	r.Use(middleware.CORSMiddleware())

	ctrl := v1.NewController(client)

	apiV1 := r.Group("/api/v1")
	{
		// objectName: string
		// bucketName: string
		// folder: file
		apiV1.POST("/objects", ctrl.PostObject)

		apiV1.GET("/objects", middleware.ParamChecker("query", map[string]*util.Error{
			"objectName": util.GetObjectNameArgumentError,
			"bucketName": util.GetBucketNameArgumentError,
			"userAdress": util.GetUserAdressArgumentError,
		}), ctrl.GetObject)

		apiV1.GET("/objects/resumable", middleware.ParamChecker("query", map[string]*util.Error{
			"objectName": util.GetObjectNameArgumentError,
			"bucketName": util.GetBucketNameArgumentError,
			"userAdress": util.GetUserAdressArgumentError,
		}), ctrl.GetObjectResumable)

		// bucketName: string
		apiV1.GET("/buckets/objects", ctrl.ListObjects)
		apiV1.GET("/helloWorld", ctrl.HelloWorld)
	}

	return r
}
