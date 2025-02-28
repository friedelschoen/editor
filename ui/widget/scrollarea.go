package widget

import (
	"image"

	"github.com/jmigpin/editor/ui/event"
	"github.com/jmigpin/editor/util/imageutil"
)

type ScrollArea struct {
	ENode
	ScrollWidth int
	LeftScroll  bool
	YBar        *ScrollBar
	XBar        *ScrollBar

	scrollable ScrollableNode
	ctx        ImageContext
}

func NewScrollArea(ctx ImageContext, scrollable ScrollableNode, xbar, ybar bool) *ScrollArea {
	sa := &ScrollArea{
		ScrollWidth: 10,
		LeftScroll:  true,
		scrollable:  scrollable,
		ctx:         ctx,
	}
	sa.SetBars(xbar, ybar)
	sa.Append(sa.scrollable)
	return sa
}

func (sa *ScrollArea) SetBars(xbar, ybar bool) {
	if xbar && sa.XBar == nil {
		sa.XBar = NewScrollBar(sa.ctx, sa)
		sa.XBar.Horizontal = true
		sa.Append(sa.XBar)

	}
	if !xbar && sa.XBar != nil {
		sa.Remove(sa.XBar)
		sa.XBar = nil
	}
	if ybar && sa.YBar == nil {
		sa.YBar = NewScrollBar(sa.ctx, sa)
		sa.Append(sa.YBar)
	}
	if !ybar && sa.YBar != nil {
		sa.Remove(sa.YBar)
		sa.YBar = nil
	}
	sa.scrollable.SetScrollable(xbar, ybar)
}

func (sa *ScrollArea) scrollPageUp()   { sa.scrollPage(true) }
func (sa *ScrollArea) scrollPageDown() { sa.scrollPage(false) }
func (sa *ScrollArea) scrollPage(up bool) {
	if sa.YBar != nil {
		sa.YBar.scrollPage(up)
	}
}

func (sa *ScrollArea) scrollJumpUp()   { sa.scrollJump(true) }
func (sa *ScrollArea) scrollJumpDown() { sa.scrollJump(false) }
func (sa *ScrollArea) scrollJump(up bool) {
	if sa.YBar != nil {
		sa.YBar.scrollWheel(up)
	}
}

func (sa *ScrollArea) OnChildMarked(child Node, newMarks Marks) {
	if child == sa.scrollable {
		if newMarks.HasAny(MarkNeedsLayout) {
			sa.MarkNeedsLayout()
		}
	}
}

func (sa *ScrollArea) Measure(hint image.Point) image.Point {
	// space to reduce due to scrollbars
	var space image.Point
	if sa.YBar != nil {
		space.X = sa.ScrollWidth
	}
	if sa.XBar != nil {
		space.Y = sa.ScrollWidth
	}

	h := hint.Sub(space)
	h = imageutil.MaxPoint(h, image.Point{0, 0})

	//m := sa.ENode.Measure(h)
	m := sa.scrollable.Measure(h)

	m = m.Add(space)
	m = imageutil.MinPoint(m, hint)

	return m
}

func (sa *ScrollArea) Layout() {
	b := sa.Bounds
	if sa.YBar != nil {
		r := b
		if sa.LeftScroll {
			r.Max.X = r.Min.X + sa.ScrollWidth
			b.Min.X = r.Max.X
		} else {
			r.Min.X = r.Max.X - sa.ScrollWidth
			b.Max.X = r.Min.X
		}
		r = r.Intersect(sa.Bounds)
		sa.YBar.Bounds = r
	}
	if sa.XBar != nil {
		r := b
		r.Min.Y = r.Max.Y - sa.ScrollWidth
		b.Max.Y = r.Min.Y
		r = r.Intersect(sa.Bounds)
		sa.XBar.Bounds = r
	}

	// scrollable bounds
	sa.scrollable.Embed().Bounds = b.Intersect(sa.Bounds)
	// ensure scrollable layout is done before the scrollbar since it might be calculated before the scrollable due to child order. The scrollable needs to be aware of its updated bounds to correctly return the viewsize.
	sa.scrollable.Layout()
}

func (sa *ScrollArea) OnInputEvent(ev0 event.Event, p image.Point) bool {
	switch evt := ev0.(type) {
	case *event.KeyDown:
		switch {
		case evt.Key.Is("PageUp"):
			sa.scrollPageUp()
		case evt.Key.Is("PageDown"):
			sa.scrollPageDown()
		default:
			// allow scrollable to receive keydown input
			if !p.In(sa.scrollable.Embed().Bounds) {
				sa.scrollable.OnInputEvent(ev0, p)
			}
		}
	case *event.MouseWheel:
		// scrolling with the wheel on the content area
		if p.In(sa.scrollable.Embed().Bounds) {
			switch {
			case evt.Y < 0:
				sa.scrollJumpUp()
			case evt.Y > 0:
				sa.scrollJumpDown()
			}
		}
	}
	return false
}
