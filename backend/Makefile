.PHONY: run build test clean migrate docker-build docker-run

# Variables
BINARY_NAME=artshop-api
MAIN_PATH=./cmd/api

# Run the server in development mode
run:
	go run $(MAIN_PATH)/main.go

# Build the binary
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$(BINARY_NAME) $(MAIN_PATH)/main.go

# Run tests
test:
	go test -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Tidy dependencies
tidy:
	go mod tidy

# Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

# Docker
docker-build:
	docker build -t artshop-api .

docker-run:
	docker run --env-file .env -p 8080:8080 artshop-api

# Run migration against Supabase (requires psql and DATABASE_URL in .env)
migrate:
	@echo "Running migrations..."
	psql "$(DATABASE_URL)" -f migrations/001_initial_schema.sql
	@echo "Migrations complete."
