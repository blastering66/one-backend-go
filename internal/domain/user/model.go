// Package user contains the User domain model.
package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a registered user in the system.
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name"          json:"name"`
	Email        string             `bson:"email"         json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"` // never serialized to JSON
	Role         string             `bson:"role"          json:"role"`
	CreatedAt    time.Time          `bson:"created_at"    json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"    json:"updated_at"`
}

// RoleUser is the default role for newly registered users.
const RoleUser = "user"

// RoleAdmin is the administrative role.
const RoleAdmin = "admin"
