package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoWith = rule.Message{
	Id: "noWith",
	Description: "A `with` block splices an object's properties into the local scope, so a bare " +
		"name inside it cannot be resolved by reading the code: whether `x` means the object's " +
		"property or an outer variable is decided at runtime by what the object happens to hold. " +
		"That defeats the checker and every reader. Bind what you need to a name instead, as in " +
		"`const { x, y } = point;`.",
}

// NoWith flags a `with` statement.
//
//	valid:   var obj = { with: 1 }; obj.with;
//	valid:   const { with: w } = { with: 4 }; w;
//	invalid: with (foo) { bar() }
//
// Purely syntactic. Both upstream implementations match on the node kind alone and ask nothing of a
// scope table or a type: oxc's `run` is a single `AstKind::WithStatement` arm, and ESLint's `create`
// returns one `WithStatement` visitor. `meta.schema` is `[]`, so there are no options, and
// `meta.messages` carries exactly one id, so an input reporting twice is two statements rather than
// two judgments about one.
//
// The span is the only place this rule can go wrong, and it is reported deliberately rather than by
// the obvious spelling. Upstream underlines the `with` keyword alone, not the statement: oxc writes
// `Span::sized(with_statement.span.start, 4)` and ESLint takes `getFirstToken(node).loc`. So this
// reaches for the first token's range instead of `ReportNode`, whose `TokenRange` extends to the
// node's end and would underline the entire block.
//
// A note on whether this rule can fire at all. `with` is a syntax error in strict mode and every ES
// module is strict, which suggests the rule is dead on a TypeScript tree. It is not: TypeScript's
// parser is error-tolerant and builds the node anyway, raising the strict-mode violation as a
// grammar diagnostic during checking, so an AST walk still sees it. What that costs upstream states
// plainly, that the rule is unnecessary when `alwaysStrict` is on. It is kept because it is free at
// runtime, it names the problem better than the compiler's grammar error does, and it holds for the
// `.js` files in the tree where no such check runs.
var NoWith = rule.Rule{
	Name: "no-with",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindWithStatement: func(node *ast.Node) {
				// The first token of the statement, which is the `with` keyword. Built from the
				// scanner rather than from `node.Pos()` directly, because `Pos()` sits before the
				// node's leading comments and whitespace and would underline those instead.
				ctx.ReportRange(
					scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos()),
					messageNoWith)
			},
		}
	},
}
