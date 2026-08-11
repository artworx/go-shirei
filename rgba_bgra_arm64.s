#include "textflag.h"

// Four copies of the RGBA -> BGRA byte permutation, one per 16-byte vector.
DATA ·rgbaToBGRAIndex+0(SB)/8, $0x0704050603000102
DATA ·rgbaToBGRAIndex+8(SB)/8, $0x0f0c0d0e0b08090a
GLOBL ·rgbaToBGRAIndex(SB), RODATA|NOPTR, $16

// func swizzleOpaqueRGBAArch(dst, src *byte, n int)
// n is a positive multiple of 16.
TEXT ·swizzleOpaqueRGBAArch(SB), NOSPLIT|NOFRAME, $0
	MOVD dst+0(FP), R0
	MOVD src+8(FP), R1
	MOVD n+16(FP), R2
	MOVD $·rgbaToBGRAIndex(SB), R3
	VLD1 (R3), [V31.B16]

	CMP $64, R2
	BLT tail
loop64:
	VLD1.P 64(R1), [V0.B16, V1.B16, V2.B16, V3.B16]
	VTBL V31.B16, [V0.B16], V0.B16
	VTBL V31.B16, [V1.B16], V1.B16
	VTBL V31.B16, [V2.B16], V2.B16
	VTBL V31.B16, [V3.B16], V3.B16
	VST1.P [V0.B16, V1.B16, V2.B16, V3.B16], 64(R0)
	SUBS $64, R2
	CMP $64, R2
	BGE loop64

tail:
	CBZ R2, done
loop16:
	VLD1.P 16(R1), [V0.B16]
	VTBL V31.B16, [V0.B16], V0.B16
	VST1.P [V0.B16], 16(R0)
	SUBS $16, R2
	BNE loop16
done:
	RET
