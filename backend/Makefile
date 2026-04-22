.PHONY: run dev build start prod-local test test-coverage clean tidy lint docker-build docker-run docker-prod migrate-up migrate-down migrate-force migrate-create migrate-version seed health

# Variables
BINARY_NAME=artshop-api
MAIN_PATH=./cmd/api
PORT?=8080

# --- Local development ---

# Run the server with go run (loads .env automatically)
run:
	go run $(MAIN_PATH)/main.go

# Alias for run
dev: run

# --- Build & run the production binary locally ---

# Build the binary using the SAME flags the Dockerfile uses (parity)
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$(BINARY_NAME) $(MAIN_PATH)/main.go

# Run the compiled binary the way Render will (binds to $PORT)
start:
	./bin/$(BINARY_NAME)

# Full prod simulation: build + start
prod-local: build start

# --- Tests ---

test:
	go test -v -race -cover ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# --- Housekeeping ---

clean:
	rm -rf bin/ coverage.out coverage.html

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

# --- Docker ---

# Build the image
docker-build:
	docker build -t artshop-api .

# Run with default port 8080 (dev-style)
docker-run:
	docker run --rm --env-file .env -p 8080:8080 artshop-api

# Run the image EXACTLY the way Render will run it (PORT=10000)
docker-prod:
	docker run --rm --env-file .env -e PORT=10000 -p 10000:10000 artshop-api

# --- Operational checks ---

# Hit /health to confirm the server is up. Override port: make health PORT=10000
health:
	curl -fsS http://localhost:$(PORT)/health && echo

# --- Database ---
#
# Schema migrations are also run automatically at server startup by the
# internal/migrate package. These targets are for MANUAL control — useful for
# rollbacks, creating new migration files, and checking version state.
#
# Requires the golang-migrate CLI for manual operations:
#   scoop install migrate             (Windows)
#   brew install golang-migrate        (macOS)
#   apt install golang-migrate         (Debian/Ubuntu)

MIGRATE_DIR=internal/migrate/schema

# Apply all pending up-migrations. Equivalent to what the server does on boot.
migrate-up:
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" up

# Roll back the MOST RECENT migration. Pass N=3 to roll back three.
migrate-down:
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" down $(if $(N),$(N),1)

# Show current migration version + whether the DB is in a dirty state.
migrate-version:
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" version

# Clear the dirty flag after fixing a failed migration manually. Pass V=<version>.
# Example: make migrate-force V=3
migrate-force:
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" force $(V)

# Create a new migration pair (up + down). Pass NAME=add_something.
# Example: make migrate-create NAME=add_user_preferences
migrate-create:
	migrate create -ext sql -dir $(MIGRATE_DIR) -seq $(NAME)

# Apply seed data files. One-shot for fresh databases; NOT idempotent.
seed:
	@echo "Applying seed data..."
	psql "$(DATABASE_URL)" -f migrations/seeds/001_core_seed.sql
	psql "$(DATABASE_URL)" -f migrations/seeds/002_extra_seed.sql
	@echo "Seeds complete."
