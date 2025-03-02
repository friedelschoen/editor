package reslocparser

import (
	"os"
	"runtime"
	"sync"

	"github.com/friedelschoen/glake/internal/command"
	"github.com/friedelschoen/glake/internal/ioutil"
	"github.com/friedelschoen/glake/internal/parser"
)

func ParseResLoc(src []byte, index int) (*ResLoc, error) {
	rlp, err := getResLocParser()
	if err != nil {
		return nil, err
	}
	return rlp.Parse(src, index)
}
func ParseResLoc2(rd ioutil.ReaderAt, index int) (*ResLoc, error) {
	src, err := ioutil.ReadFastFull(rd)
	if err != nil {
		return nil, err
	}
	min := rd.Min() // keep to restore position
	rl, err := ParseResLoc(src, index-min)
	if err != nil {
		return nil, err
	}
	// restore position
	rl.Pos += min
	rl.End += min
	return rl, nil
}

// reslocparser singleton
var rlps struct {
	once sync.Once
	p    *ResLocParser
	err  error
}

func getResLocParser() (*ResLocParser, error) {
	rlps.once.Do(func() {
		rlps.p, rlps.err = newResLocParserSingletonInstance()
	})
	return rlps.p, rlps.err
}
func newResLocParserSingletonInstance() (*ResLocParser, error) {
	rlp := NewResLocParser()

	rlp.PathSeparator = rune(os.PathSeparator)
	rlp.Escape = rune(command.EscapeCharacter())
	rlp.ParseVolume = runtime.GOOS == "windows"

	rlp.Init()

	return rlp, nil
}

// util func to replace parseutil.*

func ResLocToFilePos(rl *ResLoc) *parser.FilePos {
	return &parser.FilePos{
		Filename: rl.Path, // original string (unescaped)
		Line:     rl.Line,
		Column:   rl.Column,
		Offset:   rl.Offset,
		Len:      0,
	}
}
