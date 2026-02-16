# Clean build: stop containers, optionally remove volume, rebuild from scratch, then run
Set-Location $PSScriptRoot

Write-Host "Stopping and removing existing containers..." -ForegroundColor Cyan
docker compose down

$volumeName = "pet_medical_pet_medical_pgdata"
$volumeExists = docker volume ls -q $volumeName 2>$null
if ($volumeExists) {
    Write-Host ""
    $response = Read-Host "Delete existing database volume '$volumeName'? (y/N)"
    if ($response -eq 'y' -or $response -eq 'Y') {
        docker volume rm $volumeName 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Volume removed." -ForegroundColor Green
        } else {
            Write-Host "Could not remove volume (may be in use). Continuing anyway." -ForegroundColor Yellow
        }
    }
}

Write-Host "Clean build (no cache)..." -ForegroundColor Cyan
$env:DEVELOPMENT = "true"
docker compose build --no-cache
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed." -ForegroundColor Red
    exit 1
}
Write-Host "Starting containers..." -ForegroundColor Cyan
docker compose up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to start containers." -ForegroundColor Red
    exit 1
}
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  Pet Medical is running" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "  URL:  http://localhost:8080"
Write-Host "  Port: 8080"
Write-Host ""
Write-Host "  Default login:"
Write-Host "    Email:    admin@example.com"
Write-Host "    Password: admin123"
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
