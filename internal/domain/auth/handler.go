package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/one-backend-go/internal/domain/user"
	"github.com/one-backend-go/internal/pkg/resp"
	"github.com/one-backend-go/internal/pkg/validate"
)

// Handler holds HTTP handlers for auth endpoints.
type Handler struct {
	svc      *Service
	validate *validate.Validator
}

// NewHandler creates a new auth Handler.
func NewHandler(svc *Service, v *validate.Validator) *Handler {
	return &Handler{svc: svc, validate: v}
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", nil)
		return
	}

	if errs := h.validate.Struct(req); errs != nil {
		resp.ValidationError(c, errs)
		return
	}

	tokens, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			resp.Unauthorized(c, "invalid email or password")
			return
		}
		resp.InternalError(c)
		return
	}

	resp.Success(c, http.StatusOK, tokens)
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", nil)
		return
	}

	if errs := h.validate.Struct(req); errs != nil {
		resp.ValidationError(c, errs)
		return
	}

	tokens, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			resp.Unauthorized(c, "invalid or expired refresh token")
			return
		}
		resp.InternalError(c)
		return
	}

	resp.Success(c, http.StatusOK, tokens)
}
