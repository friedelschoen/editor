package editbuf

import (
	"unicode"

	"github.com/friedelschoen/glake/internal/ioutil"
)

func StartOfLine(ctx *EditorBuffer, sel bool) error {
	ci := ctx.C.Index()

	rd := ctx.LocalReader(ci)
	i, err := ioutil.LineStartIndex(rd, ci)
	if err != nil {
		return err
	}

	// stop at first non blank rune from the left
	n := ci - i
	for j := 0; j < n; j++ {
		ru, _, err := ioutil.ReadRuneAt(ctx.RW, i+j)
		if err != nil {
			return err
		}
		if !unicode.IsSpace(ru) {
			i += j
			break
		}
	}

	ctx.C.UpdateSelection(sel, i)
	return nil
}

func EndOfLine(ctx *EditorBuffer, sel bool) error {
	rd := ctx.LocalReader(ctx.C.Index())
	le, newline, err := ioutil.LineEndIndex(rd, ctx.C.Index())
	if err != nil {
		return err
	}
	if newline {
		le--
	}
	ctx.C.UpdateSelection(sel, le)
	return nil
}
