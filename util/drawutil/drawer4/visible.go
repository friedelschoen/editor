package drawer4

import (
	"github.com/jmigpin/editor/util/mathutil"
)

func header1PenBounds(d *Drawer, offset int) (mathutil.RectangleIntf, bool) {
	d.st = State{}
	fnIter := FnIter{}
	iters := append(d.sIters(true), &fnIter)
	d.loopInit(iters)
	d.header1()

	found := false
	pen := mathutil.RectangleIntf{}
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

func penVisibility(d *Drawer, offset int) (PenVisibility, bool) {
	pb, ok := header1PenBounds(d, offset)
	if !ok {
		return VisibilityNot, false
	} else {
		pr := pb.ToRectFloorCeil()
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
