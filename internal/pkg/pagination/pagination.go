// Package pagination provides helpers for paginated queries.
package pagination

import "math"

// Params represents validated pagination parameters.
type Params struct {
	Page     int64  `json:"page"`
	PageSize int64  `json:"page_size"`
	Sort     string `json:"sort,omitempty"`
	Order    string `json:"order,omitempty"` // "asc" or "desc"
}

// DefaultParams returns pagination params with sensible defaults.
func DefaultParams() Params {
	return Params{
		Page:     1,
		PageSize: 10,
		Sort:     "created_at",
		Order:    "desc",
	}
}

// Clamp ensures page and page_size are within acceptable bounds.
func (p *Params) Clamp() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 10
	}
	if p.PageSize > 50 {
		p.PageSize = 50
	}
}

// Skip returns the number of documents to skip for the current page.
func (p *Params) Skip() int64 {
	return (p.Page - 1) * p.PageSize
}

// TotalPages calculates the total number of pages given total items.
func TotalPages(total, pageSize int64) int64 {
	if pageSize <= 0 {
		return 0
	}
	return int64(math.Ceil(float64(total) / float64(pageSize)))
}
