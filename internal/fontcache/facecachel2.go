package fontcache

import (
	"sync"

	"golang.org/x/image/font"
)

// Same as FaceCacheL but with sync.map.
type FaceCacheL2 struct {
	font.Face
	mu  sync.RWMutex
	gc  sync.Map
	gac sync.Map
	gbc sync.Map
	kc  sync.Map // kern cache
}
