# Simple Static CDN

A high-performance, self-hosted, single-node static CDN built in pure Go. It delivers static assets with edge-grade HTTP caching, byte-range streaming for media, an authenticated REST management API, an embedded web dashboard, and security features like signed URLs, hotlink protection, and rate limiting.

---

## Features

- **High-Throughput Delivery Engine:**
  - Zero external runtime dependencies; pure Go standard library with zero-copy I/O streaming.
  - RFC 7232 compliant `ETag` generation with weak validator and `If-None-Match` `304 Not Modified` handling.
  - HTTP `206 Partial Content` (Byte-Range requests) for audio and video scrubbing.
  - In-memory `sync.Pool` Gzip compression for compressible assets (`text/*`, `application/javascript`, `image/svg+xml`, etc.).
  - Configurable `Cache-Control` header (defaults to `public, max-age=60`, customizable to `no-cache` or any desired policy).
- **Security & Protection:**
  - **HMAC-SHA256 Signed URLs:** Generate time-limited, tamper-proof download URLs (`?token=...&expires=...`).
  - **Anti-Hotlinking Guard:** Referrer and Origin whitelisting with wildcard domain support (e.g. `*.example.com`).
  - **Rate Limiting:** In-memory token-bucket rate limiter per IP address with automatic idle cleanup.
  - **CORS Support:** Configurable allowed origins, exposed streaming headers, and automated preflight handling.
  - **Safe Path Resolution:** Strict path traversal protection (`filepath.Clean`, boundary checks, null-byte rejection).
- **Management & Ingestion:**
  - **REST API:** Authenticated via `X-API-Key` or `Authorization: Bearer <key>` for uploading, listing, deleting, signing URLs, and cache purging.
  - **Embedded Web Dashboard:** Single-page drag-and-drop dashboard embedded directly in the binary via `embed.FS` (`http://localhost:8080/`).
  - **Windows Explorer Integration:** Right-click context menu integration for Windows PCs to upload files directly via the REST API and auto-copy the URL to the clipboard.
  - **Healthcheck Endpoint:** Unauthenticated `/healthz` and `/health` endpoints for Docker, Traefik, and Dokploy.
- **Deployment:**
  - Fully compatible with **Dokploy**, Docker, and Docker Compose.
  - Single static binary (<9MB) or minimal multi-stage Alpine container (<20MB).

---

## Quick Start

### 1. Deploying on Dokploy

#### Option A: Deploy via Git Repository (Dockerfile)
1. In Dokploy, create a new **Application** and connect this Git repository.
2. Select **Dockerfile** as the build type.
3. Configure the container port: `8080` (Dokploy automatically maps Traefik to this port).
4. Add a **Persistent Volume**:
   - **Mount Path:** `/data`
5. Under **Environment**, configure:
   - `CDN_BASE_URL=https://your-domain.com`
   - `CDN_API_KEY=your-secure-random-api-key`
   - `CDN_SECRET_KEY=your-secure-hmac-secret`
6. Click **Deploy**. The built-in `HEALTHCHECK` will report health status to Dokploy.

#### Option B: Deploy via Docker Compose
1. In Dokploy, create a new **Compose** project.
2. Select your repository or paste the contents of [`docker-compose.yml`](file:///c:/Users/lenovo/Desktop/Projects/antigravity/simple-cdn/docker-compose.yml).
3. Set your environment variables in Dokploy's Environment tab.
4. Click **Deploy**. The `cdn_data` volume preserves all uploaded files across redeployments.

#### Option C: Pull Pre-built Image from GHCR
Dokploy can directly pull the pre-built multi-arch image produced by GitHub Releases:
```bash
image: ghcr.io/<owner>/simple-cdn:latest
```

---

### 2. Running with Docker Compose Locally

```bash
docker compose up -d
```

Access the Web Dashboard at [http://localhost:8080](http://localhost:8080).

---

### 3. Running from Source / Binary

```bash
# Build the single standalone binary
go build -o cdn-server ./cmd/server

# Run the server
./cdn-server
```

---

### 4. Windows PC Right-Click Upload Integration

To enable right-clicking any file in Windows File Explorer and uploading it directly to your CDN:

```powershell
powershell -ExecutionPolicy Bypass -File .\tools\windows-context-menu\install.ps1
```

Configure your server URL and API key in `%USERPROFILE%\.simple-cdn\config.json`. See [Windows Context Menu Documentation](file:///c:/Users/lenovo/Desktop/Projects/antigravity/simple-cdn/tools/windows-context-menu/README.md) for complete options.

---

## Configuration

Configuration can be supplied via a JSON configuration file (set `CONFIG_PATH=/path/to/config.json`) or environment variables. Environment variables take precedence over config files.

| Environment Variable | Config Key | Default | Description |
|---|---|---|---|
| `PORT` or `CDN_PORT` | `port` | `8080` | Server listening port (supports Dokploy/PaaS `PORT`) |
| `CDN_HOST` | `host` | `0.0.0.0` | Bind IP address |
| `CDN_BASE_URL` | `base_url` | `http://localhost:8080` | Public base URL used for URL generation |
| `CDN_STORAGE_DIR` | `storage_dir` | `/data` (container) / `./data` (local) | Directory for storing hosted files |
| `CDN_MAX_UPLOAD_SIZE_MB` | `max_upload_size_mb` | `100` | Maximum upload size in megabytes |
| `CDN_CACHE_CONTROL` | `cache_control` | `public, max-age=60` | Default HTTP Cache-Control header (e.g. `public, max-age=60` or `no-cache`) |
| `CDN_ENABLE_GZIP` | `enable_gzip` | `true` | Enable on-the-fly Gzip compression |
| `CDN_API_KEY` | `api_key` | `cdn-secret-api-key-change-me` | Secret API key for REST endpoints |
| `CDN_SECRET_KEY` | `secret_key` | `hmac-signing-secret-change-me` | Secret key for HMAC-SHA256 signed URLs |
| `CDN_REQUIRE_SIGNED_URLS` | `require_signed_urls` | `false` | When true, all file delivery requires a valid signature |
| `CDN_ALLOWED_REFERRERS` | `allowed_referrers` | `[]` | Comma-separated list of allowed domains (e.g. `mysite.com,*.mysite.com`) |
| `CDN_ALLOWED_ORIGINS` | `allowed_origins` | `["*"]` | Comma-separated CORS allowed origins |
| `CDN_RATE_LIMIT_RPS` | `rate_limit_rps` | `100` | Allowed requests per second per IP |
| `CDN_RATE_LIMIT_BURST` | `rate_limit_burst` | `200` | Burst request allowance per IP |

---

## REST API Reference

All management endpoints require authentication via the `X-API-Key: <key>` header or `Authorization: Bearer <key>`.

### Upload File

```http
POST /api/upload
Content-Type: multipart/form-data
X-API-Key: your-api-key

Form fields:
- file: (binary data)
- path: documents/report.pdf (optional, defaults to filename)
```

**Response (201 Created):**
```json
{
  "status": "success",
  "path": "documents/report.pdf",
  "url": "http://localhost:8080/files/documents/report.pdf",
  "size": 1048576,
  "etag": "\"18e4f508e9b88744-2c\"",
  "content_type": "application/pdf",
  "sha256": "4b227777d4dd1fc61c6f884f48641d02b4d121d3fd328cb08b5531fcacdabf8a"
}
```

### List Files

```http
GET /api/files?prefix=documents/
X-API-Key: your-api-key
```

### Delete File

```http
DELETE /api/files/documents/report.pdf
X-API-Key: your-api-key
```

### Generate Signed URL

```http
POST /api/sign-url
Content-Type: application/json
X-API-Key: your-api-key

{
  "path": "documents/report.pdf",
  "ttl_seconds": 3600
}
```

**Response (200 OK):**
```json
{
  "url": "http://localhost:8080/files/documents/report.pdf?token=a1b2c3d4...&expires=1730761200",
  "expires_at": 1730761200,
  "ttl_sec": 3600
}
```

### Purge Cache

```http
POST /api/purge
Content-Type: application/json
X-API-Key: your-api-key

{
  "path": "*"
}
```

### Healthcheck

```http
GET /healthz
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "service": "simple-static-cdn"
}
```

---

## Automated CI/CD & Image Publishing

The repository uses a unified workflow [`.github/workflows/ci.yml`](file:///c:/Users/lenovo/Desktop/Projects/antigravity/simple-cdn/.github/workflows/ci.yml) that handles continuous integration and automated releases:

1. **On Every Pull Request:**
   - Validates code formatting (`gofmt -l .`).
   - Runs unit and integration tests with race detection and coverage (`go test -race -cover ./...`).
   - Validates multi-stage Docker build without pushing.

2. **On Push to `main`:**
   - Executes all quality and test suites.
   - **Automatic SemVer Tagging:** Automatically calculates the next version tag (`vX.Y.Z`) from commit history and pushes the git tag.
   - **Docker Multi-Arch Build & Push:** Builds multi-architecture images (`linux/amd64`, `linux/arm64`) using QEMU & Docker Buildx and publishes to GitHub Container Registry (`ghcr.io/ysn4irix/simple-cdn:<tag>` and `:latest`).
   - **Automated GitHub Release:** Cross-compiles standalone binaries for Linux (amd64, arm64), macOS (amd64, arm64), and Windows (amd64), generates SHA-256 checksums (`checksums.txt`), and attaches them to the new GitHub Release.

---

## Testing & Verification

```powershell
# Run unit tests
go test -v ./...

# Run complete end-to-end integration verification (PowerShell)
.\verify.ps1

# Or on Linux / macOS (Bash)
./verify.sh
```
