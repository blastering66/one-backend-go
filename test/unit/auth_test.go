package unit

import (
	"testing"
	"time"

	"github.com/one-backend-go/internal/domain/auth"
	"github.com/one-backend-go/internal/domain/user"
)

// ── Password hashing tests ─────────────────────────────────────────────────

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "mySecret123", false},
		{"short password", "ab1", false}, // bcrypt itself doesn't enforce length
		{"empty password", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := user.HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && hash == "" {
				t.Fatal("HashPassword() returned empty hash")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "securePass1"
	hash, err := user.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	tests := []struct {
		name   string
		hash   string
		plain  string
		wantOK bool
	}{
		{"correct password", hash, password, true},
		{"wrong password", hash, "wrongPass1", false},
		{"empty password", hash, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := user.CheckPassword(tt.hash, tt.plain); got != tt.wantOK {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.wantOK)
			}
		})
	}
}

// ── JWT tests ──────────────────────────────────────────────────────────────

func TestJWTGenerateAndValidate(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret-key-12345", 15*time.Minute)

	t.Run("valid token round-trip", func(t *testing.T) {
		token, err := mgr.GenerateAccessToken("user123", "user@example.com")
		if err != nil {
			t.Fatalf("GenerateAccessToken() error: %v", err)
		}
		if token == "" {
			t.Fatal("GenerateAccessToken() returned empty token")
		}

		claims, err := mgr.ValidateAccessToken(token)
		if err != nil {
			t.Fatalf("ValidateAccessToken() error: %v", err)
		}
		if claims.Subject != "user123" {
			t.Errorf("Subject = %q, want %q", claims.Subject, "user123")
		}
		if claims.Email != "user@example.com" {
			t.Errorf("Email = %q, want %q", claims.Email, "user@example.com")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		mgrExpired := auth.NewJWTManager("test-secret", -1*time.Second)
		token, err := mgrExpired.GenerateAccessToken("user123", "user@example.com")
		if err != nil {
			t.Fatalf("GenerateAccessToken() error: %v", err)
		}

		_, err = mgrExpired.ValidateAccessToken(token)
		if err == nil {
			t.Fatal("ValidateAccessToken() expected error for expired token")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		mgrA := auth.NewJWTManager("secret-A", 15*time.Minute)
		mgrB := auth.NewJWTManager("secret-B", 15*time.Minute)

		token, _ := mgrA.GenerateAccessToken("user1", "a@b.com")
		_, err := mgrB.ValidateAccessToken(token)
		if err == nil {
			t.Fatal("ValidateAccessToken() expected error for wrong secret")
		}
	})

	t.Run("invalid token string", func(t *testing.T) {
		_, err := mgr.ValidateAccessToken("not.a.valid.token")
		if err == nil {
			t.Fatal("ValidateAccessToken() expected error for garbage token")
		}
	})
}

func TestGenerateRefreshTokenString(t *testing.T) {
	token, err := auth.GenerateRefreshTokenString()
	if err != nil {
		t.Fatalf("GenerateRefreshTokenString() error: %v", err)
	}
	if len(token) < 32 {
		t.Errorf("refresh token too short: len=%d", len(token))
	}

	// Ensure uniqueness
	token2, _ := auth.GenerateRefreshTokenString()
	if token == token2 {
		t.Error("two generated refresh tokens should not be identical")
	}
}

func TestAccessTTLSeconds(t *testing.T) {
	mgr := auth.NewJWTManager("secret", 15*time.Minute)
	if got := mgr.AccessTTLSeconds(); got != 900 {
		t.Errorf("AccessTTLSeconds() = %d, want 900", got)
	}
}
