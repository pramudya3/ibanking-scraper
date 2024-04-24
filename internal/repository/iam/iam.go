package iam

import (
	"context"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/logger"

	"github.com/jmoiron/sqlx"
)

type defaultRepository struct {
	db *sqlx.DB
}

func NewIAMRepository(db *sqlx.DB) domain.IAMRepository {
	return defaultRepository{
		db: db,
	}
}

func (r defaultRepository) GetUserPolicies(ctx context.Context, user *domain.User) (domain.IAMPolicies, error) {
	var policies domain.IAMPolicies

	if err := r.db.SelectContext(ctx, &policies, `
		SELECT p.id, p.action, p.menu, p.asset, p.griyabayar, p.onpay
		FROM public.user_role_bindings rb
		LEFT JOIN public.user_role_policies pb ON pb.role_id = rb.role_id
		LEFT JOIN public.user_policies p ON p.id = pb.policy_id
		LEFT JOIN public.user_logins l ON l.id = rb.user_id
		WHERE rb.user_id = $1 and
		case 
			when l.griyabayar then p.griyabayar is true 
			when l.griyabayar is not true then p.onpay is true 
		end
	`, user.ID); err != nil {
		logger.Error(err)
		return nil, errors.Transform(err)
	}

	return policies, nil
}

func (r defaultRepository) FetchPolicies(ctx context.Context) ([]*domain.IAMPolicy, error) {
	var policies []*domain.IAMPolicy

	if err := r.db.SelectContext(ctx, &policies, `
		SELECT *
		FROM public.user_policies
		WHERE asset != '/'
	`); err != nil {
		return nil, err
	}

	return policies, nil
}
