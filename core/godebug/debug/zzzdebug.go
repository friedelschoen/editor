package debug

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
)

// Vars populated at an init that runs before this file (generated at compile).
var EncoderId string
var onExecSide bool // has generated config

//----------

// NOTE: init() functions declared across multiple files in a package are processed in alphabetical order of the file name

func init() {
	if !onExecSide {
		// NOTE: a built binary will only be able to run with this editor instance. On the other hand, it can self debug any part of the editor, including the debug pkg, inside an editor running from another editor.
		EncoderId = GenDigitsStr(10)
	} else {
		//EncoderId = // set by generated config
	}

	registerStructsForProtoConn(EncoderId)
	if onExecSide {
		es.init()
	}
}

//----------
//----------
//----------

// exec side options (set by generated config)
var eso struct {
	addr                Addr
	isServer            bool
	noInitMsg           bool
	srcLines            bool                 // warning at init msg
	syncSend            bool                 // don't send in chunks (slow)
	stringifyBytesRunes bool                 // "abc" instead of [97 98 99]
	filesData           []*AnnotatorFileData // all debug data

	// TODO: currently always true
	//acceptOnlyFirstConn bool // avoid possible hanging progs waiting for another connection to continue debugging (most common case)
}

//----------

// exec side
// runs before init(), needed because there could be an Exit() call throught some other init() func, before main() starts
var es = newES()

//----------

type ES struct {
	p     *Proto
	initw *InitWait
}

func newES() *ES {
	es := &ES{}
	es.initw = newInitWait()
	return es
}

//initExecSide2(initExecSide3)
// allows onload func to be defined when in syscal/js mode
//var initExecSide2 = func(fn func()) { fn() }
//func initExecSide3() {

func (es *ES) init() {
	defer es.initw.done()

	if !eso.noInitMsg {
		msg := "binary compiled with editor debug data. Use -noinitmsg to omit this msg."
		if !eso.srcLines {
			msg += fmt.Sprintf(" Note that in the case of panic, the src lines will not correspond to the original src code, but to the annotated src (-srclines=false).")
		}
		execSidePrintf("%v\n", msg)
	}

	fd := &FilesDataMsg{Data: eso.filesData}
	exs := &ProtoExecSide{fdata: fd, NoWriteBuffering: eso.syncSend}
	ctx := context.Background()
	es.p = NewProto(ctx, eso.isServer, eso.addr, exs)
	if err := es.p.Connect(); err != nil {
		execSidePrintError(err)
	}
}
func (es *ES) afterInitOk(fn func()) {
	mustBeExecSide()
	es.initw.wait()
	if !es.p.GotConnectedFastCheck() {
		return
	}
	fn()
}

//----------

func mustBeExecSide() {
	if !onExecSide {
		panic("not on exec side")
	}
}
func execSidePrintError(err error) {
	execSidePrintf("error: %v\n", err)
}
func execSidePrintf(f string, args ...interface{}) {
	mustBeExecSide()
	fmt.Fprintf(os.Stderr, "DEBUG: "+f, args...)
}

//----------

// Auto-inserted at defer main for a clean exit. Don't use.
func Close() {
	es.afterInitOk(func() {
		if err := es.p.WaitClose(); err != nil {
			execSidePrintError(err)
		}
	})
}

// Auto-inserted in annotated files to replace os.Exit calls. Don't use.
// Non-annotated files that call os.Exit will not let the editor receive all debug msgs. The sync msgs option will need to be used.
func Exit(code int) {
	Close()
	if !eso.noInitMsg {
		execSidePrintf("exit code: %v\n", code)
	}
	os.Exit(code)
}

// Auto-inserted at annotations. Don't use.
// NOTE: func name is used in annotator, don't rename.
func L(fileIndex, debugIndex, offset int, item Item) {
	lmsg := &LineMsg{
		FileIndex:  fileIndex,
		DebugIndex: debugIndex,
		Offset:     offset,
		Item:       item,
	}
	es.afterInitOk(func() {
		if err := es.p.WriteLineMsg(lmsg); err != nil {
			execSidePrintError(err)
		}
	})
}

//----------
//----------
//----------

func StartEditorSide(ctx context.Context, isServer bool, addr Addr) (*Proto, error) {
	eds := &ProtoEditorSide{}
	p := NewProto(ctx, isServer, addr, eds)
	err := p.Connect()
	return p, err
}

//----------
//----------
//----------

type InitWait struct {
	wg       *sync.WaitGroup
	waitSlow bool
}

func newInitWait() *InitWait {
	iw := &InitWait{}
	iw.wg = &sync.WaitGroup{}
	iw.wg.Add(1)
	return iw
}
func (iw *InitWait) wait() {
	if !iw.waitSlow {
		iw.waitSlow = true
		iw.wg.Wait()
	}
}
func (iw *InitWait) done() {
	iw.wg.Done()
}

//----------
//----------
//----------

func GenDigitsStr(n int) string {
	const src = "0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = src[rand.Intn(len(src))]
	}
	return string(b)
}
