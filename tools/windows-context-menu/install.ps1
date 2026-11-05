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

$regPath = "HKCU:\Software\Classes\*\shell\SimpleCDN"
$regCommandPath = "$regPath\command"

Write-Host "Registering context menu for current user..." -ForegroundColor Yellow

# 1. Create main shell key
if (-not (Test-Path $regPath)) {
    New-Item -Path $regPath -Force | Out-Null
}

Set-ItemProperty -Path $regPath -Name "(Default)" -Value "Upload to Simple CDN"
Set-ItemProperty -Path $regPath -Name "Icon" -Value "imageres.dll,-1004"

# 2. Create command key
if (-not (Test-Path $regCommandPath)) {
    New-Item -Path $regCommandPath -Force | Out-Null
}

$commandValue = "powershell.exe -WindowStyle Hidden -ExecutionPolicy Bypass -File `"$uploadScript`" `"%1`""
Set-ItemProperty -Path $regCommandPath -Name "(Default)" -Value $commandValue

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
