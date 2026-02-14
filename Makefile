.PHONY: run test lint seed build clean docker-up docker-down

# ── Run ─────────────────────────────────────────────────────────────────────
run:
	go run ./cmd/server

# ── Build ───────────────────────────────────────────────────────────────────
build:
	go build -o bin/server ./cmd/server

# ── Seed ────────────────────────────────────────────────────────────────────
seed:
	go run ./cmd/seed

# ── Test ────────────────────────────────────────────────────────────────────
test:
	go test ./... -v -count=1 -race

test-cover:
	go test ./... -v -count=1 -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# ── Lint ────────────────────────────────────────────────────────────────────
lint:
	go vet ./...
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

# ── Docker ──────────────────────────────────────────────────────────────────
docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

# ── Clean ───────────────────────────────────────────────────────────────────
clean:
	rm -rf bin/ coverage.* *.out
