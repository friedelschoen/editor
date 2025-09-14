package lsproto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"strings"
	"sync"
	"time"

	"github.com/friedelschoen/editor/util/ctxutil"
	"github.com/friedelschoen/editor/util/iout"
	"github.com/friedelschoen/editor/util/iout/iorw"
)

type Client struct {
	rcli         *rpc.Client
	li           *LangInstance
	readLoopDone chan error

	lock struct {
		sync.Mutex
		fversions map[string]int
		//folders   []*WorkspaceFolder
	}

	serverCapabilities struct {
		workspace struct {
			folders bool
			symbol  bool
		}
		rename bool
	}
}

//----------

func NewClientTCP(ctx context.Context, addr string, li *LangInstance) (*Client, error) {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	cli := NewClientIO(ctx, conn, li)
	return cli, nil
}

//----------

func NewClientIO(ctx context.Context, rwc io.ReadWriteCloser, li *LangInstance) *Client {
	cli := &Client{li: li}
	cli.lock.fversions = map[string]int{}

	cc := NewJsonCodec(rwc)
	cc.OnNotificationMessage = cli.onNotificationMessage
	cc.OnUnexpectedServerReply = cli.onUnexpectedServerReply

	cli.rcli = rpc.NewClientWithCodec(cc)

	// wait for the codec readloop
	cli.readLoopDone = make(chan error, 1)
	go func() {
		err := cc.ReadLoop()
		if err != nil {
			cli.li.cancelCtx()
		}
		cli.readLoopDone <- err
	}()

	// close when ctx is done
	go func() {
		<-ctx.Done()
		if err := cli.sendClose(); err != nil {
			// Commented: best effort, ignore errors
			//cli.li.lang.PrintWrapError(err)
		}
		if err := rwc.Close(); err != nil {
			// Commented: best effort, ignore errors
			//cli.li.lang.PrintWrapError(err)
		}

		if err := context.Cause(ctx); err != context.Canceled {
			err = fmt.Errorf("ctxcause: %w", err)
			cli.li.lang.PrintWrapError(err)
		}
	}()

	return cli
}

//----------

func (cli *Client) Wait() error {
	return <-cli.readLoopDone
}

func (cli *Client) sendClose() error {
	me := iout.MultiError{}
	if err := cli.ShutdownRequest(); err != nil {
		me.Add(err)
	} else {
		me.Add(cli.ExitNotification())
	}
	return me.Result()
}

//----------

func (cli *Client) Call(ctx context.Context, method string, args, reply any) error {
	lspResp := &Response{}
	fn := func() error {
		return cli.rcli.Call(method, args, lspResp)
	}
	lateFn := func(err error) {
		if err != nil {
			err = fmt.Errorf("call late: %w", err)
			cli.li.lang.PrintWrapError(err)
		}
	}
	err := ctxutil.Call(ctx, method, fn, lateFn)
	if err != nil {
		err = fmt.Errorf("call: %w", err)
		return cli.li.lang.WrapError(err)
	}

	// not expecting a reply
	if _, ok := noreplyMethod(method); ok {
		return nil
	}

	// soft error (rpc data with error content)
	if lspResp.Error != nil {
		return cli.li.lang.WrapError(lspResp.Error)
	}

	// decode result
	return decodeJsonRaw(lspResp.Result, reply)
}

//----------

func (cli *Client) onNotificationMessage(msg *NotificationMessage) {
	// Msgs like:
	// - a notification was sent to the srv, not expecting a reply, but it receives one because it was an error (has id)
	// {"error":{"code":-32601,"message":"method not found"},"id":2,"jsonrpc":"2.0"}

	//logJson("notification <--: ", msg)
	//spew.Dump(msg)

	switch msg.Method {
	case "window/logMessage":
		//spew.Dump(msg.Params)

		lmp := msg.Params.lmp
		if lmp != nil {
			switch lmp.Type {
			case 1: // error
				err := fmt.Errorf("%v", lmp.Message)
				cli.li.lang.PrintWrapError(err)
			}
		}
	}
}

func (cli *Client) onUnexpectedServerReply(resp *Response) {
	if resp.Error != nil {
		// json-rpc error codes: https://www.jsonrpc.org/specification
		report := false
		switch resp.Error.Code {
		case -32601: // method not found
			report = true
		case -32602: // invalid params
			report = true

			//case -32603: // internal error
			//report = true
		}
		if report {
			err := fmt.Errorf("id=%v, code=%v, msg=%q, data=%v", resp.Id, resp.Error.Code, resp.Error.Message, resp.Error.Data)
			cli.li.lang.PrintWrapError(err)
		}
	}
}

//----------

func (cli *Client) Initialize(ctx context.Context) error {
	opt, err := cli.initializeParams()
	if err != nil {
		return err
	}
	logJson("opt -->: ", opt)

	serverCapabilities := (any)(nil)
	if err := cli.Call(ctx, "initialize", opt, &serverCapabilities); err != nil {
		return err
	}
	logJson("initialize <--: ", serverCapabilities)

	cli.readServerCapabilities(serverCapabilities)

	// send "initialized" (gopls: "no views" error without this)
	opt2 := json.RawMessage("{}")
	return cli.Call(ctx, "noreply:initialized", opt2, nil)
}

func (cli *Client) initializeParams() (json.RawMessage, error) {
	opt := []string{"\"capabilities\":{}"}

	//	rootUri, err := cli.rootUri()
	//	if err != nil {
	//		return nil, err
	//	}
	//	_ = rootUri

	//	// workspace folders
	//	cli.folders = []*WorkspaceFolder{{Uri: rootUri}}
	//	foldersBytes, err := encodeJson(cli.folders)
	//	if err != nil {
	//		return nil, err
	//	}

	//	// other capabilities
	//	//"capabilities":{
	//	//	"workspace":{
	//	//		"configuration":true,
	//	//		"workspaceFolders":true
	//	//	},
	//	//	"textDocument":{
	//	//		"publishDiagnostics":{
	//	//			"relatedInformation":true
	//	//		}
	//	//	}
	//	//}

	//	raw := json.RawMessage("{" +
	//		// TODO: gopls is not allowing rooturi=null at the moment...
	//		fmt.Sprintf("%q:%q", "rootUri", rootUri) + "," +
	//		// set workspace folders to use the later as "remove" value
	//		fmt.Sprintf("%q:%s", "workspaceFolders", foldersBytes) +
	//		"}")

	raw := "{" + strings.Join(opt, ",") + "}"
	return json.RawMessage(raw), nil
}

//func (cli *Client) rootUri() (DocumentUri, error) {
//	// using a non-existent dir to prevent an lsp server to start scanning the user disk doesn't work well (ex: gopls gives "no views in the session" after the cache is gone)
//	// use initial request file
//	dir := filepath.Dir(cli.li.lang.InstanceReqFilename)
//	rootUrl, err := absFilenameToUrl(dir)
//	if err != nil {
//		return "", err
//	}
//	return DocumentUri(rootUrl), nil
//}

func (cli *Client) readServerCapabilities(caps any) {
	path := "capabilities.workspace.workspaceFolders.supported"
	v, err := JsonGetPath(caps, path)
	if err == nil {
		if b, ok := v.(bool); ok && b {
			cli.serverCapabilities.workspace.folders = true
		}
	}

	path = "capabilities.workspaceSymbolProvider"
	v, err = JsonGetPath(caps, path)
	if err == nil {
		if b, ok := v.(bool); ok && b {
			cli.serverCapabilities.workspace.symbol = true
		}
	}

	path = "capabilities.renameProvider"
	v, err = JsonGetPath(caps, path)
	if err == nil {
		if b, ok := v.(bool); ok && b {
			cli.serverCapabilities.rename = true
		}
	}
}

//----------

func (cli *Client) ShutdownRequest() error {
	// https://microsoft.github.io/language-server-protocol/specification#shutdown

	// TODO: shutdown request should expect a reply
	// * clangd is sending a reply (ok)
	// * gopls is not sending a reply (NOT OK)

	// best effort, impose timeout
	ctx := context.Background()
	ctx2, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()
	ctx = ctx2

	err := cli.Call(ctx, "shutdown", nil, nil)
	return err
}

func (cli *Client) ExitNotification() error {
	// https://microsoft.github.io/language-server-protocol/specification#exit

	// no ctx timeout needed, it's not expecting a reply
	ctx := context.Background()
	err := cli.Call(ctx, "noreply:exit", nil, nil)
	return err
}

//----------

func (cli *Client) TextDocumentDidOpen(ctx context.Context, filename, text string, version int) error {
	// https://microsoft.github.io/language-server-protocol/specification#textDocument_didOpen

	opt := &DidOpenTextDocumentParams{}
	opt.TextDocument.LanguageId = cli.li.lang.Reg.Language
	opt.TextDocument.Version = version
	opt.TextDocument.Text = text
	url, err := AbsFilenameToUrl(filename)
	if err != nil {
		return err
	}
	opt.TextDocument.Uri = DocumentUri(url)
	return cli.Call(ctx, "noreply:textDocument/didOpen", opt, nil)
}

func (cli *Client) TextDocumentDidClose(ctx context.Context, filename string) error {
	// https://microsoft.github.io/language-server-protocol/specification#textDocument_didClose

	opt := &DidCloseTextDocumentParams{}
	url, err := AbsFilenameToUrl(filename)
	if err != nil {
		return err
	}
	opt.TextDocument.Uri = DocumentUri(url)
	return cli.Call(ctx, "noreply:textDocument/didClose", opt, nil)
}

func (cli *Client) TextDocumentDidChange(ctx context.Context, filename, text string, version int) error {
	// https://microsoft.github.io/language-server-protocol/specification#textDocument_didChange

	opt := &DidChangeTextDocumentParams{}
	opt.TextDocument.Version = &version
	url, err := AbsFilenameToUrl(filename)
	if err != nil {
		return err
	}
	opt.TextDocument.Uri = DocumentUri(url)

	// text end line/column
	rd := iorw.NewStringReaderAt(text)
	pos, err := OffsetToPosition(rd, len(text))
	if err != nil {
		return err
	}

	// changes
	opt.ContentChanges = []*TextDocumentContentChangeEvent{
		{
			Range: Range{
				Start: Position{0, 0},
				End:   pos,
			},
			//RangeLength: len(text), // TODO: not working?
			Text: text,
		},
	}
	return cli.Call(ctx, "noreply:textDocument/didChange", opt, nil)
}

func (cli *Client) TextDocumentDidSave(ctx context.Context, filename string, text []byte) error {
	// https://microsoft.github.io/language-server-protocol/specification#textDocument_didSave

	opt := &DidSaveTextDocumentParams{}
	opt.Text = string(text) // has omitempty
	url, err := AbsFilenameToUrl(filename)
	if err != nil {
		return err
	}
	opt.TextDocument.Uri = DocumentUri(url)

	return cli.Call(ctx, "noreply:textDocument/didSave", opt, nil)
}

//----------

func (cli *Client) TextDocumentDefinition(ctx context.Context, filename string, pos Position) (*Location, error) {
	// https://microsoft.github.io/language-server-protocol/specification#textDocument_definition

	opt := &TextDocumentPositionParams{}
	opt.Position = pos
	url, err := AbsFilenameToUrl(filename)
	if err != nil {
		return nil, err
	}
	opt.TextDocument.Uri = DocumentUri(url)

	result := []*Location{}
	if err := cli.Call(ctx, "textDocument/definition", opt, &result); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no results")
	}
	return result[0], nil // first result only
}

//----------

func (cli *Client) TextDocumentImplementation(ctx context.Context, filename string, pos Position) (*Location, error) {
	// https://microsoft.github.io/language-server-protocol/specification#textDocument_implementation

	opt := &TextDocumentPositionParams{}
	opt.Position = pos
	url, err := AbsFilenameToUrl(filename)
	if err != nil {
		return nil, err
	}
	opt.TextDocument.Uri = DocumentUri(url)

	result := []*Location{}
	if err := cli.Call(ctx, "textDocument/implementation", opt, &result); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no results")
	}
	return result[0], nil // first result only
}

//----------

func (cli *Client) TextDocumentCompletion(ctx context.Context, filename string, pos Position) (*CompletionList, error) {
	// https://microsoft.github.io/language-server-protocol/specification#textDocument_completion

	opt := &CompletionParams{}
	opt.Context.TriggerKind = 1 // invoked
	opt.Position = pos
	url, err := AbsFilenameToUrl(filename)
	if err != nil {
		return nil, err
	}
	opt.TextDocument.Uri = DocumentUri(url)

	result := CompletionList{}
	if err := cli.Call(ctx, "textDocument/completion", opt, &result); err != nil {
		return nil, err
	}
	//logJson("textdocumentcompletion", result)
	return &result, nil
}

//----------

func (cli *Client) TextDocumentDidOpenVersion(ctx context.Context, filename string, b []byte) error {

	cli.lock.Lock()
	v, ok := cli.lock.fversions[filename]
	if !ok {
		v = 1
	} else {
		v++
	}
	cli.lock.fversions[filename] = v
	cli.lock.Unlock()

	return cli.TextDocumentDidOpen(ctx, filename, string(b), v)
}

//----------

//func (cli *Client) WorkspaceDidChangeWorkspaceFolders(ctx context.Context, added, removed []*WorkspaceFolder) error {
//	opt := &DidChangeWorkspaceFoldersParams{}
//	opt.Event = &WorkspaceFoldersChangeEvent{}
//	opt.Event.Added = added
//	opt.Event.Removed = removed
//	err := cli.Call(ctx, "noreply:workspace/didChangeWorkspaceFolders", opt, nil)
//	return err
//}

//----------

//func (cli *Client) UpdateWorkspaceFolder(ctx context.Context, dir string) error {
//	if !cli.serverCapabilities.workspace.folders {
//		return nil
//	}

//	removed := cli.folders
//	url, err := absFilenameToUrl(dir)
//	if err != nil {
//		return err
//	}
//	cli.folders = []*WorkspaceFolder{{Uri: DocumentUri(url)}}
//	return cli.WorkspaceDidChangeWorkspaceFolders(ctx, cli.folders, removed)

//}

// TODO
//return cli.WorkspaceDidChangeConfiguration(ctx, dir)
//func (cli *Client) WorkspaceDidChangeConfiguration(ctx context.Context, dir string) error {
//	url, err := absFilenameToUrl(dir)
//	if err != nil {
//		return err
//	}
//	//"settings":{"rootUri":"` + url + `"}
//	opt := json.RawMessage(`{
//		"settings":{"workspaceFolders":[{"uri":"` + url + `"}]}
//	}`)
//	return cli.Call(ctx, "noreply:workspace/didChangeConfiguration", opt, nil)
//}

//----------

func (cli *Client) TextDocumentRename(ctx context.Context, filename string, pos Position, newName string) (*WorkspaceEdit, error) {
	//// Commented: try it anyway
	//if !cli.serverCapabilities.rename {
	//	return nil, fmt.Errorf("server did not advertize rename capability")
	//}

	opt := &RenameParams{}
	opt.NewName = newName
	opt.Position = pos
	url, err := AbsFilenameToUrl(filename)
	if err != nil {
		return nil, err
	}
	opt.TextDocument.Uri = DocumentUri(url)
	result := WorkspaceEdit{}
	err = cli.Call(ctx, "textDocument/rename", opt, &result)
	return &result, err
}

//----------

func (cli *Client) TextDocumentPrepareCallHierarchy(ctx context.Context, filename string, pos Position) ([]*CallHierarchyItem, error) {
	opt := &CallHierarchyPrepareParams{}
	opt.Position = pos
	url, err := AbsFilenameToUrl(filename)
	if err != nil {
		return nil, err
	}
	opt.TextDocument.Uri = DocumentUri(url)
	result := []*CallHierarchyItem{}
	err = cli.Call(ctx, "textDocument/prepareCallHierarchy", opt, &result)
	return result, err
}
func (cli *Client) CallHierarchyCalls(ctx context.Context, typ CallHierarchyCallType, item *CallHierarchyItem) ([]*CallHierarchyCall, error) {
	method := ""
	switch typ {
	case IncomingChct:
		method = "callHierarchy/incomingCalls"
	case OutgoingChct:
		method = "callHierarchy/outgoingCalls"
	default:
		panic("bad type")
	}
	opt := &CallHierarchyCallsParams{}
	opt.Item = item
	result := []*CallHierarchyCall{}
	err := cli.Call(ctx, method, opt, &result)
	return result, err
}

//----------

func (cli *Client) TextDocumentReferences(ctx context.Context, filename string, pos Position) ([]*Location, error) {
	opt := &ReferenceParams{}
	opt.Context.IncludeDeclaration = true
	url, err := AbsFilenameToUrl(filename)
	if err != nil {
		return nil, err
	}
	opt.TextDocument.Uri = DocumentUri(url)
	opt.Position = pos
	result := []*Location{}
	err = cli.Call(ctx, "textDocument/references", opt, &result)
	return result, err
}
