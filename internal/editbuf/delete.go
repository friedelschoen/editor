package editbuf

import "github.com/friedelschoen/editor/internal/ioutil"

func Delete(ctx *EditorBuffer) error {
	a, b, ok := ctx.C.SelectionIndexes()
	if ok {
		ctx.C.SetSelectionOff()
	} else {
		a = ctx.C.Index()
		_, size, err := ioutil.ReadRuneAt(ctx.RW, a)
		if err != nil {
			return err
		}
		b = a + size
	}
	if err := ctx.RW.OverwriteAt(a, b-a, nil); err != nil {
		return err
	}
	ctx.C.SetIndex(a)
	return nil
}
