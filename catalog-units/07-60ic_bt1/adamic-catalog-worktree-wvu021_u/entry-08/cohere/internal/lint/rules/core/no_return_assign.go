package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageReturnAssignment = rule.Message{
	Id: "returnAssignment",
	Description: "This return statement assigns rather than returns a value it already has, so it " +
		"does two things at once and the assignment is the one the reader is likeliest to miss. " +
		"It is usually a `==` typed as `=`, and where it is deliberate it reads as a value being " +
		"returned rather than as a write happening on the way out. Assign on its own line and " +
		"return the variable, or wrap the assignment in its own parentheses to say the write was " +
		"meant.",
}

var messageArrowAssignment = rule.Message{
	Id: "arrowAssignment",
	Description: "This arrow function's body is an assignment, so calling it writes to something " +
		"outside itself and returns what it wrote. A concise arrow body reads as a value, which " +
		"is what makes the write easy to miss at the call site. Give the arrow a block body and " +
		"assign inside it, or wrap the assignment in its own parentheses to say the write was " +
		"meant.",
}

// NoReturnAssignMode is the option value, spelled as the config spells it.
//
// A named string type rather than a struct because this rule's option is a bare string in the
// config tuple: `["error", "always"]` puts `"always"` where other rules put an object. This mirrors
// NoCondAssignMode, which shares the option surface and the same two spellings.
type NoReturnAssignMode = string

const (
	// NoReturnAssignExceptParens permits an assignment wrapped in its own parentheses, which is the
	// established way to write "the assignment here is deliberate". This is the default.
	NoReturnAssignExceptParens NoReturnAssignMode = "except-parens"

	// NoReturnAssignAlways reports every assignment returned from a function, parenthesized or not.
	NoReturnAssignAlways NoReturnAssignMode = "always"
)

// NoReturnAssignOptions is the decoded option.
type NoReturnAssignOptions = NoReturnAssignMode

// NoReturnAssign flags an assignment in a return position.
//
//	valid:   function x() { var result = a * b; return result; }
//	valid:   function x() { return (result = a * b); }          (except-parens, the default)
//	valid:   function x() { return function y() { result = a * b }; }
//	valid:   const foo = (a,b,c) => ((a = b), c)                (except-parens)
//	invalid: function x() { return result = a * b; }
//	invalid: () => result = a * b
//	invalid: function x() { return (result = a * b); }          (always)
//
// A `return` that assigns is doing two things where the syntax advertises one. The reader sees a
// value leaving the function and does not see the write that happened on the way out, which is the
// same reason `no-cond-assign` exists one position over. Most of the time it is `==` typed as `=`.
//
// # Two messages, and what separates them
//
// A concise arrow whose whole body is the assignment reports `arrowAssignment`; everything else
// reporting reports `returnAssignment`. That is not cosmetic, because the two point at different
// nodes: the return statement in one case and the entire arrow function in the other. A rule with
// one message would pass every id fixture while pointing at the wrong span for half its findings.
//
// # The option is about parentheses, and the default is the permissive one
//
// Under `except-parens`, the default, an assignment wrapped in its own parentheses is how an author
// says the write was meant, and the rule stays silent. Under `always`, parentheses stop meaning
// anything. Same surface as `no-cond-assign`, same two spellings, same permissive default.
//
// # The parenthesis test is positional and blind to nesting, which is a real divergence source
//
// Upstream asks `astUtils.isParenthesised`, which reads the token immediately BEFORE the assignment
// and the token immediately AFTER it and answers true when they are `(` and `)`. It does not ask
// whether those two parentheses are each other's match, and it does not ask whether they belong to
// the assignment at all. So `return foo(a = b)` is SILENT under the default, because the call's own
// open paren sits immediately before the assignment and its close paren immediately after, and the
// rule reads that as the author's deliberate wrapping. Measured against the installed build.
//
// A port using our `ParenthesizedExpression` node instead would report that input, because the
// parser gives a call argument no parenthesized wrapper. It would also disagree the other way on
// `return (a = b) + 1`, where upstream is silent and there is a real wrapper only around the left
// operand of an addition that is itself unwrapped. Both directions are wrong, so this reproduces
// the positional test rather than the structural one, and `isPositionallyParenthesized` carries the
// measurements at the line.
//
// # The ancestor walk, and what stops it
//
// From the assignment, upstream climbs until it reaches a node whose ESTree type matches
// `/^(?:[a-zA-Z]+?Statement|ArrowFunctionExpression|FunctionExpression|ClassExpression)$/`, then
// asks what it landed on. A `ReturnStatement` reports; an `ArrowFunctionExpression` reports only
// when the child it came up through IS the arrow's body, which is what separates `() => a = b` from
// `(a = b) => c`, where the assignment is a parameter default. Anything else is silent.
//
// The sentinel set is what makes `return class extends (a = b) {}` and `switch (a = b)` silent:
// `ClassExpression` and `SwitchStatement` both stop the walk before any return is reached. It is
// also why a function declaration's body does not leak, even though `FunctionDeclaration` itself is
// not a sentinel: an assignment written in a body is an `ExpressionStatement`, which is.
var NoReturnAssign = rule.Rule{
	Name: "no-return-assign",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Unrecognized and absent both land on the permissive mode. A rule enabled with a bare
		// severity must not silently run the strict half, and a typo must not silently escalate.
		// Same reasoning, and the same direction, as no-cond-assign.
		always := false
		if mode, configured := rule.OptionsAs[NoReturnAssignOptions](options); configured && mode == NoReturnAssignAlways {
			always = true
		}

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !isAssignmentExpression(node) {
					return
				}
				if !always && isPositionallyParenthesized(ctx.SourceFile, node) {
					return
				}

				child, sentinel := climbToTheEnclosingSentinel(node)
				if sentinel == nil {
					return
				}

				switch sentinel.Kind {
				case ast.KindReturnStatement:
					ctx.ReportNode(sentinel, messageReturnAssignment)
				case ast.KindArrowFunction:
					// Only when the walk came up through the arrow's BODY. An assignment in a
					// parameter default reaches the same arrow through a different child, and
					// `(a = b) => c` is a default rather than a returned write.
					if sentinel.AsArrowFunction().Body == child {
						ctx.ReportNode(sentinel, messageArrowAssignment)
					}
				}
			},
		}
	},
}

// climbToTheEnclosingSentinel walks up from an assignment to the first node that ends the search,
// returning both that node and the child it was reached through.
//
// The child is returned because the arrow case needs it: an arrow reports only when the walk
// arrived through its body, and by the time the loop has the arrow it has lost which of the arrow's
// children it came from unless it kept it.
//
// Reaching the top of the file returns nil rather than the SourceFile. Upstream's loop ends when
// `parent` is nullish, and `Program` is not in its sentinel set, so a walk that runs out of
// ancestors falls through both report branches. Returning nil here reproduces that: a SourceFile is
// neither a return nor an arrow, so returning it would behave identically, but nil says the walk
// found nothing rather than leaving the caller to notice the kind does not match.
func climbToTheEnclosingSentinel(assignment *ast.Node) (child *ast.Node, sentinel *ast.Node) {
	child = assignment
	for ancestor := assignment.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if isReturnAssignSentinel(ancestor) {
			return child, ancestor
		}
		child = ancestor
	}
	return child, nil
}

// isReturnAssignSentinel reports whether a node ends the ancestor walk.
//
// This is upstream's `/^(?:[a-zA-Z]+?Statement|ArrowFunctionExpression|FunctionExpression|ClassExpression)$/`
// written out. The regex is a name test rather than a semantic one, so the set has to be enumerated
// from the ESTree type names it matches rather than derived from what a statement is. Two members
// of that set are surprising enough to be worth naming:
//
//   - `FunctionDeclaration` is NOT a sentinel, because the name ends in `Declaration`. It does not
//     have to be: an assignment written inside its body is an `ExpressionStatement`, which is, so
//     the walk stops well below the declaration and no assignment escapes a function body.
//   - `SwitchCase` and `CatchClause` are NOT sentinels, for the same naming reason, and unlike the
//     function case there is no block below them in the `case` position. That is why
//     `switch (a = b) {}` is silent through `SwitchStatement` rather than through anything nearer.
//
// `ast.IsStatement` is deliberately not used, and the difference runs in both directions. It counts
// declaration statements, which upstream's regex excludes, and it refuses a `Block` that is a
// function body or a try/catch block, which upstream's regex includes. Either difference alone
// would change which inputs report.
func isReturnAssignSentinel(node *ast.Node) bool {
	switch node.Kind {
	// The `[a-zA-Z]+?Statement` half of the regex, as typescript-go spells the same constructs.
	//
	// # Two places where the transcription is not one for one, and both are load-bearing
	//
	// `KindVariableStatement` has no counterpart in upstream's set: ESTree calls the node
	// `VariableDeclaration`, whose name ends in `Declaration`, so the regex does not match it. It is
	// here anyway because our parser wraps declarations in a statement node and because it is the
	// only thing that stops one real shape. `const f = () => { var z = a = b; };` puts an assignment
	// in a declarator initializer, which is not an expression statement, so nothing below the
	// declaration stops the walk; without this case it climbs to the arrow and reports
	// `arrowAssignment` on code upstream leaves alone. Measured: dropping this case flips that input
	// and `function f(){ return () => { var z = a = b; }; }` from silent to reporting, and both are
	// silent against the installed build.
	//
	// `KindBlock` is deliberately ABSENT, and it is the one member of upstream's set that is not
	// here. Upstream's regex matches `BlockStatement`, and it is what stops the declarator shape
	// above in ESTree. In our tree it is SUBSUMED everywhere it could apply: a walk starting at an
	// assignment cannot reach a Block without passing a nearer sentinel, because a bare assignment
	// inside a block is wrapped in an ExpressionStatement and a declared one in a VariableStatement.
	// So the two differences cancel, and the position upstream covers with Block is covered here by
	// VariableStatement.
	//
	// Measured rather than argued. A probe walking from every assignment to the first sentinel, with
	// a Block case that reports instead of stopping, reached a Block ZERO times across 21 shapes:
	// bare blocks, function and generator and async bodies, if, try, catch, finally, switch clauses,
	// every loop form, a labeled block, a class static block, a namespace, and three malformed
	// inputs whose error recovery synthesizes nodes well-formed source never produces. The same
	// probe with ExpressionStatement removed from the stop set reached a Block on 15 of those 21,
	// which is the control that says the probe could see.
	//
	// A mutation dropping Block from this set therefore survives every fixture, correctly: no input
	// can distinguish the two versions. Listing it anyway would read as a handled case and the next
	// reader would have no way to tell it was inert. That verdict enumerates the callers above it:
	// if a listener anchored on something other than an assignment ever calls this, it is void.
	//
	// # Five members are inert, and they are kept anyway
	//
	// A full sweep over this set, dropping one member at a time, found five that no input can
	// distinguish: Break, Continue, Debugger, Empty, and Labeled. The first four hold no expression
	// at all, so no assignment can sit under one; a labeled statement holds only a statement, which
	// stops the walk first. A probe over fourteen shapes, including nested labels, a labeled
	// function declaration, and two malformed inputs, found none of the five ever reached as the
	// first sentinel, and the control with ExpressionStatement removed reached Labeled immediately.
	//
	// They stay because this set is a TRANSCRIPTION of an enumerated upstream regex rather than a
	// set of independent decisions, and the next reader's job is to check it against that regex. A
	// set missing five members it should have looks like a porting gap, and proving each absence
	// inert costs the same probe again. Block is deleted rather than kept for the opposite reason:
	// it is inert AND its role is taken by a different node, so listing it would misdescribe how
	// this works.
	//
	// The three that ARE load-bearing and least obvious, each pinned by a fixture: With, through its
	// subject expression; Try, through a catch binding's destructuring default, which is the only
	// position under a `try` that is not already a statement; and VariableStatement, above.
	case ast.KindBreakStatement,
		ast.KindContinueStatement,
		ast.KindDebuggerStatement,
		ast.KindDoStatement,
		ast.KindEmptyStatement,
		ast.KindExpressionStatement,
		ast.KindForInStatement,
		ast.KindForOfStatement,
		ast.KindForStatement,
		ast.KindIfStatement,
		ast.KindLabeledStatement,
		ast.KindReturnStatement,
		ast.KindSwitchStatement,
		ast.KindThrowStatement,
		ast.KindTryStatement,
		ast.KindVariableStatement,
		ast.KindWhileStatement,
		ast.KindWithStatement:
		return true

	// The three expression forms the regex names by hand. `FunctionExpression` and `ClassExpression`
	// are `KindFunctionExpression` and `KindClassExpression`; `ArrowFunctionExpression` is
	// `KindArrowFunction`, and it is the only one of the three that can also report.
	case ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindClassExpression:
		return true
	}
	return false
}

// isPositionallyParenthesized reproduces ESLint's `astUtils.isParenthesised`, which is a test on the
// two tokens adjacent to a node rather than on the tree.
//
// Upstream reads `sourceCode.getTokenBefore(node)` and `sourceCode.getTokenAfter(node)` and answers
// true when their values are `(` and `)`. Nothing checks that the two parentheses match each other
// or that either belongs to the node, so the predicate is positional and, on several real shapes,
// answers a question different from "is this expression wrapped".
//
// Measured against the installed build (eslint 10.8.1, whose rule file is byte-identical to the
// clone), under the default mode where silence means the predicate answered true:
//
//	return (a = b);              silent    a real wrapper
//	return foo(a = b);           silent    the CALL's parens, not the assignment's
//	return new Foo(a = b);       silent    same, through `new`
//	return a[b = c];             REPORTS   brackets are not parentheses
//	return [a = b];              REPORTS   same
//	return {k: a = b};           REPORTS   same
//	return (a, b = c);           REPORTS   the token before is `,`
//	return (a) = b;              REPORTS   the token after is `;`
//	return (a = b) + 1;          silent    a real wrapper, on a node that is not the whole return
//	return (a = b).c;            silent    the token after is `)`, the member access is invisible
//
// So a structural test over our `ParenthesizedExpression` node disagrees with upstream on at least
// the two call rows and the `.c` row, in both directions, and no case in upstream's own corpus
// separates them. It writes no call-argument assignment at all.
//
// # Finding the two adjacent tokens without a token stream
//
// The previous token needs no scan at all. A node's `Pos()` is the position BEFORE its leading
// trivia, which is exactly where the previous token ended, so `text[node.Pos()-1]` is that token's
// last byte. That is the same property `rule.TokenRange` exists to work around, used in the other
// direction. Measured on the trivia shapes: `return foo(/*c*/ a = b /*d*/)` and
// `return ( /*c*/ a = b )` both answer `(` here, while `TokenRange(...).Pos()-1` answers the space
// inside the comment's trailing gap and gets both rows wrong.
//
// The next token does need a forward scan, since `End()` sits at the last byte of the node itself
// with the trailing trivia after it. `scanner.SkipTrivia` is what the scanner uses to find a token
// start, so it lands where `getTokenAfter` would.
//
// # Why comparing one byte is faithful, and not a shortcut
//
// Upstream compares a token's whole value; this compares one byte. They agree because `(` and `)`
// are single-byte punctuators that no other token can start or end with: the grammar has no longer
// token beginning with `(` or ending with `)`, so a byte found at a token boundary IS that token.
//
// A `(` inside a string, a template, or a regular expression is never at that boundary, because
// something always separates it from the assignment. Both `return "(" , a = b;` and a template
// literal holding an open paren followed by `, a = b;` answer `,` here and report, matching
// upstream. A `(` inside a COMMENT is skipped, which is also what upstream does, since its token
// reader skips comments: `return foo /*(*/ (a = b);` is silent both ways because the real `(`
// follows the comment.
func isPositionallyParenthesized(sourceFile *ast.SourceFile, node *ast.Node) bool {
	text := sourceFile.Text()

	// `node.Pos()` is before the leading trivia, so the byte at Pos()-1 closes the previous token.
	// A node starting the file has no previous token, which upstream reads as nullish and false.
	position := node.Pos()
	if position <= 0 || position > len(text) || text[position-1] != '(' {
		return false
	}

	after := scanner.SkipTrivia(text, node.End())
	return after < len(text) && text[after] == ')'
}
