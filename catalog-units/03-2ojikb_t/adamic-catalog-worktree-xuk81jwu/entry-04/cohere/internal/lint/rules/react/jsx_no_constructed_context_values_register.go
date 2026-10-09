package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `jsx-no-constructed-context-values` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. No `Decode`, because upstream's `schema: false` is the authoritative statement that
// this rule takes no options.
func init() {
	rule.Register(rule.Registration{Rule: JsxNoConstructedContextValues})
}
