package validator

import (
	"ibanking-scraper/internal/errors"
)

type validationError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type Validator struct {
	stash  map[string]bool
	errors []validationError
}

func New() *Validator {
	return &Validator{
		stash:  make(map[string]bool),
		errors: []validationError{},
	}
}

func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

func (v *Validator) AddFieldError(field, message string) {
	if _, exist := v.stash[field]; !exist {
		v.stash[field] = true
		v.errors = append(v.errors, validationError{
			Field:   field,
			Message: message,
		})
	}
}

func (v *Validator) AddError(message string) {
	v.errors = append(v.errors, validationError{
		Message: message,
	})
}

func (v *Validator) Check(ok bool, field, message string) {
	if !ok {
		v.AddFieldError(field, message)
	}
}

func (v *Validator) Err() error {
	if len(v.errors) == 0 {
		return nil
	}

	err := errors.ErrRequestPayloadInvalid
	return errors.AddErrorContext(err, v.errors)
}
