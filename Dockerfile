# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files if they exist
COPY go.* ./

# Copy source code
COPY main.go .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o debug-httpd main.go

# Final stage
FROM scratch

# Copy the binary from builder
COPY --from=builder /build/debug-httpd /debug-httpd

# Expose the default port
EXPOSE 9876

# Run the server with default port (can be overridden by CMD)
ENTRYPOINT ["/debug-httpd"]
CMD ["9876"]