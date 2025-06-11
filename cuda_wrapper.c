#include <windows.h>
#include <stdio.h>

// Function pointer typedefs for CUDA functions
typedef int (*InitCUDAFunc)();
typedef void (*CleanupCUDAFunc)();
typedef void (*UpdateVerticesFunc)(float* vertices, int numVertices, float time, float majorRadius, float minorRadius);
typedef void (*UpdateColorsFunc)(float* colors, int numVertices, float time);

// Global variables
static HMODULE cudaDLL = NULL;
static InitCUDAFunc initCUDAPtr = NULL;
static CleanupCUDAFunc cleanupCUDAPtr = NULL;
static UpdateVerticesFunc updateVerticesPtr = NULL;
static UpdateColorsFunc updateColorsPtr = NULL;
static int dllLoaded = 0;

// Load CUDA DLL and get function pointers
static int loadCUDADLL() {
    if (dllLoaded) return 1;
    
    printf("Loading CUDA DLL...\n");
    cudaDLL = LoadLibrary("cuda_export.dll");
    if (!cudaDLL) {
        printf("Failed to load cuda_export.dll\n");
        return 0;
    }
    
    initCUDAPtr = (InitCUDAFunc)GetProcAddress(cudaDLL, "initCUDA");
    cleanupCUDAPtr = (CleanupCUDAFunc)GetProcAddress(cudaDLL, "cleanupCUDA");
    updateVerticesPtr = (UpdateVerticesFunc)GetProcAddress(cudaDLL, "updateTorusVerticesCUDA");
    updateColorsPtr = (UpdateColorsFunc)GetProcAddress(cudaDLL, "updateTorusColorsCUDA");
    
    if (!initCUDAPtr || !cleanupCUDAPtr || !updateVerticesPtr || !updateColorsPtr) {
        printf("Failed to get function pointers from CUDA DLL\n");
        FreeLibrary(cudaDLL);
        cudaDLL = NULL;
        return 0;
    }
    
    printf("CUDA DLL loaded successfully\n");
    dllLoaded = 1;
    return 1;
}

// Wrapper functions
int initCUDA() {
    printf("initCUDA wrapper called\n");
    if (!loadCUDADLL()) {
        printf("Failed to load CUDA DLL in initCUDA\n");
        return 0;
    }
    
    printf("Calling CUDA initCUDA function\n");
    int result = initCUDAPtr();
    printf("CUDA initCUDA returned: %d\n", result);
    return result;
}

void cleanupCUDA() {
    printf("cleanupCUDA wrapper called\n");
    if (dllLoaded && cleanupCUDAPtr) {
        cleanupCUDAPtr();
    }
    
    if (cudaDLL) {
        FreeLibrary(cudaDLL);
        cudaDLL = NULL;
        dllLoaded = 0;
    }
    printf("CUDA cleanup completed\n");
}

void updateTorusVerticesCUDA(float* vertices, int numVertices, float time, float majorRadius, float minorRadius) {
    // printf("updateTorusVerticesCUDA wrapper called with %d vertices\n", numVertices);
    if (dllLoaded && updateVerticesPtr) {
        updateVerticesPtr(vertices, numVertices, time, majorRadius, minorRadius);
        // printf("CUDA vertex update completed\n");
    } else {
        printf("CUDA DLL not loaded, skipping vertex update\n");
    }
}

void updateTorusColorsCUDA(float* colors, int numVertices, float time) {
    // printf("updateTorusColorsCUDA wrapper called with %d vertices\n", numVertices);
    if (dllLoaded && updateColorsPtr) {
        updateColorsPtr(colors, numVertices, time);
        // printf("CUDA color update completed\n");
    } else {
        printf("CUDA DLL not loaded, skipping color update\n");
    }
}
