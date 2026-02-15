You are an expert Go backend engineer. Generate a complete, production‑ready Go project that exposes a small REST API for a food service app with features:
1) user registration
2) user login (JWT)
3) list products (food items) with pagination + category filter + text search.

### Tech stack & versions
- Go 1.22+.
- HTTP framework: gin-gonic/gin.
- Database: MongoDB using the official mongo-go-driver.
- Password hashing: golang.org/x/crypto/bcrypt.
- Auth: JWT (github.com/golang-jwt/jwt/v5).
- Validation: github.com/go-playground/validator/v10.
- Environment config: github.com/joho/godotenv.
- Logging: use Gin’s logger + a simple structured logger (log/slog) where appropriate.
- Dependency management: Go modules.
- Containerization: Docker + docker-compose (MongoDB + API).
- Testing: Go’s testing package + table-driven tests.

### Functional requirements
- **Register** (POST /api/v1/auth/register):
  - Request: { name, email, password }
  - Validate:
    - name: 2–50 chars (letters, spaces)
    - email: valid and lowercased; must be unique
    - password: min 8 chars with at least 1 letter and 1 number
  - Hash password with bcrypt (cost 12).
  - Create user document; store: _id, name, email, password_hash, created_at.
  - Response: 201 with { id, name, email, created_at } (never return password).

- **Login** (POST /api/v1/auth/login):
  - Request: { email, password }
  - Verify credentials; on success issue a **JWT access token** (HS256) with claims: sub(user_id), email, iat, exp (15m). Also return a **refresh token** (random 32+ bytes base64) that is stored server-side in Mongo (collection refresh_tokens with user_id, token, expires_at, revoked=false).
  - Response: 200 with { access_token, access_token_expires_in, refresh_token, token_type:"Bearer" }.
  - On failure: 401 with standardized error body.

- **Refresh token** (POST /api/v1/auth/refresh):
  - Request: { refresh_token }
  - Validate server-side, rotate (invalidate old, create new), and issue new access token.
  - Response: same shape as login.

- **List products** (GET /api/v1/products):
  - Public endpoint (no auth required for listing).
  - Support query params: page (default 1), page_size (default 10, max 50), q (text search on name/description), category (exact match), sort (name|price|created_at, asc/desc).
  - Response: 200 with { items: [...], page, page_size, total, total_pages }.
  - Product fields: _id, name, description, price (decimal as int cents), category, image_url (optional), is_available (bool), created_at, updated_at.
  - Create indexes:
    - products: { name:text, description:text }, category (asc), created_at (desc)
    - users: unique index on email
    - refresh_tokens: user_id + token unique; expires_at TTL index

- **(Optional) Create product** (POST /api/v1/products):
  - Protected (Bearer token).
  - Admin-only via simple role field on user (role: "admin"|"user"); default "user".
  - Validate name (2–80), description (<=1000), price >= 0, category non-empty.

### Non-functional requirements
- Clean architecture with layers: /internal/{config, db, domain, user, product, auth, http} and /cmd/server/main.go.
- Use context.Context for all DB ops and HTTP handlers.
- Timeout: HTTP server read/write timeouts; Mongo client with context timeouts.
- Secure defaults:
  - Read JWT secret & Mongo URI from environment variables.
  - Never log secrets or password hashes.
  - Use prepared statements / parameterization (mongo driver) and validate all inputs.
- Graceful shutdown with context cancellation.
- Return consistent error JSON: { "error": { "code": "string", "message": "string", "details": any } }.
- Add CORS middleware (allow configurable origins).
- Pagination & sorting parameters must be validated with sensible defaults and bounds.

### Environment & config
- .env example:
  - PORT=8080
  - MONGODB_URI=mongodb://mongo:27017
  - MONGODB_DB=foodsvc
  - JWT_SECRET=change_me_dev_only
  - ACCESS_TOKEN_TTL=15m
  - REFRESH_TOKEN_TTL=720h
  - CORS_ALLOWED_ORIGINS=*
- Config loader that supports .env in dev and pure env in prod.

### Data models (MongoDB)
- users:
  - _id (ObjectID)
  - name (string)
  - email (string, unique, lowercase)
  - password_hash (string)
  - role (string, enum: "user"|"admin", default "user")
  - created_at (time)
  - updated_at (time)
- products:
  - _id, name, description, price_cents (int64), category, image_url (string, optional),
    is_available (bool, default true), created_at, updated_at
- refresh_tokens:
  - _id, user_id (ObjectID), token (string), expires_at (time, TTL index), revoked (bool, default false), created_at

### API design details
- Base path: /api/v1
- Response content-type: application/json; charset=utf-8
- Auth:
  - Access token in Authorization: Bearer <token>
  - Middleware verifies signature, expiry, and injects user context
- Errors:
  - 400 for validation errors (include field-level messages)
  - 401 for auth failures; 403 for role violations
  - 404 for not found
  - 409 on email conflict during register
  - 500 for unhandled server errors

### Project structure (generate files)
- cmd/
  - server/main.go  (wire everything: config, logger, db, router, routes, graceful shutdown)
- internal/
  - config/config.go
  - db/mongo.go        (connect, indexes creation)
  - http/router.go     (gin engine, middleware, CORS)
  - http/middleware.go (auth, recover, request-id)
  - domain/
    - user/{model.go, repository.go, service.go, handler.go, dtos.go}
    - auth/{jwt.go, service.go, handler.go}
    - product/{model.go, repository.go, service.go, handler.go, dtos.go}
  - pkg/validate/validate.go (validator wrapper)
  - pkg/resp/resp.go         (success/error helpers)
  - pkg/pagination/pagination.go
- test/
  - e2e/ (http tests using httptest)
  - unit/ (service & repository tests with fakes)

### Endpoints summary (implement all)
- POST   /api/v1/auth/register
- POST   /api/v1/auth/login
- POST   /api/v1/auth/refresh
- GET    /api/v1/products
- (Optional admin)
  - POST   /api/v1/products
  - PUT    /api/v1/products/:id
  - DELETE /api/v1/products/:id

### Indexes & seed
- Create the Mongo indexes programmatically at startup (idempotent).
- Provide a small seed script or function to insert 10 sample food products.

### Tooling & DX
- Provide Makefile with:
  - make run
  - make test
  - make lint (go vet + staticcheck if available)
  - make seed
- Provide Dockerfile for the API and docker-compose.yml with services:
  - api (depends_on mongo)
  - mongo (expose 27017; mount a named volume)
- Provide a Postman (or HTTPie .http) collection for all endpoints with example requests.
- Provide example curl commands in README.
- Include .golangci.yml only if you also add the tool; otherwise use go vet.

### Tests
- Unit tests for:
  - password hashing/verification
  - JWT generation/validation
  - product listing service (pagination, filters)
- E2E tests using httptest for register/login and products listing.
- Aim for >70% coverage on services and handlers.

### Output expectations
- Generate all source files with idiomatic comments.
- Include a top-level README.md with:
  - project overview
  - prerequisites
  - how to run (with and without Docker)
  - environment variables
  - API docs (request/response samples)
  - testing instructions
- After code, print a short “Quickstart” section:
  - commands: `go mod tidy`, `cp .env.example .env`, `docker-compose up -d`, `go run ./cmd/server`
  - example curl for register, login, and list products.

### Conventions & notes
- Use UTC times in ISO 8601.
- Ensure secure JWT parsing (validate alg HS256 only).
- Lowercase & trim emails on input.
- Return consistent error shapes everywhere.
- Avoid global variables; pass dependencies via constructors.
- Use context with timeouts for DB calls (e.g., 5s).
- Keep handlers thin; put logic in services; repositories isolate Mongo queries.

Now, generate the full project code, Docker assets, seed data, tests, and README in one response. If something is ambiguous, make a reasonable default and document it in README.
``