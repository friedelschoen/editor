package lsproto

//godebug:annotatepackage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/friedelschoen/editor/util/iout"
	"github.com/friedelschoen/editor/util/iout/iorw"
	"github.com/friedelschoen/editor/util/osutil"
	"github.com/friedelschoen/editor/util/testutil"
)

func TestStruct1(t *testing.T) {
	{
		msg := `"abc"`
		doc := _completionItemDocumentation{}
		if err := json.Unmarshal([]byte(msg), &doc); err != nil {
			t.Fatal(err)
		}
		if doc.str == nil || *doc.str != "abc" {
			fmt.Printf("%+v\n", doc)
			t.Fail()
		}
	}
	{
		msg := `{"kind":"markup","value":"abc"}`
		doc := _completionItemDocumentation{}
		if err := json.Unmarshal([]byte(msg), &doc); err != nil {
			t.Fatal(err)
		}
		if doc.mc == nil || doc.mc.Value != "abc" {
			fmt.Printf("%+v\n", doc)
			t.Fail()
		}
	}

}

//----------
//----------
//----------

func TestScripts(t *testing.T) {
	log.SetFlags(0)
	//log.SetPrefix("lsptester: ")

	scr := testutil.NewScript(os.Args)
	scr.ScriptsDir = "testdata"
	//scr.Parallel = true // TODO: failing
	//scr.Work = true

	man := (*Manager)(nil)
	scr.ScriptStart = func(t *testing.T) error {
		man = newTestManager(t)
		return nil
	}
	scr.ScriptStop = func(t *testing.T) error {
		man.Stop()
		return nil
	}

	scr.Cmds = []*testutil.ScriptCmd{
		{"lspSourceCursor", func(st *testutil.ST, args []string) error {
			return lspSourceCursor(st, args, man)
		}},
		{"lspDefinition", func(st *testutil.ST, args []string) error {
			return lspDefinition(st, args, man)
		}},
		{"lspCompletion", func(st *testutil.ST, args []string) error {
			return lspCompletion(st, args, man)
		}},
		{"lspRename", func(st *testutil.ST, args []string) error {
			return lspRename(st, args, man)
		}},
		{"lspReferences", func(st *testutil.ST, args []string) error {
			return lspReferences(st, args, man)
		}},
		{"lspCallHierarchy", func(st *testutil.ST, args []string) error {
			return lspCallHierarchy(st, args, man)
		}},
	}

	scr.Run(t)
}

//----------

func lspSourceCursor(st *testutil.ST, args []string, man *Manager) error {
	args = args[1:] // remove cmd string
	if len(args) != 3 {
		return fmt.Errorf("sourcecursor: expecting 3 args: %v", args)
	}

	template := st.DirJoin(args[0])
	filename := st.DirJoin(args[1])
	mark := args[2]

	mark2, err := strconv.ParseInt(mark, 10, 32)
	if err != nil {
		return err
	}

	// read template
	b, err := os.ReadFile(template)
	if err != nil {
		return err
	}
	offset, src := sourceCursor(st.T, string(b), int(mark2))

	// write filename
	if err := os.WriteFile(filename, []byte(src), 0o644); err != nil {
		return err
	}

	st.Printf("%d", offset)

	return nil
}

//----------

func lspDefinition(st *testutil.ST, args []string, man *Manager) error {
	args = args[1:] // remove cmd string
	if len(args) != 2 {
		return fmt.Errorf("rename: expecting 2 args: %v", args)
	}

	filename := st.DirJoin(args[0])
	offset := args[1]

	// read offset (allow offset from env var)
	offset2, err := getIntArgPossiblyFromEnv(st, offset)
	if err != nil {
		return err
	}

	// read filename
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	rd := iorw.NewStringReaderAt(string(b))

	// full filename
	filename2, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	ctx := context.Background()
	f, rang, err := man.TextDocumentDefinition(ctx, filename2, rd, offset2)
	if err != nil {
		return err
	}
	st.Printf("%v %v", f, rang)
	return nil
}

//----------

func lspCompletion(st *testutil.ST, args []string, man *Manager) error {
	args = args[1:] // remove cmd string
	if len(args) != 2 {
		return fmt.Errorf("rename: expecting 2 args: %v", args)
	}

	filename := st.DirJoin(args[0])
	offset := args[1]

	// read offset (allow offset from env var)
	offset2, err := getIntArgPossiblyFromEnv(st, offset)
	if err != nil {
		return err
	}

	// read filename
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	rd := iorw.NewStringReaderAt(string(b))

	// full filename
	filename2, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	ctx := context.Background()
	clist, err := man.TextDocumentCompletion(ctx, filename2, rd, offset2)
	if err != nil {
		return err
	}
	w := CompletionListToString(clist)
	st.Printf("%v", w)
	return nil
}

//----------

func lspRename(st *testutil.ST, args []string, man *Manager) error {
	args = args[1:] // remove cmd string
	if len(args) != 3 {
		return fmt.Errorf("rename: expecting 3 args: %v", args)
	}

	filename := st.DirJoin(args[0])
	offsetVar := args[1]
	newName := args[2]

	// read offset (allow offset from env var)
	offset2, err := getIntArgPossiblyFromEnv(st, offsetVar)
	if err != nil {
		return err
	}

	// read filename
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	rd := iorw.NewStringReaderAt(string(b))

	// full filename
	filename2, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	ctx := context.Background()
	wecs, err := man.TextDocumentRenameAndPatch(ctx, filename2, rd, offset2, newName, nil)
	if err != nil {
		return err
	}
	for _, wec := range wecs {
		b, err := os.ReadFile(wec.Filename)
		if err != nil {
			return err
		}
		st.Printf("filename: %v\n", wec.Filename)
		st.Printf("%s\n", b)
	}

	return nil
}

//----------

func lspReferences(st *testutil.ST, args []string, man *Manager) error {
	args = args[1:] // remove cmd string
	if len(args) != 2 {
		return fmt.Errorf("rename: expecting 2 args: %v", args)
	}

	filename := st.DirJoin(args[0])
	offsetVar := args[1]

	// read offset (allow offset from env var)
	offset2, err := getIntArgPossiblyFromEnv(st, offsetVar)
	if err != nil {
		return err
	}

	// read filename
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	rd := iorw.NewStringReaderAt(string(b))

	// full filename
	filename2, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	ctx := context.Background()
	locs, err := man.TextDocumentReferences(ctx, filename2, rd, offset2)
	if err != nil {
		return err
	}

	str, err := LocationsToString(locs, "")
	if err != nil {
		return err
	}
	st.Printf("%v", str)

	return nil
}

//----------

func lspCallHierarchy(st *testutil.ST, args []string, man *Manager) error {
	args = args[1:] // remove cmd string
	if len(args) != 2 {
		return fmt.Errorf("rename: expecting 2 args: %v", args)
	}

	filename := st.DirJoin(args[0])
	offsetVar := args[1]

	// read offset (allow offset from env var)
	offset2, err := getIntArgPossiblyFromEnv(st, offsetVar)
	if err != nil {
		return err
	}

	// read filename
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	rd := iorw.NewStringReaderAt(string(b))

	// full filename
	filename2, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	ctx := context.Background()
	mcalls, err := man.CallHierarchyCalls(ctx, filename2, rd, offset2, IncomingChct)
	if err != nil {
		return err
	}
	str, err := ManagerCallHierarchyCallsToString(mcalls, IncomingChct, "")
	if err != nil {
		return err
	}
	st.Printf("result: %v", str)

	return nil
}

//----------
//----------
//----------

func goplsRegistration(tcp bool, trace bool, stderr bool) Registration {
	cmd := osutil.ExecName("gopls")
	if trace {
		cmd += " -v"
	}
	cmd += " serve"
	if trace {
		cmd += " -rpc.trace"
	}
	net := "stdio"
	if tcp {
		net = "tcp"
		cmd += " -listen={{.Addr}}"
	}

	var errOut []string
	if stderr {
		errOut = append(errOut, "stderr")
		//errOut = ",stderrmanmsg" // DEBUG
	}

	return Registration{
		Language: "go",
		Exts:     []string{".go"},
		Network:  net,
		Cmd:      cmd,
		Optional: errOut,
	}
	// return fmt.Sprintf("go,.go,%v,%q%s", net, cmd, errOut)
}

func cLangRegistration(alternateExe string, stderr bool) Registration {
	ext := []string{".c", ".h", ".cpp", ".hpp", ".cc"}
	exe := "clangd"
	if alternateExe != "" {
		exe = alternateExe
	}
	cmd := osutil.ExecName(exe)
	var errOut []string
	if stderr {
		errOut = append(errOut, "stderr")
	}
	return Registration{
		Language: "cpp",
		Exts:     ext,
		Network:  "stdio",
		Cmd:      cmd,
		Optional: errOut,
	}
}

func pylspRegistration(tcp bool, stderr bool) Registration {
	cmd := osutil.ExecName("pylsp")
	net := "stdio"
	if tcp {
		net = "tcp"
		cmd += " --tcp"
		cmd += " --host={{.Host}}"
		cmd += " --port={{.Port}}"
	}
	var errOut []string
	if stderr {
		errOut = append(errOut, "stderr")
	}
	return Registration{
		Language: "python",
		Exts:     []string{".py"},
		Network:  net,
		Cmd:      cmd,
		Optional: errOut,
	}
}

func newTestManager(t *testing.T) *Manager {
	t.Helper()

	msgFn := func(s string) {
		t.Helper()
		// can't use t.Log if already out of the test
		logPrintf("manager async msg: %v", s)
	}
	w := iout.FnWriter(func(p []byte) (int, error) {
		msgFn(string(p))
		return len(p), nil
	})

	man := NewManager(msgFn)
	man.serverWrapW = w

	// lang registrations
	u := []Registration{
		// WARNING: can't use stdio with stderr to be able to run scripts collectlog (use tcp if available)

		//GoplsRegistration(false, false,logTestVerbose()),
		goplsRegistration(true, false, verboseLog()),

		//cLangRegistration("", false),
		//cLangRegistration("", logTestVerbose()),
		cLangRegistration("clangd-19", false),

		pylspRegistration(true, false),

		// dummy
		Registration{"dummy1", []string{".dummy1"}, "stdio", "dummy_exec", nil},
		Registration{"dummy2", []string{".dummy2"}, "tcp", "dummy_exec", nil},
	}
	for _, reg := range u {
		if err := man.Register(&reg); err != nil {
			panic(err)
		}
	}

	return man
}

//----------

func getIntArgPossiblyFromEnv(st *testutil.ST, v string) (int, error) {
	// read offset (allow offset from env var)
	if v2 := st.Env.Get(v); v2 != "" {
		v = strings.TrimSpace(v2)
	}

	u, err := strconv.ParseInt(v, 10, 32)
	return int(u), err
}

//----------

func sourceCursor(t *testing.T, src string, nth int) (int, string) {
	src2, index, err := testutil.SourceCursor("●", src, nth)
	if err != nil {
		t.Fatal(err)
	}
	return index, src2
}
