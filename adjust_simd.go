//go:build goexperiment.simd && amd64

package imaging

import (
	"image"
	"simd/archsimd"
)

// adjustBrightnessImpl applies brightness adjustment using SIMD.
// Processes 2 NRGBA pixels (8 float32 channels) at a time via Float32x8.
func adjustBrightnessImpl(img *image.NRGBA, percentage float64) *image.NRGBA {
	b := img.Bounds()
	dst := image.NewNRGBA(b)
	shift := float32(percentage * 2.55)

	// Build shift vector: [shift,shift,shift,0, shift,shift,shift,0]
	// The 0.0 for alpha positions ensures alpha is not modified.
	var shiftArr [8]float32
	shiftArr[0] = shift
	shiftArr[1] = shift
	shiftArr[2] = shift
	shiftArr[3] = 0
	shiftArr[4] = shift
	shiftArr[5] = shift
	shiftArr[6] = shift
	shiftArr[7] = 0
	shiftVec := archsimd.LoadFloat32x8(&shiftArr)

	var zeroArr [8]float32
	zeroVec := archsimd.LoadFloat32x8(&zeroArr)
	var maxArr [8]float32
	for i := range maxArr {
		maxArr[i] = 255
	}
	maxVec := archsimd.LoadFloat32x8(&maxArr)

	parallel(b.Min.Y, b.Max.Y, func(ys <-chan int) {
		for y := range ys {
			sOff := img.PixOffset(b.Min.X, y)
			dOff := dst.PixOffset(b.Min.X, y)
			rowEnd := sOff + b.Dx()*4

			// Process 2 pixels at a time (8 bytes = 8 float32)
			for ; sOff+8 <= rowEnd; sOff += 8 {
				var pixel [8]float32
				pixel[0] = float32(img.Pix[sOff+0])
				pixel[1] = float32(img.Pix[sOff+1])
				pixel[2] = float32(img.Pix[sOff+2])
				pixel[3] = float32(img.Pix[sOff+3])
				pixel[4] = float32(img.Pix[sOff+4])
				pixel[5] = float32(img.Pix[sOff+5])
				pixel[6] = float32(img.Pix[sOff+6])
				pixel[7] = float32(img.Pix[sOff+7])

				pixelVec := archsimd.LoadFloat32x8(&pixel)
				result := pixelVec.Add(shiftVec)
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

			// Scalar tail for remaining pixels
			for ; sOff < rowEnd; sOff += 4 {
				dst.Pix[dOff+0] = clampUint8(float32(img.Pix[sOff+0]) + shift)
				dst.Pix[dOff+1] = clampUint8(float32(img.Pix[sOff+1]) + shift)
				dst.Pix[dOff+2] = clampUint8(float32(img.Pix[sOff+2]) + shift)
				dst.Pix[dOff+3] = img.Pix[sOff+3]
				dOff += 4
			}
		}
	})
	return dst
}

func clampUint8(v float32) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v + 0.5)
}
