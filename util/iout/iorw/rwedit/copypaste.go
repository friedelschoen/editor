package rwedit

import (
	"fmt"
)

func Copy(ctx *Ctx) error {
	if b, ok := ctx.Selection(); ok {
		ctx.Fns.SetClipboardData(string(b))
	}
	return nil
}

func Paste(ctx *Ctx) {
	s, err := ctx.Fns.GetClipboardData()
	if err != nil {
		ctx.Fns.Error(fmt.Errorf("rwedit.paste: %w", err))
		return
	}
	if err := InsertString(ctx, s); err != nil {
		ctx.Fns.Error(fmt.Errorf("rwedit.paste: insertstring: %w", err))
	}
}
