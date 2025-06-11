#include <cuda_runtime.h>
#include <math.h>
#include <stdio.h>

#ifndef M_PI
#define M_PI 3.14159265358979323846
#endif

__global__ void torusVertexKernel(float* vertices, int numVertices, float time, float majorRadius, float minorRadius) {
    int idx = blockIdx.x * blockDim.x + threadIdx.x;
    
    if (idx < numVertices) {
        int vertexOffset = idx * 6;
        
        int segments = (int)sqrt((float)numVertices);
        int rings = segments;
        
        int u = idx % segments;
        int v = idx / segments;
        
        if (v >= rings) return;
        
        float theta = 2.0f * M_PI * u / segments;
        float phi = 2.0f * M_PI * v / rings;
        
        theta += time * 0.5f;
        phi += time * 0.3f;
        
        float cosTheta = cosf(theta);
        float sinTheta = sinf(theta);
        float cosPhi = cosf(phi);
        float sinPhi = sinf(phi);
        
        float x = (majorRadius + minorRadius * cosPhi) * cosTheta;
        float y = (majorRadius + minorRadius * cosPhi) * sinTheta;
        float z = minorRadius * sinPhi;
        
        vertices[vertexOffset + 0] = x;
        vertices[vertexOffset + 1] = y;
        vertices[vertexOffset + 2] = z;
    }
}

__global__ void torusColorKernel(float* colors, int numVertices, float time) {
    int idx = blockIdx.x * blockDim.x + threadIdx.x;
    
    if (idx < numVertices) {
        int colorOffset = idx * 3;
        
        float phase = time * 2.0f + idx * 0.1f;
        
        colors[colorOffset + 0] = 0.5f + 0.5f * sinf(phase);
        colors[colorOffset + 1] = 0.5f + 0.5f * sinf(phase + 2.1f);
        colors[colorOffset + 2] = 0.5f + 0.5f * sinf(phase + 4.2f);
    }
}

static float* d_vertices = nullptr;
static float* d_colors = nullptr;
static int maxVertices = 0;
static bool cudaInitialized = false;

extern "C" {

int initCUDA() {
    if (cudaInitialized) return 1;
    
    cudaError_t error = cudaSetDevice(0);
    if (error != cudaSuccess) {
        printf("CUDA Error: %s\n", cudaGetErrorString(error));
        return 0;
    }
    
    cudaInitialized = true;
    return 1;
}

void cleanupCUDA() {
    if (!cudaInitialized) return;
    
    if (d_vertices) { cudaFree(d_vertices); d_vertices = nullptr; }
    if (d_colors) { cudaFree(d_colors); d_colors = nullptr; }
    
    maxVertices = 0;
    cudaInitialized = false;
    cudaDeviceReset();
}

int ensureDeviceMemory(int numVertices) {
    if (!cudaInitialized) return 0;
    
    if (numVertices > maxVertices) {
        if (d_vertices) cudaFree(d_vertices);
        if (d_colors) cudaFree(d_colors);
        
        size_t vertexSize = numVertices * 6 * sizeof(float);
        size_t colorSize = numVertices * 3 * sizeof(float);
        
        if (cudaMalloc(&d_vertices, vertexSize) != cudaSuccess) return 0;
        if (cudaMalloc(&d_colors, colorSize) != cudaSuccess) return 0;
        
        maxVertices = numVertices;
    }
    
    return 1;
}

void updateTorusVerticesCUDA(float* vertices, int numVertices, float time, float majorRadius, float minorRadius) {
    if (!cudaInitialized || !ensureDeviceMemory(numVertices)) return;
    
    size_t vertexSize = numVertices * 6 * sizeof(float);
    if (cudaMemcpy(d_vertices, vertices, vertexSize, cudaMemcpyHostToDevice) != cudaSuccess) return;
    
    int blockSize = 256;
    int gridSize = (numVertices + blockSize - 1) / blockSize;
    
    torusVertexKernel<<<gridSize, blockSize>>>(d_vertices, numVertices, time, majorRadius, minorRadius);
    
    if (cudaDeviceSynchronize() != cudaSuccess) return;
    
    cudaMemcpy(vertices, d_vertices, vertexSize, cudaMemcpyDeviceToHost);
}

void updateTorusColorsCUDA(float* colors, int numVertices, float time) {
    if (!cudaInitialized || !ensureDeviceMemory(numVertices)) return;
    
    size_t colorSize = numVertices * 3 * sizeof(float);
    if (cudaMemcpy(d_colors, colors, colorSize, cudaMemcpyHostToDevice) != cudaSuccess) return;
    
    int blockSize = 256;
    int gridSize = (numVertices + blockSize - 1) / blockSize;
    
    torusColorKernel<<<gridSize, blockSize>>>(d_colors, numVertices, time);
    
    if (cudaDeviceSynchronize() != cudaSuccess) return;
    
    cudaMemcpy(colors, d_colors, colorSize, cudaMemcpyDeviceToHost);
}

} // extern "C"
