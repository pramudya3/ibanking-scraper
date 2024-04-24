package iam

import (
	"context"
	"ibanking-scraper/domain"
	"ibanking-scraper/domain/rbac"
	"strings"

	"github.com/thoas/go-funk"
)

type defaultUsecase struct {
	iamRepository domain.IAMRepository
}

func NewIAMUsecase(iamRepository domain.IAMRepository) domain.IAMUsecase {
	return defaultUsecase{
		iamRepository: iamRepository,
	}
}

func (r defaultUsecase) Authorize(ctx context.Context, user *domain.User, action, asset string) *domain.IAMPolicy {
	userPolicies, err := r.iamRepository.GetUserPolicies(ctx, user)
	if err != nil {
		return nil
	}

	policies, err := r.iamRepository.FetchPolicies(ctx)
	if err != nil {
		return nil
	}

	assetOwnerRaw := funk.Find(policies, func(policy *domain.IAMPolicy) bool {
		return policy.Action.FromHTTPMethod(action) && strings.HasPrefix(asset, policy.Asset)
	})

	var assetOwner *domain.IAMPolicy
	if assetOwnerRaw != nil {
		assetOwner = assetOwnerRaw.(*domain.IAMPolicy)
	} else {
		assetOwner = &domain.IAMPolicy{
			Policy: rbac.Policy{
				Asset:      asset,
				Griyabayar: false,
				Onpay:      true,
			},
			ID: "",
		}
	}

	return userPolicies.Authorize(action, asset, assetOwner)
}
