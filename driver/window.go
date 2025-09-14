package driver

import (
	"image"
	"image/draw"

	"github.com/friedelschoen/editor/util/uiutil/event"
)

type Window interface {
	NextEvent() (_ event.Event, ok bool) // !ok = no more events

	Close() error
	WindowSetName(Name string) error
	Image() (draw.Image, error)
	ImagePut(Rect image.Rectangle) error
	ImageResize(Rect image.Rectangle) error
	CursorSet(Cursor event.Cursor) error
	PointerQuery() (image.Point, error)
	PointerWarp(P image.Point) error
	ClipboardDataGet() (string, error)
	ClipboardDataSet(str string) error
}
