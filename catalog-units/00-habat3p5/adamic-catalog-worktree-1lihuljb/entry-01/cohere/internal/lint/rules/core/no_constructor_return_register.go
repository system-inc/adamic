package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-constructor-return.
//
// A file of its own rather than a line in the package's `register.go`, which is the shared list
// every porter used to edit and conflict on. `rule.Register` appends into a map keyed by name and is
// callable from any file's init, so a rule is three new files and no shared edit.
//
// No `Decode`, because upstream's `meta.schema` is the empty array: this rule has no option surface
// at all, and the `options any` parameter its Run takes is the signature every rule shares rather
// than a hint that there is something to decode.
func init() {
	rule.Register(rule.Registration{Rule: NoConstructorReturn})
}
