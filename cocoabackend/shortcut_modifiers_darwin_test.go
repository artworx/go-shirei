//go:build darwin && !ios

package cocoabackend

import (
	"testing"

	"go.hasen.dev/shirei"
)

func TestShortcutRetainsModifiersWhenReleasedBeforeFrame(t *testing.T) {
	for _, test := range []struct {
		name    string
		flags   uint
		mods    shirei.Modifiers
		key     uint16
		wantKey shirei.KeyCode
	}{
		{"Ctrl Tab", nsControl, shirei.ModCtrl, vkTab, shirei.KeyTab},
		{"Ctrl Shift Tab", nsControl | nsShift, shirei.ModCtrl | shirei.ModShift, vkTab, shirei.KeyTab},
		{"Cmd P", nsCommand, shirei.ModCmd, 0x23, shirei.KeyP},
		{"plain Tab", 0, shirei.ModNone, vkTab, shirei.KeyTab},
		{"Shift Tab", nsShift, shirei.ModShift, vkTab, shirei.KeyTab},
	} {
		t.Run(test.name, func(t *testing.T) {
			shirei.ResetInputSession()
			t.Cleanup(shirei.ResetInputSession)
			onModifiers(test.flags)
			onKeyDown(int(test.key), "")
			onKeyUp(int(test.key), "")
			onModifiers(0)
			var got shirei.KeyCombo
			runInputFrame(func() { got = shirei.ActiveCombo() })
			want := shirei.Combo(test.wantKey, test.mods)
			if got != want {
				t.Fatalf("shortcut after native key release = %#v, want %#v", got, want)
			}
			if shirei.GetInputState().Modifiers != shirei.ModNone {
				t.Fatal("frame did not restore the released modifier state")
			}
			runInputFrame(func() { got = shirei.ActiveCombo() })
			if got != (shirei.KeyCombo{}) {
				t.Fatalf("shortcut leaked into the next frame: %#v", got)
			}
		})
	}
}
