# GhostDraft Production Build Script (PowerShell)
# This script builds the desktop app with Turso credentials embedded

# Load environment variables from .env if it exists
if (Test-Path .env) {
    Get-Content .env | ForEach-Object {
        if ($_ -match '^([^#][^=]+)=(.*)$') {
            [Environment]::SetEnvironmentVariable($matches[1], $matches[2], 'Process')
        }
    }
}

# Get variables
$TURSO_DATABASE_URL = $env:TURSO_DATABASE_URL
$TURSO_AUTH_TOKEN = $env:TURSO_AUTH_TOKEN

# Validate required variables
if ([string]::IsNullOrEmpty($TURSO_DATABASE_URL)) {
    Write-Host "Error: TURSO_DATABASE_URL is required" -ForegroundColor Red
    Write-Host "Set it in .env or as an environment variable"
    exit 1
}

if ([string]::IsNullOrEmpty($TURSO_AUTH_TOKEN)) {
    Write-Host "Error: TURSO_AUTH_TOKEN is required" -ForegroundColor Red
    Write-Host "Set it in .env or as an environment variable"
    exit 1
}

Write-Host "Building GhostDraft..."
Write-Host "  Turso URL: $TURSO_DATABASE_URL"
Write-Host "  Auth Token: $($TURSO_AUTH_TOKEN.Substring(0, [Math]::Min(8, $TURSO_AUTH_TOKEN.Length)))..."

# Build with ldflags
$ldflags = "-X 'ghostdraft/internal/data.TursoURL=$TURSO_DATABASE_URL' -X 'ghostdraft/internal/data.TursoAuthToken=$TURSO_AUTH_TOKEN'"

wails build -ldflags $ldflags

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build complete! Output: build/bin/GhostDraft.exe" -ForegroundColor Green
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
