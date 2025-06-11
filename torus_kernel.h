#ifndef TORUS_KERNEL_H
#define TORUS_KERNEL_H

#ifdef __cplusplus
extern "C" {
#endif

// CUDA function declarations
void updateTorusVerticesCUDA(float* vertices, int numVertices, float time, float majorRadius, float minorRadius);
void updateTorusColorsCUDA(float* colors, int numVertices, float time);
int initCUDA();
void cleanupCUDA();

#ifdef __cplusplus
}
#endif

#endif // TORUS_KERNEL_H
