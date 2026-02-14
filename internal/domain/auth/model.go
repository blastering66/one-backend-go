// Package auth contains the refresh token model and JWT helpers.
package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RefreshToken represents a server-side refresh token stored in MongoDB.
type RefreshToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Token     string             `bson:"token"`
	ExpiresAt time.Time          `bson:"expires_at"`
	Revoked   bool               `bson:"revoked"`
	CreatedAt time.Time          `bson:"created_at"`
}

// TokenResponse is returned by login and refresh endpoints.
type TokenResponse struct {
	AccessToken          string `json:"access_token"`
	AccessTokenExpiresIn int    `json:"access_token_expires_in"` // seconds
	RefreshToken         string `json:"refresh_token"`
	TokenType            string `json:"token_type"`
}

// RefreshRequest is the body for POST /api/v1/auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
