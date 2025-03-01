package drawer4

import (
	"image"
	"io"

	"github.com/friedelschoen/glake/internal/geometry"
	"github.com/friedelschoen/glake/internal/io/iorw"
)

type RuneReader struct {
	d *Drawer
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

	ru, size, err := iorw.ReadRuneAt(rr.d.reader, rr.d.st.runeR.ri)
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
	st.kern = geometry.Intf2(k)
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

func (rr *RuneReader) glyphAdvance(ru rune) geometry.Intf {
	adv, ok := rr.d.st.runeR.fface.GlyphAdvance(ru)
	if !ok {
		return 0
	}
	return geometry.Intf2(adv)
}

func (rr *RuneReader) tabbedGlyphAdvance(ru rune) geometry.Intf {
	adv := rr.glyphAdvance(ru)
	if ru == '\t' {
		adv = rr.nextTabStopAdvance(rr.d.st.runeR.pen.X, adv)
	}
	return adv
}

func (rr *RuneReader) nextTabStopAdvance(penx, tadv geometry.Intf) geometry.Intf {
	// avoid divide by zero
	if tadv == 0 {
		return 0
	}

	px := penx - rr.startingPen().X
	x := px + tadv
	n := int(x / tadv)
	nadv := geometry.Intf(n) * tadv
	return nadv - px
}

func (rr *RuneReader) penBounds() geometry.RectangleIntf {
	st := &rr.d.st.runeR
	minX, minY := st.pen.X, st.pen.Y
	maxX, maxY := minX+st.advance, minY+rr.d.lineHeight
	min := geometry.PointIntf{minX, minY}
	max := geometry.PointIntf{maxX, maxY}
	return geometry.RectangleIntf{min, max}
}

func (rr *RuneReader) penBoundsRect() image.Rectangle {
	pb := rr.penBounds()
	// expand min (use floor), and max (use ceil)
	return pb.ToRectFloorCeil()
}

func (rr *RuneReader) startingPen() geometry.PointIntf {
	p := rr.d.bounds.Min
	p.X += rr.d.Opt.RuneReader.StartOffsetX
	if rr.d.st.runeR.ri == 0 {
		p.X += rr.d.firstLineOffsetX
	}
	return geometry.PIntf2(p)
}

func (rr *RuneReader) maxX() geometry.Intf {
	return geometry.Intf1(rr.d.bounds.Max.X)
}
