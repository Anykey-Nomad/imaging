//go:build gpu && amd64

package imaging

import (
	"image"

	"golang.org/x/image/gpu"
)

// tryGPUResize attempts to resize using GPU. Returns nil if GPU is unavailable.
func tryGPUResize(src *image.NRGBA, dstW, dstH int) *image.NRGBA {
	if !gpu.GPUModeEnabled() || !gpu.Available() {
		return nil
	}
	return gpu.ResizeGPU(src, dstW, dstH)
}
