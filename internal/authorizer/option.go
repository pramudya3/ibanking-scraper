package authorizer

import (
	"context"
	"ibanking-scraper/domain"
)

// ForUser allow the user only if the user is granted
func ForUser(utypes ...domain.TipeUser) Option {
	return func(ctx context.Context, authz *authorization) {
		authz.allow = (authz.User.Role.Name.In(utypes...)) && (authz.Policy != nil)
	}
}

// ForUserWithoutRBAC the opposite from ForUser, it is allowing user type without RBAC check
func ForUserWithoutRBAC(utypes ...domain.TipeUser) Option {
	return func(ctx context.Context, authz *authorization) {
		authz.allow = authz.User.Role.Name.In(utypes...)
	}
}

// ForAuthenticatedUser allow any authenticated user
func ForAuthenticatedUser() Option {
	return func(ctx context.Context, authz *authorization) {
		if !authz.User.IsAnonymous() {
			authz.allow = true
		}
	}
}
