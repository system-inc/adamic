package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-empty-object-type.
//
// Its own file rather than a line in the package's register.go, so two ports landing at once do not
// conflict on a file neither is really changing. The decoder compiles allowWithName as JavaScript does.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoEmptyObjectType,
		Decode: DecodeNoEmptyObjectTypeOptions,
	})
}
