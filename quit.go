package shirei

import "sync"

var quitRequests struct {
	sync.RWMutex
	handler func()
}

// SetQuitHandler lets an application defer native window close, menu Quit and
// app.Quit for asynchronous save/recovery. A nil handler restores normal exit.
// Handlers should schedule work and return promptly, keeping the event loop live.
func SetQuitHandler(handler func()) {
	quitRequests.Lock()
	quitRequests.handler = handler
	quitRequests.Unlock()
}

// HandleQuitRequest reports whether the application took ownership of quitting.
// A handled request must leave its window open until the application exits.
func HandleQuitRequest() bool {
	quitRequests.RLock()
	handler := quitRequests.handler
	quitRequests.RUnlock()
	if handler == nil {
		return false
	}
	handler()
	return true
}
