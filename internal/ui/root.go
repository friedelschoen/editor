package ui

import (
	"image"

	"github.com/friedelschoen/glake/internal/eventregister"
	"github.com/friedelschoen/glake/internal/ui/driver"
	"github.com/friedelschoen/glake/internal/ui/widget"
)

// User Interface root (top) node.
type Root struct {
	*widget.MultiLayer
	UI              *UI
	Toolbar         *Toolbar
	MainMenuButton  *MainMenuButton
	ContextFloatBox *ContextFloatBox
	Cols            *Columns
	EvReg           eventregister.Register
}

func NewRoot(ui *UI) *Root {
	return &Root{MultiLayer: widget.NewMultiLayer(), UI: ui}
}

func (root *Root) Init() {
	//  background layer
	bgLayer := widget.NewBoxLayout()
	bgLayer.YAxis = true
	root.BgLayer.Append(bgLayer)

	// background layer
	{
		// top toolbar
		{
			ttb := widget.NewBoxLayout()
			bgLayer.Append(ttb)

			// toolbar
			root.Toolbar = NewToolbar(root.UI)
			ttb.Append(root.Toolbar)
			ttb.SetChildFlex(root.Toolbar, true, false)

			// main menu button
			mmb := NewMainMenuButton(root)
			mmb.Label.Border.Left = 1
			ttb.Append(mmb)
			ttb.SetChildFill(mmb, false, true)
			root.MainMenuButton = mmb
		}

		// columns
		root.Cols = NewColumns(root)
		bgLayer.Append(root.Cols)
	}

	root.ContextFloatBox = NewContextFloatBox(root)
}

func (l *Root) OnChildMarked(child widget.Node, newMarks widget.Marks) {
	l.MultiLayer.OnChildMarked(child, newMarks)
	// dynamic toolbar
	if l.Toolbar != nil && l.Toolbar.HasAnyMarks(widget.MarkNeedsLayout) {
		l.BgLayer.MarkNeedsLayout()
	}
}

func (l *Root) OnInputEvent(ev0 driver.Event, p image.Point) bool {
	return false
}

const (
	RootSelectAnnotationEventId = iota
)

type RootSelectAnnotationEvent struct {
	Type RootSelectAnnotationType
}

type RootSelectAnnotationType int

const (
	RootSelAnnTypeFirst RootSelectAnnotationType = iota
	RootSelAnnTypeLast
	RootSelAnnTypePrev
	RootSelAnnTypeNext
	RootSelAnnTypeClear
)
