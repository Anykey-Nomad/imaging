//go:build gpu && amd64

package imaging

import "golang.org/x/image/gpu"

// InitGPU initializes the WebGPU subsystem for image processing.
// Must be called once at application startup before any GPU operations.
func InitGPU() { gpu.Init() }
