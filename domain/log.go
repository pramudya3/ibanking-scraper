package domain

import (
	"context"
	"ibanking-scraper/domain/types"

	"github.com/indrasaputra/hashids"
)

type (
	LogUsecase interface {
		FetchByAkunBank(ctx context.Context, id uint64, filter Filter) ([]*Log, interface{}, error)
		Create(ctx context.Context, log *Log) error
	}

	LogRepository interface {
		FetchByAkunBank(ctx context.Context, id uint64, filter Filter) ([]*Log, interface{}, error)
		Create(ctx context.Context, log *Log) error
	}

	Log struct {
		ID         uint64         `json:"id" db:"id"`
		TglEntri   types.NullTime `json:"tgl_entri" db:"tgl_entri"`
		AkunBankId hashids.ID     `json:"akun_bank_id" db:"akun_bank_id"`
		Tipe       LogType        `json:"tipe" db:"tipe"`
		Keterangan string         `json:"keterangan" db:"keterangan"`
	}
)
