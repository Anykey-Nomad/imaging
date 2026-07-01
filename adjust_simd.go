//go:build goexperiment.simd && amd64

package imaging

import (
	"image"
)

func adjustBrightnessImpl(img *image.NRGBA, percentage float64) *image.NRGBA {
	b := img.Bounds()
	dst := image.NewNRGBA(b)

	// 1. Конвертируем процент в int8 сдвиг
	shift := int8(percentage * 2.55)

	// 2. ОБЪЯВЛЯЕМ переменные ЗДЕСЬ
	isSub := shift < 0

	if shift == 0 {
		copy(dst.Pix, img.Pix)
		return dst
	}

	// Готовим 32-байтный вектор сдвига
	var shiftVec [32]byte
	absShift := byte(abs(int(shift))) // используем абсолютное значение для вектора
	for i := 0; i < 32; i += 4 {
		shiftVec[i] = absShift   // R
		shiftVec[i+1] = absShift // G
		shiftVec[i+2] = absShift // B
		shiftVec[i+3] = 0        // A
	}

	parallel(b.Min.Y, b.Max.Y, func(ys <-chan int) {
		for y := range ys {
			sOff := img.PixOffset(b.Min.X, y)
			dOff := dst.PixOffset(b.Min.X, y)
			rowLen := b.Dx() * 4

			// Теперь isSub определена и доступна внутри функции
			adjustBrightnessAVX2(dst.Pix[dOff:dOff+rowLen], img.Pix[sOff:sOff+rowLen], &shiftVec, isSub)
		}
	})
	return dst
}

// Вспомогательная функция для модуля
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Заглушка для ассемблера (должна быть вне функций)
func adjustBrightnessAVX2(dst, src []byte, shiftVec *[32]byte, isSub bool)
