package imaging

import (
	"image"
	"math"
)

func gaussianBlurKernel(x, sigma float64) float64 {
	return math.Exp(-(x*x)/(2*sigma*sigma)) / (sigma * math.Sqrt(2*math.Pi))
}

// Blur produces a blurred version of the image using a Gaussian function.
// Sigma parameter must be positive and indicates how much the image will be blurred.
//
// Example:
//
//	dstImage := imaging.Blur(srcImage, 3.5)
//
func Blur(img image.Image, sigma float64) *image.NRGBA {
	if sigma <= 0 {
		return Clone(img)
	}

	// Try GPU-accelerated blur first (bilinear via WebGPU compute shader).
	if nrgba, ok := img.(*image.NRGBA); ok {
		if gpuResult := tryGPUBlur(nrgba, sigma); gpuResult != nil {
			return gpuResult
		}
	}

	radius := int(math.Ceil(sigma * 3.0))
	kernel := make([]float64, radius+1)

	for i := 0; i <= radius; i++ {
		kernel[i] = gaussianBlurKernel(float64(i), sigma)
	}

	return blurVertical(blurHorizontal(img, kernel), kernel)
}

// Sharpen produces a sharpened version of the image.
// Sigma parameter must be positive and indicates how much the image will be sharpened.
//
// Example:
//
//	dstImage := imaging.Sharpen(srcImage, 3.5)
//
func Sharpen(img image.Image, sigma float64) *image.NRGBA {
	if sigma <= 0 {
		return Clone(img)
	}

	// Try GPU-accelerated sharpen first (unsharp mask via WebGPU compute shader).
	if nrgba, ok := img.(*image.NRGBA); ok {
		if gpuResult := tryGPUSharpen(nrgba, sigma); gpuResult != nil {
			return gpuResult
		}
	}

	src := toNRGBA(img)
	blurred := Blur(src, sigma)
	dst := image.NewNRGBA(src.Bounds())

	applyUnsharpMask(src, blurred, dst)
	return dst
}
