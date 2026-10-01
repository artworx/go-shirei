//go:build darwin && !ios

package cocoabackend

import (
	"testing"

	"go.hasen.dev/shirei"
)

func resetMouseEdgeTestState(t *testing.T) {
	t.Helper()
	shirei.ResetInputSession()
	pendingMouseEdges = pendingMouseEdges[:0]
	t.Cleanup(func() {
		shirei.ResetInputSession()
		pendingMouseEdges = pendingMouseEdges[:0]
	})
}

func TestQuickMouseClickPreservesDownAndUpAcrossFrames(t *testing.T) {
	for _, test := range []struct {
		name   string
		button int
		want   shirei.MouseButton
	}{
		{name: "primary", button: 0, want: shirei.MousePrimary},
		{name: "secondary", button: 1, want: shirei.MouseSecondary},
	} {
		t.Run(test.name, func(t *testing.T) {
			resetMouseEdgeTestState(t)

			onMouse(24, 36, mouseDown, test.button)
			onMouse(24, 36, mouseUp, test.button)

			var firstAction shirei.MouseAction
			var firstButton shirei.MouseButton
			runInputFrame(func() {
				firstAction = shirei.GetFrameInput().Mouse
				firstButton = shirei.GetInputState().MouseButton
			})
			if firstAction != shirei.MouseClick || firstButton != test.want {
				t.Fatalf("first frame = (%v, %v), want (%v, %v)",
					firstAction, firstButton, shirei.MouseClick, test.want)
			}

			var secondAction shirei.MouseAction
			var secondButton shirei.MouseButton
			runInputFrame(func() {
				secondAction = shirei.GetFrameInput().Mouse
				secondButton = shirei.GetInputState().MouseButton
			})
			if secondAction != shirei.MouseRelease || secondButton != test.want {
				t.Fatalf("second frame = (%v, %v), want (%v, %v)",
					secondAction, secondButton, shirei.MouseRelease, test.want)
			}
		})
	}
}

func TestQuickMouseClickCompletesPressActionOnce(t *testing.T) {
	resetMouseEdgeTestState(t)
	shirei.GetHost().WindowSize = shirei.Vec2{100, 100}
	scope := new(int)
	presses := 0
	view := func() {
		shirei.ContainerWithKey(scope, shirei.Attrs(shirei.FixSize(100, 100)), func() {
			if shirei.PressAction() {
				presses++
			}
		})
	}

	onMouse(24, 36, mouseMove, 0)
	runInputFrame(view)
	onMouse(24, 36, mouseDown, 0)
	onMouse(24, 36, mouseUp, 0)

	runInputFrame(view)
	if presses != 0 {
		t.Fatalf("press completed on mouse-down frame: %d", presses)
	}
	runInputFrame(view)
	if presses != 1 {
		t.Fatalf("presses after mouse-up frame = %d, want 1", presses)
	}
	runInputFrame(view)
	if presses != 1 {
		t.Fatalf("neutral follow-up repeated press: %d", presses)
	}
}

func TestMouseEdgesAreDeliveredOnceInArrivalOrder(t *testing.T) {
	resetMouseEdgeTestState(t)
	edges := []struct {
		action int
		button int
		want   mouseEdge
	}{
		{mouseDown, 0, mouseEdge{action: shirei.MouseClick, button: shirei.MousePrimary}},
		{mouseUp, 0, mouseEdge{action: shirei.MouseRelease, button: shirei.MousePrimary}},
		{mouseDown, 1, mouseEdge{action: shirei.MouseClick, button: shirei.MouseSecondary}},
		{mouseUp, 1, mouseEdge{action: shirei.MouseRelease, button: shirei.MouseSecondary}},
	}
	for _, edge := range edges {
		onMouse(24, 36, edge.action, edge.button)
	}

	for index, edge := range edges {
		var got mouseEdge
		runInputFrame(func() {
			got = mouseEdge{
				action: shirei.GetFrameInput().Mouse,
				button: shirei.GetInputState().MouseButton,
			}
		})
		if got != edge.want {
			t.Fatalf("frame %d = %#v, want %#v", index, got, edge.want)
		}
	}
	var trailing shirei.MouseAction
	runInputFrame(func() { trailing = shirei.GetFrameInput().Mouse })
	if trailing != 0 || len(pendingMouseEdges) != 0 {
		t.Fatalf("queue did not drain: trailing action=%v, pending=%d", trailing, len(pendingMouseEdges))
	}
}

func TestSingleMouseEdgeIsNotRepeated(t *testing.T) {
	for _, test := range []struct {
		name   string
		action int
		want   shirei.MouseAction
	}{
		{name: "down", action: mouseDown, want: shirei.MouseClick},
		{name: "up", action: mouseUp, want: shirei.MouseRelease},
	} {
		t.Run(test.name, func(t *testing.T) {
			resetMouseEdgeTestState(t)
			onMouse(24, 36, test.action, 0)
			var first shirei.MouseAction
			runInputFrame(func() { first = shirei.GetFrameInput().Mouse })
			var second shirei.MouseAction
			runInputFrame(func() { second = shirei.GetFrameInput().Mouse })
			if first != test.want || second != 0 {
				t.Fatalf("actions = (%v, %v), want (%v, 0)", first, second, test.want)
			}
		})
	}
}
