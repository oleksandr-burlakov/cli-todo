# Build script for CLI Todo (Windows / PowerShell)
# Usage: .\build.ps1  or  .\build.ps1 -Output ".\todo.exe"

param(
    [string]$Output = ".\todo.exe"
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go is not installed or not in PATH. Install from https://go.dev/dl/"
    exit 1
}

Write-Host "Tidying modules..." -ForegroundColor Cyan
go mod tidy
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Building..." -ForegroundColor Cyan
go build -o $Output .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Done: $Output" -ForegroundColor Green
