package parseutil

import (
	"github.com/jmigpin/editor/util/iout/iorw"
)

type FilePos struct {
	Filename     string
	Offset, Len  int // length after offset for a range
	Line, Column int // bigger than zero to be considered
}

func (fp *FilePos) HasOffset() bool {
	return fp.Line == 0
}

//----------

// Parse fmt: <filename:line?:col?>. Accepts escapes but doesn't unescape.
func ParseFilePos(str string) (*FilePos, error) {
	rw := iorw.NewBytesReadWriter([]byte(str))
	res, err := ParseResource(rw, 0)
	if err != nil {
		return nil, err
	}
	return NewFilePosFromResource(res), nil
}

func NewFilePosFromResource(res *Resource) *FilePos {
	return &FilePos{
		Offset:   -1,
		Filename: res.RawPath, // original string (unescaped)
		Line:     res.Line,
		Column:   res.Column,
	}
}
