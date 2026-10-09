package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/prefer-literal-enum-member.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// The Decode entry names the rule's own decoder rather than the generic helper. This rule's single
// option defaults to FALSE, so the generic one would happen to produce the right zero value, and the
// hand-rolled decoder is used anyway: relying on that coincidence leaves the next person to add a
// default-true key here inheriting a decoder that silently inverts their rule.
func init() {
	rule.Register(rule.Registration{
		Rule:   PreferLiteralEnumMember,
		Decode: DecodePreferLiteralEnumMemberOptions,
	})
}
