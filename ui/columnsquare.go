package ui

import (
	"image"

	"github.com/jmigpin/editor/ui/driver"
	"github.com/jmigpin/editor/ui/widget"
	"github.com/jmigpin/editor/util/imageutil"
	"github.com/veandco/go-sdl2/sdl"
)

type ColumnSquare struct {
	widget.ENode
	Size image.Point
	col  *Column
}

func NewColumnSquare(col *Column) *ColumnSquare {
	sq := &ColumnSquare{col: col, Size: image.Point{5, 5}}
	sq.Cursor = sdl.SYSTEM_CURSOR_NO
	return sq
}

func (sq *ColumnSquare) Measure(hint image.Point) image.Point {
	return imageutil.MinPoint(sq.Size, hint)
}
func (sq *ColumnSquare) Paint() {
	c := sq.TreeThemePaletteColor("columnsquare")
	imageutil.FillRectangle(sq.col.ui.Image(), sq.Bounds, c)
}
func (sq *ColumnSquare) OnInputEvent(ev driver.Event, p image.Point) bool {
	switch t := ev.(type) {
	case *driver.MouseClick:
		switch t.Key.Mouse {
		case driver.ButtonLeft, driver.ButtonMiddle, driver.ButtonRight:
			sq.col.Close()
		}
	}
	return true
}
