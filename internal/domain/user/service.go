package user

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// Service contains business logic for user operations.
type Service struct {
	repo *Repository
}

// NewService creates a new user Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Register creates a new user after hashing the password.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("user service hash: %w", err)
	}

	u := &User{
		Name:         strings.TrimSpace(req.Name),
		Email:        email,
		PasswordHash: string(hash),
		Role:         RoleUser,
	}

	if err = s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	slog.Info("user registered", "id", u.ID.Hex(), "email", u.Email)
	return u, nil
}

// Authenticate verifies email/password and returns the user on success.
func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user service auth: %w", err)
	}
	if u == nil {
		return nil, ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return u, nil
}

// HashPassword hashes a plaintext password with bcrypt. Exported for testing.
func HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	return string(h), err
}

// CheckPassword compares a hash with a plaintext password. Exported for testing.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// ErrInvalidCredentials indicates wrong email or password.
var ErrInvalidCredentials = fmt.Errorf("invalid email or password")
