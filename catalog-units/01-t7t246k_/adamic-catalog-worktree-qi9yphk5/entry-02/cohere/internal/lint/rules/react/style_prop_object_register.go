package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `style-prop-object` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit.
//
// The decoder is hand-rolled rather than `rule.DecodeOptionsInto` so an empty body answers with an
// empty allow list rather than an error. Note the direction: an empty list here means the rule
// examines EVERY element, which is the opposite of the two sibling forbid rules where an empty list
// means the rule declines everything.
func init() {
	rule.Register(rule.Registration{
		Rule:   StylePropObject,
		Decode: DecodeStylePropObjectOptions,
	})
}
