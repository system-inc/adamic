package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers max-classes-per-file.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled because upstream's option is a `oneOf` over an integer and an object,
// and because its default is 1 rather than the zero value. A generic decoder would fail the integer
// shape outright and would turn an absent option into a maximum of zero, which reports every file
// holding a single class.
func init() {
	rule.Register(rule.Registration{
		Rule:   MaxClassesPerFile,
		Decode: DecodeMaxClassesPerFileOptions,
	})
}
