package unit

import (
	"testing"

	"github.com/one-backend-go/internal/pkg/validate"
)

func TestValidatorName(t *testing.T) {
	v := validate.New()

	type nameInput struct {
		Name string `validate:"required,name"`
	}

	tests := []struct {
		name    string
		input   nameInput
		wantErr bool
	}{
		{"valid name", nameInput{Name: "John Doe"}, false},
		{"too short", nameInput{Name: "J"}, true},
		{"with numbers", nameInput{Name: "John123"}, true},
		{"empty", nameInput{Name: ""}, true},
		{"exactly 2 chars", nameInput{Name: "Jo"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Struct(tt.input)
			if (errs != nil) != tt.wantErr {
				t.Errorf("Struct() errors = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestValidatorStrongPass(t *testing.T) {
	v := validate.New()

	type passInput struct {
		Password string `validate:"required,strongpass"`
	}

	tests := []struct {
		name    string
		input   passInput
		wantErr bool
	}{
		{"valid", passInput{Password: "abcdef12"}, false},
		{"no digits", passInput{Password: "abcdefgh"}, true},
		{"no letters", passInput{Password: "12345678"}, true},
		{"too short", passInput{Password: "ab1"}, true},
		{"empty", passInput{Password: ""}, true},
		{"mixed case with digits", passInput{Password: "MyPass123"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Struct(tt.input)
			if (errs != nil) != tt.wantErr {
				t.Errorf("Struct() errors = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestValidatorEmail(t *testing.T) {
	v := validate.New()

	type emailInput struct {
		Email string `validate:"required,email"`
	}

	tests := []struct {
		name    string
		input   emailInput
		wantErr bool
	}{
		{"valid", emailInput{Email: "test@example.com"}, false},
		{"no @", emailInput{Email: "testexample.com"}, true},
		{"empty", emailInput{Email: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Struct(tt.input)
			if (errs != nil) != tt.wantErr {
				t.Errorf("Struct() errors = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}
