package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-use-before-define.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled rather than `rule.DecodeOptionsInto`, for two reasons that compound.
// Six of this rule's seven options default to TRUE, and the generic helper errors on empty input,
// which the config layer turns into nil for a non-required rule, which a struct assertion turns into
// the zero value: false for all of them. That path leaves a rule that registers on every file and
// judges almost nothing. And upstream's schema is a `oneOf`, so the option can arrive as the bare
// string `"nofunc"` rather than an object, which no struct unmarshal can accept.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoUseBeforeDefine,
		Decode: DecodeNoUseBeforeDefineOptions,
	})
}
