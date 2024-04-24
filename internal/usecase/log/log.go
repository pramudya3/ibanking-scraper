package log

import (
	"context"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/types"
	"time"
)

type LogUsecase struct {
	LogRepo domain.LogRepository
	Timeout time.Duration
}

func NewLogUsecase(logRepo domain.LogRepository, timeout types.ServerContextTimeoutDuration) domain.LogUsecase {
	return &LogUsecase{
		LogRepo: logRepo,
		Timeout: time.Duration(timeout),
	}
}

func (c *LogUsecase) FetchByAkunBank(ctx context.Context, akunBank uint64, filter domain.Filter) ([]*domain.Log, interface{}, error) {
	return c.LogRepo.FetchByAkunBank(ctx, akunBank, filter)
}

func (c *LogUsecase) Create(ctx context.Context, log *domain.Log) error {
	return c.LogRepo.Create(ctx, log)
}
