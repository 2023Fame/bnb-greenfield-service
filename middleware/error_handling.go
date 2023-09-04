package middleware

import (
	"file-service/util"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		// If there are any errors
		if len(ctx.Errors) > 0 {
			allErrors := make([]util.ErrorResponse, len(ctx.Errors))

			// Loop through all errors and fill allErrors
			for i, ctxErr := range ctx.Errors {
				var customErr util.ErrorResponse
				if errors.As(ctxErr.Err, &customErr) {
					// Append original error and its stack trace
					customErr.ErrorStack = errors.Cause(ctxErr.Err).Error()
					allErrors[i] = customErr
				} else {
					// For non-custom errors, just get the stack trace
					allErrors[i] = util.ErrorResponse{
						Code:          http.StatusInternalServerError,
						Message:       "Internal Server Error",
						Detail:        ctxErr.Err.Error(),
						OriginalError: ctxErr.Err,
						ErrorStack:    errors.Cause(ctxErr.Err).Error(),
					}
				}
			}

			// Return all errors as an array in the response
			ctx.JSON(http.StatusInternalServerError, allErrors)
		}
	}
}
