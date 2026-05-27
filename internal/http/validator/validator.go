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

	t := reflect.TypeOf(i)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fields := make(map[string][]string, len(ve))
	for _, fe := range ve {
		msg := "is invalid"
		if sf, ok := t.FieldByName(fe.StructField()); ok {
			if m := sf.Tag.Get("message"); m != "" {
				msg = m
			}
		}
		fields[fe.Field()] = append(fields[fe.Field()], msg)
	}

	return &ValidationError{
		Fields: fields,
	}
}
