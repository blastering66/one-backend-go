// Package resp provides consistent JSON response helpers for the API.
package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorBody is the standard error envelope returned by the API.
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the code, message, and optional details of an error.
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Success sends a JSON success response with the given status code and data.
func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// Fail sends a JSON error response with the given status code and error detail.
func Fail(c *gin.Context, status int, code, message string, details interface{}) {
	c.JSON(status, ErrorBody{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// ValidationError returns a 400 response with field-level validation messages.
func ValidationError(c *gin.Context, fields map[string]string) {
	Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "request validation failed", fields)
}

// Unauthorized returns a 401 error response.
func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

// Forbidden returns a 403 error response.
func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

// NotFound returns a 404 error response.
func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, "NOT_FOUND", message, nil)
}

// Conflict returns a 409 error response.
func Conflict(c *gin.Context, message string) {
	Fail(c, http.StatusConflict, "CONFLICT", message, nil)
}

// InternalError returns a 500 error response.
func InternalError(c *gin.Context) {
	Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred", nil)
}
