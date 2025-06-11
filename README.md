# Go CUDA Torus Application

A rotating toroidal (donut) shape made of small cubes with a pulsating green glow effect, implemented in Go with OpenGL and CUDA integration.

## Prerequisites

- [Podman Desktop](https://podman-desktop.io/) for Windows
- NVIDIA GPU with CUDA support
- NVIDIA Container Toolkit (for GPU access in containers)

## Quick Start

The project uses a multi-stage Docker setup with separate containers for building and running:

- **Builder Stage**: Contains Go compiler, GCC, and development tools
- **Runtime Stage**: Lightweight container with only runtime dependencies  
- **Development Stage**: Full development environment with tools and source mounting

### Option 1: Using PowerShell Script (Recommended)

```powershell
# Start development environment (recommended for coding)
.\run-podman.ps1 dev

# Build and run production application
.\run-podman.ps1 run

# Build specific stages
.\run-podman.ps1 build development    # Build dev image
.\run-podman.ps1 build runtime        # Build production image
.\run-podman.ps1 build all            # Build all stages

# Just compile the binary (CI/CD style)
.\run-podman.ps1 build-only

# Use Docker Compose
.\run-podman.ps1 compose-dev          # Development with compose
.\run-podman.ps1 compose-prod         # Production with compose

# Clean up
.\run-podman.ps1 clean
```

### Option 2: Manual Podman Commands

```bash
# Build specific stages
podman build --target development -t go-cuda-torus:dev .
podman build --target runtime -t go-cuda-torus:runtime .
podman build --target builder -t go-cuda-torus:builder .

# Run development container
podman run --rm -it \
  --name go-cuda-torus-dev \
  --device nvidia.com/gpu=all \
  -v "$PWD:/workspace" \
  -p 8080:8080 -p 2345:2345 \
  go-cuda-torus:dev

# Run production container
podman run --rm -it \
  --name go-cuda-torus \
  --device nvidia.com/gpu=all \
  -p 8080:8080 \
  go-cuda-torus:runtime

# Build-only (for CI/CD)
podman run --rm \
  --name go-cuda-torus-builder \
  -v "$PWD:/workspace" \
  go-cuda-torus:builder
```

### Option 3: Docker Compose

```bash
# Development environment
podman-compose up --build torus-dev

# Production environment  
podman-compose up --build torus-app

# Build-only
podman-compose up --build torus-builder
```

## Development

The multi-stage container setup provides:

### Builder Stage
- Go 1.22.10 with CGO enabled
- GCC/G++ compilers for CGO and CUDA
- OpenGL development libraries (Mesa, GLFW, GLEW)
- CUDA 12.2 development toolkit
- Git and build tools

### Runtime Stage (Production)
- Minimal Ubuntu base with CUDA runtime
- Only runtime OpenGL libraries
- Pre-built application binary
- X11/Xvfb for display support
- ~50% smaller than development image

### Development Stage  
- Everything from builder stage
- Additional development tools (vim, nano, gdb, valgrind)
- Source code mounting for live development
- Debugger support (port 2345 for Delve)
- Go module caching for faster rebuilds

### Container Usage

**Development Workflow:**
```bash
# Start development container
.\run-podman.ps1 dev

# Inside container - live development
go build -o torus-app .
/usr/local/bin/run-with-display.sh ./torus-app

# Or rebuild and test quickly
go run .
```

**Production Deployment:**
```bash
# Build optimized runtime image
.\run-podman.ps1 build runtime

# Run production container
.\run-podman.ps1 run
```

**CI/CD Pipeline:**
```bash
# Build-only for testing/artifacts
.\run-podman.ps1 build-only
```

## Project Structure

- `main.go` - Main application with OpenGL setup and render loop
- `shaders.go` - Vertex and fragment shader sources
- `math_utils.go` - Vector/matrix math library
- `cuda_stuff.go` - CUDA integration (placeholder)
- `Dockerfile` - Container definition
- `docker-compose.yml` - Multi-service setup
- `run-podman.ps1` - PowerShell runner script

## Current Status

- ✅ Project structure and Go modules setup
- ✅ Docker containerization with all dependencies
- ✅ Basic OpenGL rendering pipeline (single cube)
- 🔄 Torus geometry generation (pending)
- 🔄 CUDA kernels for pulsating glow effect (pending)
- 🔄 CUDA-OpenGL interoperability (pending)

## Troubleshooting

### GPU Access Issues

Ensure NVIDIA Container Toolkit is installed and configured:
- [NVIDIA Container Toolkit Installation](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html)

### Display Issues

The container uses Xvfb (virtual framebuffer) for headless rendering. For actual display output, you may need to configure X11 forwarding or use VNC.

## Next Steps

1. Implement torus geometry generation
2. Add CUDA kernels for animation effects
3. Set up CUDA-OpenGL interoperability
4. Add web interface for parameter control (port 8080)
