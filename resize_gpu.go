//go:build gpu && amd64

package imaging

import (
	"image"

	"golang.org/x/image/gpu"
)

// tryGPUResize attempts to resize using GPU (bilinear). Returns nil if GPU is unavailable.
func tryGPUResize(src *image.NRGBA, dstW, dstH int) *image.NRGBA {
	if !gpu.GPUModeEnabled() || !gpu.Available() {
		return nil
	}
	return gpu.ResizeGPU(src, dstW, dstH)
}

// tryGPUResizeCatmullRom attempts to resize using GPU (Catmull-Rom bicubic).
// Returns nil if GPU is unavailable.
func tryGPUResizeCatmullRom(src *image.NRGBA, dstW, dstH int) *image.NRGBA {
	if !gpu.GPUModeEnabled() || !gpu.Available() {
		return nil
	}
	return gpu.ResizeCatmullRomGPU(src, dstW, dstH)
}
