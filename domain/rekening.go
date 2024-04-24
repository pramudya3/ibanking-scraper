package domain

import (
	"context"
	"ibanking-scraper/domain/types"

	"github.com/indrasaputra/hashids"
)

type (
	RekeningUsecase interface {
		BulkInsertRekOnpay(ctx context.Context, rekenings []*Rekening) error
		BulkInsertRekGriyabayar(ctx context.Context, rekenings []*Rekening) error
	}

	RekeningRepository interface {
		BulkInsertRekOnpay(ctx context.Context, rekenings []*Rekening) error
		BulkInsertRekGriyabayar(ctx context.Context, rekenings []*Rekening) error
	}

	Rekening struct {
		ID              hashids.ID     `json:"id" db:"id"`
		TipeBank        BankType       `json:"tipe_bank" db:"tipe_bank"`
		Rekening        string         `json:"rekening" db:"rekening"`
		PemilikRekening string         `json:"pemilik_rekening" db:"pemilik_rekening"`
		Saldo           int64          `json:"saldo" db:"saldo"`
		Griyabayar      bool           `json:"griyabayar" db:"griyabayar"`
		TglUpdate       types.NullTime `json:"tgl_update,omitempty" db:"tgl_update"`
		IDAkunBank      hashids.ID     `json:"id_akun_bank" db:"id_akun_bank"`
		SaldoStr        string         `json:"-" db:"-"`
	}
)
