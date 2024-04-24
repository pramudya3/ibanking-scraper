package domain

import (
	"database/sql/driver"
	"encoding/json"
)

const (
	MutasiRekeningTypeDebet MutasiRekeningType = iota
	MutasiRekeningTypeKredit
)

type MutasiRekeningType uint

func (c MutasiRekeningType) Value() (driver.Value, error) {
	return c.String(), nil
}

func (c *MutasiRekeningType) Scan(value interface{}) error {
	if val, ok := value.(string); ok {
		*c = c.getCode(val)
	}

	return nil
}

func (c MutasiRekeningType) MarshalJSON() ([]byte, error) {
	if c == MutasiRekeningTypeDebet {
		return json.Marshal(nil)
	}
	return json.Marshal(c.String())
}

func (c *MutasiRekeningType) UnmarshalJSON(data []byte) error {
	var t string
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	*c = c.getCode(t)
	return nil
}

func (c MutasiRekeningType) String() string {
	switch c {
	case MutasiRekeningTypeKredit:
		return "KREDIT"
	case MutasiRekeningTypeDebet:
		return "DEBET"
	default:
		return "UNKNOWN"
	}
}

func (c *MutasiRekeningType) getCode(val string) MutasiRekeningType {
	switch val {
	case "KREDIT":
		return MutasiRekeningTypeKredit
	case "DEBET":
		return MutasiRekeningTypeDebet
	default:
		return MutasiRekeningTypeDebet
	}
}
