# Simple Static CDN - Windows Context Menu Uninstaller
# Removes "Upload to Simple CDN" from Windows Explorer context menu

$ErrorActionPreference = "SilentlyContinue"

Write-Host "Removing Simple CDN context menu from Windows Registry..." -ForegroundColor Yellow

$regKey = "HKCU\Software\Classes\*\shell\SimpleCDN"

# Use reg.exe directly to avoid PowerShell wildcard expansion hangs on the literal '*' key
$null = & reg.exe query $regKey 2>&1
if ($LASTEXITCODE -eq 0) {
    & reg.exe delete $regKey /f | Out-Null
    Write-Host "Successfully removed Windows context menu." -ForegroundColor Green
} else {
    Write-Host "Context menu entry was not found in registry." -ForegroundColor Gray
}
