package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repository provides persistence operations for users.
type Repository struct {
	col *mongo.Collection
}

// NewRepository returns a new user Repository.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("users")}
}

// Create inserts a new user document.
func (r *Repository) Create(ctx context.Context, u *User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	u.ID = primitive.NewObjectID()
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now

	_, err := r.col.InsertOne(ctx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrEmailExists
		}
		return fmt.Errorf("user repo create: %w", err)
	}
	return nil
}

// FindByEmail retrieves a user by email address.
func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var u User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("user repo findByEmail: %w", err)
	}
	return &u, nil
}

// FindByID retrieves a user by their ObjectID.
func (r *Repository) FindByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var u User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("user repo findByID: %w", err)
	}
	return &u, nil
}

// ErrEmailExists indicates a duplicate email during registration.
var ErrEmailExists = fmt.Errorf("email already exists")
