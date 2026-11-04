#!/usr/bin/env bash
set -euo pipefail

echo "==============================================="
echo "  Simple Static CDN - End-to-End Verification  "
echo "==============================================="

# 1. Run unit test suite
echo -e "\n[1/3] Running full Go unit test suite..."
go test -v ./...
echo "All unit tests passed successfully."

# 2. Build executable
echo -e "\n[2/3] Building executable binary..."
BIN_PATH="./cdn-verify"
go build -o "${BIN_PATH}" ./cmd/server
echo "Binary built successfully (${BIN_PATH})."

# 3. Running live HTTP integration checks
echo -e "\n[3/3] Running live HTTP integration checks..."
TEST_PORT=8765
TEST_STORAGE="./.test-cdn-data"
rm -rf "${TEST_STORAGE}"

export CDN_PORT="${TEST_PORT}"
export CDN_HOST="127.0.0.1"
export CDN_STORAGE_DIR="${TEST_STORAGE}"
export CDN_BASE_URL="http://127.0.0.1:${TEST_PORT}"
export CDN_API_KEY="verify-test-key-999"
export CDN_SECRET_KEY="verify-secret-sign-key"
export CDN_ALLOWED_REFERRERS="trusted.com,*.company.org"

"${BIN_PATH}" &
SERVER_PID=$!

cleanup() {
    kill "${SERVER_PID}" 2>/dev/null || true
    rm -f "${BIN_PATH}" .sample-verify.txt
    rm -rf "${TEST_STORAGE}"
}
trap cleanup EXIT

sleep 2

# Check A: Web UI & Healthcheck
echo -n "  -> Probing Web UI dashboard (/)... "
UI_RES=$(curl -s "http://127.0.0.1:${TEST_PORT}/")
if [[ "${UI_RES}" != *"Simple Static CDN"* ]]; then
    echo "FAILED"
    exit 1
fi
echo "OK (200)"

echo -n "  -> Probing Healthcheck (/healthz)... "
HEALTH_RES=$(curl -s "http://127.0.0.1:${TEST_PORT}/healthz")
if [[ "${HEALTH_RES}" != *"healthy"* ]]; then
    echo "FAILED"
    exit 1
fi
echo "OK (200, healthy)"

# Check B: REST Upload
echo -n "  -> Testing file upload via /api/upload... "
echo -n "The quick brown fox jumps over the lazy dog." > .sample-verify.txt
UPLOAD_JSON=$(curl -s -X POST "http://127.0.0.1:${TEST_PORT}/api/upload" \
    -H "X-API-Key: verify-test-key-999" \
    -F "path=documents/sample.txt" \
    -F "file=@.sample-verify.txt")

FILE_ETAG=$(echo "${UPLOAD_JSON}" | grep -o '"etag":"[^"]*"' | cut -d'"' -f4)
echo "OK (Uploaded, ETag: ${FILE_ETAG})"

# Check C: Static Delivery
echo -n "  -> Probing static file delivery (/files/documents/sample.txt)... "
DELIV_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:${TEST_PORT}/files/documents/sample.txt")
if [ "${DELIV_CODE}" != "200" ]; then
    echo "FAILED (Status ${DELIV_CODE})"
    exit 1
fi
echo "OK (200)"

# Check D: ETag 304 validation
echo -n "  -> Testing ETag 304 validation... "
ETAG_CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "If-None-Match: ${FILE_ETAG}" "http://127.0.0.1:${TEST_PORT}/files/documents/sample.txt")
if [ "${ETAG_CODE}" != "304" ]; then
    echo "FAILED (Expected 304, got ${ETAG_CODE})"
    exit 1
fi
echo "OK (304 Not Modified)"

# Check E: HTTP 206 Byte-Range
echo -n "  -> Testing HTTP 206 Byte-Range streaming (bytes=0-8)... "
RANGE_CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "Range: bytes=0-8" "http://127.0.0.1:${TEST_PORT}/files/documents/sample.txt")
if [ "${RANGE_CODE}" != "206" ]; then
    echo "FAILED (Expected 206, got ${RANGE_CODE})"
    exit 1
fi
echo "OK (206 Partial Content)"

# Check F: Gzip compression
echo -n "  -> Testing Gzip compression... "
GZIP_HEADER=$(curl -s -I -H "Accept-Encoding: gzip" "http://127.0.0.1:${TEST_PORT}/files/documents/sample.txt" | grep -i "content-encoding: gzip" || true)
if [ -z "${GZIP_HEADER}" ]; then
    echo "FAILED"
    exit 1
fi
echo "OK (Content-Encoding: gzip)"

# Check G: Signed URL
echo -n "  -> Testing HMAC-SHA256 Signed URLs... "
SIGN_RES=$(curl -s -X POST "http://127.0.0.1:${TEST_PORT}/api/sign-url" \
    -H "X-API-Key: verify-test-key-999" \
    -H "Content-Type: application/json" \
    -d '{"path":"documents/sample.txt","ttl_seconds":60}')

SIGNED_URL=$(echo "${SIGN_RES}" | grep -o '"url":"[^"]*"' | cut -d'"' -f4)
SIGNED_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${SIGNED_URL}")
if [ "${SIGNED_CODE}" != "200" ]; then
    echo "FAILED (Expected 200, got ${SIGNED_CODE})"
    exit 1
fi

TAMPERED_URL="${SIGNED_URL/token=/token=tampered}"
TAMPERED_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${TAMPERED_URL}")
if [ "${TAMPERED_CODE}" != "403" ]; then
    echo "FAILED (Expected 403, got ${TAMPERED_CODE})"
    exit 1
fi
echo "OK (Valid: 200, Tampered: 403)"

# Check H: Anti-Hotlinking
echo -n "  -> Testing Anti-Hotlinking... "
LEECH_CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "Referer: https://leech-site.com" "http://127.0.0.1:${TEST_PORT}/files/documents/sample.txt")
TRUSTED_CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "Referer: https://trusted.com/page" "http://127.0.0.1:${TEST_PORT}/files/documents/sample.txt")
if [ "${LEECH_CODE}" != "403" ] || [ "${TRUSTED_CODE}" != "200" ]; then
    echo "FAILED (Leech: ${LEECH_CODE}, Trusted: ${TRUSTED_CODE})"
    exit 1
fi
echo "OK (Leech: 403, Trusted: 200)"

# Check I: Cache Purge
echo -n "  -> Testing Cache Purge API (/api/purge)... "
PURGE_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://127.0.0.1:${TEST_PORT}/api/purge" \
    -H "X-API-Key: verify-test-key-999" \
    -H "Content-Type: application/json" \
    -d '{"path":"*"}')
if [ "${PURGE_CODE}" != "200" ]; then
    echo "FAILED (Status ${PURGE_CODE})"
    exit 1
fi
echo "OK (Cache Invalidation)"

# Check J: File Deletion
echo -n "  -> Testing File Deletion (/api/files/documents/sample.txt)... "
DEL_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "http://127.0.0.1:${TEST_PORT}/api/files/documents/sample.txt" \
    -H "X-API-Key: verify-test-key-999")
AFTER_DEL_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:${TEST_PORT}/files/documents/sample.txt")
if [ "${DEL_CODE}" != "200" ] || [ "${AFTER_DEL_CODE}" != "404" ]; then
    echo "FAILED (Delete: ${DEL_CODE}, After: ${AFTER_DEL_CODE})"
    exit 1
fi
echo "OK (Deleted & 404 verified)"

echo -e "\n==============================================="
echo "  ALL VERIFICATION CHECKS PASSED (100%)       "
echo "==============================================="
