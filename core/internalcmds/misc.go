package internalcmds

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/jmigpin/editor/core"
	"github.com/jmigpin/editor/core/godebug"
	"github.com/jmigpin/editor/ui"
	"github.com/jmigpin/editor/util/ctxutil"
	"github.com/jmigpin/editor/util/iout"
	"github.com/jmigpin/editor/util/osutil"
)

//----------

func Version(args *core.InternalCmdArgs) error {
	args.Ed.Messagef("version: %v", core.Version())
	return nil
}

//----------

func Exit(args *core.InternalCmdArgs) error {
	args.Ed.Close()
	return nil
}

//----------

func SaveSession(args *core.InternalCmdArgs) error {
	core.SaveSession(args.Ed, args.Part)
	return nil
}
func OpenSession(args *core.InternalCmdArgs) error {
	core.OpenSession(args.Ed, args.Part)
	return nil
}
func DeleteSession(args *core.InternalCmdArgs) error {
	core.DeleteSession(args.Ed, args.Part)
	return nil
}
func ListSessions(args *core.InternalCmdArgs) error {
	core.ListSessions(args.Ed)
	return nil
}

//----------

func NewColumn(args *core.InternalCmdArgs) error {
	args.Ed.NewColumn()
	return nil
}
func CloseColumn(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	erow.Row.Col.Close()
	return nil
}

//----------

func CloseRow(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	erow.Row.Close()
	return nil
}
func ReopenRow(args *core.InternalCmdArgs) error {
	args.Ed.RowReopener.Reopen()
	return nil
}
func MaximizeRow(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	erow.Row.Maximize()
	return nil
}

//----------

func Save(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	return erow.Info.SaveFile()
}
func SaveAllFiles(args *core.InternalCmdArgs) error {
	var me iout.MultiError
	for _, info := range args.Ed.ERowInfos() {
		if info.IsFileButNotDir() {
			me.Add(info.SaveFile())
		}
	}
	return me.Result()
}

//----------

func Reload(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	return erow.Reload()
}
func ReloadAllFiles(args *core.InternalCmdArgs) error {
	me := &iout.MultiError{}
	for _, info := range args.Ed.ERowInfos() {
		if info.IsFileButNotDir() {
			me.Add(info.ReloadFile())
		}
	}
	return me.Result()
}
func ReloadAll(args *core.InternalCmdArgs) error {
	// reload all dirs erows
	me := &iout.MultiError{}
	for _, info := range args.Ed.ERowInfos() {
		if info.IsDir() {
			for _, erow := range info.ERows {
				me.Add(erow.Reload())
			}
		}
	}

	me.Add(ReloadAllFiles(args))

	return me.Result()
}

//----------

func Stop(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	erow.Exec.Stop()
	return nil
}

//----------

func Clear(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	erow.Row.TextArea.SetStrClearHistory("")
	return nil
}

//----------

func OpenFilemanager(args *core.InternalCmdArgs) error {
	dir := ""
	erow, ok := args.ERow()
	if ok && !erow.Info.IsSpecial() {
		dir = erow.Info.Dir()
	} else {
		d, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = d
	}

	return osutil.OpenFilemanager(dir)
}

func OpenTerminal(args *core.InternalCmdArgs) error {
	dir := ""
	erow, ok := args.ERow()
	if ok && !erow.Info.IsSpecial() {
		dir = erow.Info.Dir()
	} else {
		d, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = d
	}

	return osutil.OpenTerminal(dir)
}

func OpenExternal(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}

	if erow.Info.IsSpecial() {
		return fmt.Errorf("can't run on special row")
	}

	return osutil.OpenExternal(erow.Info.Name())
}

//----------

func GoDebug(args *core.InternalCmdArgs) error {
	args2 := args.Part.ArgsUnquoted()

	// special case: show help
	cmd := godebug.NewCmd()
	buf := &bytes.Buffer{}
	cmd.Stderr = buf
	if err := cmd.ParseFlagsOnce(args2[1:]); errors.Is(err, flag.ErrHelp) {
		return fmt.Errorf("%w\n%v", err, buf.String())
	}

	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	return args.Ed.GoDebug.RunAsync(args.Ctx, erow, args2)
}

func GoDebugFind(args *core.InternalCmdArgs) error {
	// TODO: erow needed?
	//erow, err := args.ERowOrErr()
	//if err != nil {
	//	return err
	//}

	a := args.Part.ArgsUnquoted()
	if len(a) < 2 {
		return fmt.Errorf("missing string to find")
	}
	s := ""
	if len(a) == 2 {
		s = a[1] // single arg, unquoted if quoted
	} else {
		s = args.Part.FromArgString(1) // verbatim
	}

	return args.Ed.GoDebug.AnnotationFind(s)
}

func GoDebugTrace(args *core.InternalCmdArgs) error {
	return args.Ed.GoDebug.Trace()
}

//----------

func ColorTheme(args *core.InternalCmdArgs) error {
	ui.ColorThemeCycler.Cycle(args.Ed.UI.Root)
	args.Ed.UI.Root.MarkNeedsLayoutAndPaint()
	return nil
}

//----------

func FontRunes(args *core.InternalCmdArgs) error {
	var u string
	for i := 0; i < 15000; {
		start := i
		var w string
		for j := 0; j < 25; j++ {
			w += string(rune(i))
			i++
		}
		u += fmt.Sprintf("%d: %s\n", start, w)
	}
	args.Ed.Messagef("%s", u)
	return nil
}

//----------

func LSProtoCloseAll(args *core.InternalCmdArgs) error {
	man := args.Ed.LSProtoMan
	if man.NInstances() == 0 {
		return fmt.Errorf("no instances are running")
	}
	args.Ed.LSProtoMan.Stop()
	return nil
}
func CtxutilCallsState(args *core.InternalCmdArgs) error {
	s := ctxutil.CallsState()
	args.Ed.Messagef("%s", s)
	return nil
}
