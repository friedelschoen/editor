package pscan

// WARNING: DO NOT EDIT, THIS FILE WAS AUTO GENERATED

type Wrap struct {
	sc *Scanner
	M  *Match
}

func (w *Wrap) init(sc *Scanner) {
	w.sc = sc
	w.M = sc.M
}

func (w *Wrap) And(fns ...MFn) MFn {
	return func(pos int) (int, error) {
		return w.M.And(pos, fns...)
	}
}

func (w *Wrap) AndR(fns ...MFn) MFn {
	return func(pos int) (int, error) {
		return w.M.AndR(pos, fns...)
	}
}

func (w *Wrap) Or(fns ...MFn) MFn {
	return func(pos int) (int, error) {
		return w.M.Or(pos, fns...)
	}
}

func (w *Wrap) Optional(fn MFn) MFn {
	return func(pos int) (int, error) {
		return w.M.Optional(pos, fn)
	}
}

func (w *Wrap) Rune(ru rune) MFn {
	return func(pos int) (int, error) {
		return w.M.Rune(pos, ru)
	}
}

func (w *Wrap) RuneFn(fn func(rune) bool) MFn {
	return func(pos int) (int, error) {
		return w.M.RuneFn(pos, fn)
	}
}

func (w *Wrap) RuneFnLoop(fn func(rune) bool) MFn {
	return func(pos int) (int, error) {
		return w.M.RuneFnLoop(pos, fn)
	}
}

func (w *Wrap) RuneOneOf(rs []rune) MFn {
	return func(pos int) (int, error) {
		return w.M.RuneOneOf(pos, rs)
	}
}

func (w *Wrap) RuneNoneOf(rs []rune) MFn {
	return func(pos int) (int, error) {
		return w.M.RuneNoneOf(pos, rs)
	}
}

func (w *Wrap) Sequence(seq string) MFn {
	return func(pos int) (int, error) {
		return w.M.Sequence(pos, seq)
	}
}

func (w *Wrap) SequenceMid(seq string) MFn {
	return func(pos int) (int, error) {
		return w.M.SequenceMid(pos, seq)
	}
}

func (w *Wrap) LimitedLoop(min, max int, fn MFn) MFn {
	return func(pos int) (int, error) {
		return w.M.LimitedLoop(pos, min, max, fn)
	}
}

func (w *Wrap) Loop(fn MFn) MFn {
	return func(pos int) (int, error) {
		return w.M.Loop(pos, fn)
	}
}

func (w *Wrap) LoopSep(fn, sep MFn) MFn {
	return func(pos int) (int, error) {
		return w.M.LoopSep(pos, fn, sep)
	}
}

func (w *Wrap) Spaces(includeNL bool, escape rune) MFn {
	return func(pos int) (int, error) {
		return w.M.Spaces(pos, includeNL, escape)
	}
}

func (w *Wrap) EscapeAny(escape rune) MFn {
	return func(pos int) (int, error) {
		return w.M.EscapeAny(pos, escape)
	}
}

func (w *Wrap) StringSection(openclose string, esc rune, failOnNewline bool, maxLen int, eofClose bool) MFn {
	return func(pos int) (int, error) {
		return w.M.StringSection(pos, openclose, esc, failOnNewline, maxLen, eofClose)
	}
}

func (w *Wrap) QuotedString() MFn {
	return func(pos int) (int, error) {
		return w.M.QuotedString(pos)
	}
}

func (w *Wrap) QuotedString2(esc rune, maxLen1, maxLen2 int) MFn {
	return func(pos int) (int, error) {
		return w.M.QuotedString2(pos, esc, maxLen1, maxLen2)
	}
}

func (w *Wrap) RegexpFromStartCached(res string, maxLen int) MFn {
	return func(pos int) (int, error) {
		return w.M.RegexpFromStartCached(pos, res, maxLen)
	}
}

func (w *Wrap) MustErr(fn MFn) MFn {
	return func(pos int) (int, error) {
		return w.M.MustErr(pos, fn)
	}
}

func (w *Wrap) PtrFalse(v *bool) MFn {
	return func(pos int) (int, error) {
		return w.M.PtrFalse(pos, v)
	}
}

func (w *Wrap) StaticTrue(v bool) MFn {
	return func(pos int) (int, error) {
		return w.M.StaticTrue(pos, v)
	}
}

func (w *Wrap) ReverseMode(reverse bool, fn MFn) MFn {
	return func(pos int) (int, error) {
		return w.M.ReverseMode(pos, reverse, fn)
	}
}

func (w *Wrap) OnValue(fn VFn, cb func(any)) MFn {
	return func(pos int) (int, error) {
		return w.M.OnValue(pos, fn, cb)
	}
}

func (w *Wrap) StringValue(fn MFn) VFn {
	return func(pos int) (any, int, error) {
		return w.M.StringValue(pos, fn)
	}
}

func (w *Wrap) RuneValue(fn MFn) VFn {
	return func(pos int) (any, int, error) {
		return w.M.RuneValue(pos, fn)
	}
}
