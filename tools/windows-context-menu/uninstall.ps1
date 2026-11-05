# Simple Static CDN - Windows Context Menu Uninstaller
# Removes "Upload to Simple CDN" from Windows Explorer context menu

$ErrorActionPreference = "SilentlyContinue"

Write-Host "Removing Simple CDN context menu from Windows Registry..." -ForegroundColor Yellow

$regPath = "HKCU:\Software\Classes\*\shell\SimpleCDN"

if (Test-Path $regPath) {
    Remove-Item -Path $regPath -Recurse -Force
    Write-Host "Successfully removed Windows context menu." -ForegroundColor Green
} else {
    Write-Host "Context menu entry was not found in registry." -ForegroundColor Gray
}
