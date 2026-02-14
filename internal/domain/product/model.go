// Package product contains the Product domain model.
package product

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product represents a food item in the catalog.
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"  json:"id"`
	Name        string             `bson:"name"           json:"name"`
	Description string             `bson:"description"    json:"description"`
	PriceCents  int64              `bson:"price_cents"    json:"price_cents"`
	Category    string             `bson:"category"       json:"category"`
	ImageURL    string             `bson:"image_url"      json:"image_url,omitempty"`
	IsAvailable bool               `bson:"is_available"   json:"is_available"`
	CreatedAt   time.Time          `bson:"created_at"     json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"     json:"updated_at"`
}
