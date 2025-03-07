package editbuf

import (
	"bytes"
	"image"

	"github.com/friedelschoen/editor/internal/ioutil"
)

//godebug:annotatefile

type EditorBuffer struct {
	RW  ioutil.ReadWriterAt
	C   Cursor
	Fns CtxFns
}

func NewEditorBuffer() *EditorBuffer {
	ctx := &EditorBuffer{C: &SimpleCursor{}, Fns: nil}
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

type CtxFns interface {
	Error(error)

	GetPoint(int) image.Point
	GetIndex(image.Point) int
	LineHeight() int
	CommentLineSym() any
	MakeIndexVisible(int)
	PageUp(up bool)
	ScrollUp(up bool)

	Undo() error
	Redo() error
}
