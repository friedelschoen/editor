package editbuf

import "github.com/friedelschoen/glake/internal/ioutil"

func Backspace(ctx *EditorBuffer) error {
	a, b, ok := ctx.C.SelectionIndexes()
	if ok {
		ctx.C.SetSelectionOff()
	} else {
		b = ctx.C.Index()
		_, size, err := ioutil.ReadLastRuneAt(ctx.RW, b)
		if err != nil {
			return err
		}
		a = b - size
	}
	if err := ctx.RW.OverwriteAt(a, b-a, nil); err != nil {
		return err
	}
	ctx.C.SetIndex(a)
	return nil
}
