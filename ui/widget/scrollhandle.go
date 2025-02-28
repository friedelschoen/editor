package widget

import (
	"image"
	"image/color"

	"github.com/jmigpin/editor/ui/driver"
	"github.com/jmigpin/editor/util/imageutil"
)

// Used by ScrollBar.
type ScrollHandle struct {
	ENode
	ctx    ImageContext
	sb     *ScrollBar
	inside bool
}

func NewScrollHandle(ctx ImageContext, sb *ScrollBar) *ScrollHandle {
	sh := &ScrollHandle{ctx: ctx, sb: sb}

	// the scrollbar handles the decision making, the handle only draws
	sh.AddMarks(MarkNotDraggable)

	return sh
}

func (sh *ScrollHandle) Paint() {
	var c color.Color
	if sh.sb.clicking || sh.sb.dragging {
		c = sh.TreeThemePaletteColor("scrollhandle_select")
	} else if sh.inside {
		c = sh.TreeThemePaletteColor("scrollhandle_hover")
	} else {
		c = sh.TreeThemePaletteColor("scrollhandle_normal")
	}
	imageutil.FillRectangle(sh.ctx.Image(), sh.Bounds, c)
}

func (sh *ScrollHandle) OnInputEvent(ev driver.Event, p image.Point) bool {
	switch ev.(type) {
	case *driver.MouseEnter:
		sh.inside = true
		sh.MarkNeedsPaint()
	case *driver.MouseLeave:
		sh.inside = false
		sh.MarkNeedsPaint()
	}
	return false
}
