package ui

import (
	"image"

	"github.com/friedelschoen/editor/internal/ui/driver"
	"github.com/friedelschoen/editor/internal/ui/widget"
	"github.com/veandco/go-sdl2/sdl"
)

type ColSeparator struct {
	*widget.Separator
	col *Column
}

func NewColSeparator(col *Column) *ColSeparator {
	sep := widget.NewSeparator(col.ui, col.Cols.Root.MultiLayer)
	sep.Size.X = separatorWidth
	sep.Handle.Left = 3
	sep.Handle.Right = 3
	sep.Handle.Cursor = sdl.SYSTEM_CURSOR_SIZEWE

	csep := &ColSeparator{Separator: sep, col: col}
	csep.SetThemePaletteNamePrefix("colseparator_")
	return csep
}
func (sh *ColSeparator) OnInputEvent(ev0 driver.Event, p image.Point) bool {
	switch ev := ev0.(type) {
	case *driver.MouseDragMove:
		switch {
		case ev.Key.HasMouse(driver.ButtonLeft):
			p.X += sh.Handle.DragPad.X
			sh.col.resizeWithMoveToPoint(&p)
		}
	case *driver.MouseWheel:
		if ev.X < 0 {
			sh.col.resizeWithMoveJump(true, &p)
		} else if ev.X > 0 {
			sh.col.resizeWithMoveJump(false, &p)
		}
	}
	return true // no other widget will get the event
}
