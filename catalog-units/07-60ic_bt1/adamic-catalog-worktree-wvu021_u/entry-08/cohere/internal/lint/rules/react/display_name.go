package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// DisplayNameOptions configures the rule.
//
// Both keys default to false upstream, so the zero value is the correct default and the decoder
// stays simple. That is worth stating because the sibling `self-closing-comp` defaults BOTH of its
// options to true, where the same generic decoder would silently invert the rule.
type DisplayNameOptions struct {
	// IgnoreTranspilerName requires an explicit `displayName` even when the transpiler would infer
	// one from the binding. Absent means false.
	IgnoreTranspilerName bool `json:"ignoreTranspilerName"`

	// CheckContextObjects additionally requires a `displayName` on a `createContext` result.
	// Absent means false.
	CheckContextObjects bool `json:"checkContextObjects"`
}

// DefaultDisplayNameOptions is the unconfigured answer: both checks off.
func DefaultDisplayNameOptions() DisplayNameOptions {
	return DisplayNameOptions{}
}

// DecodeDisplayNameOptions reads this rule's configuration from the config layer.
//
// Our config layer unwraps the severity tuple before dispatch, so this receives upstream's option
// object rather than upstream's one-element array. Empty input is a rule configured as a bare
// `"error"` and resolves to the default.
func DecodeDisplayNameOptions(raw []byte) (any, error) {
	options := DefaultDisplayNameOptions()
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

var messageNoDisplayName = rule.Message{
	Id:          "noDisplayName",
	Description: "Component definition is missing display name",
}

var messageNoContextDisplayName = rule.Message{
	Id:          "noContextDisplayName",
	Description: "Context definition is missing display name",
}

// DisplayName reports a component that React DevTools cannot label.
//
//	valid:   function Hello() { return <div/>; }              the binding names it
//	valid:   const C = memo(() => <div/>); C.displayName = 'C'
//	invalid: const C = memo(() => <div/>);                     anonymous, unlabelled
//	invalid: export default function() { return <div/>; }
//
// # This rule is the inverse of the in-house `structure/react-component-no-display-name`
//
// That rule is enabled today and reports a hand-written `displayName` assignment, exempting only
// anonymous wrapper results. This one reports a component that lacks a display name, and its
// transpiler-name check exempts exactly the components whose binding already supplies one.
//
// The two therefore agree on the common cases and are not duplicates: a named function needs no
// assignment and gets none, and an anonymous `memo(...)` needs one and is allowed one. They meet on
// a narrower class, a NAMED component carrying a redundant assignment, which the in-house rule
// reports and this one is silent on. That is the in-house rule's own judgment and this rule does
// not contradict it, because silence is not disagreement.
//
// Worth a look before turning `ignoreTranspilerName` on, however: that option makes this rule
// demand an assignment on every component, including the named ones the in-house rule forbids from
// carrying it, and the two would then contradict on every component in the tree. Left at its
// default of false, which is upstream's default and the only setting under which the two coexist.
//
// # What counts as having a name
//
// Either an explicit `displayName`, or a binding the transpiler infers a name from, which upstream
// calls the transpiler name and which this rule reproduces in `hasTranspilerName`.
//
// # No fix
//
// The repair is to name a thing, which means choosing the name. Upstream ships no fixer either.
var DisplayName = rule.Rule{
	Name: "react/display-name",

	// The checker resolves the object of a `C.displayName` write back to its declaration, which is
	// what separates a write against the component from a write against a shadowing binding of the
	// same name. Measured: those two cases differ only in that shadowing and upstream reports one
	// and not the other, so a name comparison would be wrong.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		settings := DefaultDisplayNameOptions()
		if configured, isConfigured := rule.OptionsAs[DisplayNameOptions](options); isConfigured {
			settings = configured
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// A typed rule handed a nil checker goes silent rather than crashing, which is the
				// more dangerous failure because every quiet fixture then passes vacuously.
				// `TestDisplayNameRequiresTheTypedHarness` fails loudly if this is reached without
				// one.
				if ctx.TypeChecker == nil {
					return
				}
				reportMissingDisplayNames(ctx, node, settings)
			},
		}
	},
}

// reportMissingDisplayNames runs the whole analysis for one file.
//
// Upstream does its work in `Program:exit`, because the verdict cannot be reached until every
// `displayName` assignment in the file has been seen, and those can appear far below the definition.
// Our walk is pre-order, so a source-file listener fires before its children and gathering has to
// happen here rather than in per-kind listeners.
func reportMissingDisplayNames(ctx rule.Context, sourceFile *ast.Node, settings DisplayNameOptions) {
	components := collectDetectedComponents(ctx, sourceFile)

	// named is the set of component nodes that have a display name by either route.
	named := map[*ast.Node]bool{}
	for _, component := range components {
		if !settings.IgnoreTranspilerName && hasTranspilerName(component.node) {
			named[component.node] = true
		}
	}

	// A `createContext` result is tracked by its declaration node, because the later
	// `Ctx.displayName = ...` write resolves back to that same declaration through the checker.
	contexts := map[*ast.Node]*ast.Node{}
	if settings.CheckContextObjects {
		collectContextObjects(ctx, sourceFile, contexts)
	}
	namedContexts := map[*ast.Node]bool{}

	byNode := map[*ast.Node]bool{}
	for _, component := range components {
		byNode[component.node] = true
	}

	// One walk for every display-name declaration in the file, whatever form it takes.
	markDeclaredDisplayNames(ctx, sourceFile, byNode, named, contexts, namedContexts)

	for _, component := range components {
		if named[component.node] {
			continue
		}
		if isNestedPragmaWrapperArgument(component.node) {
			// `memo(forwardRef(fn))` reports once, on the outer call, rather than twice. Upstream
			// skips the inner one so a single anonymous component does not produce two findings.
			continue
		}
		if isShadowedPragmaWrapper(ctx, component.node) {
			continue
		}
		reportComponentNode(ctx, component.node)
	}

	// No second `CheckContextObjects` test here: the collection above is already gated on it, so
	// with the option off `contexts` is empty and this loop runs zero times. An earlier draft
	// carried the guard and a mutation sweep could not kill it, which is the same decision written
	// twice. The single site is the collection.
	for declaration, anchor := range contexts {
		if namedContexts[declaration] {
			continue
		}
		ctx.ReportNode(anchor, messageNoContextDisplayName)
	}
}

// hasTranspilerName reports whether the binding around a component supplies a name the transpiler
// would infer, which upstream treats as satisfying the rule.
//
// Five accepting shapes, matching upstream branch for branch.
//
// The `module.exports` exclusion in the object-assignment branch is upstream's and is the opposite
// of what the sibling `no-multi-comp` does with the same identifier: there `module.exports = fn`
// COUNTS as naming a component, here `module.exports = {...}` does NOT supply a transpiler name.
// Both are measured and both are upstream's, so the two rules disagree about `module.exports` on
// purpose.
func hasTranspilerName(node *ast.Node) bool {
	parent := semanticParentOf(node)

	// A pragma wrapper call is judged by its wrapped ARGUMENT rather than by the call, because
	// upstream's CallExpression handler asks `hasTranspilerName(node.arguments[0])`. So
	// `memo(function Hello(){})` is named by the inner function's own name, and
	// `const C = memo(() => ...)` is not named by the `const C` binding at all. Measured: the first
	// is silent and the second reports.
	if node.Kind == ast.KindCallExpression && isPragmaComponentWrapper(node, map[string]bool{}) {
		// A wrapper that is ITSELF wrapped is treated as named without asking about its argument.
		// Upstream's condition is `!isWrappedInAnotherPragma && (... || !hasTranspilerName(arg))`,
		// so being wrapped short-circuits the early return and the call is marked as declared.
		//
		// That is what makes `React.memo(React.forwardRef(anon))` silent as a whole: the inner
		// `forwardRef` call is named by this branch, and the outer `memo` call is named by its
		// `const C =` binding through the argument test below. Three of upstream's passing cases
		// write exactly this shape and all three failed before this branch existed.
		if pragmaComponentWrapperFor(node, map[string]bool{}) != nil {
			return true
		}

		arguments := node.AsCallExpression().Arguments
		if arguments == nil || len(arguments.Nodes) == 0 {
			return false
		}
		wrapped := skipParenthesesOptional(arguments.Nodes[0])
		if wrapped == nil {
			return false
		}
		return hasTranspilerName(wrapped)
	}

	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		// A named class carries its own name.
		return node.Name() != nil

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		if node.Name() != nil {
			return true
		}

	case ast.KindObjectLiteralExpression:
		if parent == nil {
			return false
		}
		// `const Hello = createReactClass({...})`, where the object is the factory's argument and
		// the declaration two levels up supplies the name.
		if grandparent := semanticParentOf(parent); grandparent != nil {
			if grandparent.Kind == ast.KindVariableDeclaration {
				return true
			}
			if grandparent.Kind == ast.KindBinaryExpression {
				binary := grandparent.AsBinaryExpression()
				if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
					// Assigned to something, which names it, unless that something is
					// `module.exports`, which upstream excludes by name.
					return !isModuleExportsAssignmentTarget(binary.Left)
				}
			}
		}
		return false
	}

	// The remaining shape is a function-like expression whose binding names it: a variable
	// declaration, an object property, or a method shorthand. Upstream additionally requires the
	// grandparent NOT to be an ES5 component factory call, so an anonymous `render` inside
	// `createReactClass({...})` does not name the component that surrounds it.
	// A method shorthand in an object literal is its own node kind here rather than a function
	// expression under a property, so the component node IS the method and its name is the key.
	// Upstream reaches the same answer through `node.parent.method === true`.
	if node.Kind == ast.KindMethodDeclaration {
		return node.Name() != nil && !isEs5ComponentFactoryArgument(semanticParentOf(node))
	}

	if node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction {
		return false
	}
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration, ast.KindPropertyAssignment, ast.KindMethodDeclaration:
		grandparent := semanticParentOf(parent)
		if grandparent == nil {
			return true
		}
		return !isEs5ComponentFactoryArgument(grandparent)
	}
	return false
}

// isModuleExportsAssignmentTarget reports whether an assignment target is `module.exports`.
func isModuleExportsAssignmentTarget(target *ast.Node) bool {
	target = skipParenthesesOptional(target)
	if target == nil || target.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := target.AsPropertyAccessExpression()
	object := skipParenthesesOptional(access.Expression)
	if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "module" {
		return false
	}
	name := access.Name()
	return name != nil && name.Text() == "exports"
}

// isEs5ComponentFactoryArgument reports whether a node is the object literal handed to the ES5
// component factory, which is upstream's `isES5Component(node.parent.parent)` test.
func isEs5ComponentFactoryArgument(node *ast.Node) bool {
	if node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	parent := semanticParentOf(node)
	return parent != nil && parent.Kind == ast.KindCallExpression && isCreateReactClassCall(parent)
}

// markDeclaredDisplayNames walks the file once and records every display name it finds.
//
// Upstream spreads this across four listeners: a class property, a member expression assignment, a
// class method, and a property inside an ES5 factory object. All four are folded into one walk here
// because the gathering has to happen inside the source-file listener anyway.
func markDeclaredDisplayNames(
	ctx rule.Context,
	sourceFile *ast.Node,
	componentNodes map[*ast.Node]bool,
	named map[*ast.Node]bool,
	contexts map[*ast.Node]*ast.Node,
	namedContexts map[*ast.Node]bool,
) {
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		switch node.Kind {
		case ast.KindPropertyDeclaration:
			// `static displayName = 'x'` inside a class body names the class. `enclosingClassOf`
			// is this package's existing helper, from `state_in_constructor.go`; a build-time
			// collision is how that was found, and its version is better because it asks
			// `ast.IsClassLike` rather than naming two kinds.
			if isDisplayNameKey(node.Name()) {
				if class := enclosingClassOf(node); class != nil {
					named[class] = true
				}
			}

		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			// `static displayName() {}` counts, and so does `static get displayName()`. Upstream
			// keys on the member's name alone, and its `MethodDefinition` visitor sees an accessor
			// under that same node type. Our parser gives an accessor its own kind, so all three
			// are named here. Measured: a class whose only display name is a static getter is
			// silent upstream even under `ignoreTranspilerName`.
			if isDisplayNameKey(node.Name()) {
				if class := enclosingClassOf(node); class != nil {
					named[class] = true
				}
			}

		case ast.KindPropertyAssignment:
			// `{ displayName: 'x' }` inside an ES5 factory object names that object.
			if isDisplayNameKey(node.AsPropertyAssignment().Name()) {
				if object := semanticParentOf(node); object != nil &&
					object.Kind == ast.KindObjectLiteralExpression {
					named[object] = true
				}
			}

		case ast.KindPropertyAccessExpression:
			// `C.displayName = ...`, the form that can appear far below the definition.
			access := node.AsPropertyAccessExpression()
			if !isDisplayNameKey(access.Name()) {
				break
			}
			target := relatedComponentFor(ctx, access.Expression, componentNodes)
			if target != nil {
				named[target] = true
			}
			for _, declaration := range resolvedDeclarationsOf(ctx, access.Expression) {
				if _, isContext := contexts[declaration]; isContext {
					namedContexts[declaration] = true
				}
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile)
}

// isDisplayNameKey reports whether a property key names `displayName`.
//
// Upstream accepts an identifier and a string literal, which is why a computed `['displayName']`
// and a quoted `'displayName':` both count.
func isDisplayNameKey(key *ast.Node) bool {
	if key == nil {
		return false
	}
	switch key.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral:
		return key.Text() == "displayName"
	case ast.KindComputedPropertyName:
		inner := skipParenthesesOptional(key.AsComputedPropertyName().Expression)
		return inner != nil && inner.Kind == ast.KindStringLiteral && inner.Text() == "displayName"
	}
	return false
}

// resolvedDeclarationsOf resolves an expression to every declaration it binds to.
//
// This is what separates a write against the component from a write against a shadowing binding of
// the same name, and it is why the rule declares the checker. Measured on 2026-08-27: upstream
// reports `const C = memo(...)` when an inner scope declares its own `C` and writes displayName on
// THAT, and stays silent when the write is against the outer one. A name comparison cannot tell
// those apart; the checker hands back declarations at different positions.
//
// # Why this returns a list rather than one declaration
//
// A symbol can carry several declarations through declaration merging, and the callers here are
// asking "does ANY declaration match this component", not "what single thing is this name". An
// earlier draft took the earliest by position, which reads as the careful choice and is wrong:
// probed on 2026-08-27, `interface C {}` written ABOVE `const C = memo(...)` puts the interface at
// index zero and at the earlier position, so the component would never be found and its
// `C.displayName` write would stop registering. Upstream is silent on that input, measured, and it
// stays silent here only because every declaration is offered.
//
// Never index blindly either way; the list is what the question wants.
func resolvedDeclarationsOf(ctx rule.Context, expression *ast.Node) []*ast.Node {
	expression = skipParenthesesOptional(expression)
	if expression == nil || ctx.TypeChecker == nil {
		return nil
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(expression)
	if symbol == nil {
		return nil
	}
	return symbol.Declarations
}

// relatedComponentFor resolves the object of a `X.displayName` write to the component it names.
//
// Upstream's `getRelatedComponent` walks ESLint's scope references, which we do not have. The
// checker answers the same question directly: resolve the object to its declaration, then find the
// component whose node is that declaration's initializer or the declaration itself.
//
// Reproducing the reference walk would be reproducing a workaround for a constraint we do not have.
// What has to match is the DECISION about which writes count, and that is what the shadowing
// measurement above pins.
func relatedComponentFor(ctx rule.Context, object *ast.Node, componentNodes map[*ast.Node]bool) *ast.Node {
	for _, declaration := range resolvedDeclarationsOf(ctx, object) {
		if found := componentForDeclaration(declaration, componentNodes); found != nil {
			return found
		}
	}
	return nil
}

// componentForDeclaration finds the component a single declaration stands for, or nil.
func componentForDeclaration(declaration *ast.Node, componentNodes map[*ast.Node]bool) *ast.Node {
	// The component may BE the declaration, for a function or class declaration.
	if componentNodes[declaration] {
		return declaration
	}

	// Or the declaration may be a property holding the component, which is how a nested member
	// path resolves: `Mixins.Greetings.Hello.displayName = 'Hello'` resolves its object to the
	// `Hello: function(){}` property assignment rather than to any variable. Upstream reaches the
	// same component by walking the member path and matching its rendered text; the checker
	// answers it directly and handles the nesting depth for free.
	if declaration.Kind == ast.KindPropertyAssignment {
		value := skipParenthesesOptional(declaration.AsPropertyAssignment().Initializer)
		if value != nil && componentNodes[value] {
			return value
		}
	}
	// A method shorthand in an object literal IS the component node, so the declaration matches
	// directly through the check above; this branch is for the property-with-a-function form.

	// Or it may be what the declaration initializes, for `const C = memo(...)` and friends.
	if declaration.Kind == ast.KindVariableDeclaration {
		initializer := skipParenthesesOptional(declaration.AsVariableDeclaration().Initializer)
		if initializer != nil && componentNodes[initializer] {
			return initializer
		}
		// A pragma wrapper's component node is the wrapping call, and the initializer IS that call,
		// so the branch above already covers it. An object literal handed to the ES5 factory is the
		// component instead, so the call's first argument is checked too.
		if initializer != nil && initializer.Kind == ast.KindCallExpression {
			arguments := initializer.AsCallExpression().Arguments
			if arguments != nil && len(arguments.Nodes) > 0 {
				first := skipParenthesesOptional(arguments.Nodes[0])
				if first != nil && componentNodes[first] {
					return first
				}
			}
		}
	}
	return nil
}

// isNestedPragmaWrapperArgument reports whether a component is a pragma wrapper call that is itself
// the argument of another pragma wrapper.
//
// `memo(forwardRef(fn))` is two components to the detector, because the wrapper detector fires at
// every level of a nest, and upstream reports it once rather than twice. So the INNER call is
// skipped and the outer one carries the finding.
//
// Measured on 2026-08-27: `memo(forwardRef(anon))`, `memo(memo(anon))` and `forwardRef(memo(anon))`
// all report zero findings, while `memo(anon)` alone reports one. Zero rather than one is worth
// reading twice, and it is not this function's doing: the outer call has a transpiler name from its
// `const C =` binding, so it is already exempt, and skipping the inner one takes the count to zero.
// Without the skip the inner call would report and the pair would produce a finding upstream does
// not.
func isNestedPragmaWrapperArgument(component *ast.Node) bool {
	if component.Kind != ast.KindCallExpression {
		return false
	}
	parent := semanticParentOf(component)
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	// The component has to be the wrapped first ARGUMENT rather than the callee or a later
	// argument.
	//
	// # This check is inert today, and it is kept rather than deleted
	//
	// A mutation neutralising it survived, and an A/B on `memo(x, forwardRef(anon))` and
	// `memo(a, b, forwardRef(anon))` produced byte-identical output with and without it on
	// 2026-08-27. The reason is upstream's own gate one level up: a wrapper call is only added as a
	// component when ITS first argument is function-like, and the detector never offers a call
	// sitting in a later argument slot, so this line is not reached on any input tried.
	//
	// The grammar does produce the shape, measured with a probe that found calls at argument
	// indices one and two inside another call, so this is "no detected component reaches it" rather
	// than "the parser cannot build it". That is a narrower claim than unreachability and it rests
	// on a gate in a different function, which is exactly the kind of verdict that expires when
	// someone widens the detector. Keeping the check costs one comparison and makes this function
	// correct on its own terms rather than correct by a neighbour's invariant.
	arguments := parent.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return false
	}
	if skipParenthesesOptional(arguments.Nodes[0]) != component {
		return false
	}
	return isPragmaComponentWrapper(parent, map[string]bool{})
}

// isShadowedPragmaWrapper reports whether a wrapper call's name is a local binding rather than
// React's.
//
// # A measured divergence between the clone and the installed build
//
// The clone carries `isShadowedComponent`, which exempts a wrapper whose identifier is shadowed in
// an enclosing scope, and it handles both a bare `memo(...)` and a shadowed `React.memo(...)`. The
// installed build, which is what our gate compares against, only has the bare half.
//
// Measured on 2026-08-27, both trees reporting version 7.37.5: a shadowed bare `memo` is silent
// installed, and a shadowed `React.memo` REPORTS installed while the clone's three corpus cases
// covering it expect silence. Those three cases are the only ones in the whole 116-case corpus
// where the installed build disagrees with the clone.
//
// This follows the installed build, because that is the artifact the differential compares against.
// The three clone cases are recorded in the test file as a known drift rather than deleted.
//
// The bare half is answered by resolution rather than by a scope walk: if the callee identifier
// resolves to a declaration in this file, it is a local binding and not React's.
func isShadowedPragmaWrapper(ctx rule.Context, component *ast.Node) bool {
	if component.Kind != ast.KindCallExpression {
		return false
	}
	callee := skipParenthesesOptional(component.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return false
	}
	if !pragmaWrapperNames[callee.Text()] {
		return false
	}

	declarations := resolvedDeclarationsOf(ctx, callee)
	if len(declarations) == 0 {
		return false
	}
	// An import binding is React's own; anything else declared in source is a local shadow. Asked
	// across every declaration rather than one, so a merged symbol carrying an import alongside
	// something else still counts as React's.
	for _, declaration := range declarations {
		switch declaration.Kind {
		case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport:
			return false
		}
	}
	return true
}

// collectContextObjects records every `createContext` result in the file, keyed by the declaration
// a later `displayName` write will resolve to.
//
// Upstream matches on the CALLEE NAME alone for the member form, so `other.createContext()` counts
// as a context object exactly as `React.createContext()` does. Measured, and reproduced rather than
// narrowed: the receiver is never checked.
func collectContextObjects(contextCheckerHolder rule.Context, sourceFile *ast.Node, contexts map[*ast.Node]*ast.Node) {
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		switch node.Kind {
		case ast.KindVariableDeclaration:
			declaration := node.AsVariableDeclaration()
			if isCreateContextCall(declaration.Initializer) {
				contexts[node] = node
			}

		case ast.KindBinaryExpression:
			// `Hello = createContext()`, an assignment rather than a declaration, which the corpus
			// writes three times against a `let` or `var` declared above.
			//
			// The map is keyed on the TARGET's declaration rather than on this assignment node,
			// because a later `Hello.displayName = ...` resolves through the checker to that
			// declaration and to nothing else. Keying on the assignment made all three cases
			// report despite carrying a display name, since the two lookups could never meet.
			//
			// The anchor stays this node, so the finding points at the assignment the reader would
			// edit rather than at a bare `let Hello;`.
			binary := node.AsBinaryExpression()
			if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
				break
			}
			if !isCreateContextCall(binary.Right) {
				break
			}
			target := skipParenthesesOptional(binary.Left)
			if target == nil || target.Kind != ast.KindIdentifier {
				break
			}
			for _, declaration := range resolvedDeclarationsOf(contextCheckerHolder, target) {
				contexts[declaration] = node
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile)
}

// isCreateContextCall reports whether an expression is a call to `createContext`, bare or through
// any receiver.
func isCreateContextCall(expression *ast.Node) bool {
	expression = skipParenthesesOptional(expression)
	if expression == nil || expression.Kind != ast.KindCallExpression {
		return false
	}
	callee := skipParenthesesOptional(expression.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == "createContext"
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		return name != nil && name.Text() == "createContext"
	}
	return false
}

// reportComponentNode reports one component, anchored the way upstream anchors it.
//
// Upstream's node for `export default function(){}` is the FunctionDeclaration its parser builds
// under an ExportDefaultDeclaration, so its span starts at the `function` keyword and the
// `export default` prefix is outside the finding. Our parser folds the modifiers onto the
// declaration itself, so reporting the node directly would include them.
//
// Measured against the installed build on 2026-08-27: that input reports
// `function(){ return <div/>; }` and not `export default function(){ return <div/>; }`. The
// difference is invisible to a message-id assertion and is pinned by
// TestDisplayNameAnchorsOnTheComponent.
//
// Every other component shape has no modifiers and is reported unchanged.
func reportComponentNode(ctx rule.Context, component *ast.Node) {
	modifiers := component.Modifiers()
	if modifiers == nil || len(modifiers.Nodes) == 0 {
		ctx.ReportNode(component, messageNoDisplayName)
		return
	}
	last := modifiers.Nodes[len(modifiers.Nodes)-1]
	if last.End() >= component.End() {
		ctx.ReportNode(component, messageNoDisplayName)
		return
	}
	// The modifier's end sits before the whitespace preceding the keyword, so the range is scanned
	// forward to the next token rather than taken raw. Without that the span carries a leading
	// space and the assertion fails by one byte, which reads exactly like an off-by-one in the
	// rule.
	start := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, last.End()).Pos()
	ctx.ReportRange(core.NewTextRange(start, component.End()), messageNoDisplayName)
}
