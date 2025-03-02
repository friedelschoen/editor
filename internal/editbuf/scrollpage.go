package editbuf

func ScrollUp(ctx *EditorBuffer, up bool) {
	ctx.Fns.ScrollUp(up)
}

func PageUp(ctx *EditorBuffer, up bool) {
	ctx.Fns.PageUp(up)
}
