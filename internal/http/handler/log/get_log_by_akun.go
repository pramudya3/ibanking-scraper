package log

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/http/helper"
	"ibanking-scraper/internal/http/response"
	"ibanking-scraper/internal/validator/v1"
	"net/http"

	"github.com/indrasaputra/hashids"
	"github.com/labstack/echo/v4"
)

func GetLogByAkun(uc domain.LogUsecase) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		v := validator.New()

		str := c.Param("id")
		akunBank, err := hashids.DecodeHash([]byte(str))
		if err != nil {
			return errors.BadRequest.New(errors.ErrResourceNotFound.Error())
		}

		var (
			in struct {
				domain.Filter
			}
		)

		qs := c.QueryParams()
		in.Filter.PageSize = helper.ReadInt(qs, "pageSize", 50, v)
		in.Filter.Page = helper.ReadInt(qs, "page", 1, v)

		data, meta, err := uc.FetchByAkunBank(ctx, uint64(akunBank), in.Filter)
		if err != nil {
			return errors.InternalServerError.New(err.Error())
		}

		return c.JSON(http.StatusOK, response.NewSuccess(data, meta))
	}
}
