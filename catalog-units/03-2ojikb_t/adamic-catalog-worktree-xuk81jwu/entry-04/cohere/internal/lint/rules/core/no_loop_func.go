package core

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoLoopFunc = rule.Message{
	Id: "unsafeRefs",
	Description: "This function is created inside a loop and closes over a variable the loop " +
		"reassigns, so every iteration builds a function sharing one binding rather than capturing " +
		"the value it was made with. By the time any of them runs the variable holds whatever the " +
		"last iteration left, which is why the callbacks all report the same index. Declare the " +
		"variable with `let` or `const` inside the loop body so each iteration gets its own, or " +
		"pass the value in as an argument.",
}

// NoLoopFunc flags a function created in a loop that closes over a variable the loop can change.
//
//	valid:   for (let i=0; i<l; i++) { (function() { i; }) }      let per iteration
//	valid:   for (const i of {}) { (function() { i; }) }          a constant binding
//	valid:   let a = 0; for (let i=0; i<l; i++) { (function() { a; }); }   never written
//	valid:   for (var i=0; i<l; i++) { (function() { undeclared; }) }      not resolved here
//	valid:   for (var i = 0; i < 10; ++i) { (()=>{ i;})() }       an immediately invoked function
//	valid:   for (var i=0, a=function() { i; }; i<l; i++) { }     the initializer is outside
//	invalid: for (var i=0; i<l; i++) { (function() { i; }) }
//	invalid: for (var i in {}) { (function() { i; }) }
//	invalid: for (let i=0; i<l; i++) { (function() { i; }); i = 7; }   written after the border
//
// A `var` declared in a loop is one binding for the whole loop, so a function made in the body
// closes over that binding rather than over the value at the time. Every iteration produces a
// function that will later read whatever the last iteration left behind, which is the classic
// "all my callbacks print 10" bug. `let` and `const` in the loop head are a fresh binding per
// iteration, so they are safe, and that is the repair upstream is pointing at.
//
// # What makes a reference unsafe, and it is not simply "declared with var"
//
// Upstream's `isSafe` answers in three steps, and each is reproduced:
//
//	a constant binding is safe            const, using, await using
//	a `let` declared INSIDE the loop      a different instance each iteration
//	otherwise, safe only if no write to   the border is where the outermost containing loop starts
//	that binding sits at or after a border
//
// The third is the one that carries the real judgment, and it is why
// `let a = 0; for (let i=0; i<l; i++) { (function() { a; }); }` is clean while the same code with
// `a = 1` after the loop is not. The variable is captured either way; what differs is whether
// anything can still change it.
//
// # The border, and why it is the OUTERMOST loop
//
// `getTopLoopNode` climbs from the containing loop through every enclosing loop and takes the
// outermost one's start position, stopping at the declaration when the binding is a `let`. A write
// anywhere at or after that position is unsafe, because an outer iteration can run it and then run
// the inner loop again. Taking the innermost loop instead would call
// `for (var i=0; i<l; i++) { for (var j=0; j<m; j++) { (function() { i+j; }) } }` safe in `i`, and
// upstream reports both names on it.
//
// # Where our substrate differs from upstream's, and what replaces it
//
// Upstream reads `sourceCode.getScope(node).through`, which is eslint-scope's list of references
// that escape a function unresolved by it, then `variable.references` for every reference to a
// binding, then `reference.isWrite()`. We have no resolved-reference index, so all three are
// rebuilt on the checker:
//
//	through            every identifier under the function whose resolved declaration is NOT
//	                   itself under that function. Measured to be the same set for this rule's
//	                   purposes, because a reference resolving to a declaration inside the
//	                   function is exactly what eslint-scope resolves locally and drops.
//	variable.refs      every identifier in the file resolving to the same declaration, gathered
//	                   once per file rather than per function. `no-class-assign` uses the same
//	                   inversion of the checker's one-way answer.
//	isWrite            `reference.WritesToBinding`, which is the shelf's union of the write forms
//	                   and is what `no-class-assign` and three siblings already ask.
//
// `variable.scope.variableScope === upperRef.from.variableScope` is upstream asking whether the
// write happens in the same function the variable belongs to. Reproduced by comparing the enclosing
// function-like of the declaration against that of the write, with the source file standing in for
// the outermost scope.
//
// # Immediately invoked functions, and the second half of that arm nobody expects
//
// A function called where it is made cannot outlive the iteration, so upstream skips it. The skip
// is conditional twice over: it applies only to a function that is neither `async` nor a generator,
// and a NAMED function expression whose own name is referenced from the escaping set is not
// skipped, because the name being reachable means something else can hold onto it.
//
// The skip also feeds back into the walk: a skipped function is recorded, and `getContainingLoopNode`
// then climbs THROUGH it when judging a function nested inside it. That is what makes
// `arr.push((f => f)((() => i)()));` clean, and it is why the recording is a set on the rule rather
// than a local. Our walk visits parents outward, so the same set is consulted while climbing.
//
// # What is outside the loop, and it is not the whole head
//
// A `for` statement's initializer runs once, so a function created there is not in the loop, and
// `for...in` / `for...of` evaluate their right-hand side once for the same reason. Both are corpus
// cases: `for (var i=0, a=function() { i; }; i<l; i++) { }` and
// `for (var x in xs.filter(function(x) { return x != upper; })) { }` are clean. The test and the
// update of a `for` are NOT outside; upstream returns the loop for them, and
// `for (var i=0; (function() { i; })(), i<l; i++) { }` is clean only because it is an immediately
// invoked function.
//
// # Which node kinds are judged, and why ours are seven where upstream's are three
//
// Upstream listens on FunctionDeclaration, FunctionExpression and ArrowFunctionExpression. In
// ESTree that reaches every function-like shape in the language, because an object shorthand
// method, a getter, a setter, a class method, a class constructor and a static method ALL carry a
// FunctionExpression as their `value`. Our parser gives each its own kind, so the same three
// listeners reach none of them.
//
// Measured against the installed eslint 10.8.1 build: all eleven method-like shapes report, and
// with three listeners cohere reported two. The corpus writes no method, accessor or class in any
// of its 96 cases, so nothing imported could see it. It was found by a cross-linter comparison on
// the real tree, where eslint reported three findings cohere missed, all three object shorthand
// methods inside a `for(;;)` loop capturing reassigned outer bindings.
//
// The loop-boundary climb is widened to match. Without that half, a closure written inside a method
// would climb past the method, find the loop outside it, and be judged as though it ran once per
// iteration.
//
// # A stated divergence: where the finding points on a method
//
// ESTree's FunctionExpression for `onStatement(sql) { ... }` begins at the parameter list, because
// the key belongs to the enclosing Property rather than to the function. So upstream's finding
// starts at `(sql)` and ours starts at `onStatement`.
//
// Ours is deliberate, and it is the tree's convention rather than this rule's invention.
// `require-yield` shipped before this rule, listens on `KindMethodDeclaration` and reports the
// node, and diverges from upstream on the same axis in the same direction. Measured on
// `const o = { *onStatement(sql) { return 1; } };`:
//
//	eslint require-yield   "*onStatement"                     the ESTree node it was handed
//	cohere require-yield   "*onStatement(sql) { return 1; }"  the whole method declaration
//	eslint no-loop-func    "(statementSql) { return u; }"     the FunctionExpression value
//	cohere no-loop-func    "onStatement(statementSql) { ... }" the whole method declaration
//
// The two upstream spans do not even agree with each other, because each rule reports whichever
// ESTree node its listener happened to receive, and for a method those are different nodes. Ours
// agree, because both anchor on the one node our parser gives a method.
//
// The method name is the part a reader navigates by, and a span opening on an anonymous parameter
// list tells them which line but not which member. The gate agrees this is not a parity question:
// `internal/differential/differential.go` keys a finding on file, line and rule, and its comment
// says column is excluded because the two gates locate findings at different offsets often enough
// that comparing it would drown the real disagreements. Measured on the real tree, all ten findings
// match on file and line; three differ in column by exactly the method name.
//
// Pinned by TestNoLoopFuncMethodSpanIncludesTheName so a later reader does not narrow it to match
// upstream's node boundaries without knowing this was a choice.
//
// # No fix
//
// The repair is either changing a `var` to `let`, which changes the binding's scope and is not
// always what was meant, or restructuring to pass the value in. Both are judgment, and upstream
// ships no fixer either.
var NoLoopFunc = rule.Rule{
	Name: "no-loop-func",

	// The whole discrimination is name resolution. Four clean cases write a name spelled exactly
	// like an unsafe one and differ only in what it binds to, so a syntactic port reports them.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Recorded across the file, because a skipped immediately-invoked function has to stay
		// transparent to the loop search of every function nested inside it. Upstream keeps the
		// same set for the same reason.
		skippedImmediatelyInvoked := map[*ast.Node]bool{}

		// Every identifier in the file, grouped by the declaration it resolves to. Built once on
		// the source file rather than per function: upstream reads it off eslint-scope's index,
		// which is also built once. And built only when a function sits in a loop, since no other
		// function can report: most files have none, and the index resolves every identifier.
		var referencesByDeclaration map[*ast.Node][]*ast.Node

		check := func(node *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}
			loopNode := loopFuncContainingLoop(node, skippedImmediatelyInvoked)
			if loopNode == nil {
				return
			}
			if referencesByDeclaration == nil {
				referencesByDeclaration = loopFuncReferenceIndex(ctx)
			}
			checkLoopFunc(ctx, node, loopNode, skippedImmediatelyInvoked, referencesByDeclaration)
		}

		// Every kind whose ESTree counterpart is a FunctionExpression, which is what upstream's
		// three listeners cover. See loopFuncIsFunctionLike for the measurement.
		return rule.Listeners{
			ast.KindFunctionDeclaration: check,
			ast.KindFunctionExpression:  check,
			ast.KindArrowFunction:       check,
			ast.KindMethodDeclaration:   check,
			ast.KindGetAccessor:         check,
			ast.KindSetAccessor:         check,
			ast.KindConstructor:         check,
		}
	},
}

// checkLoopFunc judges one function, which sits in loopNode.
func checkLoopFunc(
	ctx rule.Context,
	node *ast.Node,
	loopNode *ast.Node,
	skippedImmediatelyInvoked map[*ast.Node]bool,
	referencesByDeclaration map[*ast.Node][]*ast.Node,
) {
	escaping := loopFuncEscapingReferences(ctx, node)

	// A function invoked where it is created cannot outlive the iteration. Upstream applies this
	// only to a function that is neither async nor a generator, so an async immediately-invoked
	// function is still judged.
	if !loopFuncIsAsyncOrGenerator(node) && loopFuncIsImmediatelyInvoked(node) {
		if !loopFuncNameIsReferenced(node, escaping) {
			skippedImmediatelyInvoked[node] = true
			return
		}
	}

	// Distinct names, in the order their references appear, because the message lists them and the
	// corpus asserts `'i', 'j'` rather than either alphabetically or by declaration.
	unsafeNames := []string{}
	seen := map[string]bool{}
	for _, identifier := range escaping {
		declaration := loopFuncDeclarationFor(ctx, identifier)
		if declaration == nil {
			// Upstream requires `r.resolved`: an undeclared name is no-undef's business, not this
			// rule's, and six corpus cases turn on it.
			//
			// Kept although a sweep shows it SUBSUMED rather than load-bearing. Deleting it leaves
			// every fixture green, because an unresolved name reaches `loopFuncReferenceIsSafe`
			// with a nil declaration, `loopFuncReferenceIndex` never stores a nil key, and the
			// empty reference list makes the write loop answer safe. The verdict arrives by
			// accident of the index rather than by the rule deciding it.
			//
			// It stays because it is also the guard that stops a nil declaration reaching
			// `loopFuncDeclarationStatement` and `loopFuncVariableScope`. Both tolerate nil today;
			// neither is required to, and the walk recovers per FILE rather than per rule, so a nil
			// dereference here would cost every rule its verdict on this file rather than only
			// this one.
			continue
		}
		if loopFuncReferenceIsSafe(ctx, loopNode, declaration, referencesByDeclaration, skippedImmediatelyInvoked) {
			continue
		}
		name := identifier.Text()
		if seen[name] {
			continue
		}
		seen[name] = true
		unsafeNames = append(unsafeNames, name)
	}

	if len(unsafeNames) == 0 {
		return
	}

	ctx.ReportNode(node, rule.Message{
		Id: messageNoLoopFunc.Id,
		Description: fmt.Sprintf("%s The unsafe reference is to '%s'.",
			messageNoLoopFunc.Description, strings.Join(unsafeNames, "', '")),
	})
}

// loopFuncContainingLoop returns the loop a node sits inside, or nil.
//
// Nested functions are a boundary, with one exception: a function already recorded as a skipped
// immediately-invoked one is climbed through, because it runs in the iteration that made it.
func loopFuncContainingLoop(node *ast.Node, skippedImmediatelyInvoked map[*ast.Node]bool) *ast.Node {
	for current := node; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindWhileStatement, ast.KindDoStatement:
			return parent

		case ast.KindForStatement:
			// The initializer runs once, before the loop, so a function created there is not in
			// the loop. The test and the update are in it.
			if parent.AsForStatement().Initializer != current {
				return parent
			}

		case ast.KindForInStatement, ast.KindForOfStatement:
			// The iterated expression is evaluated once.
			if parent.AsForInOrOfStatement().Expression != current {
				return parent
			}

		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindConstructor:
			if skippedImmediatelyInvoked[parent] {
				break
			}
			return nil
		}
	}
	return nil
}

// loopFuncTopLoop returns the outermost loop containing the given loop, without passing a position
// the caller wants excluded.
//
// Upstream's `getTopLoopNode`. The border is the excluded node's END, so a `let` declared inside the
// loop stops the climb at the loop that declares it: an outer iteration cannot reach a binding that
// does not exist yet.
func loopFuncTopLoop(loopNode *ast.Node, excluded *ast.Node, skippedImmediatelyInvoked map[*ast.Node]bool) *ast.Node {
	border := 0
	if excluded != nil {
		border = excluded.End()
	}
	result := loopNode
	for current := loopNode; current != nil && current.Pos() >= border; {
		result = current
		current = loopFuncContainingLoop(current, skippedImmediatelyInvoked)
	}
	return result
}

// loopFuncReferenceIsSafe answers upstream's `isSafe` for one resolved binding.
func loopFuncReferenceIsSafe(
	ctx rule.Context,
	loopNode *ast.Node,
	declaration *ast.Node,
	referencesByDeclaration map[*ast.Node][]*ast.Node,
	skippedImmediatelyInvoked map[*ast.Node]bool,
) bool {
	kind := loopFuncBindingKindOf(declaration)

	// const, using and await using cannot be reassigned, so every iteration sees the same value and
	// there is nothing to capture wrongly. `NodeFlagsConstant` is exactly upstream's
	// CONSTANT_BINDINGS set; see no-const-assign for why that mask rather than NodeFlagsAwaitUsing.
	if kind == loopFuncBindingConstant {
		return true
	}

	// A `let` declared inside the loop is a fresh binding each iteration, which is the whole point
	// of `let` in a loop head.
	declarationStatement := loopFuncDeclarationStatement(declaration)
	if kind == loopFuncBindingLet && declarationStatement != nil &&
		declarationStatement.Pos() > loopNode.Pos() && declarationStatement.End() < loopNode.End() {
		return true
	}

	excluded := (*ast.Node)(nil)
	if kind == loopFuncBindingLet {
		excluded = declarationStatement
	}
	border := loopFuncTopLoop(loopNode, excluded, skippedImmediatelyInvoked).Pos()

	// Safe only if every reference to the binding is either a read, or a write that happens before
	// the border AND in the same function the variable itself belongs to.
	declarationScope := loopFuncVariableScope(declaration)
	for _, occurrence := range referencesByDeclaration[declaration] {
		if !loopFuncWritesTo(occurrence) {
			continue
		}
		if loopFuncVariableScope(occurrence) == declarationScope && occurrence.Pos() < border {
			continue
		}
		return false
	}
	return true
}

// loopFuncWritesTo reports whether an occurrence of a name stores a value into its binding.
//
// `reference.WritesToBinding` is the shelf's answer and it covers the assignment forms, but
// eslint-scope counts three DECLARATION positions as writes too, and this rule reads the write set
// rather than the assignment set. Measured by walking eslint-scope's own reference list, which is
// the surface upstream reads:
//
//	var i = 0;              W    the initializer stores
//	let i = 5;              W    the same
//	var i;                  -    no initializer, so nothing is stored
//	for (var i in o)        W    the iteration assigns the binding each round
//	for (var i of o)        W    the same
//	function f(p) {}        -    a parameter is bound by the call, not written here
//	for (var [i, j] of xs)  W    both names, through the binding pattern
//
// The `for...in` / `for...of` row is why this exists. Without it, `for (var i in {}) { (function()
// { i; }) }` has no write anywhere and reads as safe, and it is one of upstream's reported cases.
// Four of this rule's corpus cases turn on that row alone.
func loopFuncWritesTo(identifier *ast.Node) bool {
	if reference.WritesToBinding(identifier) {
		return true
	}
	// The declaration this name is bound by, climbing out through any binding pattern. A
	// destructured head like `for (var [i, j] of xs)` puts each name on a BindingElement rather
	// than on the declaration, and both names are still written every iteration.
	declarationNode := loopFuncDeclarationBinding(identifier)
	if declarationNode == nil {
		return false
	}
	declaration := declarationNode.AsVariableDeclaration()
	if declaration.Initializer != nil {
		return true
	}
	// A binding in a `for...in` or `for...of` head takes a new value every iteration, which is a
	// write even though nothing is spelled out at the site.
	list := declarationNode.Parent
	if list == nil || list.Parent == nil {
		return false
	}
	switch list.Parent.Kind {
	case ast.KindForInStatement, ast.KindForOfStatement:
		return list.Parent.AsForInOrOfStatement().Initializer == list
	}
	return false
}

// loopFuncDeclarationBinding returns the VariableDeclaration an identifier is the bound name of,
// climbing out through array and object binding patterns, or nil when the identifier is not a
// bound name at all.
//
// The climb stops at the first node that is not a pattern or a binding element, so a name in an
// initializer or a computed property key never reaches the declaration and answers nil.
func loopFuncDeclarationBinding(identifier *ast.Node) *ast.Node {
	current := identifier
	for current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Name() != current {
				return nil
			}
			return parent
		case ast.KindArrayBindingPattern, ast.KindObjectBindingPattern:
		case ast.KindBindingElement:
			// Only the bound name climbs. A default value or a property name in
			// `{ a: b = c }` is a different position and must not read as a write.
			if parent.AsBindingElement().Name() != current {
				return nil
			}
		default:
			return nil
		}
		current = parent
	}
	return nil
}

// loopFuncBindingKind is the three-way classification upstream reads off the declaration's kind.
type loopFuncBinding int

const (
	// loopFuncBindingOther is `var`, a parameter, a function, a class, an import: anything whose
	// binding is neither block-scoped-per-iteration nor immutable.
	loopFuncBindingOther loopFuncBinding = iota
	loopFuncBindingLet
	loopFuncBindingConstant
)

// loopFuncBindingKindOf classifies a declaration node.
//
// Upstream reads `declaration.kind` off the enclosing VariableDeclaration, and anything that is not
// a variable declaration gets `""`, which falls through both the constant test and the let test. So
// a parameter, a function declaration and a class declaration are all "other" and judged on their
// writes, which is what this reproduces.
func loopFuncBindingKindOf(declaration *ast.Node) loopFuncBinding {
	list := loopFuncDeclarationListFor(declaration)
	if list == nil {
		return loopFuncBindingOther
	}
	switch {
	case list.Flags&ast.NodeFlagsConstant != 0:
		return loopFuncBindingConstant
	case list.Flags&ast.NodeFlagsLet != 0:
		return loopFuncBindingLet
	}
	return loopFuncBindingOther
}

// loopFuncDeclarationListFor climbs from a declaration node to the VariableDeclarationList whose
// flags say `var`, `let` or `const`, or nil when the node is not a variable binding at all.
//
// # The climb is what a destructured binding needs, and its absence was a false positive on real code
//
// A binding written as `for (const [, channel] of channels)` resolves to a KindBindingElement, not
// to a KindVariableDeclaration, so a test that requires the latter answers "other" for it and the
// constant arm never fires. Measured on the ahra tree before this climb existed: four findings on
// `const` destructured loop bindings that the installed eslint 10.8.1 build does not make, in
// `DiscordApi.ts` twice, `FramerCommandLineInterface.ts` and `InputTimeRange.tsx`.
//
// No corpus case could see it. Upstream reads `declaration.kind` off the VariableDeclaration that
// ESTree already puts above every declarator including a destructured one, so the shape simply does
// not arise there, and its own corpus writes exactly one destructured loop head
// (`for (var [i, j] of ...)`), which is a `var` and lands on "other" either way.
func loopFuncDeclarationListFor(declaration *ast.Node) *ast.Node {
	for current := declaration; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindVariableDeclarationList:
			return current
		case ast.KindVariableDeclaration, ast.KindBindingElement,
			ast.KindArrayBindingPattern, ast.KindObjectBindingPattern:
		default:
			return nil
		}
	}
	return nil
}

// loopFuncDeclarationStatement returns the node whose range upstream compares against the loop's.
//
// Upstream uses `definition.parent`, the whole `VariableDeclaration` (which in ESTree is the
// statement, holding every declarator), rather than the single declarator. Ours calls that node the
// VariableDeclarationList, so the equivalent is the list, or the statement that wraps it where one
// does. A `for (let i = 0; ...)` head has a list with no statement above it, which is the shape
// that has to keep working.
//
// # The lift is EQUIVALENT under measurement, and it is upstream's shape rather than ours
//
// A mutant returning the list and never lifting survived the corpus and a fixture written for it.
// The two nodes differ by the trailing semicolon at the end and by any modifier at the start
// (measured: `declare let y: number;` puts the statement at 0 and the list at 7), and both ends
// feed position comparisons. But this function's result is only read in the `let` arm of
// `loopFuncReferenceIsSafe`, and for a one-byte or modifier-width difference to flip a verdict a
// loop boundary would have to fall exactly inside that gap.
//
// Kept because it is what upstream compares, and a divergence chosen for being unobservable today
// is the kind that stops being unobservable quietly. Not kept as a claim that it discriminates.
func loopFuncDeclarationStatement(declaration *ast.Node) *ast.Node {
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
		return declaration
	}
	list := declaration.Parent
	if list == nil {
		return declaration
	}
	if list.Parent != nil && list.Parent.Kind == ast.KindVariableStatement {
		return list.Parent
	}
	return list
}

// loopFuncVariableScope returns the function-like a node belongs to, or the source file.
//
// Upstream's `variableScope` is the nearest function or the module/global scope, skipping block
// scopes, which is exactly what this walk does by only stopping at function-likes.
func loopFuncVariableScope(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindClassStaticBlockDeclaration,
			ast.KindSourceFile:
			return current
		}
	}
	return nil
}

// loopFuncEscapingReferences is every identifier under a function that does not resolve to a
// declaration under that same function.
//
// This is upstream's `scope.through`. A name declared inside the function is resolved by
// eslint-scope and never appears there, and a name resolving to nothing at all does appear, which
// is why the undeclared cases have to be filtered on `resolved` afterwards rather than here.
//
// # A named function expression's own name escapes, and it does not look like it should
//
// `declaration == node` is the exception and it is not decoration. eslint-scope puts a named
// function expression's name in a scope of its own that WRAPS the function scope, so a body
// referencing that name resolves outward and the reference lands in `through`. Our checker resolves
// the same name to the function node itself, which is under the function, so the containment test
// alone drops it.
//
// Measured by reading `sourceCode.getScope(node).through` for
// `(function a(){ i; a; })()`: both `i` and `a` are there. Five corpus cases depend on it, because
// the immediately-invoked skip is declined exactly when the name is reachable, and without this the
// name is never seen and every one of them goes silent.
func loopFuncEscapingReferences(ctx rule.Context, node *ast.Node) []*ast.Node {
	escaping := []*ast.Node{}
	body := node.Body()
	if body == nil {
		return escaping
	}

	var walk func(current *ast.Node)
	walk = func(current *ast.Node) {
		if current == nil {
			return
		}
		if current.Kind == ast.KindIdentifier {
			if loopFuncIsReadableReference(current) {
				declaration := loopFuncDeclarationFor(ctx, current)
				if declaration == nil || !loopFuncIsUnder(declaration, node) ||
					declaration == node {
					escaping = append(escaping, current)
				}
			}
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}

	// Default values in the parameter list are evaluated per call and can reference outer names, so
	// upstream's scope for the function includes them.
	for _, parameter := range node.Parameters() {
		if initializer := parameter.AsParameterDeclaration().Initializer; initializer != nil {
			walk(initializer)
		}
	}
	walk(body)
	return escaping
}

// loopFuncIsReadableReference declines the identifier positions that are not references to a
// binding at all: a property name, a member access's name side, a label, an import or export
// specifier's wire name.
//
// # Every arm here is EQUIVALENT for this rule, and it is kept anyway
//
// A sweep neutralising the property-access arm survived the whole corpus and survived two rounds of
// fixtures written for it, which is the signal to stop writing fixtures and go read. Measured, the
// reason is that resolution already answers the question twice over:
//
//	o.u          where `o` is undeclared, `u` gets NO SYMBOL at all
//	o.u          where `o` is typed, `u` resolves to the PropertyAssignment at its own position
//	({u: 1})     the key likewise resolves to its own PropertyAssignment
//
// In no arrangement does a property name resolve to the VariableDeclaration the rule is tracking,
// so the reference index never groups it with the variable and no verdict can move. Upstream is
// filtering because eslint-scope hands it a reference list where the distinction would matter; our
// checker has already made it.
//
// Kept because it is cheap, because it states the intent at the site where a reader looks for it,
// and because it does not rest on the checker continuing to answer this way. Not kept as a claim
// that it discriminates: it does not, and a future reader finding the mutant alive should read this
// rather than write a third fixture.
func loopFuncIsReadableReference(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return true
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Name() != identifier
	case ast.KindQualifiedName:
		return parent.AsQualifiedName().Right != identifier
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Name() != identifier
	case ast.KindPropertyDeclaration, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor:
		return parent.Name() != identifier
	case ast.KindBreakStatement, ast.KindContinueStatement, ast.KindLabeledStatement:
		return false
	}
	return true
}

// loopFuncIsUnder reports whether a node sits inside an ancestor.
func loopFuncIsUnder(node *ast.Node, ancestor *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}

// loopFuncDeclarationFor resolves an identifier to the declaration it binds to.
//
// `LocalSymbol` is the fallback an exported declaration needs: the checker hands an exported
// declaration its own truncated symbol while the merged list lives there instead.
func loopFuncDeclarationFor(ctx rule.Context, identifier *ast.Node) *ast.Node {
	if ctx.TypeChecker == nil {
		return nil
	}
	// This file's declarations only: a position is compared below, and positions in two files do not
	// compare.
	declarations := rule.DeclarationsIn(ctx.SourceFile, ctx.TypeChecker.GetSymbolAtLocation(identifier))
	if len(declarations) == 0 {
		return nil
	}
	// The earliest declaration by position, so every reference to one binding agrees on which node
	// names it regardless of how the symbol's list happens to be ordered. Never index zero: a
	// symbol can carry several declarations and their order is not guaranteed to be source order.
	//
	// # Why choosing the earliest is EQUIVALENT here, and why it is still written
	//
	// A mutant taking the latest instead survived the whole corpus and survived a fixture written
	// for it. The reason is that this function's answer is used two ways and only one of them can
	// see the choice:
	//
	//	as the INDEX KEY, where any consistent choice groups the same references together,
	//	  because every reference to one symbol reaches the same list and picks the same node
	//	as a POSITION, which happens in exactly one place: the `let` arm of
	//	  loopFuncReferenceIsSafe, through loopFuncDeclarationStatement
	//
	// And a block-scoped binding cannot be redeclared. Probed over six shapes including
	// `let x = 1; let x = 2;` and `for (let i=0;...) { let i = 1; }`: block-scoped symbols carrying
	// more than one declaration came back 0 while the `var` control came back 5. So the arm that
	// reads a position never sees a symbol with a choice to make.
	//
	// Written this way because "any consistent choice" is only true while that stays true, and
	// because indexing zero is the pattern this tree has already been bitten by.
	earliest := declarations[0]
	for _, declaration := range declarations[1:] {
		if declaration.Pos() < earliest.Pos() {
			earliest = declaration
		}
	}
	return earliest
}

// loopFuncReferenceIndex groups every identifier in the file by the declaration it resolves to.
//
// This replaces `variable.references`, which eslint-scope maintains and we do not have. Built once
// per file, because the alternative is resolving the whole file again for every function in it, and
// shared through the file cache with no-unmodified-loop-condition, which reads the same index. Two
// rules each building it cost 29 MB in 394K objects on a cold ahra run (#942rdnn). Neither rule
// writes to it.
func loopFuncReferenceIndex(ctx rule.Context) map[*ast.Node][]*ast.Node {
	return rule.Cached(ctx.FileCache, "core.loopFuncReferenceIndex", func() map[*ast.Node][]*ast.Node {
		return buildLoopFuncReferenceIndex(ctx)
	})
}

// buildLoopFuncReferenceIndex is loopFuncReferenceIndex's fill.
func buildLoopFuncReferenceIndex(ctx rule.Context) map[*ast.Node][]*ast.Node {
	index := map[*ast.Node][]*ast.Node{}
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier {
			if loopFuncIsReadableReference(node) {
				if declaration := loopFuncDeclarationFor(ctx, node); declaration != nil {
					index[declaration] = append(index[declaration], node)
				}
			}
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(ctx.SourceFile.AsNode())
	return index
}

// loopFuncIsImmediatelyInvoked reports whether a function is the callee of the call that wraps it.
//
// Upstream tests `parent.type === "CallExpression" && parent.callee === node`, and its own AST puts
// the parenthesis outside the function expression, so `(function(){})()` has the function as a
// direct callee. Ours wraps it in a parenthesized expression, so the climb past parentheses is
// required rather than cosmetic and every corpus case for this arm is written with them.
func loopFuncIsImmediatelyInvoked(node *ast.Node) bool {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	return parent.AsCallExpression().Expression == current
}

// loopFuncIsAsyncOrGenerator reports whether a function may outlive the call that made it.
//
// Upstream declines the immediately-invoked skip for both, because an async function returns before
// its body finishes and a generator's body runs on later `next` calls, so neither is confined to
// the iteration that created it even when it is called there.
func loopFuncIsAsyncOrGenerator(node *ast.Node) bool {
	if node.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
		return true
	}
	// An arrow function cannot be a generator, and `AsteriskToken` lives on the concrete node
	// rather than on `ast.Node`, so each kind is asked separately. Same shape as
	// `require_atomic_updates.go`'s `canSuspend`, which settled this first.
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	}
	return false
}

// loopFuncNameIsReferenced reports whether a named function expression's own name appears among the
// references escaping it.
//
// Upstream checks this only for a FunctionExpression with an id, and it is what stops the
// immediately-invoked skip when something inside can still reach the function by name.
func loopFuncNameIsReferenced(node *ast.Node, escaping []*ast.Node) bool {
	if node.Kind != ast.KindFunctionExpression {
		return false
	}
	name := node.Name()
	if name == nil {
		return false
	}
	for _, identifier := range escaping {
		if identifier.Text() == name.Text() {
			return true
		}
	}
	return false
}
