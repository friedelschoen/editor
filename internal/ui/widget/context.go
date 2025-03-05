package widget

import (
	"image/draw"

	"github.com/veandco/go-sdl2/sdl"
)

type UIContext interface {
	Error(error)

	ImageContext
	CursorContext
	//	Image() draw.Image // TODO
	//	SetCursor(sdl.SystemCursor) // TODO

	RunOnUIGoRoutine(f func())
	SetClipboardData(string) error
	GetClipboardData() (string, error)
}

type ImageContext interface {
	Image() draw.Image
}

type CursorContext interface {
	SetCursor(sdl.SystemCursor)
}
