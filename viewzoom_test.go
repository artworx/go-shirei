package shirei

import "testing"

func TestViewZoomLayoutPointerAndPublishedCoordinates(t *testing.T) {
	previousSize, previousScale, previousZoom := ui.Host.WindowSize, ui.Host.WindowScale, ui.Host.ViewZoom
	previousBudget := ui.Host.GlyphCacheBudgetBytes
	t.Cleanup(func() {
		ui.Host.WindowSize, ui.Host.WindowScale, ui.Host.ViewZoom = previousSize, previousScale, previousZoom
		ui.Host.GlyphCacheBudgetBytes = previousBudget
		ResetInputSession()
	})
	ui.Host.WindowSize, ui.Host.WindowScale, ui.Host.ViewZoom = Vec2{400, 200}, 2, 2
	ui.Host.Input.MousePoint = Vec2{60, 40}
	ui.Host.GlyphCacheBudgetBytes = 16 << 20
	scope := new(int)
	var out FrameOutputData
	var hovered bool
	for range 3 {
		out = RunFrameFn(func() {
			if ui.Host.WindowSize != (Vec2{200, 100}) || ui.Host.Input.MousePoint != (Vec2{30, 20}) || ui.Host.WindowScale != 4 {
				t.Fatal("layout/input scale mismatch")
			}
			ModAttrs(func(a *AttrSet) { a.Animations = 0 })
			ContainerWithKey(scope, Attrs(Viewport), func() {
				NextAccessName("zoom_box")
				Container(Attrs(FixSize(100, 50), Background(1, 0, 0, 1)), func() {
					AssignAccess()
					hovered = IsHovered()
					ui.Host.CaretPos, ui.Host.CompositionPos, ui.Host.CaretHeight = Vec2{10, 20}, Vec2{12, 24}, 16
					Label("Zoom", FontSize(16))
				})
			})
		})
		if ui.Host.WindowSize != (Vec2{400, 200}) || ui.Host.WindowScale != 2 || ui.Host.Input.MousePoint != (Vec2{60, 40}) {
			t.Fatal("backend geometry was not restored")
		}
	}
	if !hovered {
		t.Fatal("zoomed pointer missed the box")
	}
	if ui.Host.CaretPos != (Vec2{20, 40}) || ui.Host.CompositionPos != (Vec2{24, 48}) || ui.Host.CaretHeight != 32 {
		t.Fatal("IME anchors were not magnified")
	}
	node, ok := QueryContainer("zoom_box")
	if !ok || node.Bounds.Size != (Vec2{200, 100}) {
		t.Fatalf("accessibility bounds = %+v", node.Bounds)
	}
	foundBox, foundGlyph := false, false
	for _, s := range out.Surfaces {
		if s.Color1 == (Vec4{1, 0, 0, 1}) && s.Rect.Size == (Vec2{200, 100}) {
			foundBox = true
		}
		if s.GlyphRunCount > 0 {
			g := s.GlyphRunAt(0, out.GlyphRuns)
			if g.Rect.Size[1] != 32 {
				t.Fatalf("glyph size = %v, want 32", g.Rect.Size)
			}
			key, ok := GlyphKeyForRun(&g)
			if ok {
				if _, ok := GlyphBitmap(key); !ok {
					t.Fatal("zoomed glyph missing from cache")
				}
			}
			foundGlyph = true
		}
	}
	if !foundBox || !foundGlyph {
		t.Fatalf("missing zoomed paint: box=%v text=%v", foundBox, foundGlyph)
	}
}

func TestViewZoomSharedGlyphGeometryStaysImmutable(t *testing.T) {
	data := NewGlyphRunData([]GlyphRun{{Rect: Rect{Origin: Vec2{3, 4}, Size: Vec2{8, 16}}, GlyphOffset: Vec2{1, 2}}})
	s := Surface{Rect: Rect{Origin: Vec2{20, 30}}, GlyphData: data, GlyphDataScale: 2}
	got := s.GlyphRunAt(0, nil)
	if got.Rect != (Rect{Origin: Vec2{26, 38}, Size: Vec2{16, 32}}) || got.GlyphOffset != (Vec2{2, 4}) {
		t.Fatalf("scaled glyph = %+v", got)
	}
	if data.glyphs[0].Rect.Size[1] != 16 {
		t.Fatal("mutated shared geometry")
	}
}
