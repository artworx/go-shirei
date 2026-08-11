package shirei

// swizzleOpaqueRGBA converts premultiplied opaque RGBA pixels to the BGRA
// layout used by the software framebuffer. Source and destination must contain
// whole four-byte pixels; they may not overlap.
func swizzleOpaqueRGBA(dst, src []byte) {
	n := min(len(dst), len(src)) &^ 3
	if n == 0 {
		return
	}
	archN := n &^ 15
	if archN > 0 {
		swizzleOpaqueRGBAArch(&dst[0], &src[0], archN)
	}
	for offset := archN; offset < n; offset += 4 {
		dst[offset] = src[offset+2]
		dst[offset+1] = src[offset+1]
		dst[offset+2] = src[offset]
		dst[offset+3] = 255
	}
}
