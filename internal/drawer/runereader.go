package drawer

import (
	"image"
	"io"

	"github.com/friedelschoen/glake/internal/ioutil"
	"golang.org/x/image/math/fixed"
)

type RuneReader struct {
	d *TextDrawer
}

func (rr *RuneReader) Init() {
	rr.d.st.runeR.fface = rr.d.fface
	rr.d.st.runeR.pen = rr.startingPen()
	rr.d.st.runeR.startRi = -1
}

func (rr *RuneReader) Iter() {
	// initialize start ri
	if rr.d.st.runeR.startRi == -1 {
		rr.d.st.runeR.startRi = rr.d.st.runeR.ri
	}

	ru, size, err := ioutil.ReadRuneAt(rr.d.reader, rr.d.st.runeR.ri)
	if err != nil {
		// run last advanced position (draw/delayeddraw/selecting)
		if err == io.EOF {
			_ = rr.iter2(eofRune, 0)
		}
		rr.d.iterStop()
		return
	}
	_ = rr.iter2(ru, size)
}

func (rr *RuneReader) End() {}

func (rr *RuneReader) iter2(ru rune, size int) bool {
	st := &rr.d.st.runeR
	st.ru = ru

	// add/subtract kern with previous rune
	k := rr.d.st.runeR.fface.Kern(st.prevRu, st.ru)
	st.kern = fixed.Int52_12(k << 6)
	st.pen.X += st.kern

	st.advance = rr.tabbedGlyphAdvance(st.ru)

	if !rr.d.iterNext() {
		return false
	}

	// advance for next rune
	st.ri += size
	st.prevRu = st.ru
	st.pen.X += st.advance

	return true
}

func (rr *RuneReader) insertExtraString(s string) bool {
	rr.pushExtra()
	defer rr.popExtra()

	for _, ru := range s {
		if !rr.iter2(ru, len(string(ru))) {
			return false
		}
	}
	return true
}

func (rr *RuneReader) pushExtra() {
	rr.d.st.runeR.extra++
}
func (rr *RuneReader) popExtra() {
	rr.d.st.runeR.extra--
}
func (rr *RuneReader) isExtra() bool {
	return rr.d.st.runeR.extra > 0
}
func (rr *RuneReader) isNormal() bool {
	return !rr.isExtra()
}

func (rr *RuneReader) glyphAdvance(ru rune) fixed.Int52_12 {
	adv, ok := rr.d.st.runeR.fface.GlyphAdvance(ru)
	if !ok {
		return 0
	}
	return fixed.Int52_12(adv << 6)
}

func (rr *RuneReader) tabbedGlyphAdvance(ru rune) fixed.Int52_12 {
	adv := rr.glyphAdvance(ru)
	if ru == '\t' {
		adv = rr.nextTabStopAdvance(rr.d.st.runeR.pen.X, adv)
	}
	return adv
}

func (rr *RuneReader) nextTabStopAdvance(penx, tadv fixed.Int52_12) fixed.Int52_12 {
	// avoid divide by zero
	if tadv == 0 {
		return 0
	}

	px := penx - rr.startingPen().X
	x := px + tadv
	n := int(x / tadv)
	nadv := fixed.Int52_12(n) * tadv
	return nadv - px
}

func (rr *RuneReader) penBounds() fixed.Rectangle52_12 {
	st := &rr.d.st.runeR
	minX, minY := st.pen.X, st.pen.Y
	maxX, maxY := minX+st.advance, minY+rr.d.lineHeight
	min := fixed.Point52_12{X: minX, Y: minY}
	max := fixed.Point52_12{X: maxX, Y: maxY}
	return fixed.Rectangle52_12{Min: min, Max: max}
}

func (rr *RuneReader) penBoundsRect() image.Rectangle {
	rf := rr.penBounds()
	// expand min (use floor), and max (use ceil)

	min := image.Point{
		X: rf.Min.X.Ceil(),
		Y: rf.Min.Y.Ceil(),
	}
	max := image.Point{
		X: rf.Max.X.Floor(),
		Y: rf.Max.Y.Floor(),
	}
	return image.Rectangle{Min: min, Max: max}
}

func (rr *RuneReader) startingPen() fixed.Point52_12 {
	p := rr.d.bounds.Min
	p.X += rr.d.Opt.RuneReader.StartOffsetX
	if rr.d.st.runeR.ri == 0 {
		p.X += rr.d.firstLineOffsetX
	}
	return fixed.Point52_12{
		X: fixed.Int52_12(p.X << 12),
		Y: fixed.Int52_12(p.Y << 12),
	}
}

func (rr *RuneReader) maxX() fixed.Int52_12 {
	return fixed.Int52_12(rr.d.bounds.Max.X << 12)
}
