package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageCaseDeclaration = rule.Message{
	Id: "unexpected",
	Description: "A `let`, `const`, `class`, or `function` declared directly in a case clause is " +
		"scoped to the whole switch, not to this arm. Every other arm can see the binding, but only " +
		"this arm initializes it, so a sibling arm that reaches it throws on a binding that appears " +
		"to be in scope. Wrap this clause's body in braces to scope the binding to the arm that " +
		"declares it.",
}

var messageAddBraces = rule.Message{
	Id:          "addBraces",
	Description: "Wrap the clause body in braces.",
}

// NoCaseDeclarations flags a lexical declaration sitting directly in a `case` or `default` clause.
//
//	valid:   switch (a) { case 1: { const x = 1; break; } }
//	valid:   switch (a) { case 1: var x = 1; break; }
//	valid:   switch (a) { case 1: if (a) { const x = 1; } break; }
//	invalid: switch (a) { case 1: const x = 1; break; }
//
// The clauses of a switch share one block scope, so a binding declared in one arm exists in all of
// them while only its own arm ever runs the initializer. A sibling arm touching it hits the temporal
// dead zone and throws, on a name that is visibly in scope. A hoisted `function` is worse in the
// other direction: it is callable from arms that never declared it, which reads as a bug in whatever
// arm called it.
//
// `var` is exempt because it is function-scoped and hoisted, so it has neither hazard. It is
// arguably the worse style, but that is a different rule's argument.
//
// A declaration nested inside a block or a statement within the clause is already scoped, so only
// direct children of the clause are examined. That is the whole of the original's traversal: ESLint
// loops the clause's own statement list and never descends.
//
// Where ESLint has nothing to say, since its AST is JavaScript, we exempt TypeScript's own
// declarations. `type` and `interface` are erased before anything runs and cannot be in a dead zone.
// `enum` and `namespace` do emit runtime values, but they emit an assignment into a hoisted `var`
// binding, which puts them on the exempt side of the same line `var` sits on rather than the
// reported side. `declare` forms emit nothing at all. rslint reaches the same set.
//
// `using` and `await using` are reported: both are block-scoped exactly as `const` is, and their
// disposal is tied to the scope they belong to, so a switch-wide scope also means switch-wide
// lifetime. ESLint's own check is `kind !== 'var'`, which includes them, so this is the original's
// reading rather than an extension of it.
//
// The repair is a suggestion, not a fix. Wrapping the body changes what the code means: any binding
// a later arm was reading through fall-through stops being visible, and only the author knows
// whether some arm was doing that on purpose. Following ESLint, the braces wrap the entire clause
// body rather than the reported statement, so several declarations in one clause all offer the same
// whole-clause repair rather than proposing nested blocks that would fight each other.
var NoCaseDeclarations = rule.Rule{
	Name: "no-case-declarations",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// Listening on the case block rather than on each clause visits one node per switch instead
		// of one per arm, and hands the clause list already assembled.
		return rule.Listeners{
			ast.KindCaseBlock: func(node *ast.Node) {
				caseBlock := node.AsCaseBlock()
				if caseBlock == nil || caseBlock.Clauses == nil {
					return
				}

				for _, clauseNode := range caseBlock.Clauses.Nodes {
					clause := clauseNode.AsCaseOrDefaultClause()
					if clause == nil || clause.Statements == nil {
						continue
					}
					statements := clause.Statements.Nodes
					if len(statements) == 0 {
						continue
					}

					firstStatement := statements[0]
					lastStatement := statements[len(statements)-1]

					for _, statement := range statements {
						if !isLexicalDeclaration(statement) {
							continue
						}
						ctx.ReportNodeWithSuggestions(statement, messageCaseDeclaration, rule.Suggestion{
							Message: messageAddBraces,
							Fixes: []rule.Fix{
								ctx.InsertBefore(firstStatement, "{ "),
								ctx.InsertAfter(lastStatement, " }"),
							},
						})
					}
				}
			},
		}
	},
}

// isLexicalDeclaration says whether a statement introduces a binding scoped to the enclosing block
// rather than to the enclosing function.
//
// The variable case reads NodeFlagsBlockScoped rather than testing for `let` and `const` by name,
// which is what makes `using` and `await using` land on the correct side without naming them: the
// parser already decided which declaration kinds are block-scoped, and that decision is the one the
// rule is about. NodeFlagsAmbient excludes `declare` forms, which emit nothing and so cannot be
// reached in a dead zone.
func isLexicalDeclaration(statement *ast.Node) bool {
	switch statement.Kind {
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
		return statement.Flags&ast.NodeFlagsAmbient == 0
	case ast.KindVariableStatement:
		variableStatement := statement.AsVariableStatement()
		if variableStatement == nil || variableStatement.DeclarationList == nil {
			return false
		}
		declarationFlags := variableStatement.DeclarationList.Flags
		return declarationFlags&ast.NodeFlagsBlockScoped != 0 && declarationFlags&ast.NodeFlagsAmbient == 0
	}
	return false
}
