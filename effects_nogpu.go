//go:build !gpu || !amd64

package imaging

import "image"

// tryGPUGrayscale returns nil — caller falls back to SIMD/scalar.
func tryGPUGrayscale(src *image.NRGBA) *image.NRGBA { return nil }

// tryGPUBlur returns nil — caller falls back to SIMD/scalar.
func tryGPUBlur(src *image.NRGBA, sigma float64) *image.NRGBA { return nil }

// tryGPUSharpen returns nil — caller falls back to SIMD/scalar.
func tryGPUSharpen(src *image.NRGBA, sigma float64) *image.NRGBA { return nil }
