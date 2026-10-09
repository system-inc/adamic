package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/consistent-generic-constructors.
//
// Its own file rather than a line in the package's register.go, so two ports landing at once do not
// conflict on a file neither is really changing.
//
// The decoder is hand-rolled rather than generic, because upstream's option is a bare STRING rather
// than an object and the generic helper cannot read one. See its doc comment for the wire shapes.
func init() {
	rule.Register(rule.Registration{
		Rule:   ConsistentGenericConstructors,
		Decode: DecodeConsistentGenericConstructorsOptions,
	})
}
