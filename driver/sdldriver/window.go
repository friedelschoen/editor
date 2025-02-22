//go:build !windows

package sdldriver

import (
	"fmt"
	"image"
	"image/draw"
	"time"

	"github.com/jmigpin/editor/ui/event"
	"github.com/veandco/go-sdl2/sdl"
)

const FPS = 60

type Window struct {
	window  *sdl.Window
	running bool
	events  []event.Event
}

func NewWindow() (*Window, error) {
	win := &Window{}

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}

	win.events = make([]event.Event, 0, 100)

	var err error
	win.window, err = sdl.CreateWindow("editor", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}

	w, h := win.window.GetSize()
	win.events = append(win.events, &event.WindowResize{
		Rect: image.Rect(0, 0, int(w), int(h)),
	})

	win.running = true

	return win, nil
}

func (win *Window) Update() {
	fmt.Println("update")
	win.window.UpdateSurface()
}

func (win *Window) Resize(rect image.Rectangle) error {
	pnt := rect.Size()
	win.window.SetSize(int32(pnt.X), int32(pnt.Y))
	return nil
}

// ClipboardDataGet implements driver.Window.
func (win *Window) ClipboardDataGet() (string, error) {
	return sdl.GetClipboardText()
}

// ClipboardDataSet implements driver.Window.
func (win *Window) ClipboardDataSet(text string) error {
	return sdl.SetClipboardText(text)
}

// Close implements driver.Window.
func (win *Window) Close() error {
	win.window.Destroy()
	sdl.Quit()
	return nil
}

// CursorSet implements driver.Window.
func (win *Window) CursorSet(cur event.Cursor) error {
	id := sdl.SystemCursor(0)
	switch cur {
	case event.NoneCursor:
		/* do nothing */
		return nil
	case event.DefaultCursor:
		id = sdl.SYSTEM_CURSOR_ARROW
	case event.NSResizeCursor:
		id = sdl.SYSTEM_CURSOR_SIZENS
	case event.WEResizeCursor:
		id = sdl.SYSTEM_CURSOR_SIZEWE
	case event.CloseCursor:
		id = sdl.SYSTEM_CURSOR_NO
	case event.MoveCursor:
		id = sdl.SYSTEM_CURSOR_CROSSHAIR
	case event.PointerCursor:
		id = sdl.SYSTEM_CURSOR_HAND
	case event.BeamCursor: // text cursor
		id = sdl.SYSTEM_CURSOR_IBEAM
	case event.WaitCursor: // watch cursor
		id = sdl.SYSTEM_CURSOR_WAITARROW
	}
	sdl.SetCursor(sdl.CreateSystemCursor(id))
	return nil
}

// Image implements driver.Window.
func (win *Window) Image() (draw.Image, error) {
	return win.window.GetSurface()
}

// PointerQuery implements driver.Window.
func (win *Window) PointerQuery() (image.Point, error) {
	x, y, _ := sdl.GetMouseState()
	return image.Point{int(x), int(y)}, nil
}

// PointerWarp implements driver.Window.
func (win *Window) PointerWarp(cur image.Point) error {
	win.window.WarpMouseInWindow(int32(cur.X), int32(cur.Y))
	return nil
}

// WindowSetName implements driver.Window.
func (win *Window) WindowSetName(title string) error {
	win.window.SetTitle(title)
	return nil
}

func getModifier() (res event.KeyModifiers) {
	mod := sdl.GetModState()
	switch {
	case mod&sdl.KMOD_CTRL != 0:
		res |= event.ModCtrl
	case mod&sdl.KMOD_SHIFT != 0:
		res |= event.ModShift
	case mod&sdl.KMOD_ALT != 0:
		res |= event.ModAlt
	case mod&sdl.KMOD_GUI != 0:
		res |= event.Mod4
	case mod&sdl.KMOD_NUM != 0:
		res |= event.Mod2
	case mod&sdl.KMOD_CAPS != 0:
		res |= event.ModLock
	case mod&sdl.KMOD_MODE != 0:
		res |= event.ModAltGr
	}
	return
}

func (win *Window) NextEvent() (event.Event, bool) {
	fmt.Println("poll")
	if len(win.events) > 0 {
		var el event.Event
		el, win.events = win.events[0], win.events[1:]
		return el, true
	}

	for {
		pollevent := sdl.PollEvent()
		if pollevent == nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		switch evt := pollevent.(type) {
		case *sdl.QuitEvent:
			win.running = false
			return event.WindowClose{}, false
		case *sdl.RenderEvent:
			return event.WindowExpose{}, true
		case *sdl.WindowEvent:
			w, h := win.window.GetSize()
			return &event.WindowResize{
				Rect: image.Rect(0, 0, int(w), int(h)),
			}, true
		case *sdl.MouseButtonEvent:
			pnt := image.Point{int(evt.X), int(evt.Y)}
			btn := event.MouseButton(0)
			switch evt.Button {
			case sdl.BUTTON_LEFT:
				btn = event.ButtonLeft
			case sdl.BUTTON_MIDDLE:
				btn = event.ButtonMiddle
			case sdl.BUTTON_RIGHT:
				btn = event.ButtonRight
			}

			switch evt.Clicks {
			case 1:
				return &event.MouseClick{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(),
				}, true
			case 2:
				return &event.MouseDoubleClick{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(),
				}, true
			case 3:
				return &event.MouseTripleClick{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(),
				}, true
			}
		case *sdl.MouseWheelEvent:
			pnt := image.Point{int(evt.X), int(evt.Y)}
			btn := event.ButtonWheelDown
			if evt.Direction == sdl.MOUSEWHEEL_FLIPPED {
				btn = event.ButtonWheelUp
			}
			return &event.MouseClick{
				Point:   pnt,
				Button:  btn,
				Buttons: event.MouseButtons(btn),
				Mods:    getModifier(),
			}, true
		case *sdl.MouseMotionEvent:
			btn := event.MouseButton(0)
			switch evt.State {
			case sdl.BUTTON_LEFT:
				btn = event.ButtonLeft
			case sdl.BUTTON_MIDDLE:
				btn = event.ButtonMiddle
			case sdl.BUTTON_RIGHT:
				btn = event.ButtonRight
			}
			return &event.MouseMove{
				Point:   image.Point{int(evt.X), int(evt.Y)},
				Buttons: event.MouseButtons(btn),
				Mods:    getModifier(),
			}, true
		case *sdl.KeyboardEvent:
			if evt.State == sdl.PRESSED {
				return &event.KeyDown{
					KeySym: event.KeySym(evt.Keysym.Sym),
					Rune:   rune(evt.Keysym.Sym),
				}, true
			} else {
				return &event.KeyUp{
					KeySym: event.KeySym(evt.Keysym.Sym),
					Rune:   rune(evt.Keysym.Sym),
				}, true
			}
		}
	}
}
