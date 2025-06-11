# Quick setup script to build and run the Go CUDA Torus application
# This script handles the MinGW PATH issue and builds the application

Write-Host "Go CUDA Torus - Quick Native Build" -ForegroundColor Green
Write-Host "===================================" -ForegroundColor Green

# Add MinGW to PATH for this session
$mingwPaths = @(
    "C:\WinLibs\mingw64\bin",
    "C:\mingw64\bin",
    "$env:USERPROFILE\AppData\Local\Programs\mingw\bin",
    "$env:ProgramFiles\mingw64\bin",
    "${env:ProgramFiles(x86)}\mingw64\bin"
)

$foundMinGW = $false
foreach ($path in $mingwPaths) {
    if (Test-Path $path) {
        Write-Host "Found MinGW at: $path" -ForegroundColor Green
        $env:PATH = "$path;$env:PATH"
        $foundMinGW = $true
        break
    }
}

if (-not $foundMinGW) {
    Write-Host "MinGW not found in common locations. Checking registry..." -ForegroundColor Yellow
    
    # Check winget installation path
    try {
        $wingetPath = Get-ChildItem -Path "$env:LOCALAPPDATA\Microsoft\WinGet\Packages" -Directory | 
                     Where-Object { $_.Name -like "*WinLibs*" } | 
                     Select-Object -First 1
        
        if ($wingetPath) {
            $mingwBin = Join-Path $wingetPath.FullName "mingw64\bin"
            if (Test-Path $mingwBin) {
                Write-Host "Found MinGW via winget at: $mingwBin" -ForegroundColor Green
                $env:PATH = "$mingwBin;$env:PATH"
                $foundMinGW = $true
            }
        }
    } catch {
        # Continue without error
    }
}

if (-not $foundMinGW) {
    Write-Host "Error: Could not find MinGW installation" -ForegroundColor Red
    Write-Host "Please restart your PowerShell session after installing MinGW" -ForegroundColor Yellow
    Write-Host "Or manually add MinGW bin directory to your PATH" -ForegroundColor Yellow
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
    Write-Host "✗ GCC still not working" -ForegroundColor Red
    Write-Host "Try restarting PowerShell and running: gcc --version" -ForegroundColor Yellow
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
