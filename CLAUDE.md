# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a debug HTTP server project that creates a Docker container for debugging purposes. The server displays:
- Environment variables
- Host information (hostname, FQDN, IP addresses)
- Request details (headers, client info)
- Python version

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

### Test the Server
```bash
curl http://localhost:9876
```

## Architecture

- **Dockerfile**: Multi-stage build creating a Python-based HTTP server
- **server.py**: Embedded in Dockerfile, creates a JSON API endpoint that returns debug information
- **GitHub Actions**: Automated build and push to ghcr.io on main branch commits

## Deployment

The image is automatically built and pushed to GitHub Container Registry (ghcr.io) when changes are pushed to the main branch. The workflow is defined in `.github/workflows/docker-publish.yml`.

To pull the published image:
```bash
docker pull ghcr.io/tokuhirom/debug-httpd:latest
```