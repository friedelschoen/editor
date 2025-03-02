package editbuf

func Undo(ctx *EditorBuffer) error {
	return ctx.Fns.Undo()
}
func Redo(ctx *EditorBuffer) error {
	return ctx.Fns.Redo()
}
