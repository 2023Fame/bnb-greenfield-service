# BNB Storage Client Guide

The BNB Storage Client is a Go package providing an interface to interact with the Greenfield storage service on the BNB Chain. With this client, users can manage their storage operations such as creating buckets, uploading objects, downloading objects, and listing stored objects.

## Table of Contents:
- [Initialization](#initialization)
- [Bucket Operations](#bucket-operations)
- [Object Operations](#object-operations)
- [Utilities](#utilities)
- [Cleanup](#cleanup)

---

### Initialization:
To initialize the BNB client, you need a `privateKey`, `chainId`, and `rpcAddr`:

```go
client := storageclient.NewClient(config.PrivateKey, config.ChainId, config.RpcAddr)

app := router.SetupRouter(client)
```

We will initialize BNBclient in main.go and obtain the parameters required to connect to BNB from azure keyvalut in config.go based on the Managed Service Identity credentials: RpcAddr, ChainId, PrivateKey, etc., as well as the PrivateAESKey and other secrets required for encryption

---

### Bucket Operations:
#### Create a Bucket:
To create a new bucket:
```go
txnHash, err := client.CreateBucket(ctx, bucketName)
```
#### Check if Bucket Exists:
```go
exists := client.ExistBucket(ctx, bucketName)
```

---

### Object Operations:
#### Upload an Object:
Before uploading, the client checks if the specified bucket exists and creates it if necessary:
```go
txnHash, err := client.CreateObject(ctx, bucketName, objectName, buffer)
```
#### Download an Object:
Simple object retrieval:
```go
data, err := client.GetObject(ctx, bucketName, objectName)
```

For resumable downloads:
```go
filePath, err := client.GetObjectResumable(ctx, bucketName, objectName, userUniqueID)
```
The `GetObjectResumable` method supports interrupted downloads by checking for existing temporary files and resuming the download.

### Cleanup:
Over time, temporary files might accumulate on the system. To regularly clean up old files:
```go
CleanupOldFiles()
```

The `CleanupOldFiles` function scans the `/tmp/` directory and deletes old files based on a specified threshold (default is 24 hours).

#### List Stored Objects:
To list objects in a bucket: ()
```go
objectsList, err := client.ListObjects(ctx, bucketName)
```
---

For more details, please refer to the provided `bnb_client.go` source file or consult the Greenfield SDK documentation on the BNB Chain.
