package drawutil

type SyntaxComment struct {
	Start string
	End   string // empty for single line comment
}

func (syc *SyntaxComment) IsLine() bool {
	return syc.End == ""
}

//----------

type RangeAlignment int

const (
	RAlignKeep         RangeAlignment = iota
	RAlignKeepOrBottom                // keep if visible, or bottom
	RAlignAuto
	RAlignTop
	RAlignBottom
	RAlignCenter
)
