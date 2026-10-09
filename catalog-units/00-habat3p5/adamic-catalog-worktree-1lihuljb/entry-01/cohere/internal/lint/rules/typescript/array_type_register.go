package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/array-type.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// `Decode` names the rule's own hand-written decoder rather than `rule.DecodeOptionsInto`. Neither
// of this rule's two defaults is a Go zero value, and `readonly` falls back to the RESOLVED value of
// `default` rather than to a constant, which is a dependency between two fields no struct tag can
// express. A generic decode would hand the rule two empty settings that match no arm, and the rule
// would register on every file and report nothing while every fixture stayed green.
func init() {
	rule.Register(rule.Registration{Rule: ArrayType, Decode: DecodeArrayTypeOptions})
}
