package mutasi

import (
	"context"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/types"
	"time"
)

type MutasiUsecase struct {
	MutasiRepo domain.MutasiRepository
	timeout    time.Duration
}

func (c *MutasiUsecase) FetchByRekening(ctx context.Context, rekening string, filter domain.Filter) ([]*domain.Mutasi, error) {
	//TODO implement me
	panic("implement me")
}

func (c *MutasiUsecase) FetchByDate(ctx context.Context, startDate, endDate string, rekening string, isGriyabayar bool, filter domain.Filter) ([]*domain.Mutasi, interface{}, error) {
	return c.MutasiRepo.FetchByDate(ctx, startDate, endDate, rekening, isGriyabayar, filter)
}

func (c *MutasiUsecase) Fetch(ctx context.Context, rekening string) ([]*domain.Mutasi, error) {
	return c.MutasiRepo.Fetch(ctx, rekening)
}
func (c *MutasiUsecase) FetchRekGriyabayar(ctx context.Context) ([]*domain.Mutasi, error) {
	return c.MutasiRepo.FetchRekGriyabayar(ctx)
}

func (c *MutasiUsecase) BulkInsertRekOnpay(ctx context.Context, mutasis []*domain.Mutasi) error {
	return c.MutasiRepo.BulkInsertRekOnpay(ctx, mutasis)
}

func (c *MutasiUsecase) BulkInsertRekGriyabayar(ctx context.Context, mutasis []*domain.Mutasi) error {
	return c.MutasiRepo.BulkInsertRekGriyabayar(ctx, mutasis)
}

func NewMutasiUsecase(MutasiRepo domain.MutasiRepository, timeout types.ServerContextTimeoutDuration) domain.MutasiUsecase {
	return &MutasiUsecase{
		MutasiRepo: MutasiRepo,
		timeout:    time.Duration(timeout),
	}
}
