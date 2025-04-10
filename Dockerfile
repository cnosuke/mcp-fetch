FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git make

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application for Linux architecture
RUN make build-for-linux-amd64

# Final stage
FROM gcr.io/distroless/static:nonroot

# Set metadata
LABEL org.opencontainers.image.source="https://github.com/cnosuke/mcp-fetch"
LABEL org.opencontainers.image.description="MCP server for fetch functionality"

WORKDIR /app

# Copy configuration
COPY --from=builder /app/config.yml /app/config.yml

# Copy the binary
COPY --from=builder /app/bin/mcp-fetch-linux-amd64 /app/mcp-fetch

# Use nonroot user
USER nonroot:nonroot

# Set the entrypoint
ENTRYPOINT ["/app/mcp-fetch"]

# Default command
CMD ["server", "--config", "config.yml"]