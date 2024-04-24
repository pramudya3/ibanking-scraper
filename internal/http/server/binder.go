package server

import (
	"encoding/json"
	"ibanking-scraper/internal/errors"

	"github.com/labstack/echo/v4"
)

type standardBinder struct{}

// NewBinder create a binder that comply with echo.Binder interface
func NewBinder() echo.Binder {
	return &standardBinder{}
}

// Bind do binding JSON request body to given interface
func (b *standardBinder) Bind(i interface{}, c echo.Context) (err error) {
	req := c.Request()
	if req.ContentLength == 0 {
		return errors.BadRequest.New("Request body is empty")
	}
	if err = json.NewDecoder(req.Body).Decode(i); err != nil {
		switch v := err.(type) {
		case *json.UnmarshalTypeError:
			return errors.BadRequest.Wrap(v, "Error while unmarshalling")
		case *json.SyntaxError:
			return errors.BadRequest.Wrap(v, "Bad JSON syntax")
		default:
			return errors.BadRequest.Wrap(err, "Invalid request")
		}
	}
	return
}
