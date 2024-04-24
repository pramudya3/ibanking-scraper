package domain

import (
	"ibanking-scraper/domain/types"
)

type Auditable struct {
	CreatedAt types.NullTime   `json:"createdAt" db:"created_at"`
	CreatedBy types.NullString `json:"createdBy" db:"created_by"`
	UpdatedAt types.NullTime   `json:"updatedAt,omitempty" db:"updated_at"`
	UpdatedBy types.NullString `json:"updatedBy,omitempty" db:"updated_by"`
	DisableAt types.NullTime   `json:"disableAt,omitempty" db:"disable_at"`
	DisableBy types.NullString `json:"disableBy,omitempty" db:"disable_by"`
	DeletedAt types.NullTime   `json:"deletedAt,omitempty" db:"deleted_at"`
	DeletedBy types.NullString `json:"deletedBy,omitempty" db:"deleted_by"`
}
