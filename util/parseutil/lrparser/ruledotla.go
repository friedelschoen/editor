package lrparser

// aka "set of items"
type RuleDotsLaSet map[RuleDot]RuleSet // lookahead set

//// aka "set of items"
//type RuleDotLas []*RuleDotLa

//func (rdlas RuleDotLas) has(rdla *RuleDotLa) bool {
//	str := rdla.String()
//	for _, rdla2 := range rdlas {
//		if rdla2.String() == str {
//			return true
//		}
//	}
//	return false
//}
//func (rdlas RuleDotLas) hasAll(rdlas2 RuleDotLas) bool {
//	for _, rdla2 := range rdlas2 {
//		if rdlas.has(rdla2) {
//			return true
//		}
//	}
//	return false
//}
//func (rdlas RuleDotLas) appendUnique(rdlas2 RuleDotLas) RuleDotLas {
//	m := map[string]bool{}
//	for _, rdla := range rdlas {
//		id := rdla.String()
//		m[id] = true
//	}
//	for _, rdla := range rdlas2 {
//		id := rdla.String()
//		if !m[id] {
//			rdlas = append(rdlas, rdla)
//		}
//	}
//	return rdlas
//}
//func (rdlas RuleDotLas) String() string {
//	s := "ruledotlookaheads:\n"
//	for _, rdl := range rdlas {
//		s += fmt.Sprintf("\t%v\n", rdl)
//	}
//	return s
//}

//func rdlaLookahead(rdla *RuleDotLa, rFirst *RulesFirst) {
//	if rdla.parent == nil { // rdla is the start rule
//		rdla.looka.set(rFirst.ri.endRule()) // lookahead is the end rule
//		return
//	}

//	b := []Rule{}
//	rd, ok := rdla.parent.rd.advanceDot()
//	if ok {
//		w := rd.dotAndAfterRules()
//		b = append(b, w...)
//		//if ok {
//		//b = append(b, r)
//		//rdla.looka.add(rFirst.get(r))
//		//}
//	}

//	for _, a := range rdla.parent.looka.sorted() {
//		ha := append(b, a)
//		rset := rFirst.getSequence(ha)
//		rdla.looka.add(rset)
//	}

//	//nilr := rFirst.ri.nilRule()
//	//if len(rdla.looka) == 0 || rdla.looka.isSet(nilr) {
//	//	rdla.looka.add(rdla.parent.looka)
//	//}
//}
