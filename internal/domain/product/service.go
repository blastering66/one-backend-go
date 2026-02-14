package product

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/one-backend-go/internal/pkg/pagination"
)

// Service contains business logic for products.
type Service struct {
	repo *Repository
}

// NewService creates a new product Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List returns a paginated, filtered product listing.
func (s *Service) List(ctx context.Context, filter ListFilter, p pagination.Params) (*ListResponse, error) {
	p.Clamp()

	products, total, err := s.repo.List(ctx, filter, p)
	if err != nil {
		return nil, fmt.Errorf("product service list: %w", err)
	}

	items := make([]Response, 0, len(products))
	for i := range products {
		items = append(items, products[i].ToResponse())
	}

	return &ListResponse{
		Items:      items,
		Page:       p.Page,
		PageSize:   p.PageSize,
		Total:      total,
		TotalPages: pagination.TotalPages(total, p.PageSize),
	}, nil
}

// Create adds a new product to the catalog.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Product, error) {
	available := true
	if req.IsAvailable != nil {
		available = *req.IsAvailable
	}

	p := &Product{
		Name:        req.Name,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
		IsAvailable: available,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Update modifies an existing product.
func (s *Service) Update(ctx context.Context, idHex string, req UpdateRequest) (*Product, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return nil, fmt.Errorf("invalid product id")
	}

	update := bson.M{}
	if req.Name != nil {
		update["name"] = *req.Name
	}
	if req.Description != nil {
		update["description"] = *req.Description
	}
	if req.PriceCents != nil {
		update["price_cents"] = *req.PriceCents
	}
	if req.Category != nil {
		update["category"] = *req.Category
	}
	if req.ImageURL != nil {
		update["image_url"] = *req.ImageURL
	}
	if req.IsAvailable != nil {
		update["is_available"] = *req.IsAvailable
	}

	if len(update) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	p, err := s.repo.Update(ctx, id, update)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	return p, nil
}

// Delete removes a product from the catalog.
func (s *Service) Delete(ctx context.Context, idHex string) error {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return fmt.Errorf("invalid product id")
	}

	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrProductNotFound
	}
	return nil
}

// ErrProductNotFound indicates the product does not exist.
var ErrProductNotFound = fmt.Errorf("product not found")
