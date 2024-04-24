package healthcheck

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Fetch() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}
}
