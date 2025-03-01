package reslocparser

type ResLoc struct {
	Path string // raw path

	Line   int // 0 is nil
	Column int // 0 is nil
	Offset int // -1 is nil

	PathSep rune
	Escape  rune

	Scheme string // ex: "file://", useful to know when to translate to another path separator
	Volume string

	Pos, End int // contains reverse expansion
}

func NewResLoc() *ResLoc {
	return &ResLoc{Offset: -1}
}

//func (rl *ResLoc) ToString() string {
//	return rl.ToString2(false)
//}
//func (rl *ResLoc) ToString2(preferOffset bool) string {
//	s := rl.ClearFilename1()
//	u := ""
//	if preferOffset && rl.Offset >= 0 {
//		u = rl.offsetToString()
//	} else if rl.Line > 0 {
//		u = rl.linecolToString()
//	} else if rl.Offset >= 0 {
//		u = rl.offsetToString()
//	}
//	if u != "" {
//		u = ":" + u
//	}
//	return s + u
//}
