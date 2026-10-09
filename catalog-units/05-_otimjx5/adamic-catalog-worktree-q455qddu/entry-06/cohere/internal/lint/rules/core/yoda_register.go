package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers yoda.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// DecodeOptionList rather than Decode, because upstream's option is a two element list, an enum then
// an object: `["error", "always", {"exceptRange": true}]`. The decoder is hand rolled because that
// heterogeneous list is something the generic helper cannot express.
func init() {
	rule.Register(rule.Registration{
		Rule:             Yoda,
		DecodeOptionList: DecodeYodaOptions,
	})
}
