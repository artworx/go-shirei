package app

import (
	g "go.hasen.dev/generic"
	"go.hasen.dev/shirei"
)

// Quit requests process exit. An installed Shirei quit handler may defer it.
// Safe to call from any goroutine.
// AddExitCleanup handlers run first.
func Quit() {
	if !shirei.HandleQuitRequest() {
		g.ExitWithCleanup(0)
	}
}
