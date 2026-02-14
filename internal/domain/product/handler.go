package product

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/one-backend-go/internal/pkg/pagination"
	"github.com/one-backend-go/internal/pkg/resp"
	"github.com/one-backend-go/internal/pkg/validate"
)

// Handler holds HTTP handlers for product endpoints.
type Handler struct {
	svc      *Service
	validate *validate.Validator
}

// NewHandler creates a new product Handler.
func NewHandler(svc *Service, v *validate.Validator) *Handler {
	return &Handler{svc: svc, validate: v}
}

// List handles GET /api/v1/products.
func (h *Handler) List(c *gin.Context) {
	p := pagination.DefaultParams()

	if v := c.Query("page"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			p.Page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			p.PageSize = n
		}
	}
	if v := c.Query("sort"); v != "" {
		parts := strings.SplitN(v, ",", 2)
		p.Sort = parts[0]
		if len(parts) == 2 && (parts[1] == "asc" || parts[1] == "desc") {
			p.Order = parts[1]
		}
	}

	filter := ListFilter{
		Query:    c.Query("q"),
		Category: c.Query("category"),
	}

	result, err := h.svc.List(c.Request.Context(), filter, p)
	if err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, http.StatusOK, result)
}

// Create handles POST /api/v1/products (admin only).
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", nil)
		return
	}

	if errs := h.validate.Struct(req); errs != nil {
		resp.ValidationError(c, errs)
		return
	}

	p, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, http.StatusCreated, p.ToResponse())
}

// Update handles PUT /api/v1/products/:id (admin only).
func (h *Handler) Update(c *gin.Context) {
	idParam := c.Param("id")

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", nil)
		return
	}

	if errs := h.validate.Struct(req); errs != nil {
		resp.ValidationError(c, errs)
		return
	}

	p, err := h.svc.Update(c.Request.Context(), idParam, req)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			resp.NotFound(c, "product not found")
			return
		}
		resp.InternalError(c)
		return
	}

	resp.Success(c, http.StatusOK, p.ToResponse())
}

// Delete handles DELETE /api/v1/products/:id (admin only).
func (h *Handler) Delete(c *gin.Context) {
	idParam := c.Param("id")

	err := h.svc.Delete(c.Request.Context(), idParam)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			resp.NotFound(c, "product not found")
			return
		}
		resp.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
