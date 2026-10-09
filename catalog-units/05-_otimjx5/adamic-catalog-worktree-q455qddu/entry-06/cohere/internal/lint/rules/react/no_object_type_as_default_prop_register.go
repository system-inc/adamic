package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `no-object-type-as-default-prop` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. No decoder, because upstream declares no schema at all.
func init() {
	rule.Register(rule.Registration{Rule: NoObjectTypeAsDefaultProp})
}
