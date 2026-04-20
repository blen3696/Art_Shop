.PHONY: run dev build start prod-local test test-coverage clean tidy lint docker-build docker-run docker-prod migrate health

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

migrate:
	@echo "Running migrations..."
	psql "$(DATABASE_URL)" -f migrations/001_initial_schema.sql
	@echo "Migrations complete."
