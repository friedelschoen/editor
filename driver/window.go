package driver

import (
	"image"
	"image/draw"

	"github.com/jmigpin/editor/uiutil/event"
	"github.com/jmigpin/editor/uiutil/widget"
)

type Window interface {
	EventLoop(events chan<- interface{}) // should emit events from uiutil/event

	Close()
	SetWindowName(string)

	Image() draw.Image
	PutImage(*image.Rectangle)
	UpdateImageSize() error

	SetCursor(widget.Cursor)
	QueryPointer() (*image.Point, error)
	WarpPointer(*image.Point)

	// copypaste
	GetCPPaste(event.CopyPasteIndex) (string, error)
	SetCPCopy(event.CopyPasteIndex, string) error
}
