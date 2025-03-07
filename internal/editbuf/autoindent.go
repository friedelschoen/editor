package editbuf

import (
	"errors"
	"io"
	"unicode"

	"github.com/friedelschoen/editor/internal/ioutil"
)

func AutoIndent(ctx *EditorBuffer) error {
	ci := ctx.C.Index()

	rd1 := ioutil.NewLimitedReaderAt(ctx.RW, ci-2000, ci)
	i, err := ioutil.LineStartIndex(rd1, ci)
	if err != nil {
		return err
	}

	rd := ioutil.NewLimitedReaderAt(ctx.RW, i, ci)
	j, _, err := ioutil.RuneIndexFn(rd, i, false, unicode.IsSpace)
	if err != nil {
		if errors.Is(err, io.EOF) {
			j = ci // all spaces up to ci
		} else {
			return err
		}
	}

	// string to insert
	s, err := ctx.RW.ReadFastAt(i, j-i)
	if err != nil {
		return err
	}
	s2 := append([]byte{'\n'}, s...)

	// selection to overwrite
	n := 0
	if a, b, ok := ctx.C.SelectionIndexes(); ok {
		n = b - a
		ci = a
		ctx.C.SetSelectionOff()
	}

	if err := ctx.RW.OverwriteAt(ci, n, s2); err != nil {
		return err
	}
	ctx.C.SetIndex(ci + len(s2))
	return nil
}
