package lrparser

type RuleDot struct { // also know as "item"
	prod    Rule // producer: "prod->rule"
	rule    Rule // producee: where the dot runs
	dot     int  // rule dot
	reverse bool
}

type RuleDots []*RuleDot
