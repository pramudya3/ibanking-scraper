package bank_account

import (
	"database/sql"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"net/http"

	"github.com/indrasaputra/hashids"
	"github.com/labstack/echo/v4"
)

func DeleteBankAccount(uc domain.BankAccountUsecase) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		ctx := c.Request().Context()
		str := c.Param("id")

		id, err := hashids.DecodeHash([]byte(str))
		if err != nil {
			return errors.ErrResourceNotFound
		}

		if err = uc.Delete(ctx, uint64(id)); err != nil {
			if err != sql.ErrNoRows {
				return errors.InternalServerError.New(err.Error())
			} else {
				return errors.ErrResourceNotFound
			}
		}

		return c.NoContent(http.StatusOK)
	}
}
