package driver

import (
	"image"
	"image/draw"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/jmigpin/editor/ui/event"
	"github.com/veandco/go-sdl2/sdl"
)

const FPS = 60

type Window struct {
	window *sdl.Window
	events chan event.Event
	istext atomic.Bool
}

func NewWindow() (*Window, error) {
	win := &Window{}

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}

	win.events = make(chan event.Event, 100)

	var err error
	win.window, err = sdl.CreateWindow("editor", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}

	w, h := win.window.GetSize()
	win.events <- &event.WindowResize{
		Rect: image.Rect(0, 0, int(w), int(h)),
	}

	win.window.SetResizable(true)
	sdl.StartTextInput()

	return win, nil
}

func (win *Window) Update() {
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

func getModifier(mod sdl.Keymod) (res event.KeyModifiers) {
	if mod == 0 {
		mod = sdl.GetModState()
	}
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
	for {
		select {
		case event := <-win.events:
			return event, true
		default:
			/* do nothing, continue */
		}

		pollevent := sdl.PollEvent()
		if pollevent == nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		switch evt := pollevent.(type) {
		case *sdl.QuitEvent:
			win.window.Destroy()
			sdl.Quit()
			return event.WindowClose{}, false
		case *sdl.WindowEvent:
			switch evt.Event {
			case sdl.WINDOWEVENT_ENTER:
				return &event.MouseEnter{}, true
			case sdl.WINDOWEVENT_LEAVE:
				return &event.MouseLeave{}, true
			case sdl.WINDOWEVENT_RESIZED:
				return &event.WindowResize{
					Rect: image.Rect(0, 0, int(evt.Data1), int(evt.Data2)),
				}, true
			case sdl.WINDOWEVENT_EXPOSED:
				w, h := win.window.GetSize()
				return &event.WindowExpose{
					Rect: image.Rect(0, 0, int(w), int(h)),
				}, true
			}
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
			if evt.State == sdl.PRESSED {
				switch evt.Clicks {
				case 1:
					win.events <- &event.MouseClick{
						Point:   pnt,
						Button:  btn,
						Buttons: event.MouseButtons(btn),
						Mods:    getModifier(0),
					}
				case 2:
					win.events <- &event.MouseDoubleClick{
						Point:   pnt,
						Button:  btn,
						Buttons: event.MouseButtons(btn),
						Mods:    getModifier(0),
					}
				case 3:
					win.events <- &event.MouseTripleClick{
						Point:   pnt,
						Button:  btn,
						Buttons: event.MouseButtons(btn),
						Mods:    getModifier(0),
					}
				}
				return &event.MouseDown{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(0),
				}, true
			} else {
				return &event.MouseUp{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(0),
				}, true
			}
		case *sdl.MouseWheelEvent:
			mx, my, _ := sdl.GetMouseState()
			pnt := image.Point{int(mx), int(my)}

			btn := event.ButtonWheelDown
			if evt.Y < 0 {
				btn = event.ButtonWheelUp
				evt.Y = -evt.Y
			}
			for i := int32(0); i < evt.Y; i++ {
				win.events <- &event.MouseDown{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(0),
				}
				win.events <- &event.MouseClick{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(0),
				}
				win.events <- &event.MouseUp{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(0),
				}
			}

			btn = event.ButtonWheelRight
			if evt.X < 0 {
				btn = event.ButtonWheelLeft
				evt.X = -evt.X
			}
			for i := int32(0); i < evt.X; i++ {
				win.events <- &event.MouseDown{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(0),
				}
				win.events <- &event.MouseClick{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(0),
				}
				win.events <- &event.MouseUp{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(0),
				}
			}
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
				Mods:    getModifier(0),
			}, true
		case *sdl.TextInputEvent:
			win.istext.Store(true)
			char, _ := utf8.DecodeRune(evt.Text[:])

			win.events <- &event.KeyUp{
				KeySym: event.KeySym(char),
				Rune:   char,
				Mods:   getModifier(0),
			}

			return &event.KeyDown{
				KeySym: event.KeySym(char),
				Rune:   char,
				Mods:   getModifier(0),
			}, true
		case *sdl.KeyboardEvent:
			sym, ok := symmap[int(evt.Keysym.Sym)]
			if !ok {
				sym = event.KSymNone
			}

			/*
				from SDL2:src/events/SDL_keyboard.c:1048
				do not send textinput when text starts with
				->  (unsigned char)*text < ' ' || *text == 127
				thats for sure!
				It seems like it does not send textinput when
				ctrl is pressed.
			*/
			char := byte(evt.Keysym.Sym)
			if char >= ' ' && char != 127 && evt.Keysym.Mod&sdl.KMOD_CTRL == 0 {
				continue /* wait for TextInputEvent */
			}
			if evt.State == sdl.PRESSED {
				return &event.KeyDown{
					KeySym: sym,
					Rune:   rune(evt.Keysym.Sym),
					Mods:   getModifier(sdl.Keymod(evt.Keysym.Mod)),
				}, true
			} else {
				return &event.KeyUp{
					KeySym: sym,
					Rune:   rune(evt.Keysym.Sym),
					Mods:   getModifier(sdl.Keymod(evt.Keysym.Mod)),
				}, true
			}
		}
	}
}
