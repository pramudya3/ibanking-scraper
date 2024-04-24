package bank_account

import (
	"database/sql"
	"ibanking-scraper/domain"
	"ibanking-scraper/domain/types"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/utils"
	"net/http"

	"github.com/indrasaputra/hashids"
	"github.com/labstack/echo/v4"
)

func UpdateBankAccount(uc domain.BankAccountUsecase) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		str := c.Param("id")
		id, err := hashids.DecodeHash([]byte(str))

		if err != nil {
			return err
		}

		payload := &BankAccountRequest{}

		if err := c.Bind(&payload); err != nil {
			return errors.BadRequest.New(err.Error())
		}

		if err := c.Validate(payload); err != nil {
			return errors.Transform(err)
		}

		bankAccount, err := uc.FindByID(ctx, uint64(id))
		if err != nil {
			if err != sql.ErrNoRows {
				return errors.InternalServerError.New(err.Error())
			} else {
				return errors.ErrResourceNotFound
			}
		}

		var password string
		if payload.Password.String == "" {
			passwordDecoded := utils.DecodePassword(bankAccount.Password.String)
			password = passwordDecoded
		} else {
			password = payload.Password.String
		}

		data := &domain.BankAccount{
			TipeAkun:      payload.TipeAkun,
			CompanyId:     payload.CompanyId,
			UserId:        payload.UserId,
			Password:      types.NullString{String: password},
			RekOnpay:      payload.RekOnpay,
			RekGriyabayar: payload.RekGriyabayar,
			IntervalCek:   payload.IntervalCek,
			TotalCekHari:  payload.TotalCekHari,
			AutoLogout:    payload.AutoLogout,
			JamAktifStart: payload.JamAktifStart,
			JamAktifEnd:   payload.JamAktifEnd,
			Aktif:         payload.Aktif,
		}

		passwordEncoded := utils.EncodePassword(data.Password.String)
		data.Password.String = passwordEncoded

		err = uc.Update(ctx, data, uint64(id))
		if err != nil {
			return errors.InternalServerError.New(err.Error())
		}

		return c.NoContent(http.StatusOK)
	}
}
