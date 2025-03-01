package drawer4

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type DrawRune struct {
	d *Drawer
}

func (dr *DrawRune) Init() {}

func (dr *DrawRune) Iter() {
	dr.draw()
	if !dr.d.iterNext() {
		return
	}
}

func (dr *DrawRune) End() {}

func (dr *DrawRune) draw() {
	st := &dr.d.st.drawR

	pen := dr.d.iters.runeR.penBoundsRect().Min

	// draw now
	//dr.draw2(dr.d.st.runeR.fface, pen, dr.d.st.runeR.ru, dr.d.st.curColors.fg)
	//return

	// delayed draw
	if st.delay != nil {
		dr.draw2(st.delay.fface, st.delay.pen, st.delay.ru, st.delay.fg)
	}

	// delay drawing by one rune to allow drawing the kern bg correctly. The last position is also drawn because the runereader emits a final ru=0 at the end
	st.delay = &DrawRuneDelay{
		pen:   pen,
		ru:    dr.d.st.runeR.ru,
		fg:    dr.d.st.curColors.fg,
		fface: dr.d.st.runeR.fface,
	}
}

func (dr *DrawRune) draw2(fface font.Face, pen image.Point, ru rune, fg color.Color) {
	// skip draw
	if ru < 0 {
		return
	}

	//fmt.Printf("draw at %v \"%c\"\n", pen, ru)

	gr, mask, maskp, _, ok := fface.Glyph(fixed.Point26_6{Y: fface.Metrics().Ascent}, ru)
	if !ok {
		return
	}

	// clip
	b := dr.d.Bounds()
	gr = gr.Add(pen)
	if gr.Min.X < b.Min.X {
		maskp.X += b.Min.X - gr.Min.X
	}
	if gr.Min.Y < b.Min.Y {
		maskp.Y += b.Min.Y - gr.Min.Y
	}
	gr = gr.Intersect(b)

	draw.DrawMask(dr.d.st.drawR.img, gr, image.NewUniform(fg), image.Point{}, mask, maskp, draw.Over)
}

type DrawRuneDelay struct {
	pen   image.Point
	ru    rune
	fg    color.Color
	fface font.Face
}
