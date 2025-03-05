package editbuf

import "github.com/friedelschoen/glake/internal/ui/driver"

func SelectLine(ctx *EditorBuffer) error {
	ctx.C.SetSelectionOff()
	a, b, _, err := ctx.CursorSelectionLinesIndexes()
	if err != nil {
		return err
	}
	ctx.C.SetSelection(a, b)
	// set primary copy
	if b, ok := ctx.Selection(); ok {
		driver.SetClipboardData(string(b))
	}
	return nil
}
