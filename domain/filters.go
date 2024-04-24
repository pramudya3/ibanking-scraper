package domain

import (
	"fmt"
	"ibanking-scraper/internal/tool"
	"ibanking-scraper/internal/validator/v1"
	"math"
	"strings"
)

type (
	Filter struct {
		Page         int
		PageSize     int
		Sort         string
		SortSafeList []string
	}

	FilterMetadata struct {
		CurrentPage  int `json:"currentPage"`
		PageSize     int `json:"pageSize"`
		FirstPage    int `json:"firstPage"`
		LastPage     int `json:"lastPage"`
		TotalRecords int `json:"totalRecords"`
	}
)

func (f Filter) Validate(v *validator.Validator) error {
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", fmt.Sprintf("invalid sort value, consider this value to sort: %s", strings.Join(f.SortSafeList, ", ")))

	if !v.Valid() {
		return v.Err()
	}
	return nil
}

func (f Filter) SortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return tool.ToSnakeCase(strings.TrimPrefix(f.Sort, "-"))
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filter) SortDirection() string {
	if strings.HasPrefix(f.Sort, "_") {
		return "DESC"
	}
	return "ASC"
}

func (f Filter) Limit() int {
	return f.PageSize
}

func (f Filter) Offset() int {
	return (f.Page - 1) * f.PageSize
}

func (f Filter) BuildMetadata(totalRecords int) FilterMetadata {
	return FilterMetadata{
		CurrentPage:  f.Page,
		PageSize:     f.PageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(f.PageSize))),
		TotalRecords: totalRecords,
	}
}
