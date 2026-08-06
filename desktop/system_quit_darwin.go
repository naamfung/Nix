//go:build darwin

package main

/*
#cgo darwin LDFLAGS: -framework Cocoa
void installInxSystemQuitHook(void);
*/
import "C"

import "sync"

var installSystemQuitHookOnce sync.Once

func installSystemQuitHook() {
	installSystemQuitHookOnce.Do(func() {
		C.installInxSystemQuitHook()
	})
}

//export InxMarkSystemQuit
func InxMarkSystemQuit() {
	markSystemQuitRequested()
}
