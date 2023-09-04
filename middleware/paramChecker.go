package middleware

import (
	"file-service/util"
	"github.com/gin-gonic/gin"
)

func ParamChecker(source string, requiredParams map[string]*util.Error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		paramValues := make(map[string]string)

		switch source {
		case "query":
			for paramName, err := range requiredParams {
				value := ctx.Query(paramName)
				if value == "" {
					util.ReportError(ctx, nil, err)
					ctx.Abort()
					return
				}
				paramValues[paramName] = value
			}
		}

		ctx.Set("ValidatedParams", paramValues)
		ctx.Next()
	}
}
