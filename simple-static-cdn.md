# Self-Hosted Simple Static CDN

## Goal
Build a self-hosted, single-node static CDN in Go providing high-throughput file delivery with edge-grade caching headers, byte-range streaming, authenticated REST upload API, embedded drag-and-drop Web UI, and security controls (signed URLs, anti-hotlink, rate limiting).

## Architecture Overview
- **Core Server**: Go standard library / lightweight router with zero external heavy frameworks.
- **Storage**: Local filesystem storage with content-addressed hashing option (SHA-256) and path routing.
- **Delivery Pipeline**: Automatic MIME detection, HTTP 206 Partial Content (Byte-Range), strong ETags with 304 validation, Brotli/Gzip compression, and configurable `Cache-Control`.
- **Security & Protection**: HMAC-SHA256 URL signing with expiration (`?token=...&expires=...`), Referrer / Origin whitelisting, configurable CORS, and in-memory IP/token rate limiting.
- **Management**: Token-authenticated REST API (`/api/upload`, `/api/files`, `/api/purge`) + single-page embedded HTML/CSS/JS dashboard.
- **Deployment**: Multi-stage Docker build producing a scratch or distroless container (<20MB).

## Tasks
- [ ] Task 1: Initialize Go module and configuration subsystem (`config.go`) supporting YAML/JSON and environment variables. → Verify: Unit test configuration loader with default and overridden values.
- [ ] Task 2: Implement storage engine (`internal/storage`) managing file writes, metadata, atomic replacement, and deletion. → Verify: Test storing, retrieving, and deleting sample files on disk.
- [ ] Task 3: Implement core delivery handler (`internal/delivery`) with strong ETags, `If-None-Match`, `Cache-Control`, and HTTP 206 Byte-Range streaming. → Verify: `curl -I` and range header requests (`Range: bytes=0-100`) returning 200/206/304 statuses.
- [ ] Task 4: Implement compression middleware (Gzip and Brotli) for text and compressible assets with buffer pooling. → Verify: Requests with `Accept-Encoding: gzip, br` receive properly compressed response headers.
- [ ] Task 5: Implement security middleware (`internal/security`) for HMAC-SHA256 URL token validation, referrer hotlink blocking, CORS headers, and rate limiting. → Verify: Unit tests verifying expired/tampered tokens are rejected with 403, and valid tokens serve content.
- [ ] Task 6: Implement authenticated REST management API (`internal/api`) for file upload, file listing, deletion, and cache purge. → Verify: `curl -H "X-API-Key: ..."` successfully uploads a file and purges cache.
- [ ] Task 7: Build embedded single-file lightweight Web UI (`internal/web`) for browser drag-and-drop uploads, file browsing, and signed URL generation. → Verify: Open web dashboard in browser or test HTTP GET `/admin/` returning 200 OK.
- [ ] Task 8: Create production `Dockerfile` (multi-stage build) and `docker-compose.yml` with persistent volume mappings. → Verify: Container builds cleanly and runs locally.
- [ ] Task 9: Implement integration test suite and verification script (`verify.ps1` / `verify.sh`) exercising full end-to-end flows. → Verify: Automated script passes all delivery, security, and API checks.

## Done When
- [ ] Static files are served with sub-millisecond overhead, correct ETags, 304 handling, and byte-range support.
- [ ] Authenticated REST API allows uploading, listing, deleting, and cache purging.
- [ ] Web UI is embedded in the binary and allows visual upload and signed URL generation.
- [ ] Signed expiring URLs and hotlink referrer controls block unauthorized access.
- [ ] Docker container image builds and starts in isolated environment.

## Notes
- Keep external Go dependencies minimal to maximize maintainability and security footprint.
- All Web UI assets should be embedded directly using Go `embed.FS` to maintain a single static binary distribution.
