package drawer

import (
	"image/color"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func updateSyntaxHighlightOps(d *TextDrawer) {
	if shDone(d) {
		return
	}
	d.Opt.SyntaxHighlight.Group.Ops = SyntaxHighlight(d)
}

func shDone(d *TextDrawer) bool {
	if !d.Opt.SyntaxHighlight.On {
		d.Opt.SyntaxHighlight.Group.Ops = nil
		return true
	}
	if d.opt.syntaxH.updated {
		return true
	}
	d.opt.syntaxH.updated = true
	return false
}

func HexColor(in chroma.Colour) color.Color {
	if !in.IsSet() {
		return nil
	}
	return color.RGBA{
		R: in.Red(),
		G: in.Green(),
		B: in.Blue(),
		A: 255,
	}
}

func SyntaxHighlight(d *TextDrawer) []*ColorizeOp {
	// limit reading to be able to handle big content
	o, n, _, _ := d.visibleLen()

	/* read up, until visible end */
	bytes, err := d.reader.ReadFastAt(0, o+n)
	if err != nil {
		return nil
	}
	content := string(bytes)

	lexer := lexers.Get("go")
	if lexer == nil {
		lexer = lexers.Fallback
	}
	// fmt.Printf("lexer: %s\n", lexer.Config().Name)

	style := styles.Get("xcode")
	if style == nil {
		style = styles.Fallback
	}

	tokens, _ := lexer.Tokenise(nil, content)

	ops := make([]*ColorizeOp, 0)

	off := 0
	for {
		token := tokens()
		if token == chroma.EOF {
			break
		}
		s := style.Get(token.Type)
		ops = append(ops, &ColorizeOp{
			Offset: off,
			Fg:     HexColor(s.Colour),
		})
		off += len(token.Value)
	}

	return ops
}
