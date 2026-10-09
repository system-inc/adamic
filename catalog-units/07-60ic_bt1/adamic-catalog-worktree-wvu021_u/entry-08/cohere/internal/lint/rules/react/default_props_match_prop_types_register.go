package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `default-props-match-prop-types` rule.
//
// No `RequiresOptions`, unlike the sibling `boolean-prop-naming`. That rule reads its pattern from
// the options and does nothing without them; this one reads only a boolean that defaults to false,
// so a bare `"error"` is a working configuration. Measured: the same input reports identically with
// no options and with an empty object.
func init() {
	rule.Register(rule.Registration{
		Rule:   DefaultPropsMatchPropTypes,
		Decode: DecodeDefaultPropsMatchPropTypesOptions,
	})
}
