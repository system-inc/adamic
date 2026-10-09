package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers id-match.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// DecodeOptionList rather than Decode, because upstream's option surface is two elements, a pattern
// string followed by a flag object: `["error", "^[a-z]+$", {"properties": true}]`. The decoder is
// hand rolled because that heterogeneous list is something no single Go struct decodes.
//
// `RequiresOptions` is deliberately NOT set. Upstream's default pattern is `^.+$`, which every
// non-empty name matches, so an unconfigured rule reporting nothing is upstream's own behaviour
// rather than an inert-rule defect.
func init() {
	rule.Register(rule.Registration{
		Rule:             IdMatch,
		DecodeOptionList: DecodeIdMatchOptions,
	})
}
