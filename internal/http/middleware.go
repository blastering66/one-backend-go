// Package http provides HTTP middleware for the Gin engine.
package http

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/one-backend-go/internal/domain/auth"
	"github.com/one-backend-go/internal/domain/user"
	"github.com/one-backend-go/internal/pkg/resp"
)

// ── Context keys ───────────────────────────────────────────────────────────────

const (
	// ContextKeyUserID is the gin context key storing the authenticated user's ID.
	ContextKeyUserID = "user_id"
	// ContextKeyEmail is the gin context key storing the authenticated user's email.
	ContextKeyEmail = "user_email"
	// ContextKeyRole is the gin context key storing the authenticated user's role.
	ContextKeyRole = "user_role"
)

// ── Request-ID middleware ──────────────────────────────────────────────────────

// RequestID injects a unique request ID into each request/response.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// ── Structured logger middleware ───────────────────────────────────────────────

// StructuredLogger logs each request with slog.
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"request_id", c.GetString("request_id"),
		)
	}
}

// ── Recovery middleware ────────────────────────────────────────────────────────

// Recovery returns a middleware that recovers from panics and returns 500.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", "error", r)
				resp.InternalError(c)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ── Auth middleware ────────────────────────────────────────────────────────────

// AuthRequired returns middleware that validates a Bearer JWT token.
func AuthRequired(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			resp.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			resp.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtMgr.ValidateAccessToken(parts[1])
		if err != nil {
			resp.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.Subject)
		c.Set(ContextKeyEmail, claims.Email)
		c.Next()
	}
}

// AdminRequired ensures the authenticated user has the admin role.
// Must be placed AFTER AuthRequired in the middleware chain.
func AdminRequired(userRepo *user.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get(ContextKeyUserID)
		if !exists {
			resp.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		uid, err := primitive.ObjectIDFromHex(userIDStr.(string))
		if err != nil {
			resp.Unauthorized(c, "invalid user id in token")
			c.Abort()
			return
		}

		u, err := userRepo.FindByID(c.Request.Context(), uid)
		if err != nil || u == nil {
			resp.Unauthorized(c, "user not found")
			c.Abort()
			return
		}

		if u.Role != user.RoleAdmin {
			resp.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		c.Set(ContextKeyRole, u.Role)
		c.Next()
	}
}
