package drawutil

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/friedelschoen/glake/util/fontutil"
	"github.com/friedelschoen/glake/util/iout/iorw"
)

type Drawer interface {
	Reader() iorw.ReaderAt
	SetReader(iorw.ReaderAt)
	ContentChanged()

	FontFace() *fontutil.FontFace
	SetFontFace(*fontutil.FontFace)
	LineHeight() int
	SetFg(color.Color)

	Bounds() image.Rectangle
	SetBounds(image.Rectangle)

	// rune offset  (set text view position; save/restore view in session file)
	RuneOffset() int
	SetRuneOffset(int)

	LocalPointOf(index int) image.Point
	LocalIndexOf(image.Point) int

	Measure() image.Point
	Draw(img draw.Image)

	// specialized: covers editor row button margin
	FirstLineOffsetX() int
	SetFirstLineOffsetX(x int)

	// cursor
	SetCursorOffset(int)

	// scrollable utils
	ScrollOffset() image.Point
	SetScrollOffset(image.Point)
	ScrollSize() image.Point
	ScrollViewSize() image.Point
	ScrollPageSizeY(up bool) int
	ScrollWheelSizeY(up bool) int

	// visibility utils
	RangeVisible(offset, n int) bool
	RangeVisibleOffset(offset, n int, align RangeAlignment) int
}

type SyntaxHighlightComment struct {
	S, E   string // {start,end} sequence
	IsLine bool   // single line comment (end argument is ignored)
}

type RangeAlignment int

const (
	RAlignKeep         RangeAlignment = iota
	RAlignKeepOrBottom                // keep if visible, or bottom
	RAlignAuto
	RAlignTop
	RAlignBottom
	RAlignCenter
)
