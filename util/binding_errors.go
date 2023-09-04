package util

import "github.com/go-playground/validator/v10"

func ParseBindError(err error) *Error {
	// Cast error to validator's ValidationErrors type
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			// Check for required error type
			if e.Tag() == "required" {
				switch e.Field() {
				case "ObjectName":
					return PostObjectNameArgumentError
				case "BucketName":
					return PostBucketNameArgumentError
				}
			}
		}
	}
	return &Error{StatusCode: 400, ErrorCode: 1000, Message: err.Error()}
}
