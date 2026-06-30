//go:build !gpu || !amd64

package imaging

import "image"

// tryGPUResize returns nil — caller falls back to SIMD/scalar.
func tryGPUResize(src *image.NRGBA, dstW, dstH int) *image.NRGBA {
	return nil
}
