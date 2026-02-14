package product

import "time"

// ── Request DTOs ───────────────────────────────────────────────────────────────

// CreateRequest is the body for POST /api/v1/products (admin only).
type CreateRequest struct {
	Name        string `json:"name"         validate:"required,min=2,max=80"`
	Description string `json:"description"  validate:"max=1000"`
	PriceCents  int64  `json:"price_cents"  validate:"gte=0"`
	Category    string `json:"category"     validate:"required"`
	ImageURL    string `json:"image_url"    validate:"omitempty,url"`
	IsAvailable *bool  `json:"is_available"`
}

// UpdateRequest is the body for PUT /api/v1/products/:id (admin only).
type UpdateRequest struct {
	Name        *string `json:"name"         validate:"omitempty,min=2,max=80"`
	Description *string `json:"description"  validate:"omitempty,max=1000"`
	PriceCents  *int64  `json:"price_cents"  validate:"omitempty,gte=0"`
	Category    *string `json:"category"     validate:"omitempty,min=1"`
	ImageURL    *string `json:"image_url"    validate:"omitempty,url"`
	IsAvailable *bool   `json:"is_available"`
}

// ── Response DTOs ──────────────────────────────────────────────────────────────

// Response is the API representation of a product.
type Response struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url,omitempty"`
	IsAvailable bool      `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListResponse is the paginated product list envelope.
type ListResponse struct {
	Items      []Response `json:"items"`
	Page       int64      `json:"page"`
	PageSize   int64      `json:"page_size"`
	Total      int64      `json:"total"`
	TotalPages int64      `json:"total_pages"`
}

// ToResponse converts a Product model to its public response form.
func (p *Product) ToResponse() Response {
	return Response{
		ID:          p.ID.Hex(),
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		Category:    p.Category,
		ImageURL:    p.ImageURL,
		IsAvailable: p.IsAvailable,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
