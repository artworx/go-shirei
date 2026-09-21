package shirei

import "testing"

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
