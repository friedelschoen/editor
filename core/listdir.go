package core

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/friedelschoen/glake/util/parseutil"
)

func ListDirERow(erow *ERow, filepath string, subs, hiddens bool) {
	erow.Exec.RunAsync(func(ctx context.Context, rw io.ReadWriter) error {
		return ListDirContext(ctx, rw, erow.Info.Name(), subs, hiddens)
	})
}

func ListDirContext(ctx context.Context, w io.Writer, filepath string, subs, hiddens bool) error {
	// "../" at the top
	u := ".." + string(os.PathSeparator)
	if _, err := w.Write([]byte(u + "\n")); err != nil {
		return err
	}

	return listDirContext(ctx, w, filepath, "", subs, hiddens)
}

func listDirContext(ctx context.Context, w io.Writer, fpath, addedFilepath string, subs, hiddens bool) error {
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

	slices.SortFunc(fis, CompareFileInfos)

	for _, fi := range fis {
		// stop on context
		if ctx.Err() != nil {
			return ctx.Err()
		}

		name := fi.Name()

		if !hiddens && strings.HasPrefix(name, ".") {
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

		if fi.IsDir() && subs {
			afp := filepath.Join(addedFilepath, name)
			err := listDirContext(ctx, w, fpath, afp, subs, hiddens)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CompareFileInfos(a, b os.FileInfo) int {
	an := strings.ToLower(a.Name())
	bn := strings.ToLower(b.Name())

	cmp := func() int {
		v := strings.Compare(an, bn)
		if v == 0 {
			return strings.Compare(a.Name(), b.Name())
		}
		return v
	}

	if a.IsDir() {
		if b.IsDir() {
			return cmp()
		}
		return -1
	}
	if b.IsDir() {
		return 1
	}
	return cmp()
}
