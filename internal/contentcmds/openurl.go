package contentcmds

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"unicode"

	"github.com/friedelschoen/glake/internal/command"
	"github.com/friedelschoen/glake/internal/core"
	"github.com/friedelschoen/glake/internal/ioutil"
	"github.com/friedelschoen/glake/internal/parser"
)

// doesn't wait for the cmd to end
func openBrowser(url string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		c = exec.Command("open", url)
	default: // linux, others...
		c = exec.Command("xdg-open", url)
	}
	if err := c.Start(); err != nil {
		return err
	}
	go c.Wait() // async to let run, but wait to clear resources
	return nil
}

// Opens url lines in preferred application.
func OpenURL(ctx context.Context, erow *core.ERow, index int) (error, bool) {
	ta := erow.Row.TextArea

	//TODO: handle "//http://www" (detect "http" start?)

	isHttpRune := func(ru rune) bool {
		extra := parser.RunesExcept(parser.ExtraRunes, " []()<>")
		return unicode.IsLetter(ru) || unicode.IsDigit(ru) ||
			strings.ContainsRune(extra, ru)
	}

	rd := ioutil.NewLimitedReaderAtPad(ta.RW(), index, index, 1000)
	l, r := parser.ExpandIndexesEscape(rd, index, false, isHttpRune, command.EscapeCharacter())

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
	if err := openBrowser(ustr); err != nil {
		return err, true
	}

	return nil, true
}
