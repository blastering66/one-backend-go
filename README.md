# One Backend Go — Food Service REST API

A production-ready Go REST API for a food service application. Features user registration, JWT authentication with refresh token rotation, and a paginated product catalog with text search and filtering.

## Tech Stack

| Component | Library |
|---|---|
| HTTP framework | [gin-gonic/gin](https://github.com/gin-gonic/gin) |
| Database | MongoDB via [mongo-go-driver](https://github.com/mongodb/mongo-go-driver) |
| Auth | JWT ([golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)) + bcrypt |
| Validation | [go-playground/validator/v10](https://github.com/go-playground/validator) |
| Config | [joho/godotenv](https://github.com/joho/godotenv) + env vars |
| Logging | `log/slog` (structured JSON) |
| Containerization | Docker + Docker Compose |

## Project Structure

```
cmd/
  server/main.go          # Application entry point
  seed/main.go            # Database seed script
internal/
  config/config.go        # Environment configuration loader
  db/mongo.go             # MongoDB connection + index management
  http/
    router.go             # Gin engine, routes, CORS
    middleware.go          # Auth, recovery, request-ID, logger
  domain/
    user/                 # User model, repository, service, handler, DTOs
    auth/                 # JWT manager, refresh tokens, auth service & handler
    product/              # Product model, repository, service, handler, DTOs
  pkg/
    validate/validate.go  # Custom validator wrapper
    resp/resp.go          # Standardized JSON response helpers
    pagination/pagination.go
test/
  unit/                   # Table-driven unit tests
  e2e/                    # HTTP integration tests (httptest)
```

## Prerequisites

- **Go 1.22+**
- **MongoDB 6+** (local or Docker)
- **Docker & Docker Compose** (optional, for containerized dev)

## Environment Variables

Copy the example file and edit as needed:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `MONGODB_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGODB_DB` | `foodsvc` | Database name |
| `JWT_SECRET` | _(required)_ | HMAC-SHA256 signing secret |
| `ACCESS_TOKEN_TTL` | `15m` | Access token lifetime |
| `REFRESH_TOKEN_TTL` | `720h` | Refresh token lifetime (30 days) |
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated allowed origins |

## Running

### With Docker (recommended)

```bash
docker compose up -d --build
```

The API starts at `http://localhost:8080`. MongoDB is exposed on port `27017`.

### Without Docker

Ensure MongoDB is running locally, then:

```bash
cp .env.example .env
# Edit .env: set MONGODB_URI=mongodb://localhost:27017
go mod tidy
go run ./cmd/server
```

### Seed sample data

```bash
go run ./cmd/seed
# or
make seed
```

## API Documentation

Base URL: `http://localhost:8080/api/v1`

All responses use `Content-Type: application/json; charset=utf-8`.

### Error format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "request validation failed",
    "details": {
      "email": "invalid email address"
    }
  }
}
```

---

### POST /api/v1/auth/register

Register a new user.

**Request:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "secret123"
}
```

**Response (201):**
```json
{
  "id": "65f1a2b3c4d5e6f7a8b9c0d1",
  "name": "John Doe",
  "email": "john@example.com",
  "role": "user",
  "created_at": "2026-02-15T10:30:00Z"
}
```

**Errors:** 400 (validation), 409 (email exists)

---

### POST /api/v1/auth/login

Authenticate and receive tokens.

**Request:**
```json
{
  "email": "john@example.com",
  "password": "secret123"
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "access_token_expires_in": 900,
  "refresh_token": "dGhpcyBpcyBhIHJlZnJl...",
  "token_type": "Bearer"
}
```

**Errors:** 401 (invalid credentials)

---

### POST /api/v1/auth/refresh

Rotate refresh token and get new access token.

**Request:**
```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJl..."
}
```

**Response (200):** Same shape as login response.

**Errors:** 401 (invalid/expired refresh token)

---

### GET /api/v1/products

List products with pagination, filtering, and search. **Public endpoint — no auth required.**

**Query parameters:**

| Param | Default | Description |
|---|---|---|
| `page` | `1` | Page number |
| `page_size` | `10` | Items per page (max 50) |
| `q` | | Text search on name + description |
| `category` | | Exact category match |
| `sort` | `created_at,desc` | Sort field + direction (`name`, `price`, `created_at`) |

**Response (200):**
```json
{
  "items": [
    {
      "id": "65f1a2b3c4d5e6f7a8b9c0d1",
      "name": "Margherita Pizza",
      "description": "Traditional pizza with fresh mozzarella",
      "price_cents": 1299,
      "category": "pizza",
      "image_url": "https://example.com/img/margherita.jpg",
      "is_available": true,
      "created_at": "2026-02-15T10:00:00Z",
      "updated_at": "2026-02-15T10:00:00Z"
    }
  ],
  "page": 1,
  "page_size": 10,
  "total": 10,
  "total_pages": 1
}
```

---

### POST /api/v1/products _(admin only)_

Create a product. Requires `Authorization: Bearer <token>` from an admin user.

**Request:**
```json
{
  "name": "Pepperoni Pizza",
  "description": "Classic pepperoni pizza",
  "price_cents": 1499,
  "category": "pizza",
  "image_url": "https://example.com/img/pepperoni.jpg",
  "is_available": true
}
```

**Response (201):** Product object.

---

### PUT /api/v1/products/:id _(admin only)_

Update a product (partial update).

### DELETE /api/v1/products/:id _(admin only)_

Delete a product. Returns `{"message": "product deleted"}`.

---

## Example curl Commands

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"secret123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'

# Refresh token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'

# List products
curl http://localhost:8080/api/v1/products

# List with filters
curl "http://localhost:8080/api/v1/products?category=pizza&page=1&page_size=5"

# Text search
curl "http://localhost:8080/api/v1/products?q=chicken"

# Sorted by price ascending
curl "http://localhost:8080/api/v1/products?sort=price,asc"

# Create product (admin)
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{"name":"New Item","description":"Desc","price_cents":999,"category":"burgers"}'
```

## Testing

### Unit tests (no external dependencies)

```bash
go test ./test/unit/... -v -race
```

### E2E tests (requires MongoDB)

```bash
MONGODB_URI=mongodb://localhost:27017 JWT_SECRET=testsecret go test ./test/e2e/... -v
```

### All tests with coverage

```bash
go test ./... -v -race -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Using Make

```bash
make test          # Run all tests
make test-cover    # Tests + coverage report
make lint          # go vet + staticcheck
```

## Quickstart

```bash
# 1. Install dependencies
go mod tidy

# 2. Set up environment
cp .env.example .env

# 3. Start with Docker
docker compose up -d

# 4. Seed sample data
go run ./cmd/seed

# 5. Or run locally (with MongoDB running)
go run ./cmd/server
```

## Design Decisions

- **Clean architecture**: Handlers → Services → Repositories. No global state; all dependencies injected via constructors.
- **Refresh token rotation**: Each use invalidates the old token and issues a new pair, preventing replay attacks.
- **Bcrypt cost 12**: Good balance of security and performance for auth workloads.
- **TTL index on refresh_tokens**: MongoDB automatically removes expired tokens.
- **Consistent error envelope**: Every error response follows `{ error: { code, message, details } }`.
- **UTC timestamps**: All times are stored and returned in ISO 8601 UTC format.
- **Admin role**: Simple role-based access control via a `role` field on the user document. Default is `"user"`. Set to `"admin"` directly in MongoDB for admin access.

## License

See [LICENSE](LICENSE).
