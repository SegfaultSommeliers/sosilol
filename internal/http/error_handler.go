package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/SegfaultSommeliers/sosilol/internal/http/validator"
	"github.com/gofiber/fiber/v3"
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
	ctx fiber.Ctx,
	status int,
	resp ErrorResponse,
) error {
	if err := ctx.Status(status).JSON(resp); err != nil {
		log.Error("failed to send error response", "error", err)
		return err
	}
	return nil
}

func NewCustomErrorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(
		ctx fiber.Ctx,
		err error,
	) error {
		if err == nil {
			return nil
		}

		if validationErr, ok := errors.AsType[*validator.ValidationError](err); ok {
			return sendError(log, ctx, http.StatusBadRequest, ErrorResponse{
				Code:    "validation_error",
				Message: "validation failed",
				Errors:  validationErr.Fields,
			})
		}

		if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
			msg := fiberErr.Message
			if msg == "" {
				msg = http.StatusText(fiberErr.Code)
			}

			return sendError(log, ctx, fiberErr.Code, ErrorResponse{
				Code:    "http_error",
				Message: msg,
			})
		}

		if appErr, ok := errors.AsType[*AppError](err); ok {
			statusCode := appErr.StatusCode
			code := appErr.Code
			msg := appErr.Message

			return sendError(log, ctx, statusCode, ErrorResponse{
				Code:    code,
				Message: msg,
			})
		}

		return sendError(log, ctx, http.StatusInternalServerError, ErrorResponse{
			Code:    "internal_error",
			Message: "internal server error",
		})
	}
}
