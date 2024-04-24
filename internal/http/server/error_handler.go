package server

import (
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/http/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorHandler handling error response
func ErrorHandler(err error, c echo.Context) {
	if !c.Response().Committed {
		switch c.Request().Method {
		case http.MethodHead:
			err = c.NoContent(http.StatusOK)
		default:
			err = c.JSON(getHTTPError(err))
		}
	}

	if err != nil {
		// Log misbehaving error
		c.Logger().Error(err)
	}
}

// Convert err to known http status code
func getHTTPError(err error) (code int, resp *response.Error) {

	var internalMessage = "internal server error - please contact support"

	if echoerr, ok := err.(*echo.HTTPError); ok {
		switch echoerr.Code {
		case http.StatusNotFound:
			return http.StatusNotFound, response.NewError(errors.ErrResourceNotFound)
		default:
			err = errors.Unknown.New(internalMessage)
			return http.StatusInternalServerError, response.NewError(err)
		}
	}

	switch errors.GetType(err) {
	case errors.Unknown, errors.InternalServerError:
		err = errors.Unknown.New(internalMessage)
		return http.StatusInternalServerError, response.NewError(err)
	case errors.BadRequest:
		code = http.StatusBadRequest
	case errors.RequestFailed:
		code = http.StatusPaymentRequired
	case errors.NotFound:
		code = http.StatusNotFound
	case errors.Conflict:
		code = http.StatusConflict
	case errors.Unauthorized:
		code = http.StatusUnauthorized
	case errors.Forbidden:
		code = http.StatusForbidden
	case errors.NoType:
		code = http.StatusNotImplemented
	}

	return code, response.NewError(err)
}
