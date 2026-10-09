package nexus

import (
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const consistencyNoReturnVoidId = "returnVoid"

// consistencyNoReturnVoidText is the rule's message, whose wording lives in
// `policy/messages/consistency-no-return-void.json`.
var consistencyNoReturnVoidText = policy.MessageOf("nexus/consistency-no-return-void", consistencyNoReturnVoidId)

func consistencyNoReturnVoidMessage() rule.Message {
	return rule.Message{
		Id:          consistencyNoReturnVoidId,
		Description: consistencyNoReturnVoidText.Render(nil),
	}
}

// ConsistencyNoReturnVoid bans `return void <expression>;`.
//
//	invalid: if (!postId) return void console.log('Usage: ...');
//	valid:   if (!postId) { console.log('Usage: ...'); return; }
//	invalid: return void (first(), second());
//	valid:   onClick={() => void handleAdd()}
//	valid:   void startServer();
//
// # Where it came from
//
// Kirk's rulings of 2026-10-01 on `no-confusing-void-expression` and on the `void`/`undefined` return
// rule: ten `return void console.log('Usage: ...')` lines in one command-line interface were an
// early-exit-and-log trick, and the ruling was to write the call as a statement and then `return;`.
//
// # What it reports, and what it leaves alone
//
// A `return` statement whose expression is a `void` unary, looking through parentheses
// (`return (void f());` is the same statement). Nothing else:
//
//   - An arrow function's expression body (`() => void handleAdd()`) is not a `return` statement. It is
//     the idiom for a callback that starts a promise and returns nothing, it is on every button in the
//     app, and the ruling was about the statement. It stays out of scope.
//   - A `void` expression statement (`void startServer();`) is the dropped-promise marker the message
//     defends.
//   - `return void 0;` is reported like any other operand: it is `return undefined;` spelled to look
//     like something else, and `return;` is what it means.
//
// # The fix, and why it is safe
//
// `return void E;` evaluates `E`, discards it, and returns `undefined`. `E; return;` does exactly the
// same, so the rewrite preserves meaning, with three guards so that the text it writes parses as what
// it says:
//
//  1. The statement must be exactly `return void E;` with only whitespace between the tokens, so no
//     comment is eaten and the parenthesized form is left for a person.
//  2. `E` must begin with an identifier character and not with `function`, `class`, `let` or `async`.
//     As a statement, `E` is parsed from its first token: one opening with `(`, `[`, a template or a
//     unary operator can join the previous line under automatic semicolon insertion, one opening with
//     `{` parses as a block, and the four keywords start a declaration rather than an expression.
//  3. Where the `return` sits in a statement list (a block or a case clause), it becomes two
//     statements, the second on its own line at the same indentation. Where it is the whole body of an
//     `if`, an `else`, a loop or a label, two statements need braces, so it becomes `{ E; return; }`:
//     without them the `return;` would run unconditionally.
//
// The fix always writes `return;`. In a function whose type says it returns a value, that is the bare
// return `nexus/consistency-require-matching-return-type` then rewrites to `return undefined;`, so the
// two fixes converge on the right spelling in one run rather than this rule repeating that judgment.
var ConsistencyNoReturnVoid = rule.Rule{
	Name: "nexus/consistency-no-return-void",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				expression := node.AsReturnStatement().Expression
				if expression == nil {
					return
				}
				voidExpression := ast.SkipParentheses(expression)
				if voidExpression.Kind != ast.KindVoidExpression {
					return
				}

				replacement, fixable := consistencyNoReturnVoidReplacement(ctx, node, voidExpression)
				if !fixable {
					ctx.ReportNode(node, consistencyNoReturnVoidMessage())
					return
				}
				ctx.ReportNodeWithFixes(node, consistencyNoReturnVoidMessage(), ctx.ReplaceNode(node, replacement))
			},
		}
	},
}

// consistencyNoReturnVoidUnsafeLeadingWords start a declaration, not an expression, when they open a
// statement.
var consistencyNoReturnVoidUnsafeLeadingWords = regexp.MustCompile(`^(function|class|let|async)\b`)

// consistencyNoReturnVoidReplacement is the text that replaces the whole `return void E;` statement,
// or false when the rewrite cannot be written safely.
//
// Parentheses around the `void` (`return (void f());`) need no check of their own: the statement text
// then fails the canonical match below, and a mutant removing a separate check survived every fixture.
func consistencyNoReturnVoidReplacement(ctx rule.Context, node *ast.Node, voidExpression *ast.Node) (string, bool) {
	sourceText := ctx.SourceFile.Text()
	statementRange := rule.TokenRange(ctx.SourceFile, node)
	statementText := sourceText[statementRange.Pos():statementRange.End()]

	operand := voidExpression.AsVoidExpression().Expression
	operandRange := rule.TokenRange(ctx.SourceFile, operand)
	operandText := sourceText[operandRange.Pos():operandRange.End()]

	canonical := regexp.MustCompile(`^return\s+void\s+` + regexp.QuoteMeta(operandText) + `\s*;$`)
	if !canonical.MatchString(statementText) {
		return "", false
	}
	first := operandText[0]
	startsWithIdentifier := first == '_' || first == '$' ||
		(first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')
	if !startsWithIdentifier || consistencyNoReturnVoidUnsafeLeadingWords.MatchString(operandText) {
		return "", false
	}

	switch node.Parent.Kind {
	// A `return` is only legal inside a function body, so a file or a namespace body is never its
	// statement list.
	case ast.KindBlock, ast.KindCaseClause, ast.KindDefaultClause:
		lineStart := strings.LastIndexByte(sourceText[:statementRange.Pos()], '\n') + 1
		indentation := sourceText[lineStart:statementRange.Pos()]
		if strings.TrimLeft(indentation, " \t") != "" {
			// Something else shares the line (`case 1: return void f();`), so there is no indentation
			// to copy and the second statement stays beside the first.
			return operandText + "; return;", true
		}
		return operandText + ";\n" + indentation + "return;", true
	}
	return "{ " + operandText + "; return; }", true
}
