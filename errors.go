package sosilol

import "net/http"

type AppError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e AppError) Error() string {
	return e.Message
}

func BadRequest(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "bad_request",
		Message:    message,
	}
}
