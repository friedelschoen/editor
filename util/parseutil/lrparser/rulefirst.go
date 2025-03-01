package lrparser

// rules first terminals
type RuleFirstT struct {
	ri      *RuleIndex
	cache   map[Rule]RuleSet
	seen    map[Rule]int
	reverse bool
}

//type RuleFollow struct {
//	ri     *RuleIndex
//	rFirst *RulesFirst
//	cache  map[Rule]RuleSet
//}

//func newRuleFollow(ri *RuleIndex, rFirst *RulesFirst, r Rule) *RuleFollow {
//	rf := &RuleFollow{ri: ri, rFirst: rFirst}
//	rf.cache = map[Rule]RuleSet{}
//	rf.calc(r)
//	return rf
//}
//func (rf *RuleFollow) get(r Rule) RuleSet {
//	return rf.cache[r]
//}
//func (rf *RuleFollow) calc(r Rule) {
//	AFollow := RuleSet{}
//	AFollow.set(rf.ri.endRule())
//	rf.cache[r] = AFollow

//	seen := map[Rule]int{}
//	rf.calc2(r, AFollow, seen)
//}
//func (rf *RuleFollow) calc2(A Rule, AFollow RuleSet, seen map[Rule]int) {
//	if seen[A] >= 2 { // need to visit 2nd time to allow afollow to be used in nested rules
//		return
//	}
//	seen[A]++
//	defer func() { seen[A]-- }()

//	//rset := RuleSet{}
//	w, ok := ruleProductions(A)
//	if !ok { // terminal
//		return
//	}
//	nilr := rf.ri.nilRule()
//	for _, r2 := range w {
//		// A->r2
//		w2 := ruleRhs(r2) // sequence
//		for i, B := range w2 {
//			// A->αBβ

//			if ruleIsTerminal(B) {
//				continue
//			}

//			BFollow, ok := rf.cache[B]
//			if !ok {
//				BFollow = RuleSet{}
//				rf.cache[B] = BFollow
//			}

//			haveβ := i < len(w2)-1
//			βFirstHasNil := false
//			if haveβ {
//				β := w2[i+1]
//				βFirst := RuleSet{}
//				βFirst.add(rf.rFirst.get(β))
//				βFirstHasNil = βFirst.isSet(nilr)
//				βFirst.unset(nilr)
//				BFollow.add(βFirst)
//			}
//			if !haveβ || βFirstHasNil {
//				BFollow.add(AFollow)
//			}

//			rf.calc2(B, BFollow, seen)
//		}
//	}
//}
//func (rf *RuleFollow) String() string {
//	u := []string{}
//	for _, r := range rf.ri.sorted() {
//		u = append(u, fmt.Sprintf("%v:%v", r.id(), rf.get(r)))
//	}
//	return fmt.Sprintf("{\n\t%v\n}", strings.Join(u, ",\n\t"))
//}
