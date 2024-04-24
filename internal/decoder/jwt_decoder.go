package decoder

import (
	"context"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/logger"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/indrasaputra/hashids"
)

type jwtDecoder struct {
	userUsecase domain.UserUsecase
}

func NewJWTDecoder(uc domain.UserUsecase) domain.JWTDecoder {
	return &jwtDecoder{
		userUsecase: uc,
	}
}

func (d *jwtDecoder) Decode(ctx context.Context, tokenString string, secret string) (*domain.User, error) {
	token, err := VerifyJWT(tokenString, secret)
	if err != nil {
		return nil, err
	}

	if err := d.userUsecase.ValidateToken(ctx, token); err != nil {
		return nil, err
	}

	return &domain.User{
		ID: token.UserID,
		Role: &domain.IAMRole{
			Name: token.Tipe,
		},
		Griyabayar: token.Griyabayar,
	}, nil
}

func VerifyJWT(tokenString string, secretKey string) (*domain.TokenDetail, error) {
	jwtKeyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := errors.Newf("unexpected signing method for jwt: %v", token.Header["alg"])
			return nil, err
		}

		return []byte(secretKey), nil
	}

	jwtToken, err := jwt.Parse(tokenString, jwtKeyFunc)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "signature is invalid"):
			return nil, errors.ErrJWTInvalid
		case strings.Contains(err.Error(), "token is expired"):
			return nil, errors.ErrJWTExpired
		default:
			return nil, errors.Unauthorized.New(err.Error())
		}
	}

	if !jwtToken.Valid {
		return nil, errors.ErrJWTInvalid
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.ErrJWTInvalid
	}

	token := &domain.TokenDetail{}

	if val, ok := claims["uid"]; ok {
		if str, ok := val.(string); ok {
			id, err := hashids.DecodeHash([]byte(str))
			if err != nil {
				return nil, errors.ErrUnauthorized
			}

			token.UserID = id
		}
	}

	if val, ok := claims["access_uuid"]; ok {
		if str, ok := val.(string); ok {
			token.AccessUuid = str
		}
	}

	if val, ok := claims["refresh_uuid"]; ok {
		if str, ok := val.(string); ok {
			token.RefreshUuid = str
		}
	}

	if val, ok := claims["griyabayar"]; ok {
		if str, ok := val.(bool); ok {
			token.Griyabayar = str
		}
	}

	if val, ok := claims["utype"]; ok {
		if str, ok := val.(string); ok {
			if err := token.Tipe.Scan(str); err != nil {
				logger.Error(err)
			}
		}
	}

	return token, nil
}
