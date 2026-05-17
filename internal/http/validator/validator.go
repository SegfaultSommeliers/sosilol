package validator

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Fields map[string][]string
}

func (e *ValidationError) Error() string {
	return "validation error"
}

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			return fld.Name
		}
		return name
	})

	return &CustomValidator{validator: v}
}

func (cv *CustomValidator) Validate(i any) error {
	err := cv.validator.Struct(i)
	if err == nil {
		return nil
	}

	if _, ok := errors.AsType[*validator.InvalidValidationError](err); ok {
		return err
	}

	ve, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok {
		return err
	}

	fields := make(map[string][]string, len(ve))
	for _, fe := range ve {
		field := fe.Field()
		fields[field] = append(fields[field], validationMessage(fe))
	}

	return &ValidationError{
		Fields: fields,
	}
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "must be at least " + fe.Param()
	case "max":
		return "must be at most " + fe.Param()
	case "gte":
		return "must be greater than or equal to " + fe.Param()
	case "lte":
		return "must be less than or equal to " + fe.Param()
	case "alphanum":
		return "must be alphanumeric"
	default:
		return "is invalid"
	}
}
