//go:build goexperiment.simd && amd64

package imaging

import (
	"image"
	"simd/archsimd"
)

func blurHorizontal(img image.Image, kernel []float64) *image.NRGBA {
	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, src.w, src.h))
	radius := len(kernel) - 1
	parallel(0, src.h, func(ys <-chan int) {
		scanLine := make([]uint8, src.w*4)
		for y := range ys {
			src.scan(0, y, src.w, y+1, scanLine)
			x := 0
			for ; x+1 < src.w; x += 2 {
				var sumVec archsimd.Float32x8
				var wsum0, wsum1 float64
				min := x - radius
				if min < 0 {
					min = 0
				}
				max := x + 1 + radius
				if max > src.w-1 {
					max = src.w - 1
				}
				for ix := min; ix <= max; ix++ {
					i := ix * 4
					d0 := absint(x - ix)
					d1 := absint(x + 1 - ix)
					var w0, w1 float32
					if d0 < len(kernel) {
						w0 = float32(kernel[d0])
					}
					if d1 < len(kernel) {
						w1 = float32(kernel[d1])
					}
					a := float32(scanLine[i+3])
					aw0 := a * w0
					aw1 := a * w1
					pb := [8]float32{
						float32(scanLine[i]), float32(scanLine[i+1]), float32(scanLine[i+2]), a,
						float32(scanLine[i]), float32(scanLine[i+1]), float32(scanLine[i+2]), a,
					}
					wb := [8]float32{aw0, aw0, aw0, w0, aw1, aw1, aw1, w1}
					pv := archsimd.LoadFloat32x8(&pb)
					wv := archsimd.LoadFloat32x8(&wb)
					sumVec = sumVec.Add(pv.Mul(wv))
					wsum0 += float64(w0)
					wsum1 += float64(w1)
				}
				var result [8]float32
				sumVec.Store(&result)
				a0 := float64(result[3])
				if a0 != 0 {
					j := y*dst.Stride + x*4
					d := dst.Pix[j : j+4 : j+4]
					ai := 1.0 / a0
					d[0] = clamp(float64(result[0]) * ai)
					d[1] = clamp(float64(result[1]) * ai)
					d[2] = clamp(float64(result[2]) * ai)
					d[3] = clamp(a0 / wsum0)
				}
				a1 := float64(result[7])
				if a1 != 0 {
					j := y*dst.Stride + (x+1)*4
					d := dst.Pix[j : j+4 : j+4]
					ai := 1.0 / a1
					d[0] = clamp(float64(result[4]) * ai)
					d[1] = clamp(float64(result[5]) * ai)
					d[2] = clamp(float64(result[6]) * ai)
					d[3] = clamp(a1 / wsum1)
				}
			}
			for ; x < src.w; x++ {
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
					wa := float64(scanLine[i+3]) * weight
					r += float64(scanLine[i]) * wa
					g += float64(scanLine[i+1]) * wa
					b += float64(scanLine[i+2]) * wa
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

func blurVertical(img image.Image, kernel []float64) *image.NRGBA {
	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, src.w, src.h))
	radius := len(kernel) - 1
	parallel(0, src.w, func(xs <-chan int) {
		scanLine := make([]uint8, src.h*4)
		for x := range xs {
			src.scan(x, 0, x+1, src.h, scanLine)
			y := 0
			for ; y+1 < src.h; y += 2 {
				var sumVec archsimd.Float32x8
				var wsum0, wsum1 float64
				min := y - radius
				if min < 0 {
					min = 0
				}
				max := y + 1 + radius
				if max > src.h-1 {
					max = src.h - 1
				}
				for iy := min; iy <= max; iy++ {
					i := iy * 4
					d0 := absint(y - iy)
					d1 := absint(y + 1 - iy)
					var w0, w1 float32
					if d0 < len(kernel) {
						w0 = float32(kernel[d0])
					}
					if d1 < len(kernel) {
						w1 = float32(kernel[d1])
					}
					a := float32(scanLine[i+3])
					aw0 := a * w0
					aw1 := a * w1
					pb := [8]float32{
						float32(scanLine[i]), float32(scanLine[i+1]), float32(scanLine[i+2]), a,
						float32(scanLine[i]), float32(scanLine[i+1]), float32(scanLine[i+2]), a,
					}
					wb := [8]float32{aw0, aw0, aw0, w0, aw1, aw1, aw1, w1}
					pv := archsimd.LoadFloat32x8(&pb)
					wv := archsimd.LoadFloat32x8(&wb)
					sumVec = sumVec.Add(pv.Mul(wv))
					wsum0 += float64(w0)
					wsum1 += float64(w1)
				}
				var result [8]float32
				sumVec.Store(&result)
				a0 := float64(result[3])
				if a0 != 0 {
					j := y*dst.Stride + x*4
					d := dst.Pix[j : j+4 : j+4]
					ai := 1.0 / a0
					d[0] = clamp(float64(result[0]) * ai)
					d[1] = clamp(float64(result[1]) * ai)
					d[2] = clamp(float64(result[2]) * ai)
					d[3] = clamp(a0 / wsum0)
				}
				a1 := float64(result[7])
				if a1 != 0 {
					j := (y + 1) * dst.Stride + x*4
					d := dst.Pix[j : j+4 : j+4]
					ai := 1.0 / a1
					d[0] = clamp(float64(result[4]) * ai)
					d[1] = clamp(float64(result[5]) * ai)
					d[2] = clamp(float64(result[6]) * ai)
					d[3] = clamp(a1 / wsum1)
				}
			}
			for ; y < src.h; y++ {
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
					wa := float64(scanLine[i+3]) * weight
					r += float64(scanLine[i]) * wa
					g += float64(scanLine[i+1]) * wa
					b += float64(scanLine[i+2]) * wa
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

func Grayscale(img image.Image) *image.NRGBA {
	// Try GPU-accelerated grayscale first (BT.601 via WebGPU compute shader).
	if nrgba, ok := img.(*image.NRGBA); ok {
		if gpuResult := tryGPUGrayscale(nrgba); gpuResult != nil {
			return gpuResult
		}
	}

	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, src.w, src.h))
	weights := [8]float32{0.299, 0.587, 0.114, 0, 0.299, 0.587, 0.114, 0}
	wv := archsimd.LoadFloat32x8(&weights)
	parallel(0, src.h, func(ys <-chan int) {
		scanLine := make([]uint8, src.w*4)
		for y := range ys {
			src.scan(0, y, src.w, y+1, scanLine)
			j := y * dst.Stride
			x := 0
			for ; x+1 < src.w; x += 2 {
				si := x * 4
				pb := [8]float32{
					float32(scanLine[si]), float32(scanLine[si+1]), float32(scanLine[si+2]), float32(scanLine[si+3]),
					float32(scanLine[si+4]), float32(scanLine[si+5]), float32(scanLine[si+6]), float32(scanLine[si+7]),
				}
				pv := archsimd.LoadFloat32x8(&pb)
				rv := pv.Mul(wv)
				var result [8]float32
				rv.Store(&result)
				g0 := uint8(result[0] + result[1] + result[2] + 0.5)
				g1 := uint8(result[4] + result[5] + result[6] + 0.5)
				dst.Pix[j] = g0
				dst.Pix[j+1] = g0
				dst.Pix[j+2] = g0
				dst.Pix[j+3] = scanLine[si+3]
				dst.Pix[j+4] = g1
				dst.Pix[j+5] = g1
				dst.Pix[j+6] = g1
				dst.Pix[j+7] = scanLine[si+7]
				j += 8
			}
			for ; x < src.w; x++ {
				si := x * 4
				r := scanLine[si]
				g := scanLine[si+1]
				b := scanLine[si+2]
				f := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
				c := uint8(f + 0.5)
				dst.Pix[j] = c
				dst.Pix[j+1] = c
				dst.Pix[j+2] = c
				dst.Pix[j+3] = scanLine[si+3]
				j += 4
			}
		}
	})
	return dst
}
