package internalcmds

import (
	"fmt"

	"github.com/friedelschoen/editor/internal/core"
	"github.com/friedelschoen/editor/internal/parser"
)

func CopyFilePosition(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}

	if !erow.Info.IsFileButNotDir() {
		return fmt.Errorf("not a file")
	}

	ta := erow.Row.TextArea
	ci := ta.CursorIndex()
	rd := ta.RW()
	line, col, err := parser.IndexLineColumn(rd, ci)
	if err != nil {
		return err
	}

	s := fmt.Sprintf("copyfileposition:\n\t%v:%v:%v", erow.Info.Name(), line, col)
	erow.Ed.Messagef(s)

	return nil
}
