package lrparser

import (
	"github.com/friedelschoen/glake/util/parseutil/pscan"
)

type Rule interface {
	id() string
	isTerminal() bool
	childs() []Rule
	iterChildRefs(fn func(index int, ref *Rule) error) error
	String() string
}

// common rule
type CmnRule struct {
	childs2 []Rule
}

//godebug:annotateoff

var defRuleStartSym = "^"   // used in grammar
var defRuleNoPrintSym = "§" // used in grammar

// (0 childs)
type StringRule struct {
	BasicPNode
	CmnRule
	runes   []rune
	rranges []pscan.RuneRange
	typ     stringRType
}

// processor function call rule: allows processing rules at compile time. Ex: string operations.
// (0 childs)
type ProcRule struct {
	BasicPNode
	CmnRule
	name string
	args []ProcRuleArg // allows more then just rules (ex: ints)
}

// (0 childs)
type SingletonRule struct {
	BasicPNode
	CmnRule
	name   string
	isTerm bool
}

func newSingletonRule(name string, isTerm bool) *SingletonRule {
	return &SingletonRule{name: name, isTerm: isTerm}
}

// setup to be available in the grammars at ruleindex.go
var endRule = newSingletonRule("$", true)
var nilRule = newSingletonRule("nil", true)

// special start rule to know start/end (not a terminal)
var startRule = newSingletonRule("^^^", false)

// parenthesis rule type
type parenRType rune

const (
	parenRTNone       parenRType = 0
	parenRTOptional   parenRType = '?'
	parenRTZeroOrMore parenRType = '*'
	parenRTOneOrMore  parenRType = '+'

	// strings related
	parenRTStrOr      parenRType = '%' // individual runes
	parenRTStrOrNeg   parenRType = '!' // individual runes: not
	parenRTStrOrRange parenRType = '-' // individual runes: range
	parenRTStrMid     parenRType = '~' // sequence: middle match
)

// string rule type
type stringRType byte

const (
	stringRTAnd stringRType = iota
	stringRTOr
	stringRTOrNeg
	stringRTMid
)

// ----------

type ProcRuleFn func(args ProcRuleArgs) (Rule, error)
type ProcRuleArg any
type ProcRuleArgs []ProcRuleArg

type RuleSet map[Rule]struct{}
