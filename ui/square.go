package ui

import (
	"image"
	"image/color"

	// only for cursordef

	"github.com/jmigpin/editor/imageutil"
	"github.com/jmigpin/editor/uiutil/event"
	"github.com/jmigpin/editor/uiutil/widget"
	"github.com/jmigpin/editor/xgbutil/evreg"
)

// Used in row to move, close, and show row states.
type Square struct {
	widget.LeafEmbedNode
	EvReg    *evreg.Register
	Width    int
	Triangle bool

	ui       *UI
	values   [7]bool // bg and mini-squares
	pressPad image.Point
}

func NewSquare(ui *UI) *Square {
	sq := &Square{ui: ui}
	sq.Width = SquareWidth
	sq.EvReg = evreg.NewRegister()
	return sq
}

func (sq *Square) Measure(hint image.Point) image.Point {
	return image.Point{sq.Width, sq.Width}
}
func (sq *Square) Paint() {
	// full background
	var c color.Color = SquareColor
	if sq.values[SquareEdited] {
		c = SquareEditedColor
	}
	if sq.values[SquareNotExist] {
		c = SquareNotExistColor
	}
	if sq.values[SquareExecuting] {
		c = SquareExecutingColor
	}
	bounds := sq.Bounds()
	imageutil.FillRectangle(sq.ui.Image(), &bounds, c)

	//// triangle
	//b2 := bounds
	//b2.Max.X = b2.Min.X + 1
	//b2.Min.Y = b2.Min.Y + sq.Width
	//b2.Max.Y = b2.Min.Y + 1
	//for x := 0; x < sq.Width; x++ {
	//	imageutil.FillRectangle(sq.ui.Image(), &b2, colornames.Grey)
	//	b2 = b2.Add(image.Point{1, -1})
	//	//b2.Min.Y++
	//}

	// duplicate
	if sq.values[SquareDuplicate] {
		c2 := SquareEditedColor
		imageutil.BorderRectangle(sq.ui.Image(), &bounds, c2, 2)
	}

	// mini-squares
	miniSq := func(i int) *image.Rectangle {
		r := bounds
		w := (r.Max.X - r.Min.X) / 2
		r.Max.X = r.Min.X + w
		r.Max.Y = r.Min.Y + w
		switch i {
		case 0:
		case 1:
			r.Min.X += w
			r.Max.X += w
		case 2:
			r.Min.Y += w
			r.Max.Y += w
		case 3:
			r.Min.X += w
			r.Max.X += w
			r.Min.Y += w
			r.Max.Y += w
		}
		return &r
	}
	if sq.values[SquareDiskChanges] {
		u := 0
		if ScrollbarLeft {
			u = 1
		}
		r := miniSq(u)
		imageutil.FillRectangle(sq.ui.Image(), r, SquareDiskChangesColor)
	}
	if sq.values[SquareActive] {
		u := 1
		if ScrollbarLeft {
			u = 0
		}
		r := miniSq(u)
		imageutil.FillRectangle(sq.ui.Image(), r, SquareActiveColor)
	}
}

func (sq *Square) OnInputEvent(ev interface{}, p image.Point) bool {
	// press point pad
	switch evt := ev.(type) {
	case *event.MouseDown:
		u := evt.Point.Sub(sq.Bounds().Min)
		if !ScrollbarLeft {
			u.X = evt.Point.X - sq.Bounds().Max.X
		}
		sq.pressPad = u
	}

	sq.runCallbacks(ev, p)

	return false
}
func (sq *Square) runCallbacks(ev interface{}, p image.Point) {
	// input event for registered callbacks
	topPoint := p.Sub(sq.pressPad)
	topXPoint := image.Point{p.X, topPoint.Y}
	ev2 := &SquareInputEvent{ev, &p, &topPoint, &topXPoint}
	sq.EvReg.RunCallbacks(SquareInputEventId, ev2)
}

func (sq *Square) WarpPointer() {
	sa := sq.Bounds()
	p := sa.Min.Add(sa.Max.Sub(sa.Min).Div(2))
	sq.ui.WarpPointer(&p)
}

func (sq *Square) Value(t SquareType) bool {
	return sq.values[t]
}
func (sq *Square) SetValue(t SquareType, v bool) {
	if sq.values[t] != v {
		sq.values[t] = v
		sq.MarkNeedsPaint()
	}
}

type SquareType int

const (
	SquareNone SquareType = iota
	SquareActive
	SquareExecuting
	SquareEdited
	SquareDiskChanges
	SquareNotExist
	SquareDuplicate
)

const (
	SquareInputEventId = iota
)

type SquareInputEvent struct {
	Event     interface{}
	Point     *image.Point
	TopPoint  *image.Point
	TopXPoint *image.Point
}
