package syncutil

import (
	"sync"
	"time"
)

// Continously usable, instantiated once for many wait()/set() calls. Fails if wait() is not ready when set() is called.
// Usage:
//
//	w:=NewWaitForSet()
//	w.Start(5*time.Second)
//	...
//	// sync/async call to w.Set()
//	...
//	v,err := w.WaitForSet()
//	if err!=nil {
//	}
type WaitForSet struct {
	d struct {
		sync.Mutex
		get struct {
			timer   *time.Timer
			waiting bool
		}
		set struct {
			gotV bool
			v    any
		}
	}
	cond *sync.Cond // signals from set() or timeout()
}

// In case waitforset() is not going to be called.

// Fails if not able to set while get() is ready.
