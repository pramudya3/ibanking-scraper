package bank_account

import (
	"context"
	"ibanking-scraper/domain"
	types2 "ibanking-scraper/domain/types"
	"ibanking-scraper/internal/types"
	"time"

	"github.com/lib/pq"
)

type BankAccountUsecase struct {
	BankAccountRepo domain.BankAccountRepository
	timeout         time.Duration
}

func (c *BankAccountUsecase) FetchToken(ctx context.Context, id uint64) (pq.StringArray, error) {
	return c.BankAccountRepo.FetchToken(ctx, id)
}

func (c *BankAccountUsecase) GetToken(ctx context.Context, id uint64) (types2.NullString, error) {
	return c.BankAccountRepo.GetToken(ctx, id)
}

func (c *BankAccountUsecase) DeleteToken(ctx context.Context, token string, id uint64) error {
	return c.BankAccountRepo.DeleteToken(ctx, token, id)
}

func (c *BankAccountUsecase) AddToken(ctx context.Context, id uint64, token pq.StringArray) error {
	return c.BankAccountRepo.AddToken(ctx, id, token)
}

func (c *BankAccountUsecase) Fetch(ctx context.Context) ([]*domain.BankAccount, error) {
	return c.BankAccountRepo.Fetch(ctx)
}

func (c *BankAccountUsecase) FindByID(ctx context.Context, id uint64) (*domain.BankAccount, error) {
	return c.BankAccountRepo.FindByID(ctx, id)
}

func (c *BankAccountUsecase) Create(ctx context.Context, bankAccount *domain.BankAccount) error {
	return c.BankAccountRepo.Create(ctx, bankAccount)
}

func (c *BankAccountUsecase) Update(ctx context.Context, bankAccount *domain.BankAccount, id uint64) error {
	return c.BankAccountRepo.Update(ctx, bankAccount, id)
}

func (c *BankAccountUsecase) UpdateLoginStatus(ctx context.Context, bankAccount *domain.BankAccount, id uint64) error {
	return c.BankAccountRepo.UpdateLoginStatus(ctx, bankAccount, id)
}

func (c *BankAccountUsecase) Delete(ctx context.Context, id uint64) error {
	return c.BankAccountRepo.Delete(ctx, id)
}

func NewBankAccountUsecase(BankAccountRepo domain.BankAccountRepository, timeout types.ServerContextTimeoutDuration) domain.BankAccountUsecase {
	return &BankAccountUsecase{
		BankAccountRepo: BankAccountRepo,
		timeout:         time.Duration(timeout),
	}
}
