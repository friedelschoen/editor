package editbuf

import (
	"bytes"
	"fmt"
	"image"

	"github.com/friedelschoen/glake/internal/ioutil"
)

//godebug:annotatefile

type EditorBuffer struct {
	RW  ioutil.ReadWriterAt
	C   Cursor
	Fns CtxFns
}

func NewEditorBuffer() *EditorBuffer {
	ctx := &EditorBuffer{C: &SimpleCursor{}, Fns: EmptyCtxFns()}
	return ctx
}

func (ctx *EditorBuffer) CursorSelectionLinesIndexes() (int, int, bool, error) {
	a, b, ok := ctx.C.SelectionIndexes()
	if !ok {
		a = ctx.C.Index()
		b = a
	}
	rd := ctx.LocalReader2(a, b)
	return ioutil.LinesIndexes(rd, a, b)
}

func (ctx *EditorBuffer) Selection() ([]byte, bool) {
	a, b, ok := ctx.C.SelectionIndexes()
	if !ok {
		return nil, false
	}
	w, err := ctx.RW.ReadFastAt(a, b-a)
	if err != nil {
		return nil, false
	}
	return bytes.Clone(w), true
}

func (ctx *EditorBuffer) LocalReader(i int) ioutil.ReaderAt {
	return ctx.LocalReader2(i, i)
}
func (ctx *EditorBuffer) LocalReader2(min, max int) ioutil.ReaderAt {
	pad := 2500
	return ioutil.NewLimitedReaderAtPad(ctx.RW, min, max, pad)
}

type CtxFns struct {
	Error func(error)

	GetPoint         func(int) image.Point
	GetIndex         func(image.Point) int
	LineHeight       func() int
	CommentLineSym   func() any
	MakeIndexVisible func(int)
	PageUp           func(up bool)
	ScrollUp         func(up bool)

	SetClipboardData func(string)
	GetClipboardData func() (string, error)

	Undo func() error
	Redo func() error
}

func EmptyCtxFns() CtxFns {
	u := CtxFns{}

	u.Error = func(err error) { fmt.Println(err) }

	u.GetPoint = func(int) image.Point { return image.Point{} }
	u.GetIndex = func(image.Point) int { return 0 }
	u.LineHeight = func() int { return 0 }
	u.CommentLineSym = func() any { return nil }
	u.MakeIndexVisible = func(int) {}
	u.PageUp = func(bool) {}
	u.ScrollUp = func(bool) {}

	u.SetClipboardData = func(string) {}
	u.GetClipboardData = func() (string, error) {
		return "", nil
	}

	u.Undo = func() error { return nil }
	u.Redo = func() error { return nil }

	return u
}
