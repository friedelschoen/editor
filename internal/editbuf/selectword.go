package editbuf

import (
	"github.com/friedelschoen/editor/internal/ioutil"
	"github.com/friedelschoen/editor/internal/ui/driver"
)

func SelectWord(ctx *EditorBuffer) error {
	// index rune
	ci := ctx.C.Index()
	ru, _, err := ioutil.ReadRuneAt(ctx.RW, ci)
	if err != nil {
		return err
	}

	var index int
	var word []byte
	if !ioutil.IsWordRune(ru) {
		// select just the index rune
		index = ci
		word = []byte(string(ru))
	} else {
		// select word at index
		rd := ctx.LocalReader(ci)
		w, i, err := ioutil.WordAtIndex(rd, ci)
		if err != nil {
			return err
		}

		index = i
		word = w
	}

	ctx.C.SetSelection(index, index+len(word))

	// set primary copy
	if b, ok := ctx.Selection(); ok {
		driver.SetClipboardData(string(b))
	}

	return nil
}
