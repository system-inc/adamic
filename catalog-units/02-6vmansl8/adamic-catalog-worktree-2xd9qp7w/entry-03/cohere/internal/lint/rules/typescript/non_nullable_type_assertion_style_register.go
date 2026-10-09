package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/non-nullable-type-assertion-style.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// No Decode entry, because `meta.schema` is `[]` upstream and the rule reads no options at all.
func init() {
	rule.Register(rule.Registration{Rule: NonNullableTypeAssertionStyle})
}
