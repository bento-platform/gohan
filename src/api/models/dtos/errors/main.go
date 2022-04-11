package errors

import (
	"api/models/dtos"
	"time"
)

func CreateSimpleBadRequest(message string) dtos.GeneralErrorResponseDto {
	return dtos.GeneralErrorResponseDto{
		Code:      400,
		Message:   "Bad Request",
		Timestamp: time.Now(),
		Errors: []dtos.GeneralError{
			{
				Message: message,
			},
		},
	}
}
func CreateSimpleNotFound(message string) dtos.GeneralErrorResponseDto {
	return dtos.GeneralErrorResponseDto{
		Code:      404,
		Message:   "Not Found",
		Timestamp: time.Now(),
		Errors: []dtos.GeneralError{
			{
				Message: message,
			},
		},
	}
}
func CreateSimpleInternalServerError(message string) dtos.GeneralErrorResponseDto {
	return dtos.GeneralErrorResponseDto{
		Code:      500,
		Message:   "Internal Server Error",
		Timestamp: time.Now(),
		Errors: []dtos.GeneralError{
			{
				Message: message,
			},
		},
	}
}
