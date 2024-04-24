package domain

import (
	"context"
	"fmt"
	"ibanking-scraper/domain/types"

	"github.com/indrasaputra/hashids"
)

type (
	UserUsecase interface {
		ValidateToken(ctx context.Context, td *TokenDetail) error
	}

	UserRepository interface {
		FetchTokenSession(ctx context.Context, td *TokenDetail) (string, error)
	}

	User struct {
		ID         hashids.ID       `db:"id" json:"id"`
		Username   string           `db:"username" json:"username"`
		Nama       string           `db:"nama" json:"nama"`
		Password   types.Password   `db:"password" json:"password"`
		Nomorhp    string           `db:"nomorhp" json:"nomorhp"`
		Lokasi     types.NullString `db:"lokasi" json:"lokasi"`
		DeviceID   types.NullString `db:"device_id" json:"deviceID"`
		OTP        types.Password   `db:"otp" json:"otp"`
		ExpiredOTP types.NullTime   `db:"expired_otp" json:"expiredOTP"`
		Griyabayar bool             `db:"griyabayar" json:"griyabayar"`
		Expired    bool             `db:"expired" json:"-"`
		Disabled   bool             `db:"disabled" json:"-"`

		Role *IAMRole

		Auditable
	}

	TokenDetail struct {
		AccessToken  string
		RefreshToken string
		AccessUuid   string
		RefreshUuid  string
		AtExpires    int64
		RtExpires    int64

		UserID     hashids.ID
		Tipe       TipeUser
		Griyabayar bool
	}
)

var AnonymousUser = &User{}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func (u *User) IsLoginDisabled() bool {
	return u.DisableAt.Valid
}

func (td *TokenDetail) GetAccessSessionKey() string {
	return GetAccessSessionPattern(td.UserID) + ":" + td.AccessUuid
}

func (td *TokenDetail) GetRefreshSessionKey() string {
	return GetRefreshSessionPattern(td.UserID) + ":" + td.RefreshUuid
}

func (td *TokenDetail) MatchSession(userID string) bool {
	return userID == td.UserID.EncodeString()
}

func GetAccessSessionPattern(userID hashids.ID) string {
	return fmt.Sprintf("access:%d", userID)
}

func GetRefreshSessionPattern(userID hashids.ID) string {
	return fmt.Sprintf("refresh:%d", userID)
}
