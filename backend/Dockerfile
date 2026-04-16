# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files first (Docker layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /artshop-api ./cmd/api/main.go

# Runtime stage - minimal image
FROM gcr.io/distroless/static-debian12

COPY --from=builder /artshop-api /artshop-api

EXPOSE 8080

ENTRYPOINT ["/artshop-api"]
