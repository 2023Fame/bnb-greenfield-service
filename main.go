package main

import (
	"file-service/config"
	"file-service/router"
	"file-service/storageclient"
	common "file-service/util"
)

// main.go

func main() {

	client := storageclient.NewClient(config.PrivateKey, config.ChainId, config.RpcAddr)

	app := router.SetupRouter(client)

	// Run cleanup in background after all initializations
	go func() {
		storageclient.CleanupOldFiles()
	}()

	// Run http service.
	err := app.Run()
	common.HandleErr(err, "Gin run failed.")
}
