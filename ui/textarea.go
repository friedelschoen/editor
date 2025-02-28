package ui

import (
	"image"
	"unicode"

	"github.com/jmigpin/editor/ui/driver"
	"github.com/jmigpin/editor/ui/widget"
	"github.com/jmigpin/editor/util/drawutil/drawer4"
	"github.com/jmigpin/editor/util/evreg"
	"github.com/jmigpin/editor/util/iout/iorw"
	"github.com/jmigpin/editor/util/iout/iorw/rwedit"
	"github.com/veandco/go-sdl2/sdl"
)

type TextArea struct {
	*widget.TextEditX

	EvReg evreg.Register

	ui *UI
}

func NewTextArea(ui *UI) *TextArea {
	ta := &TextArea{ui: ui}
	ta.TextEditX = widget.NewTextEditX(ui)
	return ta
}

func (ta *TextArea) OnInputEvent(ev0 driver.Event, p image.Point) bool {
	h := false

	// input events callbacks (terminal related)
	ev2 := &TextAreaInputEvent{TextArea: ta, Event: ev0}
	ta.EvReg.RunCallbacks(TextAreaInputEventId, ev2)
	h = ev2.ReplyHandled

	// select annotation events
	if !h {
		h = ta.handleInputEvent2(ev0, p)
		// consider handled to avoid root events to select global annotations
		if h {
			return true
		}
	}

	if !h {
		h = ta.TextEditX.OnInputEvent(ev0)
		// don't consider handled to allow ui.Row to get inputevents
		if h {
			return false
		}
	}

	return h
}

func (ta *TextArea) handleInputEvent2(ev0 any, p image.Point) bool {
	switch ev := ev0.(type) {
	case *driver.MouseClick:
		if ev.Key.Is("MouseRight") {
			if !ta.PointIndexInsideSelection(ev.Point) {
				rwedit.MoveCursorToPoint(ta.EditCtx(), ev.Point, false)
			}
			i := ta.GetIndex(ev.Point)
			ev2 := &TextAreaCmdEvent{ta, i}
			ta.EvReg.RunCallbacks(TextAreaCmdEventId, ev2)
			return true
		}
	case *driver.MouseDown:
		if ev.Key.Mouse == driver.ButtonRight {
			ta.ENode.Cursor = sdl.SYSTEM_CURSOR_HAND
		}
	case *driver.MouseUp:
		if ev.Key.Mouse == driver.ButtonRight {
			ta.ENode.Cursor = sdl.SYSTEM_CURSOR_ARROW
		}
	case *driver.MouseDragStart:
		if ev.Key.Mouse == driver.ButtonRight {
			ta.ENode.Cursor = sdl.SYSTEM_CURSOR_ARROW
		}
	case *driver.KeyDown:
		if ev.Key.Is("Tab") {
			return ta.inlineCompleteEv()
		}
	}
	return false
}

func (ta *TextArea) inlineCompleteEv() bool {
	c := ta.Cursor()
	if c.HaveSelection() {
		return false
	}

	// previous rune should not be a space
	ru, _, err := iorw.ReadRuneAt(ta.RW(), c.Index()-1)
	if err != nil {
		return false
	}
	if unicode.IsSpace(ru) {
		return false
	}

	ev2 := &TextAreaInlineCompleteEvent{ta, c.Index(), false}
	ta.EvReg.RunCallbacks(TextAreaInlineCompleteEventId, ev2)
	return ev2.ReplyHandled
}

func (ta *TextArea) PointIndexInsideSelection(p image.Point) bool {
	c := ta.Cursor()
	if s, e, ok := c.SelectionIndexes(); ok {
		i := ta.GetIndex(p)
		return i >= s && i < e
	}
	return false
}

func (ta *TextArea) Layout() {
	ta.TextEditX.Layout()
	ta.setDrawer4Opts()

	ev2 := &TextAreaLayoutEvent{TextArea: ta}
	ta.EvReg.RunCallbacks(TextAreaLayoutEventId, ev2)
}

func (ta *TextArea) setDrawer4Opts() {
	if d, ok := ta.Drawer.(*drawer4.Drawer); ok {
		// scale cursor based on lineheight
		w := 1
		u := d.LineHeight()
		u2 := int(float64(u) * 0.08)
		if u2 > 1 {
			w = u2
		}
		d.Opt.Cursor.AddedWidth = w

		// set startoffsetx based on cursor
		d.Opt.RuneReader.StartOffsetX = d.Opt.Cursor.AddedWidth * 2
	}
}

const (
	TextAreaCmdEventId = iota
	TextAreaSelectAnnotationEventId
	TextAreaInlineCompleteEventId
	TextAreaInputEventId
	TextAreaLayoutEventId
)

type TextAreaCmdEvent struct {
	TextArea *TextArea
	Index    int
}

type TextAreaSelectAnnotationEvent struct {
	TextArea        *TextArea
	AnnotationIndex int
	Offset          int // annotation string click offset
	Type            TASelAnnType
}

type TASelAnnType int

const (
	TasatPrev TASelAnnType = iota
	TasatNext
	TasatMsg
	TasatMsgPrev
	TasatMsgNext
	TasatPrint
	TasatPrintPreviousAll
)

type TextAreaInlineCompleteEvent struct {
	TextArea *TextArea
	Offset   int

	ReplyHandled bool // allow callbacks to set value
}

type TextAreaInputEvent struct {
	TextArea     *TextArea
	Event        any
	ReplyHandled bool
}

type TextAreaLayoutEvent struct {
	TextArea *TextArea
}
