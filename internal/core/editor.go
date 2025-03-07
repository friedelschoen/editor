package core

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/friedelschoen/glake/internal/command"
	"github.com/friedelschoen/glake/internal/drawer"
	"github.com/friedelschoen/glake/internal/fswatcher"
	"github.com/friedelschoen/glake/internal/ioutil"
	"github.com/friedelschoen/glake/internal/lsproto"
	"github.com/friedelschoen/glake/internal/toolbarparser"
	"github.com/friedelschoen/glake/internal/ui"
	"github.com/friedelschoen/glake/internal/ui/driver"
	"github.com/friedelschoen/glake/internal/ui/widget"
	"github.com/veandco/go-sdl2/sdl"
	"golang.org/x/image/font"
)

type Editor struct {
	UI                *ui.UI
	HomeVars          *toolbarparser.HomeVarMap
	Watcher           fswatcher.Watcher
	RowReopener       *RowReopener
	LSProtoMan        *lsproto.Manager
	InlineComplete    *InlineComplete
	Plugins           *Plugins
	EEvents           *EEvents // editor events (used by plugins)
	FsCaseInsensitive bool     // filesystem

	dndh         *DndHandler
	ifbw         *InfoFloatBoxWrap
	erowInfos    map[string]*ERowInfo // use ed.ERowInfo*() to access
	preSaveHooks []*PreSaveHook

	zipSessionsFile bool
}

func RunEditor(opt *Options) error {
	ed := &Editor{}
	ed.erowInfos = map[string]*ERowInfo{}
	ed.ifbw = NewInfoFloatBox(ed)

	// TODO: osx can have a case insensitive filesystem
	ed.FsCaseInsensitive = runtime.GOOS == "windows"

	ed.HomeVars = &toolbarparser.HomeVarMap{}
	ed.RowReopener = NewRowReopener(ed)
	ed.dndh = NewDndHandler(ed)
	ed.InlineComplete = NewInlineComplete(ed)
	ed.EEvents = NewEEvents()

	if err := ed.init(opt); err != nil {
		return err
	}

	go ed.fswatcherEventLoop()
	ed.uiEventLoop() // blocks

	return nil
}

func (ed *Editor) init(opt *Options) error {
	// fs watcher + gwatcher
	w, err := fswatcher.NewFsnWatcher()
	if err != nil {
		return err
	}
	ed.Watcher = fswatcher.NewGWatcher(w)

	ed.zipSessionsFile = opt.ZipSessionsFile

	ed.setupTheme(opt)

	// user interface
	ui0, err := ui.NewUI("Editor")
	if err != nil {
		return err
	}
	ed.UI = ui0
	ed.UI.OnError = ed.Error
	ed.setupUIRoot()

	// TODO: ensure it has the window measure
	ed.EnsureOneColumn()

	// setup plugins
	setupInitialRows := true
	err = ed.setupPlugins(opt)
	if err != nil {
		ed.Error(err)
		setupInitialRows = false
	}

	if setupInitialRows {
		// enqueue setup initial rows to run after UI has window measure
		ed.UI.RunOnUIGoRoutine(func() {
			ed.setupInitialRows(opt)
		})
	}

	ed.initLSProto(opt)
	ed.initPreSaveHooks(opt)

	return nil
}

func (ed *Editor) initLSProto(opt *Options) {
	// language server protocol manager
	ed.LSProtoMan = lsproto.NewManager(ed.Message)
	for _, reg := range opt.LSProtos.regs {
		ed.LSProtoMan.Register(reg)
	}

	// NOTE: argument for not having auto-registration: don't auto add since the lsproto server could have issues, and auto-adding doesn't allow the user to have a choice to using directly some other option (like a plugin)
	// NOTE: unlikely to be using a plugin for golang since gopls is fairly stable now, allow auto registration at least for ".go" if not present

	// auto setup gopls if there is no handler for ".go" files
	_, err := ed.LSProtoMan.LangManager("a.go")
	if err != nil { // no registration exists
		s := lsproto.GoplsRegistration(false, false, false)
		reg, err := lsproto.NewRegistration(s)
		if err != nil {
			panic(err)
		}
		_ = ed.LSProtoMan.Register(reg)
	}
}

func (ed *Editor) initPreSaveHooks(opt *Options) {
	// auto register "goimports" if no entry exists for the "go" language
	found := false
	for _, r := range opt.PreSaveHooks.regs {
		if r.Language == "go" {
			found = true
			break
		}
	}
	if !found {
		exec := "goimports"
		opt.PreSaveHooks.MustSet("go,.go," + exec)
	}

	ed.preSaveHooks = opt.PreSaveHooks.regs
}

func (ed *Editor) Close() {
	ed.LSProtoMan.Close()
	ed.UI.Events <- &driver.WindowClose{}
}

func (ed *Editor) uiEventLoop() {
	defer ed.UI.Close()

	for {
		ed.UI.PollEvent()
		var ev driver.Event
		select {
		case e := <-ed.UI.Events:
			ev = e
		default:
			continue
		}
		switch t := ev.(type) {
		case *driver.WindowClose:
			return
		case *driver.DndPosition:
			ed.dndh.OnPosition(t)
		case *driver.DndDrop:
			ed.dndh.OnDrop(t)
		default:
			if !ed.handleGlobalShortcuts(ev) {
				if !ed.UI.HandleEvent(ev) {
					log.Printf("uievloop: unhandled event: %#v", ev)
				}
			}
		}
		ed.UI.LayoutMarkedAndSchedulePaint()
	}
}

func (ed *Editor) fswatcherEventLoop() {
	for {
		ev, ok := <-ed.Watcher.Events()
		if !ok {
			ed.Close()
			return
		}
		switch evt := ev.(type) {
		case error:
			ed.Error(evt)
		case *fswatcher.Event:
			ed.handleWatcherEvent(evt)
		}
	}
}

func (ed *Editor) handleWatcherEvent(ev *fswatcher.Event) {
	info, ok := ed.ERowInfo(ev.Name)
	if ok {
		ed.UI.RunOnUIGoRoutine(func() {
			info.UpdateDiskEvent()
		})
	}
}

func (ed *Editor) Errorf(f string, a ...any) {
	ed.Error(fmt.Errorf(f, a...))
}
func (ed *Editor) Error(err error) {
	header := "error"
	if errors.Is(err, flag.ErrHelp) {
		header = "usage"
	}
	ed.Messagef("%v: %v", header, err)
}

func (ed *Editor) Messagef(f string, a ...any) {
	ed.Message(fmt.Sprintf(f, a...))
}

func (ed *Editor) Message(s string) {
	// ensure newline
	if !strings.HasSuffix(s, "\n") {
		s = s + "\n"
	}

	ed.UI.RunOnUIGoRoutine(func() {
		erow := ed.messagesERow()

		// index to make visible, get before append
		ta := erow.Row.TextArea
		index := ta.Len()

		erow.AppendBytesClearHistory([]byte(s))

		// don't count spaces at the end for closer bottom alignment
		u := strings.TrimRightFunc(s, unicode.IsSpace)
		erow.MakeRangeVisibleAndFlash(index, len(u))
	})
}

func (ed *Editor) messagesERow() *ERow {
	erow, isNew := ExistingERowOrNewBasic(ed, "+Messages")
	if isNew {
		erow.ToolbarSetStrAfterNameClearHistory(" | Clear")
	}
	return erow
}

func (ed *Editor) ReadERowInfo(name string) *ERowInfo {
	return readERowInfoOrNew(ed, name)
}

func (ed *Editor) ERowInfo(name string) (*ERowInfo, bool) {
	k := ed.ERowInfoKey(name)
	info, ok := ed.erowInfos[k]
	return info, ok
}

func (ed *Editor) ERowInfos() []*ERowInfo {
	// stable list
	keys := []string{}
	for k := range ed.erowInfos {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	u := make([]*ERowInfo, len(ed.erowInfos))
	for i, k := range keys {
		u[i] = ed.erowInfos[k]
	}
	return u
}

func (ed *Editor) ERowInfoKey(name string) string {
	if ed.FsCaseInsensitive {
		return strings.ToLower(name)
	}
	return name
}

func (ed *Editor) SetERowInfo(name string, info *ERowInfo) {
	k := ed.ERowInfoKey(name)
	ed.erowInfos[k] = info
}

func (ed *Editor) DeleteERowInfo(name string) {
	k := ed.ERowInfoKey(name)
	delete(ed.erowInfos, k)
}

func (ed *Editor) ERows() []*ERow {
	w := []*ERow{}
	for _, info := range ed.ERowInfos() {
		w = append(w, info.ERows...)
	}
	return w
}

func (ed *Editor) GoodRowPos() *ui.RowPos {
	return ed.UI.GoodRowPos()
}

func (ed *Editor) ActiveERow() (*ERow, bool) {
	for _, e := range ed.ERows() {
		if e.Row.HasState(ui.RowStateActive) {
			return e, true
		}
	}
	return nil, false
}

func (ed *Editor) setupUIRoot() {
	ed.setupRootToolbar()
	ed.setupRootMenuToolbar()

	// ui.root select annotation
	// ed.UI.Root.EvReg.Add(ui.RootSelectAnnotationEventId, func(ev any) {
	// 	rowPos := ed.GoodRowPos()
	// 	ev2 := ev.(*ui.RootSelectAnnotationEvent)
	// 	ed.GoDebug.SelectAnnotation(rowPos, ev2)
	// })
}

func (ed *Editor) setupRootToolbar() {
	tb := ed.UI.Root.Toolbar
	// cmd event
	tb.EvReg.Add(ui.TextAreaCmdEventId, func(ev any) {
		InternalCmdFromRootTb(ed, tb)
	})
	// on write
	tb.RWEvReg.Add(ioutil.RWEvIdWrite, func(ev0 any) {
		ed.updateERowsToolbarsHomeVars()
	})

	s := "Exit ListSessions NewColumn NewRow ReopenRow Reload LsprotoCloseAll Stop"
	tb.SetStrClearHistory(s)
}

func (ed *Editor) setupRootMenuToolbar() {
	tb := ed.UI.Root.MainMenuButton.Toolbar
	// cmd event
	tb.EvReg.Add(ui.TextAreaCmdEventId, func(ev any) {
		InternalCmdFromRootTb(ed, tb)
	})
	// on write
	tb.RWEvReg.Add(ioutil.RWEvIdWrite, func(ev0 any) {
		ed.updateERowsToolbarsHomeVars()
	})

	w := [][]string{
		{"ColorTheme", "FontTheme"},
		{"CopyFilePosition"},
		{"CtxutilCallsState"},
		{"Find -h"},
		{"FontRunes", "RuneCodes"},
		{"GoDebug -h", "GoDebug run -h", "GoDebug connect -h"},
		{"GoDebugFind"},
		{"GoDebugTrace"},
		{"GotoLine"},
		{"ListDir", "ListDir -hidden", "ListDir -sub"},
		{"ListSessions", "OpenSession", "DeleteSession", "SaveSession"},
		{"LsprotoRename", "LsprotoCloseAll", "LsprotoCallers", "LsprotoCallees", "LsprotoReferences"},
		{"NewColumn", "NewRow", "ReopenRow", "MaximizeRow"},
		{"NewFile", "SaveAllFiles", "Save"},
		{"OpenExternal", "OpenFilemanager", "OpenTerminal"},
		{"Reload", "ReloadAll", "ReloadAllFiles"},
		{"SortTextLines", "SortTextLines -h"},
	}
	last := []string{"Exit", "Version", "Stop", "Clear"}

	// simple sorted list
	w2 := []string{}
	for _, a := range w {
		//w2 = append(w2, strings.Join(a, "|")) // TODO: ui test issue
		w2 = append(w2, a...)
	}
	sort.Slice(w2, func(a, b int) bool { return w2[a] < w2[b] })
	w2 = append(w2, "\n"+strings.Join(last, " | "))
	s1 := strings.Join(w2, "\n")

	tb.SetStrClearHistory(s1)
}

func (ed *Editor) updateERowsToolbarsHomeVars() {
	tb1 := ed.UI.Root.Toolbar.Str()
	tb2 := ed.UI.Root.MainMenuButton.Toolbar.Str()
	ed.HomeVars = toolbarparser.ParseToolbarVars([]string{tb1, tb2}, ed.FsCaseInsensitive)
	for _, erow := range ed.ERows() {
		erow.UpdateToolbarNameEncoding()
	}
}

func (ed *Editor) setupInitialRows(opt *Options) {
	if opt.SessionName != "" {
		OpenSessionFromString(ed, opt.SessionName)
		return
	}

	// cmd line filenames to open
	if len(opt.Filenames) > 0 {
		col := ed.UI.Root.Cols.FirstChildColumn()
		for _, filename := range opt.Filenames {
			// try to use absolute path
			u, err := filepath.Abs(filename)
			if err == nil {
				filename = u
			}

			info := ed.ReadERowInfo(filename)
			if len(info.ERows) == 0 {
				rowPos := ui.NewRowPos(col, nil)
				_, err := NewLoadedERow(info, rowPos)
				if err != nil {
					ed.Error(err)
				}
			}
		}
		return
	}

	// open current directory
	dir, err := os.Getwd()
	if err == nil {
		// create a second column (one should exist already)
		_ = ed.NewColumn()

		// open directory
		info := ed.ReadERowInfo(dir)
		cols := ed.UI.Root.Cols
		rowPos := ui.NewRowPos(cols.LastChildColumn(), nil)
		_ = NewLoadedERowOrNewBasic(info, rowPos)
	}
}

func (ed *Editor) setupTheme(opt *Options) {
	drawer.WrapLineRune, _ = utf8.DecodeRuneInString(opt.WrapLineRune)
	// fontcache.TabWidth = opt.TabWidth
	// fontcache.CarriageReturnRune, _ = utf8.DecodeRuneInString(opt.CarriageReturnRune)
	ui.ScrollBarLeft = opt.ScrollBarLeft
	ui.ScrollBarWidth = opt.ScrollBarWidth
	ui.ShadowsOn = opt.Shadows

	// color theme
	ui.ColorTheme = opt.ColorTheme

	// font options
	ui.TTFontOptions.DPI = opt.DPI
	ui.TTFontOptions.Size = opt.FontSize
	switch opt.FontHinting {
	case "none":
		ui.TTFontOptions.Hinting = font.HintingNone
	case "vertical":
		ui.TTFontOptions.Hinting = font.HintingVertical
	case "full":
		ui.TTFontOptions.Hinting = font.HintingFull
	default:
		fmt.Fprintf(os.Stderr, "unknown font hinting: %v\n", opt.FontHinting)
		os.Exit(2)
	}

	// font theme
	ui.CurrentFont = opt.Font
}

func (ed *Editor) setupPlugins(opt *Options) error {
	ed.Plugins = NewPlugins(ed)
	a := strings.Split(opt.Plugins, ",")
	for _, s := range a {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		err := ed.Plugins.AddPath(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ed *Editor) EnsureOneColumn() {
	if ed.UI.Root.Cols.ColsLayout.Spl.ChildsLen() == 0 {
		_ = ed.NewColumn()
	}
}

func (ed *Editor) NewColumn() *ui.Column {
	col := ed.UI.Root.Cols.NewColumn()
	// close
	col.EvReg.Add(ui.ColumnCloseEventId, func(ev0 any) {
		ed.EnsureOneColumn()
	})
	return col
}

func (ed *Editor) handleGlobalShortcuts(ev any) (handled bool) {
	switch t := ev.(type) {
	case driver.KeyDown:
		autoCloseInfo := true

		switch {
		case t.Key.Is("Escape"):
			ed.InlineComplete.CancelAndClear()
			ed.cancelERowInfosCmds()
			ed.cancelERowsContentCmds()
			ed.cancelERowsInternalCmds()
			autoCloseInfo = false
			ed.cancelInfoFloatBox()
			return true
		case t.Key.Is("F1"):
			autoCloseInfo = false
			ed.toggleInfoFloatBox()
			return true
		}

		if autoCloseInfo {
			x, y, _ := sdl.GetMouseState()
			ed.UI.Root.ContextFloatBox.AutoClose(t, image.Point{int(x), int(y)})
			if !ed.ifbw.ui().Visible() {
				ed.cancelInfoFloatBox()
			}
		}
	}
	return false
}

// example cmds canceled: openfilename, opensession, ...
func (ed *Editor) cancelERowsContentCmds() {
	for _, erow := range ed.ERows() {
		erow.CancelContentCmd()
	}
}

// example cmds canceled: GoDebug, Lsproto*, ...
func (ed *Editor) cancelERowsInternalCmds() {
	for _, erow := range ed.ERows() {
		erow.CancelInternalCmd()
	}
}

// example cmd canceled: presavehooks (goimports, src formatters, ...)
func (ed *Editor) cancelERowInfosCmds() {
	for _, info := range ed.ERowInfos() {
		info.CancelCmd()
	}
}

func (ed *Editor) cancelInfoFloatBox() {
	ed.ifbw.Cancel()
	cfb := ed.ifbw.ui()
	cfb.Hide()
}

func (ed *Editor) toggleInfoFloatBox() {
	ed.ifbw.Cancel() // cancel previous run

	// toggle
	cfb := ed.ifbw.ui()
	cfb.Toggle()
	if !cfb.Visible() {
		return
	}

	// showInfoFloatBox

	// find ta/erow under pointer
	ta, ok := cfb.FindTextAreaUnderPointer()
	if !ok {
		cfb.Hide()
		return
	}
	erow, ok := ed.NodeERow(ta)
	if !ok {
		cfb.Hide()
		return
	}

	// show util
	show := func(s string) {
		cfb.TextArea.ClearPos()
		cfb.SetStrClearHistory(s)
		cfb.Show()
	}
	showAsync := func(s string) {
		ed.UI.RunOnUIGoRoutine(func() {
			if cfb.Visible() {
				show(s)
			}
		})
	}

	// initial ui feedback at position
	cfb.SetRefPointToTextAreaCursor(ta)
	show("Loading...")

	ed.RunAsyncBusyCursor(cfb, func() {
		// there is no timeout to complete since the context can be canceled manually

		// context based on erow context
		ctx := ed.ifbw.NewCtx(erow.ctx)

		// plugin autocomplete
		showAsync("Loading plugin...")
		err, handled := ed.Plugins.RunAutoComplete(ctx, cfb)
		if handled {
			if err != nil {
				ed.Error(err)
			}
			return
		}

		// lsproto autocomplete
		filename := ""
		switch ta {
		case erow.Row.TextArea:
			if erow.Info.IsDir() {
				filename = ".editor_directory"
			} else {
				filename = erow.Info.Name()
			}
		case erow.Row.Toolbar.TextArea:
			filename = ".editor_toolbar"
		default:
			showAsync("")
			return
		}
		// handle filename
		lang, err := ed.LSProtoMan.LangManager(filename)
		if err != nil {
			showAsync(err.Error()) // err:"no registration for..."
			return
		}
		// ui feedback while loading
		v := fmt.Sprintf("Loading lsproto(%v)...", lang.Reg.Language)
		showAsync(v)
		// lsproto autocomplete
		s, err := ed.lsprotoManAutoComplete(ctx, ta, erow)
		if err != nil {
			ed.Error(err)
			showAsync("")
			return
		}
		showAsync(s)
	})
}

func (ed *Editor) lsprotoManAutoComplete(ctx context.Context, ta *ui.TextArea, erow *ERow) (string, error) {
	//ta := erow.Row.TextArea
	comps, err := ed.LSProtoMan.TextDocumentCompletionDetailStrings(ctx, erow.Info.Name(), ta.RW(), ta.CursorIndex())
	if err != nil {
		return "", err
	}
	s := "0 results"
	if len(comps) > 0 {
		s = strings.Join(comps, "\n")
	}
	return s, nil
}

func (ed *Editor) NodeERow(node widget.Node) (*ERow, bool) {
	for p := node.Embed().Parent; p != nil; p = p.Parent {
		if r, ok := p.Wrapper.(*ui.Row); ok {
			for _, erow := range ed.ERows() {
				if r == erow.Row {
					return erow, true
				}
			}
		}
	}
	return nil, false
}

func (ed *Editor) RunAsyncBusyCursor(node widget.Node, fn func()) {
	ed.RunAsyncBusyCursor2(node, func(done func()) { fn(); done() })
}

// Caller should call done function in the end.
func (ed *Editor) RunAsyncBusyCursor2(node widget.Node, fn func(done func())) {
	set := func(c sdl.SystemCursor) {
		ed.UI.RunOnUIGoRoutine(func() {
			node.Embed().Cursor = c
			ed.UI.QueueEmptyWindowInputEvent() // updates cursor tree
		})
	}
	set(sdl.SYSTEM_CURSOR_WAITARROW)
	done := func() {
		set(sdl.SYSTEM_CURSOR_ARROW)
	}
	// launch go routine to allow the UI to update the cursor
	go fn(done)
}

func (ed *Editor) SetAnnotations(req EdAnnotationsRequester, ta *ui.TextArea, on bool, selIndex int, entries *drawer.AnnotationGroup) {
	// avoid lockup:
	// godebugstart->inlinecomplete.clear->godebugrestoreannotations
	ed.UI.RunOnUIGoRoutine(func() {
		ed.setAnnotations2(req, ta, on, selIndex, entries)
	})
}
func (ed *Editor) setAnnotations2(req EdAnnotationsRequester, ta *ui.TextArea, on bool, selIndex int, entries *drawer.AnnotationGroup) {
	if !ed.CanModifyAnnotations(req, ta) {
		return
	}
	// set annotations (including clear)
	ta.Drawer.Opt.Annotations.On = on
	ta.Drawer.Opt.Annotations.Selected.EntryIndex = selIndex
	ta.Drawer.Opt.Annotations.Entries = entries
	ta.MarkNeedsLayoutAndPaint()

	// // restore godebug annotations
	// if req == EareqInlineComplete && !on {
	// 	// find erow info from textarea
	// 	for _, erow := range ed.ERows() {
	// 		if erow.Row.TextArea == ta {
	// 			ed.GoDebug.UpdateInfoAnnotations(erow.Info)
	// 		}
	// 	}
	// }
}

func (ed *Editor) CanModifyAnnotations(req EdAnnotationsRequester, ta *ui.TextArea) bool {
	switch req {
	case EareqGoDebugStart:
		ed.InlineComplete.CancelAndClear()
		return true
	case EareqGoDebug:
		if ed.InlineComplete.IsOn(ta) {
			return false
		}
		return true
	case EareqInlineComplete:
		return true
	default:
		panic(req)
	}
}

func (ed *Editor) runPreSaveHooks(ctx context.Context, info *ERowInfo, b []byte) ([]byte, error) {
	ext := filepath.Ext(info.Name())
	for _, h := range ed.preSaveHooks {
		for _, e := range h.Exts {
			if e == ext {
				b2, err := ed.runPreSaveHook(ctx, info, b, h.Cmd)
				if err != nil {
					err2 := fmt.Errorf("presavehook(%v): %w", h.Language, err)
					return nil, err2
				}
				b = b2
			}
		}
	}
	return b, nil
}

func (ed *Editor) runPreSaveHook(ctx context.Context, info *ERowInfo, content []byte, cmd string) ([]byte, error) {
	// timeout for the cmd to run
	timeout := 5 * time.Second
	ctx2, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dir := filepath.Dir(info.Name())
	r := bytes.NewReader(content)
	cmd2 := strings.Split(cmd, " ")

	return command.RunCmdStdin(ctx2, dir, r, cmd2...)
}

func (ed *Editor) loadSessions() (*Sessions, error) {
	return ed.loadSessions2()
}
func (ed *Editor) loadSessions2() (*Sessions, error) {
	loadPlain := func() (*Sessions, error) {
		filename := sessionsFilename()
		return newSessionsFromPlain(filename)
	}
	loadZip := func() (*Sessions, error) {
		zipFilename, filename := sessionsZipFilenames()
		return newSessionsFromZip(zipFilename, filename)
	}

	load0, load1 := loadPlain, loadZip
	if ed.zipSessionsFile {
		load0, load1 = loadZip, loadPlain
	}

	ss, err := load0()
	// try to load the other way to allow transition
	if errors.Is(err, os.ErrNotExist) {
		ss2, err2 := load1()
		if errors.Is(err2, os.ErrNotExist) {
			// both non-existent, allow newsessions without error
			ss, err = &Sessions{}, nil
		} else {
			ss, err = ss2, err2
		}
	}

	return ss, err
}

func (ed *Editor) saveSessions(ss *Sessions) error {
	return ed.saveSessions2(ss)
}
func (ed *Editor) saveSessions2(ss *Sessions) error {
	hasPlainFile := func() bool {
		filename := sessionsFilename()
		_, err := os.Stat(filename)
		return err == nil
	}
	hasZipFile := func() bool {
		zipFilename, _ := sessionsZipFilenames()
		_, err := os.Stat(zipFilename)
		return err == nil
	}

	if ed.zipSessionsFile || (hasZipFile() && !hasPlainFile()) {
		zipFilename, filename := sessionsZipFilenames()
		return ss.saveToZip(zipFilename, filename)
	} else {
		filename := sessionsFilename()
		return ss.saveToPlain(filename)
	}
}

type EdAnnotationsRequester int

const (
	EareqGoDebug EdAnnotationsRequester = iota
	EareqGoDebugStart
	EareqInlineComplete
)

type InfoFloatBoxWrap struct {
	ed   *Editor
	ctx  context.Context
	canc context.CancelFunc
}

func NewInfoFloatBox(ed *Editor) *InfoFloatBoxWrap {
	return &InfoFloatBoxWrap{ed: ed}
}
func (ifbw *InfoFloatBoxWrap) NewCtx(ctx context.Context) context.Context {
	ifbw.Cancel() // cancel previous
	ifbw.ctx, ifbw.canc = context.WithCancel(ctx)
	return ifbw.ctx
}
func (ifbw *InfoFloatBoxWrap) Cancel() {
	if ifbw.canc != nil {
		ifbw.canc()
		ifbw.canc = nil
	}
}
func (ifbw *InfoFloatBoxWrap) ui() *ui.ContextFloatBox {
	return ifbw.ed.UI.Root.ContextFloatBox
}
