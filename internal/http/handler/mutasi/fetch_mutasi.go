package mutasi

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/http/helper"
	"ibanking-scraper/internal/http/response"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/internal/validator/v1"
	"ibanking-scraper/pkg/constant"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func FetchMutasiByDate(uc domain.MutasiUsecase) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var (
			in struct {
				filter domain.Filter
			}
		)
		v := validator.New()
		qs := c.QueryParams()
		in.filter.PageSize = helper.ReadInt(qs, "pageSize", 50, v)
		in.filter.Page = helper.ReadInt(qs, "page", 1, v)
		startdate := helper.ReadString(qs, "startDate", time.Now().Format(constant.LayoutDateISO8601))
		startDate, err := time.Parse(constant.LayoutDateISO8601, startdate)
		if err != nil {
			logger.Debug(err)
		}
		endDate := helper.ReadString(qs, "endDate", startDate.AddDate(0, 0, -1).Format(constant.LayoutDateISO8601))
		rekening := helper.ReadString(qs, "rekening", "-")
		griyabayar := helper.ReadString(qs, "griyabayar", "false")

		var gb bool
		if griyabayar == "true" {
			gb = true
		} else if griyabayar == "false" {
			gb = false
		} else {
			return errors.BadRequest.New("the option is just true or false")
		}

		data, meta, err := uc.FetchByDate(ctx, startDate.Format(constant.LayoutDateISO8601), endDate, rekening, gb, in.filter)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, response.NewSuccess(&data, meta))
	}
}
