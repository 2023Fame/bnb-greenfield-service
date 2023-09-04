package util

import (
	"github.com/gin-gonic/gin"
	"log"
)

func (e ErrorResponse) Error() string {
	return e.Message
}

type ErrorResponse struct {
	Code          int    `json:"code"`
	ErrorCode     int    `json:"errorCode"`
	Message       string `json:"message"`
	Detail        string `json:"detail,omitempty"`
	OriginalError error  `json:"original_error,omitempty"`
	ErrorStack    string `json:"errorStack"` // new field to contain the error stack
}

func NewErrorResponse(code int, errorCode int, message string, detail string, originalError error) ErrorResponse {
	return ErrorResponse{
		Code:          code,
		ErrorCode:     errorCode,
		Message:       message,
		Detail:        detail,
		OriginalError: originalError,
	}
}

func HandleErr(err error, msg string) bool {
	if err != nil {
		log.Printf("%s - error: %v", msg, err)
		return true
	}
	return false
}

// Error represents a standard application error.
type Error struct {
	StatusCode int
	ErrorCode  int
	Message    string
}

// Error makes it compatible with `error` interface
func (e *Error) Error() string {
	return e.Message
}

var (
	// 1000-1099: Parameter errors
	PostObjectNameArgumentError = &Error{StatusCode: 400, ErrorCode: 1001, Message: "objectName is required in form data"}
	PostBucketNameArgumentError = &Error{StatusCode: 400, ErrorCode: 1002, Message: "bucketName is required in form data"}
	GetObjectNameArgumentError  = &Error{StatusCode: 400, ErrorCode: 1003, Message: "objectName is required in query parameters"}
	GetBucketNameArgumentError  = &Error{StatusCode: 400, ErrorCode: 1004, Message: "bucketName is required in query parameters"}
	GetUserAdressArgumentError  = &Error{StatusCode: 400, ErrorCode: 1005, Message: "userAdress is required in query parameters"}
	ParametersNotValidatedError = &Error{StatusCode: 400, ErrorCode: 1900, Message: "Parameters not validated"}
	GetValidatedParametersError = &Error{StatusCode: 400, ErrorCode: 1900, Message: "Get validated parameters failed"}
	ReadingParametersError      = &Error{StatusCode: 500, ErrorCode: 1901, Message: "Internal error while reading parameters"}
	InvalidRangeHeader          = &Error{StatusCode: 400, ErrorCode: 1902, Message: "Range header is invalid"}

	// 1100-1199: File errors
	FormDataFileRetrieveError = &Error{StatusCode: 500, ErrorCode: 1101, Message: "Error in retrieving the folder form file"}
	FileOpenError             = &Error{StatusCode: 500, ErrorCode: 1102, Message: "Error in opening the uploaded file"}
	FileReadError             = &Error{StatusCode: 500, ErrorCode: 1103, Message: "Reading uploaded file failed"}
	FileStatError             = &Error{StatusCode: 500, ErrorCode: 1104, Message: "Stat file failed"}
	DirectoryCreationError    = &Error{StatusCode: 500, ErrorCode: 1105, Message: "Directory create failed"}

	// 1200-1299: Encryption errors
	EncryptionError = &Error{StatusCode: 500, ErrorCode: 1201, Message: "Encryption failed"}
	DecryptionError = &Error{StatusCode: 500, ErrorCode: 1202, Message: "Decryption failed"}

	// 1300-1399: BNB Client errors
	BNBClientUploadError         = &Error{StatusCode: 500, ErrorCode: 1301, Message: "Failed to upload to BNBClient"}
	BNBClientDownloadError       = &Error{StatusCode: 500, ErrorCode: 1302, Message: "Failed to download to BNBClient"}
	BNBClientCreateTempFileError = &Error{StatusCode: 500, ErrorCode: 1303, Message: "Failed to createTempFile in BNBClient"}
	BNBFGetObjectResumableError  = &Error{StatusCode: 500, ErrorCode: 1304, Message: "Failed to call BNB Greenfield: FGetObjectResumable"}

	// 1400-1499: Zip operation errors
	ZipEntryCreationError = &Error{StatusCode: 500, ErrorCode: 1401, Message: "Failed to create new zip entry"}
	ZipWriteError         = &Error{StatusCode: 500, ErrorCode: 1402, Message: "Failed to write decrypted file to new zip"}
	ZipFinalizeError      = &Error{StatusCode: 500, ErrorCode: 1403, Message: "Failed to finalize new zip file"}
)

func ReportError(ctx *gin.Context, originalError error, err *Error) {
	if err := ctx.Error(NewErrorResponse(err.StatusCode, err.ErrorCode, "Internal Error", err.Message, originalError)); err != nil {
		log.Printf("Unexpected error when adding error to context: %v", err)
	}
}

// CheckQueryParameter checks the existence of a query parameter and returns its value.
// If the parameter is missing, it reports an error using the given error definition.
func CheckQueryParameter(ctx *gin.Context, paramName string, err *Error) (string, bool) {
	value := ctx.Query(paramName)
	if value == "" {
		ReportError(ctx, nil, err)
		return "", false
	}
	return value, true
}
