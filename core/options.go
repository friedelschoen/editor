package core

import (
	"fmt"
	"strings"

	"github.com/jmigpin/editor/core/lsproto"
	"github.com/jmigpin/editor/util/parseutil"
)

type Options struct {
	Font string `json:"font"`

	TabWidth           int    `json:"tabwidth"`
	WrapLineRune       string `json:"wrapline-rune"`
	CarriageReturnRune string `json:"cr-rune"`

	ColorTheme     string `json:"colortheme"`
	CommentsColor  int    `json:"comment-color"`
	StringsColor   int    `json:"string-color"`
	ScrollBarWidth int    `json:"scrollbar-width"`
	ScrollBarLeft  bool   `json:"scrollbar-left"`
	Shadows        bool   `json:"shadows"`

	SessionName string
	Filenames   []string

	UseMultiKey bool `json:"multikey"`

	Plugins []string `json:"plugins"`

	LSProtos     []lsproto.Registration
	PreSaveHooks []PreSaveHook

	ZipSessionsFile bool
}

//----------

type PreSaveHook struct {
	Language string   `json:"language"`
	Exts     []string `json:"extensions"`
	Cmd      string   `json:"command"`
}

func (h *PreSaveHook) String() string {
	u := []string{h.Language}

	exts := strings.Join(h.Exts, " ")
	if len(h.Exts) >= 2 {
		exts = fmt.Sprintf("%q", exts)
	}
	u = append(u, exts)

	cmd := h.Cmd
	cmd2 := parseutil.AddEscapes(cmd, '\\', " ,")
	if cmd != cmd2 {
		cmd = fmt.Sprintf("%q", cmd)
	}
	u = append(u, cmd)

	return strings.Join(u, ",")
}
