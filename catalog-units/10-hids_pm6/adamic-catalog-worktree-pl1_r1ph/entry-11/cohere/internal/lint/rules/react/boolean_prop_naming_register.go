package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `boolean-prop-naming` rule.
//
// `RequiresOptions` is the load-bearing field. Upstream reads its pattern from `context.options[0]`
// and ESLint fills a schema default only into an options object that is present, so the rule is
// entirely inert when configured as a bare `"error"`. Marking it required turns that silence into a
// configuration failure instead of a rule that passes every fixture and lints nothing.
func init() {
	rule.Register(rule.Registration{
		Rule:            BooleanPropNaming,
		Decode:          DecodeBooleanPropNamingOptions,
		RequiresOptions: true,
	})
}
