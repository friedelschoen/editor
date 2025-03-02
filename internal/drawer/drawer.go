package drawer

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"

	"github.com/friedelschoen/glake/internal/io/iorw"
	"github.com/friedelschoen/glake/internal/mathutil"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	eofRune    = -1
	noDrawRune = -2
)

type RangeAlignment int

const (
	RAlignKeep         RangeAlignment = iota
	RAlignKeepOrBottom                // keep if visible, or bottom
	RAlignAuto
	RAlignTop
	RAlignBottom
	RAlignCenter
)

type Drawer struct {
	reader iorw.ReaderAt

	fface            font.Face
	lineHeight       fixed.Int52_12
	bounds           image.Rectangle
	firstLineOffsetX int
	fg               color.Color
	smoothScroll     bool

	iters struct {
		runeR              RuneReader // init
		measure            Measure    // end
		drawR              DrawRune
		line               Line
		lineWrap           LineWrap  // init, insert
		lineStart          LineStart // init
		indent             Indent    // insert
		earlyExit          EarlyExit
		curColors          CurColors
		bgFill             BgFill
		cursor             Cursor
		pointOf            PointOf     // end
		indexOf            IndexOf     // end
		colorize           Colorize    // init
		annotations        Annotations // insert
		annotationsIndexOf AnnotationsIndexOf
	}

	st State

	loopv struct {
		iters []Iterator
		i     int
		stop  bool
	}

	// internal opt data
	opt struct {
		measure struct {
			updated bool
			measure image.Point
		}
		runeO struct {
			offset int
		}
		cursor struct {
			offset int
		}
		wordH struct {
			word        []byte
			updatedWord bool
			updatedOps  bool
		}
		parenthesisH struct {
			updated bool
		}
		syntaxH struct {
			updated bool
		}
	}

	// external options
	Opt struct {
		QuickMeasure     bool // just return the bounds size
		EarlyExitMeasure bool // allow early exit
		RuneReader       struct {
			StartOffsetX int
		}
		LineWrap struct {
			On     bool
			Fg, Bg color.Color
		}
		Cursor struct {
			On         bool
			Fg         color.Color
			AddedWidth int
		}
		Colorize struct {
			Groups []*ColorizeGroup
		}
		Annotations struct {
			On       bool
			Fg, Bg   color.Color
			Selected struct {
				EntryIndex int
				Fg, Bg     color.Color
			}
			Entries *AnnotationGroup // must be ordered by offset
		}
		WordHighlight struct {
			On     bool
			Fg, Bg color.Color
			Group  ColorizeGroup
		}
		ParenthesisHighlight struct {
			On     bool
			Fg, Bg color.Color
			Group  ColorizeGroup
		}
		SyntaxHighlight struct {
			On    bool
			Group ColorizeGroup
		}
	}
}

// State should not be stored/restored except in initializations.
// ex: runeR.extra and runeR.ru won't be correctly set if the iterators were stopped.
type State struct {
	runeR struct {
		ri            int
		ru, prevRu    rune
		pen           fixed.Point52_12 // upper left corner (not at baseline)
		kern, advance fixed.Int52_12
		extra         int
		startRi       int
		fface         font.Face
	}
	measure struct {
		penMax fixed.Point52_12
	}
	drawR struct {
		img   draw.Image
		delay *DrawRuneDelay
	}
	line struct {
		lineStart bool
	}
	lineWrap struct {
		//wrapRi       int
		wrapping     bool
		preLineWrap  bool
		postLineWrap bool
	}
	lineStart struct {
		offset     int
		nLinesUp   int
		q          []int
		ri         int
		uppedLines int
		reader     iorw.ReaderAt // limited reader
	}
	indent struct {
		notStartingSpaces bool
		indent            fixed.Int52_12
	}
	earlyExit struct {
		extraLine bool
	}
	curColors struct {
		fg, bg color.Color
		lineBg color.Color
	}
	cursor struct {
		delay *CursorDelay
	}
	pointOf struct {
		index int
		p     image.Point
	}
	indexOf struct {
		p     fixed.Point52_12
		index int
	}
	colorize struct {
		indexes []int
	}
	annotations struct {
		cei    int // current entries index (to add to q)
		indexQ []int
	}
	annotationsIndexOf struct {
		p      fixed.Point52_12
		eindex int
		offset int
		inside struct { // inside an annotation
			on      bool
			ei      int // entry index
			soffset int // start offset
		}
	}
}

func New() *Drawer {
	d := &Drawer{}
	d.Opt.LineWrap.On = true
	d.smoothScroll = true

	// iterators
	d.iters.runeR.d = d
	d.iters.measure.d = d
	d.iters.drawR.d = d
	d.iters.line.d = d
	d.iters.lineWrap.d = d
	d.iters.lineStart.d = d
	d.iters.indent.d = d
	d.iters.earlyExit.d = d
	d.iters.curColors.d = d
	d.iters.bgFill.d = d
	d.iters.cursor.d = d
	d.iters.pointOf.d = d
	d.iters.indexOf.d = d
	d.iters.colorize.d = d
	d.iters.annotations.d = d
	d.iters.annotationsIndexOf.d = d
	return d
}

func (d *Drawer) SetReader(r iorw.ReaderAt) {
	d.reader = r
	// always run since an underlying reader could have been changed
	d.ContentChanged()
}

func (d *Drawer) Reader() iorw.ReaderAt { return d.reader }

var limitedReaderPadding = 3000

func (d *Drawer) limitedReaderPad(offset int) iorw.ReaderAt {
	pad := limitedReaderPadding
	return iorw.NewLimitedReaderAtPad(d.reader, offset, offset, pad)
}

func (d *Drawer) limitedReaderPadSpace(offset int) iorw.ReaderAt {
	// adjust the padding to avoid immediate flicker for x chars for the case of long lines
	max := 1000
	pad := limitedReaderPadding // in tests it could be a small num
	if limitedReaderPadding >= max {
		u := offset - limitedReaderPadding
		diff := max - (u % max)
		pad = limitedReaderPadding - diff
	}
	return iorw.NewLimitedReaderAtPad(d.reader, offset, offset, pad)
}

func (d *Drawer) ContentChanged() {
	d.opt.measure.updated = false
	d.opt.syntaxH.updated = false
	d.opt.wordH.updatedWord = false
	d.opt.wordH.updatedOps = false
	d.opt.parenthesisH.updated = false
}

func (d *Drawer) FontFace() font.Face { return d.fface }
func (d *Drawer) SetFontFace(ff font.Face) {
	if ff == d.fface {
		return
	}
	d.fface = ff
	d.lineHeight = fixed.Int52_12(d.fface.Metrics().Height << 6)

	d.opt.measure.updated = false
}

func (d *Drawer) LineHeight() int {
	if d.fface == nil {
		return 0
	}
	return d.fface.Metrics().Height.Ceil()
}

func (d *Drawer) SetFg(fg color.Color) { d.fg = fg }

func (d *Drawer) FirstLineOffsetX() int { return d.firstLineOffsetX }
func (d *Drawer) SetFirstLineOffsetX(x int) {
	if x != d.firstLineOffsetX {
		d.firstLineOffsetX = x
		d.opt.measure.updated = false
	}
}

func (d *Drawer) Bounds() image.Rectangle { return d.bounds }
func (d *Drawer) SetBounds(r image.Rectangle) {
	//d.ContentChanged() // commented for performance
	// performance (doesn't redo d.opt.wordH.updatedWord)
	if r.Size() != d.bounds.Size() {
		d.opt.measure.updated = false
		d.opt.syntaxH.updated = false
		d.opt.wordH.updatedOps = false
		d.opt.parenthesisH.updated = false
	}

	d.bounds = r // always update value (can change min)
}

func (d *Drawer) RuneOffset() int {
	return d.opt.runeO.offset
}
func (d *Drawer) SetRuneOffset(v int) {
	d.opt.runeO.offset = v

	d.opt.syntaxH.updated = false
	d.opt.wordH.updatedOps = false
	d.opt.parenthesisH.updated = false
}

func (d *Drawer) SetCursorOffset(v int) {
	d.opt.cursor.offset = v

	d.opt.wordH.updatedWord = false
	d.opt.wordH.updatedOps = false
	d.opt.parenthesisH.updated = false
}

func (d *Drawer) ready() bool {
	return !(d.fface == nil || d.reader == nil || d.bounds == image.Rectangle{})
}

func (d *Drawer) Measure() image.Point {
	if !d.ready() {
		return image.Point{}
	}
	if d.opt.measure.updated {
		return d.opt.measure.measure
	}
	d.opt.measure.updated = true
	d.opt.measure.measure = d.measure2()
	return d.opt.measure.measure
}

func (d *Drawer) measure2() image.Point {
	if d.Opt.QuickMeasure {
		return d.bounds.Size()
	}
	return d.measureContent()
}

// Full content measure in pixels. To be used only for small content.
func (d *Drawer) measureContent() image.Point {
	d.st = State{}
	iters := d.sIters(d.Opt.EarlyExitMeasure, &d.iters.measure)
	d.loopInit(iters)
	d.loop()
	// remove bounds min and return only the measure
	pf := d.st.measure.penMax
	p := image.Point{pf.X.Ceil(), pf.Y.Ceil()}
	m := p.Sub(d.bounds.Min)
	return m
}

func (d *Drawer) Draw(img draw.Image) {
	updateSyntaxHighlightOps(d)
	updateWordHighlightWord(d)
	updateWordHighlightOps(d)
	updateParenthesisHighlight(d)

	d.st = State{}
	iters := []Iterator{
		&d.iters.runeR,
		&d.iters.curColors,
		&d.iters.colorize,
		&d.iters.line,
		&d.iters.lineWrap,
		&d.iters.earlyExit, // after iters that change pen.Y
		&d.iters.indent,
		&d.iters.annotations, // after iters that change the line
		&d.iters.bgFill,
		&d.iters.drawR,
		&d.iters.cursor,
	}
	d.loopInit(iters)
	d.header0()
	d.st.drawR.img = img
	d.loop()
}

func (d *Drawer) LocalPointOf(index int) image.Point {
	if !d.ready() {
		return image.Point{}
	}
	d.st = State{}
	d.st.pointOf.index = index
	iters := d.sIters(true, &d.iters.pointOf)
	d.loopInit(iters)
	d.header1()
	d.loop()
	return d.st.pointOf.p
}

func (d *Drawer) LocalIndexOf(p image.Point) int {
	if !d.ready() {
		return 0
	}
	d.st = State{}
	d.st.indexOf.p = fixed.Point52_12{
		X: fixed.Int52_12(p.X << 12),
		Y: fixed.Int52_12(p.Y << 12),
	}
	iters := d.sIters(true, &d.iters.indexOf)
	d.loopInit(iters)
	d.header1()
	d.loop()
	return d.st.indexOf.index
}

func (d *Drawer) AnnotationsIndexOf(p image.Point) (int, int, bool) {
	if !d.ready() {
		return 0, 0, false
	}
	d.st = State{}
	d.st.annotationsIndexOf.p = fixed.Point52_12{
		X: fixed.Int52_12(p.X << 12),
		Y: fixed.Int52_12(p.Y << 12),
	}

	iters := d.sIters(true, &d.iters.annotations, &d.iters.annotationsIndexOf)
	d.loopInit(iters)
	d.header0()
	d.loop()

	st := &d.st.annotationsIndexOf
	if st.eindex < 0 {
		return 0, 0, false
	}
	return st.eindex, st.offset, true
}

func (d *Drawer) loopInit(iters []Iterator) {
	l := &d.loopv
	l.iters = iters
	for _, iter := range l.iters {
		iter.Init()
	}
}

func (d *Drawer) loop() {
	l := &d.loopv
	l.stop = false
	for !l.stop { // loop for each rune
		l.i = 0
		_ = d.iterNext()
	}
	for _, iter := range l.iters {
		iter.End()
	}
}

// To be called from iterators, inside the Iter() func.
func (d *Drawer) iterNext() bool {
	l := &d.loopv
	if !l.stop && l.i < len(l.iters) {
		u := l.iters[l.i]
		l.i++
		u.Iter()
		l.i--
	}
	return !l.stop
}

func (d *Drawer) iterStop() {
	d.loopv.stop = true
}

func (d *Drawer) iterNextExtra() bool {
	d.iters.runeR.pushExtra()
	defer d.iters.runeR.popExtra()
	return d.iterNext()
}

func (d *Drawer) visibleLen() (int, int, int, int) {
	d.st = State{}
	iters := d.sIters(true)
	d.loopInit(iters)
	d.header0()
	startRi := d.st.runeR.ri
	d.loop()

	// from the line start
	drawOffset := startRi
	drawLen := d.st.runeR.ri - drawOffset
	// from the current offset
	offset := d.opt.runeO.offset
	offsetLen := d.st.runeR.ri - offset

	return drawOffset, drawLen, offset, offsetLen
}

func (d *Drawer) ScrollOffset() image.Point {
	return image.Point{0, d.RuneOffset()}
}
func (d *Drawer) SetScrollOffset(o image.Point) {
	d.SetRuneOffset(o.Y)
}

func (d *Drawer) ScrollSize() image.Point {
	return image.Point{0, d.reader.Max() - d.reader.Min()}
}

func (d *Drawer) ScrollViewSize() image.Point {
	nlines := d.boundsNLines()
	n := d.scrollSizeY(nlines, false) // n runes
	return image.Point{0, n}
}

func (d *Drawer) ScrollPageSizeY(up bool) int {
	nlines := d.boundsNLines()
	return d.scrollSizeY(nlines, up)
}

func (d *Drawer) ScrollWheelSizeY(up bool) int {
	nlines := d.boundsNLines()

	// limit nlines
	nlines /= 4
	if nlines < 1 {
		nlines = 1
	} else if nlines > 4 {
		nlines = 4
	}

	return d.scrollSizeY(nlines, up)
}

// integer lines
func (d *Drawer) boundsNLines() int {
	dy := fixed.Int52_12(d.bounds.Dy() << 12)
	return int(dy / d.lineHeight)
}

func (d *Drawer) scrollSizeY(nlines int, up bool) int {
	if up {
		o := d.scrollSizeYUp(nlines)
		return -(d.opt.runeO.offset - o)
	} else {
		o := d.scrollSizeYDown(nlines)
		return o - d.opt.runeO.offset
	}
}

func (d *Drawer) scrollSizeYUp(nlines int) int {
	return d.wlineStartIndex(true, d.opt.runeO.offset, nlines, nil)
}
func (d *Drawer) scrollSizeYDown(nlines int) int {
	return d.wlineStartIndexDown(d.opt.runeO.offset, nlines)
}

func (d *Drawer) RangeVisible(offset, length int) bool {
	v1, _ := penVisibility(d, offset)
	if v1 != VisibilityNot {
		return true
	}
	v2, _ := penVisibility(d, offset+length)

	return v2 != VisibilityNot
	// v1 above and v2 below view is considered not visible (will align with v1 at top on RangeVisibleOffset(...))
}

func (d *Drawer) RangeVisibleOffset(offset, length int, align RangeAlignment) int {

	// top lines visible before the offset line
	topLines := func(n int) int {
		return d.wlineStartIndex(true, offset, n, nil)
	}

	freeLines := func() int {
		rnlines := d.rangeNLines(offset, length)
		bnlines := d.boundsNLines()
		// extra lines beyond the lines ocuppied by the range
		v := bnlines - rnlines
		if v < 0 {
			v = 0
		}
		return v
	}

	switch align {
	case RAlignKeep:
		return d.alignKeep()
	case RAlignKeepOrBottom:
		if v, ok := d.rangeVisibleOffsetKeepIfVisible(offset, length); ok {
			return v
		}
		return topLines(freeLines())
	case RAlignTop:
		return topLines(0)
	case RAlignCenter:
		return topLines(freeLines() / 2)
	case RAlignBottom:
		return topLines(freeLines())
	case RAlignAuto:
		return d.rangeVisibleOffsetAuto(offset, length)
	default:
		panic(fmt.Errorf("todo: %v", align))
	}
}

func (d *Drawer) alignKeep() int {
	return mathutil.Min(d.opt.runeO.offset, d.reader.Max())
}

func (d *Drawer) rangeVisibleOffsetKeepIfVisible(offset, length int) (int, bool) {
	offset2 := d.rangeVisibleOffset2(offset, length)
	v2, _ := penVisibility(d, offset2)
	if v2 == VisibilityFull {
		return d.alignKeep(), true
	}
	return 0, false
}

func (d *Drawer) rangeVisibleOffsetAuto(offset, length int) int {
	align := func(a RangeAlignment) int {
		return d.RangeVisibleOffset(offset, length, a)
	}

	offset2 := d.rangeVisibleOffset2(offset, length)

	v1, top1 := penVisibility(d, offset)
	v2, _ := penVisibility(d, offset2)
	if v1 == VisibilityFull {
		if v2 == VisibilityFull {
			return align(RAlignKeep)
		} else if v2 == VisibilityPartial {
			return align(RAlignBottom)
		} else if v2 == VisibilityNot { // past bottom line
			return align(RAlignBottom)
		}
	} else if v1 == VisibilityPartial {
		if top1 {
			return align(RAlignTop)
		} else {
			return align(RAlignBottom)
		}
	} else if v1 == VisibilityNot {
		if v2 == VisibilityFull {
			return align(RAlignTop)
		} else if v2 == VisibilityPartial {
			return align(RAlignTop)
		} else if v2 == VisibilityNot {
			return align(RAlignCenter)
		}
	}

	// NOTE: should never get here
	log.Printf("drawer: range visible offset bad value: %v, %v", offset, length)

	return align(RAlignCenter)
}

func (d *Drawer) rangeVisibleOffset2(offset, length int) int {
	// don't let offset+length be beyond max for v2 (would give not visible)
	offset2 := offset + length
	if offset2 > d.reader.Max() {
		offset2 = offset
	}
	return offset2
}

func (d *Drawer) rangeNLines(offset, length int) int {
	pr1, pr2, ok := d.wlineRangePenBounds(offset, length)
	if ok {
		w := pr2.Max.Y - pr1.Min.Y
		u := int(w / d.lineHeight)
		if u >= 1 {
			return u
		}
	}
	return 1 // always at least one line
}

func (d *Drawer) wlineRangePenBounds(offset, length int) (fixed.Rectangle52_12, fixed.Rectangle52_12, bool) {
	var pr1, pr2 fixed.Rectangle52_12
	var ok1, ok2 bool
	d.wlineStartLoopFn(true, offset, 0,
		func() {
			ok1 = true
			pr1 = d.iters.runeR.penBounds()
		},
		func() {
			if d.st.runeR.ri == offset+length {
				ok2 = true
				pr2 = d.iters.runeR.penBounds()
				d.iterStop()
				return
			}
			if !d.iterNext() {
				return
			}
		})
	return pr1, pr2, ok1 && ok2
}

func (d *Drawer) wlineStartIndexDown(offset, nlinesDown int) int {
	count := 0
	startRi := 0
	d.wlineStartLoopFn(true, offset, 0,
		func() {
			startRi = d.st.runeR.ri
			if nlinesDown == 0 {
				d.iterStop()
			}
		},
		func() {
			if d.st.line.lineStart || d.st.lineWrap.postLineWrap {
				if d.st.runeR.ri != startRi { // bypass ri at line start
					count++
					if count >= nlinesDown {
						d.iterStop()
						return
					}
				}
			}
			if !d.iterNext() {
				return
			}
		})
	return d.st.runeR.ri
}

func (d *Drawer) header0() {
	_ = d.header(d.opt.runeO.offset, 0)
}

func (d *Drawer) header1() {
	d.st.earlyExit.extraLine = true       // extra line at bottom
	ul := d.header(d.opt.runeO.offset, 1) // extra line at top
	if ul > 0 {
		d.st.runeR.pen.Y -= d.lineHeight * fixed.Int52_12(ul<<12)
	}
}

func (d *Drawer) header(offset, nLinesUp int) int {
	// smooth scrolling
	adjustPenY := fixed.Int52_12(0)
	if d.smoothScroll {
		adjustPenY += d.smoothScrolling(offset)
	}

	// iterate to the wline start
	st1RRPen := d.st.runeR.pen // keep initialized state to refer to pen difference
	uppedLines := d.wlineStartState(false, offset, nLinesUp)
	adjustPenY += d.st.runeR.pen.Y - st1RRPen.Y
	d.st.runeR.pen.Y -= adjustPenY

	return uppedLines
}

func (d *Drawer) smoothScrolling(offset int) fixed.Int52_12 {
	// keep/restore state to avoid interfering with other running iterations
	st := d.st
	defer func() { d.st = st }()

	s, e := d.wlineStartEnd(offset)
	t := e - s
	if t <= 0 {
		return 0
	}
	k := offset - s
	perc := float64(k) / float64(t)
	return fixed.Int52_12(int64(float64(d.lineHeight) * perc * (1 << 12)))
}

func (d *Drawer) wlineStartEnd(offset int) (int, int) {
	s, e := 0, 0
	d.wlineStartLoopFn(true, offset, 0,
		func() {
			s = d.st.runeR.ri
		},
		func() {
			if d.st.line.lineStart || d.st.lineWrap.postLineWrap {
				if d.st.runeR.ri > offset {
					e = d.st.runeR.ri
					d.iterStop()
					return
				}
			}
			if !d.iterNext() {
				return
			}
		})
	if e == 0 {
		e = d.st.runeR.ri
	}
	return s, e
}

func (d *Drawer) wlineStartLoopFn(clearState bool, offset, nLinesUp int, fnInit func(), fn func()) {
	// keep/restore iters
	iters := d.loopv.iters
	defer func() { d.loopv.iters = iters }()

	d.loopv.iters = d.sIters(false, &FnIter{fn: fn})
	d.wlineStartState(clearState, offset, nLinesUp)
	fnInit()
	d.loop()
}

// Leaves the state at line start
func (d *Drawer) wlineStartState(clearState bool, offset, nLinesUp int) int {
	// keep/restore iters
	iters := d.loopv.iters
	defer func() { d.loopv.iters = iters }()

	// set limited reading here to have common limits on the next two calls
	//var rd iorw.Reader
	//rd := d.limitedReaderPad(offset)
	rd := d.limitedReaderPadSpace(offset)

	// find start (state will reach offset)
	cp := d.st // keep state
	k := d.wlineStartIndex(clearState, offset, nLinesUp, rd)
	uppedLines := d.st.lineStart.uppedLines

	// leave state at line start instead of offset
	d.st = cp // restore state
	_ = d.wlineStartIndex(clearState, k, 0, rd)

	return uppedLines
}

func (d *Drawer) wlineStartIndex(clearState bool, offset, nLinesUp int, rd iorw.ReaderAt) int {
	if clearState {
		d.st = State{}
	}
	d.st.lineStart.offset = offset
	d.st.lineStart.nLinesUp = nLinesUp
	d.st.lineStart.reader = rd
	iters := d.sIters(false, &d.iters.lineStart)
	d.loopInit(iters)
	d.loop()
	return d.st.lineStart.ri
}

// structure iterators
func (d *Drawer) sIters(earlyExit bool, more ...Iterator) []Iterator {
	iters := []Iterator{
		&d.iters.runeR,
		&d.iters.line,
		&d.iters.lineWrap,
	}
	if earlyExit {
		iters = append(iters, &d.iters.earlyExit)
	}
	iters = append(iters, &d.iters.indent)
	iters = append(iters, more...)
	return iters
}

type Iterator interface {
	Init()
	Iter()
	End()
}

type FnIter struct {
	fn func()
}

func (it *FnIter) Init() {}
func (it *FnIter) Iter() { it.fn() }
func (it *FnIter) End()  {}

func assignColor(dest *color.Color, src color.Color) {
	if src != nil {
		*dest = src
	}
}
