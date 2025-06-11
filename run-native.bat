@echo off
REM Batch script to build and run the Go CUDA Torus application natively on Windows
REM Simple alternative to the PowerShell script

setlocal enabledelayedexpansion

echo Go CUDA Torus Application - Native Windows Runner (Batch)
echo =========================================================

REM Check if Go is installed
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go not found in PATH
    echo Please install Go from: https://golang.org/dl/
    echo Ensure Go is added to your PATH environment variable
    pause
    exit /b 1
)

REM Get Go version
for /f "tokens=*" %%i in ('go version 2^>nul') do set GO_VERSION=%%i
echo Found Go: %GO_VERSION%

REM Set CGO environment
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

REM Parse command line arguments
set ACTION=%1
if "%ACTION%"=="" set ACTION=run

if /i "%ACTION%"=="check" goto :check
if /i "%ACTION%"=="deps" goto :deps
if /i "%ACTION%"=="build" goto :build
if /i "%ACTION%"=="run" goto :run
if /i "%ACTION%"=="clean" goto :clean
if /i "%ACTION%"=="help" goto :help

echo Unknown action: %ACTION%
goto :help

:check
echo Checking system requirements...
where go >nul 2>&1 && echo [OK] Go found || echo [FAIL] Go not found
where gcc >nul 2>&1 && echo [OK] GCC found || echo [WARN] GCC not found - CGO may fail
where git >nul 2>&1 && echo [OK] Git found || echo [WARN] Git not found
goto :end

:deps
echo Installing Go dependencies...
echo Running: go mod download
go mod download
if %errorlevel% neq 0 (
    echo Error: Failed to download Go modules
    pause
    exit /b 1
)

echo Running: go mod tidy
go mod tidy
if %errorlevel% neq 0 (
    echo Error: Module tidy failed
    pause
    exit /b 1
)

echo Dependencies installed successfully
goto :end

:build
echo Installing dependencies and building...
call :deps
if %errorlevel% neq 0 exit /b 1

echo Building application...
echo Running: go build -o torus-app.exe .
go build -o torus-app.exe .
if %errorlevel% neq 0 (
    echo Error: Build failed
    pause
    exit /b 1
)

echo Build completed successfully: torus-app.exe
goto :end

:run
echo Building and running application...
call :build
if %errorlevel% neq 0 exit /b 1

echo.
echo Starting application...
echo Web interface will be available at: http://localhost:8080
echo Press Ctrl+C to stop the application
echo.

REM Run the application
torus-app.exe
goto :end

:clean
echo Cleaning build artifacts...
if exist torus-app.exe del /f torus-app.exe && echo Removed: torus-app.exe
if exist *.o del /f *.o 2>nul && echo Removed: *.o files
if exist *.a del /f *.a 2>nul && echo Removed: *.a files
if exist *.lib del /f *.lib 2>nul && echo Removed: *.lib files
echo Cleanup completed
goto :end

:help
echo Usage: run-native.bat [action]
echo.
echo Actions:
echo   check    - Check system requirements (Go, GCC, Git)
echo   deps     - Install Go dependencies only
echo   build    - Install deps and build application
echo   run      - Install deps, build, and run application (default)
echo   clean    - Remove build artifacts
echo   help     - Show this help message
echo.
echo Examples:
echo   run-native.bat           # Build and run application
echo   run-native.bat check     # Check system requirements
echo   run-native.bat build     # Build only
echo   run-native.bat clean     # Clean build files
echo.
echo Requirements:
echo   - Go 1.21+ (https://golang.org/dl/)
echo   - GCC/MinGW for CGO (optional but recommended)
echo   - Git for Go module downloads
echo   - OpenGL drivers (usually pre-installed)
goto :end

:end
if not "%ACTION%"=="run" pause
