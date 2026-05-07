# EndGit Installer for Windows

param(
    [string]$InstallPath = "$env:LOCALAPPDATA\endgit"
)

$ErrorActionPreference = "Stop"

$repo = "two-tech-dev/endgit-cli"
$apiUrl = "https://api.github.com/repos/$repo/releases/latest"

Write-Host "EndGit Installer" -ForegroundColor Cyan
Write-Host "================" -ForegroundColor Cyan
Write-Host ""

try {
    Write-Host "Fetching latest release information..." -ForegroundColor Yellow

    $response = Invoke-RestMethod -Uri $apiUrl -Headers @{
        "User-Agent" = "endgit-installer"
    }

    if (-not $response -or -not $response.tag_name) {
        Write-Host "Error: Failed to fetch release information." -ForegroundColor Red
        exit 1
    }

    $version = $response.tag_name
    Write-Host "Latest version: $version" -ForegroundColor Green

    # Find Windows asset
    $asset = $response.assets | Where-Object { $_.name -eq "endgit-windows-amd64.exe" }

    if (-not $asset) {
        $asset = $response.assets | Where-Object {
            $_.name -like "*windows*.exe" -or $_.name -like "*win*.exe"
        }
    }

    if (-not $asset) {
        $asset = $response.assets | Where-Object { $_.name -like "*.exe" }
    }

    if (-not $asset) {
        Write-Host "Error: No Windows executable found in latest release." -ForegroundColor Red
        Write-Host "Available assets:" -ForegroundColor Yellow
        $response.assets | ForEach-Object { Write-Host "  - $($_.name)" }
        exit 1
    }

    $downloadUrl = $asset.browser_download_url
    $fileName = $asset.name

    Write-Host "Found asset: $fileName" -ForegroundColor Green
    Write-Host ""

    if (-not (Test-Path $InstallPath)) {
        Write-Host "Creating installation directory: $InstallPath" -ForegroundColor Yellow
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }

    $exePath = Join-Path $InstallPath "endgit.exe"

    Write-Host "Downloading $fileName..." -ForegroundColor Yellow
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $downloadUrl -OutFile $exePath -UseBasicParsing
    $ProgressPreference = 'Continue'

    Write-Host "Downloaded successfully!" -ForegroundColor Green

    if (-not (Test-Path $exePath)) {
        Write-Host "Error: Download failed." -ForegroundColor Red
        exit 1
    }

    $fileSize = (Get-Item $exePath).Length
    if ($fileSize -eq 0) {
        Write-Host "Error: Downloaded file is empty." -ForegroundColor Red
        Remove-Item $exePath -Force
        exit 1
    }

    Write-Host "Installed to: $exePath ($([math]::Round($fileSize/1KB, 2)) KB)" -ForegroundColor Green

    Write-Host ""
    Write-Host "Adding to user PATH..." -ForegroundColor Yellow

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")

    if ($userPath -notlike "*$InstallPath*") {
        $newPath = "$userPath;$InstallPath"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Host "Added $InstallPath to PATH" -ForegroundColor Green
        Write-Host "Restart terminal to apply changes." -ForegroundColor Yellow
    } else {
        Write-Host "Already in PATH" -ForegroundColor Green
    }

    Write-Host ""
    Write-Host "Installation complete!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Usage:" -ForegroundColor Cyan
    Write-Host "  endgit search <plugin>" -ForegroundColor White
    Write-Host "  endgit install <plugin>" -ForegroundColor White
    Write-Host "  endgit init" -ForegroundColor White

} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Installation failed." -ForegroundColor Red
    exit 1
}