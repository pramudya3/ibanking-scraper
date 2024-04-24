package domain

import (
	"fmt"
	"ibanking-scraper/internal/builder/query"
	"ibanking-scraper/internal/validator/v1"
	"ibanking-scraper/pkg/constant"
	"strings"
	"time"
)

type QueryParams struct {
	QueryKV                map[string]interface{}
	QuerySafeMapList       map[string][]string
	SkipKeys               []string
	SQLQueryBuilderMapFunc map[string]func(value interface{}) query.SQLQueryBuilder
}

func (q QueryParams) Validate(v *validator.Validator) error {
	for k, val := range q.QueryKV {
		if validator.In(k, q.SkipKeys...) {
			continue
		}

		if k == "daterange" {
			val, ok := val.([]string)
			if !ok {
				continue
			}

			if len(val) != 2 {
				continue
			}

			start, end := val[0], val[1]

			switch {
			case start == "-" && end == "-":
				continue
			case start != "-" && end == "-":
				v.Check(false, "startDate", "require end date param")
				continue
			case start == "-" && end != "-":
				v.Check(false, "endDate", "require start date param")
				continue
			}

			v.Check(validator.Date(start), "startDate", "bad format for date")
			v.Check(validator.Date(end), "startDate", "bad format for date")

			startDate, _ := time.Parse(constant.LayoutDateISO8601, start)
			endDate, _ := time.Parse(constant.LayoutDateISO8601, end)
			v.Check(startDate.Before(endDate), "startDate", "start date is happen after end date")
			v.Check(endDate.After(startDate), "endDate", "end date is happen before start date")

			continue
		}

		switch val := val.(type) {
		case string:
			v.Check(validator.In(strings.ToLower(val), q.QuerySafeMapList[k]...), k, fmt.Sprintf("unsupported value for '%s' param", k))
		}
	}

	if !v.Valid() {
		return v.Err()
	}

	return nil
}
