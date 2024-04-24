package domain

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

const (
	UserRoot TipeUser = iota
	UserOwner
	UserOperator
	UserAdmin
	UserCS
	UserUnknown
)

type TipeUser uint

func (u *TipeUser) Scan(value interface{}) error {
	if val, ok := value.(string); ok {
		*u = NewTipeUser(val)
	}
	return nil

}

func (u TipeUser) Value() (driver.Value, error) {
	return u.String(), nil
}

func (c TipeUser) MarshalJSON() ([]byte, error) {
	if c == UserUnknown {
		return json.Marshal(nil)
	}
	return json.Marshal(c.String())
}

func (c *TipeUser) UnmarshalJSON(data []byte) error {
	var t string
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	*c = c.getCode(t)
	return nil
}

func (c *TipeUser) getCode(val string) TipeUser {
	switch strings.ToLower(val) {
	case "root":
		return UserRoot
	case "owner":
		return UserOwner
	case "operator":
		return UserOperator
	case "admin":
		return UserAdmin
	case "cs":
		return UserCS
	default:
		return UserUnknown
	}
}

func (u TipeUser) In(userType ...TipeUser) bool {
	for _, o := range userType {
		if o == u {
			return true
		}
	}
	return false
}

func (u TipeUser) Is(userTypeStr string) bool {
	return u.String() == userTypeStr
}

func (u TipeUser) String() string {
	return []string{
		"root",
		"owner",
		"operator",
		"admin",
		"cs",
		"unknown",
	}[u]
}

func NewTipeUser(name string) TipeUser {
	var u TipeUser
	switch strings.ToLower(name) {
	case "root":
		u = UserRoot
	case "owner":
		u = UserOwner
	case "operator":
		u = UserOperator
	case "admin":
		u = UserAdmin
	case "cs":
		u = UserCS
	default:
		u = UserUnknown
	}

	return u
}
