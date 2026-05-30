@echo off
REM Docker build script for Bengkelin API (Windows)
REM Usage: scripts\docker-build.bat [environment] [database]
REM Environment: dev, prod (default: dev)
REM Database: mysql, postgres (default: mysql)

setlocal enabledelayedexpansion

REM Default values
set ENV=%1
set DB=%2
if "%ENV%"=="" set ENV=dev
if "%DB%"=="" set DB=mysql

echo 🐳 Building Bengkelin API Docker images for %ENV% environment with %DB%...

REM Validate environment
if not "%ENV%"=="dev" if not "%ENV%"=="prod" (
    echo ❌ Invalid environment: %ENV%. Use 'dev' or 'prod'
    exit /b 1
)

REM Validate database
if not "%DB%"=="mysql" if not "%DB%"=="postgres" (
    echo ❌ Invalid database: %DB%. Use 'mysql' or 'postgres'
    exit /b 1
)

REM Build based on environment and database
if "%ENV%"=="dev" (
    echo 📦 Building development image...
    docker build -f Dockerfile.dev -t bengkelin-api:dev .
    if errorlevel 1 (
        echo ❌ Failed to build development image
        exit /b 1
    )
    echo ✅ Development image built successfully!
    
    echo 🔧 Building development services with %DB%...
    if "%DB%"=="postgres" (
        docker-compose -f docker-compose.yml -f docker-compose.dev.yml -f docker-compose.postgres-dev.yml build
    ) else (
        docker-compose -f docker-compose.yml -f docker-compose.dev.yml build
    )
    if errorlevel 1 (
        echo ❌ Failed to build development services
        exit /b 1
    )
    echo ✅ Development services built successfully!
) else (
    echo 📦 Building production image...
    docker build -f Dockerfile -t bengkelin-api:prod .
    if errorlevel 1 (
        echo ❌ Failed to build production image
        exit /b 1
    )
    echo ✅ Production image built successfully!
    
    echo 🔧 Building production services with %DB%...
    if "%DB%"=="postgres" (
        docker-compose -f docker-compose.yml -f docker-compose.prod.yml -f docker-compose.postgres-prod.yml build
    ) else (
        docker-compose -f docker-compose.yml -f docker-compose.prod.yml build
    )
    if errorlevel 1 (
        echo ❌ Failed to build production services
        exit /b 1
    )
    echo ✅ Production services built successfully!
)

REM Show image sizes
echo 📊 Image sizes:
docker images | findstr bengkelin-api

echo 🎉 Build completed for %ENV% environment with %DB%!

REM Show next steps
echo 📋 Next steps:
if "%ENV%"=="dev" (
    echo   Start development: scripts\docker-run.bat dev %DB%
    if "%DB%"=="postgres" (
        echo   View logs: docker-compose -f docker-compose.yml -f docker-compose.dev.yml -f docker-compose.postgres-dev.yml logs -f
    ) else (
        echo   View logs: docker-compose -f docker-compose.yml -f docker-compose.dev.yml logs -f
    )
) else (
    echo   Start production: scripts\docker-run.bat prod %DB%
    if "%DB%"=="postgres" (
        echo   View logs: docker-compose -f docker-compose.yml -f docker-compose.prod.yml -f docker-compose.postgres-prod.yml logs -f
    ) else (
        echo   View logs: docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f
    )
)

endlocal