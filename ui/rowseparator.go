package ui

import (
	"image"

	"github.com/jmigpin/editor/util/uiutil/event"
	"github.com/jmigpin/editor/util/uiutil/widget"
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
	sep.Handle.Cursor = event.MoveCursor

	rsep := &RowSeparator{Separator: sep, row: row}
	rsep.SetThemePaletteNamePrefix("rowseparator_")
	return rsep
}
func (sh *RowSeparator) OnInputEvent(ev0 any, p image.Point) event.Handled {
	switch ev := ev0.(type) {
	case *event.MouseDragMove:
		switch {
		case ev.Buttons.Is(event.ButtonLeft):
			p.Y += sh.Handle.DragPad.Y
			sh.row.resizeWithMoveToPoint(&p)
		}
	case *event.MouseDown:
		m := ev.Mods.ClearLocks()
		if m.IsEmpty() {
			switch ev.Button {
			case event.ButtonWheelUp:
				sh.row.resizeWithPushJump(true, &p)
			case event.ButtonWheelDown:
				sh.row.resizeWithPushJump(false, &p)
			}
		}
	case *event.MouseClick:
		switch ev.Button {
		case event.ButtonMiddle:
			sh.row.Close()
		}
	}
	return true //no other widget will get the event
}
