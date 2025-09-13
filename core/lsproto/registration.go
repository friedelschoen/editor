package lsproto

import (
	"fmt"
	"strings"

	"github.com/jmigpin/editor/util/parseutil"
)

type Registration struct {
	Language string   `json:"language"`
	Exts     []string `json:"extensions"`
	Network  string   `json:"transport"` // {stdio,tcpclient,tcp}
	Cmd      string   `json:"command"`   // template values: {.Addr,.Host,.Port}
	Optional []string `json:"flags"`     // {stderr,nogotoimpl}
}

func (reg *Registration) HasOptional(s string) bool {
	for _, v := range reg.Optional {
		if v == s {
			return true
		}
	}

	// handle backwards compatibility (keep old defaults)
	if s == "nogotoimpl" {
		// - these won't use gotoimplementation, and there is no way to enable it (it would just be slower)
		// - other languages (ex:c/c++) will use gotoimplementation
		languagesToBypass := "go python javascript"
		if strings.Contains(languagesToBypass, strings.ToLower(reg.Language)) {
			return true
		}
	}

	return false
}

func (reg *Registration) String() string {
	exts := strings.Join(reg.Exts, " ")
	if len(reg.Exts) >= 2 {
		exts = fmt.Sprintf("%q", exts)
	}

	cmd := reg.Cmd
	cmd2 := parseutil.AddEscapes(cmd, '\\', " ,")
	if cmd != cmd2 {
		cmd = fmt.Sprintf("%q", cmd)
	}

	u := []string{
		reg.Language,
		exts,
		reg.Network,
		cmd,
	}
	if len(reg.Optional) >= 1 {
		h := strings.Join(reg.Optional, " ")
		if len(reg.Optional) >= 2 {
			h = fmt.Sprintf("%q", h)
		}
		u = append(u, h)
	}
	return strings.Join(u, ",")
}
