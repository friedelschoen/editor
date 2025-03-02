package editbuf

func SelectAll(ctx *EditorBuffer) error {
	ctx.C.SetSelection(ctx.RW.Min(), ctx.RW.Max())
	return nil
}
