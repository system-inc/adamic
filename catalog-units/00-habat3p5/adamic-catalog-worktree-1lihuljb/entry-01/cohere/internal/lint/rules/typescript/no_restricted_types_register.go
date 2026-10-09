package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-restricted-types.
//
// Its own file rather than a line in the package's register.go, so two ports landing at once do not
// conflict on a file neither is really changing.
//
// `Decode` names the rule's own decoder rather than `rule.DecodeOptionsInto`. Upstream's `types`
// values are a `oneOf` over a boolean, a string, and an object, and a generic decode into any single
// Go type would discard two of the three shapes and silently ban nothing.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoRestrictedTypes,
		Decode: DecodeNoRestrictedTypesOptions,
	})
}
