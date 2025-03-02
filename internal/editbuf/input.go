package editbuf

import (
	"errors"
	"io"

	"github.com/friedelschoen/glake/internal/ui/driver"
)

//godebug:annotatefile

func HandleInput(ctx *EditorBuffer, ev any) (bool, error) {
	in := &Input{ctx, ev}
	return in.handle()
}

type Input struct {
	ctx *EditorBuffer
	ev  any
}

func (in *Input) handle() (bool, error) {
	switch ev := in.ev.(type) {
	case *driver.KeyDown:
		return in.onKeyDown(ev)
	case *driver.MouseDown:
		return in.onMouseDown(ev)
	case *driver.MouseDragMove:
		return in.onMouseDragMove(ev)
	case *driver.MouseDragEnd:
		return in.onMouseDragEnd(ev)
	case *driver.MouseClick:
		return in.onMouseClick(ev)
	case *driver.MouseWheel:
		return in.onMouseWheel(ev)
	}
	return false, nil
}

func (in *Input) onMouseDown(ev *driver.MouseDown) (bool, error) {
	switch {
	case ev.Key.Is("MouseLeft"):
		MoveCursorToPoint(in.ctx, ev.Point, false)
		return true, nil
	case ev.Key.Is("shift-MouseLeft"):
		MoveCursorToPoint(in.ctx, ev.Point, true)
		return true, nil
	}
	return false, nil
}

func (in *Input) onMouseWheel(ev *driver.MouseWheel) (bool, error) {
	if ev.Y < 0 {
		ScrollUp(in.ctx, true)
	} else if ev.Y > 0 {
		ScrollUp(in.ctx, false)
	}
	return true, nil
}

func (in *Input) onMouseDragMove(ev *driver.MouseDragMove) (bool, error) {
	if ev.Key.Mouse == driver.ButtonLeft {
		MoveCursorToPoint(in.ctx, ev.Point, true)
		return true, nil
	}
	return false, nil
}
func (in *Input) onMouseDragEnd(ev *driver.MouseDragEnd) (bool, error) {
	if ev.Key.Mouse == driver.ButtonLeft {
		MoveCursorToPoint(in.ctx, ev.Point, true)
		return true, nil
	}
	return false, nil
}

func (in *Input) onMouseClick(ev *driver.MouseClick) (bool, error) {
	switch ev.Count {
	case 1:
		if ev.Key.Mouse == driver.ButtonMiddle {
			MoveCursorToPoint(in.ctx, ev.Point, false)
			Paste(in.ctx)
			return true, nil
		}
	case 2:
		if ev.Key.Mouse == driver.ButtonLeft {
			MoveCursorToPoint(in.ctx, ev.Point, false)
			err := SelectWord(in.ctx)
			// can select at EOF but avoid error msg
			if errors.Is(err, io.EOF) {
				err = nil
			}

			return true, err
		}
	case 3:
		if ev.Key.Mouse == driver.ButtonLeft {
			MoveCursorToPoint(in.ctx, ev.Point, false)
			err := SelectLine(in.ctx)
			return true, err
		}
	}
	return false, nil
}

func (in *Input) onKeyDown(ev *driver.KeyDown) (bool, error) {
	var err error
	makeCursorVisible := func() {
		if err == nil {
			in.ctx.Fns.MakeIndexVisible(in.ctx.C.Index())
		}
	}

	switch {
	case ev.Key.Is("C-S-Right"):
		err = MoveCursorJumpRight(in.ctx, true)
		makeCursorVisible()
	case ev.Key.Is("C-Right"):
		err = MoveCursorJumpRight(in.ctx, false)
		makeCursorVisible()
	case ev.Key.Is("S-Right"):
		err = MoveCursorRight(in.ctx, true)
		makeCursorVisible()
	case ev.Key.Is("Right"):
		err = MoveCursorRight(in.ctx, false)
		makeCursorVisible()

	case ev.Key.Is("C-S-Left"):
		err = MoveCursorJumpLeft(in.ctx, true)
		makeCursorVisible()
	case ev.Key.Is("C-Left"):
		err = MoveCursorJumpLeft(in.ctx, false)
		makeCursorVisible()
	case ev.Key.Is("S-Left"):
		err = MoveCursorLeft(in.ctx, true)
		makeCursorVisible()
	case ev.Key.Is("Left"):
		err = MoveCursorLeft(in.ctx, false)
		makeCursorVisible()

	case ev.Key.Is("C-A-Up"):
		err = MoveLineUp(in.ctx)
		makeCursorVisible()
	case ev.Key.Is("C-S-Up"), ev.Key.Is("S-Up"):
		MoveCursorUp(in.ctx, true)
		makeCursorVisible()
	case ev.Key.Is("Up"):
		MoveCursorUp(in.ctx, false)
		makeCursorVisible()

	case ev.Key.Is("C-A-Down"):
		err = MoveLineDown(in.ctx)
		makeCursorVisible()
	case ev.Key.Is("C-S-Down"), ev.Key.Is("S-Down"):
		MoveCursorDown(in.ctx, true)
		makeCursorVisible()
	case ev.Key.Is("Down"):
		MoveCursorDown(in.ctx, false)
		makeCursorVisible()

	case ev.Key.Is("C-S-Home"):
		StartOfString(in.ctx, true)
		makeCursorVisible()
	case ev.Key.Is("C-Home"):
		StartOfString(in.ctx, false)
		makeCursorVisible()
	case ev.Key.Is("S-Home"):
		err = StartOfLine(in.ctx, true)
		makeCursorVisible()
	case ev.Key.Is("Home"):
		err = StartOfLine(in.ctx, false)
		makeCursorVisible()

	case ev.Key.Is("C-S-End"):
		EndOfString(in.ctx, true)
		makeCursorVisible()
	case ev.Key.Is("C-End"):
		EndOfString(in.ctx, false)
		makeCursorVisible()
	case ev.Key.Is("S-End"):
		err = EndOfLine(in.ctx, true)
		makeCursorVisible()
	case ev.Key.Is("End"):
		err = EndOfLine(in.ctx, false)
		makeCursorVisible()

	case ev.Key.Is("Backspace"):
		err = Backspace(in.ctx)
		makeCursorVisible()

	case ev.Key.Is("Delete"):
		err = Delete(in.ctx)
		makeCursorVisible() // TODO: on delete?

	case ev.Key.Is("Return"):
		err = AutoIndent(in.ctx)
		makeCursorVisible()

	case ev.Key.Is("S-Tab"):
		// TODO: using KSymTabLeft case, this still needed?
		err = TabLeft(in.ctx)
		makeCursorVisible()
	case ev.Key.Is("Tab"):
		err = TabRight(in.ctx)
		makeCursorVisible()

	case ev.Key.Is("PageUp"):
		PageUp(in.ctx, true)
	case ev.Key.Is("PageDown"):
		PageUp(in.ctx, false)

	case ev.Key.Is("ctrl-D"):
		err = Comment(in.ctx)
	case ev.Key.Is("ctrl-C"):
		err = Copy(in.ctx)
	case ev.Key.Is("ctrl-X"):
		err = Cut(in.ctx)
	case ev.Key.Is("ctrl-V"):
		Paste(in.ctx)
	case ev.Key.Is("ctrl-K"):
		err = RemoveLines(in.ctx)
	case ev.Key.Is("ctrl-A"):
		err = SelectAll(in.ctx)
	case ev.Key.Is("ctrl-Z"):
		err = Undo(in.ctx)
	case ev.Key.Is("ctrl-shift-D"):
		err = Uncomment(in.ctx)
	case ev.Key.Is("ctrl-shift-Z"):
		err = Redo(in.ctx)

	case ev.Key.Rune != 0:
		err = InsertString(in.ctx, string(ev.Key.Rune))
		makeCursorVisible()

	default:
		return false, nil
	}
	return true, err
}
