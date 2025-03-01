package lrparser

type StatesData struct {
	states            []*State
	shiftOnSRConflict bool
}

//func (sd *StatesData) checkStringsConflicts2(sr1, sr2 *StringRule) error {
//	switch sr2.typ {
//	case stringRTOr:
//		for _, ru2 := range sr2.runes {
//			if has, err := sd.srHasRune(sr1, ru2); err != nil {
//				return err
//			} else if has {
//				return fmt.Errorf("rune %q already in %v", ru2, sr1)
//			}
//		}
//		for _, rr2 := range sr2.ranges {
//			for _, ru2 := range []rune{rr2[0], rr2[1]} {
//				if has, err := sd.srHasRune(sr1, ru2); err != nil {
//					return err
//				} else if has {
//					return fmt.Errorf("range %v already in %v", rr2, sr1)
//				}
//			}
//		}
//		//case stringRTOr:
//	}
//	return nil
//}

//func (sd *StatesData) srHasRune(sr *StringRule, ru rune) (bool, error) {
//	switch sr.typ {
//	case stringRTOr:
//		for _, ru2 := range sr.runes {
//			if ru == ru2 {
//				return true, nil
//			}
//		}
//		for _, rr := range sr.ranges {
//			if rr.HasRune(ru) {
//				return true, nil
//			}
//		}
//		return false, nil
//	case stringRTOrNeg:
//		for _, ru2 := range sr.runes {
//			if ru == ru2 {
//				return false, nil
//			}
//		}
//		for _, rr := range sr.ranges {
//			if !rr.HasRune(ru) {
//				return false, nil
//			}
//		}
//		return true, nil
//	}
//	return false, fmt.Errorf("not orrule")
//}

//func (sd *StatesData) srHasRune(sr *StringRule, ru rune) (bool, error) {
//	switch sr.typ {
//	case stringRTOr:
//		for _, ru2 := range sr.runes {
//			if ru == ru2 {
//				return true, nil
//			}
//		}
//		for _, rr := range sr.ranges {
//			if rr.HasRune(ru) {
//				return true, nil
//			}
//		}
//		return false, nil
//	case stringRTOrNeg:
//		for _, ru2 := range sr.runes {
//			if ru == ru2 {
//				return false, nil
//			}
//		}
//		for _, rr := range sr.ranges {
//			if !rr.HasRune(ru) {
//				return false, nil
//			}
//		}
//		return true, nil
//	}
//	return false, fmt.Errorf("not orrule")
//}

//func (sd *StatesData) runeConflict(sr *StringRule, ru rune) error {
//	switch sr.typ {
//	case stringRTOr:
//		for _, ru2 := range sr.runes {
//			if ru2 == ru {
//				return fmt.Errorf("rune %q already defined at %v", ru sr)
//			}
//		}
//		for _, rr := range sr.ranges {
//			if rr.HasRune(ru) {
//				return fmt.Errorf("rune %q already defined at %v", ru sr)
//			}
//		}
//		return false, nil
//	case stringRTOrNeg:
//		for _, ru2 := range sr.runes {
//			if ru == ru2 {
//				return false, nil
//			}
//		}
//		for _, rr := range sr.ranges {
//			if !rr.HasRune(ru) {
//				return false, nil
//			}
//		}
//		return true, nil
//	default:
//		panic(fmt.Sprintf("bad stringrule type: %q", sr.typ))
//	}
//}

//func (sd *StatesData) solveConflicts(vd *VerticesData) error {
//	// strings conflicts (runes)
//	for _, st := range sd.states {
//		orM := map[rune]Rule{}
//		orNegM := map[rune]Rule{}
//		orRangeM := map[RuneRange]Rule{}
//		orRangeNegM := map[RuneRange]Rule{}

//		hasAnyrune := false
//		for _, r := range st.rsetSorted {
//			if r == anyruneRule {
//				hasAnyrune = true
//				break
//			}
//		}

//		// check duplicates in orRules
//		for _, r := range st.rsetSorted {
//			sr, ok := r.(*StringRule)
//			if !ok {
//				continue
//			}

//			typ := sr.typ

//			// special case: check andRule as orRule
//			if typ == stringRTAnd && len(sr.runes) == 1 {
//				typ = stringRTOr
//			}

//			switch typ {
//			//case stringRTAnd: // sequence
//			//case stringRTMid: // sequence
//			case stringRTOr:
//				if err := sd.checkRuneDups(orM, st, sr, sr.runes...); err != nil {
//					return err
//				}
//				if err := sd.checkRangeDups(orRangeM, st, sr, sr.ranges...); err != nil {
//					return err
//				}
//			case stringRTOrNeg:
//				if err := sd.checkRuneDups(orNegM, st, sr, sr.runes...); err != nil {
//					return err
//				}
//				if err := sd.checkRangeDups(orRangeNegM, st, sr, sr.ranges...); err != nil {
//					return err
//				}
//			}
//		}

//		// check intersections: between individual runes and ranges
//		if err := sd.checkRunesRangesDups(orM, orRangeM, st); err != nil {
//			return err
//		}
//		if err := sd.checkRunesRangesDups(orNegM, orRangeNegM, st); err != nil {
//			return err
//		}

//		// check intersections: all "or" rules must be in "negation" if it is defined (ex: (a|b|(c|a|b)!)
//		if err := sd.checkRunesNegation(orM, orNegM, orRangeNegM, st); err != nil {
//			return err
//		}
//		//if err := sd.checkRangesNegation(orM, orNegM, st); err != nil {
//		//	return err
//		//}

//		// check conflicts: all "or" runes must be in "not"
//		if len(orNegM) > 0 {
//			for ru, r := range orM {
//				_, ok := orNegM[ru]
//				if !ok {
//					// show "not" rules
//					rs := &RuleSet{}
//					for _, r2 := range orNegM {
//						rs.set(r2)
//					}

//					return fmt.Errorf("%v: rune %q in %v is covered in %v", st.id, ru, r, rs)
//				}
//			}
//		}
//		if hasAnyrune {
//			if len(orM) > 0 || len(orNegM) > 0 {
//				return fmt.Errorf("%v: anyrune and stringrule in the same state\n%v", st.id, sd)
//			}
//		}
//	}

//}
//func (sd *StatesData) checkRuneDups(m map[rune]Rule, st *State, r Rule, rs ...rune) error {
//	for _, ru := range rs {
//		r2, ok := m[ru]
//		if ok {
//			return fmt.Errorf("%v: rune %q in %v is already defined at %v", st.id, ru, r, r2)
//		}
//		m[ru] = r
//	}
//	return nil
//}
//func (sd *StatesData) checkRangeDups(m map[RuneRange]Rule, st *State, r Rule, h ...RuneRange) error {
//	for _, rr := range h {
//		for rr2, r2 := range m {
//			if rr2.IntersectsRange(rr) {
//				return fmt.Errorf("%v: range %q in %v is already defined at %v", st.id, rr, r, r2)
//			}
//		}
//		m[rr] = r
//	}
//	return nil
//}
//func (sd *StatesData) checkRunesRangesDups(m1 map[Rune]Rule, m2 map[RuneRange]Rule, st *State) error {
//	for ru, r1 := range m1 {
//		for rr, r2 := range m2 {
//			if rr.HasRune(ru) {
//				return fmt.Errorf("%v: rune %q in %v is covered by range %v", st.id, ru, r1, rr)
//			}
//		}
//		m[rr] = r
//	}
//	return nil
//}
//func (sd *StatesData) checkRunesNegation(m, neg map[Rune]Rule, negRange map[RuneRange]Rule, st *State) error {
//	// all "or" runes must be in "neg"
//	if len(neg) > 0 {
//		for ru, r := range m {
//			_, ok := neg[ru]
//			if ok {
//				continue
//			}
//			// show "not" rules
//			rs := &RuleSet{}
//			for _, r2 := range neg {
//				rs.set(r2)
//			}
//			return fmt.Errorf("%v: rune %q in %v is covered in %v", st.id, ru, r, rs)
//		}
//	}
//	return nil
//}
//func (sd *StatesData) checkRunesNegation2(m map[Rune]Rule, neg map[RuneRange]Rule, st *State) error {
//	if len(neg) == 0 {
//		return nil
//	}
//	// all "or" runes must be in "neg"
//	for ru, r := range m {
//		for rr,r2:=range neg{
//			if rr.HasRune(ru)[

//			}
//		}
//		_, ok := neg[ru]
//		if ok {
//			continue
//		}
//		// show "not" rules
//		rs := &RuleSet{}
//		for _, r2 := range neg {
//			rs.set(r2)
//		}
//		return fmt.Errorf("%v: rune %q in %v is covered in %v", st.id, ru, r, rs)
//	}
//	return nil
//}

//godebug:annotateoff

type State struct {
	id             stateId
	action         map[Rule][]Action
	gotoSt         map[Rule]*State
	rsetSorted     []Rule // rule set to parse in this state
	rsetHasEndRule bool
}

//godebug:annotateoff

type stateId int

type Action any
