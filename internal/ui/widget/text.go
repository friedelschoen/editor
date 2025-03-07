package widget

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/friedelschoen/editor/internal/drawer"
	"github.com/friedelschoen/editor/internal/ioutil"
)

type Text struct {
	ENode
	TextScroll

	Drawer *drawer.TextDrawer

	scrollable struct{ x, y bool }
	ctx        ImageContext
	bg         color.Color

	rw ioutil.ReadWriterAt
}

func NewText(ctx ImageContext) *Text {
	t := &Text{ctx: ctx}

	t.Drawer = drawer.New()

	t.TextScroll.Text = t
	t.TextScroll.Drawer = t.Drawer

	rw := ioutil.NewBytesReadWriterAt(nil)
	t.SetRW(rw)

	return t
}

func (t *Text) RW() ioutil.ReadWriterAt {
	return t.rw
}

func (t *Text) SetRW(rw ioutil.ReadWriterAt) {
	t.rw = rw
	t.Drawer.SetReader(rw)
}

func (t *Text) Len() int {
	return t.rw.Max() - t.rw.Min()
}

// Result might not be a copy, so changes to the slice might affect the text data.
func (t *Text) Bytes() ([]byte, error) {
	return ioutil.ReadFastFull(t.rw)
}

func (t *Text) SetBytes(b []byte) error {
	if err := ioutil.SetBytes(t.rw, b); err != nil {
		return err
	}
	t.contentChanged()
	return nil
}

func (t *Text) Str() string {
	p, err := t.Bytes()
	if err != nil {
		return ""
	}
	return string(p)
}

func (t *Text) SetStr(str string) error {
	return t.SetBytes([]byte(str))
}

func (t *Text) contentChanged() {
	t.Drawer.ContentChanged()

	// content changing can influence the layout in the case of dynamic sized textareas (needs layout). Also in the case of scrollareas that need to recalc scrollbars.
	t.MarkNeedsLayoutAndPaint()
}

// implements Scrollable interface.
func (t *Text) SetScrollable(x, y bool) {
	t.scrollable.x = x
	t.scrollable.y = y
}

func (t *Text) RuneOffset() int {
	return t.Drawer.RuneOffset()
}

func (t *Text) SetRuneOffset(v int) {
	if t.scrollable.y && t.Drawer.RuneOffset() != v {
		t.Drawer.SetRuneOffset(v)
		t.MarkNeedsLayoutAndPaint()
	}
}

func (t *Text) IndexVisible(offset int) bool {
	return t.Drawer.RangeVisible(offset, 0)
}
func (t *Text) MakeIndexVisible(offset int) {
	t.MakeRangeVisible(offset, 0)
}
func (t *Text) MakeRangeVisible(offset, n int) {
	t.MakeRangeVisible2(offset, n, drawer.RAlignAuto)
}
func (t *Text) MakeRangeVisible2(offset, n int, align drawer.RangeAlignment) {
	o := t.Drawer.RangeVisibleOffset(offset, n, align)
	t.SetRuneOffset(o)
}

func (t *Text) GetPoint(i int) image.Point {
	return t.Drawer.LocalPointOf(i)
}
func (t *Text) GetIndex(p image.Point) int {
	return t.Drawer.LocalIndexOf(p)
}

func (t *Text) LineHeight() int {
	return t.Drawer.LineHeight()
}

func (t *Text) Measure(hint image.Point) image.Point {
	b := t.Bounds
	b.Max = b.Min.Add(hint)
	t.Drawer.SetBounds(b)
	m := t.Drawer.Measure()
	return drawer.MinPoint(m, hint)
}

func (t *Text) Layout() {
	if t.Bounds != t.Drawer.Bounds() {
		t.Drawer.SetBounds(t.Bounds)
		t.MarkNeedsPaint()
	}
}

func (t *Text) PaintBase() {
	draw.Draw(t.ctx.Image(), t.Bounds, image.NewUniform(t.bg), image.Point{}, draw.Src)
}
func (t *Text) Paint() {
	t.Drawer.Draw(t.ctx.Image())
}

func (t *Text) OnThemeChange() {
	// word highlight ops (contain fg/bg colors) are cached. A contentchanged() here is the easiest way to invalidate the cache and have all the colors be updated.
	t.Drawer.ContentChanged()

	fg := t.TreeThemePaletteColor("text_fg")
	t.Drawer.SetFg(fg)

	t.bg = t.TreeThemePaletteColor("text_bg")

	ff := t.TreeThemeFontFace()
	if ff != t.Drawer.FontFace() {
		t.Drawer.SetFontFace(ff)
		t.MarkNeedsLayoutAndPaint()
	} else {
		t.MarkNeedsPaint()
	}
}
