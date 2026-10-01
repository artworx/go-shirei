package shirei

import "testing"

func TestFocusRevealPreservesVisibleOversizedContent(t *testing.T) {
	ResetInputSession()
	t.Cleanup(ResetInputSession)
	ui.Host.WindowSize = Vec2{400, 300}
	var pane, content ContainerId
	view := func() {
		pane = Container(Attrs(Viewport, FixSize(400, 300), Clip), func() {
			ScrollOnInput()
			content = Container(Attrs(Focusable, FixSize(400, 1200)), func() { FocusOnClick() })
		})
	}
	ui.Host.Input.MousePoint = Vec2{100, 100}
	for range 3 {
		RunFrameFn(view)
	}
	ui.Host.FrameInput.Scroll = Vec2{0, 400}
	RunFrameFn(view)
	before := GetScrollOffsetOf(pane)
	if before[1] == 0 {
		t.Fatal("pane did not scroll")
	}
	ui.Host.FrameInput.Scroll = Vec2{}
	ui.Host.FrameInput.Mouse = MouseClick
	RunFrameFn(view)
	if !IdHasFocus(content) {
		t.Fatal("click did not focus the content")
	}
	if after := GetScrollOffsetOf(pane); after != before {
		t.Fatalf("focus on visible oversized content moved scroll from %v to %v", before, after)
	}
}

func TestRevealDelta(t *testing.T) {
	for _, tt := range []struct {
		name           string
		from, to, want float32
	}{
		{"visible", 120, 180, 0},
		{"small above", 80, 140, -20},
		{"small below", 260, 320, 20},
		{"oversized straddling", -500, 700, 0},
		{"oversized partly above", -500, 200, 0},
		{"oversized partly below", 200, 700, 0},
		{"oversized entirely above", -500, 100, -600},
		{"oversized entirely below", 300, 900, 200},
		{"viewport sized visible", 100, 300, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := revealDelta(tt.from, tt.to, 100, 300); got != tt.want {
				t.Fatalf("revealDelta(%v, %v, 100, 300) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestRunFrameTabWithNoFocusableControls(t *testing.T) {
	tests := []struct {
		name      string
		modifiers Modifiers
	}{
		{name: "Tab"},
		{name: "Shift+Tab", modifiers: ModShift},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetInputSession()
			RunFrameFn(func() {}) // establish an empty previous frame

			ui.Host.Input.Modifiers = tt.modifiers
			ui.Host.FrameInput.Key = KeyTab
			RunFrameFn(func() {})

			if ui.nextFocused != nil {
				t.Fatal("Tab with no focusable controls scheduled focus")
			}
		})
	}
}

func focusTestFrame(view FrameFn) {
	ui.Host.WindowSize = Vec2{400, 300}
	RunFrameFn(view)
}

func focusTestTab(view FrameFn, shift bool) {
	if shift {
		ui.Host.Input.Modifiers = ModShift
	}
	ui.Host.FrameInput.Key = KeyTab
	focusTestFrame(view)
	ui.Host.Input.Modifiers = 0
}

func TestTabCyclesFocusableInSourceOrder(t *testing.T) {
	ResetInputSession()
	var a, b, c ContainerId
	view := func() {
		a = Container(Attrs(Focusable, FixSize(10, 10)), func() {})
		b = Container(Attrs(Focusable, FixSize(10, 10)), func() {})
		c = Container(Attrs(Focusable, FixSize(10, 10)), func() {})
	}
	focusTestFrame(view)

	focusTestTab(view, false)
	focusTestFrame(view)
	if !IdHasFocus(a) {
		t.Fatal("Tab from nothing should land on the first focusable")
	}

	focusTestTab(view, false)
	focusTestFrame(view)
	if !IdHasFocus(b) {
		t.Fatal("Tab should move to the second focusable")
	}

	focusTestTab(view, false)
	focusTestFrame(view)
	if !IdHasFocus(c) {
		t.Fatal("Tab should move to the third focusable")
	}

	focusTestTab(view, false)
	focusTestFrame(view)
	if !IdHasFocus(a) {
		t.Fatal("Tab should wrap to the first focusable")
	}

	focusTestTab(view, true)
	focusTestFrame(view)
	if !IdHasFocus(c) {
		t.Fatal("Shift+Tab should wrap to the last focusable")
	}
}

func TestTabOrderIgnoresZ(t *testing.T) {
	ResetInputSession()
	var first, second ContainerId
	view := func() {
		first = Container(Attrs(Focusable, InFront, FixSize(10, 10)), func() {})
		second = Container(Attrs(Focusable, FixSize(10, 10)), func() {})
	}
	focusTestFrame(view)
	focusTestTab(view, false)
	focusTestFrame(view)
	if !IdHasFocus(first) {
		t.Fatal("Tab order follows source order, not InFront paint order")
	}
	focusTestTab(view, false)
	focusTestFrame(view)
	if !IdHasFocus(second) {
		t.Fatal("second source-order sibling should be next, even if first paints in front")
	}
}

func TestModalTrapExcludesBackground(t *testing.T) {
	ResetInputSession()
	var bg, inner ContainerId
	open := true
	view := func() {
		bg = Container(Attrs(Focusable, FixSize(20, 20)), func() {})
		if open {
			Modal(200, func() { open = false }, func() {
				inner = Container(Attrs(Focusable, FixSize(20, 20)), func() {})
			})
		}
	}

	focusTestFrame(view) // trap mounts, steals, first-stop schedules inner
	focusTestFrame(view) // first-stop lands
	if IdHasFocus(bg) {
		t.Fatal("background must not keep focus while the modal is open")
	}
	if !IdHasFocus(inner) {
		t.Fatal("newly mounted modal should move focus to its first focusable")
	}

	focusTestTab(view, false)
	focusTestFrame(view)
	if IdHasFocus(bg) {
		t.Fatal("Tab must not leave the modal trap for the background")
	}
	if !IdHasFocus(inner) {
		t.Fatal("sole trap focusable should keep focus across Tab")
	}

	open = false
	focusTestFrame(view)
	focusTestTab(view, false)
	focusTestFrame(view)
	if !IdHasFocus(bg) {
		t.Fatal("after dismiss, Tab should reach the background control")
	}
}

func TestTabAfterSplicesSubtreeAfterTarget(t *testing.T) {
	ResetInputSession()
	var a, b, c, d ContainerId
	view := func() {
		a = Container(Attrs(Focusable, FixSize(10, 10)), func() {})
		b = Container(Attrs(Focusable, FixSize(10, 10)), func() {})
		Container(Attrs(TabAfter(a), Gap(2)), func() {
			c = Container(Attrs(Focusable, FixSize(10, 10)), func() {})
			d = Container(Attrs(Focusable, FixSize(10, 10)), func() {})
		})
	}
	focusTestFrame(view)

	focusTestTab(view, false)
	focusTestFrame(view)
	if !IdHasFocus(a) {
		t.Fatal("Tab from nothing should land on a")
	}
	focusTestTab(view, false)
	focusTestFrame(view)
	if !IdHasFocus(c) {
		t.Fatal("TabAfter(a) should put c immediately after a, not b")
	}
	focusTestTab(view, false)
	focusTestFrame(view)
	if !IdHasFocus(d) {
		t.Fatal("Tab should continue through the TabAfter subtree")
	}
	focusTestTab(view, false)
	focusTestFrame(view)
	if !IdHasFocus(b) {
		t.Fatal("after the TabAfter run, Tab should reach b")
	}
}

func TestTabAfterFirstStopOnMount(t *testing.T) {
	ResetInputSession()
	var trigger, inner, page ContainerId
	open := false
	view := func() {
		trigger = Container(Attrs(Focusable, FixSize(10, 10)), func() {
			if FirstRender() {
				Focus()
			}
		})
		page = Container(Attrs(Focusable, FixSize(10, 10)), func() {})
		if open {
			Container(Attrs(TabAfter(trigger)), func() {
				inner = Container(Attrs(Focusable, FixSize(10, 10)), func() {})
			})
		}
	}
	focusTestFrame(view)
	focusTestFrame(view)
	if !IdHasFocus(trigger) {
		t.Fatal("setup: trigger should have focus")
	}
	open = true
	focusTestFrame(view) // TabAfter mounts, first-stop schedules inner
	focusTestFrame(view)
	if !IdHasFocus(inner) {
		t.Fatal("newly mounted TabAfter subtree should take focus from the trigger")
	}
	if IdHasFocus(page) {
		t.Fatal("first-stop should not land on the page control")
	}
}

func TestTabAfterInheritsFocusTrap(t *testing.T) {
	ResetInputSession()
	var bg, trigger, inner ContainerId
	open := true
	panel := false
	view := func() {
		bg = Container(Attrs(Focusable, FixSize(20, 20)), func() {})
		if open {
			Modal(200, func() { open = false }, func() {
				trigger = Container(Attrs(Focusable, FixSize(20, 20)), func() {})
				if panel {
					Popup(func() {
						Container(Attrs(TabAfter(trigger)), func() {
							inner = Container(Attrs(Focusable, FixSize(20, 20)), func() {})
						})
					})
				}
			})
		}
	}
	focusTestFrame(view)
	focusTestFrame(view)
	if IdHasFocus(bg) {
		t.Fatal("setup: modal should trap focus")
	}
	panel = true
	focusTestFrame(view)
	focusTestFrame(view)
	if inner == nil {
		t.Fatal("setup: panel inner was not built")
	}
	if !IdHasFocus(inner) && !IdHasFocus(trigger) {
		t.Fatal("panel inner must be allowed in the modal trap (TabAfter inherits trap owner)")
	}
	// From trigger, Tab should reach inner, not wrap to trigger-only or leak to bg.
	if IdHasFocus(trigger) {
		focusTestTab(view, false)
		focusTestFrame(view)
	}
	if IdHasFocus(bg) {
		t.Fatal("Tab must not leak to the background through a TabAfter popup")
	}
	if !IdHasFocus(inner) && !IdHasFocus(trigger) {
		t.Fatal("focus should stay on trigger or inner, not leave the trap")
	}
}

func TestIdHasFocusWithin(t *testing.T) {
	ResetInputSession()
	var root, child ContainerId
	view := func() {
		root = Container(Attrs(Pad(4)), func() {
			child = Container(Attrs(Focusable, FixSize(10, 10)), func() {
				if FirstRender() {
					Focus()
				}
			})
		})
	}
	focusTestFrame(view)
	focusTestFrame(view)
	if !IdHasFocus(child) {
		t.Fatal("setup: child should have focus")
	}
	if !IdHasFocusWithin(root) {
		t.Fatal("IdHasFocusWithin(root) when a descendant is focused")
	}
	if !IdHasFocusWithin(child) {
		t.Fatal("IdHasFocusWithin includes the node itself")
	}
}
