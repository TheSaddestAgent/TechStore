# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application including the product-service directory
COPY . .

# Compile the server and migration binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/server \
    ./product-service/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/migrate \
    ./product-service/migrate/main.go

# Stage 2: Runtime
FROM alpine:latest AS runtime

# Install necessary certificates
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy the compiled binaries
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/bin/migrate /app/migrate

# Copy the necessary migrations
COPY --from=builder /app/product-service/migrations /app/migrations

# If you have .env file in the root context
COPY --from=builder /app/.env /app/.env

# Change ownership of the app directory for security reasons
RUN chown -R appuser:appuser /app
USER appuser

# Stage 3: Server
FROM runtime AS server
EXPOSE 8082
ENTRYPOINT ["/app/server"]

# Stage 4: Migrate
FROM runtime AS migrate
ENTRYPOINT ["/app/migrate"]
