package bank_account

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"net/http"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/indrasaputra/hashids"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

type InsertTokenRequest struct {
	Token pq.StringArray `json:"token" validate:"required"`
}

func InsertToken(uc domain.BankAccountUsecase) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		str := c.Param("id")
		id, err := hashids.DecodeHash([]byte(str))
		if err != nil {
			return errors.ErrResourceNotFound
		}

		payload := InsertTokenRequest{}

		if err := c.Bind(&payload); err != nil {
			return errors.BadRequest.New(err.Error())
		}

		validate := validator.New()

		if err := validate.Struct(payload); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		for _, token := range payload.Token {
			if len(token) != 8 {
				return errors.BadRequest.New("Len token must be 8.")
			} else {
				var wg sync.WaitGroup
				wg.Add(1)
				if err := uc.AddToken(ctx, uint64(id), payload.Token); err != nil {
					return errors.InternalServerError.New(err.Error())
				}
				wg.Done()

				wg.Wait()
			}
		}

		return c.NoContent(http.StatusCreated)
	}
}
