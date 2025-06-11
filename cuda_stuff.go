//go:build !windows && cgo
// +build !windows,cgo

package main

/*
#cgo CFLAGS: -I/usr/local/cuda/include -I.
#cgo LDFLAGS: -L/usr/local/cuda/lib64 -lcudart -L. -ltorus_kernel
#include <stdlib.h>

// Declare functions that may or may not be available
extern void updateTorusVerticesCUDA(float* vertices, int numVertices, float time, float majorRadius, float minorRadius);
extern void updateTorusColorsCUDA(float* colors, int numVertices, float time);
extern int initCUDA();
extern void cleanupCUDA();
*/
import "C"
import (
	"fmt"
	"unsafe"
)

var cudaAvailable bool

// Initialize CUDA subsystem
func initCUDASystem() bool {
	// Try to initialize CUDA - if it fails, fall back to CPU
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("CUDA initialization failed, using CPU fallback:", r)
			cudaAvailable = false
		}
	}()

	result := C.initCUDA()
	cudaAvailable = (result == 1)
	if cudaAvailable {
		fmt.Println("CUDA initialized successfully")
	} else {
		fmt.Println("CUDA not available, using CPU computations")
	}
	return cudaAvailable
}

// Cleanup CUDA resources
func cleanupCUDASystem() {
	if cudaAvailable {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("CUDA cleanup warning:", r)
			}
		}()
		C.cleanupCUDA()
		fmt.Println("CUDA cleanup completed")
	}
}

// Update torus vertices using CUDA
func updateTorusVerticesCUDA(vertices []float32, time float32, majorRadius, minorRadius float32) {
	if !cudaAvailable || len(vertices) == 0 {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("CUDA vertex update failed:", r)
			cudaAvailable = false
		}
	}()

	numVertices := len(vertices) / 6 // 6 components per vertex (x,y,z,r,g,b)

	C.updateTorusVerticesCUDA(
		(*C.float)(unsafe.Pointer(&vertices[0])),
		C.int(numVertices),
		C.float(time),
		C.float(majorRadius),
		C.float(minorRadius),
	)
}

// Update torus colors using CUDA
func updateTorusColorsCUDA(vertices []float32, time float32) {
	if !cudaAvailable || len(vertices) == 0 {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("CUDA color update failed:", r)
			cudaAvailable = false
		}
	}()

	numVertices := len(vertices) / 6 // 6 components per vertex (x,y,z,r,g,b)

	C.updateTorusColorsCUDA(
		(*C.float)(unsafe.Pointer(&vertices[0])),
		C.int(numVertices),
		C.float(time),
	)
}

// Check if CUDA is available
func isCUDAAvailable() bool {
	return cudaAvailable
}
