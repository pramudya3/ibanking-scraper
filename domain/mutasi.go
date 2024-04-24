package domain

import (
	"context"
	"ibanking-scraper/domain/types"

	"github.com/indrasaputra/hashids"
)

type (
	MutasiUsecase interface {
		Fetch(ctx context.Context, rekening string) ([]*Mutasi, error)
		FetchByDate(ctx context.Context, startDate, endDate string, rekening string, isGriyabayar bool, filter Filter) ([]*Mutasi, interface{}, error)
		FetchRekGriyabayar(ctx context.Context) ([]*Mutasi, error)
		FetchByRekening(ctx context.Context, rekening string, filter Filter) ([]*Mutasi, error)
		BulkInsertRekOnpay(ctx context.Context, mutasis []*Mutasi) error
		BulkInsertRekGriyabayar(ctx context.Context, mutasis []*Mutasi) error
	}

	MutasiRepository interface {
		Fetch(ctx context.Context, rekening string) ([]*Mutasi, error)
		FetchByDate(ctx context.Context, startDate, endDate string, rekening string, isGriyabayar bool, filter Filter) ([]*Mutasi, interface{}, error)
		FetchRekGriyabayar(ctx context.Context) ([]*Mutasi, error)
		FetchByRekening(ctx context.Context, rekening string, filter Filter) ([]*Mutasi, error)
		BulkInsertRekOnpay(ctx context.Context, mutasis []*Mutasi) error
		BulkInsertRekGriyabayar(ctx context.Context, mutasis []*Mutasi) error
	}

	Mutasi struct {
		ID              hashids.ID         `json:"id" db:"id"`
		TglEntri        types.PGDate       `json:"tgl_entri" db:"tgl_entri"`
		TglBank         types.PGDate       `json:"tgl_bank" db:"tgl_bank"`
		TipeBank        BankType           `json:"tipe_bank" db:"tipe_bank"`
		Rekening        string             `json:"rekening" db:"rekening"`
		PemilikRekening string             `json:"pemilik_rekening" db:"pemilik_rekening"`
		Keterangan      string             `json:"keterangan" db:"keterangan"`
		TipeMutasi      MutasiRekeningType `json:"tipe_mutasi" db:"tipe_mutasi"`
		Jumlah          int64              `json:"jumlah" db:"jumlah"`
		Saldo           int64              `json:"saldo" db:"saldo"`
		Terklaim        bool               `json:"terklaim" db:"terklaim"`
		Griyabayar      bool               `json:"griyabayar" db:"griyabayar"`
		IdTiket         types.NullInt      `json:"id_tiket" db:"id_tiket"`
		Catatan         types.NullString   `json:"catatan" db:"catatan"`
	}
)
