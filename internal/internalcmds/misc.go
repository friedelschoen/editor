package internalcmds

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/friedelschoen/editor/internal/context"
	"github.com/friedelschoen/editor/internal/core"
	"github.com/friedelschoen/editor/internal/multierror"
	"github.com/friedelschoen/editor/internal/ui"
)

func Version(args *core.InternalCmdArgs) error {
	args.Ed.Messagef("version: %v", core.Version())
	return nil
}

func Exit(args *core.InternalCmdArgs) error {
	args.Ed.Close()
	return nil
}

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

func Save(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	return erow.Info.SaveFile()
}
func SaveAllFiles(args *core.InternalCmdArgs) error {
	var me multierror.MultiError
	for _, info := range args.Ed.ERowInfos() {
		if info.IsFileButNotDir() {
			me.Add(info.SaveFile())
		}
	}
	return me.Result()
}

func Reload(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	return erow.Reload()
}
func ReloadAllFiles(args *core.InternalCmdArgs) error {
	me := &multierror.MultiError{}
	for _, info := range args.Ed.ERowInfos() {
		if info.IsFileButNotDir() {
			me.Add(info.ReloadFile())
		}
	}
	return me.Result()
}
func ReloadAll(args *core.InternalCmdArgs) error {
	// reload all dirs erows
	me := &multierror.MultiError{}
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

func Stop(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	erow.Exec.Stop()
	return nil
}

func Clear(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}
	erow.Row.TextArea.SetStrClearHistory("")
	return nil
}

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

	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		c = exec.Command("explorer", "/select,"+dir)
	case "darwin":
		c = exec.Command("open", dir)
	default: // linux, others...
		c = exec.Command("xdg-open", dir)
	}
	if err := c.Start(); err != nil {
		return err
	}
	go c.Wait() // async to let run, but wait to clear resources
	return nil
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

	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		return errors.New("todo")
	case "darwin":
		// TODO: review
		c = exec.Command("terminal", dir)
	default: // linux, others...
		c = exec.Command("x-terminal-emulator", "--working-directory="+dir)
	}
	if err := c.Start(); err != nil {
		return err
	}
	go c.Wait() // async to let run, but wait to clear resources
	return nil
}

func OpenExternal(args *core.InternalCmdArgs) error {
	erow, err := args.ERowOrErr()
	if err != nil {
		return err
	}

	if erow.Info.IsSpecial() {
		return fmt.Errorf("can't run on special row")
	}

	name := erow.Info.Name()
	c := (*exec.Cmd)(nil)
	switch runtime.GOOS {
	case "windows":
		// TODO: review
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", name)
	case "darwin":
		c = exec.Command("open", name)
	default: // linux, others...
		c = exec.Command("xdg-open", name)
	}
	if err := c.Start(); err != nil {
		return err
	}
	go c.Wait() // async to let run, but wait to clear resources
	return nil
}

func ColorTheme(args *core.InternalCmdArgs) error {
	name := args.Part.Args[0].String()
	ui.SetColorscheme(name, args.Ed.UI.Root)
	args.Ed.UI.Root.MarkNeedsLayoutAndPaint()
	return nil
}

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

func LSProtoCloseAll(args *core.InternalCmdArgs) error {
	return args.Ed.LSProtoMan.Close()
}
func CtxutilCallsState(args *core.InternalCmdArgs) error {
	s := context.CallsState()
	args.Ed.Messagef("%s", s)
	return nil
}
