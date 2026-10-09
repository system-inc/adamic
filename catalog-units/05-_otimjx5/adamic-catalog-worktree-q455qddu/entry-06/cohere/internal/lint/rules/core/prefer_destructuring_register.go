package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers prefer-destructuring.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is changing.
//
// DecodeOptionList rather than Decode, because upstream's option surface is two schema elements and
// `enforceForRenamedProperties` lives in the second. Registered with Decode, the config layer would
// refuse that second element; with DecodeOptionList the decoder is handed both.
func init() {
	rule.Register(rule.Registration{
		Rule:             PreferDestructuring,
		DecodeOptionList: DecodePreferDestructuringOptions,
	})
}
