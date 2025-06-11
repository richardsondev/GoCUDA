# Running Go CUDA Torus Natively on Windows

This guide explains how to run the Go CUDA Torus application directly on your Windows machine without Docker/Podman.

## Quick Start

1. **Check system requirements:**
   ```powershell
   .\run-native.ps1 check
   ```

2. **Build and run the application:**
   ```powershell
   .\run-native.ps1 run
   ```

3. **Open your browser to:**
   ```
   http://localhost:8080
   ```

## System Requirements

### Required
- **Go 1.21+** - Download from [golang.org/dl](https://golang.org/dl/)
- **OpenGL drivers** - Usually pre-installed with graphics drivers
- **Internet connection** - For downloading Go modules

### Recommended
- **GCC/MinGW** - For CGO compilation support
  - Install via: `winget install mingw` (if you have winget)
  - Or download from [mingw-w64.org](https://mingw-w64.org/)
- **Git** - For Go module downloads

## Installation Options

### Option 1: PowerShell Script (Recommended)
```powershell
# Check system requirements
.\run-native.ps1 check

# Build and run (handles dependencies automatically)
.\run-native.ps1 run

# Other useful commands
.\run-native.ps1 build          # Build only
.\run-native.ps1 clean          # Clean build artifacts
.\run-native.ps1 rebuild        # Clean and rebuild
.\run-native.ps1 help           # Show all options
```

### Option 2: Batch Script (Simple)
```cmd
# Build and run
run-native.bat

# Other commands
run-native.bat check     # Check requirements
run-native.bat build     # Build only
run-native.bat clean     # Clean artifacts
run-native.bat help      # Show help
```

### Option 3: Manual Commands
```powershell
# Install dependencies
go mod download
go mod tidy

# Build application
$env:CGO_ENABLED = "1"
go build -o torus-app.exe .

# Run application
.\torus-app.exe
```

## Features

### Web Interface
- **URL:** http://localhost:8080
- **Real-time parameter control:** Adjust torus geometry and animation
- **API endpoints:**
  - `GET /api/status` - Get current parameters and CUDA status
  - `POST /api/params` - Update parameters in real-time

### Torus Parameters
- **Major Radius** - Size of the main torus ring
- **Minor Radius** - Thickness of the torus tube
- **Major Segments** - Ring detail (higher = smoother)
- **Minor Segments** - Tube detail (higher = smoother)
- **Rotation Speed X/Y** - Animation speeds

## CUDA Support

### With NVIDIA GPU
If you have:
- NVIDIA GPU with CUDA support
- CUDA Toolkit installed
- Proper NVIDIA drivers

The application will automatically use GPU acceleration for vertex and color computations.

### CPU Fallback
Without CUDA support, the application automatically falls back to CPU computations. Performance is still excellent for real-time visualization.

## Troubleshooting

### Go Not Found
```
Error: Go not found in PATH
```
**Solution:** Install Go from [golang.org/dl](https://golang.org/dl/) and ensure it's added to your PATH.

### CGO Compilation Errors
```
gcc: command not found
```
**Solution:** Install MinGW-w64 or TDM-GCC for CGO support:
- `winget install mingw` (if available)
- Or download from [mingw-w64.org](https://mingw-w64.org/)

### OpenGL Errors
```
Failed to initialize OpenGL
```
**Solution:** Update your graphics drivers. Most modern Windows systems have OpenGL support.

### Port Already in Use
```
Error: listen tcp :8080: bind: address already in use
```
**Solution:** Stop other applications using port 8080, or modify the port in the source code.

### Module Download Failures
```
Error: Failed to download Go modules
```
**Solutions:**
1. Check internet connection
2. Configure Go proxy if behind corporate firewall:
   ```powershell
   $env:GOPROXY = "https://proxy.golang.org,direct"
   ```
3. Disable Go modules verification temporarily:
   ```powershell
   $env:GOSUMDB = "off"
   ```

## Performance Tips

### For Best Performance
1. **Enable hardware acceleration:** Ensure graphics drivers are up to date
2. **Use higher segment counts:** For smoother torus geometry (may impact performance)
3. **CUDA acceleration:** Install CUDA toolkit if you have an NVIDIA GPU

### For Slower Systems
1. **Reduce segment counts:** Lower values for better performance
2. **Slower rotation speeds:** Reduce animation load
3. **Close other applications:** Free up system resources

## Building for Distribution

### Create standalone executable:
```powershell
# Build optimized binary
$env:CGO_ENABLED = "1"
go build -ldflags="-s -w" -o torus-app.exe .
```

### Include dependencies:
The Go binary includes all dependencies. Just distribute the `torus-app.exe` file.

## File Structure

```
GoCUDA/
├── main.go              # Main application with OpenGL rendering
├── web_interface.go     # Web server and HTML interface
├── math_utils.go        # Math utilities for 3D transformations
├── shaders.go          # OpenGL vertex and fragment shaders
├── cuda_stuff.go       # CUDA integration and CGO bindings
├── torus_kernel.cu     # CUDA kernels for GPU acceleration
├── torus_kernel.h      # C interface for CUDA
├── dummy_cuda.c        # Fallback for systems without CUDA
├── run-native.ps1      # PowerShell runner script
├── run-native.bat      # Batch runner script
├── go.mod              # Go module dependencies
└── README-native.md    # This file
```

## Advanced Usage

### Custom Build Options
```powershell
# Build with debug info
go build -o torus-app.exe .

# Build optimized for release
go build -ldflags="-s -w" -o torus-app.exe .

# Build with race detection (development)
go build -race -o torus-app.exe .
```

### Environment Variables
```powershell
# Force CPU mode (disable CUDA)
$env:FORCE_CPU = "1"

# Enable verbose logging
$env:DEBUG = "1"

# Custom port
$env:PORT = "9090"
```

## Support

For issues or questions:
1. Check this README for common solutions
2. Verify system requirements are met
3. Try rebuilding: `.\run-native.ps1 rebuild`
4. Check Go and driver installations
