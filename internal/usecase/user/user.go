package user

import (
	"context"
	"ibanking-scraper/config"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/types"
	"time"
)

type userUsecase struct {
	config         *config.Config
	userRepository domain.UserRepository
	timeout        time.Duration
}

func NewUserUsecase(
	config *config.Config,
	userRepository domain.UserRepository,
	timeout types.ServerContextTimeoutDuration,
) domain.UserUsecase {
	return &userUsecase{
		config:         config,
		userRepository: userRepository,
		timeout:        time.Duration(timeout),
	}
}

func (u *userUsecase) ValidateToken(ctx context.Context, td *domain.TokenDetail) error {
	var (
		userID string
		err    error
	)

	userID, err = u.userRepository.FetchTokenSession(ctx, td)
	if err != nil {
		return errors.ErrUnauthorized
	}

	if !td.MatchSession(userID) {
		return errors.ErrUnauthorized
	}

	return nil
}
