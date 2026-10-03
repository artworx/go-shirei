package shirei

import "testing"

func TestQuitRequestDelegatesWithoutHoldingRegistrationLock(t *testing.T) {
	SetQuitHandler(nil)
	t.Cleanup(func() { SetQuitHandler(nil) })
	if HandleQuitRequest() {
		t.Fatal("unregistered request intercepted")
	}
	calls := 0
	SetQuitHandler(func() { calls++; SetQuitHandler(nil) })
	if !HandleQuitRequest() || calls != 1 {
		t.Fatal("request not delegated")
	}
	if HandleQuitRequest() {
		t.Fatal("callback could not update handler")
	}
}
