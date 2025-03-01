package contentcmds

import (
	"context"
	"os"
	"time"

	"github.com/friedelschoen/glake/internal/core"
	"github.com/friedelschoen/glake/internal/io/iorw"
	"github.com/friedelschoen/glake/internal/lsproto"
	"github.com/friedelschoen/glake/internal/parser"
)

func GoToDefinitionLSProto(ctx context.Context, erow *core.ERow, index int) (error, bool) {
	if erow.Info.IsDir() {
		return nil, false
	}

	// timeout for the cmd to run
	timeout := 8 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ed := erow.Ed
	ta := erow.Row.TextArea
	rw := ta.RW()

	// must have a registration that handles the filename
	_, err := ed.LSProtoMan.LangManager(erow.Info.Name())
	if err != nil {
		return nil, false
	}

	filename, rang, err := ed.LSProtoMan.TextDocumentDefinition(ctx, erow.Info.Name(), rw, index)
	if err != nil {
		return err, true
	}

	// content reader
	var rd iorw.ReaderAt
	info, ok := ed.ERowInfo(filename)
	if ok {
		// file is in memory already
		if erow0, ok := info.FirstERow(); ok {
			rd = erow0.Row.TextArea.RW()
		}
	}
	if rd == nil {
		// read file
		b, err := os.ReadFile(filename)
		if err != nil {
			return err, true
		}
		rd = iorw.NewBytesReadWriterAt(b)
	}

	// translate range
	offset, length, err := lsproto.RangeToOffsetLen(rd, rang)
	if err != nil {
		return err, true
	}

	// build filepos
	filePos := &parser.FilePos{
		Filename: filename,
		Offset:   offset,
		Len:      length,
	}

	erow.Ed.UI.RunOnUIGoRoutine(func() {
		// place the file under the calling row
		rowPos := erow.Row.PosBelow() // needs ui goroutine

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
