//go:build !goexperiment.simd || !amd64

package imaging

import "image"

// applyUnsharpMask applies sharpen = clamp(2*original - blurred) (scalar fallback).
func applyUnsharpMask(src, blurred *image.NRGBA, dst *image.NRGBA) {
	b := src.Bounds()
	parallel(b.Min.Y, b.Max.Y, func(ys <-chan int) {
		for y := range ys {
			sOff := src.PixOffset(b.Min.X, y)
			dOff := dst.PixOffset(b.Min.X, y)
			rowEnd := sOff + b.Dx()*4
			for ; sOff < rowEnd; sOff += 4 {
				for c := 0; c < 3; c++ {
					val := 2*int(src.Pix[sOff+c]) - int(blurred.Pix[sOff+c])
					if val < 0 {
						val = 0
					} else if val > 255 {
						val = 255
					}
					dst.Pix[dOff+c] = uint8(val)
				}
				dst.Pix[dOff+3] = src.Pix[sOff+3]
				dOff += 4
			}
		}
	})
}
