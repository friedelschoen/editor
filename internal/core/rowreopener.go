package core

import (
	"github.com/friedelschoen/glake/internal/ui"
)

const (
	QueueMax = 5
)

type RowReopener struct {
	ed    *Editor
	queue []*RowState
}

func NewRowReopener(ed *Editor) *RowReopener {
	return &RowReopener{ed: ed, queue: make([]*RowState, 0, QueueMax+1),}
}

func (rr *RowReopener) Add(row *ui.Row) {
	rstate := NewRowState(rr.ed, row)

	rr.queue = append(rr.queue, rstate)

	// limit entries
	for len(rr.queue) > QueueMax {
		rr.queue = rr.queue[1:]
	}
}
func (rr *RowReopener) Reopen() {
	if len(rr.queue) == 0 {
		rr.ed.Errorf("no rows to reopen")
		return
	}

	// pop state from queue
	rstate := rr.queue[0]
	rr.queue = rr.queue[1:]

	rowPos := rr.ed.GoodRowPos()
	erow, ok, err := rstate.OpenERow(rr.ed, rowPos)
	if err != nil {
		rr.ed.Error(err)
	}
	if !ok {
		return
	}
	rstate.RestorePos(erow)
	erow.Flash()
}
