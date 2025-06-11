#include <cuda_runtime.h>
#include <math.h>

#ifndef M_PI
#define M_PI 3.14159265358979323846
#endif

extern "C" {
    void updateTorusVerticesCUDA(float* vertices, int numVertices, float time, float majorRadius, float minorRadius);
    void updateTorusColorsCUDA(float* colors, int numVertices, float time);
    int initCUDA();
    void cleanupCUDA();
}

// CUDA kernel for updating torus vertex positions
__global__ void torusVertexKernel(float* vertices, int numVertices, float time, float majorRadius, float minorRadius) {
    int idx = blockIdx.x * blockDim.x + threadIdx.x;
    
    if (idx < numVertices) {
        // Calculate vertex index for position data (6 floats per vertex: x,y,z,r,g,b)
        int vertexOffset = idx * 6;
        
        // Generate torus parameters
        int segments = (int)sqrt((float)numVertices);
        int rings = segments;
        
        int u = idx % segments;
        int v = idx / segments;
        
        if (v >= rings) return;
        
        float theta = 2.0f * M_PI * u / segments;
        float phi = 2.0f * M_PI * v / rings;
        
        // Add rotation animation
        theta += time * 0.5f;
        phi += time * 0.3f;
        
        // Calculate torus coordinates
        float cosTheta = cosf(theta);
        float sinTheta = sinf(theta);
        float cosPhi = cosf(phi);
        float sinPhi = sinf(phi);
        
        // Torus parametric equations
        float x = (majorRadius + minorRadius * cosPhi) * cosTheta;
        float y = (majorRadius + minorRadius * cosPhi) * sinTheta;
        float z = minorRadius * sinPhi;
        
        // Update vertex positions
        vertices[vertexOffset + 0] = x;
        vertices[vertexOffset + 1] = y;
        vertices[vertexOffset + 2] = z;
    }
}

// CUDA kernel for updating torus vertex colors with dynamic effects
__global__ void torusColorKernel(float* vertices, int numVertices, float time) {
    int idx = blockIdx.x * blockDim.x + threadIdx.x;
    
    if (idx < numVertices) {
        // Calculate vertex index for color data (6 floats per vertex: x,y,z,r,g,b)
        int vertexOffset = idx * 6;
        
        // Get vertex position
        float x = vertices[vertexOffset + 0];
        float y = vertices[vertexOffset + 1];
        float z = vertices[vertexOffset + 2];
        
        // Calculate distance from center for color effect
        float distance = sqrtf(x*x + y*y + z*z);
        
        // Create pulsating color effect based on distance and time
        float pulse = sinf(time * 2.0f + distance * 0.5f) * 0.5f + 0.5f;
        float wave = sinf(time + idx * 0.1f) * 0.3f + 0.7f;
        
        // Generate RGB colors
        float r = pulse * wave;
        float g = sinf(time * 1.5f + distance * 0.3f) * 0.4f + 0.6f;
        float b = cosf(time * 0.8f + distance * 0.7f) * 0.5f + 0.5f;
        
        // Update vertex colors
        vertices[vertexOffset + 3] = r;
        vertices[vertexOffset + 4] = g;
        vertices[vertexOffset + 5] = b;
    }
}

// Host function to update torus vertices
void updateTorusVerticesCUDA(float* vertices, int numVertices, float time, float majorRadius, float minorRadius) {
    float* d_vertices;
    size_t size = numVertices * 6 * sizeof(float); // 6 floats per vertex (x,y,z,r,g,b)
    
    // Allocate GPU memory
    cudaMalloc(&d_vertices, size);
    
    // Copy data to GPU
    cudaMemcpy(d_vertices, vertices, size, cudaMemcpyHostToDevice);
    
    // Launch kernel
    int threadsPerBlock = 256;
    int blocksPerGrid = (numVertices + threadsPerBlock - 1) / threadsPerBlock;
    
    torusVertexKernel<<<blocksPerGrid, threadsPerBlock>>>(d_vertices, numVertices, time, majorRadius, minorRadius);
    
    // Wait for kernel to complete
    cudaDeviceSynchronize();
    
    // Copy result back to host
    cudaMemcpy(vertices, d_vertices, size, cudaMemcpyDeviceToHost);
    
    // Free GPU memory
    cudaFree(d_vertices);
}

// Host function to update torus colors
void updateTorusColorsCUDA(float* vertices, int numVertices, float time) {
    float* d_vertices;
    size_t size = numVertices * 6 * sizeof(float); // 6 floats per vertex (x,y,z,r,g,b)
    
    // Allocate GPU memory
    cudaMalloc(&d_vertices, size);
    
    // Copy data to GPU
    cudaMemcpy(d_vertices, vertices, size, cudaMemcpyHostToDevice);
    
    // Launch kernel
    int threadsPerBlock = 256;
    int blocksPerGrid = (numVertices + threadsPerBlock - 1) / threadsPerBlock;
    
    torusColorKernel<<<blocksPerGrid, threadsPerBlock>>>(d_vertices, numVertices, time);
    
    // Wait for kernel to complete
    cudaDeviceSynchronize();
    
    // Copy result back to host
    cudaMemcpy(vertices, d_vertices, size, cudaMemcpyDeviceToHost);
    
    // Free GPU memory
    cudaFree(d_vertices);
}

// Initialize CUDA
int initCUDA() {
    int deviceCount;
    cudaError_t error = cudaGetDeviceCount(&deviceCount);
    
    if (error != cudaSuccess) {
        return 0; // CUDA not available
    }
    
    if (deviceCount == 0) {
        return 0; // No CUDA devices
    }
    
    // Set device 0 as default
    cudaSetDevice(0);
    return 1; // Success
}

// Cleanup CUDA resources
void cleanupCUDA() {
    cudaDeviceReset();
}
