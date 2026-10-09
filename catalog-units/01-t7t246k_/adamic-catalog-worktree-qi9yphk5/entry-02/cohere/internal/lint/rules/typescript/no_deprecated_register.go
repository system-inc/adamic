package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-deprecated.
//
// Its own file rather than a line in a shared list: that list is one file every concurrent port has
// to edit, so two ports landing at once conflict on a file neither is really changing.
//
// The Decode entry is required rather than defensive. `allow` is one heterogeneous array whose two
// wire forms, a bare string and a specifier object, land in two different fields of the struct the
// rule reads, and whose `from` is an enum rather than a string. The generic decoder cannot express
// either, so without this the option would silently do nothing.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoDeprecated,
		Decode: DecodeNoDeprecatedOptions,
	})
}
