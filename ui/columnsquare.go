package ui

import (
	"image"
	"image/draw"

	"github.com/friedelschoen/editor/util/imageutil"
	"github.com/friedelschoen/editor/util/uiutil/event"
	"github.com/friedelschoen/editor/util/uiutil/widget"
)

type ColumnSquare struct {
	widget.ENode
	Size image.Point
	col  *Column
}

func NewColumnSquare(col *Column) *ColumnSquare {
	sq := &ColumnSquare{col: col, Size: image.Point{5, 5}}
	sq.Cursor = event.CloseCursor
	return sq
}

func (sq *ColumnSquare) Measure(hint image.Point) image.Point {
	return imageutil.MinPoint(sq.Size, hint)
}
func (sq *ColumnSquare) Paint() {
	c := sq.TreeThemePaletteColor("columnsquare")
	draw.Draw(sq.col.ui.Image(), sq.Bounds, image.NewUniform(c), image.Point{}, draw.Src)
}
func (sq *ColumnSquare) OnInputEvent(ev any, p image.Point) event.Handled {
	switch t := ev.(type) {
	case *event.MouseClick:
		switch t.Button {
		case event.ButtonLeft, event.ButtonMiddle, event.ButtonRight:
			sq.col.Close()
		}
	}
	return true
}
