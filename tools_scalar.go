//go:build !goexperiment.simd || !amd64

package imaging

import (
	"image"
	"math"
)

// Overlay blends the img onto the background at the given position with the given opacity.
// Scalar fallback for non-SIMD builds.
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
			for x := interRect.Min.X; x < interRect.Max.X; x++ {
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
