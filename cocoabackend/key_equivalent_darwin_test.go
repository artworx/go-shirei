//go:build darwin && !ios

package cocoabackend

import (
	"runtime"
	"testing"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
	"go.hasen.dev/shirei"
)

func TestControlTabKeyEquivalentReachesFrame(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := ensureAppkit(); err != nil {
		t.Fatal(err)
	}
	libobjc, err := dlopen("/usr/lib/libobjc.A.dylib")
	if err != nil {
		t.Fatal(err)
	}
	// A fixed signature preserves Darwin ARM64 packing for the trailing BOOL
	// and unsigned short, which objc.Send's variadic convenience call cannot.
	var keyEvent func(objc.ID, objc.SEL, uint, nsPoint, uint, float64, int64, objc.ID, objc.ID, objc.ID, bool, uint16) objc.ID
	purego.RegisterLibFunc(&keyEvent, libobjc, "objc_msgSend")
	pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(sel("new"))
	defer pool.Send(sel("drain"))
	view := objc.ID(viewClass).Send(sel("alloc")).Send(sel("initWithFrame:"), nsMakeRect(0, 0, 100, 100))
	defer view.Send(sel("release"))
	for _, test := range []struct {
		name  string
		flags uint
		key   uint16
		want  shirei.KeyCombo
	}{
		{"Ctrl Tab", nsControl, vkTab, shirei.Combo(shirei.KeyTab, shirei.ModCtrl)},
		{"Ctrl Shift Tab", nsControl | nsShift, vkTab, shirei.Combo(shirei.KeyTab, shirei.ModCtrl|shirei.ModShift)},
		{"plain Tab", 0, vkTab, shirei.KeyCombo{}},
		{"Shift Tab", nsShift, vkTab, shirei.KeyCombo{}},
		{"Cmd Tab", nsCommand, vkTab, shirei.KeyCombo{}},
		{"Ctrl Alt Tab", nsControl | nsOption, vkTab, shirei.KeyCombo{}},
		{"Ctrl P", nsControl, 0x23, shirei.KeyCombo{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			shirei.ResetInputSession()
			t.Cleanup(shirei.ResetInputSession)
			gWantsFrame = false
			event := keyEvent(objc.ID(objc.GetClass("NSEvent")),
				sel("keyEventWithType:location:modifierFlags:timestamp:windowNumber:context:characters:charactersIgnoringModifiers:isARepeat:keyCode:"),
				uint(10), nsPoint{}, test.flags, float64(0), int64(0), objc.ID(0), nsString("\t"), nsString("\t"), false, test.key,
			)
			if got := objc.Send[uint](event, sel("modifierFlags")); got != test.flags {
				t.Fatalf("event flags = %#x, want %#x", got, test.flags)
			}
			if got := objc.Send[uint16](event, sel("keyCode")); got != test.key {
				t.Fatalf("event key = %#x, want %#x", got, test.key)
			}
			handled := objc.Send[bool](view, sel("performKeyEquivalent:"), event)
			wantHandled := test.want.Key != shirei.KeyCodeNone
			if handled != wantHandled {
				t.Fatalf("native key equivalent handled = %v, want %v", handled, wantHandled)
			}
			if gWantsFrame != wantHandled {
				t.Fatalf("frame requested = %v, want %v", gWantsFrame, wantHandled)
			}
			onModifiers(0)
			var got shirei.KeyCombo
			runInputFrame(func() { got = shirei.ActiveCombo() })
			if got != test.want {
				t.Fatalf("native key equivalent delivered %#v, want %#v", got, test.want)
			}
		})
	}
}
