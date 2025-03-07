package ui

import (
	"image"
	"image/draw"

	"github.com/friedelschoen/editor/internal/drawer"
	"github.com/friedelschoen/editor/internal/ui/driver"
	"github.com/friedelschoen/editor/internal/ui/widget"
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
	return drawer.MinPoint(sq.Size, hint)
}
func (sq *ColumnSquare) Paint() {
	c := sq.TreeThemePaletteColor("columnsquare")
	draw.Draw(sq.col.ui.Image(), sq.Bounds, image.NewUniform(c), image.Point{}, draw.Src)
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
