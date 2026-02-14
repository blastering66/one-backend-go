package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles creation and validation of JWT access tokens.
type JWTManager struct {
	secret    []byte
	accessTTL time.Duration
}

// NewJWTManager creates a new JWTManager.
func NewJWTManager(secret string, accessTTL time.Duration) *JWTManager {
	return &JWTManager{
		secret:    []byte(secret),
		accessTTL: accessTTL,
	}
}

// Claims are the custom JWT claims embedded in access tokens.
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed JWT for the given user.
func (j *JWTManager) GenerateAccessToken(userID, email string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateAccessToken parses and validates a JWT string, returning the claims.
func (j *JWTManager) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// Ensure only HS256 is accepted.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// AccessTTLSeconds returns the access token TTL in whole seconds.
func (j *JWTManager) AccessTTLSeconds() int {
	return int(j.accessTTL.Seconds())
}

// GenerateRefreshTokenString creates a cryptographically random base64 token (32 bytes).
func GenerateRefreshTokenString() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
