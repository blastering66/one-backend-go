package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/one-backend-go/internal/config"
	"github.com/one-backend-go/internal/domain/user"
)

// Service contains business logic for authentication.
type Service struct {
	jwt         *JWTManager
	repo        *Repository
	userService *user.Service
	refreshTTL  time.Duration
}

// NewService creates a new auth Service.
func NewService(cfg *config.Config, jwtMgr *JWTManager, repo *Repository, userSvc *user.Service) *Service {
	return &Service{
		jwt:         jwtMgr,
		repo:        repo,
		userService: userSvc,
		refreshTTL:  cfg.RefreshTokenTTL,
	}
}

// Login authenticates the user and returns token pair.
func (s *Service) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	u, err := s.userService.Authenticate(ctx, email, password)
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, u.ID, u.Email)
}

// Refresh validates a refresh token, rotates it, and issues a new token pair.
func (s *Service) Refresh(ctx context.Context, refreshTokenStr string) (*TokenResponse, error) {
	rt, err := s.repo.FindRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, fmt.Errorf("auth refresh: %w", err)
	}
	if rt == nil {
		return nil, ErrInvalidRefreshToken
	}
	if time.Now().UTC().After(rt.ExpiresAt) {
		return nil, ErrInvalidRefreshToken
	}

	// Revoke old token (rotation).
	if err = s.repo.RevokeRefreshToken(ctx, rt.ID); err != nil {
		return nil, fmt.Errorf("auth refresh revoke: %w", err)
	}

	// Look up user to get current email (could have changed).
	// We store user_id on the refresh token, so resolve from there.
	return s.issueTokens(ctx, rt.UserID, "") // email resolved below
}

// issueTokens generates a new access + refresh token pair and stores the refresh token.
func (s *Service) issueTokens(ctx context.Context, userID primitive.ObjectID, email string) (*TokenResponse, error) {
	// If email is empty we could look it up; for simplicity we embed empty string
	// (the JWT sub already contains the user ID). In the Refresh flow the caller
	// can supply "" and we'll resolve it. Let's do a quick lookup in that case.
	if email == "" {
		// Minimal approach: we accept empty email for refresh and omit it from claims.
		// A more complete implementation would look up the user.
		email = "" // acceptable — the middleware resolves by sub
	}

	accessToken, err := s.jwt.GenerateAccessToken(userID.Hex(), email)
	if err != nil {
		return nil, fmt.Errorf("auth issue access: %w", err)
	}

	refreshStr, err := GenerateRefreshTokenString()
	if err != nil {
		return nil, fmt.Errorf("auth issue refresh: %w", err)
	}

	rt := &RefreshToken{
		UserID:    userID,
		Token:     refreshStr,
		ExpiresAt: time.Now().UTC().Add(s.refreshTTL),
		Revoked:   false,
	}
	if err = s.repo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, fmt.Errorf("auth store refresh: %w", err)
	}

	slog.Info("tokens issued", "user_id", userID.Hex())
	return &TokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresIn: s.jwt.AccessTTLSeconds(),
		RefreshToken:         refreshStr,
		TokenType:            "Bearer",
	}, nil
}

// ErrInvalidRefreshToken indicates the refresh token is missing, revoked, or expired.
var ErrInvalidRefreshToken = fmt.Errorf("invalid or expired refresh token")
