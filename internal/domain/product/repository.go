package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/one-backend-go/internal/pkg/pagination"
)

// Repository provides persistence operations for products.
type Repository struct {
	col *mongo.Collection
}

// NewRepository returns a new product Repository.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("products")}
}

// ListFilter holds optional filters for the product listing.
type ListFilter struct {
	Query    string // text search
	Category string // exact match
}

// List returns a paginated, filtered, and sorted list of products.
func (r *Repository) List(ctx context.Context, filter ListFilter, p pagination.Params) ([]Product, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	f := bson.M{}
	if filter.Query != "" {
		f["$text"] = bson.M{"$search": filter.Query}
	}
	if filter.Category != "" {
		f["category"] = filter.Category
	}

	total, err := r.col.CountDocuments(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("product repo count: %w", err)
	}

	sortOrder := -1
	if p.Order == "asc" {
		sortOrder = 1
	}

	sortField := "created_at"
	switch p.Sort {
	case "name", "price_cents", "created_at":
		sortField = p.Sort
	case "price":
		sortField = "price_cents"
	}

	opts := options.Find().
		SetSkip(p.Skip()).
		SetLimit(p.PageSize).
		SetSort(bson.D{{Key: sortField, Value: sortOrder}})

	cursor, err := r.col.Find(ctx, f, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("product repo find: %w", err)
	}
	defer cursor.Close(ctx)

	var products []Product
	if err = cursor.All(ctx, &products); err != nil {
		return nil, 0, fmt.Errorf("product repo decode: %w", err)
	}

	return products, total, nil
}

// Create inserts a new product document.
func (r *Repository) Create(ctx context.Context, p *Product) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	p.ID = primitive.NewObjectID()
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := r.col.InsertOne(ctx, p)
	if err != nil {
		return fmt.Errorf("product repo create: %w", err)
	}
	return nil
}

// FindByID retrieves a single product by its ObjectID.
func (r *Repository) FindByID(ctx context.Context, id primitive.ObjectID) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var p Product
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("product repo findByID: %w", err)
	}
	return &p, nil
}

// Update modifies an existing product document.
func (r *Repository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update["updated_at"] = time.Now().UTC()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var p Product
	err := r.col.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": update}, opts).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("product repo update: %w", err)
	}
	return &p, nil
}

// Delete removes a product by its ObjectID. Returns true if a document was deleted.
func (r *Repository) Delete(ctx context.Context, id primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, fmt.Errorf("product repo delete: %w", err)
	}
	return res.DeletedCount > 0, nil
}

// InsertMany bulk-inserts products (used for seeding).
func (r *Repository) InsertMany(ctx context.Context, products []Product) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	docs := make([]interface{}, len(products))
	now := time.Now().UTC()
	for i := range products {
		products[i].ID = primitive.NewObjectID()
		products[i].CreatedAt = now
		products[i].UpdatedAt = now
		docs[i] = products[i]
	}

	_, err := r.col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("product repo insertMany: %w", err)
	}
	return nil
}
