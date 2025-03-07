package ui

import (
	"image"

	"github.com/friedelschoen/editor/internal/ui/driver"
)

type Toolbar struct {
	*TextArea
	warpPointerOnNextLayout bool
}

func NewToolbar(ui *UI) *Toolbar {
	tb := &Toolbar{}
	tb.TextArea = NewTextArea(ui)
	tb.SetThemePaletteNamePrefix("toolbar_")
	tb.TextArea.Drawer.Opt.EarlyExitMeasure = true // performance
	return tb
}

func (tb *Toolbar) OnInputEvent(ev driver.Event, p image.Point) bool {
	switch ev.(type) {
	case *driver.KeyDown, *driver.KeyUp:
		// allow typing in the toolbar (dynamic size) without losing focus
		// It is incorrect to do this via rw callback since, for example, restoring a session (writes the toolbar) would trigger the possibility of warping the pointer.
		tb.keepPointerInsideToolbar()
	}
	return tb.TextArea.OnInputEvent(ev, p)
}

func (tb *Toolbar) keepPointerInsideToolbar() {
	p, err := tb.ui.QueryPointer()
	if err == nil && p.In(tb.Bounds) {
		tb.warpPointerOnNextLayout = true
		tb.MarkNeedsLayout()
	}
}

func (tb *Toolbar) Layout() {
	tb.TextArea.Layout()

	// warp pointer to inside the toolbar
	if tb.warpPointerOnNextLayout {
		tb.warpPointerOnNextLayout = false
		p, err := tb.ui.QueryPointer()
		if err == nil && !p.In(tb.Bounds) {
			tb.ui.WarpPointerToRectanglePad(tb.Bounds)
		}
	}
}
