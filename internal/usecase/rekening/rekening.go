package rekening

import (
	"context"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/types"
	"time"
)

type RekeningUsecase struct {
	RekeningRepo domain.RekeningRepository
	timeout      time.Duration
}

func (c *RekeningUsecase) BulkInsertRekOnpay(ctx context.Context, rekenings []*domain.Rekening) error {
	return c.RekeningRepo.BulkInsertRekOnpay(ctx, rekenings)
}

func (c *RekeningUsecase) BulkInsertRekGriyabayar(ctx context.Context, rekenings []*domain.Rekening) error {
	return c.RekeningRepo.BulkInsertRekGriyabayar(ctx, rekenings)
}

func NewRekeningUsecase(RekeningRepo domain.RekeningRepository, timeout types.ServerContextTimeoutDuration) domain.RekeningUsecase {
	return &RekeningUsecase{
		RekeningRepo: RekeningRepo,
		timeout:      time.Duration(timeout),
	}
}
