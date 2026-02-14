// Package e2e provides end-to-end HTTP tests using httptest.
//
// These tests require a running MongoDB instance. Set MONGODB_URI and JWT_SECRET
// environment variables, or create a .env file before running.
//
//	MONGODB_URI=mongodb://localhost:27017 JWT_SECRET=testsecret go test ./test/e2e/... -v
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/one-backend-go/internal/config"
	"github.com/one-backend-go/internal/db"
	"github.com/one-backend-go/internal/domain/auth"
	"github.com/one-backend-go/internal/domain/product"
	"github.com/one-backend-go/internal/domain/user"
	apphttp "github.com/one-backend-go/internal/http"
	"github.com/one-backend-go/internal/pkg/validate"
)

// setupRouter creates a test router backed by a real MongoDB.
// It drops the test database before each call to guarantee isolation.
func setupRouter(t *testing.T) *httptest.Server {
	t.Helper()

	// Use test-specific env if not set
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "e2e-test-secret")
	}
	if os.Getenv("MONGODB_URI") == "" {
		os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	}
	if os.Getenv("MONGODB_DB") == "" {
		os.Setenv("MONGODB_DB", "foodsvc_test")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load: %v", err)
	}
	cfg.MongoDB = "foodsvc_test" // force test db

	ctx := context.Background()
	mongoDB, err := db.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		t.Skipf("skipping e2e tests: cannot connect to MongoDB: %v", err)
	}

	// Drop test database for clean state
	if err := mongoDB.Drop(ctx); err != nil {
		t.Fatalf("drop test db: %v", err)
	}
	if err := db.EnsureIndexes(ctx, mongoDB); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}

	v := validate.New()
	userRepo := user.NewRepository(mongoDB)
	authRepo := auth.NewRepository(mongoDB)
	productRepo := product.NewRepository(mongoDB)

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.AccessTokenTTL)
	userSvc := user.NewService(userRepo)
	authSvc := auth.NewService(cfg, jwtMgr, authRepo, userSvc)
	productSvc := product.NewService(productRepo)

	userHandler := user.NewHandler(userSvc, v)
	authHandler := auth.NewHandler(authSvc, v)
	productHandler := product.NewHandler(productSvc, v)

	router := apphttp.NewRouter(cfg, jwtMgr, userRepo, userHandler, authHandler, productHandler)

	// Seed some products
	seedProducts(t, productRepo)

	ts := httptest.NewServer(router)
	t.Cleanup(func() {
		ts.Close()
		_ = db.Disconnect(ctx, mongoDB)
	})
	return ts
}

func seedProducts(t *testing.T, repo *product.Repository) {
	t.Helper()
	products := []product.Product{
		{Name: "Classic Burger", Description: "Beef burger", PriceCents: 999, Category: "burgers", IsAvailable: true},
		{Name: "Margherita Pizza", Description: "Fresh pizza", PriceCents: 1299, Category: "pizza", IsAvailable: true},
		{Name: "Caesar Salad", Description: "Romaine salad", PriceCents: 799, Category: "salads", IsAvailable: true},
	}
	if err := repo.InsertMany(context.Background(), products); err != nil {
		t.Fatalf("seed products: %v", err)
	}
}

func jsonBody(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return bytes.NewBuffer(b)
}

// ── Tests ──────────────────────────────────────────────────────────────────

func TestHealthCheck(t *testing.T) {
	ts := setupRouter(t)
	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRegister(t *testing.T) {
	ts := setupRouter(t)

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			"valid registration",
			map[string]string{"name": "Jane Doe", "email": "jane@example.com", "password": "secret123"},
			http.StatusCreated,
		},
		{
			"duplicate email",
			map[string]string{"name": "Jane Again", "email": "jane@example.com", "password": "secret456"},
			http.StatusConflict,
		},
		{
			"invalid email",
			map[string]string{"name": "Bob", "email": "not-an-email", "password": "secret123"},
			http.StatusBadRequest,
		},
		{
			"weak password",
			map[string]string{"name": "Bob Smith", "email": "bob@example.com", "password": "short"},
			http.StatusBadRequest,
		},
		{
			"name too short",
			map[string]string{"name": "A", "email": "a@example.com", "password": "secret123"},
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", jsonBody(t, tt.body))
			if err != nil {
				t.Fatalf("POST /register error: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tt.wantStatus {
				var body map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&body)
				t.Errorf("status = %d, want %d, body = %v", resp.StatusCode, tt.wantStatus, body)
			}
		})
	}
}

func TestLoginAndRefresh(t *testing.T) {
	ts := setupRouter(t)

	// Register first
	regBody := map[string]string{"name": "Test User", "email": "test@example.com", "password": "password123"}
	resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", jsonBody(t, regBody))
	if err != nil {
		t.Fatalf("register error: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register status = %d, want 201", resp.StatusCode)
	}

	// Login
	loginBody := map[string]string{"email": "test@example.com", "password": "password123"}
	resp, err = http.Post(ts.URL+"/api/v1/auth/login", "application/json", jsonBody(t, loginBody))
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d, want 200", resp.StatusCode)
	}

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	if tokenResp["access_token"] == nil || tokenResp["refresh_token"] == nil {
		t.Fatalf("missing tokens in response: %v", tokenResp)
	}
	if tokenResp["token_type"] != "Bearer" {
		t.Errorf("token_type = %v, want Bearer", tokenResp["token_type"])
	}

	// Refresh
	refreshBody := map[string]string{"refresh_token": tokenResp["refresh_token"].(string)}
	resp2, err := http.Post(ts.URL+"/api/v1/auth/refresh", "application/json", jsonBody(t, refreshBody))
	if err != nil {
		t.Fatalf("refresh error: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("refresh status = %d, want 200", resp2.StatusCode)
	}

	var refreshResp map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&refreshResp)
	if refreshResp["access_token"] == nil {
		t.Error("refresh did not return new access_token")
	}

	// Old refresh token should now be invalid
	resp3, err := http.Post(ts.URL+"/api/v1/auth/refresh", "application/json", jsonBody(t, refreshBody))
	if err != nil {
		t.Fatalf("refresh (old token) error: %v", err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusUnauthorized {
		t.Errorf("old refresh token: status = %d, want 401", resp3.StatusCode)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	ts := setupRouter(t)

	body := map[string]string{"email": "nonexistent@example.com", "password": "password123"}
	resp, err := http.Post(ts.URL+"/api/v1/auth/login", "application/json", jsonBody(t, body))
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestListProducts(t *testing.T) {
	ts := setupRouter(t)

	t.Run("default listing", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/products")
		if err != nil {
			t.Fatalf("GET /products error: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		items, ok := body["items"].([]interface{})
		if !ok {
			t.Fatalf("items is not an array: %v", body)
		}
		if len(items) != 3 {
			t.Errorf("items count = %d, want 3", len(items))
		}
		if body["total"].(float64) != 3 {
			t.Errorf("total = %v, want 3", body["total"])
		}
	})

	t.Run("category filter", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/products?category=pizza")
		if err != nil {
			t.Fatalf("GET /products?category=pizza error: %v", err)
		}
		defer resp.Body.Close()

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		if body["total"].(float64) != 1 {
			t.Errorf("total = %v, want 1", body["total"])
		}
	})

	t.Run("pagination", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/products?page=1&page_size=2")
		if err != nil {
			t.Fatalf("GET /products?page=1&page_size=2 error: %v", err)
		}
		defer resp.Body.Close()

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		items := body["items"].([]interface{})
		if len(items) != 2 {
			t.Errorf("items count = %d, want 2", len(items))
		}
		if body["total_pages"].(float64) != 2 {
			t.Errorf("total_pages = %v, want 2", body["total_pages"])
		}
	})

	t.Run("text search", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/products?q=pizza")
		if err != nil {
			t.Fatalf("GET /products?q=pizza error: %v", err)
		}
		defer resp.Body.Close()

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		total := body["total"].(float64)
		if total < 1 {
			t.Errorf("text search for 'pizza' returned %v results, want >= 1", total)
		}
	})
}
