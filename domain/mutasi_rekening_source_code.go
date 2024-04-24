package domain

import (
	"database/sql/driver"
	"encoding/json"
)

const (
	MutasiRekeningSourceUnknown MutasiRekeningSource = iota
	MutasiRekeningSourceBRIPersonal
	MutasiRekeningSourceBRICMS
	MutasiRekeningSourceKlikBCA
	MutasiRekeningSourceBNIDirect
	MutasiRekeningSourceBNIPersonal
)

type MutasiRekeningSource uint

func (c MutasiRekeningSource) Value() (driver.Value, error) {
	return c.String(), nil
}

func (c *MutasiRekeningSource) Scan(value interface{}) error {
	if val, ok := value.(string); ok {
		*c = c.getCode(val)
	}

	return nil
}

func (c MutasiRekeningSource) MarshalJSON() ([]byte, error) {
	if c == MutasiRekeningSourceUnknown {
		return json.Marshal(nil)
	}
	return json.Marshal(c.String())
}

func (c *MutasiRekeningSource) UnmarshalJSON(data []byte) error {
	var t string
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	*c = c.getCode(t)
	return nil
}

func (c MutasiRekeningSource) String() string {
	switch c {
	case MutasiRekeningSourceBRICMS:
		return "BRI CORPORATE"
	case MutasiRekeningSourceBRIPersonal:
		return "BRI PERSONAL"
	case MutasiRekeningSourceKlikBCA:
		return "KLIK BCA"
	case MutasiRekeningSourceBNIDirect:
		return "BNI DIRECT"
	case MutasiRekeningSourceBNIPersonal:
		return "BNI PERSONAL"
	default:
		return "UNKNOWN"
	}
}

func (c *MutasiRekeningSource) getCode(val string) MutasiRekeningSource {
	switch val {
	case "BRI CORPORATE":
		return MutasiRekeningSourceBRICMS
	case "BRI PERSONAL":
		return MutasiRekeningSourceBRIPersonal
	case "KLIK BCA":
		return MutasiRekeningSourceKlikBCA
	case "BNI DIRECT":
		return MutasiRekeningSourceBNIDirect
	case "BNI PERSONAL":
		return MutasiRekeningSourceBNIPersonal
	default:
		return MutasiRekeningSourceUnknown
	}
}
