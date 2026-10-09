package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers prefer-promise-reject-errors.
//
// Its own file rather than a line in the package's register.go, which is one file every concurrent
// port has to edit.
//
// The generic decoder is safe here because the rule's single option defaults to false, so a rule
// handed nil options gets a zero value that already behaves correctly.
func init() {
	rule.Register(rule.Registration{
		Rule:   PreferPromiseRejectErrors,
		Decode: rule.DecodeOptionsInto[PreferPromiseRejectErrorsOptions](),
	})
}
