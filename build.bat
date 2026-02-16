@echo off
cd /d "%~dp0"

echo Stopping and removing existing containers...
docker compose down

docker volume ls -q pet_medical_pet_medical_pgdata >nul 2>&1
if %errorlevel% equ 0 (
    echo.
    set /p response="Delete existing database volume? (y/N): "
    if /i "%response%"=="y" (
        docker volume rm pet_medical_pet_medical_pgdata 2>nul
        echo Volume removed.
    )
)

echo Clean build (no cache)...
set DEVELOPMENT=true
docker compose build --no-cache
if errorlevel 1 (
    echo Build failed.
    exit /b 1
)
echo Starting containers...
docker compose up -d
if errorlevel 1 (
    echo Failed to start containers.
    exit /b 1
)
echo.
echo ========================================
echo   Pet Medical is running
echo ========================================
echo   URL:  http://localhost:8080
echo   Port: 8080
echo.
echo   Default login:
echo     Email:    admin@example.com
echo     Password: admin123
echo ========================================
echo.
