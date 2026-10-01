package shirei

import (
	"fmt"
	"testing"
)

func TestTabShortcutDoesNotCycleControlFocus(t *testing.T) {
	for _, mods := range []Modifiers{ModCtrl, ModCtrl | ModShift, ModCmd, ModAlt, ModSuper} {
		t.Run(fmt.Sprint(mods), func(t *testing.T) {
			ResetInputSession()
			t.Cleanup(ResetInputSession)
			var first ContainerId
			view := func() {
				first = Container(Attrs(Focusable, FixSize(10, 10)), func() {})
				Container(Attrs(Focusable, FixSize(10, 10)), func() {})
			}
			focusTestFrame(view)
			focusTestTab(view, false)
			focusTestFrame(view)
			if !IdHasFocus(first) {
				t.Fatal("initial Tab did not focus the first control")
			}
			ui.Host.Input.Modifiers = mods
			ui.Host.FrameInput.Key = KeyTab
			focusTestFrame(func() {
				if got := ActiveCombo(); got != Combo(KeyTab, mods) {
					t.Fatalf("app received %v, want %v", got, Combo(KeyTab, mods))
				}
				// The application consumes the shortcut during its frame builder.
				ui.Host.FrameInput.Key = KeyCodeNone
				view()
			})
			ui.Host.Input.Modifiers = ModNone
			focusTestFrame(view)
			if !IdHasFocus(first) {
				t.Fatal("Tab shortcut advanced control focus before the app could consume it")
			}
		})
	}
}
