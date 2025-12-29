@echo off
REM Docker run script for Bengkelin API (Windows)
REM Usage: scripts\docker-run.bat [environment] [database] [action]
REM Environment: dev, prod (default: dev)
REM Database: mysql, postgres (default: mysql)
REM Action: up, down, restart, logs, status (default: up)

setlocal enabledelayedexpansion

REM Default values
set ENV=%1
set DB=%2
set ACTION=%3
if "%ENV%"=="" set ENV=dev
if "%DB%"=="" set DB=mysql
if "%ACTION%"=="" set ACTION=up

echo 🐳 Managing Bengkelin API Docker containers...

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

REM Validate action
if not "%ACTION%"=="up" if not "%ACTION%"=="down" if not "%ACTION%"=="restart" if not "%ACTION%"=="logs" if not "%ACTION%"=="status" (
    echo ❌ Invalid action: %ACTION%. Use 'up', 'down', 'restart', 'logs', or 'status'
    exit /b 1
)

REM Set compose files based on environment and database
if "%ENV%"=="dev" (
    if "%DB%"=="postgres" (
        set COMPOSE_FILES=-f docker-compose.yml -f docker-compose.dev.yml -f docker-compose.postgres-dev.yml
        set PROJECT_NAME=bengkelin-dev-postgres
    ) else (
        set COMPOSE_FILES=-f docker-compose.yml -f docker-compose.dev.yml
        set PROJECT_NAME=bengkelin-dev
    )
) else (
    if "%DB%"=="postgres" (
        set COMPOSE_FILES=-f docker-compose.yml -f docker-compose.prod.yml -f docker-compose.postgres-prod.yml
        set PROJECT_NAME=bengkelin-prod-postgres
    ) else (
        set COMPOSE_FILES=-f docker-compose.yml -f docker-compose.prod.yml
        set PROJECT_NAME=bengkelin-prod
    )
)

REM Execute action
if "%ACTION%"=="up" (
    echo 🚀 Starting %ENV% environment with %DB%...
    
    REM Check if .env file exists
    if not exist .env (
        echo ⚠️  .env file not found. Creating from .env.example...
        copy .env.example .env
        echo 📝 Please edit .env file with your configuration before running again.
        exit /b 1
    )
    
    REM Start services
    docker-compose %COMPOSE_FILES% -p %PROJECT_NAME% up -d
    if errorlevel 1 (
        echo ❌ Failed to start services
        exit /b 1
    )
    
    echo ✅ %ENV% environment with %DB% started successfully!
    
    REM Show status
    echo 📊 Container status:
    docker-compose %COMPOSE_FILES% -p %PROJECT_NAME% ps
    
    REM Show URLs
    echo 🌐 Available services:
    if "%ENV%"=="dev" (
        echo   API: http://localhost:3000
        if "%DB%"=="postgres" (
            echo   pgAdmin: http://localhost:8082 ^(admin@bengkelin.com / admin123^)
            echo   PostgreSQL: localhost:5432
        ) else (
            echo   phpMyAdmin: http://localhost:8080
            echo   MySQL: localhost:3306
        )
        echo   Redis Commander: http://localhost:8081
        echo   MailHog: http://localhost:8025
    ) else (
        echo   API: https://localhost ^(via Nginx^)
        echo   API Direct: http://localhost:3000
    )
) else if "%ACTION%"=="down" (
    echo 🛑 Stopping %ENV% environment with %DB%...
    docker-compose %COMPOSE_FILES% -p %PROJECT_NAME% down
    echo ✅ %ENV% environment stopped successfully!
) else if "%ACTION%"=="restart" (
    echo 🔄 Restarting %ENV% environment with %DB%...
    docker-compose %COMPOSE_FILES% -p %PROJECT_NAME% restart
    echo ✅ %ENV% environment restarted successfully!
) else if "%ACTION%"=="logs" (
    echo 📋 Showing logs for %ENV% environment with %DB%...
    docker-compose %COMPOSE_FILES% -p %PROJECT_NAME% logs -f
) else if "%ACTION%"=="status" (
    echo 📊 Status for %ENV% environment with %DB%:
    docker-compose %COMPOSE_FILES% -p %PROJECT_NAME% ps
    
    echo.
    echo 💾 Volume usage:
    docker system df
    
    echo.
    echo 🔍 Health checks:
    docker-compose %COMPOSE_FILES% -p %PROJECT_NAME% ps
)

REM Show helpful commands
echo.
echo 💡 Helpful commands:
echo   View logs: scripts\docker-run.bat %ENV% %DB% logs
echo   Check status: scripts\docker-run.bat %ENV% %DB% status
echo   Restart: scripts\docker-run.bat %ENV% %DB% restart
echo   Stop: scripts\docker-run.bat %ENV% %DB% down

endlocal