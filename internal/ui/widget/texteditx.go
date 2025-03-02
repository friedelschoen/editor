package widget

import (
	"image/color"
	"time"

	"github.com/friedelschoen/glake/internal/drawer"
	"github.com/friedelschoen/glake/internal/ioutil"
	"github.com/friedelschoen/glake/internal/shadow"
)

// textedit with extensions
type TextEditX struct {
	*TextEdit

	flash struct {
		start time.Time
		now   time.Time
		dur   time.Duration
		line  struct {
			on bool
		}
		index struct {
			on    bool
			index int
			len   int
		}
	}
}

func NewTextEditX(uiCtx UIContext) *TextEditX {
	te := &TextEditX{
		TextEdit: NewTextEdit(uiCtx),
	}

	te.Text.Drawer.Opt.Cursor.On = true

	// setup colorize order
	te.Text.Drawer.Opt.Colorize.Groups = []*drawer.ColorizeGroup{
		&te.Text.Drawer.Opt.SyntaxHighlight.Group,
		&te.Text.Drawer.Opt.WordHighlight.Group,
		&te.Text.Drawer.Opt.ParenthesisHighlight.Group,
		{}, // 3=terminal
		{}, // 4=selection
		{}, // 5=flash
	}

	return te
}

func (te *TextEditX) PaintBase() {
	te.TextEdit.PaintBase()
	te.iterateFlash()
}

func (te *TextEditX) Paint() {
	te.updateSelectionOpt()
	te.updateFlashOpt()
	te.TextEdit.Paint()
}

func (te *TextEditX) updateSelectionOpt() {
	g := te.Drawer.Opt.Colorize.Groups[4]
	c := te.Cursor()
	if s, e, ok := c.SelectionIndexes(); ok {
		// colors
		pcol := te.TreeThemePaletteColor
		fg := pcol("text_selection_fg")
		bg := pcol("text_selection_bg")
		// colorize ops
		g.Ops = []*drawer.ColorizeOp{
			{Offset: s, Fg: fg, Bg: bg},
			{Offset: e},
		}
		// don't draw other colorizations
		te.Drawer.Opt.WordHighlight.Group.Off = true
		te.Drawer.Opt.ParenthesisHighlight.Group.Off = true
	} else {
		g.Ops = nil
		// draw other colorizations
		te.Drawer.Opt.WordHighlight.Group.Off = false
		te.Drawer.Opt.ParenthesisHighlight.Group.Off = false
	}
}

func (te *TextEditX) FlashLine(index int) {
	te.startFlash(index, 0, true)
}

func (te *TextEditX) FlashIndexLen(index int, len int) {
	te.startFlash(index, len, len == 0)
}

// Safe to use concurrently. If line is true then len is calculated.
func (te *TextEditX) startFlash(index, len int, line bool) {
	te.uiCtx.RunOnUIGoRoutine(func() {
		te.flash.start = time.Now()
		te.flash.dur = 500 * time.Millisecond

		if line {
			// recalc index/len
			i0, i1 := te.flashLineIndexes(index)
			index = i0
			len = i1 - index

			te.flash.line.on = true
			// need at least len 1 or the colorize op will be canceled
			if len == 0 {
				len = 1
			}
		}

		// flash index (accurate runes)
		te.flash.index.on = true
		te.flash.index.index = index
		te.flash.index.len = len

		te.MarkNeedsPaint()
	})
}

func (te *TextEditX) flashLineIndexes(offset int) (int, int) {
	rd := te.EditCtx().LocalReader(offset)
	s, e, newline, err := ioutil.LinesIndexes(rd, offset, offset)
	if err != nil {
		return 0, 0
	}
	if newline {
		e--
	}
	return s, e
}

func (te *TextEditX) iterateFlash() {
	if !te.flash.line.on && !te.flash.index.on {
		return
	}

	te.flash.now = time.Now()
	end := te.flash.start.Add(te.flash.dur)

	// animation time ended
	if te.flash.now.After(end) {
		te.flash.index.on = false
		te.flash.line.on = false
	} else {
		te.uiCtx.RunOnUIGoRoutine(func() {
			te.MarkNeedsPaint()
		})
	}
}

func (te *TextEditX) updateFlashOpt() {
	te.updateFlashOpt4(te.Drawer)
}

func (te *TextEditX) updateFlashOpt4(d *drawer.TextDrawer) {
	g := d.Opt.Colorize.Groups[5]
	if !te.flash.index.on {
		g.Ops = nil
		return
	}

	// tint percentage
	t := te.flash.now.Sub(te.flash.start)
	perc := 1.0 - (float64(t) / float64(te.flash.dur))

	// process color function
	bg3 := te.TreeThemePaletteColor("text_bg")
	pc := func(fg, bg color.Color) (_, _ color.Color) {
		fg2 := shadow.TintOrShade(fg, perc)
		if bg == nil {
			bg = bg3
		}
		bg2 := shadow.TintOrShade(bg, perc)
		return fg2, bg2
	}

	s := te.flash.index.index
	e := s + te.flash.index.len
	line := te.flash.line.on
	g.Ops = []*drawer.ColorizeOp{
		{Offset: s, ProcColor: pc, Line: line},
		{Offset: e},
	}
}

func (te *TextEditX) EnableParenthesisMatch(v bool) {
	te.Drawer.Opt.ParenthesisHighlight.On = v
}

func (te *TextEditX) EnableSyntaxHighlight(v bool) {
	te.Drawer.Opt.SyntaxHighlight.On = v
}

func (te *TextEditX) EnableCursorWordHighlight(v bool) {
	te.Drawer.Opt.WordHighlight.On = v
}

func (te *TextEditX) OnThemeChange() {
	te.Text.OnThemeChange()

	pcol := te.TreeThemePaletteColor

	te.Drawer.Opt.Cursor.Fg = pcol("text_cursor_fg")
	te.Drawer.Opt.LineWrap.Fg = pcol("text_wrapline_fg")
	te.Drawer.Opt.LineWrap.Bg = pcol("text_wrapline_bg")

	// annotations
	te.Drawer.Opt.Annotations.Fg = pcol("text_annotations_fg")
	te.Drawer.Opt.Annotations.Bg = pcol("text_annotations_bg")
	te.Drawer.Opt.Annotations.Selected.Fg = pcol("text_annotations_select_fg")
	te.Drawer.Opt.Annotations.Selected.Bg = pcol("text_annotations_select_bg")

	// word highlight
	te.Drawer.Opt.WordHighlight.Fg = pcol("text_highlightword_fg")
	te.Drawer.Opt.WordHighlight.Bg = pcol("text_highlightword_bg")

	// parenthesis highlight
	te.Drawer.Opt.ParenthesisHighlight.Fg = pcol("text_parenthesis_fg")
	te.Drawer.Opt.ParenthesisHighlight.Bg = pcol("text_parenthesis_bg")
}
