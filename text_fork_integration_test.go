package shirei

import (
	"reflect"
	"strings"
	"testing"
)

func TestLargeIncrementalEditWithTabsAndRubyMatchesFullShape(t *testing.T) {
	style := requireTextShaping(t)
	WaitForSystemFontScan()
	for _, line := range []string{"iiii\tend\n", "prefix " + Ruby("base", "reading") + " tail\n", Ruby("iiii", "reading") + " tail\n"} {
		t.Run(line, func(t *testing.T) {
			source := strings.Repeat(line, 2200)
			ShapeTextMax(source, style, 900)
			edited := strings.Replace(source, "i", "W", 1)
			if edited == source {
				edited = strings.Replace(source, "base", "case", 1)
			}
			got := ShapeTextMax(edited, style, 900)
			runes := []rune(edited)
			segments := produceRubySegments(runes, style, nil)
			want := wrapUnwrapped(unwrappedShaped{Runes: runes, BaseDir: segments[0].Dir, Segments: segments}, style, 900)
			if !reflect.DeepEqual(got, want) {
				t.Fatal("incremental geometry differs from full shaping")
			}
		})
	}
}

func TestRubyRemappingDoesNotMutateCachedGlyphs(t *testing.T) {
	style := requireTextShaping(t)
	const text = "same"
	before := ShapeText(text, style)
	glyphs := append([]Glyph(nil), before.Lines[0].Segments[0].Glyphs...)
	for _, source := range []string{Ruby("base", "reading") + text, "prefix " + Ruby("base", text)} {
		ShapeText(source, style)
	}
	if !reflect.DeepEqual(glyphs, before.Lines[0].Segments[0].Glyphs) {
		t.Fatal("ruby remapping mutated shared glyphs")
	}
}

func TestRubyViewportIncludesReadingLineBox(t *testing.T) {
	style := requireTextShaping(t)
	shaped := ShapeText(Ruby("base", "reading"), style)
	line := &shaped.Lines[0]
	want := line.lineEm + 2*line.descenderPad
	if got := ShapedTextLayoutHeight(shaped, style); got != want {
		t.Fatalf("height = %g, want %g including ruby", got, want)
	}
}
