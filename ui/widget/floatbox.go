package widget

import (
	"image"

	"github.com/friedelschoen/glake/util/drawutil"
	"github.com/veandco/go-sdl2/sdl"
)

// Should be a child of FloatLayer.
type FloatBox struct {
	ENode
	RefPoint image.Point
	content  Node
	ml       *MultiLayer
	MaxSize  image.Point
}

func NewFloatBox(ml *MultiLayer, content Node) *FloatBox {
	fb := &FloatBox{content: content, ml: ml}
	fb.Cursor = sdl.SYSTEM_CURSOR_ARROW
	fb.Append(content)
	fb.AddMarks(MarkNotDraggable | MarkInBoundsHandlesEvent)
	return fb
}

func (fb *FloatBox) Visible() bool {
	return !fb.HasAnyMarks(MarkForceZeroBounds)
}

func (fb *FloatBox) Hide() {
	fb.ml.AddMarkRect(fb.Bounds)
	fb.AddMarks(MarkForceZeroBounds)
	fb.MarkNeedsLayoutAndPaint()
}

func (fb *FloatBox) Show() {
	fb.RemoveMarks(MarkForceZeroBounds)
	fb.MarkNeedsLayoutAndPaint()
}

func (fb *FloatBox) Toggle() {
	if !fb.Visible() {
		fb.Show()
	} else {
		fb.Hide()
	}
}

func (fb *FloatBox) Measure(hint image.Point) image.Point {
	panic("calling measure on floatbox")
}

func (fb *FloatBox) Layout() {
	b := fb.Bounds

	// max size
	max := b.Size()
	if fb.MaxSize != (image.Point{}) {
		max = drawutil.MinPoint(max, fb.MaxSize)
	}
	m := fb.content.Measure(max)

	// calc bounds attached to the reference point
	r := image.Rectangle{}
	r = r.Add(fb.RefPoint)
	r.Max = r.Min.Add(m)
	if r.Max.X > b.Max.X {
		diffX := r.Max.X - b.Max.X
		r = r.Sub(image.Point{diffX, 0})
	}
	r = r.Intersect(b)

	fb.content.Embed().Bounds = r
	fb.Bounds = r // reduce own bounds to contain events
}
