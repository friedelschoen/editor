package ui

import (
	"image"

	"github.com/friedelschoen/editor/internal/ui/driver"
	"github.com/friedelschoen/editor/internal/ui/widget"
	"github.com/veandco/go-sdl2/sdl"
)

type RowSeparator struct {
	*widget.Separator
	row *Row
}

func NewRowSeparator(row *Row) *RowSeparator {
	sep := widget.NewSeparator(row.ui, row.Col.Cols.Root.MultiLayer)
	sep.Size.Y = separatorWidth
	sep.Handle.Top = 3
	sep.Handle.Bottom = 3
	sep.Handle.Cursor = sdl.SYSTEM_CURSOR_CROSSHAIR

	rsep := &RowSeparator{Separator: sep, row: row}
	rsep.SetThemePaletteNamePrefix("rowseparator_")
	return rsep
}
func (sh *RowSeparator) OnInputEvent(ev0 driver.Event, p image.Point) bool {
	switch ev := ev0.(type) {
	case *driver.MouseDragMove:
		if ev.Key.HasMouse(driver.ButtonLeft) {
			p.Y += sh.Handle.DragPad.Y
			sh.row.resizeWithMoveToPoint(&p)
		}
	case *driver.MouseWheel:
		if ev.Y < 0 {
			sh.row.resizeWithPushJump(true, &p)
		} else if ev.Y > 0 {
			sh.row.resizeWithPushJump(false, &p)
		}
	case *driver.MouseClick:
		if ev.Key.Mouse == driver.ButtonMiddle {
			sh.row.Close()
		}
	}
	return true //no other widget will get the event
}
