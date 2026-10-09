package react

import (
	"strings"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoMultiCompOptions configures the rule.
//
// Upstream's schema is a single object with one boolean, `ignoreStateless`, defaulting to false.
// Our config layer unwraps the severity tuple before dispatch, so the decoder receives that object
// directly rather than upstream's one-element array.
type NoMultiCompOptions struct {
	// IgnoreStateless excludes function components and pragma wrappers from the count, leaving only
	// class and factory components. Absent means false, which is upstream's default.
	IgnoreStateless bool `json:"ignoreStateless"`
}

// DefaultNoMultiCompOptions is the unconfigured answer.
//
// Upstream reads `configuration.ignoreStateless || false`, so an absent option and an explicit
// false are the same rule. The corpus states that agreement directly: it ships cases under an
// explicit `{ignoreStateless: false}` and cases with no options at all, and both report.
func DefaultNoMultiCompOptions() NoMultiCompOptions {
	return NoMultiCompOptions{IgnoreStateless: false}
}

// DecodeNoMultiCompOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` only so a bare `"error"` configuration, which
// arrives as empty input, resolves to the documented default instead of erroring. The default here
// is false, so the zero value happens to be right; the decoder is explicit anyway because a future
// option defaulting to true would silently invert the rule through the generic helper, and this is
// the line where that would have to be noticed.
func DecodeNoMultiCompOptions(raw []byte) (any, error) {
	options := DefaultNoMultiCompOptions()
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

var messageOnlyOneComponent = rule.Message{
	Id:          "onlyOneComponent",
	Description: "Declare only one React component per file",
}

// detectedComponent is one component the file declares, with the shape that found it.
type detectedComponent struct {
	// node is what the finding anchors on, which differs per shape and is measured rather than
	// chosen. A function declaration reports on itself, an arrow assigned to a const reports on the
	// arrow rather than the declaration, a factory call reports on its object argument, and a
	// component wrapped in `memo` or `forwardRef` reports on the whole wrapping call.
	node *ast.Node

	// stateless is whether `ignoreStateless` excludes it. Upstream tests
	// `/Function/.test(component.node.type) || utils.isPragmaComponentWrapper(component.node)`,
	// which covers function declarations, function expressions, arrows, and pragma wrapper calls,
	// and leaves classes and factory objects counted.
	stateless bool
}

// NoMultiComp keeps a file from declaring more than one React component.
//
//	valid:   one component alone in a file
//	valid:   a component beside a plain helper function that returns no JSX
//	valid:   two function components under `{"ignoreStateless": true}`
//	invalid: two function components in one file
//	invalid: a function component and a class component in one file
//
// Every component after the first reports, so a file holding three components produces two
// findings. The finding anchors on the component that should move rather than on the top of the
// file, which is what makes the count the real cost of adopting the rule.
//
// # What counts as a component, and why this does not use the shelf's name predicate
//
// Upstream decides component-hood in `util/Components.js`, whose `isFirstLetterCapitalized` strips
// leading underscores and then asks whether the first character equals its own uppercase. That is
// true for every character with no case at all, so `$foo`, `_1`, and `테스트` are components
// upstream and `_Foo` is one too. `react.IsLikelyComponentName` asks `unicode.IsUpper`, which
// answers false for all four.
//
// Measured against the installed build on 2026-08-27, one name per run beside a control component:
// `$foo`, `_1`, `테스트`, `_Foo`, and `__Foo` all report, and `foo` and `_foo` do not. So the shelf
// helper would silence this rule on five shapes upstream counts. `hasComponentCapitalization` below
// reproduces upstream's predicate instead, and is deliberately local rather than pushed onto the
// shelf: the shelf's answer is the one three other rules already depend on, and this rule wanting a
// different one is not evidence that they want it too.
//
// # No file suffix gate
//
// Upstream has no filename condition anywhere, and three rules in this package carry a `.tsx` gate
// inherited from an oxc port rather than from this authority. Reproducing it here would blind the
// rule to every `.ts` file. `TestNoMultiCompHasNoFileSuffixGate` pins the absence.
//
// # No fix
//
// Moving a component to its own file means creating that file, naming it, moving the imports it
// needs, and updating every reference. Upstream ships no fixer either.
var NoMultiComp = rule.Rule{
	Name: "react/no-multi-comp",

	// The checker is reached through this package's `isPragmaCreateElementCall`, to answer whether a
	// bare `createElement(...)` resolves to a binding introduced by the React module. Upstream
	// answers the same question with scope analysis. Measured: without it, upstream's own passing
	// case 5, which writes `import React, { createElement } from "react"` and returns
	// `createElement("img")`, would be judged not a component.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		settings := DefaultNoMultiCompOptions()
		if configured, isConfigured := rule.OptionsAs[NoMultiCompOptions](options); isConfigured {
			settings = configured
		}

		// The verdict is about the file as a whole and cannot be reached until every component is
		// known, which is what upstream's `Program:exit` buys. This walk is pre-order, so a
		// source-file listener fires before its children and would see nothing collected. Gathering
		// inside the source-file listener puts the collection and the verdict in one place.
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// A typed rule handed a nil checker goes silent rather than crashing, which is the
				// more dangerous of the two failure modes because every quiet fixture passes
				// vacuously. `TestNoMultiCompRequiresTheTypedHarness` fails loudly if this listener
				// is ever reached without one.
				if ctx.TypeChecker == nil {
					return
				}

				components := collectDetectedComponents(ctx, node)
				if len(components) <= 1 {
					return
				}

				kept := components[:0:0]
				for _, component := range components {
					if settings.IgnoreStateless && component.stateless {
						continue
					}
					kept = append(kept, component)
				}

				// Upstream filters first and slices second, so `ignoreStateless` can remove the
				// component that would otherwise have been the exempt first one. A file holding a
				// function component and then a class reports nothing under the option, because
				// after filtering the class is the only entry and the slice leaves it out.
				if len(kept) <= 1 {
					return
				}
				for _, component := range kept[1:] {
					ctx.ReportNode(component.node, messageOnlyOneComponent)
				}
			},
		}
	},
}

// collectDetectedComponents finds every component the file declares, in source order.
//
// Upstream registers per-kind visitors and lets ESLint order them, which for a single traversal is
// source order. This walks the tree once and appends as it goes, which produces the same order.
func collectDetectedComponents(ctx rule.Context, sourceFile *ast.Node) []detectedComponent {
	components := []detectedComponent{}
	seen := map[*ast.Node]bool{}

	// detectedNames is the running list `nodeWrapsComponent` consults, holding the names of the
	// class and const-arrow components found so far. Upstream reads its live component list at the
	// moment the wrapper is judged, so the list is built as the walk proceeds rather than up front,
	// and a wrapper written above the component it names does not see it.
	detectedNames := map[string]bool{}

	add := func(node *ast.Node, stateless bool) {
		if node == nil || seen[node] {
			return
		}
		seen[node] = true
		components = append(components, detectedComponent{node: node, stateless: stateless})
		if name, hasName := detectedComponentName(node); hasName {
			detectedNames[name] = true
		}
	}

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		switch node.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if isComponentClass(node) {
				add(node, false)
			}

		case ast.KindCallExpression:
			// A pragma wrapper is added by upstream's own `CallExpression` detector, independently
			// of whether the function inside it is a component. That detector fires at every level
			// of a nest, so `memo(forwardRef(fn))` contributes two components rather than one, and
			// both are reported when a third component shares the file.
			//
			// Measured against the installed build on 2026-08-27: that input beside one other
			// component produces two findings, spanning `memo(forwardRef((p,r)=><span/>))` and
			// `forwardRef((p,r)=><span/>)`. Upstream's corpus never nests two wrappers, so this is
			// invisible to every imported fixture, and the port counted one until the differential
			// sweep found it.
			//
			// Upstream requires the first argument to be function-like here, which is what keeps
			// `memo(SomeComponent)` from counting as a new component.
			if isPragmaComponentWrapper(node, detectedNames) {
				arguments := node.AsCallExpression().Arguments
				if arguments != nil && len(arguments.Nodes) > 0 {
					if first := skipParenthesesOptional(arguments.Nodes[0]); first != nil &&
						isFunctionLikeExpression(first) {
						add(node, true)
					}
				}
			}

			// The factory form. Upstream anchors on the object literal argument rather than on the
			// call, because its `ObjectExpression` visitor is what adds it.
			if isCreateReactClassCall(node) {
				arguments := node.AsCallExpression().Arguments
				if arguments != nil && len(arguments.Nodes) > 0 {
					if first := arguments.Nodes[0]; first.Kind == ast.KindObjectLiteralExpression {
						add(first, false)
					}
				}
			}

		case ast.KindMethodDeclaration:
			// `{ Foo(props) { return <div/>; } }`. A method shorthand in an object literal is its
			// own node kind here rather than a function expression under a property, so it never
			// reaches statelessComponentFor and is judged in place. Upstream's `Property` arm
			// covers it through `node.parent.method`.
			if node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
				name := node.Name()
				if name != nil && name.Kind == ast.KindIdentifier &&
					hasComponentCapitalization(name.Text()) && returnsJsx(ctx, node) {
					add(node, true)
				}
			}

		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
			if component, isComponent := statelessComponentFor(ctx, node, detectedNames); isComponent {
				// A pragma wrapper reports on the wrapping call, and upstream still classifies it
				// as stateless for `ignoreStateless`, through the second half of its `isIgnored`
				// test. So the flag is read from what was found rather than from what was reported.
				add(component, true)
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile)

	return components
}

// statelessComponentFor reproduces upstream's `getStatelessComponent`, narrowed to what this rule
// consumes: whether the function is a component, and which node the finding anchors on.
//
// Upstream returns the node itself in most arms and the wrapping `memo(...)` or `forwardRef(...)`
// call in the pragma arm, and this rule reports on whatever it returns. That is why the anchor
// differs per shape rather than always being the declaration.
//
// The arms below follow upstream's order, because several of them overlap and the first match wins
// there too. Reordering them changes the answer on inputs where two arms both apply.
func statelessComponentFor(ctx rule.Context, node *ast.Node, detectedNames map[string]bool) (*ast.Node, bool) {
	if node.Kind == ast.KindFunctionDeclaration {
		// An async generator is banned outright upstream, with confidence 0.
		if isAsyncGenerator(node) {
			return nil, false
		}
		name := node.Name()
		// An unnamed function declaration is a default export and still counts.
		if name != nil && !hasComponentCapitalization(name.Text()) {
			return nil, false
		}
		if !returnsJsxOrNull(ctx, node) {
			return nil, false
		}
		return node, true
	}

	if node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction {
		return nil, false
	}
	if node.Kind == ast.KindFunctionExpression && isAsyncGenerator(node) {
		return nil, false
	}

	parent := semanticParentOf(node)
	if parent == nil {
		return nil, false
	}

	// A function RETURNED by another function is not a component unless it returns real JSX.
	//
	//	const any = () => { return (props) => null }
	//	const any = () => (props) => null
	//
	// Upstream places this arm before every naming test, so the inner function is abandoned no
	// matter what the outer binding is called. Without it a curried function whose inner half
	// returns null is detected as a component, because the later arms reach it through the outer
	// binding's name.
	//
	// Added while porting `display-name`, which has five passing cases writing exactly this shape.
	// `no-multi-comp` had the same gap and its own corpus could not see it: measured against the
	// installed build, `demo = () => () => null;` beside one real component reports zero findings
	// upstream and reported one here. `TestNoMultiCompCurriedFunctionsAreNotComponents` pins it.
	//
	// The predicate is real JSX rather than JSX-or-null, which is upstream's `isReturningJSX`, so
	// an inner function returning only null is declined while one returning an element is kept.
	if parent.Kind == ast.KindReturnStatement ||
		(parent.Kind == ast.KindArrowFunction && parent.AsArrowFunction().Body == node) {
		if !returnsJsx(ctx, node) {
			return nil, false
		}
	}

	switch parent.Kind {
	case ast.KindExportAssignment:
		// `export default () => <div/>`. Upstream requires real JSX here rather than JSX-or-null,
		// which is the one arm where the two predicates differ, so a default-exported function
		// returning only null is not a component.
		if returnsJsx(ctx, node) {
			return node, true
		}
		return nil, false

	case ast.KindVariableDeclaration:
		if !returnsJsxOrNull(ctx, node) {
			break
		}
		name := parent.AsVariableDeclaration().Name()
		if name != nil && name.Kind == ast.KindIdentifier && hasComponentCapitalization(name.Text()) {
			return node, true
		}
		return nil, false

	case ast.KindPropertyAssignment:
		// `{ Foo: () => <div/> }`, judged by the property key. Upstream requires real JSX here.
		//
		// Upstream's condition is `!node.id && !node.parent.computed`, so this arm claims only an
		// ANONYMOUS function under a plain key. A named function expression written as
		// `{ a: function A() {...} }` deliberately falls past this arm and is judged by its own
		// name further down, which is why this breaks rather than returning.
		//
		// Measured against the installed build on 2026-08-27: that input reports on `function B`,
		// naming the function rather than the key, while the same object with anonymous functions
		// under the same lowercase keys is silent. An earlier draft returned here and went silent
		// on the first input; only the differential sweep saw it, since upstream's corpus writes no
		// named function expression under a property.
		if node.Kind == ast.KindFunctionExpression && node.Name() != nil {
			break
		}
		if parent.AsPropertyAssignment().Name() == nil || !returnsJsx(ctx, node) {
			return nil, false
		}
		key := parent.AsPropertyAssignment().Name()
		if key.Kind == ast.KindIdentifier && hasComponentCapitalization(key.Text()) {
			return node, true
		}
		return nil, false
	}

	// `React.memo(...)` and `React.forwardRef(...)`, including the destructured, imported, required
	// and aliased spellings the corpus writes in eleven separate cases. Reports on the call.
	if wrapper := pragmaComponentWrapperFor(node, detectedNames); wrapper != nil && returnsJsxOrNull(ctx, node) {
		return wrapper, true
	}

	// A method shorthand in an object literal, `{ Foo() { return <div/>; } }`, arrives as a
	// MethodDeclaration parent rather than a PropertyAssignment, so it is handled here rather than
	// in the switch above.
	if parent.Kind == ast.KindMethodDeclaration {
		if !returnsJsx(ctx, node) {
			return nil, false
		}
		key := parent.Name()
		if key != nil && key.Kind == ast.KindIdentifier && hasComponentCapitalization(key.Text()) {
			return node, true
		}
		return nil, false
	}

	if !isInAllowedPositionForComponent(node) || !returnsJsxOrNull(ctx, node) {
		return nil, false
	}

	// A named function expression is judged by its own name.
	if node.Kind == ast.KindFunctionExpression {
		if name := node.Name(); name != nil {
			if hasComponentCapitalization(name.Text()) {
				return node, true
			}
			return nil, false
		}
	}

	// An anonymous function assigned somewhere is judged by the name it is assigned to.
	//
	// Upstream splits this in two. A plain `foo = function () {...}` is decided in its own earlier
	// arm on `parent.left.name`, and a member assignment `exports.foo = function () {...}` is
	// decided here on `parent.left.property.name`, with `module.exports` exempted by name because
	// its property is deliberately lowercase and the assignment is still a component.
	//
	// Measured against the installed build on 2026-08-27, each row beside a control component:
	// `exports.Foo` reports and `exports.foo` does not, `Foo =` reports and `foo =` does not, and
	// `module.exports =` reports. Upstream's corpus writes none of these five, so this whole
	// paragraph is invisible to the imported fixtures; without it the rule reported two of them
	// wrongly and every imported case still passed.
	if parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			target := skipParenthesesOptional(binary.Left)
			switch {
			case target == nil:
				return nil, false

			case target.Kind == ast.KindIdentifier:
				if hasComponentCapitalization(target.Text()) {
					return node, true
				}
				return nil, false

			case target.Kind == ast.KindPropertyAccessExpression:
				access := target.AsPropertyAccessExpression()
				property := access.Name()
				if property == nil {
					return nil, false
				}
				if isModuleExportsTarget(access) {
					return node, true
				}
				if hasComponentCapitalization(property.Text()) {
					return node, true
				}
				return nil, false
			}
		}
	}

	return node, true
}

// isModuleExportsTarget reports whether a member assignment target is `module.exports`.
//
// Upstream exempts it by name from the capitalization test, because `exports` is lowercase by the
// language's own spelling rather than by the author's choice, and the assignment is still the
// file's component.
func isModuleExportsTarget(access *ast.PropertyAccessExpression) bool {
	object := skipParenthesesOptional(access.Expression)
	if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "module" {
		return false
	}
	property := access.Name()
	return property != nil && property.Text() == "exports"
}

// isAsyncGenerator reports whether a function is both async and a generator, which upstream bans
// from being a component outright rather than merely declining to add it.
//
// An arrow function can be neither, so it is absent from the switch rather than defaulting to
// false by accident.
func isAsyncGenerator(node *ast.Node) bool {
	if node == nil {
		return false
	}
	var asterisk *ast.Node
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		asterisk = node.AsFunctionDeclaration().AsteriskToken
	case ast.KindFunctionExpression:
		asterisk = node.AsFunctionExpression().AsteriskToken
	default:
		return false
	}
	if asterisk == nil {
		return false
	}
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindAsyncKeyword {
			return true
		}
	}
	return false
}

// hasComponentCapitalization reproduces upstream's `isFirstLetterCapitalized`.
//
// Upstream strips leading underscores and then asks whether the first character equals its own
// uppercase form. That second test is true for every character with no case at all, so this accepts
// far more than "starts with an uppercase letter": `$foo`, `_1`, and `테스트` are all components
// upstream, and `_Foo` is one because the underscores come off first.
//
// Deliberately not `react.IsLikelyComponentName`, which asks `unicode.IsUpper` and therefore
// answers false for every one of those. Measured against the installed build on 2026-08-27: the
// five names above report and `foo` and `_foo` do not, so following the shelf would have silenced
// this rule on five real shapes. The reasoning is on the rule's doc comment as well, because that
// is where a reader deciding to "simplify" this back to the shelf would be looking.
func hasComponentCapitalization(name string) bool {
	trimmed := strings.TrimLeft(name, "_")
	if trimmed == "" {
		return false
	}
	first := []rune(trimmed)[0]
	return unicode.ToUpper(first) == first
}

// isInAllowedPositionForComponent reproduces upstream's predicate of the same name.
//
// A sequence expression is allowed only through its last operand, which is what makes the corpus's
// `const HelloComponent = (0, (props) => {...})` a component: the arrow is the final operand of a
// comma expression whose own parent is a variable declaration. Our parser spells a sequence as a
// binary expression with a comma operator rather than as its own kind, so the walk up is written
// against that shape.
func isInAllowedPositionForComponent(node *ast.Node) bool {
	parent := semanticParentOf(node)
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindVariableDeclaration,
		ast.KindBinaryExpression,
		ast.KindPropertyAssignment,
		ast.KindReturnStatement,
		ast.KindExportAssignment,
		ast.KindArrowFunction:
		if parent.Kind == ast.KindBinaryExpression {
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken == nil {
				return false
			}
			// Our parser spells both an assignment and a sequence as a binary expression, where
			// upstream has distinct `AssignmentExpression` and `SequenceExpression` kinds and
			// allows both. The operator is what separates them here.
			if binary.OperatorToken.Kind == ast.KindEqualsToken {
				return true
			}
			if binary.OperatorToken.Kind != ast.KindCommaToken {
				return false
			}
			// Only the last operand carries the value of a comma expression, so only it can be the
			// component. `(f, 0)` is not a component even though `f` is a function. The operand is
			// compared through the parenthesis skip for the same reason the parent walk uses it.
			if skipParenthesesOptional(binary.Right) != node {
				return false
			}
			return isInAllowedPositionForComponent(parent)
		}
		return true
	}
	return false
}

// returnsJsx reports whether a function returns real JSX, which is upstream's `isReturningJSX`.
func returnsJsx(ctx rule.Context, node *ast.Node) bool {
	return returnsJsxValue(ctx, node, false)
}

// returnsJsxOrNull additionally accepts a bare `null` return, which is `isReturningJSXOrNull`.
//
// The two differ in exactly three arms of `statelessComponentFor` and the difference is real: a
// default-exported function returning only null is not a component, while the same function
// assigned to a capitalized const is.
func returnsJsxOrNull(ctx rule.Context, node *ast.Node) bool {
	return returnsJsxValue(ctx, node, true)
}

// returnsJsxValue walks the returns of one function and asks whether any returns a JSX value.
//
// Upstream's `traverseReturns` descends only through blocks, if, for, while and switch, so it never
// enters a nested function. That is what keeps `const outer = () => { function Inner() { return
// <div/>; } }` from making `outer` a component, and it is reproduced here rather than approximated
// with a full subtree walk.
//
// # One narrowing, stated rather than silent
//
// Upstream's identifier arm resolves a name back to its declaration and asks whether that variable
// holds JSX, so `const jsx = <div/>; const Foo = () => jsx;` is a component there. This does not
// resolve identifiers, so that shape is silent here. It costs findings rather than adding them, no
// corpus case writes it, and adding it means a scope walk this rule otherwise has no use for. Named
// as a divergence so the next reader does not have to rediscover whether it was deliberate.
func returnsJsxValue(ctx rule.Context, node *ast.Node, acceptNull bool) bool {
	if node == nil {
		return false
	}

	found := false
	var considerReturned func(*ast.Node)
	considerReturned = func(returned *ast.Node) {
		if found || returned == nil {
			return
		}
		if isJsxValue(ctx, returned, acceptNull) {
			found = true
		}
	}

	// A concise arrow body is its own return.
	if node.Kind == ast.KindArrowFunction {
		body := node.AsArrowFunction().Body
		if body != nil && body.Kind != ast.KindBlock {
			considerReturned(body)
			return found
		}
	}

	body := functionBodyBlock(node)
	if body == nil {
		return false
	}

	var walkStatements func(*ast.Node)
	walkStatements = func(statement *ast.Node) {
		if found || statement == nil {
			return
		}
		switch statement.Kind {
		case ast.KindReturnStatement:
			considerReturned(statement.AsReturnStatement().Expression)
			return

		// This list is upstream's `traverseReturns` verbatim and is deliberately incomplete.
		//
		// Upstream descends only through `BlockStatement`, `IfStatement`, `ForStatement`,
		// `WhileStatement`, `SwitchStatement` and `SwitchCase`, so a return written inside a
		// do-while, a for-of, a for-in, or a try block is invisible to it and the function is not a
		// component. That reads like an upstream oversight and it is reproduced rather than
		// corrected, because widening it changes which files report.
		//
		// Measured against the installed build on 2026-08-27: a component whose only return sits in
		// a `do {...} while(0)` or a `for (const x of y)` is silent, while the same return in a
		// `for(;;)` or a `while(1)` is a component. An earlier draft of this port added the three
		// missing kinds as an obvious improvement and reported on two inputs upstream leaves alone;
		// the differential sweep is what caught it, since no imported fixture writes any of them.
		//
		// `CaseBlock` is included because our parser interposes it between a switch and its clauses
		// where upstream's `SwitchStatement` holds the cases directly, so omitting it would break
		// the switch case upstream does descend into.
		case ast.KindBlock,
			ast.KindIfStatement,
			ast.KindForStatement,
			ast.KindWhileStatement,
			ast.KindSwitchStatement,
			ast.KindCaseBlock,
			ast.KindCaseClause,
			ast.KindDefaultClause:
			statement.ForEachChild(func(child *ast.Node) bool {
				walkStatements(child)
				return found
			})
			return
		}
		// Anything else, a nested function most of all, is not descended into. Upstream's traversal
		// skips it, so a return written inside it belongs to that function rather than to this one.
	}
	walkStatements(body)

	return found
}

// functionBodyBlock returns the block body of a function-like node, or nil if it has none.
func functionBodyBlock(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		body := node.AsArrowFunction().Body
		if body != nil && body.Kind == ast.KindBlock {
			return body
		}
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Body
	}
	return nil
}

// isJsxValue reports whether one returned expression counts as JSX.
//
// Parentheses are unwrapped in a loop because our parser keeps them and upstream's folds them away,
// so `return (<div/>)` reaches upstream's switch as the element itself and would otherwise reach
// this one as a parenthesized expression. `((<div/>))` nests, which is why this is a loop and not a
// single step. `ast.SkipParentheses` is not used because it dereferences its argument and a return
// with no expression is a real shape here.
func isJsxValue(ctx rule.Context, node *ast.Node, acceptNull bool) bool {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return true

	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return isJsxValue(ctx, conditional.WhenTrue, acceptNull) || isJsxValue(ctx, conditional.WhenFalse, acceptNull)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return isJsxValue(ctx, binary.Left, acceptNull) || isJsxValue(ctx, binary.Right, acceptNull)
		case ast.KindCommaToken:
			// A sequence carries the value of its last operand and nothing else.
			return isJsxValue(ctx, binary.Right, acceptNull)
		}
		return false

	case ast.KindCallExpression:
		return createElementCallCounts(ctx, node)

	case ast.KindNullKeyword:
		return acceptNull
	}
	return false
}

// pragmaWrapperNames are the two React functions that wrap a function into a component.
//
// Upstream builds this from the configured pragma as `[{property: 'forwardRef', object: pragma},
// {property: 'memo', object: pragma}]`. The pragma is fixed to React here, as it is across this
// package; upstream's corpus has one case setting it to `Foo`, which is recorded as an unreproduced
// divergence in the test file.
var pragmaWrapperNames = map[string]bool{"forwardRef": true, "memo": true}

// pragmaComponentWrapperFor walks up from a function to the outermost `memo(...)` or
// `forwardRef(...)` call wrapping it, or returns nil.
//
// Upstream loops rather than checking one level, so `memo(forwardRef(fn))` reports on the outer
// call. The loop is reproduced for the same reason.
func pragmaComponentWrapperFor(node *ast.Node, detectedNames map[string]bool) *ast.Node {
	var wrapper *ast.Node
	for current := node.Parent; current != nil; current = current.Parent {
		if !isPragmaComponentWrapper(current, detectedNames) {
			break
		}
		wrapper = current
	}
	return wrapper
}

// isPragmaComponentWrapper reports whether a call is `React.memo(f)` or `React.forwardRef(f)`, in
// any of the spellings that demonstrably reach the function from React.
//
// The namespaced form is decided syntactically. The bare form is not: upstream accepts a bare
// `memo(...)` only when `isDestructuredFromPragmaImport` can show the name came from React, which
// keeps a locally defined `function memo(f) {...}` from turning its argument into a component.
//
// Measured against the installed build on 2026-08-27, one head per run over the corpus's own
// `memo((props) => <HelloComponent/>)` tail. Reporting: `import {memo} from 'react'`,
// `const memo = React.memo`, `const {memo} = React`, `const memo = require('react').memo`, and
// `const {memo} = require('react')`. Silent: no import at all, `import {memo} from 'preact'`, and a
// local `function memo(f) { return f; }`. The corpus writes seven of those eight and this rule
// reproduces all eight.
func isPragmaComponentWrapper(node *ast.Node, detectedNames map[string]bool) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	// Upstream's CallExpression detector additionally requires the first argument to be function
	// like before adding the call as a component, and its `getPragmaComponentWrapper` does not.
	// Only the wrapper question is asked here; the caller supplies the function.
	callee := call.Expression
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		if !react.IsNamespacedMember(callee, func(name string) bool {
			return pragmaWrapperNames[name]
		}) {
			return false
		}
		// The namespaced spelling alone is not enough: upstream additionally requires that the call
		// is not merely re-wrapping a component it has already detected. This gate is on the
		// member-expression branch only, and the asymmetry is upstream's rather than an oversight
		// here; see nodeWrapsComponent.
		return !nodeWrapsComponent(node, detectedNames)

	case ast.KindIdentifier:
		name := callee.Text()
		if !pragmaWrapperNames[name] {
			return false
		}
		return isNameFromReact(node, name)
	}
	return false
}

// isNameFromReact reports whether a bare identifier was bound from React somewhere in this file.
//
// Upstream reaches for scope analysis, resolving the identifier to its variable and inspecting the
// declaration. We have no resolved-reference index, so this scans the file's declarations for the
// name instead. The difference is that a shadowing binding in an inner scope is invisible to this,
// which widens the rule on a file that both imports `memo` from React and shadows it locally. No
// corpus case writes that shape, and it is recorded here rather than left for a reader to find.
//
// The five accepted spellings are the ones measured against the installed build; see
// `isPragmaComponentWrapper` for the table and the date.
func isNameFromReact(context *ast.Node, name string) bool {
	sourceFile := ast.GetSourceFileOfNode(context)
	if sourceFile == nil {
		return false
	}

	found := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if found || node == nil {
			return
		}

		switch node.Kind {
		case ast.KindImportDeclaration:
			// `import { memo } from 'react'`, and the aliased `import { memo as m }` binds `m`, so
			// the local name is what is compared.
			if !isReactModuleSpecifier(node.AsImportDeclaration().ModuleSpecifier) {
				break
			}
			// `imports.BindingsOf` decides which of the three shapes an import clause carries, so
			// the named list arrives already separated from a default or a namespace binding.
			for _, element := range imports.BindingsOf(node).Named {
				// The element's own name is the LOCAL binding, so `import {memo as m}` is matched
				// on `m` rather than on `memo`, which is what upstream's scope lookup answers too.
				if local := element.Name(); local != nil && local.Text() == name {
					found = true
					return
				}
			}

		case ast.KindVariableDeclaration:
			declaration := node.AsVariableDeclaration()
			initializer := declaration.Initializer
			if initializer == nil {
				break
			}
			target := declaration.Name()
			if target == nil {
				break
			}

			switch target.Kind {
			case ast.KindIdentifier:
				// `const memo = React.memo` and `const memo = require('react').memo`. The property
				// read is what carries the name, and the local name has to match too.
				if target.Text() != name {
					break
				}
				if initializer.Kind != ast.KindPropertyAccessExpression {
					break
				}
				access := initializer.AsPropertyAccessExpression()
				property := access.Name()
				if property == nil || property.Text() != name {
					break
				}
				if isReactValueExpression(access.Expression) {
					found = true
					return
				}

			case ast.KindObjectBindingPattern:
				// `const {memo} = React` and `const {memo} = require('react')`.
				if !isReactValueExpression(initializer) {
					break
				}
				for _, element := range target.AsBindingPattern().Elements.Nodes {
					if local := element.Name(); local != nil && local.Text() == name {
						found = true
						return
					}
				}
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(sourceFile.AsNode())

	return found
}

// isReactValueExpression reports whether an expression evaluates to the React module: the bare
// `React` identifier, or a `require('react')` call.
func isReactValueExpression(expression *ast.Node) bool {
	for expression != nil && expression.Kind == ast.KindParenthesizedExpression {
		expression = expression.AsParenthesizedExpression().Expression
	}
	if expression == nil {
		return false
	}

	if expression.Kind == ast.KindIdentifier {
		return expression.Text() == reactIdentifier
	}
	if expression.Kind != ast.KindCallExpression {
		return false
	}
	call := expression.AsCallExpression()
	if call.Expression == nil || call.Expression.Kind != ast.KindIdentifier ||
		call.Expression.Text() != "require" {
		return false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
		return false
	}
	return isReactModuleSpecifier(call.Arguments.Nodes[0])
}

// isReactModuleSpecifier reports whether a module specifier names React itself.
func isReactModuleSpecifier(specifier *ast.Node) bool {
	return specifier != nil && ast.IsStringLiteralLike(specifier) && specifier.Text() == "react"
}

// reactIdentifier is the pragma name the namespaced spellings are read against, matching the rest
// of this package.
const reactIdentifier = "React"

// semanticParentOf returns the nearest ancestor that upstream's parser would have made the parent.
//
// Our parser keeps parentheses as real `KindParenthesizedExpression` nodes and upstream's folds
// them away, so `const A = (0, (props) => ...)` gives the arrow a binary-expression parent whose own
// parent is a parenthesized expression rather than the variable declaration. Every arm of
// `statelessComponentFor` reads the parent's kind, so without this skip the whole sequence family is
// silent, and no imported fixture can flag it because upstream's corpus cannot express the shape.
//
// The loop nests because `((x))` does. `ast.SkipParentheses` is not used because it dereferences its
// argument and a node with no parent is a real shape here.
func semanticParentOf(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent
}

// skipParenthesesOptional unwraps parentheses from an expression, tolerating nil.
func skipParenthesesOptional(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// detectedComponentName is the name upstream's `getDetectedComponents` would list for a component.
//
// Upstream lists only two shapes: a class declaration, by its own name, and an arrow function
// assigned to a variable, by the variable's name. Everything else, function declarations included,
// contributes no name, so a wrapper around a function-declared component is still a wrapper.
func detectedComponentName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindClassDeclaration:
		if name := node.Name(); name != nil {
			return name.Text(), true
		}

	case ast.KindArrowFunction:
		parent := semanticParentOf(node)
		if parent == nil || parent.Kind != ast.KindVariableDeclaration {
			return "", false
		}
		if name := parent.AsVariableDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
			return name.Text(), true
		}
	}
	return "", false
}

// nodeWrapsComponent reports whether a namespaced wrapper call is merely re-wrapping a component
// already detected in this file, in which case upstream does not count the call as a new component.
//
// The name read is the outermost JSX element of the wrapped function's return, and only that one.
// So `React.forwardRef((props, ref) => <StoreListItem {...props} />)` beside a `StoreListItem`
// class is silent, while wrapping the same element in a `<div>` reports, because the name read is
// then `div`. Three of upstream's passing cases and one of its failing ones turn on exactly that
// distinction.
//
// # The asymmetry, measured rather than assumed
//
// Upstream applies this gate inside the `MemberExpression` branch of `isPragmaComponentWrapper`
// and not in the bare-identifier branch. Measured against the installed build on 2026-08-27:
// `React.forwardRef` wrapping a known class is silent, the same wrapping an unknown name reports,
// and a bare imported `forwardRef` wrapping a known component reports. So the bare spelling is
// never exempted. It reads like an oversight upstream and it is reproduced rather than corrected,
// because the corpus asserts all three rows.
func nodeWrapsComponent(call *ast.Node, detectedNames map[string]bool) bool {
	arguments := call.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return false
	}

	name, hasName := outermostReturnedElementName(arguments.Nodes[0])
	return hasName && detectedNames[name]
}

// outermostReturnedElementName reads the tag name of the JSX element a wrapped function returns.
//
// Upstream reads `node[0].body` directly, so it sees a concise arrow body or the first return of a
// block body, and nothing deeper. A function returning anything other than a JSX element yields no
// name.
func outermostReturnedElementName(wrapped *ast.Node) (string, bool) {
	if wrapped == nil {
		return "", false
	}

	var body *ast.Node
	switch wrapped.Kind {
	case ast.KindArrowFunction:
		body = wrapped.AsArrowFunction().Body
	case ast.KindFunctionExpression:
		body = wrapped.AsFunctionExpression().Body
	default:
		return "", false
	}
	if body == nil {
		return "", false
	}

	if body.Kind == ast.KindBlock {
		found := false
		for _, statement := range body.AsBlock().Statements.Nodes {
			if statement.Kind == ast.KindReturnStatement {
				body = statement.AsReturnStatement().Expression
				found = true
				break
			}
		}
		if !found {
			return "", false
		}
	}

	return jsxElementTagName(body)
}

// jsxElementTagName reads the tag name of a JSX element, or reports absence.
//
// A fragment has no name, which is what makes upstream's fragment-wrapped case report.
func jsxElementTagName(node *ast.Node) (string, bool) {
	node = skipParenthesesOptional(node)
	if node == nil {
		return "", false
	}

	// `jsx.ElementParts` reads the tag name off either opening form, which matters because a
	// self-closing `<Foo />` never produces a JsxOpeningElement and is the shape most wrapped
	// components are written in. A JsxElement carries its opening node, so it is unwrapped first.
	if node.Kind == ast.KindJsxElement {
		node = node.AsJsxElement().OpeningElement
		if node == nil {
			return "", false
		}
	}

	tagName, _ := jsx.ElementParts(node)
	// A member tag (`<Foo.Bar />`) and a namespaced one carry no plain name, and upstream reads
	// `openingElement.name.name`, which is undefined for both. A fragment has no tag node at all,
	// which is what makes upstream's fragment-wrapped case report.
	if tagName == nil || tagName.Kind != ast.KindIdentifier {
		return "", false
	}
	return tagName.Text(), true
}

// isComponentClass reports whether a class extends a React component base.
//
// Deliberately local rather than `react.IsEs6ComponentClass`, which answers false for
// `class A extends (React.Component) {}`. That helper's doc comment describes skipping parentheses
// on the heritage receiver, and it does so through `isIdentifierNamed` on the object of a property
// access; the parenthesis in that shape wraps the whole `React.Component` expression instead, so
// `isComponentBase` switches on `KindParenthesizedExpression` and falls through to false.
//
// Measured against the installed build on 2026-08-27: two classes both written
// `extends (React.Component)` report one finding, exactly as two plainly written ones do, so
// upstream counts the parenthesized form. Verified in our parser too, with a probe printing the
// heritage type kind, which is why this unwraps rather than trusting either doc comment.
//
// Upstream's corpus writes no parenthesized heritage anywhere, so no imported fixture can see this.
func isComponentClass(node *ast.Node) bool {
	var heritage *ast.NodeList
	switch node.Kind {
	case ast.KindClassDeclaration:
		heritage = node.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		heritage = node.AsClassExpression().HeritageClauses
	default:
		return false
	}
	if heritage == nil {
		return false
	}

	for _, clause := range heritage.Nodes {
		if clause.Kind != ast.KindHeritageClause {
			continue
		}
		// An `implements` clause is spelled with a different type kind by our parser, so the kind
		// test below already declines it, but the token is checked so the intent is on the page.
		if clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
			continue
		}
		types := clause.AsHeritageClause().Types
		if types == nil {
			continue
		}
		for _, typeNode := range types.Nodes {
			if typeNode.Kind != ast.KindExpressionWithTypeArguments {
				continue
			}
			if isReactComponentBase(typeNode.AsExpressionWithTypeArguments().Expression) {
				return true
			}
		}
	}
	return false
}

// isReactComponentBase reports whether an extends target names a React component base class.
//
// `React.Component` and `React.PureComponent`, plus the bare `Component` and `PureComponent` that a
// file importing them directly writes. Parentheses are unwrapped first, which is the whole reason
// this exists beside the shelf's version.
func isReactComponentBase(expression *ast.Node) bool {
	expression = skipParenthesesOptional(expression)
	if expression == nil {
		return false
	}

	switch expression.Kind {
	case ast.KindIdentifier:
		return isReactComponentBaseName(expression.Text())

	case ast.KindPropertyAccessExpression:
		access := expression.AsPropertyAccessExpression()
		object := skipParenthesesOptional(access.Expression)
		if object == nil || object.Kind != ast.KindIdentifier || object.Text() != reactIdentifier {
			return false
		}
		name := access.Name()
		return name != nil && isReactComponentBaseName(name.Text())
	}
	return false
}

// isReactComponentBaseName accepts the two base classes a component may extend.
func isReactComponentBaseName(name string) bool {
	return name == "Component" || name == "PureComponent"
}

// isCreateReactClassCall reports whether a call is the ES5 component factory.
//
// Upstream's `getCreateClass` defaults to the single name `createReactClass`, and its member branch
// compares the property against that same default. So `createReactClass({...})` is the only
// spelling that counts.
//
// Deliberately not `react.IsEs5ComponentCall`, which additionally accepts `createClass` in both the
// bare and namespaced forms. Measured against the installed build on 2026-08-27, two factory calls
// per file: `createReactClass` reports one finding, while `React.createClass` and a bare
// `createClass` both report none. Following the shelf would report on two shapes upstream leaves
// alone, and the corpus writes only `createReactClass` so it could not see the difference. This is
// the same narrowing `prefer-es6-class` in this package already had to make, for the same reason.
func isCreateReactClassCall(node *ast.Node) bool {
	if node.Kind != ast.KindCallExpression {
		return false
	}
	callee := skipParenthesesOptional(node.AsCallExpression().Expression)
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == createReactClassName

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		object := skipParenthesesOptional(access.Expression)
		if object == nil || object.Kind != ast.KindIdentifier || object.Text() != reactIdentifier {
			return false
		}
		name := access.Name()
		return name != nil && name.Text() == createReactClassName
	}
	return false
}

// createReactClassName is the only factory spelling upstream's default configuration accepts.
const createReactClassName = "createReactClass"

// createElementCallCounts reports whether a returned call constructs a React element.
//
// Delegates to this package's existing `isPragmaCreateElementCall`, which already reproduces
// upstream's `isCreateElement` for both accepting shapes, including resolving a bare
// `createElement` back to a binding introduced by the React module through the type checker.
//
// Reaching for it rather than writing a fourth copy, after a build-time name collision showed the
// decision already had a home here. The reuse is not merely convenient: my own first reading was
// wrong. Probing a bare `createElement("div")` with no import measured silent and I nearly encoded
// "bare calls do not count", but the same call under `import React, { createElement } from 'react'`
// reports, measured against the installed build on 2026-08-27. Upstream's own passing case 5 writes
// exactly that import, so a narrowed copy would have failed it.
func createElementCallCounts(ctx rule.Context, node *ast.Node) bool {
	return isPragmaCreateElementCall(ctx, node)
}

// isFunctionLikeExpression reports whether a node is a function expression or an arrow function,
// which is upstream's `astUtil.isFunctionLikeExpression`. A function declaration is deliberately
// excluded, because it cannot appear in an argument position.
func isFunctionLikeExpression(node *ast.Node) bool {
	return node != nil &&
		(node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction)
}
