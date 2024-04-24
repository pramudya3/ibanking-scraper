package bank_account

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/domain/types"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/utils"
	"net/http"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"

	"github.com/labstack/echo/v4"
)

type BankAccountRequest struct {
	TipeAkun      domain.AccountType `json:"tipe_akun"`
	CompanyId     types.NullString   `json:"company_id"`
	UserId        types.NullString   `json:"user_id" validate:"required"`
	Password      types.NullString   `json:"password" validate:"required"`
	RekOnpay      pq.StringArray     `json:"rek_onpay" validate:"unique"`
	RekGriyabayar pq.StringArray     `json:"rek_griyabayar" validate:"unique"`
	IntervalCek   types.NullInt      `json:"interval_cek"`
	TotalCekHari  types.NullInt      `json:"total_cek_hari"`
	AutoLogout    types.NullBool     `json:"auto_logout"`
	JamAktifStart types.PGTime       `json:"jam_aktif_start"`
	JamAktifEnd   types.PGTime       `json:"jam_aktif_end"`
	Aktif         bool               `json:"aktif"`
}

func AddBankAccount(uc domain.BankAccountUsecase) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		payload := BankAccountRequest{}

		if err := c.Bind(&payload); err != nil {
			return errors.BadRequest.New(err.Error())
		}

		validate := validator.New()

		data := &domain.BankAccount{
			TipeAkun:      payload.TipeAkun,
			CompanyId:     payload.CompanyId,
			UserId:        payload.UserId,
			Password:      payload.Password,
			RekOnpay:      payload.RekOnpay,
			RekGriyabayar: payload.RekGriyabayar,
			IntervalCek:   payload.IntervalCek,
			TotalCekHari:  payload.TotalCekHari,
			AutoLogout:    payload.AutoLogout,
			JamAktifStart: payload.JamAktifStart,
			JamAktifEnd:   payload.JamAktifEnd,
			Aktif:         payload.Aktif,
		}

		if err := validate.Struct(data); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		if len(data.RekOnpay) > 0 && len(data.RekGriyabayar) > 0 {
			for i := 0; i < len(data.RekOnpay); i++ {
				for j := 0; j < len(data.RekGriyabayar); j++ {
					if data.RekOnpay[i] == data.RekGriyabayar[j] {
						return errors.BadRequest.New("Rekening tidak boleh sama")
					}
				}
			}
		}

		passwordEncoded := utils.EncodePassword(payload.Password.String)
		data.Password.String = passwordEncoded

		var wg sync.WaitGroup

		wg.Add(1)
		if err := uc.Create(ctx, data); err != nil {
			return errors.InternalServerError.New(err.Error())
		}
		wg.Done()

		wg.Wait()

		return c.NoContent(http.StatusCreated)
	}
}
