//go:build goexperiment.simd && amd64

package imaging

import (
	"image"
	"simd/archsimd"
)

// resizeHorizontal resizes the image horizontally using SIMD-accelerated accumulation.
// Uses Float32x4 to process RGBA channels in parallel on amd64 with AVX/SSE.
func resizeHorizontal(img image.Image, width int, filter ResampleFilter) *image.NRGBA {
	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, width, src.h))
	weights := precomputeWeights(width, src.w, filter)

	parallel(0, src.h, func(ys <-chan int) {
		scanLine := make([]uint8, src.w*4)
		var buf [4]float32

		for y := range ys {
			src.scan(0, y, src.w, y+1, scanLine)
			j0 := y * dst.Stride

			for x := range weights {
				var sum archsimd.Float32x4

				for _, w := range weights[x] {
					i := w.index * 4
					buf[0] = float32(scanLine[i])
					buf[1] = float32(scanLine[i+1])
					buf[2] = float32(scanLine[i+2])
					buf[3] = float32(scanLine[i+3])
					pixel := archsimd.LoadFloat32x4(&buf)
					aw := float32(scanLine[i+3]) * float32(w.weight)
					awVec := archsimd.BroadcastFloat32x4(aw)
					product := pixel.Mul(awVec)
					product = product.SetElem(3, aw)
					sum = sum.Add(product)
				}

				a := sum.GetElem(3)
				if a != 0 {
					aInv := float32(1.0) / a
					j := j0 + x*4
					dst.Pix[j] = clamp(float64(sum.GetElem(0) * aInv))
					dst.Pix[j+1] = clamp(float64(sum.GetElem(1) * aInv))
					dst.Pix[j+2] = clamp(float64(sum.GetElem(2) * aInv))
					dst.Pix[j+3] = clamp(float64(a))
				}
			}
		}
	})
	return dst
}

// resizeVertical resizes the image vertically using SIMD-accelerated accumulation.
// Uses Float32x4 to process RGBA channels in parallel on amd64 with AVX/SSE.
func resizeVertical(img image.Image, height int, filter ResampleFilter) *image.NRGBA {
	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, src.w, height))
	weights := precomputeWeights(height, src.h, filter)

	parallel(0, src.w, func(xs <-chan int) {
		scanLine := make([]uint8, src.h*4)
		var buf [4]float32

		for x := range xs {
			src.scan(x, 0, x+1, src.h, scanLine)

			for y := range weights {
				var sum archsimd.Float32x4

				for _, w := range weights[y] {
					i := w.index * 4
					buf[0] = float32(scanLine[i])
					buf[1] = float32(scanLine[i+1])
					buf[2] = float32(scanLine[i+2])
					buf[3] = float32(scanLine[i+3])
					pixel := archsimd.LoadFloat32x4(&buf)
					aw := float32(scanLine[i+3]) * float32(w.weight)
					awVec := archsimd.BroadcastFloat32x4(aw)
					product := pixel.Mul(awVec)
					product = product.SetElem(3, aw)
					sum = sum.Add(product)
				}

				a := sum.GetElem(3)
				if a != 0 {
					aInv := float32(1.0) / a
					j := y*dst.Stride + x*4
					dst.Pix[j] = clamp(float64(sum.GetElem(0) * aInv))
					dst.Pix[j+1] = clamp(float64(sum.GetElem(1) * aInv))
					dst.Pix[j+2] = clamp(float64(sum.GetElem(2) * aInv))
					dst.Pix[j+3] = clamp(float64(a))
				}
			}
		}
	})
	return dst
}
