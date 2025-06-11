# ===== BUILD STAGE =====
FROM ubuntu:22.04 AS builder

# Set environment variables for build
ENV DEBIAN_FRONTEND=noninteractive
ENV CGO_ENABLED=1
ENV GO_VERSION=1.22.10

# Install basic dependencies first
RUN apt-get update && apt-get install -y \
    wget \
    curl \
    git \
    build-essential \
    gcc \
    g++ \
    pkg-config \
    ca-certificates \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Install OpenGL and X11 dependencies (required for Go OpenGL bindings)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgl1-mesa-dev \
    libglu1-mesa-dev \
    libglfw3-dev \
    libglew-dev \
    libx11-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxcursor-dev \
    libxi-dev \
    libxxf86vm-dev \
    && rm -rf /var/lib/apt/lists/*

# Add NVIDIA CUDA repository and install CUDA (optional)
RUN wget https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/x86_64/cuda-keyring_1.0-1_all.deb && \
    dpkg -i cuda-keyring_1.0-1_all.deb && \
    apt-get update && \
    (apt-get install -y cuda-toolkit-11-8 || echo "CUDA installation failed, continuing without CUDA support") && \
    rm -rf /var/lib/apt/lists/* && \
    rm -f cuda-keyring_1.0-1_all.deb

# Set CUDA environment if available
ENV PATH=$PATH:/usr/local/cuda/bin
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/cuda/lib64

# Install Go
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz

# Set Go environment
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go
ENV PATH=$PATH:$GOPATH/bin

# Create workspace directory
WORKDIR /workspace

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy source code
COPY . .

# Try to compile CUDA kernel if CUDA is available, otherwise create dummy library
RUN if command -v nvcc >/dev/null 2>&1; then \
        echo "CUDA found, compiling kernel..." && \
        nvcc -c torus_kernel.cu -o torus_kernel.o -Xcompiler -fPIC && \
        ar rcs libtorus_kernel.a torus_kernel.o; \
    else \
        echo "CUDA not found, creating dummy library..." && \
        echo "void updateTorusVerticesCUDA() {}" > dummy.c && \
        echo "void updateTorusColorsCUDA() {}" >> dummy.c && \
        echo "int initCUDA() { return 0; }" >> dummy.c && \
        echo "void cleanupCUDA() {}" >> dummy.c && \
        gcc -c dummy.c -o torus_kernel.o && \
        ar rcs libtorus_kernel.a torus_kernel.o; \
    fi

# Build the application
RUN go build -o torus-app .

# ===== RUNTIME STAGE =====
FROM ubuntu:22.04 AS runtime

# Set environment variables for runtime
ENV DEBIAN_FRONTEND=noninteractive
ENV DISPLAY=:0

# Install only runtime dependencies
RUN apt-get update && apt-get install -y \
    libgl1-mesa-glx \
    libglu1-mesa \
    libglfw3 \
    libglew2.2 \
    libx11-6 \
    libxrandr2 \
    libxinerama1 \
    libxcursor1 \
    libxi6 \
    xvfb \
    x11-apps \
    mesa-utils \
    && rm -rf /var/lib/apt/lists/*

# Create workspace directory
WORKDIR /workspace

# Copy the built application from builder stage
COPY --from=builder /workspace/torus-app .

# Create a script to run with virtual display
RUN echo '#!/bin/bash\n\
Xvfb :0 -screen 0 1024x768x24 -ac +extension GLX +render -noreset &\n\
export DISPLAY=:0\n\
sleep 2\n\
exec "$@"' > /usr/local/bin/run-with-display.sh && \
    chmod +x /usr/local/bin/run-with-display.sh

# Expose any ports if needed (for future web interface)
EXPOSE 8080

# Default command
CMD ["/usr/local/bin/run-with-display.sh", "./torus-app"]

# ===== DEVELOPMENT STAGE =====
FROM builder AS development

# Install additional development tools
RUN apt-get update && apt-get install -y \
    vim \
    nano \
    htop \
    gdb \
    valgrind \
    strace \
    && rm -rf /var/lib/apt/lists/*

# Create the display script in development too
RUN echo '#!/bin/bash\n\
Xvfb :0 -screen 0 1024x768x24 -ac +extension GLX +render -noreset &\n\
export DISPLAY=:0\n\
sleep 2\n\
exec "$@"' > /usr/local/bin/run-with-display.sh && \
    chmod +x /usr/local/bin/run-with-display.sh

# Expose ports
EXPOSE 8080

# Default to bash for development
CMD ["/bin/bash"]
