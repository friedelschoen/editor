package lrparser

// unique rule index
type RuleIndex struct {
	m  map[string]*Rule
	pm map[string]ProcRuleFn

	deref struct {
		once bool
		err  error
	}
}

//godebug:annotateoff
