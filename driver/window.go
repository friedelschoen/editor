package driver

import (
	"image"
	"image/draw"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jmigpin/editor/ui/event"
	"github.com/veandco/go-sdl2/sdl"
)

const FPS = 60

type Window struct {
	window  *sdl.Window
	events  chan event.Event
	running bool
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
	win.running = true

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

var eventnames = map[uint32]string{
	// Application events
	sdl.QUIT: "quit", // user-requested quit

	// Display events
	sdl.DISPLAYEVENT: "displayevent", // Display state change

	// Window events
	sdl.WINDOWEVENT: "windowevent", // window state change
	sdl.SYSWMEVENT:  "syswmevent",  // system specific event

	// Keyboard events
	sdl.KEYDOWN:         "keydown",         // key pressed
	sdl.KEYUP:           "keyup",           // key released
	sdl.TEXTEDITING:     "textediting",     // keyboard text editing (composition)
	sdl.TEXTINPUT:       "textinput",       // keyboard text input
	sdl.TEXTEDITING_EXT: "textediting_ext", // keyboard text editing (composition)
	sdl.KEYMAPCHANGED:   "keymapchanged",   // keymap changed due to a system event such as an input language or keyboard layout change (>= SDL 2.0.4)

	// Mouse events
	sdl.MOUSEMOTION:     "mousemotion",     // mouse moved
	sdl.MOUSEBUTTONDOWN: "mousebuttondown", // mouse button pressed
	sdl.MOUSEBUTTONUP:   "mousebuttonup",   // mouse button released
	sdl.MOUSEWHEEL:      "mousewheel",      // mouse wheel motion

	// Joystick events
	sdl.JOYAXISMOTION:    "joyaxismotion",    // joystick axis motion
	sdl.JOYBALLMOTION:    "joyballmotion",    // joystick trackball motion
	sdl.JOYHATMOTION:     "joyhatmotion",     // joystick hat position change
	sdl.JOYBUTTONDOWN:    "joybuttondown",    // joystick button pressed
	sdl.JOYBUTTONUP:      "joybuttonup",      // joystick button released
	sdl.JOYDEVICEADDED:   "joydeviceadded",   // joystick connected
	sdl.JOYDEVICEREMOVED: "joydeviceremoved", // joystick disconnected

	// Game controller events
	sdl.CONTROLLERAXISMOTION:     "controlleraxismotion",     // controller axis motion
	sdl.CONTROLLERBUTTONDOWN:     "controllerbuttondown",     // controller button pressed
	sdl.CONTROLLERBUTTONUP:       "controllerbuttonup",       // controller button released
	sdl.CONTROLLERDEVICEADDED:    "controllerdeviceadded",    // controller connected
	sdl.CONTROLLERDEVICEREMOVED:  "controllerdeviceremoved",  // controller disconnected
	sdl.CONTROLLERDEVICEREMAPPED: "controllerdeviceremapped", // controller mapping updated

	// Touch events
	sdl.FINGERDOWN:   "fingerdown",   // user has touched input device
	sdl.FINGERUP:     "fingerup",     // user stopped touching input device
	sdl.FINGERMOTION: "fingermotion", // user is dragging finger on input device

	// Gesture events
	sdl.DOLLARGESTURE: "dollargesture",
	sdl.DOLLARRECORD:  "dollarrecord",
	sdl.MULTIGESTURE:  "multigesture",

	// Clipboard events
	sdl.CLIPBOARDUPDATE: "clipboardupdate", // the clipboard changed

	// Drag and drop events
	sdl.DROPFILE:     "dropfile",     // the system requests a file open
	sdl.DROPTEXT:     "droptext",     // text/plain drag-and-drop event
	sdl.DROPBEGIN:    "dropbegin",    // a new set of drops is beginning (NULL filename)
	sdl.DROPCOMPLETE: "dropcomplete", // current set of drops is now complete (NULL filename)

	// Audio hotplug events
	sdl.AUDIODEVICEADDED:   "audiodeviceadded",   // a new audio device is available (>= SDL 2.0.4)
	sdl.AUDIODEVICEREMOVED: "audiodeviceremoved", // an audio device has been removed (>= SDL 2.0.4)

	// Sensor events
	sdl.SENSORUPDATE: "sensorupdate", // a sensor was updated

	// Render events
	sdl.RENDER_TARGETS_RESET: "render_targets_reset", // the render targets have been reset and their contents need to be updated (>= SDL 2.0.2)
	sdl.RENDER_DEVICE_RESET:  "render_device_reset",  // the device has been reset and all textures need to be recreated (>= SDL 2.0.4)

	// These are for your use, and should be allocated with RegisterEvents()
	sdl.USEREVENT: "userevent", // a user-specified event
	sdl.LASTEVENT: "lastevent", // (only for bounding internal arrays)
}

func (win *Window) NextEvent() (event.Event, bool) {
	// if !win.running {
	// 	return nil, false
	// }
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
			win.running = false
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
						Mods:    getModifier(),
					}
				case 2:
					win.events <- &event.MouseDoubleClick{
						Point:   pnt,
						Button:  btn,
						Buttons: event.MouseButtons(btn),
						Mods:    getModifier(),
					}
				case 3:
					win.events <- &event.MouseTripleClick{
						Point:   pnt,
						Button:  btn,
						Buttons: event.MouseButtons(btn),
						Mods:    getModifier(),
					}
				}
				return &event.MouseDown{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(),
				}, true
			} else {
				return &event.MouseUp{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(),
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
					Mods:    getModifier(),
				}
				win.events <- &event.MouseClick{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(),
				}
				win.events <- &event.MouseUp{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(),
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
					Mods:    getModifier(),
				}
				win.events <- &event.MouseClick{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(),
				}
				win.events <- &event.MouseUp{
					Point:   pnt,
					Button:  btn,
					Buttons: event.MouseButtons(btn),
					Mods:    getModifier(),
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
				Mods:    getModifier(),
			}, true
		case *sdl.TextInputEvent:
			char, _ := utf8.DecodeRune(evt.Text[:])

			win.events <- &event.KeyUp{
				KeySym: event.KeySym(char),
				Rune:   char,
			}

			return &event.KeyDown{
				KeySym: event.KeySym(char),
				Rune:   char,
			}, true
		case *sdl.KeyboardEvent:
			if unicode.IsPrint(rune(evt.Keysym.Sym)) {
				continue /* wait for TextInputEvent */
			}

			sym, ok := symmap[int(evt.Keysym.Sym)]
			if !ok {
				sym = event.KSymNone
			}
			if evt.State == sdl.PRESSED {
				return &event.KeyDown{
					KeySym: sym,
					Rune:   rune(evt.Keysym.Sym),
				}, true
			} else {
				return &event.KeyUp{
					KeySym: sym,
					Rune:   rune(evt.Keysym.Sym),
				}, true
			}
		}
	}
}
