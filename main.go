package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	initialWidth  = 800
	initialHeight = 600
)

var (
	currentWidth  = initialWidth
	currentHeight = initialHeight

	// FPS tracking
	frameCount  = 0
	lastFPSTime = time.Now()
	currentFPS  = 0.0

	// Window state tracking
	projectionNeedsUpdate = false
)

// TorusParams holds the configurable parameters for the torus
type TorusParams struct {
	MajorRadius    float32 `json:"majorRadius"`
	MinorRadius    float32 `json:"minorRadius"`
	MajorSegments  int     `json:"majorSegments"`
	MinorSegments  int     `json:"minorSegments"`
	RotationSpeedX float32 `json:"rotationSpeedX"`
	RotationSpeedY float32 `json:"rotationSpeedY"`
}

// AsteroidParams holds configurable parameters for the asteroid field
type AsteroidParams struct {
	SpawnRate     float32 `json:"spawnRate"`
	MaxAsteroids  int     `json:"maxAsteroids"`
	MinSize       float32 `json:"minSize"`
	MaxSize       float32 `json:"maxSize"`
	MinSpeed      float32 `json:"minSpeed"`
	MaxSpeed      float32 `json:"maxSpeed"`
	RotationSpeed float32 `json:"rotationSpeed"`
}

// Scene types
type SceneType int

const (
	SceneDonut SceneType = iota
	SceneAsteroidField
)

// Asteroid represents a single asteroid in the field
type Asteroid struct {
	Position    mgl32.Vec3
	Velocity    mgl32.Vec3
	Rotation    mgl32.Vec3
	RotationVel mgl32.Vec3
	Scale       float32
	Vertices    []float32
	Created     time.Time
}

// Scene configuration
type SceneConfig struct {
	CurrentScene      SceneType
	AsteroidSpawnRate float64 // seconds between spawns
	MaxAsteroids      int
	ViewportBounds    struct {
		MinX, MaxX float32
		MinY, MaxY float32
		MinZ, MaxZ float32
	}
}

var sceneConfig = SceneConfig{
	CurrentScene:      SceneDonut,
	AsteroidSpawnRate: 1.0, // 1 second
	MaxAsteroids:      100,
}

var asteroids []Asteroid
var lastAsteroidSpawn time.Time

// SetCurrentScene allows external modules to change the current scene
func SetCurrentScene(sceneName string) {
	switch sceneName {
	case "donut":
		sceneConfig.CurrentScene = SceneDonut
		LogInfo("Scene", "Scene changed to Donut via API")
	case "asteroids":
		sceneConfig.CurrentScene = SceneAsteroidField
		LogInfo("Scene", "Scene changed to Asteroid Field via API")
	default:
		LogError("Scene", fmt.Sprintf("Unknown scene name: %s", sceneName))
	}
}

// GetCurrentSceneName returns the current scene as a string
func GetCurrentSceneName() string {
	switch sceneConfig.CurrentScene {
	case SceneDonut:
		return "donut"
	case SceneAsteroidField:
		return "asteroids"
	default:
		return "unknown"
	}
}

// UpdateAsteroidConfig allows external modules to update asteroid parameters
func UpdateAsteroidConfig(params AsteroidParams) {
	sceneConfig.AsteroidSpawnRate = float64(params.SpawnRate)
	sceneConfig.MaxAsteroids = params.MaxAsteroids
	LogInfo("Asteroids", fmt.Sprintf("Asteroid config updated: spawn rate=%.1f, max=%d", params.SpawnRate, params.MaxAsteroids))
}

// ClearAllAsteroids removes all asteroids from the field
func ClearAllAsteroids() {
	asteroidCount := len(asteroids)
	asteroids = nil
	LogInfo("Asteroids", fmt.Sprintf("Cleared %d asteroids from field", asteroidCount))
}

// GetCurrentFPS returns the current frames per second
func GetCurrentFPS() float64 {
	return currentFPS
}

// GetWindowSize returns the current window dimensions
func GetWindowSize() (int, int) {
	return currentWidth, currentHeight
}

// updateFPS calculates and updates the current FPS
func updateFPS() {
	frameCount++
	now := time.Now()
	elapsed := now.Sub(lastFPSTime).Seconds()

	if elapsed >= 1.0 { // Update FPS every second
		currentFPS = float64(frameCount) / elapsed
		frameCount = 0
		lastFPSTime = now
	}
}

func init() {
	runtime.LockOSThread()
	// Initialize scene viewport bounds
	sceneConfig.ViewportBounds.MinX = -10.0
	sceneConfig.ViewportBounds.MaxX = 10.0
	sceneConfig.ViewportBounds.MinY = -7.5
	sceneConfig.ViewportBounds.MaxY = 7.5
	sceneConfig.ViewportBounds.MinZ = -5.0
	sceneConfig.ViewportBounds.MaxZ = 5.0

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
}

func createWindow() (*glfw.Window, error) {
	glfw.WindowHint(glfw.Resizable, glfw.True) // Enable resizing
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(currentWidth, currentHeight, "CUDA-Accelerated 3D Torus", nil, nil)
	if err != nil {
		return nil, err
	}

	window.MakeContextCurrent()
	// Set up resize callback
	window.SetSizeCallback(func(w *glfw.Window, width, height int) {
		currentWidth = width
		currentHeight = height
		projectionNeedsUpdate = true
		gl.Viewport(0, 0, int32(width), int32(height))
		LogInfo("Window", fmt.Sprintf("Window resized to %dx%d", width, height))
	})

	return window, nil
}

// Generate torus vertices with triangulation
func generateTorusVertices(majorRadius, minorRadius float32, majorSegments, minorSegments int) []float32 {
	vertices := make([]float32, 0, majorSegments*minorSegments*6*6) // 6 triangles per quad, 6 floats per vertex

	for i := 0; i < majorSegments; i++ {
		for j := 0; j < minorSegments; j++ {
			// Current segment parameters
			theta := 2.0 * math.Pi * float64(i) / float64(majorSegments)
			phi := 2.0 * math.Pi * float64(j) / float64(minorSegments)

			// Next segment parameters
			nextTheta := 2.0 * math.Pi * float64((i+1)%majorSegments) / float64(majorSegments)
			nextPhi := 2.0 * math.Pi * float64((j+1)%minorSegments) / float64(minorSegments)

			// Calculate vertices for quad
			v1 := calculateTorusVertex(majorRadius, minorRadius, float32(theta), float32(phi))
			v2 := calculateTorusVertex(majorRadius, minorRadius, float32(nextTheta), float32(phi))
			v3 := calculateTorusVertex(majorRadius, minorRadius, float32(nextTheta), float32(nextPhi))
			v4 := calculateTorusVertex(majorRadius, minorRadius, float32(theta), float32(nextPhi))

			// Add two triangles for each quad
			// Triangle 1: v1, v2, v3
			vertices = append(vertices, v1...)
			vertices = append(vertices, v2...)
			vertices = append(vertices, v3...)

			// Triangle 2: v1, v3, v4
			vertices = append(vertices, v1...)
			vertices = append(vertices, v3...)
			vertices = append(vertices, v4...)
		}
	}

	return vertices
}

// Calculate a single torus vertex (position + color)
func calculateTorusVertex(majorRadius, minorRadius, theta, phi float32) []float32 {
	cosTheta := float32(math.Cos(float64(theta)))
	sinTheta := float32(math.Sin(float64(theta)))
	cosPhi := float32(math.Cos(float64(phi)))
	sinPhi := float32(math.Sin(float64(phi)))

	x := (majorRadius + minorRadius*cosPhi) * cosTheta
	y := (majorRadius + minorRadius*cosPhi) * sinTheta
	z := minorRadius * sinPhi

	// Base colors (will be modified by CUDA)
	r := float32(0.5 + 0.5*math.Sin(float64(theta)))
	g := float32(0.5 + 0.5*math.Sin(float64(phi)))
	b := float32(0.5 + 0.5*math.Cos(float64(theta+phi)))

	return []float32{x, y, z, r, g, b}
}

// CPU fallback for animation when CUDA is not available
func updateTorusCPU(vertices []float32, time float32) {
	numVertices := len(vertices) / 6

	for i := 0; i < numVertices; i++ {
		vertexOffset := i * 6

		// Get original position
		x := vertices[vertexOffset+0]
		y := vertices[vertexOffset+1]
		z := vertices[vertexOffset+2]

		// Calculate distance for color effect
		distance := float32(math.Sqrt(float64(x*x + y*y + z*z)))

		// Simple CPU-based color animation
		pulse := float32(math.Sin(float64(time*2.0+distance*0.5)))*0.5 + 0.5
		wave := float32(math.Sin(float64(time+float32(i)*0.1)))*0.3 + 0.7

		vertices[vertexOffset+3] = pulse * wave                                                // R
		vertices[vertexOffset+4] = float32(math.Sin(float64(time*1.5+distance*0.3)))*0.4 + 0.6 // G
		vertices[vertexOffset+5] = float32(math.Cos(float64(time*0.8+distance*0.7)))*0.5 + 0.5 // B
	}
}

// Generate a detailed asteroid with complex geometry
func generateAsteroidVertices(complexity int, baseRadius float32) []float32 {
	// Create an icosphere-like asteroid with random perturbations
	vertices := make([]float32, 0, complexity*complexity*6*6)

	segments := complexity
	rings := complexity / 2

	for i := 0; i < rings; i++ {
		for j := 0; j < segments; j++ {
			// Spherical coordinates with random perturbations
			phi := math.Pi * float64(i) / float64(rings)
			theta := 2.0 * math.Pi * float64(j) / float64(segments)

			// Next ring/segment
			nextPhi := math.Pi * float64(i+1) / float64(rings)
			nextTheta := 2.0 * math.Pi * float64((j+1)%segments) / float64(segments)

			// Generate 4 vertices for a quad with random perturbations
			vertices1 := generateAsteroidVertex(phi, theta, baseRadius)
			vertices2 := generateAsteroidVertex(phi, nextTheta, baseRadius)
			vertices3 := generateAsteroidVertex(nextPhi, theta, baseRadius)
			vertices4 := generateAsteroidVertex(nextPhi, nextTheta, baseRadius)

			// Create two triangles from the quad
			// Triangle 1
			vertices = append(vertices, vertices1...)
			vertices = append(vertices, vertices2...)
			vertices = append(vertices, vertices3...)

			// Triangle 2
			vertices = append(vertices, vertices2...)
			vertices = append(vertices, vertices4...)
			vertices = append(vertices, vertices3...)
		}
	}

	return vertices
}

func generateAsteroidVertex(phi, theta float64, baseRadius float32) []float32 {
	// Add random perturbations to create irregular asteroid shape
	radiusVariation := 0.3 + rand.Float32()*0.7 // 30-100% of base radius
	bumpiness := 0.8 + rand.Float32()*0.4       // Additional surface variation

	// Calculate base spherical position
	r := float64(baseRadius * radiusVariation * bumpiness)
	x := r * math.Sin(phi) * math.Cos(theta)
	y := r * math.Sin(phi) * math.Sin(theta)
	z := r * math.Cos(phi)

	// Generate random colors (rocky asteroid colors)
	colorVariation := 0.7 + rand.Float32()*0.3
	red := (0.4 + rand.Float32()*0.3) * colorVariation // Brown/gray tones
	green := (0.3 + rand.Float32()*0.2) * colorVariation
	blue := (0.2 + rand.Float32()*0.2) * colorVariation

	return []float32{
		float32(x), float32(y), float32(z), // position
		red, green, blue, // color
	}
}

// Create a new asteroid entering from off-screen
func createNewAsteroid() Asteroid {
	bounds := sceneConfig.ViewportBounds

	// Get current asteroid parameters for configuration
	asteroidParams := GetCurrentAsteroidParams()

	// Spawn from random edge of viewport
	var pos mgl32.Vec3
	var vel mgl32.Vec3

	edge := rand.Intn(4) // 0=left, 1=right, 2=top, 3=bottom

	// Generate speed based on configurable parameters
	baseSpeed := asteroidParams.MinSpeed + rand.Float32()*(asteroidParams.MaxSpeed-asteroidParams.MinSpeed)

	switch edge {
	case 0: // Left edge
		pos = mgl32.Vec3{bounds.MinX - 2, bounds.MinY + rand.Float32()*(bounds.MaxY-bounds.MinY), bounds.MinZ + rand.Float32()*(bounds.MaxZ-bounds.MinZ)}
		vel = mgl32.Vec3{baseSpeed, (rand.Float32() - 0.5) * baseSpeed * 0.5, (rand.Float32() - 0.5) * baseSpeed * 0.3}
	case 1: // Right edge
		pos = mgl32.Vec3{bounds.MaxX + 2, bounds.MinY + rand.Float32()*(bounds.MaxY-bounds.MinY), bounds.MinZ + rand.Float32()*(bounds.MaxZ-bounds.MinZ)}
		vel = mgl32.Vec3{-baseSpeed, (rand.Float32() - 0.5) * baseSpeed * 0.5, (rand.Float32() - 0.5) * baseSpeed * 0.3}
	case 2: // Top edge
		pos = mgl32.Vec3{bounds.MinX + rand.Float32()*(bounds.MaxX-bounds.MinX), bounds.MaxY + 2, bounds.MinZ + rand.Float32()*(bounds.MaxZ-bounds.MinZ)}
		vel = mgl32.Vec3{(rand.Float32() - 0.5) * baseSpeed * 0.5, -baseSpeed, (rand.Float32() - 0.5) * baseSpeed * 0.3}
	case 3: // Bottom edge
		pos = mgl32.Vec3{bounds.MinX + rand.Float32()*(bounds.MaxX-bounds.MinX), bounds.MinY - 2, bounds.MinZ + rand.Float32()*(bounds.MaxZ-bounds.MinZ)}
		vel = mgl32.Vec3{(rand.Float32() - 0.5) * baseSpeed * 0.5, baseSpeed, (rand.Float32() - 0.5) * baseSpeed * 0.3}
	}

	// Random rotation
	rotation := mgl32.Vec3{
		rand.Float32() * 2 * math.Pi,
		rand.Float32() * 2 * math.Pi,
		rand.Float32() * 2 * math.Pi,
	}

	// Use configurable rotation speed
	rotationSpeed := asteroidParams.RotationSpeed
	rotationVel := mgl32.Vec3{
		(rand.Float32() - 0.5) * rotationSpeed,
		(rand.Float32() - 0.5) * rotationSpeed,
		(rand.Float32() - 0.5) * rotationSpeed,
	}

	// Use configurable size range
	scale := asteroidParams.MinSize + rand.Float32()*(asteroidParams.MaxSize-asteroidParams.MinSize)

	// Generate complex geometry
	complexity := 8 + rand.Intn(8) // 8-15 complexity
	baseRadius := 0.5 + rand.Float32()*0.5
	vertices := generateAsteroidVertices(complexity, baseRadius)

	return Asteroid{
		Position:    pos,
		Velocity:    vel,
		Rotation:    rotation,
		RotationVel: rotationVel,
		Scale:       scale,
		Vertices:    vertices,
		Created:     time.Now(),
	}
}

// Transform asteroid vertices by position, rotation, and scale
func transformAsteroidVertices(asteroid Asteroid) []float32 {
	transformed := make([]float32, len(asteroid.Vertices))
	copy(transformed, asteroid.Vertices)

	// Apply transformations to each vertex
	for i := 0; i < len(transformed); i += 6 { // 6 floats per vertex (x,y,z,r,g,b)
		// Get original position
		x, y, z := transformed[i], transformed[i+1], transformed[i+2]

		// Apply scale
		x *= asteroid.Scale
		y *= asteroid.Scale
		z *= asteroid.Scale

		// Apply rotation (simplified rotation around Y-axis primarily)
		rotY := asteroid.Rotation.Y()
		cosY, sinY := float32(math.Cos(float64(rotY))), float32(math.Sin(float64(rotY)))

		newX := x*cosY - z*sinY
		newZ := x*sinY + z*cosY

		// Apply translation
		transformed[i] = newX + asteroid.Position.X()
		transformed[i+1] = y + asteroid.Position.Y()
		transformed[i+2] = newZ + asteroid.Position.Z()

		// Colors remain unchanged (i+3, i+4, i+5)
	}

	return transformed
}

// Update asteroid positions and handle bouncing
func updateAsteroids(deltaTime float32) {
	bounds := sceneConfig.ViewportBounds

	for i := range asteroids {
		asteroid := &asteroids[i]

		// Update position
		asteroid.Position = asteroid.Position.Add(asteroid.Velocity.Mul(deltaTime))

		// Update rotation
		asteroid.Rotation = asteroid.Rotation.Add(asteroid.RotationVel.Mul(deltaTime))

		// Boundary bouncing
		bounced := false
		if asteroid.Position.X() < bounds.MinX || asteroid.Position.X() > bounds.MaxX {
			asteroid.Velocity[0] = -asteroid.Velocity[0]
			bounced = true
		}
		if asteroid.Position.Y() < bounds.MinY || asteroid.Position.Y() > bounds.MaxY {
			asteroid.Velocity[1] = -asteroid.Velocity[1]
			bounced = true
		}
		if asteroid.Position.Z() < bounds.MinZ || asteroid.Position.Z() > bounds.MaxZ {
			asteroid.Velocity[2] = -asteroid.Velocity[2]
			bounced = true
		}

		// Add some randomness to bouncing to prevent repetitive patterns
		if bounced {
			asteroid.Velocity = asteroid.Velocity.Mul(0.8 + rand.Float32()*0.4) // 80-120% speed retention
			// Slight random direction change
			asteroid.Velocity[0] += (rand.Float32() - 0.5) * 0.2
			asteroid.Velocity[1] += (rand.Float32() - 0.5) * 0.2
			asteroid.Velocity[2] += (rand.Float32() - 0.5) * 0.1
		}
	}

	// Remove old asteroids if we have too many
	if len(asteroids) > sceneConfig.MaxAsteroids {
		// Remove oldest asteroids
		copy(asteroids, asteroids[len(asteroids)-sceneConfig.MaxAsteroids:])
		asteroids = asteroids[:sceneConfig.MaxAsteroids]
	}
}

func main() {
	// Initialize logger first
	initLogger()

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatalf("Failed to initialize GLFW: %v", err)
	}
	defer glfw.Terminate()
	// Create window
	window, err := createWindow()
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	// Set up keyboard callback for scene switching
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			switch key {
			case glfw.Key1:
				sceneConfig.CurrentScene = SceneDonut
				LogInfo("Scene", "Switched to Donut scene")
			case glfw.Key2:
				sceneConfig.CurrentScene = SceneAsteroidField
				LogInfo("Scene", "Switched to Asteroid Field scene")
			case glfw.KeyEscape:
				window.SetShouldClose(true)
			}
		}
	})

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatalf("Failed to initialize OpenGL: %v", err)
	}

	// Print OpenGL version
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Printf("OpenGL version %s\n", version)

	// Initialize CUDA
	cudaEnabled := initCUDASystem()
	defer cleanupCUDASystem()

	// Initialize torus parameters
	torusParams := TorusParams{
		MajorRadius:    2.0,
		MinorRadius:    0.5,
		MajorSegments:  50,
		MinorSegments:  30,
		RotationSpeedX: 0.3,
		RotationSpeedY: 0.5,
	}

	// Start web server in a goroutine
	go StartWebServer(&torusParams, &cudaEnabled)

	// Create shader program
	program, err := newProgram(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		log.Fatalf("Failed to create shader program: %v", err)
	}

	// Generate initial torus vertices
	vertices := generateTorusVertices(torusParams.MajorRadius, torusParams.MinorRadius,
		torusParams.MajorSegments, torusParams.MinorSegments)

	// Create and bind VAO
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// Create and bind VBO
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	// Set up vertex attributes
	// Position attribute (location = 0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Color attribute (location = 1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	// Configure OpenGL
	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)

	// Get uniform locations
	modelLoc := gl.GetUniformLocation(program, gl.Str("model\x00"))
	viewLoc := gl.GetUniformLocation(program, gl.Str("view\x00"))
	projLoc := gl.GetUniformLocation(program, gl.Str("projection\x00")) // Set up matrices
	model := mgl32.Ident4()
	view := mgl32.LookAtV(mgl32.Vec3{0, 0, 10}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(currentWidth)/float32(currentHeight), 0.1, 100.0)

	startTime := time.Now()
	lastParams := torusParams
	needsRegeneration := false
	var allVertices []float32 // Combined vertices for current scene

	// Initialize with torus vertices
	allVertices = vertices

	// Main render loop
	for !window.ShouldClose() {
		// Poll events
		glfw.PollEvents()

		currentTime := time.Now()
		elapsed := float32(currentTime.Sub(startTime).Seconds())
		deltaTime := float32(currentTime.Sub(startTime).Seconds()) // For physics updates		 // Update FPS
		updateFPS()

		// Update projection matrix if window was resized
		if projectionNeedsUpdate {
			projection = mgl32.Perspective(mgl32.DegToRad(45.0), float32(currentWidth)/float32(currentHeight), 0.1, 100.0)
			projectionNeedsUpdate = false
		}

		// Scene-specific logic
		switch sceneConfig.CurrentScene {
		case SceneDonut:
			// Check if parameters changed
			currentParams := GetCurrentParams()
			if currentParams.MajorRadius != lastParams.MajorRadius ||
				currentParams.MinorRadius != lastParams.MinorRadius ||
				currentParams.MajorSegments != lastParams.MajorSegments ||
				currentParams.MinorSegments != lastParams.MinorSegments {
				needsRegeneration = true
				lastParams = currentParams
			}

			// Always update torusParams to include rotation speed changes
			torusParams = currentParams

			// Regenerate torus if parameters changed
			if needsRegeneration {
				vertices = generateTorusVertices(torusParams.MajorRadius, torusParams.MinorRadius,
					torusParams.MajorSegments, torusParams.MinorSegments)
				allVertices = vertices
				needsRegeneration = false
			}

			// Update vertices using CUDA if available
			if cudaEnabled {
				LogInfo("RenderPath", fmt.Sprintf("Torus update: Using CUDA. Params: R=%.2f, r=%.2f, MajorSeg=%d, MinorSeg=%d, RotXSpeed=%.2f, RotYSpeed=%.2f",
					torusParams.MajorRadius, torusParams.MinorRadius, torusParams.MajorSegments, torusParams.MinorSegments, torusParams.RotationSpeedX, torusParams.RotationSpeedY))
				updateTorusVertices(vertices, elapsed, torusParams.MajorRadius, torusParams.MinorRadius)
				updateTorusColors(vertices, elapsed)
			} else {
				LogInfo("RenderPath", fmt.Sprintf("Torus update: Using CPU. Params: R=%.2f, r=%.2f, MajorSeg=%d, MinorSeg=%d, RotXSpeed=%.2f, RotYSpeed=%.2f",
					torusParams.MajorRadius, torusParams.MinorRadius, torusParams.MajorSegments, torusParams.MinorSegments, torusParams.RotationSpeedX, torusParams.RotationSpeedY))
				updateTorusCPU(vertices, elapsed)
			}
			allVertices = vertices

		case SceneAsteroidField:
			// Spawn new asteroids
			if currentTime.Sub(lastAsteroidSpawn).Seconds() >= sceneConfig.AsteroidSpawnRate {
				if len(asteroids) < sceneConfig.MaxAsteroids {
					newAsteroid := createNewAsteroid()
					asteroids = append(asteroids, newAsteroid)
					LogInfo("Asteroid", fmt.Sprintf("Spawned asteroid #%d. Total: %d", len(asteroids), len(asteroids)))
				}
				lastAsteroidSpawn = currentTime
			}

			// Update asteroid physics
			updateAsteroids(deltaTime)

			// Combine all asteroid vertices into a single buffer
			totalFloats := 0
			for _, asteroid := range asteroids {
				totalFloats += len(asteroid.Vertices)
			}

			if totalFloats > 0 {
				allVertices = make([]float32, 0, totalFloats) // Pre-allocate with capacity

				if cudaEnabled && len(asteroids) > 0 {
					// TODO: Implement CUDA batch transformation for asteroids.
					// This function would take all asteroid data (positions, rotations, scales, base vertices)
					// and output the fully transformed allVertices slice.
					// e.g., allVertices = transformAllAsteroidsCUDA(asteroids, deltaTime) // Future implementation
					LogInfo("RenderPath", fmt.Sprintf("Asteroid transform: Using CUDA path (placeholder with CPU logic for %d asteroids)", len(asteroids)))

					// Placeholder: still uses CPU logic but under CUDA flag for now
					for _, asteroid := range asteroids {
						transformedVertices := transformAsteroidVertices(asteroid) // Existing CPU func
						allVertices = append(allVertices, transformedVertices...)
					}
				} else if len(asteroids) > 0 { // Handles !cudaEnabled or (cudaEnabled && len(asteroids) == 0)
					LogInfo("RenderPath", fmt.Sprintf("Asteroid transform: Using CPU path for %d asteroids", len(asteroids)))
					for _, asteroid := range asteroids {
						transformedVertices := transformAsteroidVertices(asteroid)
						allVertices = append(allVertices, transformedVertices...)
					}
				} else {
					// No asteroids to render
					allVertices = nil
				}
			} else {
				// No asteroids, so no vertices
				allVertices = nil
			}

		}

		// Update VBO with vertex data for current scene
		if len(allVertices) > 0 {
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
			gl.BufferData(gl.ARRAY_BUFFER, len(allVertices)*4, gl.Ptr(allVertices), gl.DYNAMIC_DRAW)
		}

		// Clear screen
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Use shader program
		gl.UseProgram(program)

		// Set different camera and rendering for each scene
		switch sceneConfig.CurrentScene {
		case SceneDonut:
			// Update model matrix with rotation animation using dynamic speeds
			currentParams := GetCurrentParams()
			rotY := float32(elapsed * currentParams.RotationSpeedY)
			rotX := float32(elapsed * currentParams.RotationSpeedX)

			// Create rotation matrices manually
			cosY, sinY := float32(math.Cos(float64(rotY))), float32(math.Sin(float64(rotY)))
			cosX, sinX := float32(math.Cos(float64(rotX))), float32(math.Sin(float64(rotX)))

			// Y rotation matrix
			rotMatY := mgl32.Mat4{
				cosY, 0, sinY, 0,
				0, 1, 0, 0,
				-sinY, 0, cosY, 0,
				0, 0, 0, 1,
			}

			// X rotation matrix
			rotMatX := mgl32.Mat4{
				1, 0, 0, 0,
				0, cosX, -sinX, 0,
				0, sinX, cosX, 0,
				0, 0, 0, 1,
			}

			// Combine rotations
			model = rotMatY.Mul4(rotMatX)

		case SceneAsteroidField:
			// Static camera view for asteroid field
			model = mgl32.Ident4()
			// Slightly different camera angle for asteroid field
			view = mgl32.LookAtV(mgl32.Vec3{0, 2, 8}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
		}
		// Set matrices
		gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])
		gl.UniformMatrix4fv(viewLoc, 1, false, &view[0])
		gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

		// Draw the current scene
		if len(allVertices) > 0 {
			gl.DrawArrays(gl.TRIANGLES, 0, int32(len(allVertices)/6))
		}

		// Swap buffers
		window.SwapBuffers()
	}
}
