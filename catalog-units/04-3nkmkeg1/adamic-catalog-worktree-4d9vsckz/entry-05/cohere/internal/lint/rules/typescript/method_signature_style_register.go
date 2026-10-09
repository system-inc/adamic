package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/method-signature-style.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// `Decode` names the rule's own hand-written decoder rather than `rule.DecodeOptionsInto`. Upstream's
// option is a bare JSON string rather than an object, and its default is `property` rather than a Go
// zero value, so a generic decode would hand the rule an empty style matching neither arm and the
// rule would register on every file and report nothing while every fixture stayed green.
func init() {
	rule.Register(rule.Registration{Rule: MethodSignatureStyle, Decode: DecodeMethodSignatureStyleOptions})
}
