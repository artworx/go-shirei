//go:build arm64

package shirei

//go:noescape
func swizzleOpaqueRGBAArch(dst, src *byte, n int)
