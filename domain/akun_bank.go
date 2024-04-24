package domain

import (
	"context"
	"ibanking-scraper/domain/types"

	"github.com/lib/pq"

	"github.com/indrasaputra/hashids"
)

type (
	BankAccountUsecase interface {
		Fetch(ctx context.Context) ([]*BankAccount, error)
		FindByID(ctx context.Context, id uint64) (*BankAccount, error)
		Create(ctx context.Context, rekening *BankAccount) error
		Update(ctx context.Context, rekening *BankAccount, id uint64) error
		UpdateLoginStatus(ctx context.Context, rekening *BankAccount, id uint64) error
		Delete(ctx context.Context, id uint64) error
		AddToken(ctx context.Context, id uint64, token pq.StringArray) error
		FetchToken(ctx context.Context, id uint64) (pq.StringArray, error)
		GetToken(ctx context.Context, id uint64) (types.NullString, error)
		DeleteToken(ctx context.Context, token string, id uint64) error
	}

	BankAccountRepository interface {
		Fetch(ctx context.Context) ([]*BankAccount, error)
		FindByID(ctx context.Context, id uint64) (*BankAccount, error)
		Create(ctx context.Context, rekening *BankAccount) error
		Update(ctx context.Context, rekening *BankAccount, id uint64) error
		UpdateLoginStatus(ctx context.Context, rekening *BankAccount, id uint64) error
		Delete(ctx context.Context, id uint64) error
		AddToken(ctx context.Context, id uint64, token pq.StringArray) error
		FetchToken(ctx context.Context, id uint64) (pq.StringArray, error)
		GetToken(ctx context.Context, id uint64) (types.NullString, error)
		DeleteToken(ctx context.Context, token string, id uint64) error
	}

	BankAccount struct {
		ID            hashids.ID       `json:"id" db:"id"`
		TipeAkun      AccountType      `json:"tipe_akun" db:"tipe_akun"`
		CompanyId     types.NullString `json:"company_id" db:"company_id"`
		UserId        types.NullString `json:"user_id" db:"user_id"`
		Password      types.NullString `json:"password" db:"password"`
		RekOnpay      pq.StringArray   `json:"rek_onpay" db:"rek_onpay" validate:"unique"`
		RekGriyabayar pq.StringArray   `json:"rek_griyabayar" db:"rek_griyabayar" validate:"unique"`
		TotalCekHari  types.NullInt    `json:"total_cek_hari" db:"total_cek_hari"`
		IntervalCek   types.NullInt    `json:"interval_cek" db:"interval_cek"`
		AutoLogout    types.NullBool   `json:"auto_logout" db:"auto_logout"`
		JamAktifStart types.PGTime     `json:"jam_aktif_start" db:"jam_aktif_start"`
		JamAktifEnd   types.PGTime     `json:"jam_aktif_end" db:"jam_aktif_end"`
		Aktif         bool             `json:"aktif" db:"aktif"`
		Token         pq.StringArray   `json:"token" db:"token"`
		StatusLogin   LoginStatus      `json:"status_login" db:"status_login"`
		TglAktivitas  types.NullTime   `json:"tgl_aktivitas" db:"tgl_aktivitas"`
		TglUpdate     types.NullTime   `json:"tgl_update,omitempty" db:"tgl_update"`
	}
)
