package response

import (
	"ibanking-scraper/internal/errors"

	"github.com/labstack/echo/v4"
)

// Success represent an success response
type Success struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta"`
}

// Error represent an error response
type Error struct {
	Errors interface{} `json:"error"`
	Meta   interface{} `json:"meta,omitempty"`
}

// EmptyMeta represents an empty struct.
type EmptyMeta struct{}

// NewSuccess creates an instance of Success response.
func NewSuccess(data, meta interface{}) *Success {
	return &Success{
		Data: data,
		Meta: meta,
	}
}

// NewError creates an instance of Error response.
func NewError(err error) *Error {
	if echoerr, ok := err.(*echo.HTTPError); ok {
		err = errors.New(echoerr.Message.(string))
	}

	return &Error{
		Errors: errors.GetErrorContext(err),
		Meta:   nil,
	}
}
