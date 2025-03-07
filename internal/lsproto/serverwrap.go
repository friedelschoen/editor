package lsproto

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"text/template"

	"github.com/friedelschoen/editor/internal/command"
	"github.com/friedelschoen/editor/internal/multierror"
)

type ServerWrap struct {
	Cmd command.CmdI
	rwc *rwc // just for IO mode (can be nil)
}

func getFreeTcpPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	p := l.Addr().(*net.TCPAddr).Port
	return p, nil
}

func StartServerWrapTCP(ctx context.Context, cmdTmpl string, w io.Writer) (*ServerWrap, string, error) {
	host := "127.0.0.1"

	// multiple editors can have multiple server wraps, need unique port
	port, err := getFreeTcpPort()
	if err != nil {
		return nil, "", err
	}

	// run cmd template
	cmd, addr, err := cmdTemplate(cmdTmpl, host, port)
	if err != nil {
		return nil, "", err
	}

	sw := newServerWrapCommon(ctx, cmd)

	// get lsp server output in tcp mode
	if w != nil {
		sw.Cmd.Cmd().Stdout = w
		sw.Cmd.Cmd().Stderr = w
	}

	if err := sw.Cmd.Start(); err != nil {
		return nil, "", err
	}
	return sw, addr, nil
}

func StartServerWrapIO(ctx context.Context, cmd string, stderr io.Writer, li *LangInstance) (*ServerWrap, io.ReadWriteCloser, error) {
	sw := newServerWrapCommon(ctx, cmd)

	pr1, pw1 := io.Pipe()
	pr2, pw2 := io.Pipe()

	sw.Cmd.Cmd().Stdin = pr1
	sw.Cmd.Cmd().Stdout = pw2
	sw.Cmd.Cmd().Stderr = stderr

	sw.rwc = &rwc{} // also keep for later close
	sw.rwc.WriteCloser = pw1
	sw.rwc.ReadCloser = pr2

	if err := sw.Cmd.Start(); err != nil {
		sw.rwc.Close() // wait will not be called, clear resources
		return nil, nil, err
	}

	return sw, sw.rwc, nil
}

func newServerWrapCommon(ctx context.Context, cmd string) *ServerWrap {
	sw := &ServerWrap{}
	args := strings.Split(cmd, " ") // TODO: escapes
	sw.Cmd = command.NewCmdIShell(ctx, args...)
	return sw
}

func (sw *ServerWrap) Wait() error {
	if sw.rwc != nil { // can be nil if in tcp mode
		// was set outside cmd, close after cmd.wait
		defer sw.rwc.Close()
	}

	return sw.Cmd.Wait()
}

type rwc struct {
	io.ReadCloser
	io.WriteCloser
}

func (rwc *rwc) Close() error {
	me := multierror.MultiError{}
	me.Add(rwc.ReadCloser.Close())
	me.Add(rwc.WriteCloser.Close())
	return me.Result()
}

func cmdTemplate(cmdTmpl string, host string, port int) (string, string, error) {
	// build template
	tmpl, err := template.New("").Parse(cmdTmpl)
	if err != nil {
		return "", "", err
	}

	// template data
	type tdata struct {
		Addr string
		Host string
		Port int
	}
	data := &tdata{}
	data.Host = host
	data.Port = port
	data.Addr = fmt.Sprintf("%s:%d", host, port)

	// fill template
	out := &bytes.Buffer{}
	if err := tmpl.Execute(out, data); err != nil {
		return "", "", err
	}
	return out.String(), data.Addr, nil
}
