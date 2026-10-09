package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers logical-assignment-operators. DecodeOptionList rather than Decode, because
// upstream's option surface is a two-element list: `["error", "always", {"enforceForIfStatements":
// true}]`.
func init() {
	rule.Register(rule.Registration{
		Rule:             LogicalAssignmentOperators,
		DecodeOptionList: DecodeLogicalAssignmentOperatorsOptions,
	})
}
