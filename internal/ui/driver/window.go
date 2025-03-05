package driver

import (
	"fmt"
	"image"
	"image/draw"
	"runtime"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/veandco/go-sdl2/sdl"
)

type Window struct {
	Events chan Event

	window     *sdl.Window
	running    bool
	lastkey    Key
	lastmotion time.Time
	cursor     sdl.SystemCursor
	dragging   bool
	dragStart  image.Point
}

// GetClipboardData implements driver.Window.
func GetClipboardData() (string, error) {
	return sdl.GetClipboardText()
}

// SetClipboardData implements driver.Window.
func SetClipboardData(text string) error {
	return sdl.SetClipboardText(text)
}

func NewWindow() (*Window, error) {
	win := &Window{}

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}

	win.Events = make(chan Event, 100)

	var err error
	win.window, err = sdl.CreateWindow("editor", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}

	w, h := win.window.GetSize()
	win.Events <- &WindowResize{
		Rect: image.Rect(0, 0, int(w), int(h)),
	}

	win.window.SetResizable(true)
	sdl.StartTextInput()

	win.running = true
	go win.eventLoop()

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

// Close implements driver.Window.
func (win *Window) Close() error {
	win.running = false
	close(win.Events)
	win.window.Destroy()
	sdl.Quit()
	time.AfterFunc(time.Second, func() {
		out := make([]byte, 100000)
		n := runtime.Stack(out, true)
		fmt.Println(string(out[:n]))
	})
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

func (win *Window) eventLoop() {
	for win.running {
		pollevent := sdl.PollEvent()
		if pollevent == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		switch evt := pollevent.(type) {
		case *sdl.QuitEvent:
			win.Events <- &WindowClose{}
		case *sdl.WindowEvent:
			switch evt.Event {
			case sdl.WINDOWEVENT_ENTER:
				win.Events <- &MouseEnter{}
			case sdl.WINDOWEVENT_LEAVE:
				win.Events <- &MouseLeave{}
			case sdl.WINDOWEVENT_RESIZED:
				win.Events <- &WindowResize{
					Rect: image.Rect(0, 0, int(evt.Data1), int(evt.Data2)),
				}
			case sdl.WINDOWEVENT_EXPOSED:
				w, h := win.window.GetSize()
				win.Events <- &WindowExpose{
					Rect: image.Rect(0, 0, int(w), int(h)),
				}
			}
		case *sdl.MouseButtonEvent:
			pnt := image.Point{int(evt.X), int(evt.Y)}
			key := NewKey(KeyMouse)
			key.Mouse = 1 << (int(evt.Button) - 1)

			if evt.State == sdl.PRESSED {
				win.Events <- &MouseClick{
					Point: pnt,
					Count: int(evt.Clicks),
					Key:   key,
				}
				win.Events <- &MouseDown{
					Point: pnt,
					Key:   key,
				}
			} else {
				win.Events <- &MouseUp{
					Point: pnt,
					Key:   key,
				}
			}
		case *sdl.MouseWheelEvent:
			win.Events <- &MouseWheel{
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

						win.Events <- &MouseDragStart{
							Point:  win.dragStart,
							Point2: pnt,
							Key:    key,
						}
						return
					}
				} else {
					win.Events <- &MouseDragMove{
						Point: pnt,
						Key:   key,
					}
					continue
				}
			} else if win.dragging {
				win.dragging = false
				win.Events <- &MouseDragEnd{
					Point: pnt,
					Key:   key,
				}
				continue
			}

			now := time.Now()
			if now.Sub(win.lastmotion) >= 500*time.Millisecond {
				win.lastmotion = now
				win.Events <- &MouseMove{
					Point: pnt,
					Key:   key,
				}
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
					win.Events <- &KeyDown{
						Key: win.lastkey,
					}
				} else {
					win.Events <- &KeyUp{
						Key: win.lastkey,
					}
				}
			}
		case *sdl.TextInputEvent:
			win.lastkey.Rune, _ = utf8.DecodeRune(evt.Text[:])

			win.Events <- &KeyUp{
				Key: win.lastkey,
			}

			win.Events <- &KeyDown{
				Key: win.lastkey,
			}
		}
	}
}
