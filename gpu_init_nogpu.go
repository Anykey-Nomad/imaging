//go:build !gpu || !amd64

package imaging

// InitGPU is a no-op when GPU support is disabled.
func InitGPU() {}
