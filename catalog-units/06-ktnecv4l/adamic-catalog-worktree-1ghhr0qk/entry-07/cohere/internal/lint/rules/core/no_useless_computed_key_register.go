package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-useless-computed-key.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// The decoder is hand rolled because the single option defaults to TRUE. `rule.DecodeOptionsInto`
// yields the zero value on an absent option, which here would silently narrow the rule to object
// literals rather than disable it, and no fixture routed through a struct could see the difference.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoUselessComputedKey,
		Decode: DecodeNoUselessComputedKeyOptions,
	})
}
