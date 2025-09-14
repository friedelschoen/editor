package copypaste

// https://tronche.com/gui/x/icccm/

import (
	"bytes"
	"encoding/binary"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jmigpin/editor/driver/xdriver/xutil"
)

type Copy struct {
	conn *xgb.Conn
	win  xproto.Window

	// Data to transfer
	clipboardStr string
}

func NewCopy(conn *xgb.Conn, win xproto.Window) (*Copy, error) {
	c := &Copy{conn: conn, win: win}
	if err := xutil.LoadAtoms(conn, &CopyAtoms, false); err != nil {
		return nil, err
	}
	return c, nil
}

//----------

func (c *Copy) Set(str string) error {
	c.clipboardStr = str
	return c.set(CopyAtoms.Clipboard)
}
func (c *Copy) set(selection xproto.Atom) error {
	t := xproto.Timestamp(xproto.TimeCurrentTime)
	c1 := xproto.SetSelectionOwnerChecked(c.conn, c.win, selection, t)
	if err := c1.Check(); err != nil {
		return err
	}

	//// ensure the owner was set
	//c2 := xproto.GetSelectionOwner(c.conn, selection)
	//r, err := c2.Reply()
	//if err != nil {
	//	return err
	//}
	//if r.Owner != c.win {
	//	return fmt.Errorf("unable to get selection ownership")
	//}

	return nil
}

//----------

// Another application is asking for the data
func (c *Copy) OnSelectionRequest(ev *xproto.SelectionRequestEvent) error {
	//// DEBUG
	//target, _ := xutil.GetAtomName(c.conn, ev.Target)
	//sel, _ := xutil.GetAtomName(c.conn, ev.Selection)
	//prop, _ := xutil.GetAtomName(c.conn, ev.Property)
	//log.Printf("on selection request: %v %v %v", target, sel, prop)

	switch ev.Target {
	case CopyAtoms.String,
		CopyAtoms.Utf8String,
		CopyAtoms.Text,
		CopyAtoms.TextPlain,
		CopyAtoms.TextPlainCharsetUtf8:
		if err := c.transferBytes(ev); err != nil {
			return err
		}
	case CopyAtoms.Targets:
		if err := c.transferTargets(ev); err != nil {
			return err
		}
	default:
		// DEBUG
		//c.debugRequest(ev)

		// try to transfer bytes anyway
		if err := c.transferBytes(ev); err != nil {
			return err
		}
	}
	return nil
}

//----------

func (c *Copy) transferBytes(ev *xproto.SelectionRequestEvent) error {
	b := []byte(c.clipboardStr)

	// change property on the requestor
	c1 := xproto.ChangePropertyChecked(
		c.conn,
		xproto.PropModeReplace,
		ev.Requestor, // requestor window
		ev.Property,  // property
		ev.Target,
		8, // format
		uint32(len(b)),
		b)
	if err := c1.Check(); err != nil {
		return err
	}

	// notify the server
	sne := xproto.SelectionNotifyEvent{
		Requestor: ev.Requestor,
		Selection: ev.Selection,
		Target:    ev.Target,
		Property:  ev.Property,
		Time:      ev.Time,
	}
	c2 := xproto.SendEventChecked(
		c.conn,
		false,
		sne.Requestor,
		xproto.EventMaskNoEvent,
		string(sne.Bytes()))
	return c2.Check()
}

//----------

// testing: $ xclip -o -target TARGETS -selection primary

func (c *Copy) transferTargets(ev *xproto.SelectionRequestEvent) error {
	targets := []xproto.Atom{
		CopyAtoms.Targets,
		CopyAtoms.String,
		CopyAtoms.Utf8String,
		CopyAtoms.Text,
		CopyAtoms.TextPlain,
		CopyAtoms.TextPlainCharsetUtf8,
	}

	tbuf := new(bytes.Buffer)
	for _, t := range targets {
		binary.Write(tbuf, binary.LittleEndian, t)
	}

	// change property on the requestor
	c1 := xproto.ChangePropertyChecked(
		c.conn,
		xproto.PropModeReplace,
		ev.Requestor,   // requestor window
		ev.Property,    // property
		CopyAtoms.Atom, // (would not work in some cases with ev.Target)
		32,             // format
		uint32(len(targets)),
		tbuf.Bytes())
	if err := c1.Check(); err != nil {
		return err
	}

	// notify the server
	sne := xproto.SelectionNotifyEvent{
		Requestor: ev.Requestor,
		Selection: ev.Selection,
		Target:    ev.Target,
		Property:  ev.Property,
		Time:      ev.Time,
	}
	c2 := xproto.SendEventChecked(
		c.conn,
		false,
		sne.Requestor,
		xproto.EventMaskNoEvent,
		string(sne.Bytes()))
	return c2.Check()
}

//----------

// Another application now owns the selection.
func (c *Copy) OnSelectionClear(ev *xproto.SelectionClearEvent) {
	c.clipboardStr = ""
}

//----------

var CopyAtoms struct {
	Atom      xproto.Atom `loadAtoms:"ATOM"`
	Primary   xproto.Atom `loadAtoms:"PRIMARY"`
	Clipboard xproto.Atom `loadAtoms:"CLIPBOARD"`
	Targets   xproto.Atom `loadAtoms:"TARGETS"`

	Utf8String           xproto.Atom `loadAtoms:"UTF8_STRING"`
	String               xproto.Atom `loadAtoms:"STRING"`
	Text                 xproto.Atom `loadAtoms:"TEXT"`
	TextPlain            xproto.Atom `loadAtoms:"text/plain"`
	TextPlainCharsetUtf8 xproto.Atom `loadAtoms:"text/plain;charset=utf-8"`
}
