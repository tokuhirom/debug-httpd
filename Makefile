.PHONY: test build run docker-build docker-run clean

# Default target
all: test build

# Run tests
test:
	go test -v ./...

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