package core

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	type_checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// RequireAwait flags an `async` function that never awaits anything.
//
//	valid:   async function f() { await g(); }
//	valid:   async function f() { for await (const x of y) {} }
//	valid:   async function f() {}                 // an empty body is a stub, not a mistake
//	valid:   async function* f() { doSomething(); } // a generator is exempt entirely
//	invalid: async function f() { doSomething(); }
//
// # What the marker costs when it is not doing anything
//
// `async` changes a function's contract: it wraps the return value in a promise and makes every
// caller either await it or handle a floating promise. A function that never awaits pays that cost
// and buys nothing, and a reader seeing `async` reasonably assumes there is asynchrony inside to
// find. Removing the keyword is offered as a SUGGESTION rather than applied, because it changes the
// function's return type from `Promise<T>` to `T` and every caller has to agree.
//
// # Three things count as awaiting, and missing any one reports a legitimate function
//
//	await x                  the obvious one
//	for await (const x of y) an await hidden in a loop header, which a search for
//	                         KindAwaitExpression alone never sees
//	await using x = y()      an await hidden in a declaration's flags rather than in any
//	                         expression node at all
//
// That last one is the sharpest and it carries a trap worth stating: `NodeFlagsAwaitUsing` is
// `Const|Using` rather than a bit of its own, so `flags&NodeFlagsAwaitUsing != 0` is TRUE for an
// ordinary `const`. Measured: a plain `const x` in a `for...of` header answers true to the naive
// test. The correct discrimination is the mask idiom typescript-go itself uses at
// `ast/utilities.go:1214`, comparing the block-scoped bits for equality.
//
// # Two exclusions, both deliberate upstream
//
// An EMPTY body is never reported. Upstream declines it on the reasoning that an empty async
// function is a stub or an interface implementation rather than an oversight, and reporting it
// would fire on every unimplemented method in a codebase.
//
// A GENERATOR is never reported, whatever its body. An async generator's `async` governs its
// iteration protocol rather than its return value, so the keyword is doing work even with no
// `await` inside.
//
// # Nested functions do not count for their parent
//
// An `await` inside an inner function belongs to that function. Upstream keeps a stack and pushes a
// fresh frame per function, so the inner await sets the inner frame's flag and the outer one stays
// unset. Here the search stops at every nested function boundary, which is the same rule stated as
// a walk. Measured on upstream's own case: `async function foo() { async () => { await x; } }`
// reports the OUTER function, and the inner arrow is separately clean because it does await.
var RequireAwait = rule.Rule{
	Name: "require-await",
	// The contract check reads the checker. Without this the checker is nil, the guard silently
	// never fires, and the rule quietly reverts to upstream's blindness.
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		check := func(node *ast.Node) {
			checkRequireAwait(ctx, node)
		}
		return rule.Listeners{
			ast.KindFunctionDeclaration: check,
			ast.KindFunctionExpression:  check,
			ast.KindArrowFunction:       check,
			// A method is a function expression under a MethodDefinition upstream, so its
			// listeners cover it through `FunctionExpression`. Our parser folds the two into one
			// node, so the method kinds are listened for directly. Accessors are included because
			// an accessor can carry `async` here and upstream reaches them the same way.
			ast.KindMethodDeclaration: check,
			ast.KindGetAccessor:       check,
			ast.KindSetAccessor:       check,
			ast.KindConstructor:       check,
		}
	},
}

// checkRequireAwait judges one function-like node.
func checkRequireAwait(ctx rule.Context, node *ast.Node) {
	if node.ModifierFlags()&ast.ModifierFlagsAsync == 0 {
		return
	}
	if requireAwaitIsGenerator(node) {
		return
	}
	body := node.Body()
	if body == nil {
		// An overload signature or a declaration with no body. Nothing to search and nothing to
		// report; upstream never sees these because ESTree gives them no function node.
		return
	}
	if requireAwaitBodyIsEmpty(body) {
		return
	}
	if requireAwaitContainsAwait(body) {
		return
	}
	if requireAwaitSatisfiesPromiseContract(ctx, node) {
		return
	}

	head := requireAwaitHeadRange(ctx, node)
	asyncRange, found := requireAwaitAsyncKeywordRange(ctx, node)
	if !found {
		// The `async` keyword could not be located, which should not happen for a node whose
		// modifier flag is set. Reported without a suggestion rather than skipped: the finding is
		// still true, and offering a repair whose span was not found would be worse.
		ctx.ReportRange(head, requireAwaitMessage(ctx, node))
		return
	}

	ctx.ReportRangeWithSuggestions(head, requireAwaitMessage(ctx, node), rule.Suggestion{
		Message: rule.Message{
			Id: "removeAsync",
			Description: "Remove `async`. This changes the function's return type from a " +
				"promise to the value itself, so every caller has to stop awaiting it, which " +
				"is why it is offered rather than applied.",
		},
		Fixes: []rule.Fix{rule.ReplaceRange(asyncRange, requireAwaitReplacement(ctx, node))},
	})
}

// requireAwaitMessage renders the finding, naming the function the way upstream names it.
func requireAwaitMessage(ctx rule.Context, node *ast.Node) rule.Message {
	return rule.Message{
		Id: "missingAwait",
		Description: fmt.Sprintf(
			"%s has no `await` expression. The `async` keyword wraps the return value in a "+
				"promise and obliges every caller to await it or handle a floating promise, so "+
				"a function that never awaits pays that cost for nothing and tells the reader "+
				"there is asynchrony here to find. Remove `async`, or await what this was "+
				"waiting for.",
			requireAwaitDescribe(ctx, node)),
	}
}

// requireAwaitIsGenerator answers whether a function-like node is a generator.
//
// The asterisk is a field on each concrete type rather than a method on `Node`, so the kinds are
// enumerated. An arrow function cannot be a generator at all.
func requireAwaitIsGenerator(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

// requireAwaitBodyIsEmpty answers whether a function body contains no statements.
//
// Upstream's `isEmptyFunction`, which tests for a block with an empty body. A concise arrow body is
// an expression rather than a block and is never empty, so it always answers false. A block holding
// only an empty statement is NOT empty -- measured, `async function f() { ; }` reports.
func requireAwaitBodyIsEmpty(body *ast.Node) bool {
	if body.Kind != ast.KindBlock {
		return false
	}
	statements := body.AsBlock().Statements
	return statements == nil || len(statements.Nodes) == 0
}

// requireAwaitContainsAwait searches a function body for anything that awaits.
//
// The search stops at every nested function boundary, which is what makes an inner function's await
// belong to the inner function. Upstream reaches the same place with a stack of frames pushed per
// function; this is the same rule stated as a walk, and it is the piece a naive descendant search
// gets backwards.
func requireAwaitContainsAwait(node *ast.Node) bool {
	found := false
	var walk func(current *ast.Node)
	walk = func(current *ast.Node) {
		if current == nil || found {
			return
		}
		switch current.Kind {
		case ast.KindAwaitExpression:
			found = true
			return
		case ast.KindForOfStatement:
			if current.AsForInOrOfStatement().AwaitModifier != nil {
				found = true
				return
			}
		case ast.KindVariableDeclarationList:
			if requireAwaitIsAwaitUsing(current) {
				found = true
				return
			}
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindConstructor, ast.KindClassDeclaration, ast.KindClassExpression:
			// A nested function owns its own awaits. Classes are included because their members
			// are functions of their own, and a class body cannot contain a bare await belonging
			// to the enclosing function.
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	// The body itself is walked rather than passed to the switch, so a body that IS one of the
	// stopping kinds -- which it never is, being a block or an expression -- could not end the
	// search before it starts.
	node.ForEachChild(func(child *ast.Node) bool {
		walk(child)
		return found
	})
	if !found && node.Kind != ast.KindBlock {
		// A concise arrow body is a single expression rather than a container, so it has to be
		// examined itself and not only its children: `async () => await x` awaits at the top.
		walk(node)
	}
	return found
}

// requireAwaitIsAwaitUsing answers whether a declaration list is `await using`.
//
// # The flag is a composite and the obvious test is wrong
//
// `NodeFlagsAwaitUsing` is defined as `NodeFlagsConst | NodeFlagsUsing`, not as a bit of its own,
// and typescript-go's own definition carries the note that on a single node those flags would
// otherwise be mutually exclusive. So `flags&NodeFlagsAwaitUsing != 0` is true for ANY const and
// for any using, which measured true on a plain `const` in a for-of header.
//
// The correct test compares the block-scoped bits for equality, which is the idiom typescript-go
// uses at `ast/utilities.go:1214`. Measured across all five declaration kinds, with the raw flag
// value recorded so the next reader can check this without re-probing:
//
//	await using   flags=8198   naive=true    correct=true
//	using         flags=8196   naive=true    correct=false
//	const         flags=8194   naive=true    correct=false
//	let           flags=8193   naive=false   correct=false
//	var           flags=8192   naive=false   correct=false
//
// So the naive test is wrong on three of the five rather than only on const, and the two it gets
// right are the two nobody would have thought to check.
func requireAwaitIsAwaitUsing(list *ast.Node) bool {
	return list.Flags&ast.NodeFlagsBlockScoped == ast.NodeFlagsAwaitUsing
}

// requireAwaitHeadRange returns the span upstream reports on.
//
// Upstream calls `getFunctionHeadLoc`, which is three branches in a specific order, and the order
// is the part worth stating because the branches overlap:
//
//  1. If the function's PARENT is a property, a method definition or a property definition, the
//     span starts at the parent and runs to the opening parenthesis of the parameters. This branch
//     is tested first, so it wins over the arrow branch below.
//  2. Otherwise, if the function is an arrow, the span is the arrow token alone.
//  3. Otherwise the span starts at the function and runs to the opening parenthesis.
//
// Reproducing only branches two and three puts the start on the `async` keyword whenever a function
// is a property value, which is five of the eleven shapes below and 490 of the 767 findings this
// rule produced on the ahra tree. The line was right in every one of them and only the column moved,
// which is why a differential that compared lines reported perfect agreement.
//
//	const o = { run: async function(a) {} }         "run: async function"
//	const o = { async run(a) {} }                   "async run"
//	const o = { run: async (a) => {} }              "run: async "       <- not the arrow
//	class A { foo = async function(a) {} }          "foo = async function"
//	class A { static foo = async function(a) {} }   "static foo = async function"
//	class A { foo = async (a) => {} }               "foo = async "      <- not the arrow
//	class A { static async foo(a) {} }              "static async foo"
//	async function foo(a) {}                        "async function foo"
//	const f = async function(a) {}                  "async function"
//	const f = async function named(a) {}            "async function named"
//	const f = async (a) => {}                       "=>"                <- the arrow, no parent
//
// Every span above was measured against the installed rule rather than read off the source.
func requireAwaitHeadRange(ctx rule.Context, node *ast.Node) core.TextRange {
	// Branch one, and it is first for the reason above: a property value takes the parent's start
	// even when the function is an arrow. Our parser models a method as one node rather than as a
	// property holding a function, so the method shapes reach this through the node itself and
	// only the value-holding shapes have a parent to climb to.
	if parent := node.Parent; parent != nil && requireAwaitParentOwnsTheHead(parent) {
		start := rule.TokenRange(ctx.SourceFile, parent).Pos()
		return core.NewTextRange(start, requireAwaitParametersStart(ctx, node, start))
	}

	// Branch two.
	if node.Kind == ast.KindArrowFunction {
		if arrow := node.AsArrowFunction().EqualsGreaterThanToken; arrow != nil {
			return rule.TokenRange(ctx.SourceFile, arrow)
		}
	}

	// Branch three. The head runs from the first token of the node that carries the `async` keyword
	// through the end of the name. A method's `async` sits on the method rather than on any inner
	// function, which our parser already gives us since the two are one node.
	start := requireAwaitHeadStart(ctx, node)
	end := start
	if name := node.Name(); name != nil {
		end = name.End()
	} else {
		end = requireAwaitParametersStart(ctx, node, start)
	}
	if end <= start {
		end = start + len("async")
	}
	return core.NewTextRange(start, end)
}

// requireAwaitHeadStart finds where the reported span begins for a function that owns its own head.
//
// Upstream starts at `node.loc.start`, and the subtlety is what that node CONTAINS. ESTree treats a
// class member's modifiers as part of the function's parent and an `export` as a wrapper AROUND the
// declaration, so the two kinds of leading keyword fall on opposite sides of the start:
//
//	class A { static async foo() {} }           "static async foo"
//	class A { private static async foo() {} }   "private static async foo"
//	export async function foo() {}              "async function foo"     <- not `export`
//	export default async function foo() {}      "async function foo"     <- not `export default`
//
// Our parser hangs both kinds off the same node, so taking its start includes the `export` and puts
// the span a keyword too early. This was measured rather than reasoned: a differential against the
// installed rule over the ahra tree disagreed on 154 findings after the parent branch was fixed,
// every one of them an exported function, and every one agreeing on the line.
//
// So only the export-ish modifiers are skipped, and the class member modifiers are deliberately
// kept. Skipping all of them would move `static async foo` to `async foo`, which trades one column
// defect for another and looks equally plausible from the inside.
func requireAwaitHeadStart(ctx rule.Context, node *ast.Node) int {
	start := rule.TokenRange(ctx.SourceFile, node).Pos()
	modifiers := node.Modifiers()
	if modifiers == nil {
		return start
	}
	for _, modifier := range modifiers.Nodes {
		if modifier == nil {
			continue
		}
		switch modifier.Kind {
		case ast.KindExportKeyword, ast.KindDefaultKeyword, ast.KindDeclareKeyword:
			// Skip past it and keep looking, since `export default` is two of them.
			if end := rule.TokenRange(ctx.SourceFile, modifier).End(); end > start {
				start = end
			}
		default:
			// The first modifier that belongs to the head stops the scan, so a `static` after an
			// `export` still starts the span.
			return requireAwaitSkipSpaces(ctx, start)
		}
	}
	return requireAwaitSkipSpaces(ctx, start)
}

// requireAwaitSkipSpaces advances past whitespace, so a skipped keyword does not leave its space.
func requireAwaitSkipSpaces(ctx rule.Context, start int) int {
	text := ctx.SourceFile.Text()
	for start < len(text) {
		switch text[start] {
		case ' ', '\t', '\n', '\r':
			start++
			continue
		}
		break
	}
	return start
}

// requireAwaitParentOwnsTheHead answers whether the reported span starts at the function's parent.
//
// Upstream names five parent types. Two of them, `TSPropertySignature` and `TSMethodSignature`,
// describe declarations with no body, which this rule cannot reach because a function with no body
// has nothing to search for an await. The three that remain are a property in an object literal and
// a property declaration in a class.
//
// Upstream also has no value test and needs none, and neither does this. A function nested deeper
// inside an initializer, `foo = [async function(){}]`, has the ARRAY as its parent rather than the
// property, so the kind test alone already declines it. A guard comparing the initializer against
// the node was written here first and then removed: it survived every mutation because nothing can
// reach it, which is the signature of a condition that only looks like it is doing work. The four
// nested shapes in `TestRequireAwaitReportsOnTheFunctionHead` pin the behaviour either way.
func requireAwaitParentOwnsTheHead(parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindPropertyAssignment, ast.KindPropertyDeclaration:
		return true
	}
	return false
}

// requireAwaitParametersStart finds the opening parenthesis of a function's parameter list.
//
// Upstream reaches it through `getOpeningParenOfParams`. Found by scanning here rather than through
// the parameter list node, whose type carries a position but no `*ast.Node` to hand `TokenRange`.
// A type parameter list can put an earlier `<` in the way but never an earlier `(`, so the first
// one after the start is the right one.
func requireAwaitParametersStart(ctx rule.Context, node *ast.Node, start int) int {
	text := ctx.SourceFile.Text()
	for scan := start; scan < node.End() && scan < len(text); scan++ {
		if text[scan] == '(' {
			return scan
		}
	}
	return start
}

// requireAwaitAsyncKeywordRange returns the span the suggestion replaces.
//
// Upstream's range runs from the `async` token to the start of the NEXT token including comments,
// so the trailing whitespace goes with the keyword while a comment after it survives. Measured:
// `async /* c */ function foo` becomes `/* c */ function foo`.
func requireAwaitAsyncKeywordRange(ctx rule.Context, node *ast.Node) (core.TextRange, bool) {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return core.TextRange{}, false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier == nil || modifier.Kind != ast.KindAsyncKeyword {
			continue
		}
		keyword := rule.TokenRange(ctx.SourceFile, modifier)
		text := ctx.SourceFile.Text()
		end := keyword.End()
		// Consume the whitespace that follows, stopping at the next non-space byte. That byte
		// begins either the next token or a comment, and upstream stops before both.
		for end < len(text) {
			switch text[end] {
			case ' ', '\t', '\n', '\r':
				end++
				continue
			}
			break
		}
		return core.NewTextRange(keyword.Pos(), end), true
	}
	return core.TextRange{}, false
}

// requireAwaitReplacement returns what the suggestion writes in place of `async`.
//
// Almost always the empty string. Upstream substitutes a SEMICOLON instead when removing the
// keyword would let automatic semicolon insertion join this construct to the previous line, which
// its corpus exercises three times:
//
//	class A { a = 0 \n async [b](){} }    the field and the method would merge
//	class A { a = 0 \n async in(){} }     same, with a keyword-named method
//	foo \n async () => {}                 the call and the arrow would merge
//
// The test is whether the token after `async` could continue the preceding expression. Reproduced
// narrowly rather than by porting upstream's full `needsPrecedingSemicolon`, because the shapes
// that reach it here are exactly the two the corpus names: a class member whose predecessor is a
// property with no semicolon, and an expression statement beginning with a parenthesis or a
// bracket.
func requireAwaitReplacement(ctx rule.Context, node *ast.Node) string {
	if requireAwaitNeedsPrecedingSemicolon(ctx, node) {
		return ";"
	}
	return ""
}

// requireAwaitNeedsPrecedingSemicolon answers whether removing `async` would merge two statements.
//
// # Both halves are required, and each was measured
//
// Upstream asks two questions and emits a semicolon only when both answer yes: whether the token
// after `async` could CONTINUE the preceding expression, and whether the preceding construct
// actually needs a semicolon. Testing only the second reports a semicolon on shapes upstream leaves
// alone, which is how this arm was wrong on two of upstream's own cases before it was measured:
//
//	class A { a = 0 \n async [b](){} }    ";"   the bracket continues `0` as an index
//	class A { a = 0 \n async in(){} }     ";"   `in` continues `0` as a relational operator
//	class A { a = 0 \n async m(){} }      ""    a plain identifier cannot continue it
//	class A { a \n async [b](){} }        ""    the property has no initializer to continue
//	class A { foo() {} \n async [bar](){} } ""  a method needs no semicolon after it
//
// So the continuation test is about the token AFTER `async`, and the semicolon test is about what
// comes BEFORE the member. Reproduced narrowly rather than by porting upstream's full
// `needsPrecedingSemicolon`, because the shapes that reach it here are exactly these.
func requireAwaitNeedsPrecedingSemicolon(ctx rule.Context, node *ast.Node) bool {
	if !requireAwaitTokenAfterAsyncContinuesExpression(ctx, node) {
		return false
	}
	previous := requireAwaitPreviousSibling(node)
	if previous == nil {
		return false
	}
	// Only a construct that could still be an expression needs the guard. A method declaration is
	// already terminated by its body, so nothing can continue it.
	if !requireAwaitCanBeContinued(previous) {
		return false
	}
	// A previous sibling that already ends in a semicolon cannot be continued either.
	text := ctx.SourceFile.Text()
	end := previous.End()
	for end > 0 && (text[end-1] == ' ' || text[end-1] == '\t' || text[end-1] == '\n' ||
		text[end-1] == '\r') {
		end--
	}
	return end == 0 || text[end-1] != ';'
}

// requireAwaitCanBeContinued answers whether a preceding member could still be an open expression.
//
// A property declaration WITH an initializer is the only class member that can, because the
// initializer is an expression and the next line can extend it. A property with no initializer, a
// method, an accessor and a constructor are all closed already.
func requireAwaitCanBeContinued(previous *ast.Node) bool {
	switch previous.Kind {
	case ast.KindPropertyDeclaration:
		return previous.AsPropertyDeclaration().Initializer != nil
	case ast.KindExpressionStatement, ast.KindVariableStatement:
		return true
	}
	return false
}

// requireAwaitTokenAfterAsyncContinuesExpression answers whether the token following `async` could
// extend the expression on the line above.
//
// Three shapes can: an opening bracket (an index), an opening parenthesis (a call), and a keyword
// that is also a binary operator, of which `in` and `instanceof` are the ones a method can be named.
// A plain identifier cannot, because two identifiers in a row are not an expression.
func requireAwaitTokenAfterAsyncContinuesExpression(ctx rule.Context, node *ast.Node) bool {
	asyncRange, found := requireAwaitAsyncKeywordRange(ctx, node)
	if !found {
		return false
	}
	text := ctx.SourceFile.Text()
	scan := asyncRange.End()
	for scan < len(text) && (text[scan] == ' ' || text[scan] == '\t' || text[scan] == '\n' ||
		text[scan] == '\r') {
		scan++
	}
	if scan >= len(text) {
		return false
	}
	if text[scan] == '[' || text[scan] == '(' {
		return true
	}
	// A keyword-named method whose name is also a binary operator. `in` and `instanceof` are the
	// two; upstream reaches the same set through its own token classification.
	rest := text[scan:]
	for _, keyword := range []string{"instanceof", "in"} {
		if strings.HasPrefix(rest, keyword) {
			after := scan + len(keyword)
			if after >= len(text) || !requireAwaitIsIdentifierByte(text[after]) {
				return true
			}
		}
	}
	return false
}

// requireAwaitIsIdentifierByte answers whether a byte can continue an identifier.
func requireAwaitIsIdentifierByte(character byte) bool {
	switch {
	case character >= 'a' && character <= 'z':
		return true
	case character >= 'A' && character <= 'Z':
		return true
	case character >= '0' && character <= '9':
		return true
	case character == '_' || character == '$':
		return true
	}
	return false
}

// requireAwaitPreviousSibling returns the statement or member preceding a node, or nil.
//
// Only two containers matter: a class body, where a property with an initializer can run into the
// next member, and a statement list, where an expression statement can continue onto the next line.
//
// The walk climbs through intermediate nodes so a function EXPRESSION reaches the statement that
// holds it: `foo \n async () => {}` has the arrow nested inside an expression statement rather than
// sitting in the statement list itself.
func requireAwaitPreviousSibling(node *ast.Node) *ast.Node {
	subject := node
	for subject != nil {
		parent := subject.Parent
		if parent == nil {
			return nil
		}
		var siblings []*ast.Node
		switch parent.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			siblings = parent.Members()
		case ast.KindBlock:
			siblings = parent.AsBlock().Statements.Nodes
		case ast.KindSourceFile:
			siblings = parent.AsSourceFile().Statements.Nodes
		}
		if siblings != nil {
			for index, sibling := range siblings {
				if sibling == subject {
					if index == 0 {
						return nil
					}
					return siblings[index-1]
				}
			}
			return nil
		}
		subject = parent
	}
	return nil
}

// requireAwaitDescribe renders a function the way upstream's `getFunctionNameWithKind` does, with
// the first letter capitalized.
//
// Nine renderings appear in the corpus and each is a distinct combination:
//
//	Async function 'foo'      a named declaration or named expression
//	Async function            an anonymous function expression
//	Async arrow function      any arrow, named or not
//	Async method 'foo'        a class or object-literal method
//	Async method ''           a method whose name is the empty string literal
//	Async method              a computed method name, which has no readable name
//
// The property name wins over the function's own where both exist: `{ async: async function foo() }`
// renders `Async method 'async'`, because the object property is what a reader searches for.
func requireAwaitDescribe(ctx rule.Context, node *ast.Node) string {
	kind := "function"
	switch node.Kind {
	case ast.KindArrowFunction:
		return "Async arrow function"
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		kind = "method"
	case ast.KindConstructor:
		return "Async method 'constructor'"
	}

	// A function expression that is the value of a property or a method definition is described by
	// the property's name rather than its own.
	nameHolder := node
	if parent := node.Parent; parent != nil {
		switch parent.Kind {
		case ast.KindPropertyAssignment:
			if parent.AsPropertyAssignment().Initializer == node {
				nameHolder = parent
				kind = "method"
			}
		}
	}

	if name := nameHolder.Name(); name != nil {
		if text, ok := requireAwaitNameText(name); ok {
			return fmt.Sprintf("Async %s '%s'", kind, text)
		}
	}
	// Go whitespace: kind is one of this function's own Go literals, never source text.
	return "Async " + strings.TrimSpace(kind)
}

// requireAwaitNameText reads a name node's text when the syntax settles it.
//
// A computed name answers nothing, which is what produces the bare `Async method` rendering. A
// string-literal name answers its cooked value, so a method named with an empty string
// literal renders WITH that empty name rather than as an unnamed method -- the
// empty string is a NAME rather than an absent one, and the two render differently.
func requireAwaitNameText(name *ast.Node) (string, bool) {
	switch name.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return name.Text(), true
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return name.Text(), true
	case ast.KindNumericLiteral:
		return name.Text(), true
	}
	return "", false
}

// requireAwaitSatisfiesPromiseContract answers whether this function's `async` is required by the
// contract it is being written into, rather than being a stray keyword.
//
// # Why the core rule needs this and upstream's does not have it
//
// ESLint judges one function at a time, so it can see that a body holds no `await` but not why the
// keyword is there. Both the core rule and the typed variant therefore report the same shape: a
// stub written into a position whose declared type is `=> Promise<T>`, sitting beside siblings that
// genuinely do I/O. Measured in ahra, that shape is most of what survives an honest sweep: a
// not-yet-wired ads collector in a map typed `=> Promise<...>`, a test double standing in for a
// dependency whose real implementation awaits, a condition function matching its interface.
//
// Upstream already concedes the case in a narrower form. An EMPTY body is exempt precisely because
// reporting it "would fire on every unimplemented method in a codebase" — the same reasoning,
// drawn by body shape rather than by contract. A stub with one `return emptyReport(...)` in it is
// the same animal and falls outside that line.
//
// We hold a resident type graph and a real checker, so we can ask the question upstream cannot:
// what type is this function expected to have where it is written? If that position demands a
// promise, the `async` is load-bearing and removing it would break the assignment. Reporting it
// would be telling the author to make their code wrong.
//
// The check is deliberately narrow. It fires only when a contextual type exists AND every call
// signature it offers returns a thenable. A position that accepts `T | Promise<T>` is NOT exempt,
// because there the keyword really is optional and the finding is real — that union is exactly
// what a widened interface looks like, and widening is the right repair when no implementation
// awaits. Absent contextual type means absent exemption.
func requireAwaitSatisfiesPromiseContract(ctx rule.Context, node *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}

	// A position inside a generic call is judged by what the call DECLARES, not by what it inferred.
	// See requireAwaitDeclaredGenericDemand for why the inferred answer is the function grading its
	// own homework.
	if declared, substitutions, throughGenericCall := requireAwaitDeclaredGenericDemand(ctx, node); throughGenericCall {
		return requireAwaitTypesDemandPromise(ctx, node, declared, substitutions)
	}

	contextualType := checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone)
	if contextualType == nil {
		// A class method has no contextual type of its own; its contract comes from whatever the
		// class implements or extends. Without this, an adapter satisfying an interface whose
		// declared member returns a promise reads as a stray keyword, and the repair the rule
		// suggests would stop the class from satisfying its own interface.
		contextualType = requireAwaitHeritageMemberType(ctx, node)
	}
	if contextualType == nil {
		return false
	}

	return requireAwaitTypesDemandPromise(ctx, node, []*checker.Type{contextualType}, nil)
}

// requireAwaitTypesDemandPromise answers whether every call signature the position's types offer
// returns a thenable.
//
// Every signature the position offers has to demand a promise. If any one of them accepts a plain
// value, the author had a choice and the keyword is not required of them. A type parameter with a
// substitution is read as the types standing in for it, and one without is read through its
// apparent type, which is its constraint: an unconstrained `U` offers no signature and no `then`.
func requireAwaitTypesDemandPromise(ctx rule.Context, node *ast.Node, types []*checker.Type, substitutions map[*checker.Type][]*checker.Type) bool {
	sawSignature := false
	for _, part := range requireAwaitExpandTypes(types, substitutions) {
		for _, signature := range type_checking.GetCallSignatures(ctx.TypeChecker, part) {
			sawSignature = true
			returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)

			// Split the RETURN type, not the function type. A position declared
			// `T | Promise<T>` offers one signature whose return is a union, and the author may
			// legitimately answer with either, so the keyword is optional and the finding stands.
			// Only a return type that is thenable in every branch actually demands a promise.
			for _, returnPart := range requireAwaitExpandTypes([]*checker.Type{returnType}, substitutions) {
				if !type_checking.IsThenableType(ctx.TypeChecker, node, returnPart) {
					return false
				}
			}
		}
	}

	return sawSignature
}

// requireAwaitExpandTypes splits each type into its union parts and replaces a substituted type
// parameter with the types standing in for it.
func requireAwaitExpandTypes(types []*checker.Type, substitutions map[*checker.Type][]*checker.Type) []*checker.Type {
	var expanded []*checker.Type
	for _, t := range types {
		for part := range type_checking.UnionTypePartsSeq(t) {
			if standIns, substituted := substitutions[part]; substituted {
				for _, standIn := range standIns {
					expanded = append(expanded, type_checking.UnionTypeParts(standIn)...)
				}
				continue
			}
			expanded = append(expanded, part)
		}
	}
	return expanded
}

// requireAwaitContextStepKind is one way a contextual type is handed from a container to a child.
type requireAwaitContextStepKind int

const (
	// requireAwaitStepProperty reads a named property: an object literal hands it to its member.
	requireAwaitStepProperty requireAwaitContextStepKind = iota
	// requireAwaitStepElement reads an array or tuple element at an index.
	requireAwaitStepElement
	// requireAwaitStepReturn reads the return type of the container's call signatures: a function
	// hands it to the expression it returns.
	requireAwaitStepReturn
)

// requireAwaitContextStep is one step from a container's contextual type down to a child's.
type requireAwaitContextStep struct {
	kind  requireAwaitContextStepKind
	name  string
	index int
}

// requireAwaitDeclaredGenericDemand finds the type a generic call DECLARES for the position this
// function is written into, when the position's contextual type comes from a generic call.
//
// # Why the inferred contextual type is the wrong question here
//
// A generic call infers its type parameters from its arguments, and this function is one of them.
// After inference, `[1, 2].map(async (n) => n + sync())` resolves to `map<Promise<number>>`, so the
// contextual type of the callback is `(value: number) => Promise<number>` and the position appears
// to demand exactly the promise the callback produced. It demands nothing: drop the keyword and `U`
// becomes `number`, and the call still compiles. Measured on ahra before this existed, the same
// self-inference exempted `onSelected: async function () {...}` inside `React.useMemo(() => [...])`
// at `useMetricsExportMenu.tsx:74` and `:92`, both of which ESLint reported and was right about.
//
// TypeScript's own `ContextFlagsIgnoreNodeInferences` does not close this. It blocks an inference
// only when the inferred source type's own symbol is the blocked node, which catches
// `identity(async () => 1)` and misses `map`, whose `U` is inferred from the callback's RETURN type
// (a `Promise` whose symbol is the global one). It also works by re-resolving the enclosing calls
// with inference blocked, which leaves expression types cached from that blocked run on a checker
// every other rule reads afterwards. This walk only reads.
//
// # What the walk does
//
// From the function up to the nearest call it is an argument of, record each step a contextual
// type takes on the way down: an object literal's property, an array's element, a contextually
// typed function's return. If that call resolved to an instantiation of a generic signature and
// wrote no explicit type arguments, replay the steps over the DECLARED parameter type instead of
// the instantiated one. A type parameter met on the way is read through its constraint, which is
// how `(api) => Promise<T>` and `T extends () => Promise<void>` still demand a promise and `() => U`
// does not, with two stand-ins where the parameter is fixed by something other than this function:
//
//   - another argument whose declared parameter type is exactly the type parameter, as in
//     `pick<T>(first: T, second: T)`, stands in with its instantiated type. That type still
//     carries this function's contribution, so the stand-in errs toward silence, never toward a
//     report the code cannot satisfy;
//   - a call whose declared return type is exactly the type parameter and which sits in a typed
//     position, as in `const items: Item[] = useMemo(() => [...])`, stands in with that position's
//     type. If `Item` declares the member as returning a promise, removing the keyword breaks the
//     assignment, so the demand is real.
//
// A deeper mention of the parameter in another argument is not followed, and an overload chosen
// BECAUSE this function returns a promise is not undone; both are named here rather than claimed.
//
// The third result is false when the position does not come from a generic call, or when the walk
// meets a shape it does not model; the caller then asks for the ordinary contextual type, which is
// what this rule did before the walk existed.
func requireAwaitDeclaredGenericDemand(ctx rule.Context, node *ast.Node) ([]*checker.Type, map[*checker.Type][]*checker.Type, bool) {
	var steps []requireAwaitContextStep
	current := node

	if ast.IsMethodDeclaration(node) {
		// An object literal's method takes its contextual type from the literal, by name. A class
		// method is not in a call and its contract comes from the heritage clause instead.
		object := node.Parent
		if object == nil || !ast.IsObjectLiteralExpression(object) {
			return nil, nil, false
		}
		name := node.Name()
		if name == nil {
			return nil, nil, false
		}
		text, ok := requireAwaitNameText(name)
		if !ok {
			return nil, nil, false
		}
		steps = append(steps, requireAwaitContextStep{kind: requireAwaitStepProperty, name: text})
		current = object
	}

	for {
		parent := current.Parent
		if parent == nil {
			return nil, nil, false
		}

		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			// A parenthesis hands its contextual type straight through.

		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if conditional.WhenTrue != current && conditional.WhenFalse != current {
				return nil, nil, false
			}

		case ast.KindBinaryExpression:
			// `a || b` and `a ?? b` give both operands the expression's contextual type.
			operator := parent.AsBinaryExpression().OperatorToken.Kind
			if operator != ast.KindBarBarToken && operator != ast.KindQuestionQuestionToken {
				return nil, nil, false
			}

		case ast.KindPropertyAssignment:
			if parent.AsPropertyAssignment().Initializer != current {
				return nil, nil, false
			}
			name := parent.Name()
			if name == nil {
				return nil, nil, false
			}
			text, ok := requireAwaitNameText(name)
			if !ok {
				return nil, nil, false
			}
			steps = append(steps, requireAwaitContextStep{kind: requireAwaitStepProperty, name: text})
			parent = parent.Parent

		case ast.KindArrayLiteralExpression:
			index := -1
			for elementIndex, element := range parent.Elements() {
				if element == current {
					index = elementIndex
					break
				}
				// A spread before this element moves its position by an amount the syntax does
				// not state.
				if element.Kind == ast.KindSpreadElement {
					return nil, nil, false
				}
			}
			if index < 0 {
				return nil, nil, false
			}
			steps = append(steps, requireAwaitContextStep{kind: requireAwaitStepElement, index: index})

		case ast.KindReturnStatement:
			function := ast.GetContainingFunction(parent)
			if function == nil || !requireAwaitReturnTakesContextualType(function) {
				return nil, nil, false
			}
			steps = append(steps, requireAwaitContextStep{kind: requireAwaitStepReturn})
			parent = function

		case ast.KindArrowFunction:
			if parent.Body() != current || !requireAwaitReturnTakesContextualType(parent) {
				return nil, nil, false
			}
			steps = append(steps, requireAwaitContextStep{kind: requireAwaitStepReturn})

		case ast.KindCallExpression, ast.KindNewExpression:
			return requireAwaitDemandFromGenericCall(ctx, parent, current, steps)

		default:
			return nil, nil, false
		}

		current = parent
	}
}

// requireAwaitReturnTakesContextualType answers whether a function's returned expression is typed
// by the function's contextual signature, which is what makes a return a step in the walk.
//
// A declared return type fixes the demand without any call's help, and an async or generator
// function hands its returned expression the awaited or yielded type rather than the signature's
// return type. Neither is modelled, so both end the walk.
func requireAwaitReturnTakesContextualType(function *ast.Node) bool {
	if !ast.IsFunctionExpression(function) && !ast.IsArrowFunction(function) {
		return false
	}
	if function.Type() != nil {
		return false
	}
	return ast.GetFunctionFlags(function) == ast.FunctionFlagsNormal
}

// requireAwaitDemandFromGenericCall replays the recorded steps over the parameter type a generic
// call declares at the argument's position.
func requireAwaitDemandFromGenericCall(ctx rule.Context, call *ast.Node, argument *ast.Node, steps []requireAwaitContextStep) ([]*checker.Type, map[*checker.Type][]*checker.Type, bool) {
	// Explicit type arguments are written by the author, so nothing was inferred from this function.
	if len(call.TypeArguments()) != 0 {
		return nil, nil, false
	}

	arguments := call.Arguments()
	argumentIndex := -1
	for index, candidate := range arguments {
		if candidate == argument {
			argumentIndex = index
			break
		}
		if candidate.Kind == ast.KindSpreadElement {
			return nil, nil, false
		}
	}
	if argumentIndex < 0 {
		// The callee, not an argument.
		return nil, nil, false
	}

	resolved := checker.Checker_getResolvedSignature(ctx.TypeChecker, call, nil, checker.CheckModeNormal)
	if resolved == nil {
		return nil, nil, false
	}
	declared := resolved.Target()
	if declared == nil || len(declared.TypeParameters()) == 0 {
		// Not an instantiation of a generic signature, so the contextual type was not inferred
		// from anything and the ordinary question is the right one.
		return nil, nil, false
	}

	parameterType := requireAwaitParameterTypeAt(ctx, declared, argumentIndex)
	if parameterType == nil {
		return nil, nil, false
	}

	substitutions := requireAwaitTypeParameterStandIns(ctx, call, resolved, declared, argumentIndex)

	types := []*checker.Type{parameterType}
	for stepIndex := len(steps) - 1; stepIndex >= 0; stepIndex-- {
		types = requireAwaitApplyContextStep(ctx, requireAwaitExpandTypes(types, substitutions), steps[stepIndex])
	}
	return types, substitutions, true
}

// requireAwaitParameterTypeAt is the type a signature declares for the argument at an index,
// reading a rest parameter's element type.
func requireAwaitParameterTypeAt(ctx rule.Context, signature *checker.Signature, index int) *checker.Type {
	parameters := signature.Parameters()
	if signature.HasRestParameter() && index >= len(parameters)-1 {
		restType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameters[len(parameters)-1])
		return checker.Checker_getIndexTypeOfType(ctx.TypeChecker, restType, checker.Checker_numberType(ctx.TypeChecker))
	}
	if index >= len(parameters) {
		return nil
	}
	return checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameters[index])
}

// requireAwaitTypeParameterStandIns finds, for each of the declared signature's type parameters,
// the types that fix it independently of the argument being judged.
func requireAwaitTypeParameterStandIns(ctx rule.Context, call *ast.Node, resolved *checker.Signature, declared *checker.Signature, argumentIndex int) map[*checker.Type][]*checker.Type {
	substitutions := map[*checker.Type][]*checker.Type{}
	typeParameters := declared.TypeParameters()
	isTypeParameter := func(t *checker.Type) bool {
		for _, typeParameter := range typeParameters {
			if t == typeParameter {
				return true
			}
		}
		return false
	}

	declaredParameters := declared.Parameters()
	resolvedParameters := resolved.Parameters()
	for index := range call.Arguments() {
		if index == argumentIndex || index >= len(declaredParameters) || index >= len(resolvedParameters) {
			continue
		}
		if declared.HasRestParameter() && index >= len(declaredParameters)-1 {
			continue
		}
		declaredType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, declaredParameters[index])
		if !isTypeParameter(declaredType) {
			continue
		}
		resolvedType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, resolvedParameters[index])
		substitutions[declaredType] = append(substitutions[declaredType], resolvedType)
	}

	declaredReturn := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, declared)
	if isTypeParameter(declaredReturn) {
		if positionType := checker.Checker_getContextualType(ctx.TypeChecker, call, checker.ContextFlagsNone); positionType != nil {
			substitutions[declaredReturn] = append(substitutions[declaredReturn], positionType)
		}
	}

	return substitutions
}

// requireAwaitApplyContextStep takes one step down from a container's types to a child's.
//
// Property and signature lookups read a type parameter through its apparent type, so an
// unconstrained one yields nothing here, which is the point: nothing is demanded of the child.
func requireAwaitApplyContextStep(ctx rule.Context, types []*checker.Type, step requireAwaitContextStep) []*checker.Type {
	var next []*checker.Type
	for _, part := range types {
		switch step.kind {
		case requireAwaitStepProperty:
			if property := checker.Checker_getPropertyOfType(ctx.TypeChecker, part, step.name); property != nil {
				next = append(next, checker.Checker_getTypeOfSymbol(ctx.TypeChecker, property))
				continue
			}
			if indexed := checker.Checker_getIndexTypeOfType(ctx.TypeChecker, part, checker.Checker_stringType(ctx.TypeChecker)); indexed != nil {
				next = append(next, indexed)
			}
		case requireAwaitStepElement:
			if checker.Checker_isArrayOrTupleType(ctx.TypeChecker, part) && !checker.Checker_isArrayType(ctx.TypeChecker, part) {
				elements := checker.Checker_getTypeArguments(ctx.TypeChecker, part)
				if step.index < len(elements) {
					next = append(next, elements[step.index])
				}
				continue
			}
			if indexed := checker.Checker_getIndexTypeOfType(ctx.TypeChecker, part, checker.Checker_numberType(ctx.TypeChecker)); indexed != nil {
				next = append(next, indexed)
			}
		case requireAwaitStepReturn:
			for _, signature := range type_checking.GetCallSignatures(ctx.TypeChecker, part) {
				next = append(next, checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature))
			}
		}
	}
	return next
}

// requireAwaitHeritageMemberType finds the type a class's own interface declares for this method.
//
// `class CoinbaseAdapter implements FinanceAdapterInterface` puts the contract on the class rather
// than on the method, so `getContextualType` on the method returns nothing. Walking to the heritage
// clause and asking for the member of the same name recovers it. Measured on ahra: five finance
// adapters implement one interface whose fetch members return promises, and two of the five
// genuinely await, so the promise is earned and the other three have no choice about the keyword.
func requireAwaitHeritageMemberType(ctx rule.Context, node *ast.Node) *checker.Type {
	if !ast.IsMethodDeclaration(node) || ast.IsStatic(node) {
		return nil
	}
	name := node.Name()
	if name == nil || !ast.IsIdentifier(name) {
		return nil
	}
	class := node.Parent
	if class == nil || !ast.IsClassLike(class) {
		return nil
	}
	heritageClauses := type_checking.GetHeritageClauses(class)
	if heritageClauses == nil {
		return nil
	}

	for _, clause := range heritageClauses.Nodes {
		for _, typeNode := range clause.AsHeritageClause().Types.Nodes {
			heritageType := ctx.TypeChecker.GetTypeAtLocation(typeNode)
			if heritageType == nil {
				continue
			}
			member := checker.Checker_getPropertyOfType(ctx.TypeChecker, heritageType, name.Text())
			if member == nil {
				continue
			}
			return ctx.TypeChecker.GetTypeOfSymbolAtLocation(member, node)
		}
	}
	return nil
}
