package shirei

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLookupClosestFaceUsesAvailableWeightInFamily(t *testing.T) {
	const family = "Shirei Weight Fallback"
	registerTestFace(t, family, FontAspect{
		Weight:  WeightNormal,
		Style:   StyleNormal,
		Stretch: StretchNormal,
	})
	bold := registerTestFace(t, family, FontAspect{
		Weight:  WeightBold,
		Style:   StyleNormal,
		Stretch: StretchNormal,
	})

	got := LookupClosestFace(FaceLookupKey{family, FontAspect{
		Weight:  WeightSemibold,
		Style:   StyleNormal,
		Stretch: StretchNormal,
	}})
	if got != bold {
		t.Fatalf("semibold lookup = %d, want available bold face %d", got, bold)
	}
}

func registerTestFace(t *testing.T, family string, aspect FontAspect) FontId {
	t.Helper()
	faceRegistryMu.Lock()
	face := _nextFaceLocked()
	face.Family = family
	face.FaceLookupKey = FaceLookupKey{Family: family, Aspect: aspect}
	_mapFaceLocked(face.FaceLookupKey, face.FontId)
	faceRegistryMu.Unlock()
	t.Cleanup(func() {
		faceRegistryMu.Lock()
		delete(res.faceMap, FaceLookupKey{Family: strings.ToLower(family), Aspect: aspect})
		clear(res.closestFaceMap)
		faceRegistryMu.Unlock()
	})
	return face.FontId
}

func TestSystemFontParserPanicIsMemoized(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("ZitherIndia parser regression fixture is a macOS system font")
	}
	wantPath := filepath.Clean("/System/Library/Fonts/ZitherIndia.otf")
	var id FontId
	for _, face := range AllFontFaces() {
		if filepath.Clean(face.Filepath) == wantPath {
			id = face.FontId
			break
		}
	}
	if id == 0 {
		t.Skip("ZitherIndia system font is not installed")
	}

	_ = GetParsedFont(id)
	if GetFace(id).parseError == nil {
		t.Fatal("parser panic was not memoized; the font will be retried every frame")
	}
}
