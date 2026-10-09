package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-restricted-globals.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// DecodeOptionList rather than Decode, because upstream's option surface is variadic: each
// restricted global is its own element, `["error", "event", "fdescribe"]`, or a single element is
// an object carrying that list plus two flags. Each entry is itself either a string or an object,
// and the generic helper expresses none of those alternatives.
func init() {
	rule.Register(rule.Registration{
		Rule:             NoRestrictedGlobals,
		DecodeOptionList: DecodeNoRestrictedGlobalsOptions,
	})
}
