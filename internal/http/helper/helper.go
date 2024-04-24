package helper

import (
	"context"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/pkg/constant"
)

func ExtractUserFromContext(ctx context.Context) (*domain.User, error) {
	val := ctx.Value(constant.ContextKeyUser)

	if user, ok := val.(*domain.User); ok {
		return user, nil
	}

	return nil, errors.ErrUnauthorized
}
