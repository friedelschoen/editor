package ui

import (
	"image"

	"github.com/jmigpin/editor/util/uiutil/event"
	"github.com/jmigpin/editor/util/uiutil/widget"
)

type ContextFloatBox struct {
	*widget.FloatBox

	root     *Root
	sa       *widget.ScrollArea
	TextArea *TextArea
}

func NewContextFloatBox(root *Root) *ContextFloatBox {
	cfb := &ContextFloatBox{root: root}

	cfb.TextArea = NewTextArea(root.UI)
	cfb.SetStrClearHistory("")

	cfb.sa = widget.NewScrollArea(root.UI, cfb.TextArea, false, true)
	cfb.sa.LeftScroll = ScrollBarLeft

	border := widget.NewBorder(root.UI, cfb.sa)
	border.SetAll(1)

	container := WrapInBottomShadowOrNone(root.UI, border)

	cfb.FloatBox = widget.NewFloatBox(root.MultiLayer, container)
	cfb.FloatBox.MaxSize = image.Point{800, 100000}
	root.MultiLayer.ContextLayer.Append(cfb)
	cfb.FloatBox.Hide()

	cfb.SetThemePaletteNamePrefix("contextfloatbox_")

	return cfb
}

func (cfb *ContextFloatBox) SetStrClearHistory(s string) {
	if s == "" {
		s = "No content provided."
	}
	cfb.TextArea.SetStrClearHistory(s)
}

func (cfb *ContextFloatBox) Layout() {
	ff := cfb.TextArea.TreeThemeFontFace()
	cfb.sa.ScrollWidth = UIThemeUtil.GetScrollBarWidth(ff)
	cfb.FloatBox.Layout()
}

func (cfb *ContextFloatBox) OnInputEvent(ev any, p image.Point) event.Handled {
	switch ev.(type) {
	case *event.KeyUp,
		*event.KeyDown:
		// let lower layers get events
		return false
	}
	return true
}

func (cfb *ContextFloatBox) AutoClose(ev any, p image.Point) {
	if cfb.Visible() && !p.In(cfb.Bounds) {
		switch ev.(type) {
		case *event.KeyDown,
			*event.MouseDown:
			cfb.Hide()
			return
		case *event.MouseMove:
		default:
			//fmt.Printf("%T\n", ev)
		}
	}
}

func (cfb *ContextFloatBox) Toggle() {
	visible := cfb.Visible()
	if !visible {
		cfb.Show()
	} else {
		cfb.Hide()
	}
}

func (cfb *ContextFloatBox) SetRefPointToTextAreaCursor(ta *TextArea) {
	p := ta.GetPoint(ta.CursorIndex())
	p.Y += ta.LineHeight()
	cfb.RefPoint = p
	// compensate scrollwidth for a better position
	if cfb.sa.LeftScroll {
		cfb.RefPoint.X -= cfb.sa.ScrollWidth
	}
}

func (cfb *ContextFloatBox) FindTextAreaUnderPointer() (*TextArea, bool) {
	// pointer position
	p, err := cfb.root.UI.QueryPointer()
	if err != nil {
		return nil, false
	}
	ta := cfb.visitToFindTA(p, cfb.root)
	return ta, ta != nil
}

func (cfb *ContextFloatBox) visitToFindTA(p image.Point, node widget.Node) (ta *TextArea) {
	if p.In(node.Embed().Bounds) {
		if u, ok := node.(*TextArea); ok {
			return u
		}
		if u, ok := node.(*Toolbar); ok {
			return u.TextArea
		}
		if u, ok := node.(*RowToolbar); ok {
			return u.TextArea
		}
	}
	node.Embed().IterateWrappersReverse(func(n widget.Node) bool {
		u := cfb.visitToFindTA(p, n)
		if u != nil {
			ta = u
			return false
		}
		return true
	})
	return ta
}
