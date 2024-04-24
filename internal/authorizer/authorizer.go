package authorizer

import (
	"context"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/http/helper"
)

type Permit interface {
	GetPolicy() *domain.IAMPolicy
	GetUser() *domain.User
	IsPermitted() bool
}

type Authorizer interface {
	Authorize(ctx context.Context, action, asset string, opts ...Option) Permit
}

type Option func(ctx context.Context, authz *authorization)

type DefaultAuthz struct {
	iamUsecase domain.IAMUsecase
}

type authorization struct {
	User   *domain.User
	Policy *domain.IAMPolicy

	allow bool
}

func (a *authorization) GetPolicy() *domain.IAMPolicy {
	return a.Policy
}

func (a *authorization) GetUser() *domain.User {
	return a.User
}

func (a *authorization) IsPermitted() bool {
	return a.Policy != nil && a.allow
}

func New(iamUsecase domain.IAMUsecase) *DefaultAuthz {
	return &DefaultAuthz{
		iamUsecase: iamUsecase,
	}
}

func (a DefaultAuthz) Authorize(ctx context.Context, action, asset string, opts ...Option) Permit {
	authz := &authorization{}

	user, err := helper.ExtractUserFromContext(ctx)
	if err != nil {
		return authz
	}

	authz.User = user

	if user.IsAnonymous() {
		return authz
	}

	authz.Policy = a.iamUsecase.Authorize(ctx, user, action, asset)
	for _, opt := range opts {
		opt(ctx, authz)
	}

	return authz
}
