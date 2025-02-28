package ui

import (
	"image"

	"github.com/jmigpin/editor/ui/event"
	"github.com/jmigpin/editor/ui/widget"
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
func (sh *RowSeparator) OnInputEvent(ev0 event.Event, p image.Point) bool {
	switch ev := ev0.(type) {
	case *event.MouseDragMove:
		if ev.Key.HasMouse(event.ButtonLeft) {
			p.Y += sh.Handle.DragPad.Y
			sh.row.resizeWithMoveToPoint(&p)
		}
	case *event.MouseWheel:
		if ev.Y < 0 {
			sh.row.resizeWithPushJump(true, &p)
		} else if ev.Y > 0 {
			sh.row.resizeWithPushJump(false, &p)
		}
	case *event.MouseClick:
		if ev.Key.Mouse == event.ButtonMiddle {
			sh.row.Close()
		}
	}
	return true //no other widget will get the event
}
