package utils

import (
	"ibanking-scraper/domain/types"

	"github.com/lib/pq"
)

func ToStringArray(datas []types.NullString) pq.StringArray {
	pqArrays := pq.StringArray{}
	for _, data := range datas {
		if data.Valid {
			pqArrays = append(pqArrays, data.String)
		}
	}
	return pqArrays
}
