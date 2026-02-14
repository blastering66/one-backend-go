package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repository handles persistence for refresh tokens.
type Repository struct {
	col *mongo.Collection
}

// NewRepository returns a new auth Repository.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("refresh_tokens")}
}

// CreateRefreshToken stores a new refresh token document.
func (r *Repository) CreateRefreshToken(ctx context.Context, rt *RefreshToken) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rt.ID = primitive.NewObjectID()
	rt.CreatedAt = time.Now().UTC()

	_, err := r.col.InsertOne(ctx, rt)
	if err != nil {
		return fmt.Errorf("auth repo create: %w", err)
	}
	return nil
}

// FindRefreshToken finds a valid (non-revoked, non-expired) refresh token.
func (r *Repository) FindRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var rt RefreshToken
	err := r.col.FindOne(ctx, bson.M{
		"token":   token,
		"revoked": false,
	}).Decode(&rt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("auth repo find: %w", err)
	}
	return &rt, nil
}

// RevokeRefreshToken marks an existing refresh token as revoked.
func (r *Repository) RevokeRefreshToken(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.col.UpdateByID(ctx, id, bson.M{"$set": bson.M{"revoked": true}})
	if err != nil {
		return fmt.Errorf("auth repo revoke: %w", err)
	}
	return nil
}

// RevokeAllForUser revokes all refresh tokens for a given user.
func (r *Repository) RevokeAllForUser(ctx context.Context, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.col.UpdateMany(ctx,
		bson.M{"user_id": userID, "revoked": false},
		bson.M{"$set": bson.M{"revoked": true}},
	)
	if err != nil {
		return fmt.Errorf("auth repo revokeAll: %w", err)
	}
	return nil
}
