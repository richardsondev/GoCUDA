# Quick setup script to build and run the Go CUDA Torus application
# This script handles the MinGW PATH issue and builds the application

Write-Host "Go CUDA Torus - Quick Native Build" -ForegroundColor Green
Write-Host "===================================" -ForegroundColor Green

# Add known MinGW path for this system
$mingwPath = "C:\Users\Billy\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.MCF.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin"

if (Test-Path $mingwPath) {
    Write-Host "Found MinGW at: $mingwPath" -ForegroundColor Green
    $env:PATH = "$mingwPath;$env:PATH"
} else {
    Write-Host "Error: MinGW not found at expected location" -ForegroundColor Red
    Write-Host "Please check MinGW installation" -ForegroundColor Yellow
    exit 1
}

# Test GCC
try {
    $gccVersion = gcc --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ GCC found and working" -ForegroundColor Green
    } else {
        throw "GCC test failed"
    }
} catch {
    Write-Host "✗ GCC not working" -ForegroundColor Red
    exit 1
}

# Check Go
try {
    $goVersion = go version
    Write-Host "✓ Go found: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Go not found" -ForegroundColor Red
    exit 1
}

# Set CGO environment
$env:CGO_ENABLED = "1"
Write-Host "✓ CGO enabled" -ForegroundColor Green

# Install dependencies
Write-Host "Installing Go dependencies..." -ForegroundColor Cyan
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to download dependencies" -ForegroundColor Red
    exit 1
}

go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to tidy modules" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Dependencies installed" -ForegroundColor Green

# Build application
Write-Host "Building application..." -ForegroundColor Cyan
go build -o torus-app.exe .

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Build successful!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Starting application..." -ForegroundColor Cyan
    Write-Host "Web interface will be available at: http://localhost:8080" -ForegroundColor Yellow
    Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
    Write-Host ""
    
    # Run the application
    .\torus-app.exe
} else {
    Write-Host "✗ Build failed" -ForegroundColor Red
    Write-Host "Check the error messages above" -ForegroundColor Yellow
}
