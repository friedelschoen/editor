package ui

import (
	"github.com/friedelschoen/editor/internal/ui/widget"
)

type MainMenuButton struct {
	*widget.FloatBoxButton
	sa      *widget.ScrollArea
	Toolbar *Toolbar
}

func NewMainMenuButton(root *Root) *MainMenuButton {
	mmb := &MainMenuButton{}

	content := &widget.ENode{}

	mmb.FloatBoxButton = widget.NewFloatBoxButton(root.UI, root.MultiLayer, root.MenuLayer, content)
	mmb.FloatBoxButton.Label.Text.SetStr(string(rune(8801))) // 3 lines rune
	mmb.FloatBoxButton.Label.Pad.Left = 5
	mmb.FloatBoxButton.Label.Pad.Right = 5

	// theme
	mmb.SetThemePaletteNamePrefix("mm_")
	content.SetThemePaletteNamePrefix("mm_content_")

	// float content
	mmb.Toolbar = NewToolbar(root.UI)
	mmb.Toolbar.Drawer.Opt.EarlyExitMeasure = false // full measure to avoid flicker (want the menu size stable)
	mmb.sa = widget.NewScrollArea(root.UI, mmb.Toolbar, false, true)
	mmb.sa.LeftScroll = ScrollBarLeft
	border := widget.NewBorder(root.UI, mmb.sa)
	border.SetAll(1)
	n1 := WrapInBottomShadowOrNone(root.UI, border)
	content.Append(n1)

	return mmb
}
