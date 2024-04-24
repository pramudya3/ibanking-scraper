package errors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// ErrorType is application-level error type
type ErrorType uint

const (
	// NoType error (451)
	NoType ErrorType = iota

	// BadRequest error (400)
	// The request was unacceptable, often due to missing a required parameter.
	BadRequest

	// Unauthorized error (401)
	// No valid authorization/authentication used
	Unauthorized

	// RequestFailed error (402)
	// The parameters were valid but the request failed.
	RequestFailed

	// Forbidden error (403)
	// Not permitted to continue the request
	Forbidden

	// NotFound error (404)
	// The requested resource doesn't exist.
	NotFound

	// Conflict error (409)
	Conflict

	// InternalServerError error (500)
	// Something went wrong on the server side
	InternalServerError

	// Unknown error (503)
	// We still unable to identified the error
	Unknown
)

var (

	// ErrInternalServer indicates there is unexpected problem occurs in the system itself.
	// The detail of the error/problem should be known in internal message.
	ErrInternalServer = InternalServerError.New("internal server error")

	// ErrUnauthorized is returned when a request doesn't include authorization in its header.
	// The authorization must be using bearer authorization.
	// It also can be returned if the authorization is invalid.
	ErrUnauthorized = Unauthorized.New("request is unauthorized")

	// ErrLoginDisabled is returned when user is not allowed to login
	ErrLoginDisabled = Forbidden.New("user access is temporarily restricted, contact support for more information")

	// ErrUserNotAllowed is returned when user is not allowed to log in
	ErrUserNotAllowed = Forbidden.New("user not allowed access this application, contact support for more information")

	// ErrWrongContentType is returned when content-type in request's header is not as expected.
	ErrWrongContentType = BadRequest.New("wrong content type")

	// ErrResourceNotFound is returned when requested resource not found
	ErrResourceNotFound = NotFound.New("requested resource not found")

	// ErrRequestPayloadInvalid is returned when given payload on request is invalid
	ErrRequestPayloadInvalid = BadRequest.New("invalid requested payload")

	// ErrJWTInvalid is returned when given jwt is not valid
	ErrJWTInvalid = Unauthorized.New("JSON Web Token is not valid")

	// ErrJWTExpired is returned when given jwt is expired
	ErrJWTExpired = Unauthorized.New("JSON Web Token is expired")
)

type ValidationErrors []error

func (ve ValidationErrors) Error() string {
	buff := bytes.NewBufferString("")

	for i := 0; i < len(ve); i++ {

		if err, ok := ve[i].(applicationError); ok {
			buff.WriteString(err.Error())
			buff.WriteString("\n")
		}
	}

	return strings.TrimSpace(buff.String())
}

type applicationError struct {
	errorType     ErrorType
	originalError error
	errContext    interface{}
}

func (e applicationError) Error() string {
	return e.originalError.Error()
}

func (e applicationError) MarshalJSON() ([]byte, error) {
	if e.errContext != nil {
		return json.Marshal(e.errContext)
	}

	return json.Marshal(map[string]interface{}{
		"message": e.Error(),
	})
}

func (e applicationError) Cause() error {
	return errors.Cause(e.originalError)
}

func (e applicationError) Is(target error) bool {
	if err, ok := target.(applicationError); ok {
		if e.Cause() == err.Cause() {
			return true
		}
	}

	return false
}

// New creates a new applicationError
func (errorType ErrorType) New(msg string) error {
	return errorType.Newf(msg)
}

// Newf creates a new applicationError with formatted message
func (errorType ErrorType) Newf(msg string, args ...interface{}) error {
	return applicationError{
		errorType:     errorType,
		originalError: fmt.Errorf(msg, args...),
	}
}

// Wrap creates a new wrapped error
func (errorType ErrorType) Wrap(err error, msg string) error {
	return errorType.Wrapf(err, msg)
}

// Wrapf creates a new wrapped error with formatted message
func (errorType ErrorType) Wrapf(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return applicationError{
		errorType:     errorType,
		originalError: errors.Wrapf(err, msg, args...),
	}
}

// New creates a no type error
func New(msg string) error {
	return Newf(msg)
}

// Newf creates a no type error with formatted message
func Newf(msg string, args ...interface{}) error {
	return NoType.Newf(msg, args...)
}

// Wrap an error with a string
func Wrap(err error, msg string) error {
	return Wrapf(err, msg)
}

// Wrapf an error with format string
func Wrapf(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(applicationError); ok {
		return applicationError{
			errorType:     appErr.errorType,
			originalError: errors.Wrapf(err, msg, args...),
			errContext:    appErr.errContext,
		}
	}

	return NoType.Wrapf(err, msg, args...)
}

// Is give error type check
func Is(err, target error) bool {
	if appErr, ok := err.(applicationError); ok {
		return appErr.Is(target)
	}

	return errors.Is(err, target)
}

// AddErrorContext add param error context to an error
func AddErrorContext(err error, errContext interface{}) error {
	if appErr, ok := err.(applicationError); ok {
		appErr.errContext = errContext
		return appErr
	}

	return err
}

// GetErrorContext returns the error context
func GetErrorContext(err error) interface{} {
	if appErr, ok := err.(applicationError); ok {
		if appErr.errContext == nil {
			return map[string]interface{}{
				"message": appErr.originalError.Error(),
			}
		}
		return appErr.errContext
	}

	return Unknown.New(err.Error())
}
