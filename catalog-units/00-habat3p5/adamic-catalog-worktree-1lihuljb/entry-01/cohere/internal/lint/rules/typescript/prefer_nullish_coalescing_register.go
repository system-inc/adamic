package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/prefer-nullish-coalescing.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is changing.
//
// The Decode entry is required rather than defensive, for two reasons beyond the usual one.
// `ignoreConditionalTests` defaults TRUE, so a rule handed a nil options value must still get that
// default or it reports on every conditional test in the tree. And `ignorePrimitives` is a
// `oneOf(object, enum:[true])`, which no struct tag can express and which the generic decoder would
// reject outright.
func init() {
	rule.Register(rule.Registration{
		Rule:   PreferNullishCoalescing,
		Decode: DecodePreferNullishCoalescingOptions,
	})
}
