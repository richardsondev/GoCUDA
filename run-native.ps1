# PowerShell script to build and run the Go CUDA Torus application natively on Windows
# This script handles dependency installation, building, and running the application

param(
    [string]$Action = "help",
    [switch]$SkipDeps = $false,
    [switch]$Force = $false
)

# Colors for output
$Green = "Green"
$Red = "Red"
$Yellow = "Yellow"
$Cyan = "Cyan"

Write-Host "Go CUDA Torus Application - Native Windows Runner" -ForegroundColor $Green
Write-Host "=================================================" -ForegroundColor $Green

# Check if Go is installed
function Test-GoInstallation {
    try {
        $goVersion = go version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ Go found: $goVersion" -ForegroundColor $Green
            return $true
        }
    } catch {
        # Go command not found
    }
    
    Write-Host "✗ Go not found in PATH" -ForegroundColor $Red
    Write-Host "Please install Go from: https://golang.org/dl/" -ForegroundColor $Yellow
    Write-Host "Ensure Go is added to your PATH environment variable" -ForegroundColor $Yellow
    return $false
}

# Check if GCC/MinGW is available for CGO
function Test-CGOCompiler {
    try {
        $gccVersion = gcc --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ GCC found for CGO compilation" -ForegroundColor $Green
            return $true
        }
    } catch {
        # GCC not found
    }
    
    Write-Host "✗ GCC not found - CGO compilation required for OpenGL" -ForegroundColor $Red
    Write-Host "Installing MinGW-w64 automatically..." -ForegroundColor $Yellow
    
    # Try to install MinGW via different methods
    if (Install-MinGW) {
        return $true
    }
    
    Write-Host "Please install MinGW-w64 manually:" -ForegroundColor $Yellow
    Write-Host "1. Download from: https://www.mingw-w64.org/downloads/" -ForegroundColor $Cyan
    Write-Host "2. Or use winget: winget install mingw" -ForegroundColor $Cyan
    Write-Host "3. Or use chocolatey: choco install mingw" -ForegroundColor $Cyan
    return $false
}

# Install MinGW-w64 automatically
function Install-MinGW {
    Write-Host "Attempting automatic MinGW installation..." -ForegroundColor $Yellow
    
    # Try winget first
    try {
        winget --version | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Installing MinGW via winget..." -ForegroundColor $Cyan
            winget install --id=mingw --accept-package-agreements --accept-source-agreements
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✓ MinGW installed via winget" -ForegroundColor $Green
                Write-Host "Please restart your terminal and try again" -ForegroundColor $Yellow
                return $true
            }
        }
    } catch {
        # winget not available
    }
    
    # Try chocolatey
    try {
        choco --version | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Installing MinGW via chocolatey..." -ForegroundColor $Cyan
            choco install mingw -y
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✓ MinGW installed via chocolatey" -ForegroundColor $Green
                Write-Host "Please restart your terminal and try again" -ForegroundColor $Yellow
                return $true
            }
        }
    } catch {
        # chocolatey not available
    }
    
    # Try scoop
    try {
        scoop --version | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Installing MinGW via scoop..." -ForegroundColor $Cyan
            scoop install mingw
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✓ MinGW installed via scoop" -ForegroundColor $Green
                Write-Host "Please restart your terminal and try again" -ForegroundColor $Yellow
                return $true
            }
        }
    } catch {
        # scoop not available
    }
    
    return $false
}

# Check if Git is available
function Test-GitInstallation {
    try {
        $gitVersion = git --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ Git found: $gitVersion" -ForegroundColor $Green
            return $true
        }
    } catch {
        # Git not found
    }
    
    Write-Host "✗ Git not found - may affect Go module downloads" -ForegroundColor $Yellow
    return $false
}

# Install Go dependencies
function Install-Dependencies {
    Write-Host "Installing Go dependencies..." -ForegroundColor $Cyan
    
    # Enable CGO for OpenGL libraries
    $env:CGO_ENABLED = "1"
    
    # Download dependencies
    Write-Host "Running: go mod download" -ForegroundColor $Yellow
    go mod download
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ Failed to download Go modules" -ForegroundColor $Red
        return $false
    }
    
    # Verify dependencies
    Write-Host "Running: go mod verify" -ForegroundColor $Yellow
    go mod verify
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ Module verification failed" -ForegroundColor $Red
        return $false
    }
    
    # Tidy up modules
    Write-Host "Running: go mod tidy" -ForegroundColor $Yellow
    go mod tidy
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ Module tidy failed" -ForegroundColor $Red
        return $false
    }
    
    Write-Host "✓ Dependencies installed successfully" -ForegroundColor $Green
    return $true
}

# Build the application
function Build-Application {
    param([bool]$Force = $false)
    
    $binaryName = "torus-app.exe"
    
    # Check if binary exists and is newer than source files
    if (-not $Force -and (Test-Path $binaryName)) {
        $binaryTime = (Get-Item $binaryName).LastWriteTime
        $sourceFiles = Get-ChildItem -Filter "*.go"
        $newestSource = ($sourceFiles | Sort-Object LastWriteTime -Descending | Select-Object -First 1).LastWriteTime
        
        if ($binaryTime -gt $newestSource) {
            Write-Host "✓ Binary is up to date, skipping build" -ForegroundColor $Green
            return $true
        }
    }
    
    Write-Host "Building application..." -ForegroundColor $Cyan
    
    # Set build environment
    $env:CGO_ENABLED = "1"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    
    # Build command
    Write-Host "Running: go build -o $binaryName ." -ForegroundColor $Yellow
    go build -o $binaryName .
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Build completed successfully: $binaryName" -ForegroundColor $Green
        return $true
    } else {
        Write-Host "✗ Build failed" -ForegroundColor $Red
        return $false
    }
}

# Run the application
function Start-Application {
    $binaryName = "torus-app.exe"
    
    if (-not (Test-Path $binaryName)) {
        Write-Host "✗ Binary not found: $binaryName" -ForegroundColor $Red
        return $false
    }
    
    Write-Host "Starting application..." -ForegroundColor $Cyan
    Write-Host "Web interface will be available at: http://localhost:8080" -ForegroundColor $Yellow
    Write-Host "Press Ctrl+C to stop the application" -ForegroundColor $Yellow
    Write-Host ""
    
    # Run the application
    & ".\$binaryName"
}

# Clean build artifacts
function Clear-BuildArtifacts {
    Write-Host "Cleaning build artifacts..." -ForegroundColor $Cyan
    
    $filesToRemove = @("torus-app.exe", "*.o", "*.a", "*.lib")
    
    foreach ($pattern in $filesToRemove) {
        $files = Get-ChildItem -Filter $pattern -ErrorAction SilentlyContinue
        foreach ($file in $files) {
            Remove-Item $file.FullName -Force
            Write-Host "Removed: $($file.Name)" -ForegroundColor $Yellow
        }
    }
    
    Write-Host "✓ Cleanup completed" -ForegroundColor $Green
}

# Main script logic
switch ($Action.ToLower()) {
    "check" {
        Write-Host "Checking system requirements..." -ForegroundColor $Cyan
        $goOk = Test-GoInstallation
        $gccOk = Test-CGOCompiler
        $gitOk = Test-GitInstallation
        
        if ($goOk) {
            Write-Host "✓ System ready for Go development" -ForegroundColor $Green
        } else {
            Write-Host "✗ System requirements not met" -ForegroundColor $Red
        }
    }
    
    "deps" {
        if (-not (Test-GoInstallation)) { exit 1 }
        if (-not (Install-Dependencies)) { exit 1 }
    }
    
    "build" {
        if (-not (Test-GoInstallation)) { exit 1 }
        if (-not $SkipDeps -and -not (Install-Dependencies)) { exit 1 }
        if (-not (Build-Application -Force $Force)) { exit 1 }
    }
    
    "run" {
        if (-not (Test-GoInstallation)) { exit 1 }
        if (-not $SkipDeps -and -not (Install-Dependencies)) { exit 1 }
        if (-not (Build-Application -Force $Force)) { exit 1 }
        Start-Application
    }
    
    "clean" {
        Clear-BuildArtifacts
    }
    
    "rebuild" {
        Clear-BuildArtifacts
        if (-not (Test-GoInstallation)) { exit 1 }
        if (-not (Install-Dependencies)) { exit 1 }
        if (-not (Build-Application -Force $true)) { exit 1 }
    }
    
    "help" {
        Write-Host "Usage: .\run-native.ps1 [action] [options]" -ForegroundColor $Cyan
        Write-Host ""
        Write-Host "Actions:" -ForegroundColor $Cyan
        Write-Host "  check    - Check system requirements (Go, GCC, Git)" -ForegroundColor $Yellow
        Write-Host "  deps     - Install Go dependencies only" -ForegroundColor $Yellow
        Write-Host "  build    - Install deps and build application" -ForegroundColor $Yellow
        Write-Host "  run      - Install deps, build, and run application" -ForegroundColor $Yellow
        Write-Host "  clean    - Remove build artifacts" -ForegroundColor $Yellow
        Write-Host "  rebuild  - Clean, install deps, and rebuild" -ForegroundColor $Yellow
        Write-Host "  help     - Show this help message" -ForegroundColor $Yellow
        Write-Host ""
        Write-Host "Options:" -ForegroundColor $Cyan
        Write-Host "  -SkipDeps    - Skip dependency installation" -ForegroundColor $Yellow
        Write-Host "  -Force       - Force rebuild even if binary is up to date" -ForegroundColor $Yellow
        Write-Host ""
        Write-Host "Examples:" -ForegroundColor $Green
        Write-Host "  .\run-native.ps1 check                    # Check system requirements" -ForegroundColor $Green
        Write-Host "  .\run-native.ps1 run                      # Build and run application" -ForegroundColor $Green
        Write-Host "  .\run-native.ps1 build -Force             # Force rebuild" -ForegroundColor $Green
        Write-Host "  .\run-native.ps1 run -SkipDeps            # Run without updating deps" -ForegroundColor $Green
        Write-Host ""
        Write-Host "Requirements:" -ForegroundColor $Cyan
        Write-Host "  - Go 1.21+ (https://golang.org/dl/)" -ForegroundColor $Yellow
        Write-Host "  - GCC/MinGW for CGO (optional but recommended)" -ForegroundColor $Yellow
        Write-Host "  - Git for Go module downloads" -ForegroundColor $Yellow
        Write-Host "  - OpenGL drivers (usually pre-installed)" -ForegroundColor $Yellow
    }
    
    default {
        Write-Host "Unknown action: $Action" -ForegroundColor $Red
        Write-Host "Use '.\run-native.ps1 help' for usage information" -ForegroundColor $Yellow
        exit 1
    }
}
