package sosilol

import "net/http"

type AppError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrExchangeCodeFailed = &AppError{
		StatusCode: http.StatusUnauthorized,
		Code:       "authorization_failed",
		Message:    "failed to exchange code for token",
	}
	ErrGetGithubClientFailed = &AppError{
		StatusCode: http.StatusUnauthorized,
		Code:       "failed_get_profile",
		Message:    "failed to get github client",
	}
	ErrUserNotFound = &AppError{
		StatusCode: http.StatusUnauthorized,
		Code:       "user_not_found",
		Message:    "failed to find github user",
	}
	ErrPasteNotFound = &AppError{
		StatusCode: http.StatusNotFound,
		Code:       "paste_not_found",
		Message:    "failed to find paste",
	}
)

func BadRequest(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "bad_request",
		Message:    message,
	}
}
