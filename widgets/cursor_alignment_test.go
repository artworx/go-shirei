package widgets

import (
	"testing"

	. "go.hasen.dev/shirei"
)

func TestComputeCursorIndexAlignedUsesVisualLineOffset(t *testing.T) {
	InitFontSubsystem()
	style := DefaultTextStyle()
	style.FontSize = 16
	shaped := ShapeText("aligned", style)
	lineWidth := shaped.Lines[0].Width
	rect := Rect{Size: Vec2{200, 40}}
	visualStart := rect.Size[0] - lineWidth

	start := ComputeCursorIndexAligned(rect, Vec2{visualStart + 1, 8}, Vec2{}, shaped, AlignEnd)
	end := ComputeCursorIndexAligned(rect, Vec2{rect.Size[0] - 1, 8}, Vec2{}, shaped, AlignEnd)
	if start != 0 {
		t.Fatalf("right-aligned visual start maps to %d, want 0", start)
	}
	if end != len(shaped.Runes) {
		t.Fatalf("right-aligned visual end maps to %d, want %d", end, len(shaped.Runes))
	}
}
