package unit

import (
	"testing"
	"time"

	"github.com/one-backend-go/internal/domain/product"
	"github.com/one-backend-go/internal/domain/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserToResponse(t *testing.T) {
	now := time.Now().UTC()
	id := primitive.NewObjectID()
	u := &user.User{
		ID:        id,
		Name:      "Jane Doe",
		Email:     "jane@example.com",
		Role:      user.RoleUser,
		CreatedAt: now,
	}

	resp := u.ToResponse()
	if resp.ID != id.Hex() {
		t.Errorf("ID = %q, want %q", resp.ID, id.Hex())
	}
	if resp.Name != "Jane Doe" {
		t.Errorf("Name = %q, want %q", resp.Name, "Jane Doe")
	}
	if resp.Email != "jane@example.com" {
		t.Errorf("Email = %q, want %q", resp.Email, "jane@example.com")
	}
	if resp.Role != "user" {
		t.Errorf("Role = %q, want %q", resp.Role, "user")
	}
}

func TestProductToResponse(t *testing.T) {
	now := time.Now().UTC()
	id := primitive.NewObjectID()
	p := &product.Product{
		ID:          id,
		Name:        "Pizza",
		Description: "Delicious",
		PriceCents:  1299,
		Category:    "pizza",
		ImageURL:    "https://example.com/pizza.jpg",
		IsAvailable: true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	resp := p.ToResponse()
	if resp.ID != id.Hex() {
		t.Errorf("ID = %q, want %q", resp.ID, id.Hex())
	}
	if resp.PriceCents != 1299 {
		t.Errorf("PriceCents = %d, want %d", resp.PriceCents, 1299)
	}
	if resp.Category != "pizza" {
		t.Errorf("Category = %q, want %q", resp.Category, "pizza")
	}
	if resp.ImageURL != "https://example.com/pizza.jpg" {
		t.Errorf("ImageURL = %q, want %q", resp.ImageURL, "https://example.com/pizza.jpg")
	}
}
