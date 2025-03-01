package testutil

import (
	"testing"
)

// based on txtar (txt archive)
type Script struct {
	ScriptsDir     string
	Args           []string
	Cmds           []*ScriptCmd // user cmds (provided)
	Work           bool         // don't remove work dir at end
	NoFilepathsFix bool         // don't rewrite filepaths for current dir

	ScriptStart func(*testing.T) error // each script init
	ScriptStop  func(*testing.T) error // each script close

	ucmds map[string]*ScriptCmd // user cmds (mapped)
	icmds map[string]*ScriptCmd // internal cmds

	workDir string
	lastCmd struct {
		stdout []byte
		stderr []byte
		err    []byte
	}
}

type ScriptCmd struct {
	Name string
	Fn   func(t *testing.T, args []string) error
}
