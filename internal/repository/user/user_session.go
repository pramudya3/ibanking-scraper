package user

import (
	"context"
	"fmt"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"

	"github.com/indrasaputra/hashids"
)

func (c *userRepository) FetchTokenSession(ctx context.Context, td *domain.TokenDetail) (string, error) {
	var (
		query  string
		userID hashids.ID
	)

	if td.RefreshUuid != "" {
		query = fmt.Sprintf(`WHERE refresh_token='%s'`, td.GetRefreshSessionKey())
	}

	if td.AccessUuid != "" {
		query = fmt.Sprintf(`WHERE access_token='%s'`, td.GetAccessSessionKey())
	}

	if err := c.db.GetContext(ctx, &userID, `SELECT id FROM public.user_logins `+query); err != nil {
		return "", errors.Transform(err)
	}

	return userID.EncodeString(), nil
}
