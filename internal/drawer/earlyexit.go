package drawer

import "golang.org/x/image/math/fixed"

type EarlyExit struct {
	d *TextDrawer
}

func (ee *EarlyExit) Init() {}

func (ee *EarlyExit) Iter() {
	maxY := fixed.Int52_12(ee.d.bounds.Max.Y << 12)

	// extra line iterated (not visible, useful for header1)
	if ee.d.st.earlyExit.extraLine {
		maxY += ee.d.lineHeight
	}

	if ee.d.st.runeR.pen.Y >= maxY {
		ee.d.iterStop()
		return
	}
	if !ee.d.iterNext() {
		return
	}
}

func (ee *EarlyExit) End() {}
