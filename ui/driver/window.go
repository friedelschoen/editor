package driver

import (
	"image"
	"image/draw"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/veandco/go-sdl2/sdl"
)

const FPS = 60

type Window struct {
	window  *sdl.Window
	events  chan Event
	lastkey Key
}

func NewWindow() (*Window, error) {
	win := &Window{}

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}

	win.events = make(chan Event, 100)

	var err error
	win.window, err = sdl.CreateWindow("editor", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}

	w, h := win.window.GetSize()
	win.events <- &WindowResize{
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
func (win *Window) CursorSet(cur sdl.SystemCursor) error {
	sdl.SetCursor(sdl.CreateSystemCursor(cur))
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

func (win *Window) NextEvent() (Event, bool) {
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
			return WindowClose{}, false
		case *sdl.WindowEvent:
			switch evt.Event {
			case sdl.WINDOWEVENT_ENTER:
				return &MouseEnter{}, true
			case sdl.WINDOWEVENT_LEAVE:
				return &MouseLeave{}, true
			case sdl.WINDOWEVENT_RESIZED:
				return &WindowResize{
					Rect: image.Rect(0, 0, int(evt.Data1), int(evt.Data2)),
				}, true
			case sdl.WINDOWEVENT_EXPOSED:
				w, h := win.window.GetSize()
				return &WindowExpose{
					Rect: image.Rect(0, 0, int(w), int(h)),
				}, true
			}
		case *sdl.MouseButtonEvent:
			pnt := image.Point{int(evt.X), int(evt.Y)}
			key := NewKey(KeyMouse)
			key.Mouse = 1 << (int(evt.Button) - 1)

			if evt.State == sdl.PRESSED {
				win.events <- &MouseClick{
					Point: pnt,
					Count: int(evt.Clicks),
					Key:   key,
				}
				return &MouseDown{
					Point: pnt,
					Key:   key,
				}, true
			} else {
				return &MouseUp{
					Point: pnt,
					Key:   key,
				}, true
			}
		case *sdl.MouseWheelEvent:
			return &MouseWheel{
				X: int(evt.X),
				Y: int(evt.Y),
			}, true
		case *sdl.MouseMotionEvent:
			key := NewKey(KeyMouse)
			key.Mouse = 1 << (evt.State - 1)

			return &MouseMove{
				Point: image.Point{int(evt.X), int(evt.Y)},
				Key:   key,
			}, true
		case *sdl.KeyboardEvent:
			/*
				from SDL2:src/events/SDL_keyboard.c:1048
				do not send textinput when text starts with
				->  (unsigned char)*text < ' ' || *text == 127
				thats for sure!
				It seems like it does not send textinput when
				ctrl is pressed.
			*/
			win.lastkey = NewKeyFromKeysym(evt.Keysym)
			char := rune(evt.Keysym.Sym)
			if !unicode.IsPrint(char) || evt.Keysym.Mod&sdl.KMOD_CTRL != 0 {
				if evt.State == sdl.PRESSED {
					return &KeyDown{
						Key: win.lastkey,
					}, true
				} else {
					return &KeyUp{
						Key: win.lastkey,
					}, true
				}
			}
		case *sdl.TextInputEvent:
			win.lastkey.Rune, _ = utf8.DecodeRune(evt.Text[:])

			win.events <- &KeyUp{
				Key: win.lastkey,
			}

			return &KeyDown{
				Key: win.lastkey,
			}, true
		}
	}
}
