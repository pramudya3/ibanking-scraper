package bank_account

import (
	"database/sql"
	"ibanking-scraper/domain"
	"ibanking-scraper/domain/types"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/http/response"
	"ibanking-scraper/utils"
	"net/http"

	"github.com/lib/pq"

	"github.com/indrasaputra/hashids"
	"github.com/labstack/echo/v4"
)

type Response struct {
	ID            hashids.ID         `json:"id"`
	TipeAkun      domain.AccountType `json:"tipe_akun"`
	CompanyId     types.NullString   `json:"company_id"`
	UserId        types.NullString   `json:"user_id"`
	Password      types.NullString   `json:"password"`
	RekOnpay      pq.StringArray     `json:"rek_onpay"`
	RekGriyabayar pq.StringArray     `json:"rek_griyabayar"`
	TotalCekHari  types.NullInt      `json:"total_cek_hari"`
	IntervalCek   types.NullInt      `json:"interval_cek"`
	AutoLogout    types.NullBool     `json:"auto_logout"`
	JamAktifStart types.PGTime       `json:"jam_aktif_start"`
	JamAktifEnd   types.PGTime       `json:"jam_aktif_end"`
	Aktif         bool               `json:"aktif"`
	Token         pq.StringArray     `json:"token"`
}

func GetBankAccount(uc domain.BankAccountUsecase) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		ctx := c.Request().Context()
		str := c.Param("id")
		id, err := hashids.DecodeHash([]byte(str))
		if err != nil {
			return errors.ErrResourceNotFound
		}

		bankAccount, err := uc.FindByID(ctx, uint64(id))
		if err != nil {
			if err != sql.ErrNoRows {
				return errors.InternalServerError.New(err.Error())
			} else {
				return errors.ErrResourceNotFound
			}
		}

		resp := &Response{
			ID:            bankAccount.ID,
			TipeAkun:      bankAccount.TipeAkun,
			CompanyId:     bankAccount.CompanyId,
			UserId:        bankAccount.UserId,
			Password:      bankAccount.Password,
			RekOnpay:      bankAccount.RekOnpay,
			RekGriyabayar: bankAccount.RekGriyabayar,
			IntervalCek:   bankAccount.IntervalCek,
			TotalCekHari:  bankAccount.TotalCekHari,
			AutoLogout:    bankAccount.AutoLogout,
			JamAktifStart: bankAccount.JamAktifStart,
			JamAktifEnd:   bankAccount.JamAktifEnd,
			Aktif:         bankAccount.Aktif,
			Token:         bankAccount.Token,
		}

		passwordDecoded := utils.DecodePassword(bankAccount.Password.String)
		resp.Password.String = passwordDecoded

		return c.JSON(http.StatusOK, response.NewSuccess(resp, response.EmptyMeta{}))
	}
}
