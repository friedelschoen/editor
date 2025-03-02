package editbuf

func SelectLine(ctx *EditorBuffer) error {
	ctx.C.SetSelectionOff()
	a, b, _, err := ctx.CursorSelectionLinesIndexes()
	if err != nil {
		return err
	}
	ctx.C.SetSelection(a, b)
	// set primary copy
	if b, ok := ctx.Selection(); ok {
		ctx.Fns.SetClipboardData(string(b))
	}
	return nil
}
