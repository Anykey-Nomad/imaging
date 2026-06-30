//go:build goexperiment.simd && amd64

package imaging

import (
	"image"
	"math"
	"simd/archsimd"
)

// Overlay blends the img onto the background at the given position with the given opacity.
// SIMD-accelerated version using Float32x8 to process 2 pixels (RGBA+RGBA) simultaneously.
func Overlay(background, img image.Image, pos image.Point, opacity float64) *image.NRGBA {
	opacity = math.Min(math.Max(opacity, 0.0), 1.0)
	dst := Clone(background)
	pos = pos.Sub(background.Bounds().Min)
	pasteRect := image.Rectangle{Min: pos, Max: pos.Add(img.Bounds().Size())}
	interRect := pasteRect.Intersect(dst.Bounds())
	if interRect.Empty() {
		return dst
	}
	src := newScanner(img)
	opF32 := float32(opacity)

	parallel(interRect.Min.Y, interRect.Max.Y, func(ys <-chan int) {
		scanLine := make([]uint8, interRect.Dx()*4)
		for y := range ys {
			x1 := interRect.Min.X - pasteRect.Min.X
			x2 := interRect.Max.X - pasteRect.Min.X
			y1 := y - pasteRect.Min.Y
			y2 := y1 + 1
			src.scan(x1, y1, x2, y2, scanLine)
			i := y*dst.Stride + interRect.Min.X*4
			j := 0
			n := interRect.Dx()

			// Process 2 pixels at a time with Float32x8
			px := 0
			for ; px+1 < n; px += 2 {
				// Load 2 bg pixels and 2 fg pixels as raw uint8
				bgR0, bgG0, bgB0, bgA0 := dst.Pix[i], dst.Pix[i+1], dst.Pix[i+2], dst.Pix[i+3]
				bgR1, bgG1, bgB1, bgA1 := dst.Pix[i+4], dst.Pix[i+5], dst.Pix[i+6], dst.Pix[i+7]
				fgR0, fgG0, fgB0, fgA0 := scanLine[j], scanLine[j+1], scanLine[j+2], scanLine[j+3]
				fgR1, fgG1, fgB1, fgA1 := scanLine[j+4], scanLine[j+5], scanLine[j+6], scanLine[j+7]

				// Compute blend coefficients (matching original algorithm exactly)
				// Original: coef2 = opacity * a2 / 255; coef1 = (1-coef2) * a1 / 255
				coef2_0 := opF32 * float32(fgA0) / 255.0
				coef1_0 := (1.0 - coef2_0) * float32(bgA0) / 255.0
				coefSum_0 := coef1_0 + coef2_0
				if coefSum_0 > 0 {
					coef1_0 /= coefSum_0
					coef2_0 /= coefSum_0
				}
				coef2_1 := opF32 * float32(fgA1) / 255.0
				coef1_1 := (1.0 - coef2_1) * float32(bgA1) / 255.0
				coefSum_1 := coef1_1 + coef2_1
				if coefSum_1 > 0 {
					coef1_1 /= coefSum_1
					coef2_1 /= coefSum_1
				}

				// Build coefficient vectors: [coef1,coef1,coef1,0, coef1,coef1,coef1,0]
				c1Buf := [8]float32{coef1_0, coef1_0, coef1_0, 0, coef1_1, coef1_1, coef1_1, 0}
				c2Buf := [8]float32{coef2_0, coef2_0, coef2_0, 0, coef2_1, coef2_1, coef2_1, 0}

				// Load bg and fg as float32 for SIMD blend
				bgBuf := [8]float32{
					float32(bgR0), float32(bgG0), float32(bgB0), 0,
					float32(bgR1), float32(bgG1), float32(bgB1), 0,
				}
				fgBuf := [8]float32{
					float32(fgR0), float32(fgG0), float32(fgB0), 0,
					float32(fgR1), float32(fgG1), float32(fgB1), 0,
				}

				// SIMD: result = bg * coef1 + fg * coef2
				bgVec := archsimd.LoadFloat32x8(&bgBuf)
				fgVec := archsimd.LoadFloat32x8(&fgBuf)
				c1Vec := archsimd.LoadFloat32x8(&c1Buf)
				c2Vec := archsimd.LoadFloat32x8(&c2Buf)
				resultVec := bgVec.Mul(c1Vec).Add(fgVec.Mul(c2Vec))
				var result [8]float32
				resultVec.Store(&result)

				// Store R,G,B for both pixels (truncate to match original uint8 behavior)
				dst.Pix[i] = uint8(result[0])
				dst.Pix[i+1] = uint8(result[1])
				dst.Pix[i+2] = uint8(result[2])
				// Alpha: original formula a1 + a2*opacity*(255-a1)/255
				dst.Pix[i+3] = uint8(math.Min(float64(bgA0)+float64(fgA0)*opacity*(255.0-float64(bgA0))/255.0, 255.0))
				dst.Pix[i+4] = uint8(result[4])
				dst.Pix[i+5] = uint8(result[5])
				dst.Pix[i+6] = uint8(result[6])
				dst.Pix[i+7] = uint8(math.Min(float64(bgA1)+float64(fgA1)*opacity*(255.0-float64(bgA1))/255.0, 255.0))

				i += 8
				j += 8
			}

			// Handle remaining odd pixel (scalar, exact original algorithm)
			for ; px < n; px++ {
				d := dst.Pix[i : i+4 : i+4]
				r1 := float64(d[0])
				g1 := float64(d[1])
				b1 := float64(d[2])
				a1 := float64(d[3])
				s := scanLine[j : j+4 : j+4]
				r2 := float64(s[0])
				g2 := float64(s[1])
				b2 := float64(s[2])
				a2 := float64(s[3])
				coef2 := opacity * a2 / 255
				coef1 := (1 - coef2) * a1 / 255
				coefSum := coef1 + coef2
				coef1 /= coefSum
				coef2 /= coefSum
				d[0] = uint8(r1*coef1 + r2*coef2)
				d[1] = uint8(g1*coef1 + g2*coef2)
				d[2] = uint8(b1*coef1 + b2*coef2)
				d[3] = uint8(math.Min(a1+a2*opacity*(255-a1)/255, 255))
				i += 4
				j += 4
			}
		}
	})
	return dst
}
