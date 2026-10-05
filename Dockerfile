# syntax=docker/dockerfile:1

# Step 1: Build static Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build static non-CGO binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -extldflags '-static'" \
    -trimpath \
    -o /app/fake-api-server ./cmd/server

# Step 2: Minimal non-root runtime container
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=builder /app/fake-api-server /fake-api-server

# Default environment variables
ENV ADDR=":8080" \
    TARPIT_DELAY_MS="100ms" \
    MAX_CONCURRENT="1000" \
    LOG_LEVEL="info" \
    ENABLE_GZIP_BOMB="false"

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/fake-api-server"]
