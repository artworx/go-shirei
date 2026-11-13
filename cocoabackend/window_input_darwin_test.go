//go:build darwin && !ios

package cocoabackend

import (
	"github.com/ebitengine/purego/objc"
	"go.hasen.dev/shirei"
	"testing"
)

func TestPendingInputIsOwnedByItsWindow(t *testing.T) {
	previous := currentWindow
	t.Cleanup(func() { currentWindow = previous })
	first := newWindowState(shirei.NewUI())
	second := newWindowState(shirei.NewUI())
	enter := enterWindow(first)
	onModifiers(nsControl)
	onKeyDown(int(vkTab), "")
	onModifiers(0)
	onMouse(10, 10, mouseDown, 0)
	onMouse(10, 10, mouseUp, 0)
	enter()
	enter = enterWindow(second)
	runInputFrameWith(func() {
		if shirei.ActiveCombo().Key != shirei.KeyCodeNone || shirei.GetFrameInput().Mouse != 0 {
			t.Error("input leaked into another window")
		}
	}, shirei.RunFrameFnLocked)
	enter()
	enter = enterWindow(first)
	runInputFrameWith(func() {
		if shirei.ActiveCombo() != shirei.Combo(shirei.KeyTab, shirei.ModCtrl) || shirei.GetFrameInput().Mouse != shirei.MouseClick {
			t.Error("original window lost pending input")
		}
	}, shirei.RunFrameFnLocked)
	runInputFrameWith(func() {
		if shirei.GetFrameInput().Mouse != shirei.MouseRelease {
			t.Error("original window lost release")
		}
	}, shirei.RunFrameFnLocked)
	enter()
}

func TestPopupCloseDoesNotRequestApplicationQuit(t *testing.T) {
	calls := 0
	shirei.SetQuitHandler(func() { calls++ })
	t.Cleanup(func() { shirei.SetQuitHandler(nil) })
	const window objc.ID = 0x1235
	windowsByObject[window] = newWindowState(shirei.NewUI())
	t.Cleanup(func() { delete(windowsByObject, window) })
	if appWindowShouldClose(0, 0, window) || calls != 0 {
		t.Fatal("popup close requested application quit")
	}
}
