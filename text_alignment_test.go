package shirei

import (
	"math"
	"testing"
)

func TestShapedTextLayoutAlignedPositionsTextWithinAvailableWidth(t *testing.T) {
	InitFontSubsystem()
	style := DefaultTextStyle()
	style.FontSize = 16
	shaped := ShapeTextMax("aligned", style, 180)

	origin := func(alignment Alignment) float32 {
		ResetInputSession()
		GetHost().WindowSize = Vec2{240, 80}
		var output FrameOutputData
		for range 3 {
			output = RunFrameFn(func() {
				Container(Attrs(NoAnimate, FixSize(200, 60)), func() {
					ShapedTextLayoutAligned(shaped, style, alignment, 0, 0)
				})
			})
		}
		for _, surface := range output.Surfaces {
			if surface.GlyphData != nil && surface.GlyphRunCount > 0 {
				return surface.Rect.Origin[0]
			}
		}
		t.Fatalf("alignment %v did not paint text", alignment)
		return 0
	}

	start := origin(AlignStart)
	middle := origin(AlignMiddle)
	end := origin(AlignEnd)
	if !(start < middle && middle < end) {
		t.Fatalf("text origins start/middle/end = %.1f/%.1f/%.1f, want increasing positions", start, middle, end)
	}
}

// Selection/decorations must follow the aligned glyph origin, including the
// v0.8 literal-color path and a fully transparent per-field selection.
func TestShapedTextStyledAlignmentKeepsSelectionOnGlyphs(t *testing.T) {
	style := requireTextShaping(t)
	style.FontSize = 16
	defer ResetInputSession()
	for _, text := range []string{"aligned", "aligned\nsecond line"} {
		shaped := ShapeTextMax(text, style, 180)
		for _, alignment := range []Alignment{AlignStart, AlignMiddle, AlignEnd} {
			for _, selection := range []Vec4{{120, 65, 45, .7}, {}} {
				ResetInputSession()
				GetHost().WindowSize = Vec2{240, 120}
				var output FrameOutputData
				for range 3 {
					output = RunFrameFn(func() {
						Container(Attrs(NoAnimate, FixSize(200, 100)), func() {
							ShapedTextLayoutStyledAligned(shaped, style, alignment, 0, len(shaped.Runes), selection)
						})
					})
				}
				var glyphs, highlights []Surface
				for _, surface := range output.Surfaces {
					if surface.GlyphData != nil && surface.GlyphRunCount > 0 {
						glyphs = append(glyphs, surface)
					}
					if surface.GlyphData == nil && surface.Color1[3] > 0 && surface.Stroke == 0 {
						highlights = append(highlights, surface)
					}
				}
				if len(glyphs) != len(shaped.Lines) {
					t.Fatalf("text %q: painted %d glyph lines, want %d", text, len(glyphs), len(shaped.Lines))
				}
				if selection == (Vec4{}) {
					if len(highlights) != 0 {
						t.Fatal("transparent selection fell back to package color")
					}
					continue
				}
				if len(highlights) != len(glyphs) {
					t.Fatalf("text %q: highlights=%d glyph lines=%d", text, len(highlights), len(glyphs))
				}
				for i, highlight := range highlights {
					if highlight.Color1 != selection {
						t.Fatalf("selection color=%v, want %v", highlight.Color1, selection)
					}
					if math.Abs(float64(highlight.Rect.Origin[0]-glyphs[i].Rect.Origin[0])) > .01 {
						t.Fatalf("text %q alignment %v: highlight x=%v glyph x=%v", text, alignment, highlight.Rect.Origin[0], glyphs[i].Rect.Origin[0])
					}
				}
			}
		}
	}
}
