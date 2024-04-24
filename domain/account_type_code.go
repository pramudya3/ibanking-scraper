package domain

import (
	"database/sql/driver"
	"encoding/json"
)

const (
	AccountTypeUnknown AccountType = iota
	AccountTypeBCAPersonal
	AccountTypeBCABisnis
	AccountTypeBNIPersonal
	AccountTypeBNIDirect
	AccountTypeBRI
	AccountTypeBRICMS
	AccountTypeBRIIBBIZ
	AccountTypeMandiriMCM
)

type AccountType uint

func (c AccountType) Value() (driver.Value, error) {
	return c.String(), nil
}

func (c *AccountType) Scan(value interface{}) error {
	if val, ok := value.(string); ok {
		*c = c.getCode(val)
	}

	return nil
}

func (c AccountType) MarshalJSON() ([]byte, error) {
	if c == AccountTypeUnknown {
		return json.Marshal(nil)
	}
	return json.Marshal(c.String())
}

func (c *AccountType) UnmarshalJSON(data []byte) error {
	var t string
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	*c = c.getCode(t)
	return nil
}

func (c AccountType) String() string {
	switch c {
	case AccountTypeBCAPersonal:
		return "BCA Personal"
	case AccountTypeBCABisnis:
		return "BCA Bisnis"
	case AccountTypeBNIPersonal:
		return "BNI Personal"
	case AccountTypeBNIDirect:
		return "BNI Direct"
	case AccountTypeBRI:
		return "BRI"
	case AccountTypeBRICMS:
		return "BRI CMS"
	case AccountTypeBRIIBBIZ:
		return "BRI IBBIZ"
	case AccountTypeMandiriMCM:
		return "Mandiri MCM"
	default:
		return "Unknown"
	}
}

func (c *AccountType) getCode(val string) AccountType {
	switch val {
	case "BCA Personal":
		return AccountTypeBCAPersonal
	case "BCA Bisnis":
		return AccountTypeBCABisnis
	case "BNI Personal":
		return AccountTypeBNIPersonal
	case "BNI Direct":
		return AccountTypeBNIDirect
	case "BRI":
		return AccountTypeBRI
	case "BRI CMS":
		return AccountTypeBRICMS
	case "BRI IBBIZ":
		return AccountTypeBRIIBBIZ
	case "Mandiri MCM":
		return AccountTypeMandiriMCM
	default:
		return AccountTypeUnknown
	}
}
