//go:build goexperiment.simd && amd64

package imaging

import (
	"image"
	"simd/archsimd"
)

// applyUnsharpMask computes sharpen = clamp(2*original - blurred) using SIMD.
// Processes 2 pixels (8 float32 channels) at a time.
func applyUnsharpMask(src, blurred *image.NRGBA, dst *image.NRGBA) {
	b := src.Bounds()

	var twoArr [8]float32
	for i := range twoArr {
		twoArr[i] = 2
	}
	twoVec := archsimd.LoadFloat32x8(&twoArr)

	var zeroArr [8]float32
	zeroVec := archsimd.LoadFloat32x8(&zeroArr)
	var maxArr [8]float32
	for i := range maxArr {
		maxArr[i] = 255
	}
	maxVec := archsimd.LoadFloat32x8(&maxArr)

	parallel(b.Min.Y, b.Max.Y, func(ys <-chan int) {
		for y := range ys {
			sOff := src.PixOffset(b.Min.X, y)
			dOff := dst.PixOffset(b.Min.X, y)
			rowEnd := sOff + b.Dx()*4

			for ; sOff+8 <= rowEnd; sOff += 8 {
				var orig [8]float32
				orig[0] = float32(src.Pix[sOff+0])
				orig[1] = float32(src.Pix[sOff+1])
				orig[2] = float32(src.Pix[sOff+2])
				orig[3] = float32(src.Pix[sOff+3])
				orig[4] = float32(src.Pix[sOff+4])
				orig[5] = float32(src.Pix[sOff+5])
				orig[6] = float32(src.Pix[sOff+6])
				orig[7] = float32(src.Pix[sOff+7])

				var blur [8]float32
				blur[0] = float32(blurred.Pix[sOff+0])
				blur[1] = float32(blurred.Pix[sOff+1])
				blur[2] = float32(blurred.Pix[sOff+2])
				blur[3] = float32(blurred.Pix[sOff+3])
				blur[4] = float32(blurred.Pix[sOff+4])
				blur[5] = float32(blurred.Pix[sOff+5])
				blur[6] = float32(blurred.Pix[sOff+6])
				blur[7] = float32(blurred.Pix[sOff+7])

				origVec := archsimd.LoadFloat32x8(&orig)
				blurVec := archsimd.LoadFloat32x8(&blur)

				// sharpen = clamp(2*original - blurred)
				result := origVec.Mul(twoVec).Sub(blurVec)
				result = result.Max(zeroVec).Min(maxVec)

				var out [8]float32
				result.Store(&out)
				dst.Pix[dOff+0] = uint8(out[0])
				dst.Pix[dOff+1] = uint8(out[1])
				dst.Pix[dOff+2] = uint8(out[2])
				dst.Pix[dOff+3] = uint8(out[3])
				dst.Pix[dOff+4] = uint8(out[4])
				dst.Pix[dOff+5] = uint8(out[5])
				dst.Pix[dOff+6] = uint8(out[6])
				dst.Pix[dOff+7] = uint8(out[7])
				dOff += 8
			}

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
