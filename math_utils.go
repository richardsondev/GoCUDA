package main

import "math"

// In a real application, use a dedicated math library like 'gonum/mat' or 'go-gl/mathgl'

type vec3 struct{ x, y, z float32 }
type mat4 [16]float32

func identity() mat4 {
	return mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func perspective(fovY, aspect, near, far float32) mat4 {
	f := float32(1.0 / math.Tan(float64(fovY*(math.Pi/180.0))/2.0))
	return mat4{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (far + near) / (near - far), -1,
		0, 0, (2 * far * near) / (near - far), 0,
	}
}

func lookAt(eye, center, up vec3) mat4 {
	f := normalize(sub(center, eye))
	s := normalize(cross(f, up))
	u := cross(s, f)

	return mat4{
		s.x, u.x, -f.x, 0,
		s.y, u.y, -f.y, 0,
		s.z, u.z, -f.z, 0,
		-dot(s, eye), -dot(u, eye), dot(f, eye), 1,
	}
}

func rotate(m mat4, angle float32, axis vec3) mat4 {
	c := float32(math.Cos(float64(angle * (math.Pi / 180.0))))
	s := float32(math.Sin(float64(angle * (math.Pi / 180.0))))
	t := 1 - c
	x, y, z := axis.x, axis.y, axis.z

	// Normalize axis
	lenSq := x*x + y*y + z*z
	if lenSq > 0 && (lenSq < 0.9999 || lenSq > 1.0001) { // Check if not already normalized
		invLen := 1.0 / float32(math.Sqrt(float64(lenSq)))
		x *= invLen
		y *= invLen
		z *= invLen
	}

	rot := mat4{
		t*x*x + c, t*x*y - s*z, t*x*z + s*y, 0,
		t*x*y + s*z, t*y*y + c, t*y*z - s*x, 0,
		t*x*z - s*y, t*y*z + s*x, t*z*z + c, 0,
		0, 0, 0, 1,
	}
	return multiply(m, rot) // Apply rotation to existing model matrix
}

func sub(a, b vec3) vec3 { return vec3{a.x - b.x, a.y - b.y, a.z - b.z} }
func cross(a, b vec3) vec3 {
	return vec3{a.y*b.z - a.z*b.y, a.z*b.x - a.x*b.z, a.x*b.y - a.y*b.x}
}
func dot(a, b vec3) float32 { return a.x*b.x + a.y*b.y + a.z*b.z }
func length(v vec3) float32 { return float32(math.Sqrt(float64(dot(v, v)))) }
func normalize(v vec3) vec3 {
	l := length(v)
	if l == 0 {
		return vec3{0, 0, 0}
	}
	return vec3{v.x / l, v.y / l, v.z / l}
}

func multiply(a, b mat4) mat4 {
	var out mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			sum := float32(0.0)
			for k := 0; k < 4; k++ {
				sum += a[i*4+k] * b[k*4+j]
			}
			out[i*4+j] = sum
		}
	}
	return out
}
