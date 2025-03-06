package ui

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/friedelschoen/glake/internal/findfont"
	"github.com/friedelschoen/glake/internal/ui/widget"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var ScrollBarLeft = true
var ScrollBarWidth int = 0 // 0=based on a portion of the font size

const separatorWidth = 1 // col/row separators width

const themeExtension = ".glaketheme"

// Predefined color names mapped to hex values
var colors = map[string]string{
	"red":      "#FF0000",
	"green":    "#008000",
	"blue":     "#0000FF",
	"black":    "#000000",
	"white":    "#FFFFFF",
	"yellow":   "#FFFF00",
	"orange":   "#FFA500",
	"cyan":     "#00FFFF",
	"magenta":  "#FF00FF",
	"gray":     "#808080",
	"darkgray": "#A9A9A9",
}

// parseColor converts a color name or hex string to an RGBA color.
func parseColor(text string) (color.Color, error) {
	if len(text) == 0 {
		return nil, nil
	}

	// Lookup named colors
	if hex, found := colors[text]; found {
		text = hex
	}

	// Validate hex format
	if !strings.HasPrefix(text, "#") {
		return nil, fmt.Errorf("invalid color format: expected '#' prefix or color name in `%s`", text)
	}

	text = text[1:] // Remove #

	// Expand 3-digit and 4-digit hex codes
	if len(text) == 3 || len(text) == 4 {
		expanded := strings.Builder{}
		for _, char := range text {
			expanded.WriteByte(byte(char))
			expanded.WriteByte(byte(char))
		}
		text = expanded.String()
	}

	// Validate hex string
	if !regexp.MustCompile(`^[0-9A-Fa-f]+$`).MatchString(text) {
		return nil, fmt.Errorf("invalid color format: non-hex characters found")
	}

	switch len(text) {
	case 6: // #RRGGBB
		c, err := strconv.ParseUint(text, 16, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid color value: %w", err)
		}
		return color.RGBA{uint8(c >> 16 & 0xFF), uint8(c >> 8 & 0xFF), uint8(c & 0xFF), 255}, nil

	case 8: // #RRGGBBAA
		c, err := strconv.ParseUint(text, 16, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid color value: %w", err)
		}
		return color.RGBA{uint8(c >> 24 & 0xFF), uint8(c >> 16 & 0xFF), uint8(c >> 8 & 0xFF), uint8(c & 0xFF)}, nil
	}

	return nil, fmt.Errorf("invalid color format: incorrect length")
}

// ParsePalette reads a color palette from an INI-like file.
func parsePalette(filename string) (map[string]color.Color, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open palette file: %w", err)
	}
	defer file.Close()

	palette := make(map[string]color.Color)
	scanner := bufio.NewScanner(file)
	var section string
	lineNum := 0

	for scanner.Scan() {
		lineNum++

		text, _, _ := strings.Cut(scanner.Text(), ";")
		text = strings.TrimSpace(text)

		// Ignore comments and empty lines
		if text == "" {
			continue
		}

		// Handle section headers [section]
		if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") {
			section = text[1:len(text)-1] + "_"
			continue
		}

		// Parse key-value pairs
		key, value, found := strings.Cut(text, "=")
		if !found {
			return nil, fmt.Errorf("malformed line %d: missing '='", lineNum)
		}

		key = section + strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		// Convert to color
		col, err := parseColor(value)
		if err != nil {
			return nil, fmt.Errorf("malformed line %d (key: %s): %w", lineNum, key, err)
		}

		palette[key] = col
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return palette, nil
}

func exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil /* || !errors.Is(err, os.ErrNotExist) */
}

func filepathTheme(name string) string {
	if exist(name) {
		return name
	}
	if exist(name + themeExtension) {
		return name + themeExtension
	}
	cdir, err := os.UserConfigDir()
	if err == nil {
		cname := path.Join(cdir, "glake", "themes", name)
		if exist(cname) {
			return cname
		}
		cname = path.Join(cdir, "glake", "themes", name+themeExtension)
		if exist(cname) {
			return cname
		}
	}
	return ""
}

func SetColorscheme(name string, node widget.Node) {
	path := filepathTheme(name)
	fmt.Printf("using %s\n", path)
	pal, err := parsePalette(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to set theme: %s\n", err)
		os.Exit(1)
		return
	}
	node.Embed().SetThemePalette(pal)
}

func settheme(name string) func(node widget.Node) {
	return func(node widget.Node) {
		SetColorscheme(name, node)
	}
}

var ColorThemeCycler cycler = cycler{
	entries: []cycleEntry{
		{"light", settheme("light")},
		{"acme", settheme("acme")},
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
