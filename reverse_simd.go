//go:build goexperiment.simd && amd64

package imaging

import (
	"unsafe"
	"simd/archsimd"
)

// reverse reverses the order of 4-byte NRGBA pixels in the given byte slice.
// SIMD version: swaps 8 pixels (32 bytes) at a time using Float32x8.
// Each float32 covers exactly one 4-byte pixel (R,G,B,A), so the
// byte-level representation is preserved through Load/Store.
func reverse(pix []uint8) {
	n := len(pix)
	if n <= 4 {
		return
	}

	chunkBytes := 32 // 8 pixels * 4 bytes
	left := 0
	right := n

	// SIMD: swap 8-pixel blocks from both ends
	for left+chunkBytes <= right-chunkBytes {
		leftPtr := (*[8]float32)(unsafe.Pointer(&pix[left]))
		rightPtr := (*[8]float32)(unsafe.Pointer(&pix[right-chunkBytes]))

		leftVec := archsimd.LoadFloat32x8(leftPtr)
		rightVec := archsimd.LoadFloat32x8(rightPtr)

		// Swap: store left data to right position, right data to left position
		leftVec.Store(rightPtr)
		rightVec.Store(leftPtr)

		left += chunkBytes
		right -= chunkBytes
	}

	// Scalar: handle remaining middle pixels
	for left < right-4 {
		pi := pix[left : left+4 : left+4]
		pj := pix[right-4 : right : right]
		pi[0], pj[0] = pj[0], pi[0]
		pi[1], pj[1] = pj[1], pi[1]
		pi[2], pj[2] = pj[2], pi[2]
		pi[3], pj[3] = pj[3], pi[3]
		left += 4
		right -= 4
	}
}
