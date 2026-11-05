# Simple Static CDN - Windows Context Menu Upload Handler
param(
    [Parameter(Mandatory = $false, Position = 0)]
    [string]$TargetFile
)

Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName Microsoft.VisualBasic

function Show-Notification {
    param (
        [string]$Title,
        [string]$Message,
        [string]$Icon = "Info"
    )
    
    $notifyIcon = New-Object System.Windows.Forms.NotifyIcon
    $notifyIcon.Icon = [System.Drawing.SystemIcons]::Information
    if ($Icon -eq "Error") {
        $notifyIcon.Icon = [System.Drawing.SystemIcons]::Error
    }
    $notifyIcon.Visible = $true
    
    $tooltipIcon = [System.Windows.Forms.ToolTipIcon]::Info
    if ($Icon -eq "Error") {
        $tooltipIcon = [System.Windows.Forms.ToolTipIcon]::Error
    }
    
    $notifyIcon.ShowBalloonTip(5000, $Title, $Message, $tooltipIcon)
    Start-Sleep -Milliseconds 200
    $notifyIcon.Dispose()
}

if (-not $TargetFile -or -not (Test-Path -Path $TargetFile -PathType Leaf)) {
    [System.Windows.Forms.MessageBox]::Show("No valid file selected for upload.", "Simple CDN", "OK", "Warning") | Out-Null
    exit 1
}

# 1. Load or initialize configuration
$configDir = Join-Path $env:USERPROFILE ".simple-cdn"
$configFile = Join-Path $configDir "config.json"

if (-not (Test-Path $configDir)) {
    New-Item -ItemType Directory -Path $configDir -Force | Out-Null
}

$config = $null
if (Test-Path $configFile) {
    try {
        $config = Get-Content -Path $configFile -Raw | ConvertFrom-Json
    } catch {
        $config = $null
    }
}

if (-not $config -or -not $config.api_url) {
    $defaultUrl = "http://localhost:8080"
    $serverUrl = [Microsoft.VisualBasic.Interaction]::InputBox(
        "Enter your Simple CDN Base URL (e.g., https://cdn.yourdomain.com):",
        "Simple CDN Setup",
        $defaultUrl
    )
    if (-not $serverUrl) { exit 1 }

    $apiKey = [Microsoft.VisualBasic.Interaction]::InputBox(
        "Enter your CDN X-API-Key:",
        "Simple CDN Setup",
        "cdn-secret-api-key-change-me"
    )
    if (-not $apiKey) { exit 1 }

    $configObj = @{
        api_url = $serverUrl.TrimEnd('/')
        api_key = $apiKey.Trim()
        copy_to_clipboard = $true
    }

    $configObj | ConvertTo-Json | Set-Content -Path $configFile -Force
    $config = [PSCustomObject]$configObj
}

$apiUrl = $config.api_url.TrimEnd('/')
$apiKey = $config.api_key
$fileName = [System.IO.Path]::GetFileName($TargetFile)

# 2. Upload using curl.exe
$uploadEndpoint = "$apiUrl/api/upload"

try {
    $curlOutput = & "C:\Windows\System32\curl.exe" -s -X POST $uploadEndpoint `
        -H "X-API-Key: $apiKey" `
        -F "file=@$TargetFile"

    if (-not $curlOutput) {
        throw "Failed to connect to Simple CDN at $uploadEndpoint"
    }

    $res = $curlOutput | ConvertFrom-Json

    if ($res.status -eq "success" -and $res.url) {
        Set-Clipboard -Value $res.url
        Show-Notification -Title "CDN Upload Successful" -Message "$fileName uploaded!`nURL copied to clipboard: $($res.url)" -Icon "Info"
    } elseif ($res.error) {
        Show-Notification -Title "CDN Upload Failed" -Message "Server error: $($res.error)" -Icon "Error"
    } else {
        Show-Notification -Title "CDN Upload Failed" -Message "Unexpected response: $curlOutput" -Icon "Error"
    }
} catch {
    Show-Notification -Title "CDN Upload Error" -Message $_.Exception.Message -Icon "Error"
}
