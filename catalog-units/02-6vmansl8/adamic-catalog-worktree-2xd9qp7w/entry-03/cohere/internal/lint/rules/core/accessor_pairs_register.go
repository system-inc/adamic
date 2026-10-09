package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers accessor-pairs.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled rather than `rule.DecodeOptionsInto`, and that is the part worth
// reading. The generic helper errors on empty input, the config layer turns that error into nil for
// a non-required rule, and a nil handed to a rule reading bool fields yields false for all of them.
// Two of this rule's four options default to TRUE, so that path would silently switch off the class
// check and the setter-without-getter check, leaving a rule that registers everywhere and reports
// almost nothing. `DecodeAccessorPairsOptions` returns the defaults on empty input instead, and the
// fields behind it are pointers so an absent key stays distinguishable from an explicit false.
func init() {
	rule.Register(rule.Registration{
		Rule:   AccessorPairs,
		Decode: DecodeAccessorPairsOptions,
	})
}
