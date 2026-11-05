# Windows Explorer Context Menu for Simple Static CDN

This tool integrates Simple Static CDN into the native Windows File Explorer right-click context menu. Right-clicking any file allows you to upload it directly to your CDN instance via the REST API and have the public URL automatically copied to your clipboard.

---

## Features

- **Native Windows Explorer Integration:** Adds an **"Upload to Simple CDN"** action when right-clicking any file.
- **Automatic Clipboard Copy:** The public URL returned by the CDN is immediately copied to your clipboard.
- **Windows Notifications:** Displays a native balloon tip/notification with the uploaded file name and public CDN URL.
- **No Administrator Privileges Required:** Installs directly under `HKEY_CURRENT_USER\Software\Classes`, requiring zero UAC prompts or admin rights.
- **Flexible Configuration:** Server URL and `X-API-Key` are stored in `%USERPROFILE%\.simple-cdn\config.json`.

---

## Installation

1. Open PowerShell and run:
   ```powershell
   cd tools/windows-context-menu
   powershell -ExecutionPolicy Bypass -File .\install.ps1
   ```
2. The installer registers the right-click menu entry and creates an initial configuration file at:
   ```
   %USERPROFILE%\.simple-cdn\config.json
   ```

---

## Configuration

Edit `%USERPROFILE%\.simple-cdn\config.json` to match your deployed CDN instance:

```json
{
  "api_url": "https://cdn.yourdomain.com",
  "api_key": "your-secret-api-key",
  "copy_to_clipboard": true
}
```

- `api_url`: The base URL of your CDN server (local or production domain).
- `api_key`: The secret key matching your `CDN_API_KEY`.
- `copy_to_clipboard`: Automatically copy the returned file URL to your clipboard upon upload (`true` or `false`).

---

## Usage

1. Open Windows File Explorer.
2. Right-click any file (image, video, document, archive, etc.).
3. Click **"Upload to Simple CDN"**.
4. A notification appears confirming the upload, and the CDN URL is ready to paste anywhere (`Ctrl + V`).

---

## Uninstallation

To remove the context menu entry from Windows:

```powershell
powershell -ExecutionPolicy Bypass -File .\uninstall.ps1
```
