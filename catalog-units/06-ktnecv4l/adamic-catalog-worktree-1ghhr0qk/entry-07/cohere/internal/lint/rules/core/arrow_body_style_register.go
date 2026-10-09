package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers arrow-body-style.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// DecodeOptionList rather than Decode, because upstream's option surface is a two-element list:
// `["error", "as-needed", {"requireReturnForObjectLiteral": true}]`. The decoder is hand-rolled so an
// unrecognized mode fails loudly. `rule.DecodeOptionsInto` would
// leave an unknown string in the field, and this rule's three modes disagree about every input, so a
// typo would select a silent fourth behaviour of reporting nothing rather than producing an error.
// The same function also supplies the default mode, which upstream carries as `defaultOptions` and
// which a nil handed to a bare-severity rule would otherwise leave empty.
func init() {
	rule.Register(rule.Registration{
		Rule:             ArrowBodyStyle,
		DecodeOptionList: DecodeArrowBodyStyleOptions,
	})
}
