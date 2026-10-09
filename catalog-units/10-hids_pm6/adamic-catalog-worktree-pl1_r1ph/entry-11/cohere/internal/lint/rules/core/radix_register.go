package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers radix.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled because upstream's single option is a bare string enum, `"always"` or
// `"as-needed"`, which `rule.DecodeOptionsInto` cannot read. The option is deprecated upstream and
// changes nothing the rule reports, but its schema still accepts it, so `["error", "as-needed"]` has
// to load rather than stop the run. A value outside the enum is still refused.
func init() {
	rule.Register(rule.Registration{
		Rule:   Radix,
		Decode: DecodeRadixOptions,
	})
}
