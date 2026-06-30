//go:build !goexperiment.simd || !amd64

package imaging

import (
	"image"
)

// blurHorizontal is the scalar fallback for horizontal Gaussian blur.
func blurHorizontal(img image.Image, kernel []float64) *image.NRGBA {
	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, src.w, src.h))
	radius := len(kernel) - 1

	parallel(0, src.h, func(ys <-chan int) {
		scanLine := make([]uint8, src.w*4)
		scanLineF := make([]float64, len(scanLine))
		for y := range ys {
			src.scan(0, y, src.w, y+1, scanLine)
			for i, v := range scanLine {
				scanLineF[i] = float64(v)
			}
			for x := 0; x < src.w; x++ {
				min := x - radius
				if min < 0 {
					min = 0
				}
				max := x + radius
				if max > src.w-1 {
					max = src.w - 1
				}
				var r, g, b, a, wsum float64
				for ix := min; ix <= max; ix++ {
					i := ix * 4
					weight := kernel[absint(x-ix)]
					wsum += weight
					s := scanLineF[i : i+4 : i+4]
					wa := s[3] * weight
					r += s[0] * wa
					g += s[1] * wa
					b += s[2] * wa
					a += wa
				}
				if a != 0 {
					aInv := 1 / a
					j := y*dst.Stride + x*4
					d := dst.Pix[j : j+4 : j+4]
					d[0] = clamp(r * aInv)
					d[1] = clamp(g * aInv)
					d[2] = clamp(b * aInv)
					d[3] = clamp(a / wsum)
				}
			}
		}
	})
	return dst
}

// blurVertical is the scalar fallback for vertical Gaussian blur.
func blurVertical(img image.Image, kernel []float64) *image.NRGBA {
	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, src.w, src.h))
	radius := len(kernel) - 1

	parallel(0, src.w, func(xs <-chan int) {
		scanLine := make([]uint8, src.h*4)
		scanLineF := make([]float64, len(scanLine))
		for x := range xs {
			src.scan(x, 0, x+1, src.h, scanLine)
			for i, v := range scanLine {
				scanLineF[i] = float64(v)
			}
			for y := 0; y < src.h; y++ {
				min := y - radius
				if min < 0 {
					min = 0
				}
				max := y + radius
				if max > src.h-1 {
					max = src.h - 1
				}
				var r, g, b, a, wsum float64
				for iy := min; iy <= max; iy++ {
					i := iy * 4
					weight := kernel[absint(y-iy)]
					wsum += weight
					s := scanLineF[i : i+4 : i+4]
					wa := s[3] * weight
					r += s[0] * wa
					g += s[1] * wa
					b += s[2] * wa
					a += wa
				}
				if a != 0 {
					aInv := 1 / a
					j := y*dst.Stride + x*4
					d := dst.Pix[j : j+4 : j+4]
					d[0] = clamp(r * aInv)
					d[1] = clamp(g * aInv)
					d[2] = clamp(b * aInv)
					d[3] = clamp(a / wsum)
				}
			}
		}
	})
	return dst
}

// Grayscale is the scalar fallback for grayscale conversion.
// Uses BT.601 luminance weights: 0.299*R + 0.587*G + 0.114*B.
func Grayscale(img image.Image) *image.NRGBA {
	// Try GPU-accelerated grayscale first (BT.601 via WebGPU compute shader).
	if nrgba, ok := img.(*image.NRGBA); ok {
		if gpuResult := tryGPUGrayscale(nrgba); gpuResult != nil {
			return gpuResult
		}
	}

	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, src.w, src.h))
	parallel(0, src.h, func(ys <-chan int) {
		for y := range ys {
			i := y * dst.Stride
			src.scan(0, y, src.w, y+1, dst.Pix[i:i+src.w*4])
			for x := 0; x < src.w; x++ {
				d := dst.Pix[i : i+3 : i+3]
				r := d[0]
				g := d[1]
				b := d[2]
				f := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
				y := uint8(f + 0.5)
				d[0] = y
				d[1] = y
				d[2] = y
				i += 4
			}
		}
	})
	return dst
}
