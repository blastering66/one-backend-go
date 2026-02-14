package user

import "time"

// ── Request DTOs ───────────────────────────────────────────────────────────────

// RegisterRequest is the body for POST /api/v1/auth/register.
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,name"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,strongpass"`
}

// LoginRequest is the body for POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ── Response DTOs ──────────────────────────────────────────────────────────────

// UserResponse is the safe representation of a user (no password).
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a User model to its public response form.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID.Hex(),
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
