package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers id-denylist.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// DecodeOptionList rather than Decode, because upstream's option surface is variadic: each denied
// name is its own element, `["error", "data", "err"]`. The decoder is hand rolled because that list
// is of names rather than an object, which `rule.DecodeOptionsInto` cannot express through a struct.
//
// `RequiresOptions` is deliberately NOT set. The rule denies nothing when unconfigured, which is
// upstream's own behaviour rather than an inert-rule defect: its denylist is built from
// `context.options` and an unconfigured rule simply has none.
func init() {
	rule.Register(rule.Registration{
		Rule:             IdDenylist,
		DecodeOptionList: DecodeIdDenylistOptions,
	})
}
