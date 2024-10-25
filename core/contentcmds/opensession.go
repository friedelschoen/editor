package contentcmds

import (
	"context"
	"strings"
	"unicode"

	"github.com/jmigpin/editor/core"
	"github.com/jmigpin/editor/util/iout/iorw"
	"github.com/jmigpin/editor/util/parseutil/pscan"
)

func OpenSession(ctx context.Context, erow *core.ERow, index int) (error, bool) {
	ta := erow.Row.TextArea

	// limit reading
	rd := iorw.NewLimitedReaderAtPad(ta.RW(), index, index, 1000)

	sname, err := sessionName(rd, index)
	if err != nil {
		return nil, false
	}

	erow.Ed.UI.RunOnUIGoRoutine(func() {
		core.OpenSessionFromString(erow.Ed, sname)
	})

	return nil, true
}

//----------

func sessionName(rd iorw.ReaderAt, index int) (string, error) {
	sc := iorw.NewScanner(rd)
	//sc.SetSrc2(rd, index)

	parseName := sc.W.RuneFnLoop(sessionNameRune)
	cmdStr := "OpenSession"
	name := ""
	parseCmdAndName := sc.W.And(
		sc.W.Sequence(cmdStr),
		sc.M.SpacesExceptNewline,
		pscan.WKeep(&name, sc.W.StrValue(parseName)),
	)

	if p2, err := sc.M.Or(index,
		// index at: "●OpenSession● sessionname"
		sc.W.And(
			sc.W.ReverseMode(true, sc.W.Optional(sc.W.Or(
				sc.W.Sequence(cmdStr),
				sc.W.SequenceMid(cmdStr),
			))),
			parseCmdAndName,
		),
		// index at: "OpenSession ●sessionname●"
		sc.W.And(
			sc.W.ReverseMode(true, sc.W.And(
				sc.W.Sequence(cmdStr),
				sc.M.SpacesExceptNewline,
				sc.W.Optional(parseName),
			)),
			parseCmdAndName,
		),
	); err != nil {
		return "", sc.SrcError(p2, err)
	} else {
		return name, nil
	}
}

func sessionNameRune(ru rune) bool {
	return unicode.IsLetter(ru) ||
		unicode.IsDigit(ru) ||
		strings.ContainsRune("_-.", ru)
}
