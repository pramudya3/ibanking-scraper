package domain

import (
	"context"
	"ibanking-scraper/domain/rbac"

	"github.com/indrasaputra/hashids"
)

type (
	IAMUsecase interface {
		Authorize(ctx context.Context, user *User, action, asset string) *IAMPolicy
	}

	IAMRepository interface {
		GetUserPolicies(ctx context.Context, user *User) (IAMPolicies, error)
		FetchPolicies(ctx context.Context) ([]*IAMPolicy, error)
	}

	IAMRole struct {
		ID   hashids.ID `db:"id" json:"id"`
		Name TipeUser   `db:"name" json:"name"`

		Policies IAMPolicies `db:"policy" json:"-"`
	}

	IAMPolicy struct {
		rbac.Policy

		ID string `db:"id" json:"id"`
	}

	IAMPolicies []*IAMPolicy
)

func (ps IAMPolicies) Authorize(httpMethod, asset string, assetOwner *IAMPolicy) *IAMPolicy {
	for _, policy := range ps {
		if granted := policy.IsGranted(httpMethod, asset, assetOwner.Griyabayar, assetOwner.Onpay); granted {
			return policy
		}
	}

	return nil
}
