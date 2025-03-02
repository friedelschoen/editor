package core

import (
	"context"
	"strings"
	"sync"
	"unicode"

	"github.com/friedelschoen/glake/internal/drawer"
	"github.com/friedelschoen/glake/internal/io/iorw"
	"github.com/friedelschoen/glake/internal/mathutil"
	"github.com/friedelschoen/glake/internal/ui"
)

type InlineComplete struct {
	ed *Editor

	mu struct {
		sync.Mutex
		cancel context.CancelFunc
		ta     *ui.TextArea // if not nil, inlinecomplete is on
		index  int          // cursor index
	}
}

func NewInlineComplete(ed *Editor) *InlineComplete {
	ic := &InlineComplete{ed: ed}
	ic.mu.cancel = func() {} // avoid nil call (called in many places)
	return ic
}

func (ic *InlineComplete) Complete(erow *ERow, ev *ui.TextAreaInlineCompleteEvent) bool {

	// early pre-check if filename is supported
	_, err := ic.ed.LSProtoMan.LangManager(erow.Info.Name())
	if err != nil {
		return false // not handled
	}

	ta := ev.TextArea

	ic.mu.Lock()
	defer ic.mu.Unlock()

	ic.mu.cancel() // cancel previous run

	// clear annotations at other textarea
	if ic.mu.ta != nil && ic.mu.ta != ta {
		// run async to avoid lockup
		go ic.setAnnotations(ic.mu.ta, nil)
	}

	ctx, cancel := context.WithCancel(erow.ctx)
	ic.mu.cancel = cancel
	ic.mu.ta = ta
	ic.mu.index = ta.CursorIndex()

	go ic.complete2(ctx, erow.Info.Name(), ta, ev)
	return true
}

func (ic *InlineComplete) complete2(ctx context.Context, filename string, ta *ui.TextArea, ev *ui.TextAreaInlineCompleteEvent) {
	cleanup := func() {
		ic.mu.cancel()
		ic.ed.UI.EnqueueNoOpEvent()
	}
	handleErr := func(err error) {
		ic.setAnnotations(ta, nil)
		ic.ed.Error(err)
	}

	comps, err := ic.lsprotoCompletions(ctx, filename, ta)
	if err != nil {
		defer cleanup()
		handleErr(err)
		return
	}

	// insert completions uses BeginUndoGroup, needs to run in sync
	ic.ed.UI.RunOnUIGoRoutine(func() {
		defer cleanup()
		if err := ic.insertCompletions(ta, ev, comps); err != nil {
			handleErr(err)
		}
	})
}
func (ic *InlineComplete) insertCompletions(ta *ui.TextArea, ev *ui.TextAreaInlineCompleteEvent, comps []string) error {
	// insert complete
	completed, comps, err := ic.insertCompletions2(comps, ta)
	if err != nil {
		return err
	}

	switch len(comps) {
	case 0:
		ic.setAnnotationsMsg(ta, "0 results")
	case 1:
		if completed {
			ic.setAnnotations(ta, nil)
		} else {
			ic.setAnnotationsMsg(ta, "already complete")
		}
	default:
		// show completions
		entries := drawer.NewAnnotationGroup(len(comps))
		for i, v := range comps {
			u := &drawer.Annotation{Offset: ev.Offset, Bytes: []byte(v)}
			entries.Anns[i] = u
		}
		ic.setAnnotations(ta, entries)
	}
	return nil
}

func (ic *InlineComplete) insertCompletions2(comps []string, ta *ui.TextArea) (completed bool, _ []string, _ error) {
	ta.BeginUndoGroup()
	defer ta.EndUndoGroup()

	newIndex, completed, comps2, err := insertComplete(comps, ta.RW(), ta.CursorIndex())
	if err != nil {
		return completed, comps2, err
	}
	//if newIndex != 0 {
	if completed {
		ta.SetCursorIndex(newIndex)
		// update index for CancelOnCursorChange
		ic.mu.Lock()
		ic.mu.index = newIndex
		ic.mu.Unlock()
	}
	return completed, comps2, err
}

func (ic *InlineComplete) lsprotoCompletions(ctx context.Context, filename string, ta *ui.TextArea) ([]string, error) {
	compList, err := ic.ed.LSProtoMan.TextDocumentCompletion(ctx, filename, ta.RW(), ta.CursorIndex())
	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, ci := range compList.Items {
		// trim labels (clangd: has some entries prefixed with space)
		label := strings.TrimSpace(ci.Label)

		res = append(res, label)
	}

	//// NOTE: this loses the provided order
	//sort.Strings(res)

	return res, nil
}

func (ic *InlineComplete) setAnnotationsMsg(ta *ui.TextArea, s string) {
	offset := ta.CursorIndex()
	entries := drawer.NewAnnotationGroup(1)
	entries.Anns = []*drawer.Annotation{{Offset: offset, Bytes: []byte(s)}}
	ic.setAnnotations(ta, entries)
}

func (ic *InlineComplete) setAnnotations(ta *ui.TextArea, entries *drawer.AnnotationGroup) {
	on := entries != nil && len(entries.Anns) > 0
	if !on {
		ic.setOff(ta)
	}
	ic.ed.SetAnnotations(EareqInlineComplete, ta, on, -1, entries)
}

func (ic *InlineComplete) IsOn(ta *ui.TextArea) bool {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	return ic.mu.ta != nil && ic.mu.ta == ta
}

func (ic *InlineComplete) setOff(ta *ui.TextArea) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	if ic.mu.ta == ta {
		ic.mu.ta = nil
		// possible early cancel for this textarea
		ic.mu.cancel()
	}
}

func (ic *InlineComplete) CancelAndClear() {
	ic.mu.Lock()
	ta := ic.mu.ta
	ic.mu.Unlock()
	if ta != nil {
		ic.setAnnotations(ta, nil)
	}
}

//func (ic *InlineComplete) CancelOnCursorChange() {
//	ic.mu.Lock()
//	ta := ic.mu.ta
//	index := ic.mu.index
//	ic.mu.Unlock()
//	if ta != nil {
//		if index != ta.CursorIndex() {
//			ic.setAnnotations(ta, nil)
//		}
//	}
//}

func insertComplete(comps []string, rw iorw.ReadWriterAt, index int) (newIndex int, completed bool, _ []string, _ error) {
	// build prefix from start of string
	start, prefix, ok := readLastUntilStart(rw, index)
	if !ok {
		return 0, false, comps, nil
	}

	expandStr, comps2 := expandAndFilter(prefix, comps)
	if len(comps2) == 0 {
		return 0, false, comps2, nil
	}
	canComplete := expandStr != ""

	if canComplete {
		// try to expand the index to the existing text
		n := len(prefix)
		expand := len(expandStr) - len(prefix)
		for i := 0; i < expand; i++ {
			b, err := rw.ReadFastAt(index+i, 1)
			if err != nil {
				break
			}
			if b[0] != expandStr[n] {
				break
			}
			n++
		}

		// insert completion
		if expandStr != prefix {
			err := rw.OverwriteAt(start, n, []byte(expandStr))
			if err != nil {
				return 0, false, nil, err
			}
			newIndex = start + len(expandStr)
			return newIndex, true, comps2, nil
		}
	}

	return 0, false, comps2, nil
}

func expandAndFilter(prefix string, comps []string) (expand string, _ []string) {
	// find prefix matches (case insensitive)
	strLow := strings.ToLower(prefix)
	comps2 := []string{}
	for _, v := range comps {
		vLow := strings.ToLower(v)
		if strings.HasPrefix(vLow, strLow) {
			comps2 = append(comps2, v)
		}
	}
	if len(comps2) == 0 {
		return "", nil
	}

	// longest prefix
	lcp := longestCommonPrefix(comps2)

	// choose next in line: keep first not eq to prefix after the one eq to prefix
	if len(lcp) == len(prefix) {
		// find first equal to prefix
		n := len(prefix)
		j := 0
		for i, s := range comps2 {
			if s[:n] == prefix {
				j = i
				break
			}
		}
		// find next in line not eq to prefix
		k := 0 // default
		for i := 0; i < len(comps2); i++ {
			u := (j + i) % len(comps2)
			s := comps2[u]
			if s[:n] != prefix {
				k = u
				break
			}
		}
		lcp = comps2[k][:n]
	}

	return lcp, comps2
}
func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strings.ToLower(strs[0])
	for i := 1; i < len(strs); i++ {
		str := strings.ToLower(strs[i])
		n := mathutil.Min(len(prefix), len(str))
		for str[:n] != prefix[:n] {
			n--
		}
		prefix = prefix[:n]
	}
	//return prefix
	return strs[0][:len(prefix)] // use original string
}

func readLastUntilStart(rd iorw.ReaderAt, index int) (int, string, bool) {
	sc := iorw.NewScanner(rd)
	sc.Reverse = true
	max := 1000
	if v, p2, err := sc.M.StringValue(index, sc.W.RuneFnLoop(func(ru rune) bool {
		max--
		if max <= 0 {
			return false
		}
		return ru == '_' ||
			unicode.IsLetter(ru) ||
			unicode.IsNumber(ru) ||
			unicode.IsDigit(ru)
	})); err != nil {
		return 0, "", false
	} else {
		s := v.(string)
		if s == "" {
			return 0, "", false
		}
		return p2, s, true
	}
}
