package driver

import (
	"image"
	"image/draw"

	"github.com/jmigpin/editor/ui/event"
)

type Window interface {
	NextEvent() (event.Event, bool) // !ok = no more events

	Close() error
	Update()
	WindowSetName(Name string) error
	Resize(rect image.Rectangle) error
	Image() (draw.Image, error)
	CursorSet(Cursor event.Cursor) error

	PointerQuery() (image.Point, error)
	PointerWarp(P image.Point) error

	ClipboardDataGet() (string, error)
	ClipboardDataSet(text string) error
}
