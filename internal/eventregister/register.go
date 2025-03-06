package eventregister

import (
	"slices"
	"sync"
)

// The zero register is empty and ready for use.
type Register struct {
	sync.RWMutex
	m map[int][]*Callback
}

// Remove is done via *Regist.Unregister().
func (reg *Register) Add(evId int, fn func(any)) *Regist {
	return reg.AddCallback(evId, &Callback{fn})
}

func (reg *Register) AddCallback(evId int, cb *Callback) *Regist {
	reg.Lock()
	defer reg.Unlock()
	if reg.m == nil {
		reg.m = map[int][]*Callback{}
	}
	_, ok := reg.m[evId]
	if !ok {
		reg.m[evId] = make([]*Callback, 0)
	}
	reg.m[evId] = append(reg.m[evId], cb)
	return &Regist{reg, evId, cb}
}

func (reg *Register) RemoveCallback(evId int, cb *Callback) {
	reg.Lock()
	defer reg.Unlock()
	if reg.m == nil {
		return
	}
	l, ok := reg.m[evId]
	if !ok {
		return
	}
	// iterate to remove since the callback doesn't keep the element (allows callback to be added more then once, or at different evId's - this is probably a useless feature unless the *callback is being used to also be set in a map)
	newl := slices.DeleteFunc(l, func(cb2 *Callback) bool { return cb2 == cb })
	if len(newl) > 0 {
		reg.m[evId] = newl
	} else {
		delete(reg.m, evId)
	}
}

// Returns number of callbacks done.
func (reg *Register) RunCallbacks(evId int, ev any) int {
	reg.RLock()
	defer reg.RUnlock()
	if reg.m == nil {
		return 0
	}
	l, ok := reg.m[evId]
	if !ok {
		return 0
	}
	c := 0
	for _, cb := range l {
		cb.F(ev)
		c++
	}
	return c
}

// Number of registered callbacks for an event id.
func (reg *Register) NCallbacks(evId int) int {
	reg.RLock()
	defer reg.RUnlock()
	if reg.m == nil {
		return 0
	}
	l, ok := reg.m[evId]
	if !ok {
		return 0
	}
	return len(l)
}

type Callback struct {
	F func(ev any)
}

type Regist struct {
	evReg *Register
	id    int
	cb    *Callback
}

func (reg *Regist) Unregister() {
	reg.evReg.RemoveCallback(reg.id, reg.cb)
}

// Utility to unregister big number of regists.
type Unregister struct {
	v []*Regist
}
