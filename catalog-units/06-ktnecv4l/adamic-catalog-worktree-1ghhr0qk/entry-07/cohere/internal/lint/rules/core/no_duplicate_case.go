package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageDuplicateCase = rule.Message{
	Id: "unexpected",
	Description: "Duplicate case label: an earlier arm in this switch tests the same expression, so " +
		"this one can never run. Almost always a clause was copied and its test never updated, which " +
		"means the body here is dead and the case it was meant to handle falls through to `default`. " +
		"Change the test to the value this arm was written for, or delete the arm.",
}

// NoDuplicateCase flags a `case` whose test is written the same way as an earlier one in the same
// switch.
//
//	valid:   switch (a) { case 1: break; case 2: break; }
//	valid:   switch (a) { case 1: break; case '1': break; }
//	valid:   switch (a) { case one: break; case 2: break; }
//	invalid: switch (a) { case 1: break; case 1: break; }
//
// The second arm is unreachable, because a switch takes the first arm that matches. That is a
// mistake in every codebase, which is why this rule carries no configuration.
//
// Comparison is by how the test is written, not by what it evaluates to. `case 1` and `case one`
// stay distinct even where `one` is 1, and `case 1` and `case '1'` stay distinct because a switch
// compares with `===`. Going the other way, by value, would need the test expressions evaluated,
// which a linter cannot do and which would be wrong anyway for a test that has side effects.
//
// The rule is deliberately silent about `default`, which has no test and so cannot duplicate one.
// A switch with two `default` clauses is a syntax error the parser already refuses.
//
// Divergence from ESLint, both in our favor and both worth knowing:
//
//   - ESLint reports the whole `SwitchCase` node, which spans the clause body. We report the test
//     expression, because the body is not what is wrong and a finding whose range covers fifty
//     lines is one a reader has to hunt through. The message names the defect precisely enough that
//     the narrower range loses nothing.
//   - ESLint parses to ESTree, which does not keep ParenthesizedExpression, so `case (1)` and
//     `case 1` are one expression there and two here. Ours is the stricter reading; it can only miss
//     a duplicate that was written with different parentheses, never invent one.
//
// No fix. Which arm is the mistake and what its test should have been is the author's knowledge,
// and deleting the wrong one deletes a body that was meant to run.
var NoDuplicateCase = rule.Rule{
	Name: "no-duplicate-case",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				switchStatement := node.AsSwitchStatement()
				if switchStatement == nil || switchStatement.CaseBlock == nil {
					return
				}
				caseBlock := switchStatement.CaseBlock.AsCaseBlock()
				if caseBlock == nil || caseBlock.Clauses == nil {
					return
				}

				// Scoped to one switch: an arm only shadows an arm it competes with, so two switches
				// in the same file testing the same value are both correct.
				seenTests := map[string]bool{}

				for _, clauseNode := range caseBlock.Clauses.Nodes {
					if clauseNode.Kind != ast.KindCaseClause {
						continue
					}
					clause := clauseNode.AsCaseOrDefaultClause()
					if clause == nil || clause.Expression == nil {
						continue
					}

					signature := tokenSignature(ctx.SourceFile, clause.Expression)
					if seenTests[signature] {
						// Report the later arm. The first one is the one that runs, so it is the
						// one that is probably right.
						ctx.Report(rule.Diagnostic{
							Range:      rule.TokenRange(ctx.SourceFile, clause.Expression),
							Message:    messageDuplicateCase,
							SourceFile: ctx.SourceFile,
						})
						continue
					}
					seenTests[signature] = true
				}
			},
		}
	},
}
