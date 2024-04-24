package domain

import (
	"context"
)

type JWTDecoder interface {
	Decode(ctx context.Context, token, secret string) (*User, error)
}
