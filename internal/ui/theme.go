package ui

import (
	"fmt"
	"image"
	"image/color"

	"github.com/friedelschoen/glake/internal/findfont"
	"github.com/friedelschoen/glake/internal/shadow"
	"github.com/friedelschoen/glake/internal/ui/widget"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var ScrollBarLeft = true
var ScrollBarWidth int = 0 // 0=based on a portion of the font size

const separatorWidth = 1 // col/row separators width

func lightThemeColors(node widget.Node) {
	pal := lightThemeColorsPal()
	pal.Merge(rowSquarePalette())
	node.Embed().SetThemePalette(pal)
}
func lightThemeColorsPal() widget.Palette {
	pal := widget.Palette{
		"text_cursor_fg":            cint(0x0),
		"text_fg":                   cint(0x0),
		"text_bg":                   cint(0xffffff),
		"text_selection_fg":         nil,
		"text_selection_bg":         cint(0xeeee9e), // yellow
		"text_colorize_string_fg":   cint(0x8b0000), // red
		"text_colorize_comments_fg": cint(0x008b00), // green
		"text_highlightword_fg":     nil,
		"text_highlightword_bg":     cint(0xc6ee9e), // green
		"text_wrapline_fg":          cint(0x0),
		"text_wrapline_bg":          cint(0xd8d8d8),
		"text_parenthesis_fg":       nil,
		"text_parenthesis_bg":       cint(0xd8d8d8),

		"toolbar_text_bg":          cint(0xecf0f1), // "clouds" grey
		"toolbar_text_wrapline_bg": cint(0xccccd8),

		"scrollbar_bg":        cint(0xf2f2f2),
		"scrollhandle_normal": shadow.Shade(cint(0xf2f2f2), 0.20),
		"scrollhandle_hover":  shadow.Shade(cint(0xf2f2f2), 0.30),
		"scrollhandle_select": shadow.Shade(cint(0xf2f2f2), 0.40),

		"column_norows_rect":  cint(0xffffff),
		"columns_nocols_rect": cint(0xffffff),
		"colseparator_rect":   cint(0x0),
		"rowseparator_rect":   cint(0x0),
		"shadow.sep_rect":     cint(0x0),

		"columnsquare": cint(0xccccd8),
		"rowsquare":    cint(0xccccd8),

		"mm_text_bg":          cint(0xecf0f1),
		"mm_button_hover_bg":  cint(0xcccccc),
		"mm_button_down_bg":   cint(0xbbbbbb),
		"mm_button_sticky_fg": cint(0xffffff),
		"mm_button_sticky_bg": cint(0x0),
		"mm_border":           cint(0x0),
		"mm_content_pad":      cint(0xecf0f1),
		"mm_content_border":   cint(0x0),

		"contextfloatbox_border": cint(0x0),
	}
	pal.Merge(rowSquarePalette())
	return pal
}

func acmeThemeColors(node widget.Node) {
	pal := acmeThemeColorsPal()
	pal.Merge(rowSquarePalette())
	node.Embed().SetThemePalette(pal)
}
func acmeThemeColorsPal() widget.Palette {
	pal := widget.Palette{
		"text_cursor_fg":            cint(0x0),
		"text_fg":                   cint(0x0),
		"text_bg":                   cint(0xffffea),
		"text_selection_fg":         nil,
		"text_selection_bg":         cint(0xeeee9e), // yellow
		"text_colorize_string_fg":   cint(0x8b0000), // red
		"text_colorize_comments_fg": cint(0x007500), // green
		"text_highlightword_fg":     nil,
		"text_highlightword_bg":     cint(0xc6ee9e), // green
		"text_wrapline_fg":          cint(0x0),
		"text_wrapline_bg":          cint(0xd8d8c6),

		"toolbar_text_bg":          cint(0xeaffff),
		"toolbar_text_wrapline_bg": cint(0xc6d8d8),

		"scrollbar_bg":        cint(0xf2f2de),
		"scrollhandle_normal": cint(0xc1c193),
		"scrollhandle_hover":  cint(0xadad6f),
		"scrollhandle_select": cint(0x99994c),

		"column_norows_rect":  cint(0xffffea),
		"columns_nocols_rect": cint(0xffffff),
		"colseparator_rect":   cint(0x0),
		"rowseparator_rect":   cint(0x0),
		"shadow.sep_rect":     cint(0x0),

		"columnsquare": cint(0xc6d8d8),
		"rowsquare":    cint(0xc6d8d8),

		"mm_text_bg":          cint(0xeaffff),
		"mm_button_hover_bg":  shadow.Shade(cint(0xeaffff), 0.10),
		"mm_button_down_bg":   shadow.Shade(cint(0xeaffff), 0.20),
		"mm_button_sticky_bg": shadow.Shade(cint(0xeaffff), 0.40),
		"mm_border":           cint(0x0),
		"mm_content_pad":      cint(0xeaffff),
		"mm_content_border":   cint(0x0),

		"contextfloatbox_border": cint(0x0),
	}
	pal.Merge(rowSquarePalette())
	return pal
}

func rowSquarePalette() widget.Palette {
	pal := widget.Palette{
		"rs_active":              cint(0x0),
		"rs_executing":           cint(0x0fad00),                    // dark green
		"rs_edited":              cint(0x0000ff),                    // blue
		"rs_disk_changes":        cint(0xff0000),                    // red
		"rs_not_exist":           cint(0xff9900),                    // orange
		"rs_duplicate":           cint(0x8888cc),                    // blueish
		"rs_duplicate_highlight": cint(0xffff00),                    // yellow
		"rs_annotations":         cint(0xd35400),                    // pumpkin
		"rs_annotations_edited":  shadow.Tint(cint(0xd35400), 0.45), // pumpkin (brighter)
	}
	return pal
}

var ColorThemeCycler cycler = cycler{
	entries: []cycleEntry{
		{"light", lightThemeColors},
		{"acme", acmeThemeColors},
	},
}

var CurrentFont = ""

func loadThemeFont(name string, node widget.Node) error {
	ff, err := ThemeFontFace(name, 0)
	if err != nil {
		return err
	}
	node.Embed().SetThemeFontFace(ff)
	return nil
}

var TTFontOptions opentype.FaceOptions

func ThemeFontFace(name string, size float64) (font.Face, error) {
	b, err := findfont.GetFontData(name)
	if err != nil {
		b = defaultFont
	}
	font, err := opentype.Parse(b)
	if err != nil {
		return nil, err
	}
	opt := TTFontOptions
	if size != 0 {
		opt.Size = size
	}
	return opentype.NewFace(font, &opt)
}

var defaultFont = gomono.TTF

type cycler struct {
	CurName string
	entries []cycleEntry
}

func (c *cycler) GetIndex(name string) (int, bool) {
	for i, e := range c.entries {
		if e.name == name {
			return i, true
		}
	}
	return -1, false
}

func (c *cycler) Cycle(node widget.Node) {
	i := 0
	if c.CurName != "" {
		k, ok := c.GetIndex(c.CurName)
		if !ok {
			panic(fmt.Sprintf("cycle name not found: %v", c.CurName))
		}
		i = (k + 1) % len(c.entries)
	}
	c.Set(c.entries[i].name, node)
}

func (c *cycler) Set(name string, node widget.Node) {
	i, ok := c.GetIndex(name)
	if !ok {
		panic(fmt.Sprintf("cycle name not found: %v", name))
	}
	c.CurName = name
	c.entries[i].fn(node)
}

func (c *cycler) Names() []string {
	w := []string{}
	for _, e := range c.entries {
		w = append(w, e.name)
	}
	return w
}

type cycleEntry struct {
	name string
	fn   func(widget.Node)
}

var UIThemeUtil uiThemeUtil

type uiThemeUtil struct{}

func (uitu *uiThemeUtil) RowMinimumHeight(ff font.Face) int {
	return ff.Metrics().Height.Ceil()
}

func (uitu *uiThemeUtil) RowSquareSize(ff font.Face) image.Point {
	lh := ff.Metrics().Height
	w := lh.Mul(fixed.Int26_6(64 * 3 / 4)) // 3/4
	return image.Point{w.Ceil(), lh.Ceil()}
}

func (uitu *uiThemeUtil) GetScrollBarWidth(ff font.Face) int {
	if ScrollBarWidth != 0 {
		return ScrollBarWidth
	}
	lh := ff.Metrics().Height
	w := lh.Mul(fixed.Int26_6(64 * 3 / 4)) // 3/4
	return w.Ceil()
}

func (uitu *uiThemeUtil) ShadowHeight(ff font.Face) int {
	lh := ff.Metrics().Height
	w := lh.Mul(fixed.Int26_6(64 * 2 / 5)) // 2/5
	return w.Ceil()
}

func cint(c int) color.RGBA {
	v := c & 0xffffff
	r := uint8(v >> 16)
	g := uint8(v >> 8)
	b := uint8(v >> 0)
	return color.RGBA{r, g, b, 255}
}
