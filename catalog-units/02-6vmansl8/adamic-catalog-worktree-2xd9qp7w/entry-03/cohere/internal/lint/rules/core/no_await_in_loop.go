package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedAwaitInLoop = rule.Message{
	Id: "unexpectedAwait",
	Description: "This waits inside a loop, so each iteration blocks on the one before it and the " +
		"work runs one at a time. Where the iterations do not depend on each other, collecting " +
		"the promises and awaiting them together turns a sum of latencies into a maximum. Where " +
		"they do depend on each other, the sequencing is deliberate and the loop is the right " +
		"shape, which is why this reports rather than repairs.",
}

// NoAwaitInLoop flags an await that runs once per iteration of an enclosing loop.
//
//	valid:   async function foo() { await bar; }
//	valid:   async function foo() { for (var bar of await baz) { } }
//	valid:   async function foo() { for await (var x of xs) { await f(x) } }
//	valid:   async function foo() { while (true) { var y = async () => await foo; } }
//	valid:   while (true) { using resource = getResource(); }
//	invalid: async function foo() { while (baz) { await bar; } }
//	invalid: async function foo() { for (var i; i < n; i = await bar) {  } }
//	invalid: while (true) { await using resource = getResource(); }
//
// # Three things can wait, not one
//
// Upstream hooks three node types and the rule is wrong if any is missing. An `await` expression is
// the obvious one. A `for await...of` loop waits on every iteration of its own iterator, so it
// reports when it is itself nested in an ordinary loop. And an `await using` declaration waits when
// its scope ends, so it reports in a loop body exactly as an `await` expression would, while a plain
// `using` does not.
//
// The last two are why the corpus carries four clean cases that look like they should report:
// `for await (var x of xs) { await f(x) }`, `while (true) { using resource = getResource(); }`,
// `await using resource = getResource();` outside any loop, and
// `for (await using resource = getResource(); ;) {}`.
//
// # The walk is upward from the waiting node, and where it stops is the rule
//
// From the node that waits, climb through parents until either a loop is found with this node in a
// position the loop RE-ENTERS, or a boundary is reached. The boundaries are the three function
// forms and a `for await...of` loop, the latter because asynchronous iteration is the deliberate
// case rather than the mistake. That is why `while (true) { var y = async () => await foo; }` is
// clean: the arrow function stops the climb before the while is seen.
//
// A class method body is a function expression in ESTree and a method declaration here, so
// `while (true) { class Foo { async foo() { await bar; } } }` is clean upstream. Our boundary test
// therefore names the method and accessor forms as well, which are the shapes the same source
// parses into.
//
// # Which POSITION inside a loop counts, which is the half that is easy to get wrong
//
// Being inside a loop node is not enough, because parts of a loop header run once. Upstream's
// `isLooped` names the re-entered positions explicitly and every one of these was measured against
// the installed build at 10.8.1 rather than read off the source:
//
//	for            test, update, body       reports; the INITIALIZER does not
//	for...of       body                     reports; the ITERABLE does not
//	for...in       body                     reports; the OBJECT does not
//	while          test, body               reports
//	do...while     test, body               reports
//
// So `for (var i = await bar; i < n; i++) {}` and `for (var bar of await baz) {}` are both clean,
// and both are in upstream's corpus for exactly this reason. A port testing only "is there a loop
// above me" reports all three and passes nothing that would tell it.
//
// The one exception is an `await using` in a for-of or for-in LEFT position, which upstream adds to
// `isLooped`: `for (await using resource of resources) {}` reports, because the resource is disposed
// once per iteration. `for (await using resource = getResource(); ;) {}` is the for-loop initializer
// and stays clean.
//
// # Where the finding points
//
// Upstream reports the waiting node itself. For an `await using` that node is ESTree's
// `VariableDeclaration`, whose span INCLUDES a trailing semicolon when one is written and excludes
// it in a for-of header where none exists. Measured on all four shapes: our `VariableStatement`
// reproduces the first and our `VariableDeclarationList` reproduces the second, so the report picks
// between them by which parent the list has rather than always taking one.
//
// # No fixer, matching upstream
//
// Upstream declares none, and the repair is a design decision: whether the iterations are actually
// independent is not visible from the syntax, and rewriting a dependent loop into
// `Promise.all` changes what the program does.
var NoAwaitInLoop = rule.Rule{
	Name: "no-await-in-loop",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(waitingNode *ast.Node, reportNode *ast.Node) {
			if !awaitInLoopIsInsideALoop(waitingNode) {
				return
			}
			ctx.ReportNode(reportNode, messageUnexpectedAwaitInLoop)
		}

		return rule.Listeners{
			ast.KindAwaitExpression: func(node *ast.Node) {
				report(node, node)
			},

			// A `for await...of` waits on its own iterator every iteration, so it reports when it
			// is nested inside an ordinary loop. The climb starts at the loop itself, and the
			// boundary test below stops at a `for await...of` PARENT rather than at this node, so
			// starting here does not immediately stop.
			ast.KindForOfStatement: func(node *ast.Node) {
				if node.AsForInOrOfStatement().AwaitModifier == nil {
					return
				}
				report(node, node)
			},

			// `await using` waits at scope exit. The flag lives on the declaration LIST here rather
			// than on the declaration, which is where ESTree carries `kind`, so this anchors on the
			// list and reports whichever node reproduces upstream's span.
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				// Measured against every declaration spelling rather than trusted by name: this
				// answers true for `await using` in all three positions it can be written and false
				// for `using`, `const`, `let` and `var`, which are four of upstream's clean cases.
				if !ast.IsVarAwaitUsing(node) {
					return
				}
				report(node, awaitUsingReportNode(node))
			},
		}
	},
}

// awaitUsingReportNode picks the node whose span matches upstream's.
//
// Upstream reports ESTree's `VariableDeclaration`, and its span includes a trailing semicolon when
// the declaration is a statement and excludes it when the declaration sits in a for-of header. Those
// are two different nodes here, so the choice is made by which parent the list has. Measured against
// the installed build on all three shapes upstream's corpus writes, and on the same source with the
// semicolon removed, which is what showed the semicolon was the difference rather than a constant
// offset.
func awaitUsingReportNode(declarationList *ast.Node) *ast.Node {
	if declarationList.Parent != nil && declarationList.Parent.Kind == ast.KindVariableStatement {
		return declarationList.Parent
	}
	return declarationList
}

// awaitInLoopIsInsideALoop climbs from a waiting node looking for a loop that re-enters it.
//
// Upstream's loop, transcribed. The climb carries both the current node and its parent, because the
// question at each step is not "is this inside a loop" but "which FIELD of this loop holds the
// branch I came up through", and only the child answers that.
func awaitInLoopIsInsideALoop(waitingNode *ast.Node) bool {
	node := waitingNode
	parent := node.Parent

	for parent != nil && !awaitInLoopIsBoundary(parent) {
		if awaitInLoopIsReEntered(node, parent) {
			return true
		}
		node = parent
		parent = parent.Parent
	}
	return false
}

// awaitInLoopIsBoundary answers whether the climb stops here.
//
// The three function forms, because a function written inside a loop is not called by being
// written, and a `for await...of`, because asynchronous iteration is the deliberate case.
//
// Upstream names only `FunctionDeclaration`, `FunctionExpression` and `ArrowFunctionExpression`,
// and that is not a shorter list than this one: ESTree represents a class method's body and an
// object method's body as a `FunctionExpression`, while this parser gives them their own kinds. The
// clean case `while (true) { class Foo { async foo() { await bar; } } }` is upstream's own, and it
// is silent there through `FunctionExpression`. Omitting the method kinds here would report it.
func awaitInLoopIsBoundary(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindConstructor,
		ast.KindClassStaticBlockDeclaration:
		return true
	case ast.KindForOfStatement:
		return node.AsForInOrOfStatement().AwaitModifier != nil
	}
	return false
}

// awaitInLoopIsReEntered answers whether `child` sits in a part of `parent` that runs once per
// iteration.
//
// This is upstream's `isLooped` and it is the discriminating half of the rule. A loop header has
// parts that run once, and a port that skips this reports three of upstream's own clean cases: the
// for-loop initializer, the for-of iterable and the for-in object.
func awaitInLoopIsReEntered(child *ast.Node, parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindForStatement:
		statement := parent.AsForStatement()
		// The initializer is deliberately absent: it runs once, which is why
		// `for (var i = await bar; i < n; i++) {}` is clean.
		return child == statement.Condition ||
			child == statement.Incrementor ||
			child == statement.Statement

	case ast.KindForOfStatement, ast.KindForInStatement:
		statement := parent.AsForInOrOfStatement()
		if child == statement.Statement {
			return true
		}
		// Upstream's one addition beyond the body: an `await using` in the LEFT position is
		// disposed once per iteration, so `for (await using resource of resources) {}` reports
		// while `for (var bar of await baz) {}` stays clean because its await is on the RIGHT.
		return child == statement.Initializer && ast.IsVarAwaitUsing(child)

	case ast.KindWhileStatement:
		statement := parent.AsWhileStatement()
		return child == statement.Expression || child == statement.Statement

	case ast.KindDoStatement:
		statement := parent.AsDoStatement()
		return child == statement.Expression || child == statement.Statement
	}
	return false
}
