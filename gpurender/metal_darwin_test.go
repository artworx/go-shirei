//go:build darwin && !ios

package gpurender

import (
	"image"
	"testing"

	"go.hasen.dev/shirei"
)

func TestMetalTopLeftFill(t *testing.T) {
	if err := Init(); err != nil {
		t.Skipf("Metal init: %v", err)
	}
	const w, h = 32, 32
	surf := testSurface(w, h)
	if surf == nil {
		t.Fatal("test IOSurface is nil")
	}
	defer testRelease(surf)

	red := shirei.Vec4{0, 100, 50, 1}
	surfaces := []shirei.Surface{{
		Rect:   shirei.Rect{Size: shirei.Vec2{10, 10}},
		Color1: red, Color2: red,
	}}
	if err := Render(surf, w, h, 1, surfaces, nil, nil, nil, true); err != nil {
		t.Fatal(err)
	}

	tl := testPixel(surf, 0, 0)
	// BGRA, opaque red ≈ (0, 0, 255, 255)
	if int(tl[2]) < 200 || int(tl[1]) > 40 || int(tl[0]) > 40 || tl[3] < 200 {
		t.Fatalf("top-left BGRA=%v want red", tl)
	}
	br := testPixel(surf, w-1, h-1)
	// cleared white
	if br[0] < 200 || br[1] < 200 || br[2] < 200 || br[3] < 200 {
		t.Fatalf("bottom-right BGRA=%v want white (Y flip?)", br)
	}
	// A pixel just below the 10x10 stamp should still be white if y-down is correct.
	below := testPixel(surf, 0, 20)
	if below[0] < 200 || below[1] < 200 || below[2] < 200 {
		t.Fatalf("pixel (0,20) BGRA=%v want white", below)
	}
}

func TestMetalRectClip(t *testing.T) {
	if err := Init(); err != nil {
		t.Skipf("Metal init: %v", err)
	}
	const w, h = 32, 32
	surf := testSurface(w, h)
	if surf == nil {
		t.Fatal("test IOSurface is nil")
	}
	defer testRelease(surf)

	red := shirei.Vec4{0, 100, 50, 1}
	surfaces := []shirei.Surface{
		{
			Rect: shirei.Rect{Origin: shirei.Vec2{10, 10}, Size: shirei.Vec2{10, 10}},
			Clip: shirei.ClipPush,
		},
		{Rect: shirei.Rect{Size: shirei.Vec2{32, 32}}, Color1: red, Color2: red},
		{
			Rect: shirei.Rect{Origin: shirei.Vec2{10, 10}, Size: shirei.Vec2{10, 10}},
			Clip: shirei.ClipPop,
		},
	}
	if err := Render(surf, w, h, 1, surfaces, nil, nil, nil, true); err != nil {
		t.Fatal(err)
	}
	inside := testPixel(surf, 15, 15)
	if int(inside[2]) < 200 || inside[3] < 200 {
		t.Fatalf("clipped interior BGRA=%v want red", inside)
	}
	outside := testPixel(surf, 0, 0)
	if outside[0] < 200 || outside[1] < 200 || outside[2] < 200 {
		t.Fatalf("outside clip BGRA=%v want white", outside)
	}
}

func TestMetalTranslucentOverWhite(t *testing.T) {
	if err := Init(); err != nil {
		t.Skipf("Metal init: %v", err)
	}
	const w, h = 16, 16
	surf := testSurface(w, h)
	if surf == nil {
		t.Fatal("test IOSurface is nil")
	}
	defer testRelease(surf)

	blackHalf := shirei.Vec4{0, 0, 0, 0.5}
	surfaces := []shirei.Surface{
		{Rect: shirei.Rect{Size: shirei.Vec2{w, h}}, Color1: blackHalf, Color2: blackHalf},
	}
	if err := Render(surf, w, h, 1, surfaces, nil, nil, nil, true); err != nil {
		t.Fatal(err)
	}
	p := testPixel(surf, 8, 8)
	// premul black 50% over white → ~127 gray
	for i := 0; i < 3; i++ {
		if int(p[i]) < 90 || int(p[i]) > 160 {
			t.Fatalf("BGRA=%v want ~50%% gray over white", p)
		}
	}
}

func TestMetalRoundedClip(t *testing.T) {
	if err := Init(); err != nil {
		t.Skipf("Metal init: %v", err)
	}
	const w, h = 32, 32
	surf := testSurface(w, h)
	if surf == nil {
		t.Fatal("test IOSurface is nil")
	}
	defer testRelease(surf)

	red := shirei.Vec4{0, 100, 50, 1}
	surfaces := []shirei.Surface{
		{
			Rect:    shirei.Rect{Origin: shirei.Vec2{6, 6}, Size: shirei.Vec2{20, 20}},
			Corners: shirei.Vec4{10, 10, 10, 10},
			Clip:    shirei.ClipPush,
		},
		{Rect: shirei.Rect{Size: shirei.Vec2{32, 32}}, Color1: red, Color2: red},
		{
			Rect: shirei.Rect{Origin: shirei.Vec2{6, 6}, Size: shirei.Vec2{20, 20}},
			Clip: shirei.ClipPop,
		},
	}
	if err := Render(surf, w, h, 1, surfaces, nil, nil, nil, true); err != nil {
		t.Fatal(err)
	}
	cut := testPixel(surf, 6, 6)
	if cut[0] < 200 || cut[1] < 200 || cut[2] < 200 {
		t.Fatalf("rounded-clip outer corner BGRA=%v want white (scissor-only?)", cut)
	}
	inside := testPixel(surf, 16, 16)
	if int(inside[2]) < 200 || inside[3] < 200 {
		t.Fatalf("rounded-clip interior BGRA=%v want red", inside)
	}
}

func TestMetalRoundedCornerCut(t *testing.T) {
	if err := Init(); err != nil {
		t.Skipf("Metal init: %v", err)
	}
	const w, h = 32, 32
	surf := testSurface(w, h)
	if surf == nil {
		t.Fatal("test IOSurface is nil")
	}
	defer testRelease(surf)

	red := shirei.Vec4{0, 100, 50, 1}
	surfaces := []shirei.Surface{{
		Rect:    shirei.Rect{Origin: shirei.Vec2{6, 6}, Size: shirei.Vec2{20, 20}},
		Color1:  red,
		Color2:  red,
		Corners: shirei.Vec4{10, 10, 10, 10},
	}}
	if err := Render(surf, w, h, 1, surfaces, nil, nil, nil, true); err != nil {
		t.Fatal(err)
	}
	// Outer TL of the corner square is outside the quarter-disk → white.
	cut := testPixel(surf, 6, 6)
	if cut[0] < 200 || cut[1] < 200 || cut[2] < 200 {
		t.Fatalf("rounded outer corner BGRA=%v want white (still square?)", cut)
	}
	inside := testPixel(surf, 16, 16)
	if int(inside[2]) < 200 || inside[3] < 200 {
		t.Fatalf("rounded interior BGRA=%v want red", inside)
	}
}

// TestMetalGradientCircleSmooth renders a white→black vertical gradient on a
// full circle. A circle decomposes into four corner coverage stamps with no
// interior rects, so the gradient must survive the coverage path (mode 1):
// each stamp carries band colors that the shader interpolates per pixel. If
// the coverage path drops color2, the circle renders as two flat halves —
// which is exactly how the regression looked on slider/toggle knobs.
func TestMetalGradientCircleSmooth(t *testing.T) {
	if err := Init(); err != nil {
		t.Skipf("Metal init: %v", err)
	}
	const w, h = 32, 32
	surf := testSurface(w, h)
	if surf == nil {
		t.Fatal("test IOSurface is nil")
	}
	defer testRelease(surf)

	white := shirei.Vec4{0, 0, 100, 1}
	black := shirei.Vec4{0, 0, 0, 1}
	// Circle: 20×20 rect at (6,6) with radius 10 on every corner.
	surfaces := []shirei.Surface{{
		Rect:    shirei.Rect{Origin: shirei.Vec2{6, 6}, Size: shirei.Vec2{20, 20}},
		Color1:  white,
		Color2:  black,
		Corners: shirei.Vec4{10, 10, 10, 10},
	}}
	if err := Render(surf, w, h, 1, surfaces, nil, nil, nil, true); err != nil {
		t.Fatal(err)
	}

	// Down the center column the gradient t = (y+0.5-6)/20; the fill is gray
	// so any channel works. Expect ≈ 255·(1−t) within a loose tolerance.
	gray := func(y int) int { return int(testPixel(surf, 16, y)[0]) }
	checks := []struct{ y, want int }{
		{8, 223}, {12, 172}, {15, 134}, {17, 108}, {20, 70}, {24, 19},
	}
	for _, c := range checks {
		got := gray(c.y)
		if got < c.want-25 || got > c.want+25 {
			t.Errorf("center column y=%d gray=%d, want ≈%d (flat corner bands?)", c.y, got, c.want)
		}
	}
	// The two-flat-halves failure mode also shows as a hard step across the
	// vertical midpoint; the true gradient changes gently there.
	if step := gray(15) - gray(17); step > 40 || step < 0 {
		t.Errorf("gray step across midline = %d, want small positive (two flat bands?)", step)
	}
}

func TestMetalImage(t *testing.T) {
	if err := Init(); err != nil {
		t.Skipf("Metal init: %v", err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = 255
		img.Pix[i+1] = 0
		img.Pix[i+2] = 0
		img.Pix[i+3] = 255
	}
	id := shirei.UseImage("gpurender-test-red", img)
	if id == 0 {
		t.Fatal("UseImage returned 0")
	}
	const w, h = 32, 32
	surf := testSurface(w, h)
	if surf == nil {
		t.Fatal("test IOSurface is nil")
	}
	defer testRelease(surf)

	surfaces := []shirei.Surface{{
		Rect:    shirei.Rect{Size: shirei.Vec2{8, 8}},
		ImageId: id,
	}}
	if err := Render(surf, w, h, 1, surfaces, nil, nil, nil, true); err != nil {
		t.Fatal(err)
	}
	tl := testPixel(surf, 0, 0)
	if int(tl[2]) < 200 || int(tl[1]) > 40 || int(tl[0]) > 40 || tl[3] < 200 {
		t.Fatalf("image top-left BGRA=%v want red", tl)
	}
	outside := testPixel(surf, 20, 20)
	if outside[0] < 200 || outside[1] < 200 || outside[2] < 200 {
		t.Fatalf("outside image BGRA=%v want white", outside)
	}
}

func TestMetalTranslucentBlackImageOverWhite(t *testing.T) {
	if err := Init(); err != nil {
		t.Skipf("Metal init: %v", err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = 0
		img.Pix[i+1] = 0
		img.Pix[i+2] = 0
		img.Pix[i+3] = 31 // ~12% premul black
	}
	id := shirei.UseImage("gpurender-test-shadow-black", img)
	const w, h = 16, 16
	surf := testSurface(w, h)
	if surf == nil {
		t.Fatal("test IOSurface is nil")
	}
	defer testRelease(surf)

	surfaces := []shirei.Surface{{
		Rect:    shirei.Rect{Size: shirei.Vec2{8, 8}},
		ImageId: id,
	}}
	if err := Render(surf, w, h, 1, surfaces, nil, nil, nil, true); err != nil {
		t.Fatal(err)
	}
	p := testPixel(surf, 4, 4)
	// premul black 12% over white → ~224
	for i := 0; i < 3; i++ {
		if int(p[i]) < 200 || int(p[i]) > 240 {
			t.Fatalf("BGRA=%v want ~12%% darken of white (~224)", p)
		}
	}
	outside := testPixel(surf, 12, 12)
	if outside[0] < 200 || outside[1] < 200 || outside[2] < 200 {
		t.Fatalf("outside BGRA=%v want white", outside)
	}
}
