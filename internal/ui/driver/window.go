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
	cursor  sdl.SystemCursor

	dragging  bool
	dragStart image.Point
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

// GetClipboardData implements driver.Window.
func (win *Window) GetClipboardData() (string, error) {
	return sdl.GetClipboardText()
}

// SetClipboardData implements driver.Window.
func (win *Window) SetClipboardData(text string) error {
	return sdl.SetClipboardText(text)
}

// Close implements driver.Window.
func (win *Window) Close() error {
	win.window.Destroy()
	sdl.Quit()
	return nil
}

// SetCursor implements driver.Window.
func (win *Window) SetCursor(cur sdl.SystemCursor) {
	if win.cursor == cur {
		return
	}
	win.cursor = cur

	sdl.SetCursor(sdl.CreateSystemCursor(cur))
}

// Image implements driver.Window.
func (win *Window) Image() draw.Image {
	img, err := win.window.GetSurface()
	if err != nil {
		return image.NewRGBA(image.Rect(0, 0, 0, 0))
	}
	return img
}

// QueryPointer implements driver.Window.
func (win *Window) QueryPointer() (image.Point, error) {
	x, y, _ := sdl.GetMouseState()
	return image.Point{int(x), int(y)}, nil
}

// WarpPointer implements driver.Window.
func (win *Window) WarpPointer(cur image.Point) {
	win.window.WarpMouseInWindow(int32(cur.X), int32(cur.Y))
}

// WindowSetName implements driver.Window.
func (win *Window) WindowSetName(title string) error {
	win.window.SetTitle(title)
	return nil
}

func (win *Window) PushEvent(ev Event) {
	win.events <- ev
}

func (win *Window) NextEvent() Event {
	for {
		select {
		case event := <-win.events:
			return event
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
			return WindowClose{}
		case *sdl.WindowEvent:
			switch evt.Event {
			case sdl.WINDOWEVENT_ENTER:
				return &MouseEnter{}
			case sdl.WINDOWEVENT_LEAVE:
				return &MouseLeave{}
			case sdl.WINDOWEVENT_RESIZED:
				return &WindowResize{
					Rect: image.Rect(0, 0, int(evt.Data1), int(evt.Data2)),
				}
			case sdl.WINDOWEVENT_EXPOSED:
				w, h := win.window.GetSize()
				return &WindowExpose{
					Rect: image.Rect(0, 0, int(w), int(h)),
				}
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
				}
			} else {
				return &MouseUp{
					Point: pnt,
					Key:   key,
				}
			}
		case *sdl.MouseWheelEvent:
			return &MouseWheel{
				X: int(evt.X),
				Y: int(evt.Y),
			}
		case *sdl.MouseMotionEvent:
			pnt := image.Point{int(evt.X), int(evt.Y)}
			key := NewKey(KeyMouse)
			key.Mouse = 1 << (evt.State - 1)

			defer func() {
				win.dragStart = pnt
			}()

			if evt.State != 0 {
				if !win.dragging {
					if win.dragStart.X == 0 && win.dragStart.Y == 0 {
						win.dragStart = pnt
					} else {
						win.dragging = true

						return &MouseDragStart{
							Point:  win.dragStart,
							Point2: pnt,
							Key:    key,
						}
					}
				} else {
					return &MouseDragMove{
						Point: pnt,
						Key:   key,
					}
				}
			} else {
				if win.dragging {
					win.dragging = false
					return &MouseDragEnd{
						Point: pnt,
						Key:   key,
					}
				}
			}

			return &MouseMove{
				Point: pnt,
				Key:   key,
			}
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
					}
				} else {
					return &KeyUp{
						Key: win.lastkey,
					}
				}
			}
		case *sdl.TextInputEvent:
			win.lastkey.Rune, _ = utf8.DecodeRune(evt.Text[:])

			win.events <- &KeyUp{
				Key: win.lastkey,
			}

			return &KeyDown{
				Key: win.lastkey,
			}
		}
	}
}
