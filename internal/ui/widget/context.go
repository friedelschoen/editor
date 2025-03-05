package widget

import (
	"image/draw"

	"github.com/veandco/go-sdl2/sdl"
)

type UIContext interface {
	Error(error)

	ImageContext
	CursorContext

	RunOnUIGoRoutine(f func())
}

type ImageContext interface {
	Image() draw.Image
}

type CursorContext interface {
	SetCursor(sdl.SystemCursor)
}
