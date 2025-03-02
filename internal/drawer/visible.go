package drawer

import (
	"image"

	"golang.org/x/image/math/fixed"
)

func header1PenBounds(d *TextDrawer, offset int) (fixed.Rectangle52_12, bool) {
	d.st = State{}
	fnIter := FnIter{}
	iters := append(d.sIters(true), &fnIter)
	d.loopInit(iters)
	d.header1()

	found := false
	pen := fixed.Rectangle52_12{}
	fnIter.fn = func() {
		if d.iters.runeR.isNormal() {
			if d.st.runeR.ri >= offset {
				if d.st.runeR.ri == offset {
					found = true
					pen = d.iters.runeR.penBounds()
				}
				d.iterStop()
				return
			}
		}
		if !d.iterNext() {
			return
		}
	}

	d.loop()

	return pen, found
}

type PenVisibility int

const (
	VisibilityNot PenVisibility = iota
	VisibilityFull
	VisibilityPartial
)

// type PenVisibility struct {
// 	not     bool // not visible
// 	full    bool // fully visible
// 	partial bool // partially visible
// 	top     bool // otherwise is bottom, valid in "full" and "partial"
// }

func penVisibility(d *TextDrawer, offset int) (PenVisibility, bool) {
	pb, ok := header1PenBounds(d, offset)
	if !ok {
		return VisibilityNot, false
	} else {
		min := image.Point{
			X: pb.Min.X.Ceil(),
			Y: pb.Min.Y.Ceil(),
		}
		max := image.Point{
			X: pb.Max.X.Floor(),
			Y: pb.Max.Y.Floor(),
		}
		pr := image.Rectangle{Min: min, Max: max}
		// allow intersection of empty x in penbounds (case of eof)
		if pr.Dx() == 0 {
			pr.Max.X = pr.Min.X + 1
		}

		// consider previous/next lines (allows cursor up/down to move 1 line instead of jumping the view aligned to the center)
		b := d.bounds // copy
		b.Min.Y--
		b.Max.Y++

		ir := b.Intersect(pr)
		if ir.Empty() {
			return VisibilityNot, false
		} else if ir == pr {
			return VisibilityFull, false
		} else {
			if pr.Min.Y < b.Min.Y {
				return VisibilityPartial, true
			}
			return VisibilityPartial, false
		}
	}
}
