package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoDynamicDelete flags `delete` applied to a computed member whose key is not a plain literal.
//
//	valid:   const container: { [i: string]: 0 } = {}; delete container.aaa;
//	valid:   const container: { [i: string]: 0 } = {}; delete container[7];
//	valid:   const container: { [i: string]: 0 } = {}; delete container[-7];
//	valid:   const container: { [i: string]: 0 } = {}; delete container['aaa'];
//	invalid: const container: { [i: string]: 0 } = {}; delete container['aa' + 'b'];
//	invalid: const container: { [i: string]: 0 } = {}; delete container[+7];
//	invalid: const container: { [i: string]: 0 } = {}; delete container[name];
//
// Deleting a statically-written key is a spelling the reader can check against the type. Deleting a
// computed one hides which property leaves the object, and on a type with declared members it
// usually means the object is being used as a map and should have been one.
//
// The rule ships no fixer: every one of upstream's ten failing cases carries `output: null`, and
// there is no rewrite that preserves meaning, since the whole point is that the key is unknown
// until run time.
//
// # What decides a finding
//
// The anchor is `KindDeleteExpression`. The operand must be an element access, and its key must not
// be an ACCEPTABLE INDEX, which upstream defines narrowly:
//
//	a Literal whose runtime value is a number or a string
//	a unary `-` applied to a Literal whose value is a number
//
// Everything else reports, including several shapes that read as static to a human. `+7` is a unary
// PLUS and upstream accepts only minus, so it reports while `-7` is clean. `-Infinity` and `NaN`
// are identifiers rather than literals, so both report. A template literal is not a Literal in the
// estree grammar even when it has no substitutions, so a delete whose key is a backtick-quoted
// string reports while the same key in single quotes does not.
//
// # Where the estree grammar and our parse tree differ, and it is load-bearing three times
//
// Upstream tests `property.type === Literal`, and estree's `Literal` covers string, number,
// boolean, null, regex and bigint under one node type, discriminating on `typeof property.value`.
// Our parser gives each of those its own kind, so the port names the two accepted kinds directly
// rather than testing a value. That is the same decision expressed against a different grammar, and
// the shapes that separate them were measured against the installed 8.x build rather than reasoned
// about:
//
//	delete c[true]   REPORT   typeof true is 'boolean', not in the accepted pair
//	delete c[null]   REPORT   typeof null is 'object'
//	delete c[/re/]   REPORT   typeof a regex is 'object'
//	delete c[7n]     REPORT   typeof 7n is 'bigint'
//
// All four report upstream and all four report here, and none of them appears in the corpus.
//
// PARENTHESES ARE INVISIBLE IN ESTREE and are real nodes here, which is the divergence that would
// have shipped silently. estree has no parenthesized-expression node at all, so upstream's `Literal`
// test sees straight through `delete c[(7)]` and stays clean. A port testing our kinds without
// skipping parentheses would report it. Measured against the installed build, and the corpus writes
// no parenthesized form anywhere:
//
//	delete c[(7)]        clean    the paren is not part of the estree tree
//	delete c[-(7)]       clean    skipped under the minus arm too
//	delete ((c[x]))      REPORT   parens around the whole operand, span excludes them
//	delete c[(y, 'b')]   REPORT   a sequence is not a literal; the span is `y, 'b'` without parens
//
// So parentheses are skipped in three places: the delete operand, the key, and the minus operand.
// The last of those is the one no reading of the rule file suggests.
//
// AN OPTIONAL CHAIN IS EXCLUDED, and not through the key test. estree wraps `c?.[x]` in a
// `ChainExpression`, so `node.argument.type !== MemberExpression` is already false and the rule
// returns before it ever looks at the key. Both spellings are clean upstream regardless of what the
// key is, which is what proves the exclusion is structural rather than a consequence of the index:
//
//	delete c?.[x]   clean
//	delete c?.[7]   clean
//
// Our parse gives the element access a non-nil `QuestionDotToken` instead of a wrapper node, so the
// exclusion is written as that test. Reproduced rather than improved: deleting through an optional
// chain is exactly as dynamic as deleting without one, and upstream's silence looks like a
// consequence of its grammar rather than a judgment. It is reproduced because the differential
// harness compares against upstream, and a finding upstream does not produce is a divergence
// whichever direction it points.
//
// # The span is the KEY, not the delete expression
//
// Upstream reports on `node.argument.property`, so `delete container['aa' + 'b']` underlines
// `'aa' + 'b'` and nothing else. That is a deliberate choice worth pinning: the surrounding delete
// is fine, the key is the problem, and a finding anchored on the whole statement would point the
// reader at the wrong token. Every reporting fixture in this package asserts the sliced span, since
// a message-id assertion cannot see where a finding lands.
//
// # Cost
//
// `KindDeleteExpression` is a rare anchor and the listener exits on the first test for anything that
// is not an element access. No checker, no program, no per-file state.
var NoDynamicDelete = rule.Rule{
	Name: "@typescript-eslint/no-dynamic-delete",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindDeleteExpression: func(node *ast.Node) {
				// SkipParentheses dereferences its argument, so the operand is read from the node
				// rather than from a helper that could answer nil. A delete expression always
				// carries an operand: the parser synthesizes a missing identifier when the source
				// is `delete ;`, so there is no shape where this field is absent.
				operand := ast.SkipParentheses(node.AsDeleteExpression().Expression)

				if !ast.IsElementAccessExpression(operand) {
					return
				}
				elementAccess := operand.AsElementAccessExpression()

				// An optional chain is silent upstream through its grammar rather than through its
				// key, so this test comes before the key test and not after it.
				if elementAccess.QuestionDotToken != nil {
					return
				}

				key := ast.SkipParentheses(elementAccess.ArgumentExpression)
				if isAcceptableDeleteIndexExpression(key) {
					return
				}

				ctx.ReportNode(key, buildDynamicDeleteMessage())
			},
		}
	},
}

// isAcceptableDeleteIndexExpression answers upstream's `isAcceptableIndexExpression`.
//
// The two accepted shapes are a string or numeric literal, and a unary minus over a numeric
// literal. Written against our kinds rather than against an estree `typeof`, for the reasons in the
// rule's doc comment.
//
// The key is already parenthesis-skipped by the caller; the minus operand is skipped here, because
// `delete c[-(7)]` is clean upstream and nothing about the rule file says so.
func isAcceptableDeleteIndexExpression(key *ast.Node) bool {
	if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNumericLiteral {
		return true
	}

	if ast.IsPrefixUnaryExpression(key) {
		unary := key.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindMinusToken {
			return ast.SkipParentheses(unary.Operand).Kind == ast.KindNumericLiteral
		}
	}

	return false
}

func buildDynamicDeleteMessage() rule.Message {
	return rule.Message{
		Id:          "dynamicDelete",
		Description: "Do not delete dynamically computed property keys.",
	}
}
