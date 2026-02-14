package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/one-backend-go/internal/pkg/resp"
	"github.com/one-backend-go/internal/pkg/validate"
)

// Handler holds HTTP handlers for user-related endpoints.
type Handler struct {
	svc      *Service
	validate *validate.Validator
}

// NewHandler creates a new user Handler.
func NewHandler(svc *Service, v *validate.Validator) *Handler {
	return &Handler{svc: svc, validate: v}
}

// Register handles POST /api/v1/auth/register.
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", nil)
		return
	}

	if errs := h.validate.Struct(req); errs != nil {
		resp.ValidationError(c, errs)
		return
	}

	u, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			resp.Conflict(c, "a user with this email already exists")
			return
		}
		resp.InternalError(c)
		return
	}

	resp.Success(c, http.StatusCreated, u.ToResponse())
}
