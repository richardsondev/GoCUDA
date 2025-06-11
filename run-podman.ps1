# PowerShell script to build and run the Go CUDA Torus application with Podman

param(
    [string]$Action = "help",
    [string]$Target = "runtime"
)

# Set Podman executable path for Windows
$PODMAN_EXE = "C:\Program Files\RedHat\Podman\podman.exe"

# Verify Podman is available
try {
    & $PODMAN_EXE --version | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Podman command failed"
    }
} catch {
    Write-Host "Error: Podman not found at $PODMAN_EXE" -ForegroundColor Red
    Write-Host "Please verify Podman Desktop is installed correctly." -ForegroundColor Yellow
    exit 1
}

Write-Host "Go CUDA Torus Application - Podman Runner" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Using Podman at: $PODMAN_EXE" -ForegroundColor Cyan

switch ($Action.ToLower()) {
    "build" {
        Write-Host "Building Docker image (target: $Target)..." -ForegroundColor Yellow
        
        if ($Target -eq "all") {
            Write-Host "Building all stages..." -ForegroundColor Cyan
            & $PODMAN_EXE build --target builder -t go-cuda-torus:builder .
            & $PODMAN_EXE build --target runtime -t go-cuda-torus:runtime .
            & $PODMAN_EXE build --target development -t go-cuda-torus:dev .
        } else {
            if ($Target -eq "runtime" -or $Target -eq "development" -or $Target -eq "builder") {
                & $PODMAN_EXE build --target $Target -t go-cuda-torus:$Target .
            } else {
                Write-Host "Invalid target: $Target. Valid targets: builder, runtime, development, all" -ForegroundColor Red
                exit 1
            }
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Build completed successfully!" -ForegroundColor Green
        } else {
            Write-Host "Build failed!" -ForegroundColor Red
            exit 1
        }
    }
    
    "run" {
        Write-Host "Building and running application (runtime container)..." -ForegroundColor Yellow
        
        # Build runtime stage
        & $PODMAN_EXE build --target runtime -t go-cuda-torus:runtime .
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Build failed!" -ForegroundColor Red
            exit 1
        }
        
        Write-Host "Starting runtime container..." -ForegroundColor Yellow
        
        & $PODMAN_EXE run --rm -it `
            --name go-cuda-torus `
            --device nvidia.com/gpu=all `
            -p 8080:8080 `
            go-cuda-torus:runtime
    }
    
    "dev" {
        Write-Host "Starting development container..." -ForegroundColor Yellow
        
        # Build development stage
        & $PODMAN_EXE build --target development -t go-cuda-torus:dev .
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Build failed!" -ForegroundColor Red
            exit 1
        }
        
        # Run in development mode with shell access and source mounting
        & $PODMAN_EXE run --rm -it `
            --name go-cuda-torus-dev `
            --device nvidia.com/gpu=all `
            -v "${PWD}:/workspace" `
            -p 8080:8080 `
            -p 2345:2345 `
            go-cuda-torus:dev
    }
    
    "build-only" {
        Write-Host "Running build-only container..." -ForegroundColor Yellow
        
        # Build builder stage
        & $PODMAN_EXE build --target builder -t go-cuda-torus:builder .
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Build failed!" -ForegroundColor Red
            exit 1
        }
        
        # Run builder to compile the application
        & $PODMAN_EXE run --rm `
            --name go-cuda-torus-builder `
            -v "${PWD}:/workspace" `
            go-cuda-torus:builder
            
        Write-Host "Build completed! Check for torus-app binary." -ForegroundColor Green
    }
    
    "compose-prod" {
        Write-Host "Starting production environment with Docker Compose..." -ForegroundColor Yellow
        & $PODMAN_EXE compose up --build torus-app
    }
    
    "compose-dev" {
        Write-Host "Starting development environment with Docker Compose..." -ForegroundColor Yellow
        & $PODMAN_EXE compose up --build torus-dev
    }
    
    "compose-build" {
        Write-Host "Running build-only with Docker Compose..." -ForegroundColor Yellow
        & $PODMAN_EXE compose up --build torus-builder
    }
    
    "clean" {
        Write-Host "Cleaning up containers and images..." -ForegroundColor Yellow
        & $PODMAN_EXE container prune -f
        & $PODMAN_EXE image prune -f
        & $PODMAN_EXE rmi go-cuda-torus:runtime -f 2>$null
        & $PODMAN_EXE rmi go-cuda-torus:dev -f 2>$null
        & $PODMAN_EXE rmi go-cuda-torus:builder -f 2>$null
        & $PODMAN_EXE volume prune -f
        Write-Host "Cleanup completed!" -ForegroundColor Green
    }
    
    "help" {
        Write-Host "Usage: .\run-podman.ps1 [action] [target]" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Available actions:" -ForegroundColor Cyan
        Write-Host "  build [target]   - Build Docker image for specific target"
        Write-Host "                     Targets: builder, runtime, development, all"
        Write-Host "  run              - Build and run production application"
        Write-Host "  dev              - Start development container with tools"
        Write-Host "  build-only       - Run build-only container to compile app"
        Write-Host "  compose-prod     - Use docker-compose for production"
        Write-Host "  compose-dev      - Use docker-compose for development"
        Write-Host "  compose-build    - Use docker-compose for build-only"
        Write-Host "  clean            - Clean up containers and images"
        Write-Host ""
        Write-Host "Examples:" -ForegroundColor Yellow
        Write-Host "  .\run-podman.ps1 build runtime      # Build production image"
        Write-Host "  .\run-podman.ps1 build development  # Build dev image"
        Write-Host "  .\run-podman.ps1 build all          # Build all stages"
        Write-Host "  .\run-podman.ps1 run                # Run production app"
        Write-Host "  .\run-podman.ps1 dev                # Start development environment"
        Write-Host "  .\run-podman.ps1 build-only         # Just compile the binary"
        Write-Host "  .\run-podman.ps1 compose-dev        # Use compose for development"
    }
    
    default {
        Write-Host "Unknown action: $Action" -ForegroundColor Red
        & $MyInvocation.MyCommand.Path "help"
    }
}
