package validator

import (
	"ibanking-scraper/internal/errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/labstack/echo/v4"
)

type reqValidator struct {
	validator *validator.Validate
}

type reqValidatonErrorContext struct {
	Message string `json:"message"`
	Field   string `json:"field"`
}

func (v *reqValidator) Validate(i interface{}) error {
	err := v.validator.Struct(i)
	if err != nil {
		return NewValidationErrors(err)
	}
	return nil
}

func New() echo.Validator {
	validator := validator.New()

	validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return &reqValidator{
		validator: validator,
	}
}

func NewValidationErrors(err error) error {
	errs := errors.ValidationErrors{}

	for _, e := range err.(validator.ValidationErrors) {
		message := getErrorMessage(e)
		ve := &reqValidatonErrorContext{
			Message: message,
			Field:   e.Field(),
		}

		fieldErr := errors.New(message)
		fieldErr = errors.AddErrorContext(fieldErr, ve)
		errs = append(errs, fieldErr)
	}

	err = errors.ErrRequestPayloadInvalid
	err = errors.AddErrorContext(err, errs)

	return err
}
