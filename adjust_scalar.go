//go:build !goexperiment.simd || !amd64

package imaging

import (
	"image"
	"math"
)

// adjustBrightnessImpl applies brightness using a LUT (scalar fallback).
// img is already *image.NRGBA — applies LUT directly without double conversion.
func adjustBrightnessImpl(img *image.NRGBA, percentage float64) *image.NRGBA {
	percentage = math.Min(math.Max(percentage, -100.0), 100.0)
	shift := 255.0 * percentage / 100.0
	lut := make([]uint8, 256)
	for i := 0; i < 256; i++ {
		lut[i] = clamp(float64(i) + shift)
	}
	b := img.Bounds()
	dst := image.NewNRGBA(b)
	parallel(b.Min.Y, b.Max.Y, func(ys <-chan int) {
		for y := range ys {
			sOff := img.PixOffset(b.Min.X, y)
			dOff := dst.PixOffset(b.Min.X, y)
			end := sOff + b.Dx()*4
			for ; sOff < end; sOff += 4 {
				dst.Pix[dOff+0] = lut[img.Pix[sOff+0]]
				dst.Pix[dOff+1] = lut[img.Pix[sOff+1]]
				dst.Pix[dOff+2] = lut[img.Pix[sOff+2]]
				dst.Pix[dOff+3] = img.Pix[sOff+3]
				dOff += 4
			}
		}
	})
	return dst
}
