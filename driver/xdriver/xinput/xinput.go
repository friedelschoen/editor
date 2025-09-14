package xinput

import (
	"image"

	"github.com/friedelschoen/editor/util/uiutil/event"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type XInput struct {
	km *KMap
}

func NewXInput(conn *xgb.Conn) (*XInput, error) {
	km, err := NewKMap(conn)
	if err != nil {
		return nil, err
	}
	xi := &XInput{km: km}
	return xi, nil
}

//----------

func (xi *XInput) ReadMapping() error {
	return xi.km.ReadMapping()
}

//----------

func (xi *XInput) KeyPress(ev *xproto.KeyPressEvent) *event.WindowInput {
	p := image.Point{int(ev.EventX), int(ev.EventY)}
	_, ks, ru := xi.km.Lookup(ev.Detail, ev.State)
	m := xi.km.modifiersToEventModifiers(ev.State)
	bs := translateModifiersToEventMouseButtons(ev.State)
	ev2 := &event.KeyDown{p, ks, m, bs, ru}
	return &event.WindowInput{p, ev2}
}
func (xi *XInput) KeyRelease(ev *xproto.KeyReleaseEvent) *event.WindowInput {
	p := image.Point{int(ev.EventX), int(ev.EventY)}
	_, ks, ru := xi.km.Lookup(ev.Detail, ev.State)
	m := xi.km.modifiersToEventModifiers(ev.State)
	bs := translateModifiersToEventMouseButtons(ev.State)
	ev2 := &event.KeyUp{p, ks, m, bs, ru}
	return &event.WindowInput{p, ev2}
}

func (xi *XInput) ButtonPress(ev *xproto.ButtonPressEvent) *event.WindowInput {
	p := image.Point{int(ev.EventX), int(ev.EventY)}
	b := translateButtonToEventButton(ev.Detail)
	bs := translateModifiersToEventMouseButtons(ev.State)
	m := xi.km.modifiersToEventModifiers(ev.State)
	ev2 := &event.MouseDown{p, b, bs, m}
	return &event.WindowInput{p, ev2}
}
func (xi *XInput) ButtonRelease(ev *xproto.ButtonReleaseEvent) *event.WindowInput {
	p := image.Point{int(ev.EventX), int(ev.EventY)}
	b := translateButtonToEventButton(ev.Detail)
	bs := translateModifiersToEventMouseButtons(ev.State)
	m := xi.km.modifiersToEventModifiers(ev.State)
	ev2 := &event.MouseUp{p, b, bs, m}
	return &event.WindowInput{p, ev2}
}

func (xi *XInput) MotionNotify(ev *xproto.MotionNotifyEvent) *event.WindowInput {
	p := image.Point{int(ev.EventX), int(ev.EventY)}
	bs := translateModifiersToEventMouseButtons(ev.State)
	m := xi.km.modifiersToEventModifiers(ev.State)
	ev2 := &event.MouseMove{p, bs, m}
	return &event.WindowInput{p, ev2}
}
