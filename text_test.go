package shirei

import (
	"strings"
	"testing"
)

func requireTextShaping(t *testing.T) TextStyleAttrs {
	t.Helper()
	InitFontSubsystem()
	attrs := DefaultTextStyle()
	probe := ShapeText("alpha", attrs)
	if len(probe.Lines) != 1 || len(probe.Lines[0].Segments) == 0 {
		t.Skip("no usable system fonts for text shaping")
	}
	return attrs
}

func TestShapeTextNewlineHasNoAdvance(t *testing.T) {
	attrs := requireTextShaping(t)
	bWidth := ShapeText("b", attrs).Lines[0].Width

	shaped := ShapeText("a\n\nb", attrs)
	if len(shaped.Lines) != 3 {
		t.Fatalf("line count = %d, want 3", len(shaped.Lines))
	}
	if shaped.Lines[1].Width != 0 {
		t.Errorf("empty hard line width = %.2f, want 0", shaped.Lines[1].Width)
	}
	if shaped.Lines[1].Height <= 0 {
		t.Errorf("empty hard line height = %.2f, want positive", shaped.Lines[1].Height)
	}
	if shaped.Lines[1].Height != shaped.Lines[0].Height {
		t.Errorf("empty hard line height = %.2f, filled line height = %.2f", shaped.Lines[1].Height, shaped.Lines[0].Height)
	}
	if shaped.Lines[2].Width != bWidth {
		t.Errorf("line after hard break width = %.2f, want %.2f", shaped.Lines[2].Width, bWidth)
	}

	trailing := ShapeText("ab\n", attrs)
	if len(trailing.Lines) != 2 {
		t.Fatalf("trailing newline line count = %d, want 2", len(trailing.Lines))
	}
	if trailing.Lines[1].Width != 0 {
		t.Errorf("trailing phantom line width = %.2f, want 0", trailing.Lines[1].Width)
	}
	if trailing.Lines[1].Height <= 0 {
		t.Errorf("trailing phantom line height = %.2f, want positive", trailing.Lines[1].Height)
	}
	if trailing.Lines[1].Height != trailing.Lines[0].Height {
		t.Errorf("trailing phantom line height = %.2f, filled line height = %.2f", trailing.Lines[1].Height, trailing.Lines[0].Height)
	}
}

func TestShapeCacheStampsAreGeometry(t *testing.T) {
	attrs := requireTextShaping(t)
	shaped := ShapeText("Hello", attrs)
	line := shaped.Lines[0]
	if len(line.stamps) == 0 || len(line.runs) == 0 {
		t.Fatal("expected stamps and runs at shape time")
	}
	n := 0
	for _, seg := range line.Segments {
		n += len(seg.Glyphs)
	}
	if len(line.stamps) != n || len(line.runs) != n {
		t.Fatalf("stamps=%d runs=%d glyphs=%d", len(line.stamps), len(line.runs), n)
	}
	for i, r := range line.runs {
		// Cached runs must stay colorless: the shape-cache key excludes
		// render-tier color, so a baked color would leak across entries.
		if r.Color != (Vec4{}) {
			t.Fatalf("run[%d] color = %v, want zero", i, r.Color)
		}
		if r.FontId == 0 || r.GlyphId == 0 {
			t.Fatalf("run[%d] missing font/glyph", i)
		}
	}
	if line.lineEm != attrs.FontSize {
		t.Fatalf("lineEm=%v want %v", line.lineEm, attrs.FontSize)
	}
}

func TestShapeCacheStampsDoNotLeakColor(t *testing.T) {
	requireTextShaping(t)
	red := Vec4{0, 80, 50, 1}
	blue := Vec4{210, 80, 50, 1}
	ui.Host.WindowSize = Vec2{240, 100}
	ui.Host.WindowScale = 1
	sid := softScope("shape_stamp_color")
	var out FrameOutputData
	for range 2 {
		out = RunFrameFn(func() {
			ModAttrs(func(a *AttrSet) { a.Animations = 0 })
			ContainerWithKey(sid, AttrSet{}, func() {
				Label("Hello", TextColor(red[0], red[1], red[2], red[3]))
				Label("Hello", TextColor(blue[0], blue[1], blue[2], blue[3]))
			})
		})
	}
	var sawRed, sawBlue bool
	for _, s := range out.Surfaces {
		for i := 0; i < int(s.GlyphRunCount); i++ {
			g := s.GlyphRunAt(i, out.GlyphRuns)
			if g.Color == red {
				sawRed = true
			}
			if g.Color == blue {
				sawBlue = true
			}
		}
	}
	if !sawRed || !sawBlue {
		t.Fatalf("glyph colors missing or leaked: red=%v blue=%v runs=%d", sawRed, sawBlue, len(out.GlyphRuns))
	}
}

func TestInternFamilyList(t *testing.T) {
	var a, b TextStyleAttrs
	a.SetFontFamilies("Menlo", "Monaco")
	b.SetFontFamilies("menlo", "MONACO")
	if a.fontFamilies == nil || a.fontFamilies != b.fontFamilies {
		t.Fatal("equal names (any case) intern to one list")
	}
	if a.fontFamilies.id == 0 {
		t.Fatal("non-empty list must not be id 0")
	}

	var c TextStyleAttrs
	c.fontFamilies = emptyFamilyList
	Fonts("Menlo", "Monaco")(&c)
	if c.fontFamilies != a.fontFamilies {
		t.Fatal("Fonts intern matches SetFontFamilies")
	}

	var d TextStyleAttrs
	d.SetFontFamilies("Noto Sans")
	Fonts("Menlo")(&d)
	var e TextStyleAttrs
	e.SetFontFamilies("Menlo", "Noto Sans")
	if d.fontFamilies != e.fontFamilies {
		t.Fatal("Fonts prepend intern")
	}

	var empty TextStyleAttrs
	empty.SetFontFamilies()
	if empty.fontFamilies != emptyFamilyList {
		t.Fatal("empty list is the singleton")
	}
}

// Shape-cache keys hash interned family-list ids. Two SetFontFamilies
// clones (including different case) must hit; a different name must miss
// even if both LookupFace answers are 0.
func TestShapeCacheKeyHashesFamilyNames(t *testing.T) {
	attrs := requireTextShaping(t)
	const text = "shape-cache-family-names"

	a := attrs
	a.SetFontFamilies("Menlo", "Monaco")
	b := attrs
	b.SetFontFamilies("menlo", "MONACO")

	ShapeStats.Calls = 0
	ShapeStats.Hits = 0
	_ = ShapeText(text, a)
	_ = ShapeText(text, b)
	if ShapeStats.Calls != 2 || ShapeStats.Hits != 1 {
		t.Fatalf("interned family lists: calls=%d hits=%d (want 2/1)", ShapeStats.Calls, ShapeStats.Hits)
	}

	c := attrs
	c.SetFontFamilies("DefinitelyNotARegisteredFamilyXXXX")
	_ = ShapeText(text, c)
	if ShapeStats.Hits != 1 {
		t.Fatalf("different family names must miss, hits=%d", ShapeStats.Hits)
	}
}

func TestShapeCacheMissesOnEpochBump(t *testing.T) {
	attrs := requireTextShaping(t)
	const text = "shape-cache-epoch-bump"

	ShapeStats.Calls = 0
	ShapeStats.Hits = 0
	_ = ShapeText(text, attrs)
	_ = ShapeText(text, attrs)
	if ShapeStats.Hits != 1 {
		t.Fatalf("want hit before bump, calls=%d hits=%d", ShapeStats.Calls, ShapeStats.Hits)
	}

	faceRegistryMu.Lock()
	res.fontLookupEpoch++
	faceRegistryMu.Unlock()
	t.Cleanup(func() {
		faceRegistryMu.Lock()
		res.fontLookupEpoch--
		faceRegistryMu.Unlock()
	})

	_ = ShapeText(text, attrs)
	if ShapeStats.Hits != 1 {
		t.Fatalf("epoch bump must miss, calls=%d hits=%d", ShapeStats.Calls, ShapeStats.Hits)
	}
	_ = ShapeText(text, attrs)
	if ShapeStats.Hits != 2 {
		t.Fatalf("want settle after wipe, calls=%d hits=%d", ShapeStats.Calls, ShapeStats.Hits)
	}
}

func TestShapeCacheSharesUnwrappedAcrossWidths(t *testing.T) {
	attrs := requireTextShaping(t)
	ui.Host.WindowScale = 1
	const text = "shape-cache-unwrapped-width-reuse ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	ShapeStats.Calls = 0
	ShapeStats.Hits = 0
	ShapeStats.ShapeHits = 0
	a := ShapeTextMax(text, attrs, 40)
	b := ShapeTextMax(text, attrs, 4000)
	if ShapeStats.Calls != 2 {
		t.Fatalf("calls=%d", ShapeStats.Calls)
	}
	if ShapeStats.Hits != 0 {
		t.Fatalf("different wrap widths must miss the wrap cache, hits=%d", ShapeStats.Hits)
	}
	if ShapeStats.ShapeHits != 1 {
		t.Fatalf("second width should reuse unwrapped shape, ShapeHits=%d", ShapeStats.ShapeHits)
	}
	if len(a.Lines) <= len(b.Lines) {
		t.Fatalf("narrow wrap %d lines, wide wrap %d; want narrow > wide", len(a.Lines), len(b.Lines))
	}

	// 40.2 and 40.4 round to the same device px at scale 1.
	const qtext = "quantize-wrap-width-probe-XXXX"
	ShapeStats.Calls = 0
	ShapeStats.Hits = 0
	ShapeStats.ShapeHits = 0
	_ = ShapeTextMax(qtext, attrs, 40.2)
	_ = ShapeTextMax(qtext, attrs, 40.4)
	if ShapeStats.Hits != 1 || ShapeStats.ShapeHits != 1 {
		t.Fatalf("quantized widths: hits=%d shapeHits=%d (want 1/1)", ShapeStats.Hits, ShapeStats.ShapeHits)
	}
}

func TestShapeSegmentReusesCachedResult(t *testing.T) {
	attrs := requireTextShaping(t)
	probe := ShapeText("segment-cache-probe", attrs)
	if len(probe.Lines) == 0 || len(probe.Lines[0].Segments) == 0 {
		t.Fatal("probe produced no shaped segment")
	}
	props := probe.Lines[0].Segments[0].GlyphSegmentProps
	runes := []rune("segment-cache-value")

	SegmentShapeStats.Calls = 0
	SegmentShapeStats.Hits = 0
	first := shapeSegment(props, runes, 0, len(runes))
	second := shapeSegment(props, runes, 0, len(runes))

	if SegmentShapeStats.Calls != 2 || SegmentShapeStats.Hits != 1 {
		t.Fatalf("segment cache calls=%d hits=%d, want calls=2 hits=1", SegmentShapeStats.Calls, SegmentShapeStats.Hits)
	}
	if first.Width != second.Width || len(first.Glyphs) != len(second.Glyphs) {
		t.Fatalf("cached segment geometry changed: first=%+v second=%+v", first, second)
	}
}

func TestLargeShapeCacheIsBoundedAndLRU(t *testing.T) {
	attrs := requireTextShaping(t)
	a := strings.Repeat("a", 16*1024+1)
	b := strings.Repeat("b", 16*1024+1)
	c := strings.Repeat("c", 16*1024+1)

	ShapeStats.Calls = 0
	ShapeStats.Hits = 0
	_ = ShapeText(a, attrs)
	_ = ShapeText(b, attrs)
	_ = ShapeText(a, attrs) // refresh a; b is now least recently used
	_ = ShapeText(c, attrs) // evicts b
	_ = ShapeText(b, attrs) // must miss

	if ShapeStats.Calls != 5 || ShapeStats.Hits != 1 {
		t.Fatalf("large shape cache calls=%d hits=%d, want calls=5 hits=1", ShapeStats.Calls, ShapeStats.Hits)
	}
}

func TestLargeSameWidthEditReshapesOneSegment(t *testing.T) {
	attrs := requireTextShaping(t)
	text := strings.Repeat("aaaaaaaaaaaaaaaa\n", 1100)
	initial := ShapeText(text, attrs)
	if len(initial.Runes) <= 16*1024 {
		t.Fatal("fixture must use the large-document path")
	}

	edited := []byte(text)
	edited[5] = 'b'
	SegmentShapeStats.Calls = 0
	SegmentShapeStats.Hits = 0
	shaped := ShapeText(string(edited), attrs)

	if SegmentShapeStats.Calls != 1 {
		t.Fatalf("edited document shaped %d segments, want exactly 1", SegmentShapeStats.Calls)
	}
	if shaped.Runes[5] != 'b' {
		t.Fatalf("edited rune = %q, want b", shaped.Runes[5])
	}
	if len(shaped.Lines) != len(initial.Lines) {
		t.Fatalf("line count changed from %d to %d", len(initial.Lines), len(shaped.Lines))
	}
}

func TestBidiLineDirectionsReuseCachedSlice(t *testing.T) {
	line := "abc אבג"
	first := bidiLineDirections(line)
	second := bidiLineDirections(line)
	if len(first) == 0 || len(second) == 0 {
		t.Fatal("bidi line produced no directions")
	}
	if &first[0] != &second[0] {
		t.Fatal("repeated bidi line did not reuse the cached direction slice")
	}
}

func TestCalculateShapedTextViewportWindow(t *testing.T) {
	style := DefaultTextStyle()
	style.FontSize = 20
	shaped := ShapedText{Lines: make([]ShapedTextLine, 100)}
	for index := range shaped.Lines {
		shaped.Lines[index] = ShapedTextLine{
			Height:   20,
			Segments: []GlyphsSegment{{GlyphSegmentProps: GlyphSegmentProps{size: 20}}},
		}
	}

	window := calculateShapedTextViewportWindow(shaped, style, 800, 200)
	if window.start != 34 || window.end != 56 {
		t.Fatalf("visible line range = [%d,%d), want [34,56)", window.start, window.end)
	}
	if window.topSpacer != 680 || window.bottomSpacer != 880 {
		t.Fatalf("spacers = top %.1f bottom %.1f, want 680 and 880", window.topSpacer, window.bottomSpacer)
	}
	visibleHeight := float32(window.end-window.start) * 20
	if window.topSpacer+visibleHeight+window.bottomSpacer != 2000 {
		t.Fatalf("virtualized height = %.1f, want full height 2000", window.topSpacer+visibleHeight+window.bottomSpacer)
	}
}
