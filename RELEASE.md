# Release Process

## Overview

Releases are automated through GitHub Actions when a version tag is pushed.

## Version Format

We use semantic versioning: `vMAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes
- **MINOR**: New features (backwards compatible)
- **PATCH**: Bug fixes

## Release Steps

### 1. Update Version in Documentation (if needed)

Make sure any version references in documentation are updated.

### 2. Create and Push Tag

```bash
# For a new patch release (e.g., v1.0.1)
git tag v1.0.1 -m "Fix: Description of fixes"
git push origin v1.0.1

# For a new minor release (e.g., v1.1.0)
git tag v1.1.0 -m "Feature: Description of new features"
git push origin v1.1.0

# For a new major release (e.g., v2.0.0)
git tag v2.0.0 -m "Breaking: Description of breaking changes"
git push origin v2.0.0
```

### 3. Automated Process

Once the tag is pushed, GitHub Actions will:

1. Run all tests (unit and integration)
2. Build binaries for multiple platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)
3. Build Docker images for multiple architectures
4. Push Docker images with appropriate tags:
   - `ghcr.io/tokuhirom/debug-httpd:v1.0.0` (exact version)
   - `ghcr.io/tokuhirom/debug-httpd:1.0` (minor version)
   - `ghcr.io/tokuhirom/debug-httpd:1` (major version)
   - `ghcr.io/tokuhirom/debug-httpd:latest` (latest stable)
5. Create a GitHub Release with:
   - Pre-built binaries
   - SHA256 checksums
   - Auto-generated release notes

## Pre-release Versions

For pre-release versions (alpha, beta, rc):

```bash
# Alpha release
git tag v1.0.0-alpha.1 -m "Alpha release"

# Beta release  
git tag v1.0.0-beta.1 -m "Beta release"

# Release candidate
git tag v1.0.0-rc.1 -m "Release candidate"
```

Pre-releases will be marked as such in GitHub Releases and won't update the `latest` Docker tag.

## Rollback

If a release has issues:

1. Delete the tag locally and remotely:
   ```bash
   git tag -d v1.0.0
   git push origin :refs/tags/v1.0.0
   ```

2. Fix the issues

3. Create a new patch version

## Docker Image Tags

After a successful release, the following Docker tags will be available:

- **Exact version**: `ghcr.io/tokuhirom/debug-httpd:v1.2.3`
- **Minor version**: `ghcr.io/tokuhirom/debug-httpd:1.2` (points to latest patch)
- **Major version**: `ghcr.io/tokuhirom/debug-httpd:1` (points to latest minor)
- **Latest**: `ghcr.io/tokuhirom/debug-httpd:latest` (latest stable release)

## Verifying a Release

After release:

1. Check GitHub Actions for successful workflow
2. Verify GitHub Release page
3. Test Docker image:
   ```bash
   docker pull ghcr.io/tokuhirom/debug-httpd:v1.0.0
   docker run --rm ghcr.io/tokuhirom/debug-httpd:v1.0.0 --version
   ```