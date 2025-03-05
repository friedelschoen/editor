package editbuf

import "github.com/friedelschoen/glake/internal/ui/driver"

func Cut(ctx *EditorBuffer) error {
	a, b, ok := ctx.C.SelectionIndexes()
	if !ok {
		return nil
	}

	s, err := ctx.RW.ReadFastAt(a, b-a)
	if err != nil {
		return err
	}
	driver.SetClipboardData(string(s))

	if err := ctx.RW.OverwriteAt(a, b-a, nil); err != nil {
		return err
	}
	ctx.C.SetSelectionOff()
	ctx.C.SetIndex(a)
	return nil
}
