package domain

import (
	"database/sql/driver"
	"encoding/json"
)

const (
	BankTypeUnknown BankType = iota
	BankTypeBCA
	BankTypeBNI
	BankTypeBRI
	BankTypeMandiri
)

type BankType uint

func (c BankType) Value() (driver.Value, error) {
	return c.String(), nil
}

func (c *BankType) Scan(value interface{}) error {
	if val, ok := value.(string); ok {
		*c = c.getCode(val)
	}

	return nil
}

func (c BankType) MarshalJSON() ([]byte, error) {
	if c == BankTypeUnknown {
		return json.Marshal(nil)
	}
	return json.Marshal(c.String())
}

func (c *BankType) UnmarshalJSON(data []byte) error {
	var t string
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	*c = c.getCode(t)
	return nil
}

func (c BankType) String() string {
	switch c {
	case BankTypeBCA:
		return "BCA"
	case BankTypeBNI:
		return "BNI"
	case BankTypeBRI:
		return "BRI"
	case BankTypeMandiri:
		return "Mandiri"
	default:
		return "Unknown"
	}
}

func (c *BankType) getCode(val string) BankType {
	switch val {
	case "BCA":
		return BankTypeBCA
	case "BNI":
		return BankTypeBNI
	case "BRI":
		return BankTypeBRI
	case "Mandiri":
		return BankTypeMandiri
	default:
		return BankTypeUnknown
	}
}
