# Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install ca-certificates (needed for HTTPS)
RUN apk add --no-cache ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Generate sqlc code
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN sqlc generate

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main /main
COPY --from=builder /app/db/schema.sql /db/schema.sql

CMD ["/main"]