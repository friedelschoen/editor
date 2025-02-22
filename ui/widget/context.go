package widget

import (
	"image/draw"

	"github.com/jmigpin/editor/ui/event"
)

type UIContext interface {
	Error(error)

	ImageContext
	CursorContext
	//	Image() draw.Image // TODO
	//	SetCursor(event.Cursor) // TODO

	RunOnUIGoRoutine(f func())
	SetClipboardData(string)
	GetClipboardData() (string, error)
}

type ImageContext interface {
	Image() draw.Image
}

type CursorContext interface {
	SetCursor(event.Cursor)
}
