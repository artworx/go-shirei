//go:build !arm64

package shirei

import "unsafe"

func swizzleOpaqueRGBAArch(dst, src *byte, n int) {
	destination := unsafe.Slice(dst, n)
	source := unsafe.Slice(src, n)
	for offset := 0; offset < n; offset += 4 {
		destination[offset] = source[offset+2]
		destination[offset+1] = source[offset+1]
		destination[offset+2] = source[offset]
		destination[offset+3] = 255
	}
}
