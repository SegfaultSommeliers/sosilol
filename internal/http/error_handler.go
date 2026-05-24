package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/SegfaultSommeliers/sosilol/internal/http/validator"
	"github.com/labstack/echo/v5"
)

type ErrorResponse struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

type AppError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}

func sendError(
	log *slog.Logger,
	c *echo.Context,
	status int,
	resp ErrorResponse,
) {
	if err := c.JSON(status, resp); err != nil {
		log.Error("failed to send error response", "error", err)
	}
}

func CustomErrorHandler(
	c *echo.Context,
	err error,
) {
	log := c.Logger()

	resp, uErr := echo.UnwrapResponse(c.Response())
	if uErr != nil {
		log.Error("failed to unwrap response", "error", uErr)
		return
	}
	if resp.Committed {
		log.Warn("response already committed, cannot send error response")
		return
	}

	if validationErr, ok := errors.AsType[*validator.ValidationError](err); ok {
		sendError(log, c, http.StatusBadRequest, ErrorResponse{
			Code:    "validation_error",
			Message: "validation failed",
			Errors:  validationErr.Fields,
		})
		return
	}

	if httpErr, ok := errors.AsType[*echo.HTTPError](err); ok {
		msg := httpErr.Message
		if msg == "" {
			msg = http.StatusText(httpErr.Code)
		}

		sendError(log, c, httpErr.Code, ErrorResponse{
			Code:    "http_error",
			Message: msg,
		})
		return
	}

	if appErr, ok := errors.AsType[*AppError](err); ok {
		statusCode := appErr.StatusCode
		code := appErr.Code
		msg := appErr.Message

		sendError(log, c, statusCode, ErrorResponse{
			Code:    code,
			Message: msg,
		})
		return
	}

	code := echo.StatusCode(err)
	if code != 0 {
		sendError(log, c, code, ErrorResponse{
			Code:    "http_error",
			Message: http.StatusText(code),
		})
		return
	}

	sendError(log, c, http.StatusInternalServerError, ErrorResponse{
		Code:    "internal_error",
		Message: "internal server error",
	})
}
