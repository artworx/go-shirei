//go:build darwin && !ios

package cocoabackend

import (
	"github.com/ebitengine/purego/objc"
	"go.hasen.dev/shirei"
	"testing"
)

func TestNativeQuitSelectorsKeepWindowOpenWhenHandled(t *testing.T) {
	calls := 0
	shirei.SetQuitHandler(func() { calls++ })
	t.Cleanup(func() { shirei.SetQuitHandler(nil) })
	if err := registerClasses(); err != nil {
		t.Fatal(err)
	}
	delegate := objc.ID(appDelegateClass).Send(sel("alloc")).Send(sel("init"))
	defer delegate.Send(sel("release"))
	if objc.Send[bool](delegate, sel("windowShouldClose:"), objc.ID(0)) {
		t.Fatal("window closed before save")
	}
	if objc.Send[uint](delegate, sel("applicationShouldTerminate:"), objc.ID(0)) != 0 {
		t.Fatal("application terminated before save")
	}
	if calls != 2 {
		t.Fatalf("requests=%d", calls)
	}
	shirei.SetQuitHandler(nil)
	if !objc.Send[bool](delegate, sel("windowShouldClose:"), objc.ID(0)) {
		t.Fatal("default close behavior changed")
	}
}
