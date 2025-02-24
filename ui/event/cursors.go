package event

type Cursor int

const (
	DefaultCursor Cursor = iota // none means not set
	NSResizeCursor
	WEResizeCursor
	CloseCursor
	MoveCursor
	PointerCursor
	BeamCursor // text cursor
	WaitCursor // watch cursor
)
