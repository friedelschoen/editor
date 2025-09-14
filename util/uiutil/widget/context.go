package widget

import (
	"image/draw"

	"github.com/friedelschoen/editor/util/uiutil/event"
)

type UIContext interface {
	Error(error)

	ImageContext
	CursorContext
	//	Image() draw.Image // TODO
	//	SetCursor(event.Cursor) // TODO

	RunOnUIGoRoutine(f func())
	SetClipboardData(string)
	GetClipboardData() string
}

type ImageContext interface {
	Image() draw.Image
}

type CursorContext interface {
	SetCursor(event.Cursor)
}
