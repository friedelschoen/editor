package event

import (
	"image"
)

type Event interface {
	IsEvent()
}

type InputEvent interface {
	Event
	At() image.Point
}

type WindowClose struct{}

func (evt WindowClose) IsEvent() {}

type WindowResize struct {
	Rect image.Rectangle
}

func (evt WindowResize) IsEvent() {}

type WindowExpose struct {
	Rect image.Rectangle
}

func (evt WindowExpose) IsEvent() {}

type MouseEnter struct{}

func (evt MouseEnter) IsEvent() {}

type MouseLeave struct{}

func (evt MouseLeave) IsEvent() {}

type MouseDown struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}

func (evt MouseDown) IsEvent() {}
func (evt MouseDown) At() image.Point {
	return evt.Point
}

type MouseUp struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}

func (evt MouseUp) IsEvent() {}
func (evt MouseUp) At() image.Point {
	return evt.Point
}

type MouseMove struct {
	Point   image.Point
	Buttons MouseButtons
	Mods    KeyModifiers
}

func (evt MouseMove) IsEvent() {}
func (evt MouseMove) At() image.Point {
	return evt.Point
}

type MouseDragStart struct {
	Point   image.Point // starting (press) point (older then point2)
	Point2  image.Point // current point (move detection) (newest point)
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}

func (evt MouseDragStart) IsEvent() {}
func (evt MouseDragStart) At() image.Point {
	return evt.Point
}

type MouseDragEnd struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}

func (evt MouseDragEnd) IsEvent() {}
func (evt MouseDragEnd) At() image.Point {
	return evt.Point
}

type MouseDragMove struct {
	Point   image.Point
	Buttons MouseButtons
	Mods    KeyModifiers
}

func (evt MouseDragMove) IsEvent() {}
func (evt MouseDragMove) At() image.Point {
	return evt.Point
}

type MouseClick struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}

func (evt MouseClick) IsEvent() {}
func (evt MouseClick) At() image.Point {
	return evt.Point
}

type MouseDoubleClick struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}

func (evt MouseDoubleClick) IsEvent() {}
func (evt MouseDoubleClick) At() image.Point {
	return evt.Point
}

type MouseTripleClick struct {
	Point   image.Point
	Button  MouseButton
	Buttons MouseButtons // contains Button
	Mods    KeyModifiers
}

func (evt MouseTripleClick) IsEvent() {}
func (evt MouseTripleClick) At() image.Point {
	return evt.Point
}

type KeyDown struct {
	Point   image.Point
	KeySym  KeySym
	Mods    KeyModifiers
	Buttons MouseButtons
	Rune    rune
}

func (evt KeyDown) IsEvent() {}
func (evt KeyDown) At() image.Point {
	return evt.Point
}

type KeyUp struct {
	Point   image.Point
	KeySym  KeySym
	Mods    KeyModifiers
	Buttons MouseButtons
	Rune    rune
}

func (evt KeyUp) IsEvent() {}
func (evt KeyUp) At() image.Point {
	return evt.Point
}

// drag and drop

type DndPosition struct {
	Point image.Point
	Types []DndType
	Reply func(DndAction)
}

func (evt DndPosition) IsEvent() {}
func (evt DndPosition) At() image.Point {
	return evt.Point
}

type DndDrop struct {
	Point       image.Point
	ReplyAccept func(bool)
	RequestData func(DndType) ([]byte, error)
}

func (evt DndDrop) IsEvent() {}
func (evt DndDrop) At() image.Point {
	return evt.Point
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
