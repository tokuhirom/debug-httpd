.PHONY: test build run docker-build docker-run clean integration-test

# Default target
all: test build

# Run unit tests
test:
	go test -v ./...

# Run integration tests (requires Docker)
integration-test:
	go test -v -tags=integration -timeout=5m ./...

# Run all tests
test-all: test integration-test

# Build binary
build:
	go build -ldflags="-w -s" -o debug-httpd main.go

# Run locally
run: build
	./debug-httpd

# Build Docker image
docker-build:
	docker build -t debug-httpd:latest .

# Run Docker container
docker-run: docker-build
	docker run -p 9876:9876 debug-httpd:latest

# Clean up
clean:
	rm -f debug-httpd
	go clean -cache
	docker rmi debug-httpd:integration-test 2>/dev/null || true