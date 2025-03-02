package drawer

import (
	"github.com/friedelschoen/glake/internal/ioutil"
)

func updateWordHighlightWord(d *TextDrawer) {
	if !d.Opt.WordHighlight.On {
		return
	}
	if !d.Opt.Cursor.On {
		return
	}

	if d.opt.wordH.updatedWord {
		return
	}
	d.opt.wordH.updatedWord = true

	// find word
	d.opt.wordH.word = nil
	ci := d.opt.cursor.offset
	rd := ioutil.NewLimitedReaderAtPad(d.reader, ci, ci, 250)
	word, _, err := ioutil.WordAtIndex(rd, ci)
	if err != nil {
		return
	}
	d.opt.wordH.word = word
}

func updateWordHighlightOps(d *TextDrawer) {
	if !d.Opt.WordHighlight.On {
		d.Opt.WordHighlight.Group.Ops = nil
		return
	}

	if d.opt.wordH.updatedOps {
		return
	}
	d.opt.wordH.updatedOps = true

	d.Opt.WordHighlight.Group.Ops = wordHOps(d)
}

func wordHOps(d *TextDrawer) []*ColorizeOp {
	word := d.opt.wordH.word
	if word == nil {
		return nil
	}

	// offsets to search
	o, n, _, _ := d.visibleLen()
	a, b := o, o+n
	a -= len(word)
	b += len(word)

	// limits
	a0 := d.reader.Min()
	if a < a0 {
		a = a0
	}
	b0 := d.reader.Max()
	if b > b0 {
		b = b0
	}

	// search
	var ops []*ColorizeOp
	for i := a; i < b; {
		// find word
		rd := ioutil.NewLimitedReaderAt(d.reader, i, b)
		j, _, err := ioutil.Index(rd, i, word, false)
		if err != nil {
			return nil
		}
		if j < 0 {
			break
		}

		// isolated word
		if ioutil.WordIsolated(d.reader, j, len(word)) {
			op1 := &ColorizeOp{
				Offset: j,
				Fg:     d.Opt.WordHighlight.Fg,
				Bg:     d.Opt.WordHighlight.Bg,
			}
			op2 := &ColorizeOp{Offset: j + len(word)}
			ops = append(ops, op1, op2)
		}

		i = j + len(word)
	}
	return ops
}
