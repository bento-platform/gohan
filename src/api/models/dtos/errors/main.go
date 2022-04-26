package errors

import (
	"api/models/dtos"
	"time"
)

/*
	Utility functions to facillitate returning error responses to HTTP clients
*/

// -- Simplest: 1 error with message
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
func CreateSimpleUnauthorized(message string) dtos.GeneralErrorResponseDto {
	return dtos.GeneralErrorResponseDto{
		Code:      401,
		Message:   "Unauthorized",
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

// --
