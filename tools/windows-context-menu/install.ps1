# Simple Static CDN - Windows Context Menu Installer
# Installs right-click "Upload to Simple CDN" for all files in Windows Explorer

$ErrorActionPreference = "Stop"

$scriptDir = $PSScriptRoot
$uploadScript = Join-Path $scriptDir "upload.ps1"

if (-not (Test-Path $uploadScript)) {
    Write-Error "upload.ps1 not found in $scriptDir"
    exit 1
}

Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "  Simple Static CDN - Windows Context Menu Setup  " -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan

$regKey = "HKCU\Software\Classes\*\shell\SimpleCDN"
$regCommandKey = "HKCU\Software\Classes\*\shell\SimpleCDN\command"

Write-Host "Registering context menu for current user..." -ForegroundColor Yellow

# Use reg.exe directly to avoid PowerShell wildcard expansion hangs on the literal '*' key
& reg.exe add $regKey /ve /d "Upload to Simple CDN" /f | Out-Null
& reg.exe add $regKey /v "Icon" /d "imageres.dll,-1004" /f | Out-Null

$commandValue = "powershell.exe -WindowStyle Hidden -ExecutionPolicy Bypass -File \`"$uploadScript\`" \`"%1\`""
& reg.exe add $regCommandKey /ve /d $commandValue /f | Out-Null

# 3. Create default configuration template if not exists
$configDir = Join-Path $env:USERPROFILE ".simple-cdn"
$configFile = Join-Path $configDir "config.json"

if (-not (Test-Path $configFile)) {
    New-Item -ItemType Directory -Path $configDir -Force | Out-Null
    $defaultConfig = @{
        api_url = "http://localhost:8080"
        api_key = "cdn-secret-api-key-change-me"
        copy_to_clipboard = $true
    }
    $defaultConfig | ConvertTo-Json | Set-Content -Path $configFile -Force
    Write-Host "Default configuration created at $configFile" -ForegroundColor Green
} else {
    Write-Host "Existing configuration found at $configFile" -ForegroundColor Green
}

Write-Host "`nSuccessfully installed Windows context menu!" -ForegroundColor Green
Write-Host "You can now right-click any file in Windows Explorer and select 'Upload to Simple CDN'." -ForegroundColor Cyan
Write-Host "Configuration file: $configFile" -ForegroundColor Gray
