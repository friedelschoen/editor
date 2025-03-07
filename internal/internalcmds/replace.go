package internalcmds

import (
	"fmt"

	"github.com/friedelschoen/editor/internal/core"
	"github.com/friedelschoen/editor/internal/editbuf"
)

func Replace(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}

	args2 := args.Part.Args[1:]
	if len(args2) != 2 {
		return fmt.Errorf("expecting 2 arguments")
	}

	old, new := args2[0].UnquotedString(), args2[1].UnquotedString()

	ta := erow.Row.TextArea
	ta.BeginUndoGroup()
	defer ta.EndUndoGroup()
	replaced, err := editbuf.Replace(ta.EditCtx(), old, new)
	if err != nil {
		return err
	}
	if !replaced {
		return fmt.Errorf("string not replaced: %q", old)
	}
	return nil
}
