package core

import (
	"context"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/jmigpin/editor/ui"
	"github.com/jmigpin/editor/util/drawutil/drawer4"
	"github.com/jmigpin/editor/util/iout/iorw"
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
	ic.mu.cancel = func() {} // avoid testing for nil
	return ic
}

//----------

func (ic *InlineComplete) Complete(erow *ERow, ev *ui.TextAreaInlineCompleteEvent) bool {

	// early pre-check if filename is supported
	_, err := ic.ed.LSProtoMan.LangManager(erow.Info.Name())
	if err != nil {
		return false // not handled
	}

	ta := ev.TextArea

	ic.mu.Lock()

	// cancel previous run
	ic.mu.cancel()
	if ic.mu.ta != nil && ic.mu.ta != ta {
		defer ic.setAnnotations(ic.mu.ta, nil)
	}

	ctx, cancel := context.WithCancel(erow.ctx)
	ic.mu.cancel = cancel
	ic.mu.ta = ta
	ic.mu.index = ta.CursorIndex()

	ic.mu.Unlock()

	go func() {
		defer cancel()
		ic.setAnnotationsMsg(ta, "loading...")
		err := ic.complete2(ctx, erow.Info.Name(), ta, ev)
		if err != nil {
			ic.setAnnotations(ta, nil)
			ic.ed.Error(err)
		}
		// TODO: not necessary in all cases
		// ensure UI update
		ic.ed.UI.EnqueueNoOpEvent()
	}()
	return true
}

func (ic *InlineComplete) complete2(ctx context.Context, filename string, ta *ui.TextArea, ev *ui.TextAreaInlineCompleteEvent) error {
	comps, err := ic.completions(ctx, filename, ta)
	if err != nil {
		return err
	}

	// insert complete
	completed, comps, err := ic.insertComplete(comps, ta)
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
		entries := []*drawer4.Annotation{}
		for _, v := range comps {
			u := &drawer4.Annotation{Offset: ev.Offset, Bytes: []byte(v)}
			entries = append(entries, u)
		}
		ic.setAnnotations(ta, entries)
	}
	return nil
}

func (ic *InlineComplete) insertComplete(comps []string, ta *ui.TextArea) (completed bool, _ []string, _ error) {
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

//----------

func (ic *InlineComplete) completions(ctx context.Context, filename string, ta *ui.TextArea) ([]string, error) {
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
	return res, nil
}

//----------

func (ic *InlineComplete) setAnnotationsMsg(ta *ui.TextArea, s string) {
	offset := ta.CursorIndex()
	entries := []*drawer4.Annotation{{Offset: offset, Bytes: []byte(s)}}
	ic.setAnnotations(ta, entries)
}

func (ic *InlineComplete) setAnnotations(ta *ui.TextArea, entries []*drawer4.Annotation) {
	on := entries != nil && len(entries) > 0
	ic.ed.SetAnnotations(EareqInlineComplete, ta, on, -1, entries)
	if !on {
		ic.setOff(ta)
	}
}

//----------

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

//----------

func (ic *InlineComplete) CancelAndClear() {
	ic.mu.Lock()
	ta := ic.mu.ta
	ic.mu.Unlock()
	if ta != nil {
		ic.setAnnotations(ta, nil)
	}
}

func (ic *InlineComplete) CancelOnCursorChange() {
	ic.mu.Lock()
	ta := ic.mu.ta
	index := ic.mu.index
	ic.mu.Unlock()
	if ta != nil {
		if index != ta.CursorIndex() {
			ic.setAnnotations(ta, nil)
		}
	}
}

//----------

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
		// original string
		origStr := prefix

		// string to insert
		n := len(origStr)
		insStr := expandStr

		// try to expand the index to the existing text
		expand := len(expandStr) - len(prefix)
		for i := 0; i < expand; i++ {
			b, err := rw.ReadFastAt(index+i, 1)
			if err != nil {
				break
			}
			if b[0] != insStr[n] {
				break
			}
			n++
		}

		// insert completion
		if insStr != origStr {
			err := rw.OverwriteAt(start, n, []byte(insStr))
			if err != nil {
				return 0, false, nil, err
			}
			newIndex = start + len(insStr)
			return newIndex, true, comps2, nil
		}
	}

	return 0, false, comps2, nil
}

//----------

func expandAndFilter(prefix string, comps []string) (expand string, comps5 []string) {
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

	//// NOTE: this loses the provided order, but better results?
	//sort.Strings(comps2)

	// longest prefix
	lcp := longestCommonPrefix(comps2)

	// choose next in line
	if len(lcp) == len(prefix) {
		for i, s := range comps2 {
			if strings.HasPrefix(s, prefix) {
				k := (i + 1) % len(comps2) // next
				lcp = comps2[k][:len(prefix)]
			}
		}
	}

	return lcp, comps2
}

func longestCommonPrefix(strs0 []string) string {
	// make copy
	strs := make([]string, len(strs0))
	copy(strs, strs0)

	sort.Strings(strs) // allows needing to compare only first/last
	first := strings.ToLower(strs[0])
	last := strings.ToLower(strs[len(strs)-1])
	longestPrefix := ""
	for i := 0; i < len(first) && i < len(last); i++ {
		if first[i] == last[i] {
			//longestPrefix += string(first[i])
			longestPrefix = strs0[0][:i+1] // use original order
			continue
		}
		break
	}
	return longestPrefix
}

//----------

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
