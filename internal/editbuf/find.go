package editbuf

import (
	"context"
	"errors"
	"io"

	"github.com/friedelschoen/glake/internal/ioutil"
)

func Find(cctx context.Context, ectx *EditorBuffer, str string, reverse bool, opt *ioutil.IndexOpt) (bool, error) {
	if str == "" {
		return false, nil
	}
	if reverse {
		i, n, err := find2Rev(cctx, ectx, []byte(str), opt)
		if err != nil || i < 0 {
			return false, err
		}
		ectx.C.SetSelection(i+n, i) // cursor at start to allow searching next
	} else {
		i, n, err := find2(cctx, ectx, []byte(str), opt)
		if err != nil || i < 0 {
			return false, err
		}
		ectx.C.SetSelection(i, i+n) // cursor at end to allow searching next
	}

	return true, nil
}
func find2(cctx context.Context, ectx *EditorBuffer, b []byte, opt *ioutil.IndexOpt) (int, int, error) {
	ci := ectx.C.Index()
	// index to end
	i, n, err := ioutil.IndexCtx(cctx, ectx.RW, ci, b, opt)
	if err != nil || i >= 0 {
		return i, n, err
	}
	// start to index
	e := ci + len(b) - 1
	rd := ioutil.NewLimitedReaderAt(ectx.RW, 0, e)
	k, n, err := ioutil.IndexCtx(cctx, rd, 0, b, opt)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return -1, 0, nil
		}
		return -1, 0, err
	}
	return k, n, nil
}
func find2Rev(cctx context.Context, ectx *EditorBuffer, b []byte, opt *ioutil.IndexOpt) (int, int, error) {
	ci := ectx.C.Index()
	// start to index (in reverse)
	i, n, err := ioutil.LastIndexCtx(cctx, ectx.RW, ci, b, opt)
	if err != nil || i >= 0 {
		return i, n, err
	}
	// index to end (in reverse)
	s := ci - len(b) + 1
	e := ectx.RW.Max()
	rd2 := ioutil.NewLimitedReaderAt(ectx.RW, s, e)
	k, n, err := ioutil.LastIndexCtx(cctx, rd2, e, b, opt)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return -1, 0, nil
		}
		return -1, 0, err
	}
	return k, n, nil
}
