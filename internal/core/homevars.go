package core

import (
	"os"

	"github.com/friedelschoen/glake/internal/core/toolbarparser"
)

type HomeVars struct {
	hvm *toolbarparser.HomeVarMap
}

func NewHomeVars() *HomeVars {
	return &HomeVars{}
}

func (hv *HomeVars) ParseToolbarVars(strs []string, caseInsensitive bool) {
	// merge strings maps
	m := toolbarparser.VarMap{}
	for _, str := range strs {
		data := toolbarparser.Parse(str)
		m2 := toolbarparser.ParseVars(data)
		// merge
		for k, v := range m2 {
			m[k] = v
		}
	}
	// add env home var at the end to enforce value
	h, err := os.UserHomeDir()
	if err != nil {
		m["~"] = h
	}

	hv.hvm = toolbarparser.NewHomeVarMap(m, caseInsensitive)
}

func (hv *HomeVars) Encode(filename string) string {
	return hv.hvm.Encode(filename)
}

func (hv *HomeVars) Decode(filename string) string {
	return hv.hvm.Decode(filename)
}
