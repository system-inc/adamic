package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `class-methods-use-this` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. The decoder is hand-written because `enforceForClassFields` defaults to TRUE and the
// generic helper would decode an absent option to false, dropping every class-field finding; see
// DecodeClassMethodsUseThisOptions.
func init() {
	rule.Register(rule.Registration{Rule: ClassMethodsUseThis, Decode: DecodeClassMethodsUseThisOptions})
}
