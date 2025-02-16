//go:build js && editorDebugExecSide

// NOTE: ex: browser client is always js/execside

package debug

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"syscall/js"
	"time"
)

// NOTE: not supporting being a server in the browser
func listen2(ctx context.Context, addr Addr) (Listener, error) {
	return nil, fmt.Errorf("not supported")
}

func dial2(ctx context.Context, addr Addr) (Conn, error) {
	// TODO: ctx

	u := websocketEntryPathUrl(addr.String())
	ws := js.Global().Get("WebSocket").New(u)

	// easier to deal with then a blob (could use blob?)
	ws.Set("binaryType", "arraybuffer")

	return newWsConn(addr, ws)
}

type WsConn struct {
	addr    Addr
	ws      js.Value
	readCh  chan any
	readBuf bytes.Buffer
}

func newWsConn(addr Addr, ws js.Value) (*WsConn, error) {
	wsc := &WsConn{addr: addr, ws: ws}
	wsc.readCh = make(chan any, 1)

	openCh := make(chan any, 1)
	openDone := false

	jsEvListen(ws, "error", jsFuncOf2(func(args []js.Value) {
		//jsLog(args)
		//jsLog(args[0])
		jsErr := args[0]
		message := jsErr.Get("message").String()
		err := errors.New(message)
		if !openDone {
			openCh <- err
		} else {
			wsc.readCh <- err
		}
	}))
	jsEvListen(ws, "open", jsFuncOf2Release(func(args []js.Value) {
		openDone = true
		openCh <- struct{}{}
	}))
	jsEvListen(ws, "message", jsFuncOf2(func(args []js.Value) {
		//jsLog(args[0])
		jsArr := args[0].Get("data")
		b := arrayBufferToBytes(jsArr)
		wsc.readCh <- b
	}))

	v := <-openCh
	switch t := v.(type) {
	case error:
		return nil, t
	}

	return wsc, nil
}

func (wsc *WsConn) Read(b []byte) (int, error) {
	if len(b) <= wsc.readBuf.Len() {
		return wsc.readBuf.Read(b)
	}

	v := <-wsc.readCh
	switch t := v.(type) {
	case error:
		return 0, t
	case []byte:
		// append to buffer first
		if _, err := wsc.readBuf.Write(t); err != nil {
			err2 := fmt.Errorf("wsc.read: %v", err)
			panic(err2)
		}
		// try again
		return wsc.Read(b)
	default:
		panic("!")
	}
}
func (wsc *WsConn) Write(b []byte) (int, error) {
	jsArr := bytesToJsArray(b)
	wsc.ws.Call("send", jsArr)
	return len(b), nil // TODO: error (exception)
}
func (wsc *WsConn) Close() error {
	wsc.ws.Call("close")
	return nil // TODO: error (exception)
}
func (wsc *WsConn) LocalAddr() Addr {
	addr := &AddrImpl{"", "<local addr not available>"}
	return addr
}
func (wsc *WsConn) RemoteAddr() Addr {
	return wsc.addr
}

func (wsc *WsConn) SetDeadline(time.Time) error {
	return fmt.Errorf("setdeadline: todo")
}
func (wsc *WsConn) SetReadDeadline(time.Time) error {
	return fmt.Errorf("setreaddeadline: todo")
}
func (wsc *WsConn) SetWriteDeadline(time.Time) error {
	return fmt.Errorf("setwritedeadline: todo")
}

func arrayBufferToBytes(arrBuf js.Value) []byte {
	// js arraybuffer to js array
	jsArr := js.Global().Get("Uint8Array").New(arrBuf)
	// js array to go slice
	b := make([]byte, jsArr.Get("byteLength").Int())
	js.CopyBytesToGo(b, jsArr)
	return b
}
func bytesToJsArray(b []byte) js.Value {
	// go slice to js array
	jsArr := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(jsArr, b)
	return jsArr
}

// simplifies to not need to return a value
func jsFuncOf2(fn func(args []js.Value)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		fn(args)
		return nil
	})
}

func jsFuncOf2Release(fn func(args []js.Value)) js.Func {
	fn2 := js.Func{}
	fn2 = jsFuncOf2(func(args []js.Value) {
		fn(args)
		fn2.Release()
	})
	return fn2
}

func jsEvListen(v js.Value, evName string, fn js.Func) {
	v.Call("addEventListener", evName, fn)
}

func jsLog(args ...any) {
	js.Global().Get("console").Call("log", args...)
}
