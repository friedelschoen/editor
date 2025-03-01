package driver

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/veandco/go-sdl2/sdl"
)

type KeyType int

type MouseButton int

const (
	KeyRune KeyType = iota
	KeyControl
	KeyMouse
)

/*
#define SDL_BUTTON_MASK(X)  (1u << ((X)-1))
#define SDL_BUTTON_LMASK    SDL_BUTTON_MASK(SDL_BUTTON_LEFT)
#define SDL_BUTTON_MMASK    SDL_BUTTON_MASK(SDL_BUTTON_MIDDLE)
#define SDL_BUTTON_RMASK    SDL_BUTTON_MASK(SDL_BUTTON_RIGHT)
#define SDL_BUTTON_X1MASK   SDL_BUTTON_MASK(SDL_BUTTON_X1)
#define SDL_BUTTON_X2MASK   SDL_BUTTON_MASK(SDL_BUTTON_X2)
*/
const (
	ButtonLeft MouseButton = 1 << iota
	ButtonMiddle
	ButtonRight
	ButtonX1
	ButtonX2
)

type Key struct {
	Type     KeyType
	KeyMod   sdl.Keymod
	MouseMod MouseButton
	Sym      sdl.Scancode
	Mouse    MouseButton
	Rune     rune
}

func NewKey(typ KeyType) Key {
	_, _, mousemod := sdl.GetMouseState()
	return Key{
		Type:     typ,
		KeyMod:   sdl.GetModState(),
		MouseMod: MouseButton(mousemod),
	}
}

func NewKeyFromKeysym(sym sdl.Keysym) Key {
	_, _, mousemod := sdl.GetMouseState()
	return Key{
		Type:     KeyControl,
		KeyMod:   sdl.Keymod(sym.Mod),
		MouseMod: MouseButton(mousemod),
		Sym:      sym.Scancode,
	}
}

var scancodes = []string{
	/* 0 */ "",
	/* 1 */ "",
	/* 2 */ "",
	/* 3 */ "",
	/* 4 */ "A",
	/* 5 */ "B",
	/* 6 */ "C",
	/* 7 */ "D",
	/* 8 */ "E",
	/* 9 */ "F",
	/* 10 */ "G",
	/* 11 */ "H",
	/* 12 */ "I",
	/* 13 */ "J",
	/* 14 */ "K",
	/* 15 */ "L",
	/* 16 */ "M",
	/* 17 */ "N",
	/* 18 */ "O",
	/* 19 */ "P",
	/* 20 */ "Q",
	/* 21 */ "R",
	/* 22 */ "S",
	/* 23 */ "T",
	/* 24 */ "U",
	/* 25 */ "V",
	/* 26 */ "W",
	/* 27 */ "X",
	/* 28 */ "Y",
	/* 29 */ "Z",
	/* 30 */ "1",
	/* 31 */ "2",
	/* 32 */ "3",
	/* 33 */ "4",
	/* 34 */ "5",
	/* 35 */ "6",
	/* 36 */ "7",
	/* 37 */ "8",
	/* 38 */ "9",
	/* 39 */ "0",
	/* 40 */ "Return",
	/* 41 */ "Escape",
	/* 42 */ "Backspace",
	/* 43 */ "Tab",
	/* 44 */ "Space",
	/* 45 */ "-",
	/* 46 */ "=",
	/* 47 */ "[",
	/* 48 */ "]",
	/* 49 */ "\\",
	/* 50 */ "#",
	/* 51 */ ";",
	/* 52 */ "'",
	/* 53 */ "`",
	/* 54 */ ",",
	/* 55 */ ".",
	/* 56 */ "/",
	/* 57 */ "CapsLock",
	/* 58 */ "F1",
	/* 59 */ "F2",
	/* 60 */ "F3",
	/* 61 */ "F4",
	/* 62 */ "F5",
	/* 63 */ "F6",
	/* 64 */ "F7",
	/* 65 */ "F8",
	/* 66 */ "F9",
	/* 67 */ "F10",
	/* 68 */ "F11",
	/* 69 */ "F12",
	/* 70 */ "PrintScreen",
	/* 71 */ "ScrollLock",
	/* 72 */ "Pause",
	/* 73 */ "Insert",
	/* 74 */ "Home",
	/* 75 */ "PageUp",
	/* 76 */ "Delete",
	/* 77 */ "End",
	/* 78 */ "PageDown",
	/* 79 */ "Right",
	/* 80 */ "Left",
	/* 81 */ "Down",
	/* 82 */ "Up",
	/* 83 */ "Numlock",
	/* 84 */ "Keypad/",
	/* 85 */ "Keypad*",
	/* 86 */ "Keypad-",
	/* 87 */ "Keypad+",
	/* 88 */ "KeypadEnter",
	/* 89 */ "Keypad1",
	/* 90 */ "Keypad2",
	/* 91 */ "Keypad3",
	/* 92 */ "Keypad4",
	/* 93 */ "Keypad5",
	/* 94 */ "Keypad6",
	/* 95 */ "Keypad7",
	/* 96 */ "Keypad8",
	/* 97 */ "Keypad9",
	/* 98 */ "Keypad0",
	/* 99 */ "Keypad.",
	/* 100 */ "NonUSBackslash",
	/* 101 */ "Application",
	/* 102 */ "Power",
	/* 103 */ "Keypad=",
	/* 104 */ "F13",
	/* 105 */ "F14",
	/* 106 */ "F15",
	/* 107 */ "F16",
	/* 108 */ "F17",
	/* 109 */ "F18",
	/* 110 */ "F19",
	/* 111 */ "F20",
	/* 112 */ "F21",
	/* 113 */ "F22",
	/* 114 */ "F23",
	/* 115 */ "F24",
	/* 116 */ "Execute",
	/* 117 */ "Help",
	/* 118 */ "Menu",
	/* 119 */ "Select",
	/* 120 */ "Stop",
	/* 121 */ "Again",
	/* 122 */ "Undo",
	/* 123 */ "Cut",
	/* 124 */ "Copy",
	/* 125 */ "Paste",
	/* 126 */ "Find",
	/* 127 */ "Mute",
	/* 128 */ "VolumeUp",
	/* 129 */ "VolumeDown",
	/* 130 */ "",
	/* 131 */ "",
	/* 132 */ "",
	/* 133 */ "Keypad,",
	/* 134 */ "Keypad= (AS400)",
	/* 135 */ "International1",
	/* 136 */ "International2",
	/* 137 */ "International3",
	/* 138 */ "International4",
	/* 139 */ "International5",
	/* 140 */ "International6",
	/* 141 */ "International7",
	/* 142 */ "International8",
	/* 143 */ "International9",
	/* 144 */ "Language1",
	/* 145 */ "Language2",
	/* 146 */ "Language3",
	/* 147 */ "Language4",
	/* 148 */ "Language5",
	/* 149 */ "Language6",
	/* 150 */ "Language7",
	/* 151 */ "Language8",
	/* 152 */ "Language9",
	/* 153 */ "AltErase",
	/* 154 */ "SysReq",
	/* 155 */ "Cancel",
	/* 156 */ "Clear",
	/* 157 */ "Prior",
	/* 158 */ "Return",
	/* 159 */ "Separator",
	/* 160 */ "Out",
	/* 161 */ "Oper",
	/* 162 */ "ClearAgain",
	/* 163 */ "CrSel",
	/* 164 */ "ExSel",
	/* 165 */ "",
	/* 166 */ "",
	/* 167 */ "",
	/* 168 */ "",
	/* 169 */ "",
	/* 170 */ "",
	/* 171 */ "",
	/* 172 */ "",
	/* 173 */ "",
	/* 174 */ "",
	/* 175 */ "",
	/* 176 */ "Keypad00",
	/* 177 */ "Keypad000",
	/* 178 */ "ThousandsSeparator",
	/* 179 */ "DecimalSeparator",
	/* 180 */ "CurrencyUnit",
	/* 181 */ "CurrencySubUnit",
	/* 182 */ "Keypad(",
	/* 183 */ "Keypad)",
	/* 184 */ "Keypad{",
	/* 185 */ "Keypad}",
	/* 186 */ "KeypadTab",
	/* 187 */ "KeypadBackspace",
	/* 188 */ "KeypadA",
	/* 189 */ "KeypadB",
	/* 190 */ "KeypadC",
	/* 191 */ "KeypadD",
	/* 192 */ "KeypadE",
	/* 193 */ "KeypadF",
	/* 194 */ "KeypadXOR",
	/* 195 */ "Keypad^",
	/* 196 */ "Keypad%",
	/* 197 */ "Keypad<",
	/* 198 */ "Keypad>",
	/* 199 */ "Keypad&",
	/* 200 */ "Keypad&&",
	/* 201 */ "Keypad|",
	/* 202 */ "Keypad||",
	/* 203 */ "Keypad:",
	/* 204 */ "Keypad#",
	/* 205 */ "KeypadSpace",
	/* 206 */ "Keypad@",
	/* 207 */ "Keypad!",
	/* 208 */ "KeypadMemStore",
	/* 209 */ "KeypadMemRecall",
	/* 210 */ "KeypadMemClear",
	/* 211 */ "KeypadMemAdd",
	/* 212 */ "KeypadMemSubtract",
	/* 213 */ "KeypadMemMultiply",
	/* 214 */ "KeypadMemDivide",
	/* 215 */ "Keypad+/-",
	/* 216 */ "KeypadClear",
	/* 217 */ "KeypadClearEntry",
	/* 218 */ "KeypadBinary",
	/* 219 */ "KeypadOctal",
	/* 220 */ "KeypadDecimal",
	/* 221 */ "KeypadHexadecimal",
	/* 222 */ "",
	/* 223 */ "",
	/* 224 */ "LeftCtrl",
	/* 225 */ "LeftShift",
	/* 226 */ "LeftAlt",
	/* 227 */ "LeftGUI",
	/* 228 */ "RightCtrl",
	/* 229 */ "RightShift",
	/* 230 */ "RightAlt",
	/* 231 */ "RightGUI",
	/* 232 */ "",
	/* 233 */ "",
	/* 234 */ "",
	/* 235 */ "",
	/* 236 */ "",
	/* 237 */ "",
	/* 238 */ "",
	/* 239 */ "",
	/* 240 */ "",
	/* 241 */ "",
	/* 242 */ "",
	/* 243 */ "",
	/* 244 */ "",
	/* 245 */ "",
	/* 246 */ "",
	/* 247 */ "",
	/* 248 */ "",
	/* 249 */ "",
	/* 250 */ "",
	/* 251 */ "",
	/* 252 */ "",
	/* 253 */ "",
	/* 254 */ "",
	/* 255 */ "",
	/* 256 */ "",
	/* 257 */ "ModeSwitch",
	/* 258 */ "Sleep",
	/* 259 */ "Wake",
	/* 260 */ "ChannelUp",
	/* 261 */ "ChannelDown",
	/* 262 */ "MediaPlay",
	/* 263 */ "MediaPause",
	/* 264 */ "MediaRecord",
	/* 265 */ "MediaFastForward",
	/* 266 */ "MediaRewind",
	/* 267 */ "MediaTrackNext",
	/* 268 */ "MediaTrackPrevious",
	/* 269 */ "MediaStop",
	/* 270 */ "Eject",
	/* 271 */ "MediaPlayPause",
	/* 272 */ "MediaSelect",
	/* 273 */ "AC New",
	/* 274 */ "AC Open",
	/* 275 */ "AC Close",
	/* 276 */ "AC Exit",
	/* 277 */ "AC Save",
	/* 278 */ "AC Print",
	/* 279 */ "AC Properties",
	/* 280 */ "AC Search",
	/* 281 */ "AC Home",
	/* 282 */ "AC Back",
	/* 283 */ "AC Forward",
	/* 284 */ "AC Stop",
	/* 285 */ "AC Refresh",
	/* 286 */ "AC Bookmarks",
	/* 287 */ "SoftLeft",
	/* 288 */ "SoftRight",
	/* 289 */ "Call",
	/* 290 */ "EndCall",
}

var mousebuttons = []string{
	"",
	"MouseLeft",   /* sdl.BUTTON_LEFT */
	"MouseMiddle", /* sdl.BUTTON_MIDDLE */
	"MouseRight",  /* sdl.BUTTON_RIGHT */
	"MouseX1",     /* SDL_BUTTON_X1 */
	"MouseX2",     /* SDL_BUTTON_X2 */
}

func (k Key) Is(name string) bool {
	modifiers := []struct {
		name string
		code sdl.Keymod
	}{
		{"ctrl-", sdl.KMOD_CTRL},
		{"C-", sdl.KMOD_CTRL},
		{"shift-", sdl.KMOD_SHIFT},
		{"S-", sdl.KMOD_SHIFT},
		{"alt-", sdl.KMOD_ALT},
		{"A-", sdl.KMOD_ALT},
		{"meta-", sdl.KMOD_GUI},
		{"M-", sdl.KMOD_GUI},
	}

	mousemod := []struct {
		name string
		code MouseButton
	}{
		{"LMB-", ButtonLeft},
		{"MMB-", ButtonMiddle},
		{"RMB-", ButtonRight},
	}

	qkeymod := sdl.Keymod(0)
	qmousemod := MouseButton(0)
nameloop:
	for len(name) > 0 {
		for _, mod := range modifiers {
			if strings.HasPrefix(name, mod.name) {
				modlen := len(mod.name)
				name = name[modlen:]
				qkeymod |= mod.code
				continue nameloop
			}
		}
		for _, mod := range mousemod {
			if strings.HasPrefix(name, mod.name) {
				modlen := len(mod.name)
				name = name[modlen:]
				qmousemod |= mod.code
				continue nameloop
			}
		}
		break
	}
	if k.KeyMod&qkeymod != qkeymod {
		return false
	}
	if qmousemod > 0 && k.MouseMod&qmousemod != qmousemod {
		return false
	}
	if len(name) == 0 {
		return k.Rune == 0 && k.Mouse == 0 && k.Sym == 0
	} else if utf8.RuneCountInString(name) == 1 {
		query, _ := utf8.DecodeRuneInString(name)
		return query == k.Rune
	} else if mbn := slices.Index(mousebuttons, name); mbn != -1 {
		return k.Mouse == 1<<(mbn-1)
	} else if sym := slices.Index(scancodes, name); sym != -1 {
		return sym == int(k.Sym)
	} else {
		return false
	}
}

func (k Key) HasMod(mod sdl.Keymod) bool {
	return k.KeyMod&mod != 0
}

func (k Key) HasMouse(mod MouseButton) bool {
	return k.MouseMod&mod != 0
}
