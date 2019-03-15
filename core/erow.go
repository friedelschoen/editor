package core

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/jmigpin/editor/core/toolbarparser"
	"github.com/jmigpin/editor/ui"
	"github.com/jmigpin/editor/util/iout"
	"github.com/jmigpin/editor/util/uiutil/event"
)

//----------

type ERow struct {
	Ed     *Editor
	Row    *ui.Row
	Info   *ERowInfo
	Exec   *ERowExec
	TbData toolbarparser.Data

	highlightDuplicates           bool
	disableTextAreaSetStrCallback bool
}

//----------

func NewERow(ed *Editor, info *ERowInfo, rowPos *ui.RowPos) *ERow {
	// create row
	row := rowPos.Column.NewRowBefore(rowPos.NextRow)

	erow := &ERow{Ed: ed, Row: row, Info: info}
	erow.Exec = NewERowExec(erow)

	// TODO: join with updateToolbarPart0
	s2 := ed.HomeVars.Encode(erow.Info.Name())
	erow.Row.Toolbar.SetStrClearHistory(s2)

	erow.initHandlers()
	erow.parseToolbar() // after handlers are set
	erow.setupTextAreaCommentString()

	return erow
}

//----------

func (erow *ERow) initHandlers() {
	row := erow.Row

	// register with the editor
	erow.Ed.ERowInfos[erow.Info.Name()] = erow.Info
	erow.Info.AddERow(erow)

	// update row state
	erow.Info.UpdateDuplicateRowState()
	erow.Info.UpdateDuplicateHighlightRowState()
	erow.Info.UpdateExistsRowState()
	erow.Info.UpdateFsDifferRowState()

	// register with watcher
	if !erow.Info.IsSpecial() && len(erow.Info.ERows) == 1 {
		erow.Ed.Watcher.Add(erow.Info.Name())
	}

	// toolbar set str
	row.Toolbar.EvReg.Add(ui.TextAreaSetStrEventId, func(ev0 interface{}) {
		erow.parseToolbar()
	})
	// toolbar cmds
	row.Toolbar.EvReg.Add(ui.TextAreaCmdEventId, func(ev0 interface{}) {
		RowToolbarCmd(erow)
	})
	// textarea set str
	row.TextArea.EvReg.Add(ui.TextAreaSetStrEventId, func(ev0 interface{}) {
		//ev := ev0.(*ui.TextAreaSetStrEvent)

		if erow.disableTextAreaSetStrCallback {
			return
		}

		erow.Info.SetRowsStrFromMaster(erow)
	})
	// textarea edit
	row.TextArea.EvReg.Add(ui.TextAreaEditEventId, func(ev0 interface{}) {
		ev := ev0.(*ui.TextAreaWriteOpEvent)
		// update duplicate edits to keep offset/cursor in position
		if erow.Info.IsFileButNotDir() {
			for _, e := range erow.Info.ERows {
				if e == erow {
					continue
				}
				e.Row.TextArea.UpdateWriteOp(ev.WriteOp)
			}
		}
	})
	// textarea content cmds
	row.TextArea.EvReg.Add(ui.TextAreaCmdEventId, func(ev0 interface{}) {
		ev := ev0.(*ui.TextAreaCmdEvent)
		runContentCmds(erow, ev.Index)
	})
	// textarea select annotation
	row.TextArea.EvReg.Add(ui.TextAreaSelectAnnotationEventId, func(ev0 interface{}) {
		ev := ev0.(*ui.TextAreaSelectAnnotationEvent)
		GoDebugSelectAnnotation(erow, ev.AnnotationIndex, ev.Offset, ev.Type)
	})
	// key shortcuts
	row.EvReg.Add(ui.RowInputEventId, func(ev0 interface{}) {
		ev := ev0.(*ui.RowInputEvent)
		switch evt := ev.Event.(type) {
		case *event.KeyDown:
			mods := evt.Mods.ClearLocks()
			switch {
			case mods.Is(event.ModCtrl) && evt.LowerRune() == 's':
				if err := erow.Info.SaveFile(); err != nil {
					erow.Ed.Error(err)
				}
			case mods.Is(event.ModCtrl) && evt.LowerRune() == 'f':
				FindShortcut(erow)
			}
		case *event.MouseEnter:
			erow.highlightDuplicates = true
			erow.Info.UpdateDuplicateHighlightRowState()
		case *event.MouseLeave:
			erow.highlightDuplicates = false
			erow.Info.UpdateDuplicateHighlightRowState()
		}
	})
	// close
	row.EvReg.Add(ui.RowCloseEventId, func(ev0 interface{}) {
		// ensure execution (if any) is stopped
		erow.Exec.Stop()

		// unregister from editor
		erow.Info.RemoveERow(erow)
		if len(erow.Info.ERows) == 0 {
			delete(erow.Ed.ERowInfos, erow.Info.Name())
		}

		// update row state
		erow.Info.UpdateDuplicateRowState()
		erow.Info.UpdateDuplicateHighlightRowState()

		// unregister with watcher
		if !erow.Info.IsSpecial() && len(erow.Info.ERows) == 0 {
			erow.Ed.Watcher.Remove(erow.Info.Name())
		}

		// add to reopener to allow to reopen later if needed
		if !erow.Info.IsSpecial() {
			erow.Ed.RowReopener.Add(row)
		}
	})
}

//----------

func (erow *ERow) parseToolbar() {
	str := erow.Row.Toolbar.Str()

	data := toolbarparser.Parse(str)

	// don't allow toolbar edit of the name
	ename := erow.Ed.HomeVars.Encode(erow.Info.Name())
	arg0, ok := data.Part0Arg0()
	if !ok {
		erow.Row.Toolbar.TextHistory.Undo()
		erow.Row.Toolbar.TextHistory.ClearForward()
		erow.Ed.Errorf("unable to get toolbar name")
		return
	}
	ename2 := arg0.UnquotedStr()
	if ename2 != ename {
		erow.Row.Toolbar.TextHistory.Undo()
		erow.Row.Toolbar.TextHistory.ClearForward()
		erow.Ed.Errorf("can't change toolbar name")
		return
	}

	erow.TbData = *data
}

//----------

func (erow *ERow) updateToolbarPart0() {
	str := erow.Row.Toolbar.Str()
	data := toolbarparser.Parse(str)
	arg0, ok := data.Part0Arg0()
	if !ok {
		return
	}

	// replace part0 arg0 with encoded name
	ename := erow.Ed.HomeVars.Encode(erow.Info.Name())
	str2 := ename + str[arg0.End:]
	if str2 != str {
		erow.Row.Toolbar.SetStrClearHistory(str2)
	}
}

//----------

func (erow *ERow) Reload() {
	if err := erow.reload(); err != nil {
		erow.Ed.Error(err)
	}
}

func (erow *ERow) reload() error {
	switch {
	case erow.Info.IsDir():
		return erow.Info.ReloadDir(erow)
	case erow.Info.IsFileButNotDir():
		return erow.Info.ReloadFile()
	default:
		return errors.New("unexpected type to reload")
	}
}

//----------

func (erow *ERow) ToolbarSetStrAfterNameClearHistory(s string) {
	arg, ok := erow.TbData.Part0Arg0()
	if !ok {
		return
	}
	i := arg.End
	str := erow.Row.Toolbar.Str()[:i] + s
	erow.Row.Toolbar.SetStrClearHistory(str)
}

//----------

func (erow *ERow) TextAreaAppendAsync(str string) <-chan struct{} {
	comm := make(chan struct{})
	erow.Ed.UI.RunOnUIGoRoutine(func() {
		erow.textAreaAppend(str)
		close(comm)
	})
	return comm
}

func (erow *ERow) textAreaAppend(str string) {
	// TODO: unlimited output? some xterms have more or less 64k limit. Bigger limits will slow down the ui since it will be calculating the new string content. This will be improved once the textarea drawer supports append/cutTop operations.

	maxSize := 64 * 1024

	ta := erow.Row.TextArea
	if err := ta.AppendStrClearHistory(str, maxSize); err != nil {
		erow.Ed.Error(err)
	}
}

//----------

// Caller is responsible for closing the writer at the end.
func (erow *ERow) TextAreaWriter() io.WriteCloser {
	pr, pw := io.Pipe()
	go func() {
		erow.readLoopToTextArea(pr)
	}()

	// terminal escape sequences filter
	var wc io.WriteCloser = pw
	if erow.Info.IsDir() {
		wc = NewTerminalFilter(erow, wc)
	}

	return iout.NewAutoBufWriter(wc)
}

func (erow *ERow) readLoopToTextArea(reader io.Reader) {
	var buf [4 * 1024]byte
	for {
		n, err := reader.Read(buf[:])
		if n > 0 {
			str := string(buf[:n])
			c := erow.TextAreaAppendAsync(str)

			// Wait for the ui to have handled the content. This prevents a tight loop program from leaving the UI unresponsive.
			<-c
		}
		if err != nil {
			break
		}
	}
}

//----------

func (erow *ERow) Flash() {
	p, ok := erow.TbData.PartAtIndex(0)
	if ok {
		if len(p.Args) > 0 {
			a := p.Args[0]
			erow.Row.Toolbar.FlashIndexLen(a.Pos, a.End-a.Pos)
		}
	}
}

//----------

func (erow *ERow) MakeIndexVisibleAndFlash(index int) {
	erow.MakeRangeVisibleAndFlash(index, 0)
}
func (erow *ERow) MakeRangeVisibleAndFlash(index int, len int) {
	erow.Row.EnsureTextAreaMinimumHeight()
	erow.Row.TextArea.MakeRangeVisible(index, len)
	erow.Row.TextArea.FlashIndexLen(index, len)

	// flash toolbar as last resort
	//if !erow.Row.TextArea.IsRangeVisible(index, len) {
	b := &erow.Row.TextArea.Bounds
	if b.Dx() < 10 || b.Dy() < 10 { // TODO: use dpi instead of fixed pixels
		erow.Flash()
	}
}

//----------

func (erow *ERow) setupTextAreaCommentString() {
	ta := erow.Row.TextArea
	switch filepath.Ext(erow.Info.Name()) {
	default:
		fallthrough
	case "", ".sh", ".conf", ".list", ".txt":
		ta.SetCommentStrings("#", [2]string{})
	case ".go", ".c", ".cpp", ".h", ".hpp":
		ta.SetCommentStrings("//", [2]string{"/*", "*/"})
	}
}
