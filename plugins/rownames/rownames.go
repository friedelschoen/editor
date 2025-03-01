package main

import (
	"sort"
	"strings"

	"github.com/friedelschoen/glake/core"
	"github.com/friedelschoen/glake/core/toolbarparser"
)

func ToolbarCmd(ed *core.Editor, erow *core.ERow, part *toolbarparser.Part) bool {
	arg0 := part.Args[0].UnquotedString()
	switch arg0 {
	case "RowNames":
		rowNames(ed)
		return true
	default:
		return false
	}
}

func rowNames(ed *core.Editor) {
	u := []string{}
	for _, info := range ed.ERowInfos() {
		u = append(u, info.Name())
	}
	sort.Strings(u)
	msg := "rownames:\n\t" + strings.Join(u, "\n\t")
	ed.Messagef(msg)
}
