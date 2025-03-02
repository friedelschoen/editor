package parser

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/friedelschoen/glake/internal/command"
	"github.com/friedelschoen/glake/internal/ioutil"
	"github.com/friedelschoen/glake/internal/parser/pscan"
)

// TODO: review

var ExtraRunes = "_-~.%@&?!=#+:^" + "(){}[]<>" + "\\/" + " "

var excludeResourceRunes = "" +
	" " + // word separator
	"=" + // usually around filenames (ex: -arg=/a/b.txt)
	"(){}[]<>" // usually used around filenames in various outputs
// escaped when outputing filenames
var escapedInFilenames = excludeResourceRunes +
	":" // note: in windows will give "C^:/"

func AddEscapes(str string, escape rune, escapeRunes string) string {
	w := []rune{}
	er := []rune(escapeRunes)
	for _, ru := range str {
		if ContainsRune(er, ru) {
			w = append(w, escape)
		}
		w = append(w, ru)
	}
	return string(w)
}

func RemoveEscapes(str string, escape rune) string {
	return RemoveEscapesEscapable(str, escape, "")
}
func RemoveEscapesEscapable(str string, escape rune, escapable string) string {
	return string(RemoveEscapes2([]rune(str), []rune(escapable), escape))
}
func RemoveEscapes2(rs []rune, escapable []rune, escape rune) []rune {
	res := make([]rune, 0, len(rs))
	escaping := false
	for i := 0; i < len(rs); i++ {
		ru := rs[i]
		if !escaping {
			if ru == escape { // remove escapes
				escaping = true
				continue
			}
		} else {
			escaping = false

			// re-add escape if not one of the escapable
			if len(escapable) > 0 {
				if !ContainsRune(escapable, ru) {
					res = append(res, escape)
				}
			}
		}
		res = append(res, ru)
	}
	return res
}

func EscapeFilename(str string) string {
	mustBeEscaped := escapedInFilenames + "^"
	return AddEscapes(str, command.EscapeCharacter(), mustBeEscaped)
}

func RemoveFilenameEscapes(f string, escape, pathSep rune) string {
	f = RemoveEscapes(f, escape)
	f = CleanMultiplePathSeps(f, pathSep)
	if u, err := url.PathUnescape(f); err == nil {
		f = u
	}
	return f
}

func CleanMultiplePathSeps(str string, sep rune) string {
	w := []rune{}
	added := false
	for _, ru := range str {
		if ru == sep {
			if !added {
				added = true
				w = append(w, ru)
			}
		} else {
			added = false
			w = append(w, ru)
		}
	}
	return string(w)
}

func ExpandIndexesEscape(rd ioutil.ReaderAt, index int, truth bool, fn func(rune) bool, escape rune) (int, int) {
	// ensure the index is not in the middle of an escape
	index = ImproveExpandIndexEscape(rd, index, escape)

	l := ExpandLastIndexEscape(rd, index, false, fn, escape)
	r := ExpandIndexEscape(rd, index, false, fn, escape)
	return l, r
}

func ExpandIndexEscape(r ioutil.ReaderAt, i int, truth bool, fn func(rune) bool, escape rune) int {
	sc := ioutil.NewScanner(r)
	return expandEscape(sc, i, truth, fn, escape)
}

func ExpandLastIndexEscape(r ioutil.ReaderAt, i int, truth bool, fn func(rune) bool, escape rune) int {
	sc := ioutil.NewScanner(r)

	// read direction
	tmp := sc.Reverse
	sc.Reverse = true
	defer func() { sc.Reverse = tmp }() // restore

	return expandEscape(sc, i, truth, fn, escape)
}

func expandEscape(sc *pscan.Scanner, i int, truth bool, fn func(rune) bool, escape rune) int {
	p2, _ := sc.M.Loop(i, sc.W.Or(
		sc.W.EscapeAny(escape),
		sc.W.RuneFn(func(ru rune) bool {
			return fn(ru) != truth
		}),
	))
	return p2
}

func ImproveExpandIndexEscape(r ioutil.ReaderAt, i int, escape rune) int {
	sc := ioutil.NewScanner(r)
	p2, _ := sc.M.ReverseMode(i, true,
		sc.W.Loop(sc.W.Rune(escape)),
	)
	return p2
}

// Line/col args are one-based.
func LineColumnIndex(rd ioutil.ReaderAt, line, column int) (int, error) {
	// must have a good line
	if line <= 0 {
		return 0, fmt.Errorf("bad line: %v", line)
	}
	line-- // make line 0 the first line

	// tolerate bad columns
	if column <= 0 {
		column = 1
	}
	column-- // make column 0 the first column

	index := -1
	l, lStart := 0, 0
	ri := 0
	for {
		if l == line {
			index = lStart // keep line start in case it is a bad col

			c := ri - lStart
			if c >= column {
				index = ri // keep line/col
				break
			}
		} else if l > line {
			break
		}

		ru, size, err := ioutil.ReadRuneAt(rd, ri)
		if err != nil {
			// be tolerant about the column
			if index >= 0 {
				return index, nil
			}
			return 0, err
		}
		ri += size
		if ru == '\n' {
			l++
			lStart = ri
		}
	}
	if index < 0 {
		return 0, fmt.Errorf("line not found: %v", line)
	}
	return index, nil
}

// Returned line/col values are one-based.
func IndexLineColumn(rd ioutil.ReaderAt, index int) (int, int, error) {
	line, lineStart := 0, 0
	ri := 0
	for ri < index {
		ru, size, err := ioutil.ReadRuneAt(rd, ri)
		if err != nil {
			return 0, 0, err
		}
		ri += size
		if ru == '\n' {
			line++
			lineStart = ri
		}
	}
	line++                    // first line is 1
	col := ri - lineStart + 1 // first column is 1
	return line, col, nil
}

// Returned line/col values are one-based.

func DetectEnvVar(str, name string) bool {
	vstr := "$" + name
	i := strings.Index(str, vstr)
	if i < 0 {
		return false
	}

	e := i + len(vstr)
	if e > len(str) {
		return false
	}

	// validate rune after the name
	ru, _ := utf8.DecodeRuneInString(str[e:])
	if ru != utf8.RuneError {
		if unicode.IsLetter(ru) || unicode.IsDigit(ru) || ru == '_' {
			return false
		}
	}

	return true
}

func RunesExcept(runes, except string) string {
	drop := func(ru rune) rune {
		if strings.ContainsRune(except, ru) {
			return -1
		}
		return ru
	}
	return strings.Map(drop, runes)
}

// Useful to compare src code lines.

func UrlToAbsFilename(url2 string) (string, error) {
	u, err := url.Parse(string(url2))
	if err != nil {
		return "", err
	}
	if u.Scheme != "file" {
		return "", fmt.Errorf("expecting file scheme: %v", u.Scheme)
	}
	filename := u.Path // unescaped

	if runtime.GOOS == "windows" {
		// remove leading slash in windows returned by url.parse: https://github.com/golang/go/issues/6027
		if len(filename) > 0 && filename[0] == '/' {
			filename = filename[1:]
		}

		filename = filepath.FromSlash(filename)
	}

	if !filepath.IsAbs(filename) {
		return "", fmt.Errorf("filename not absolute: %v", filename)
	}
	return filename, nil
}

func AbsFilenameToUrl(filename string) (string, error) {
	if !filepath.IsAbs(filename) {
		return "", fmt.Errorf("filename not absolute: %v", filename)
	}

	if runtime.GOOS == "windows" {
		filename = filepath.ToSlash(filename)
		// add leading slash to match UrlToAbsFilename behaviour
		if len(filename) > 0 && filename[0] != '/' {
			filename = "/" + filename
		}
	}

	u := &url.URL{Scheme: "file", Path: filename}
	return u.String(), nil // path is escaped
}

// unquote string with backslash as escape
func UnquoteStringBs(s string) (string, error) {
	return UnquoteString(s, '\\')
}

// removes escapes runes (keeping the escaped) if quoted
func UnquoteString(s string, esc rune) (string, error) {
	rs := []rune(s)
	_, err := RunesQuote(rs)
	if err != nil {
		return "", err
	}
	u := RemoveEscapes2(rs[1:len(rs)-1], nil, esc)
	return string(u), nil
}
func RunesQuote(rs []rune) (rune, error) {
	if len(rs) < 2 {
		return 0, fmt.Errorf("len<2")
	}
	quotes := []rune("\"'`") // allowed quotes
	quote := rs[0]
	if !ContainsRune(quotes, quote) {
		return 0, fmt.Errorf("unexpected starting quote: %q", quote)
	}
	if rs[len(rs)-1] != quote {
		return 0, fmt.Errorf("missing ending quote: %q", quote)
	}
	return quote, nil
}
func IsQuoted(s string) bool {
	_, err := RunesQuote([]rune(s))
	return err == nil
}

func ContainsRune(rs []rune, ru rune) bool {
	for _, ru2 := range rs {
		if ru2 == ru {
			return true
		}
	}
	return false
}
