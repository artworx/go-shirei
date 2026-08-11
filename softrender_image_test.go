package shirei

import (
	"bytes"
	"image"
	"testing"
)

var benchmarkBlitFramebuffer []byte

func TestUseOpaqueImageCarriesRendererPromise(t *testing.T) {
	pixels := image.NewRGBA(image.Rect(0, 0, 1, 1))
	pixels.Pix[3] = 255
	id := UseOpaqueImage(t.Name(), pixels)
	defer freeImage(id)
	data := LookupImage(id)
	if data == nil || !data.Opaque {
		t.Fatal("opaque image registration did not retain its renderer promise")
	}
}

func TestBlitOpaqueMatchesPremultipliedPathThroughRoundedClip(t *testing.T) {
	const width, height = 6, 3
	source := image.NewRGBA(image.Rect(0, 0, width, height))
	for offset := 0; offset < len(source.Pix); offset += 4 {
		source.Pix[offset] = byte(10 + offset)
		source.Pix[offset+1] = byte(20 + offset)
		source.Pix[offset+2] = byte(30 + offset)
		source.Pix[offset+3] = 255
	}
	mask := []byte{
		0, 128, 255, 255, 128, 0,
		255, 255, 255, 255, 255, 255,
		0, 128, 255, 255, 128, 0,
	}
	clip := clipState{
		rect: image.Rect(0, 0, width, height), mask: mask,
		corners: &clipCorners{cornerOnly: true, squares: [4]image.Rectangle{
			image.Rect(0, 0, 2, 1), image.Rect(4, 0, 6, 1),
			image.Rect(4, 2, 6, 3), image.Rect(0, 2, 2, 3),
		}},
	}
	background := make([]byte, width*height*4)
	for offset := 0; offset < len(background); offset += 4 {
		background[offset], background[offset+1], background[offset+2], background[offset+3] = 5, 7, 9, 255
	}
	optimized := SoftRenderer{
		fb:   Framebuffer{W: width, H: height, Stride: width * 4, Pix: append([]byte(nil), background...)},
		clip: clip, alpha: 1,
	}
	reference := SoftRenderer{
		fb:   Framebuffer{W: width, H: height, Stride: width * 4, Pix: append([]byte(nil), background...)},
		clip: clip, alpha: 1,
	}
	optimized.blitOpaque(source.Bounds(), source)
	reference.blitPremul(source.Bounds(), source, false)
	if !bytes.Equal(optimized.fb.Pix, reference.fb.Pix) {
		t.Fatalf("opaque rounded blit = %v, want %v", optimized.fb.Pix, reference.fb.Pix)
	}
}

func BenchmarkBlitPremulOpaqueDiagram(b *testing.B) {
	const width, height = 1600, 1000
	source := image.NewRGBA(image.Rect(0, 0, width, height))
	for offset := 0; offset < len(source.Pix); offset += 4 {
		source.Pix[offset] = byte(offset)
		source.Pix[offset+1] = byte(offset >> 3)
		source.Pix[offset+2] = byte(offset >> 7)
		source.Pix[offset+3] = 255
	}
	renderer := SoftRenderer{
		fb:    Framebuffer{W: width, H: height, Stride: width * 4, Pix: make([]byte, width*height*4)},
		clip:  clipState{rect: image.Rect(0, 0, width, height)},
		alpha: 1,
	}
	b.SetBytes(width * height * 4)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		renderer.blitPremul(image.Rect(0, 0, width, height), source, true)
	}
	benchmarkBlitFramebuffer = renderer.fb.Pix
}

func BenchmarkBlitOpaqueDiagram(b *testing.B) {
	const width, height = 1600, 1000
	source := image.NewRGBA(image.Rect(0, 0, width, height))
	for offset := 0; offset < len(source.Pix); offset += 4 {
		source.Pix[offset] = byte(offset)
		source.Pix[offset+1] = byte(offset >> 3)
		source.Pix[offset+2] = byte(offset >> 7)
		source.Pix[offset+3] = 255
	}
	renderer := SoftRenderer{
		fb:    Framebuffer{W: width, H: height, Stride: width * 4, Pix: make([]byte, width*height*4)},
		clip:  clipState{rect: image.Rect(0, 0, width, height)},
		alpha: 1,
	}
	b.SetBytes(width * height * 4)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		renderer.blitOpaque(image.Rect(0, 0, width, height), source)
	}
	benchmarkBlitFramebuffer = renderer.fb.Pix
}

func BenchmarkBlitOpaqueDiagramRoundedClip(b *testing.B) {
	const width, height, radius = 1600, 1000, 10
	source := image.NewRGBA(image.Rect(0, 0, width, height))
	for offset := 0; offset < len(source.Pix); offset += 4 {
		source.Pix[offset] = byte(offset)
		source.Pix[offset+1] = byte(offset >> 3)
		source.Pix[offset+2] = byte(offset >> 7)
		source.Pix[offset+3] = 255
	}
	mask := make([]byte, width*height)
	for index := range mask {
		mask[index] = 255
	}
	for y := 0; y < radius; y++ {
		for x := 0; x < radius-y; x++ {
			mask[y*width+x] = byte(255 * x / radius)
			mask[y*width+width-1-x] = byte(255 * x / radius)
			mask[(height-1-y)*width+x] = byte(255 * x / radius)
			mask[(height-1-y)*width+width-1-x] = byte(255 * x / radius)
		}
	}
	clip := clipState{
		rect: image.Rect(0, 0, width, height), mask: mask,
		corners: &clipCorners{cornerOnly: true, squares: [4]image.Rectangle{
			image.Rect(0, 0, radius, radius), image.Rect(width-radius, 0, width, radius),
			image.Rect(width-radius, height-radius, width, height), image.Rect(0, height-radius, radius, height),
		}},
	}
	for _, benchmark := range []struct {
		name string
		blit func(*SoftRenderer, image.Rectangle, *image.RGBA)
	}{
		{name: "General", blit: func(r *SoftRenderer, d image.Rectangle, s *image.RGBA) { r.blitPremul(d, s, false) }},
		{name: "OpaqueCopy", blit: (*SoftRenderer).blitOpaque},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			renderer := SoftRenderer{
				fb:   Framebuffer{W: width, H: height, Stride: width * 4, Pix: make([]byte, width*height*4)},
				clip: clip, alpha: 1,
			}
			b.SetBytes(width * height * 4)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				benchmark.blit(&renderer, source.Bounds(), source)
			}
			benchmarkBlitFramebuffer = renderer.fb.Pix
		})
	}
}

func TestBlitPremulHandlesOpaqueAndTranslucentPixels(t *testing.T) {
	renderer := SoftRenderer{
		fb: Framebuffer{
			W: 2, H: 1, Stride: 8,
			Pix: []byte{4, 8, 12, 255, 4, 8, 12, 255},
		},
		clip:  clipState{rect: image.Rect(0, 0, 2, 1)},
		alpha: 1,
	}
	source := &image.RGBA{
		Pix:    []byte{10, 20, 30, 255, 64, 32, 16, 128},
		Stride: 8,
		Rect:   image.Rect(0, 0, 2, 1),
	}

	renderer.blitPremul(image.Rect(0, 0, 2, 1), source, false)

	want := []byte{10, 20, 30, 255, 65, 35, 21, 255}
	for index := range want {
		if renderer.fb.Pix[index] != want[index] {
			t.Fatalf("pixel bytes = %v, want %v", renderer.fb.Pix, want)
		}
	}
}
