//go:build windows
// +build windows

package main

/*
#cgo LDFLAGS: -L. -lcuda_wrapper
#include <stdlib.h>

extern int initCUDA();
extern void cleanupCUDA();
extern void updateTorusVerticesCUDA(float* vertices, int numVertices, float time, float majorRadius, float minorRadius);
extern void updateTorusColorsCUDA(float* colors, int numVertices, float time);
*/
import "C"
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var cudaAvailable bool
var cudartDLL *syscall.DLL

// Initialize CUDA subsystem
func initCUDASystem() bool {
	LogInfo("CUDA", "Starting CUDA initialization...")
	LogInfo("System", "Running on Windows platform")
	// Try to detect CUDA runtime first
	detectionResult := detectCUDASoft()
	LogInfo("CUDA", fmt.Sprintf("CUDA detection result: %v", detectionResult))

	if !detectionResult {
		LogWarning("CUDA", "CUDA detection failed, but attempting initialization anyway...")
	}

	// Try to initialize CUDA runtime
	defer func() {
		if r := recover(); r != nil {
			LogError("CUDA", fmt.Sprintf("CUDA initialization failed: %v", r))
			LogWarning("System", "Falling back to CPU mode")
			cudaAvailable = false
		}
	}()

	LogInfo("CUDA", "Attempting to call C.initCUDA()...")
	result := C.initCUDA()
	LogInfo("CUDA", fmt.Sprintf("C.initCUDA() returned: %d", result))
	cudaAvailable = (result == 1)

	if cudaAvailable {
		LogSuccess("CUDA", "CUDA runtime initialized successfully")
		LogInfo("CUDA", "NVIDIA GPU drivers: Available")
		LogInfo("CUDA", "CUDA runtime library: Found")
		LogInfo("CUDA", "GPU compute capability: Sufficient")
		LogSuccess("System", "Current Mode: CUDA Acceleration")
		return true
	} else {
		LogError("CUDA", "CUDA runtime initialization failed")
		LogWarning("System", "Current Mode: CPU Fallback")
		return false
	}
}

// Detect CUDA using multiple methods
func detectCUDASoft() bool {
	LogInfo("CUDA", "Detecting CUDA installation...")

	// Method 1: Check nvidia-smi
	if checkNvidiaSMI() {
		LogInfo("CUDA", "✓ nvidia-smi detected - GPU drivers available")
	} else {
		LogWarning("CUDA", "✗ nvidia-smi not found - no GPU drivers")
		return false
	}

	// Method 2: Check nvcc compiler
	if checkNVCC() {
		LogInfo("CUDA", "✓ nvcc compiler detected - CUDA toolkit available")
	} else {
		LogWarning("CUDA", "✗ nvcc not found - CUDA toolkit not installed")
		return false
	}

	// Method 3: Check CUDA runtime DLL
	if checkCUDARuntimeDLL() {
		LogInfo("CUDA", "✓ CUDA runtime DLL found")
	} else {
		LogWarning("CUDA", "✗ CUDA runtime DLL not accessible")
		return false
	}

	// Method 4: Check CUDA toolkit path
	if checkCUDAPath() {
		LogInfo("CUDA", "✓ CUDA toolkit path detected")
	} else {
		LogInfo("CUDA", "~ CUDA toolkit path not in environment (not critical)")
	}

	return true
}

// Check if nvidia-smi is available
func checkNvidiaSMI() bool {
	cmd := exec.Command("nvidia-smi", "--version")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}

// Check if nvcc compiler is available
func checkNVCC() bool {
	cmd := exec.Command("nvcc", "--version")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}

// Check if CUDA runtime DLL is accessible
func checkCUDARuntimeDLL() bool {
	// Try to load CUDA runtime DLL
	dllNames := []string{
		"cudart64_12.dll",
		"cudart64_11.dll",
		"cudart64_10.dll",
		"cudart.dll",
	}

	for _, dllName := range dllNames {
		dll, err := syscall.LoadDLL(dllName)
		if err == nil {
			LogInfo("CUDA", fmt.Sprintf("✓ Found CUDA runtime: %s", dllName))
			cudartDLL = dll
			return true
		}
	}

	LogDebug("CUDA", "No CUDA runtime DLL found in system")
	return false
}

// Check CUDA toolkit path
func checkCUDAPath() bool {
	cudaPath := os.Getenv("CUDA_PATH")
	if cudaPath != "" {
		LogInfo("CUDA", fmt.Sprintf("CUDA_PATH: %s", cudaPath))
		return true
	}

	// Check common installation paths
	commonPaths := []string{
		"C:\\Program Files\\NVIDIA GPU Computing Toolkit\\CUDA",
		"C:\\Program Files (x86)\\NVIDIA GPU Computing Toolkit\\CUDA",
	}

	for _, basePath := range commonPaths {
		if _, err := os.Stat(basePath); err == nil {
			// Find version subdirectories
			entries, err := os.ReadDir(basePath)
			if err == nil {
				for _, entry := range entries {
					if entry.IsDir() && strings.HasPrefix(entry.Name(), "v") {
						fullPath := filepath.Join(basePath, entry.Name())
						LogInfo("CUDA", fmt.Sprintf("Found CUDA installation: %s", fullPath))
						return true
					}
				}
			}
		}
	}

	return false
}

// Cleanup CUDA resources
func cleanupCUDASystem() {
	if cudaAvailable {
		defer func() {
			if r := recover(); r != nil {
				LogWarning("CUDA", fmt.Sprintf("CUDA cleanup warning: %v", r))
			}
		}()
		LogInfo("CUDA", "Cleaning up CUDA resources...")
		C.cleanupCUDA()
		LogInfo("CUDA", "CUDA cleanup completed")
	} else {
		LogInfo("System", "No CUDA resources to cleanup")
	}
}

// Check if CUDA is available
func isCUDAAvailable() bool {
	return cudaAvailable
}

// Update vertices using CUDA acceleration
func updateTorusVertices(vertices []float32, time float32, majorRadius, minorRadius float32) {
	if !cudaAvailable || len(vertices) == 0 {
		LogDebug("CPU", fmt.Sprintf("CUDA not available, skipping vertex update for %d vertices", len(vertices)/6))
		return
	}

	defer func() {
		if r := recover(); r != nil {
			LogError("CUDA", fmt.Sprintf("CUDA vertex update failed: %v", r))
			cudaAvailable = false
		}
	}()
	numVertices := len(vertices) / 6 // 6 components per vertex (x,y,z,r,g,b)
	// Quiet logging - only log occasionally to avoid spam
	// LogDebug("CUDA", fmt.Sprintf("Updating %d vertices using CUDA acceleration", numVertices))

	C.updateTorusVerticesCUDA(
		(*C.float)(unsafe.Pointer(&vertices[0])),
		C.int(numVertices),
		C.float(time),
		C.float(majorRadius),
		C.float(minorRadius),
	)
}

// Update colors using CUDA acceleration
func updateTorusColors(colors []float32, time float32) {
	if !cudaAvailable || len(colors) == 0 {
		LogDebug("CPU", fmt.Sprintf("CUDA not available, skipping color update for %d colors", len(colors)/3))
		return
	}

	defer func() {
		if r := recover(); r != nil {
			LogError("CUDA", fmt.Sprintf("CUDA color update failed: %v", r))
			cudaAvailable = false
		}
	}()
	numVertices := len(colors) / 3 // 3 components per vertex (r,g,b)
	// Quiet logging - only log occasionally to avoid spam
	// LogDebug("CUDA", fmt.Sprintf("Updating %d colors using CUDA acceleration", numVertices))

	C.updateTorusColorsCUDA(
		(*C.float)(unsafe.Pointer(&colors[0])),
		C.int(numVertices),
		C.float(time),
	)
}
