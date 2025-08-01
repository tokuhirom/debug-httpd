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

The image is automatically built and pushed to GitHub Container Registry (ghcr.io) when changes are pushed to the main branch. The workflow is defined in `.github/workflows/docker-publish.yml`.

To pull the published image:
```bash
docker pull ghcr.io/tokuhirom/debug-httpd:latest
```