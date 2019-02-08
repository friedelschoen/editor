package drawer4

import (
	"github.com/jmigpin/editor/util/iout"
)

func updateWordHighlightWord(d *Drawer) {
	d.Opt.WordHighlight.word = nil

	if !d.Opt.WordHighlight.On {
		return
	}
	if !d.Opt.Cursor.On {
		return
	}
	if !d.Opt.RuneOffset.On {
		return
	}

	// find word
	ci := d.Opt.Cursor.index
	n := d.runeOffsetViewLen()
	word, _, err := iout.WordAtIndex(d.reader, ci, n)
	if err != nil {
		return
	}
	d.Opt.WordHighlight.word = word
}

//----------

func updateWordHighlightOps(d *Drawer) {
	opt := &d.Opt.WordHighlight
	opt.Group.Ops = WordHighlightOps(d)
}

func WordHighlightOps(d *Drawer) []*ColorizeOp {
	word := d.Opt.WordHighlight.word
	if word == nil {
		return nil
	}

	// offsets to search
	o := d.Opt.RuneOffset.offset
	n := d.runeOffsetViewLen()
	a, b := o, o+n
	a -= len(word)
	b += len(word)
	if a < 0 {
		a = 0
	}
	l := d.reader.Len()
	if b > l {
		b = l
	}

	// search
	var ops []*ColorizeOp
	for i := a; i < b; {
		// find word
		j, err := iout.Index(d.reader, i, b-i, word, false)
		if err != nil {
			return nil
		}
		if j < 0 {
			break
		}

		// isolated word
		if iout.WordIsolated(d.reader, j, len(word)) {
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
