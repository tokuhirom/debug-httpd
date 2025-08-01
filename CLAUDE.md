# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a lightweight debug HTTP server project written in Go. It creates a minimal Docker container (≈6.5MB) for debugging purposes. The server provides:
- Environment variables
- Host information (hostname, FQDN, IP addresses)
- Request details (headers, client info)
- Go version
- Access logging (last 100 requests)
- Health check endpoint

## Common Commands

### Build Docker Image Locally
```bash
docker build -t debug-httpd:latest .
```

### Run Container Locally
```bash
# Run on default port 9876
docker run -p 9876:9876 debug-httpd:latest

# Run on custom port using CMD
docker run -p 8080:8080 debug-httpd:latest 8080

# Run on custom port using environment variable
docker run -p 9000:9000 -e PORT=9000 debug-httpd:latest
```

### Build and Test Locally
```bash
# Run unit tests
make test

# Run integration tests (requires Docker)
make integration-test

# Run all tests
make test-all

# Build binary
make build

# Run locally
make run
```

### Test Endpoints
```bash
# Get debug info
curl http://localhost:9876

# Health check
curl http://localhost:9876/ping

# Get access logs
curl http://localhost:9876/logs
```

## Architecture

- **main.go**: Go HTTP server with /ping, /logs, and debug endpoints
- **main_test.go**: Comprehensive unit tests
- **integration_test.go**: Integration tests that build and test the Docker container
- **Dockerfile**: Multi-stage build creating minimal scratch-based container (≈6.5MB)
- **Makefile**: Build, test, and run targets
- **GitHub Actions**: Automated testing (unit + integration), build, and push to ghcr.io on main branch commits

## Deployment

### Continuous Integration
The CI workflow (`.github/workflows/docker-publish.yml`) runs on every push and PR to ensure code quality.

### Releases
Releases are automated through the release workflow (`.github/workflows/release.yml`):

1. Create a version tag: `git tag v1.0.0`
2. Push the tag: `git push origin v1.0.0`
3. GitHub Actions will automatically:
   - Run all tests
   - Build multi-platform binaries
   - Build and push Docker images with semantic version tags
   - Create a GitHub Release

### Docker Images
```bash
# Latest stable version
docker pull ghcr.io/tokuhirom/debug-httpd:latest

# Specific version
docker pull ghcr.io/tokuhirom/debug-httpd:v1.0.0
```

See `RELEASE.md` for detailed release process.