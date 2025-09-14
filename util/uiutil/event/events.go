package event

import (
	"image"
)

type Event any

//----------

type WindowClose struct{}
type WindowResize struct{ Rect image.Rectangle }
type WindowExpose struct{ Rect image.Rectangle } // empty = full area

type WindowInput struct {
	Point image.Point
	Event Event
}

//----------

type MouseEnter struct{}
type MouseLeave struct{}

type MouseDown struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}
type MouseUp struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}
type MouseMove struct {
	Point   image.Point
	Buttons MouseButtons
	Mods    KeyModifiers
}

type MouseDragStart struct {
	Point   image.Point // starting (press) point (older then point2)
	Point2  image.Point // current point (move detection) (newest point)
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}
type MouseDragEnd struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}
type MouseDragMove struct {
	Point   image.Point
	Buttons MouseButtons
	Mods    KeyModifiers
}

type MouseClick struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}
type MouseDoubleClick struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}
type MouseTripleClick struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}

//----------

type KeyDown struct {
	Point   image.Point
	KeySym  KeySym
	Mods    KeyModifiers
	Buttons MouseButtons
	Rune    rune
}

type KeyUp struct {
	Point   image.Point
	KeySym  KeySym
	Mods    KeyModifiers
	Buttons MouseButtons
	Rune    rune
}

//----------

// drag and drop

type DndPosition struct {
	Point image.Point
	Types []DndType
	Reply func(DndAction)
}
type DndDrop struct {
	Point       image.Point
	ReplyAccept func(bool)
	RequestData func(DndType) ([]byte, error)
}

type DndAction int

const (
	DndADeny DndAction = iota
	DndACopy
	DndAMove
	DndALink
	DndAAsk
	DndAPrivate
)

type DndType int

const (
	TextURLListDndT DndType = iota // a list separated by '\n'
)

//----------

type Handled bool
