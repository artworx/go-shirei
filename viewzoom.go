package shirei

import "math"

// The scale captured at frame entry also applies when an app changes ViewZoom
// during a build. That change takes effect on its requested next frame.
func viewZoom() float32 {
	z := ui.Host.ViewZoom
	if z <= 0 || math.IsNaN(float64(z)) || math.IsInf(float64(z), 0) {
		return 1
	}
	return z
}

func beginViewZoom() func() {
	z := viewZoom()
	if z == 1 {
		return func() {}
	}
	size, scale := ui.Host.WindowSize, ui.Host.WindowScale
	mouse, touches := ui.Host.Input.MousePoint, ui.Host.Input.Touches
	ui.Host.WindowSize = Vec2Mul(size, 1/z)
	ui.Host.WindowScale = scale * z
	ui.Host.Input.MousePoint = Vec2Mul(mouse, 1/z)
	ui.Host.FrameInput.Motion = Vec2Mul(ui.Host.FrameInput.Motion, 1/z)
	ui.Host.FrameInput.Scroll = Vec2Mul(ui.Host.FrameInput.Scroll, 1/z)
	for i := range ui.Host.Input.Touches {
		ui.Host.Input.Touches[i].Pos = Vec2Mul(touches[i].Pos, 1/z)
	}
	restored := false
	return func() {
		if restored {
			return
		}
		restored = true
		ui.Host.WindowSize, ui.Host.WindowScale = size, scale
		ui.Host.Input.MousePoint = Vec2Mul(ui.Host.Input.MousePoint, z)
		for i := range ui.Host.Input.Touches {
			ui.Host.Input.Touches[i].Pos = Vec2Mul(ui.Host.Input.Touches[i].Pos, z)
		}
		// Paint is published at the scale captured before the app's build.
		scaleFrameArtifacts(z)
	}
}

func scaleZoomRect(r Rect, z float32) Rect {
	return Rect{Origin: Vec2Mul(r.Origin, z), Size: Vec2Mul(r.Size, z)}
}

func surfaceGlyphScale(s *Surface) float32 {
	if s.GlyphDataScale == 0 {
		return 1
	}
	return s.GlyphDataScale
}

func scaleFrameArtifacts(z float32) {
	if z == 1 {
		return
	}
	for i := range ui.surfaces {
		s := &ui.surfaces[i]
		s.Rect = scaleZoomRect(s.Rect, z)
		for j := range s.Corners {
			s.Corners[j] *= z
		}
		s.Stroke *= z
		s.GlyphOffset = Vec2Mul(s.GlyphOffset, z)
		if s.GlyphData != nil {
			s.GlyphDataScale = surfaceGlyphScale(s) * z
		}
	}
	for i := range ui.glyphRuns {
		g := &ui.glyphRuns[i]
		g.Rect = scaleZoomRect(g.Rect, z)
		g.GlyphOffset = Vec2Mul(g.GlyphOffset, z)
	}
	for i := range ui.access {
		a := &ui.access[i]
		a.Bounds = scaleZoomRect(a.Bounds, z)
		a.Rect = scaleZoomRect(a.Rect, z)
		a.Scroll = Vec2Mul(a.Scroll, z)
		a.ScrollMax = Vec2Mul(a.ScrollMax, z)
	}
	ui.Host.CaretPos = Vec2Mul(ui.Host.CaretPos, z)
	ui.Host.CaretHeight *= z
	ui.Host.CompositionPos = Vec2Mul(ui.Host.CompositionPos, z)
}
