package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers unicode-bom.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled rather than `rule.DecodeOptionsInto` because this rule's option
// defaults to `never` and its wire value is a bare string. The generic helper errors on empty input,
// the config layer turns that into nil, and the rule would read nil as a zero-valued setting that
// matches neither arm and reports nothing at all. That is the live config's own shape, so the
// fallback is the ordinary path here rather than a defensive one.
func init() {
	rule.Register(rule.Registration{
		Rule:   UnicodeBom,
		Decode: DecodeUnicodeBomOptions,
	})
}
