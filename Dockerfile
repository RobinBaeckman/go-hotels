# syntax=docker/dockerfile:1.6

# ---------------------
# Stage 1: Builder
# ---------------------
FROM golang:1.24-alpine AS builder

# Enable Go modules and configure build cache
ENV CGO_ENABLED=0 GO111MODULE=on

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Cache Go modules separately to speed up builds
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the source code
COPY . .

# Build the application binary with optimizations
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -o /go-hotels ./cmd/rest

# ---------------------
# Stage 2: Runtime
# ---------------------
FROM gcr.io/distroless/static:nonroot
# Alternatively, use: FROM alpine:3.20 if you need a shell in the image

WORKDIR /app

# Copy binary from builder
COPY --from=builder /go-hotels /app/go-hotels

# Expose API port
EXPOSE 8080

# Add a healthcheck for Docker / K8s
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
  CMD wget -qO- http://localhost:8080/health || exit 1

# Run as non-root user for security
USER 65532:65532

ENTRYPOINT ["/app/go-hotels"]
