package historybuf

import (
	"github.com/friedelschoen/glake/internal/editbuf"
	"github.com/friedelschoen/glake/internal/ioutil"
)

type RWUndo struct {
	ioutil.ReadWriterAt
	History *History
}

func NewRWUndo(rw ioutil.ReadWriterAt, hist *History) *RWUndo {
	rwu := &RWUndo{ReadWriterAt: rw, History: hist}
	return rwu
}

func (rw *RWUndo) OverwriteAt(i, n int, p []byte) error {
	// don't add to history if the result is equal
	changed := true
	if eq, err := ioutil.REqual(rw, i, n, p); err == nil && eq {
		changed = false
	}

	ur, err := NewUndoRedoOverwrite(rw.ReadWriterAt, i, n, p)
	if err != nil {
		return err
	}

	if changed {
		edits := &Edits{}
		edits.Append(ur)
		rw.History.Append(edits)
	}
	return nil
}

func (rw *RWUndo) UndoRedo(redo, peek bool) (editbuf.SimpleCursor, bool, error) {
	edits, ok := rw.History.UndoRedo(redo, peek)
	if !ok {
		return editbuf.SimpleCursor{}, false, nil
	}
	c, err := edits.WriteUndoRedo(redo, rw.ReadWriterAt)
	if err != nil {
		// TODO: restore the undo/redo since it was not successful?
		return editbuf.SimpleCursor{}, false, err
	}
	return c, true, nil
}

// used in tests
