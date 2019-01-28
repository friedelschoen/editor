package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmigpin/editor/core/parseutil"
	"github.com/jmigpin/editor/core/toolbarparser"
)

func ListDirCmd(erow *ERow, part *toolbarparser.Part) error {
	if !erow.Info.IsDir() {
		return fmt.Errorf("not a directory")
	}

	tree, hidden := false, false

	args := part.Args[1:]
	for _, a := range args {
		s := a.UnquotedStr()
		switch s {
		case "-sub":
			tree = true
		case "-hidden":
			hidden = true
		}
	}

	ListDirERow(erow, erow.Info.Name(), tree, hidden)

	return nil
}

//----------

func ListDirERow(erow *ERow, filepath string, tree, hidden bool) {
	// clear
	erow.Row.TextArea.SetStrClearHistory("")
	erow.Row.TextArea.ClearPos()

	erow.Exec.Start(func(ctx context.Context, w io.Writer) error {
		return ListDirContext(ctx, w, erow.Info.Name(), tree, hidden)
	})
}

//----------

func ListDirContext(ctx context.Context, w io.Writer, filepath string, tree, hidden bool) error {
	// "../" at the top
	u := ".." + string(os.PathSeparator)
	if _, err := w.Write([]byte(u + "\n")); err != nil {
		return err
	}

	return listDirContext(ctx, w, filepath, "", tree, hidden)
}

func listDirContext(ctx context.Context, w io.Writer, fpath, addedFilepath string, tree, hidden bool) error {
	fp2 := filepath.Join(fpath, addedFilepath)

	out := func(s string) bool {
		_, err := w.Write([]byte(s))
		return err == nil
	}

	f, err := os.Open(fp2)
	if err != nil {
		out(err.Error())
		return nil
	}

	fis, err := f.Readdir(-1)
	f.Close() // close as soon as possible
	if err != nil {
		out(err.Error())
		return nil
	}

	// stop on context
	if ctx.Err() != nil {
		return ctx.Err()
	}

	sort.Sort(ByListOrder(fis))

	for _, fi := range fis {
		// stop on context
		if ctx.Err() != nil {
			return ctx.Err()
		}

		name := fi.Name()

		if !hidden && strings.HasPrefix(name, ".") {
			continue
		}

		name2 := filepath.Join(addedFilepath, name)
		if fi.IsDir() {
			name2 += string(os.PathSeparator)
		}
		s := parseutil.EscapeFilename(name2) + "\n"
		if !out(s) {
			return nil
		}

		if fi.IsDir() && tree {
			afp := filepath.Join(addedFilepath, name)
			err := listDirContext(ctx, w, fpath, afp, tree, hidden)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//----------

type ByListOrder []os.FileInfo

func (a ByListOrder) Len() int {
	return len(a)
}
func (a ByListOrder) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByListOrder) Less(i, j int) bool {
	ei := a[i]
	ej := a[j]
	iname := strings.ToLower(ei.Name())
	jname := strings.ToLower(ej.Name())
	if ei.IsDir() && ej.IsDir() {
		return iname < jname
	}
	if ei.IsDir() {
		return true
	}
	if ej.IsDir() {
		return false
	}
	return iname < jname
}
