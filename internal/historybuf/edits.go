package historybuf

import (
	"github.com/friedelschoen/editor/internal/editbuf"
	"github.com/friedelschoen/editor/internal/ioutil"
)

////godebug:annotatefile

type Edits struct {
	list       []*UndoRedo
	preCursor  editbuf.SimpleCursor
	postCursor editbuf.SimpleCursor
}

func (edits *Edits) Append(ur *UndoRedo) {
	// set pre cursor once
	if len(edits.list) == 0 {
		if len(ur.D) > 0 {
			edits.preCursor.SetSelection(ur.Index, ur.Index+len(ur.D))
		} else {
			edits.preCursor.SetIndex(ur.Index)
		}
	}

	edits.list = append(edits.list, ur)

	// renew post cursor on each append
	if len(ur.I) > 0 {
		edits.postCursor.SetSelection(ur.Index, ur.Index+len(ur.I))
	} else {
		edits.postCursor.SetIndexSelectionOff(ur.Index)
	}
}

func (edits *Edits) MergeEdits(edits2 *Edits) {
	// append list
	for _, ur := range edits.list {
		edits.Append(ur)
	}
	// merge cursor position
	if len(edits.list) == 0 {
		edits.preCursor = edits2.preCursor
	}
	edits.postCursor = edits2.postCursor
}

func (edits *Edits) WriteUndoRedo(redo bool, w ioutil.WriterAt) (editbuf.SimpleCursor, error) {
	if redo {
		for _, ur := range edits.list {
			if err := ur.Apply(redo, w); err != nil {
				return editbuf.SimpleCursor{}, err
			}
		}
		return edits.postCursor, nil
	} else {
		for i := len(edits.list) - 1; i >= 0; i-- {
			ur := edits.list[i]
			if err := ur.Apply(redo, w); err != nil {
				return editbuf.SimpleCursor{}, err
			}
		}
		return edits.preCursor, nil
	}
}

func (edits *Edits) Entries() []*UndoRedo {
	return edits.list
}

func (edits *Edits) Empty() bool {
	for _, ur := range edits.list {
		if len(ur.D) > 0 || len(ur.I) > 0 {
			return false
		}
	}
	return true
}
