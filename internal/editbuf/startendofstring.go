package editbuf

func StartOfString(ctx *EditorBuffer, sel bool) {
	ctx.C.UpdateSelection(sel, 0)
}

func EndOfString(ctx *EditorBuffer, sel bool) {
	ctx.C.UpdateSelection(sel, ctx.RW.Max())
}
