package ui

import (
	"fmt"
	"image"
	"log"
	"sync"
	"time"

	"github.com/friedelschoen/glake/internal/ui/driver"
	"github.com/friedelschoen/glake/internal/ui/widget"
	"github.com/veandco/go-sdl2/sdl"
)

type UI struct {
	*driver.Window

	DrawFrameRate int // frames per second

	curCursor sdl.SystemCursor

	closeOnce sync.Once

	applyEv *widget.ApplyEvent

	pendingPaint   bool
	lastPaintStart time.Time

	Root    *Root
	OnError func(error)
}

func NewUI(winName string) (*UI, error) {
	ui := &UI{
		DrawFrameRate: 37,
	}

	ui.Root = NewRoot(ui)

	// bui, err := uiutil.NewUI(winName, ui.Root)

	var err error
	ui.Window, err = driver.NewWindow()
	if err != nil {
		return nil, err
	}

	if err := ui.WindowSetName(winName); err != nil {
		return nil, err
	}

	ui.applyEv = widget.NewApplyEvent(ui)

	// Embed nodes have their wrapper nodes set when they are appended to another node. The root node is not appended to any other node, therefore it needs to be set here.
	ui.Root.Embed().SetWrapperForRoot(ui.Root)

	// set theme before root init
	c1 := &ColorThemeCycler
	c1.Set(c1.CurName, ui.Root)

	loadThemeFont(CurrentFont, ui.Root)

	// build ui - needs ui.UI to be set
	ui.Root.Init()

	return ui, nil
}

func (ui *UI) HandleEvent(ev driver.Event) (handled bool) {
	if ev == nil {
		return true
	}
	switch t := ev.(type) {
	case *driver.WindowResize:
		ui.resizeImage(t.Rect)
	case *driver.WindowExpose:
		fmt.Println("exposed!")
		ui.Root.Embed().MarkNeedsPaint()
	case *UIRunFuncEvent:
		t.Func()
	default:
		ui.handleWindowInput(t)
	}
	return true
}

func (ui *UI) handleWindowInput(wi driver.Event) {
	var p image.Point
	if inevt, ok := wi.(driver.InputEvent); ok {
		p = inevt.At()
	} else {
		x, y, _ := sdl.GetMouseState()
		p = image.Point{int(x), int(y)}
	}
	ui.applyEv.Apply(ui.Root, wi, p)
}

func (ui *UI) LayoutMarkedAndSchedulePaint() {
	ui.Root.LayoutMarked()
	ui.schedulePaintMarked()
}

func (ui *UI) resizeImage(r image.Rectangle) {
	if err := ui.Resize(r); err != nil {
		log.Println(err)
		return
	}

	ib := ui.Image().Bounds()
	en := ui.Root.Embed()
	if !en.Bounds.Eq(ib) {
		en.Bounds = ib
		en.MarkNeedsLayout()
		en.MarkNeedsPaint()
	}
}

func (ui *UI) schedulePaintMarked() {
	if ui.Root.Embed().TreeNeedsPaint() {
		ui.schedulePaint()
	}
}
func (ui *UI) schedulePaint() {
	if ui.pendingPaint {
		return
	}
	ui.pendingPaint = true
	// schedule
	time.AfterFunc(ui.durationToNextPaint(), func() {
		ui.PushEvent(&UIRunFuncEvent{ui.paint})
	})
}

func (ui *UI) durationToNextPaint() time.Duration {
	now := time.Now()
	frameDur := time.Second / time.Duration(ui.DrawFrameRate)
	d := now.Sub(ui.lastPaintStart)
	return frameDur - d
}

func (ui *UI) paint() {
	// DEBUG: print fps
	now := time.Now()
	//d := now.Sub(ui.lastPaintStart)
	//fmt.Printf("paint: fps %v\n", int(time.Second/d))
	ui.lastPaintStart = now

	ui.paintMarked()
}

func (ui *UI) paintMarked() {
	ui.pendingPaint = false
	u := ui.Root.PaintMarked()
	r := u.Intersect(ui.Image().Bounds())
	if !r.Empty() {
		ui.Update()
	}
}

func (ui *UI) EnqueueNoOpEvent() {
	ui.PushEvent(&UIRunFuncEvent{})
}

func (ui *UI) RunOnUIGoRoutine(f func()) {
	ui.PushEvent(&UIRunFuncEvent{f})
}

// Use with care to avoid UI deadlock (waiting within another wait).
func (ui *UI) WaitRunOnUIGoRoutine(f func()) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ui.RunOnUIGoRoutine(func() {
		defer wg.Done()
		f()
	})
	wg.Wait()
}

// Allows triggering a run of applyevent (ex: useful for cursor update, needs point or it won't work).
func (ui *UI) QueueEmptyWindowInputEvent() {
	p, err := ui.QueryPointer()
	if err != nil {
		return
	}
	ui.PushEvent(&driver.MouseClick{Point: p})
}

func (ui *UI) WarpPointerToRectanglePad(r image.Rectangle) {
	p, err := ui.QueryPointer()
	if err != nil {
		return
	}

	pad := 5

	set := func(v *int, min, max int) {
		if max-min < pad*2 {
			*v = min + (max-min)/2
		} else {
			if *v < min+pad {
				*v = min + pad
			} else if *v > max-pad {
				*v = max - pad
			}
		}
	}

	set(&p.X, r.Min.X, r.Max.X)
	set(&p.Y, r.Min.Y, r.Max.Y)

	ui.WarpPointer(p)
}

func (ui *UI) resizeRowToGoodSize(row *Row) {
	if row.PrevSibling() == nil {
		return
	}
	prevRow := row.PrevSiblingWrapper().(*Row)
	b := ui.rowInsertionBounds(prevRow)
	col := row.Col
	colDy := col.Bounds.Dy()
	perc := float64(b.Min.Y-col.Bounds.Min.Y) / float64(colDy)
	col.RowsLayout.Spl.Resize(row, perc)
}

func (ui *UI) GoodRowPos() *RowPos {
	var best struct {
		area    int
		col     *Column
		nextRow *Row
	}

	// default position if nothing better is found
	best.col = ui.Root.Cols.FirstChildColumn()

	for _, c := range ui.Root.Cols.Columns() {
		rows := c.Rows()

		// space before first row
		s := c.Bounds.Size()
		if len(rows) > 0 {
			s.Y = rows[0].Bounds.Min.Y - c.Bounds.Min.Y
		}
		a := s.X * s.Y
		if a > best.area {
			best.area = a
			best.col = c
			best.nextRow = nil
			if len(rows) > 0 {
				best.nextRow = rows[0]
			}
		}

		// space between rows
		for _, r := range rows {
			b := ui.rowInsertionBounds(r)
			s := b.Size()
			a := s.X * s.Y
			if a > best.area {
				best.area = a
				best.col = c
				best.nextRow = r.NextRow()
			}
		}
	}

	return NewRowPos(best.col, best.nextRow)
}

func (ui *UI) rowInsertionBounds(prevRow *Row) image.Rectangle {
	ta := prevRow.TextArea
	if r2, ok := ui.boundsAfterVisibleCursor(ta); ok {
		return *r2
	} else if r3, ok := ui.boundsOfTwoThirds(ta); ok {
		return *r3
	} else {
		b := prevRow.Bounds
		b.Max = b.Max.Sub(b.Size().Div(2)) // half size from StartPercentLayout
		return b
	}
}

func (ui *UI) boundsAfterVisibleCursor(ta *TextArea) (*image.Rectangle, bool) {
	ci := ta.CursorIndex()
	if !ta.IndexVisible(ci) {
		return nil, false
	}
	p := ta.GetPoint(ci)
	lh := ta.LineHeight()
	r := ta.Bounds
	r.Min.Y = p.Y + lh
	r = ta.Bounds.Intersect(r)
	if r.Dy() < lh*2 {
		return nil, false
	}
	return &r, true
}

func (ui *UI) boundsOfTwoThirds(ta *TextArea) (*image.Rectangle, bool) {
	lh := ta.LineHeight()
	if ta.Bounds.Size().Y < lh {
		return nil, false
	}
	b := ta.Bounds
	b.Min.Y = b.Max.Y - (ta.Bounds.Dy() * 2 / 3)
	r := ta.Bounds.Intersect(b)
	return &r, true
}

func (ui *UI) Error(err error) {
	if ui.OnError != nil {
		ui.OnError(err)
	}
}

type RowPos struct {
	Column  *Column
	NextRow *Row
}

func NewRowPos(col *Column, nextRow *Row) *RowPos {
	return &RowPos{col, nextRow}
}

type UIRunFuncEvent struct {
	Func func()
}

func (UIRunFuncEvent) IsEvent() {}
