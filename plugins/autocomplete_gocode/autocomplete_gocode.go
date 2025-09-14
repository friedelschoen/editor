/*
Build with:
$ go build -buildmode=plugin autocomplete_gocode.go
*/

package main

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"path/filepath"
	"time"

	"github.com/friedelschoen/editor/core"
	"github.com/friedelschoen/editor/ui"
	"github.com/friedelschoen/editor/util/osutil"
)

func AutoComplete(ctx context.Context, ed *core.Editor, cfb *ui.ContextFloatBox) (_ error, handled bool) {
	ta, ok := cfb.FindTextAreaUnderPointer()
	if !ok {
		cfb.Hide()
		return nil, false
	}

	erow, ok := ed.NodeERow(ta)
	if ok {
		ok = autoCompleteERow(ed, cfb, erow)
		if ok {
			return nil, true
		}
	}

	cfb.SetRefPointToTextAreaCursor(ta)
	cfb.TextArea.SetStr("no results")
	return nil, true
}

func autoCompleteERow(ed *core.Editor, cfb *ui.ContextFloatBox, erow *core.ERow) bool {
	if erow.Info.IsFileButNotDir() && path.Ext(erow.Info.Name()) == ".go" {
		autoCompleteERowGolang(ed, cfb, erow)
		return true
	}
	return false
}

//----------

func autoCompleteERowGolang(ed *core.Editor, cfb *ui.ContextFloatBox, erow *core.ERow) {
	// timeout for the cmd to run
	timeout := 8000 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// gocode args
	filename := erow.Info.Name()
	offset := erow.Row.TextArea.CursorIndex()
	args := []string{osutil.ExecName("gocode"), "autocomplete", fmt.Sprintf("%v", offset)}

	// gocode can read from stdin: use textarea bytes
	bin, err := erow.Row.TextArea.Bytes()
	if err != nil {
		ed.Error(err)
		return
	}
	in := bytes.NewBuffer(bin)

	// execute external cmd
	dir := filepath.Dir(filename)
	bout, err := osutil.RunCmdStdin(ctx, dir, in, args...)
	if err != nil {
		ed.Error(err)
		return
	}

	cfb.SetRefPointToTextAreaCursor(erow.Row.TextArea)
	cfb.TextArea.SetStr(string(bout))
	cfb.TextArea.ClearPos()
}
