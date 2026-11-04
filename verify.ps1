# End-to-end integration test and verification script for Simple Static CDN

$ErrorActionPreference = "Stop"

$goExe = "C:\Users\lenovo\go-toolchain\go\bin\go.exe"
if (-not (Test-Path $goExe)) {
    $goExe = "go"
}

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Simple Static CDN - End-to-End Verification  " -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan

# 1. Run unit test suite
Write-Host "`n[1/3] Running full Go unit test suite..." -ForegroundColor Yellow
$testOutput = & $goExe test -v ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "Unit tests failed!" -ForegroundColor Red
    exit 1
}
Write-Host "All unit tests passed successfully." -ForegroundColor Green

# 2. Build executable
Write-Host "`n[2/3] Building executable binary..." -ForegroundColor Yellow
$binPath = "$PSScriptRoot\cdn-verify.exe"
& $goExe build -o $binPath ./cmd/server
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "Binary built successfully ($binPath)." -ForegroundColor Green

# 3. Launch server in background on test port
Write-Host "`n[3/3] Running live HTTP integration checks..." -ForegroundColor Yellow

$testPort = 8765
$testStorage = "$PSScriptRoot\.test-cdn-data"
if (Test-Path $testStorage) {
    Remove-Item -Path $testStorage -Recurse -Force
}

$env:CDN_PORT = "$testPort"
$env:CDN_HOST = "127.0.0.1"
$env:CDN_STORAGE_DIR = $testStorage
$env:CDN_BASE_URL = "http://127.0.0.1:$testPort"
$env:CDN_API_KEY = "verify-test-key-999"
$env:CDN_SECRET_KEY = "verify-secret-sign-key"
$env:CDN_ALLOWED_REFERRERS = "trusted.com,*.company.org"

$serverProcess = Start-Process -FilePath $binPath -PassThru -NoNewWindow
Start-Sleep -Seconds 1

$sampleFile = "$PSScriptRoot\.sample-verify.txt"

try {
    # Check A: Web UI & Healthcheck
    Write-Host "  -> Probing Web UI dashboard (/)..." -NoNewline
    Start-Sleep -Seconds 2
    $uiRes = curl.exe -s -i "http://127.0.0.1:$testPort/"
    if (-not ($uiRes -match "Simple Static CDN")) {
        throw "Web UI check failed. Response: $uiRes"
    }
    Write-Host " OK (200)" -ForegroundColor Green

    Write-Host "  -> Probing Healthcheck endpoint (/healthz)..." -NoNewline
    $healthRes = curl.exe -s -i "http://127.0.0.1:$testPort/healthz"
    if (-not ($healthRes -match "healthy")) {
        throw "Healthcheck failed. Response: $healthRes"
    }
    Write-Host " OK (200, healthy)" -ForegroundColor Green

    # Check B: REST Upload
    Write-Host "  -> Testing file upload via /api/upload..." -NoNewline
    $sampleFile = "$PSScriptRoot\.sample-verify.txt"
    "The quick brown fox jumps over the lazy dog." | Out-File -FilePath $sampleFile -Encoding ascii -NoNewline

    $curlOutput = curl.exe -s -X POST "http://127.0.0.1:$testPort/api/upload" `
        -H "X-API-Key: verify-test-key-999" `
        -F "path=documents/sample.txt" `
        -F "file=@$sampleFile" | ConvertFrom-Json

    if ($curlOutput.status -ne "success" -or $curlOutput.path -ne "documents/sample.txt") {
        throw "Upload failed or returned unexpected payload: $curlOutput"
    }
    $fileEtag = $curlOutput.etag
    Write-Host " OK (Uploaded, ETag: $fileEtag)" -ForegroundColor Green

    # Check C: Static Delivery & Headers
    Write-Host "  -> Probing static file delivery (/files/documents/sample.txt)..." -NoNewline
    $delivRes = Invoke-WebRequest -Uri "http://127.0.0.1:$testPort/files/documents/sample.txt" -UseBasicParsing
    if ($delivRes.StatusCode -ne 200 -or $delivRes.Headers["Accept-Ranges"] -ne "bytes") {
        throw "Delivery headers check failed"
    }
    Write-Host " OK (200, Accept-Ranges: bytes)" -ForegroundColor Green

    # Check D: ETag Caching (304 Not Modified)
    Write-Host "  -> Testing ETag 304 validation..." -NoNewline
    $etagCurl = curl.exe -s -i -H "If-None-Match: $fileEtag" "http://127.0.0.1:$testPort/files/documents/sample.txt"
    if (-not ($etagCurl -match "304 Not Modified")) {
        # Check with unquoted fallback in case powershell stripped quotes
        $etagCurl = curl.exe -s -i -H ('If-None-Match: ' + $fileEtag) "http://127.0.0.1:$testPort/files/documents/sample.txt"
    }
    if (-not ($etagCurl -match "304 Not Modified")) {
        throw "ETag conditional check failed. Output: $etagCurl"
    }
    Write-Host " OK (304 Not Modified)" -ForegroundColor Green

    # Check E: HTTP 206 Byte-Range Streaming
    Write-Host "  -> Testing HTTP 206 Byte-Range streaming (bytes=0-8)..." -NoNewline
    $rangeOutput = curl.exe -s -i -H "Range: bytes=0-8" "http://127.0.0.1:$testPort/files/documents/sample.txt"
    if (-not ($rangeOutput -match "206 Partial Content")) {
        throw "Byte-Range streaming failed. Output: $rangeOutput"
    }
    Write-Host " OK (206 Partial Content)" -ForegroundColor Green

    # Check F: Gzip Compression
    Write-Host "  -> Testing Gzip compression..." -NoNewline
    $gzipHeaders = curl.exe -s -I -H "Accept-Encoding: gzip" "http://127.0.0.1:$testPort/files/documents/sample.txt"
    if (-not ($gzipHeaders -match "Content-Encoding: gzip")) {
        throw "Gzip compression failed. Output: $gzipHeaders"
    }
    Write-Host " OK (Content-Encoding: gzip)" -ForegroundColor Green

    # Check G: Signed URL generation and enforcement
    Write-Host "  -> Testing HMAC-SHA256 Signed URLs..." -NoNewline
    $signRes = Invoke-RestMethod -Uri "http://127.0.0.1:$testPort/api/sign-url" `
        -Method Post `
        -Headers @{"X-API-Key"="verify-test-key-999"} `
        -ContentType "application/json" `
        -Body '{"path":"documents/sample.txt","ttl_seconds":60}'

    $signedUrl = $signRes.url
    $signedGet = curl.exe -s -o NUL -w "%{http_code}" $signedUrl
    if ($signedGet -ne "200") {
        throw "Signed URL access failed, got HTTP $signedGet"
    }

    # Tampered signature check
    $tamperedUrl = $signedUrl -replace "token=", "token=tampered"
    $tamperedGet = curl.exe -s -o NUL -w "%{http_code}" $tamperedUrl
    if ($tamperedGet -ne "403") {
        throw "Tampered token should return 403 Forbidden, got $tamperedGet"
    }
    Write-Host " OK (Valid: 200, Tampered: 403)" -ForegroundColor Green

    # Check H: Anti-Hotlinking Referrer check
    Write-Host "  -> Testing Anti-Hotlinking..." -NoNewline
    $leechGet = curl.exe -s -o NUL -w "%{http_code}" -H "Referer: https://pirate-site.com" "http://127.0.0.1:$testPort/files/documents/sample.txt"
    if ($leechGet -ne "403") {
        throw "Unauthorized referer should be blocked with 403, got $leechGet"
    }
    $trustedGet = curl.exe -s -o NUL -w "%{http_code}" -H "Referer: https://trusted.com/page" "http://127.0.0.1:$testPort/files/documents/sample.txt"
    if ($trustedGet -ne "200") {
        throw "Authorized referer should return 200, got $trustedGet"
    }
    Write-Host " OK (Leech: 403, Trusted: 200)" -ForegroundColor Green

    # Check I: Cache Purge
    Write-Host "  -> Testing Cache Purge API (/api/purge)..." -NoNewline
    $purgeOutput = Invoke-RestMethod -Uri "http://127.0.0.1:$testPort/api/purge" `
        -Method Post `
        -Headers @{"X-API-Key"="verify-test-key-999"} `
        -ContentType "application/json" `
        -Body '{"path":"*"}'
    if ($purgeOutput.status -ne "purged") {
        throw "Purge API failed"
    }
    Write-Host " OK (Cache Invalidation)" -ForegroundColor Green

    # Check J: File Deletion
    Write-Host "  -> Testing File Deletion (/api/files/documents/sample.txt)..." -NoNewline
    $delOutput = Invoke-RestMethod -Uri "http://127.0.0.1:$testPort/api/files/documents/sample.txt" `
        -Method Delete `
        -Headers @{"X-API-Key"="verify-test-key-999"}
    if ($delOutput.status -ne "deleted") {
        throw "Delete failed"
    }
    $afterDelGet = curl.exe -s -o NUL -w "%{http_code}" "http://127.0.0.1:$testPort/files/documents/sample.txt"
    if ($afterDelGet -ne "404") {
        throw "Deleted file should return 404, got $afterDelGet"
    }
    Write-Host " OK (Deleted & 404 verified)" -ForegroundColor Green

    Write-Host "`n===============================================" -ForegroundColor Cyan
    Write-Host "  ALL VERIFICATION CHECKS PASSED (100%)       " -ForegroundColor Green
    Write-Host "===============================================" -ForegroundColor Cyan
}
finally {
    if ($serverProcess -and -not $serverProcess.HasExited) {
        Stop-Process -Id $serverProcess.Id -Force -ErrorAction SilentlyContinue
    }
    if ($sampleFile -and (Test-Path $sampleFile)) {
        Remove-Item -Path $sampleFile -Force -ErrorAction SilentlyContinue
    }
    if ($binPath -and (Test-Path $binPath)) {
        Remove-Item -Path $binPath -Force -ErrorAction SilentlyContinue
    }
    if ($testStorage -and (Test-Path $testStorage)) {
        Remove-Item -Path $testStorage -Recurse -Force -ErrorAction SilentlyContinue
    }
}
