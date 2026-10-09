package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageDefaultCaseLastNotLast = rule.Message{
	Id: "notLast",
	Description: "This `default` clause is not the last one in its `switch`. A reader scanning " +
		"the clauses in order expects the fallback at the bottom, and a `default` sitting in the " +
		"middle also falls through into whatever case follows it when its body has no `break`, " +
		"which is a control flow almost nobody writes on purpose. Move it to the end.",
}

// DefaultCaseLast flags a `switch` whose `default` clause is not written last.
//
//	valid:   switch (foo) { case 1: break; default: break; }
//	valid:   switch (foo) { default: }
//	valid:   switch (foo) { case 1: case 2: default: }
//	invalid: switch (foo) { default: break; case 1: break; }
//	invalid: switch (foo) { case 1: default: case 2: break; }
//
// # Position is the entire judgment, and a body is not part of it
//
// Upstream finds the index of the clause with no test and compares it against the last index. It
// never looks inside a clause, so an empty `default:` sitting last is clean and a `default:` with a
// full body sitting second-to-last reports. Twelve of the twenty-three clean cases in the corpus
// exist to pin that: they vary the bodies, the break statements and the number of neighbouring
// cases while keeping the position legal, and every one of them would still be clean if the rule
// read no body at all.
//
// The corpus also carries `switch (foo) { case 1: default: }`, where `default` is last while sharing
// a fallthrough group with `case 1`. Upstream reports on index rather than on grouping, so that is
// clean, and a port reasoning about fallthrough groups instead of about clause order would report
// it.
//
// # Why this anchors on the switch rather than on the clause
//
// A `KindDefaultClause` listener would fire on the right node and could not answer the question:
// deciding whether it is last needs the sibling list, which means climbing back to the case block
// anyway. Anchoring on the statement reads the clause list once and asks the only question there
// is.
//
// # The reported span is the default clause, not the switch
//
// Upstream passes `node: defaultClause`, and its corpus asserts the column of the `default` keyword
// on all fourteen reporting cases rather than the column of `switch`. A port reporting the whole
// statement satisfies every message-id assertion and points the reader at the wrong line in a long
// switch, which is exactly the case where the finding is worth having.
var DefaultCaseLast = rule.Rule{
	Name: "default-case-last",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				clauses := node.AsSwitchStatement().CaseBlock.AsCaseBlock().Clauses.Nodes
				for index, clause := range clauses {
					if clause.Kind != ast.KindDefaultClause {
						continue
					}
					// Upstream takes the FIRST clause with no test and stops. A switch holding two
					// `default` clauses is a syntax error the parser still recovers from, so the
					// shape is reachable here, and upstream would report it once rather than twice.
					if index != len(clauses)-1 {
						ctx.ReportNode(clause, messageDefaultCaseLastNotLast)
					}
					return
				}
			},
		}
	},
}
