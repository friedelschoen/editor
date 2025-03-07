package editbuf

import (
	"fmt"

	"github.com/friedelschoen/editor/internal/ui/driver"
)

func Copy(ctx *EditorBuffer) error {
	if b, ok := ctx.Selection(); ok {
		driver.SetClipboardData(string(b))
	}
	return nil
}

func Paste(ctx *EditorBuffer) {
	s, err := driver.GetClipboardData()
	if err != nil {
		ctx.Fns.Error(fmt.Errorf("rwedit.paste: %w", err))
		return
	}
	if err := InsertString(ctx, s); err != nil {
		ctx.Fns.Error(fmt.Errorf("rwedit.paste: insertstring: %w", err))
	}
}
