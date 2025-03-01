package contentcmds

import (
	"context"
	"fmt"
	"strings"

	"github.com/friedelschoen/glake/core"
	"github.com/friedelschoen/glake/util/iout/iorw"
	"github.com/friedelschoen/glake/util/parseutil"
	"github.com/friedelschoen/glake/util/parseutil/reslocparser"
)

// Detects compilers output file format <string(:int)?(:int)?>, and goes to line/column.
func OpenFilename(ctx context.Context, erow *core.ERow, index int) (error, bool) {
	ta := erow.Row.TextArea
	var rd iorw.ReaderAt
	considerMiddle := false
	if a, b, ok := ta.Cursor().SelectionIndexes(); ok {
		// consider only the selection
		rd = iorw.NewLimitedReaderAt(ta.RW(), a, b)
	} else {
		considerMiddle = true
		// limit reading
		rd = iorw.NewLimitedReaderAtPad(ta.RW(), index, index, 1000)
	}

	rl, err := reslocparser.ParseResLoc2(rd, index)
	if err != nil {
		return err, false
	}
	filePos := reslocparser.ResLocToFilePos(rl)

	// consider middle path (index position) if pos is not present
	if considerMiddle && !filePos.HasPos() {
		k := index - rl.Pos
		if k <= 0 {
			// don't consider middle for these cases
			// k<0: index was before filename (fil^e:///a/b/c.txt)
			// k=0: index at filename start (empty string) (file://^/a/b/c.txt)
		} else {
			// index was beyond filename (/a/b/c.txt:1^:2)
			if k > len(filePos.Filename) {
				k = len(filePos.Filename)
			}
			// cut filename
			i := strings.Index(filePos.Filename[k:], string(rl.PathSep))
			if i >= 0 {
				filePos.Filename = filePos.Filename[:k+i]
			}
		}
	}

	// detected it's a filename, return true from here

	// remove escapes
	filePos.Filename = parseutil.RemoveFilenameEscapes(filePos.Filename, rl.Escape, rl.PathSep)

	// decode home vars
	filePos.Filename = erow.Ed.HomeVars.Decode(filePos.Filename)

	// find full filename
	filename, fi, ok := core.FindFileInfo(filePos.Filename, erow.Info.Dir())
	if !ok {
		err := fmt.Errorf("fileinfo not found: %q", filePos.Filename)
		return err, true
	}
	filePos.Filename = filename

	erow.Ed.UI.RunOnUIGoRoutine(func() {
		// place new under the calling row
		rowPos := erow.Row.PosBelow() // needs ui goroutine

		// if calling erow is dir, and new is not dir, place at a good place
		if erow.Info.IsDir() && !fi.IsDir() {
			rowPos = erow.Ed.GoodRowPos() // needs ui goroutine
		}

		conf := &core.OpenFileERowConfig{
			FilePos:               filePos,
			RowPos:                rowPos,
			FlashVisibleOffsets:   true,
			NewIfNotExistent:      true,
			NewIfOffsetNotVisible: true,
		}
		core.OpenFileERow(erow.Ed, conf) // needs ui goroutine
	})

	return nil, true
}
