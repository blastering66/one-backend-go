// Package validate provides a shared validator instance with custom rules.
package validate

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator with custom registrations.
type Validator struct {
	v *validator.Validate
}

// New creates a Validator with custom validations registered.
func New() *Validator {
	v := validator.New()

	// name: letters and spaces, 2-50 chars
	_ = v.RegisterValidation("name", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if len(val) < 2 || len(val) > 50 {
			return false
		}
		return regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(val)
	})

	// strongpass: min 8 chars, at least 1 letter and 1 number
	_ = v.RegisterValidation("strongpass", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if len(val) < 8 {
			return false
		}
		hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(val)
		hasDigit := regexp.MustCompile(`[0-9]`).MatchString(val)
		return hasLetter && hasDigit
	})

	return &Validator{v: v}
}

// Struct validates a struct and returns a map of field-level error messages.
func (va *Validator) Struct(s interface{}) map[string]string {
	err := va.v.Struct(s)
	if err == nil {
		return nil
	}

	errs := make(map[string]string)
	for _, fe := range err.(validator.ValidationErrors) {
		field := strings.ToLower(fe.Field())
		switch fe.Tag() {
		case "required":
			errs[field] = field + " is required"
		case "email":
			errs[field] = "invalid email address"
		case "name":
			errs[field] = "must be 2-50 characters, letters and spaces only"
		case "strongpass":
			errs[field] = "min 8 chars with at least 1 letter and 1 number"
		case "min":
			errs[field] = field + " must be at least " + fe.Param() + " characters"
		case "max":
			errs[field] = field + " must be at most " + fe.Param() + " characters"
		case "gte":
			errs[field] = field + " must be >= " + fe.Param()
		case "lte":
			errs[field] = field + " must be <= " + fe.Param()
		default:
			errs[field] = field + " is invalid"
		}
	}
	return errs
}

// Engine returns the underlying validator for direct use if needed.
func (va *Validator) Engine() *validator.Validate {
	return va.v
}
