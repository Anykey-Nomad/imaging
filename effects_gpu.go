//go:build gpu && amd64

package imaging

import (
	"image"

	"golang.org/x/image/gpu"
)

// tryGPUGrayscale attempts grayscale conversion on GPU. Returns nil if unavailable.
func tryGPUGrayscale(src *image.NRGBA) *image.NRGBA {
	if !gpu.GPUModeEnabled() || !gpu.Available() {
		return nil
	}
	return gpu.GrayscaleGPU(src)
}

// tryGPUBlur attempts Gaussian blur on GPU. Returns nil if unavailable.
func tryGPUBlur(src *image.NRGBA, sigma float64) *image.NRGBA {
	if !gpu.GPUModeEnabled() || !gpu.Available() {
		return nil
	}
	return gpu.BlurGPU(src, sigma)
}

// tryGPUSharpen attempts unsharp-mask sharpening on GPU. Returns nil if unavailable.
func tryGPUSharpen(src *image.NRGBA, sigma float64) *image.NRGBA {
	if !gpu.GPUModeEnabled() || !gpu.Available() {
		return nil
	}
	return gpu.SharpenGPU(src, sigma)
}
