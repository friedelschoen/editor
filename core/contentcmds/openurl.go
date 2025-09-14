package contentcmds

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/friedelschoen/editor/core"
	"github.com/friedelschoen/editor/util/iout/iorw"
	"github.com/friedelschoen/editor/util/osutil"
	"github.com/friedelschoen/editor/util/parseutil"
)

// Opens url lines in preferred application.
func OpenURL(ctx context.Context, erow *core.ERow, index int) (error, bool) {
	ta := erow.Row.TextArea

	//TODO: handle "//http://www" (detect "http" start?)

	isHttpRune := func(ru rune) bool {
		extra := parseutil.RunesExcept(parseutil.ExtraRunes, " []()<>")
		return unicode.IsLetter(ru) || unicode.IsDigit(ru) ||
			strings.ContainsRune(extra, ru)
	}

	rd := iorw.NewLimitedReaderAtPad(ta.RW(), index, index, 1000)
	l, r := parseutil.ExpandIndexesEscape(rd, index, false, isHttpRune, osutil.EscapeRune)

	b, err := rd.ReadFastAt(l, r-l)
	if err != nil {
		return err, false // not handled
	}
	str := string(b)

	u, err := url.Parse(str)
	if err != nil {
		return err, false
	}

	// accepted schemes
	switch u.Scheme {
	case "http", "https", "ftp", "mailto":
		// ok
	default:
		err := fmt.Errorf("unsupported scheme: %v", u.Scheme)
		return err, false
	}

	ustr := u.String()
	if err := osutil.OpenBrowser(ustr); err != nil {
		return err, true
	}

	return nil, true
}
