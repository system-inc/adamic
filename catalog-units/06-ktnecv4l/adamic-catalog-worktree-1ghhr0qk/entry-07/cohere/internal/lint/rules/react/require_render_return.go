package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoRenderReturn = rule.Message{
	Id: "noRenderReturn",
	Description: "A class component's `render` returns nothing, so React receives `undefined` " +
		"where it expected an element. React treats that as an error rather than as an empty " +
		"component, and the whole subtree fails to mount with a message naming the class rather " +
		"than the missing return. The usual cause is a body that builds the element and forgets " +
		"to hand it back, or an arrow whose braces turned an implicit return into a statement. " +
		"Return the element, or return `null` to render nothing on purpose.",
}

// RequireRenderReturn flags a class component whose `render` never returns.
//
//	valid:   class H extends React.Component { render() { return <div/>; } }
//	valid:   class H extends React.Component { render = () => <div/> }
//	valid:   var H = createReactClass({ render: function() { return <div/> } });
//	valid:   class H { render() {} }                          (not a component)
//	valid:   class H extends React.Component {}                (no render at all)
//	valid:   var H = createReactClass({ render });             (render is not function-like)
//	invalid: class H extends React.Component { render() {} }
//	invalid: var H = createReactClass({ render: function() {} });
//	invalid: class H extends React.Component { render() { [].map(function(){ return 1 }) } }
//
// Ported from `react/require-render-return` in `eslint-plugin-react`, which is the implementation
// that defined this rule. Every behavior recorded below was measured by driving the installed
// build, version 7.37.5, through the ESLint Linter API on inputs written for the question, because
// this rule's corpus is thirteen clean cases and four failing ones and is silent about most of what
// the port has to decide.
//
// `meta.schema` is `[]`, so there is no option surface, and `meta.fixable` is unset, so there is no
// repair to port. One message id.
//
// # The rule is a scope count, not a containment test
//
// The name reads as "does `render` contain a return" and answering that question is how a port gets
// this wrong in both directions at once. Upstream's `ReturnStatement` listener walks the ancestors
// of every return, counting how many function-like scopes it crosses, and marks the return as
// satisfying `render` only when the property named `render` is reached at a depth of one or less.
// So a return written inside a callback in `render` does not count, which is upstream's third
// failing case, and a return written inside an `if` or a `switch` in `render` does count, which is
// two of its clean ones. Blocks, conditionals, loops and switch bodies are transparent; functions
// are not.
//
// The regular expression upstream counts with is `/Function(Expression|Declaration)$/`, and it is
// **unanchored at the front**, so `ArrowFunctionExpression` matches it. A reading that treats
// arrows as transparent is wrong, and it is the natural reading: measured, `render() { var f = () =>
// { return 1; }; }` reports, so an arrow costs a scope exactly as a function expression does.
//
// # Where our AST loses a count, which is the same off-by-one the sibling rules carry
//
// ESLint's `MethodDefinition` is a wrapper whose `value` is a separate `FunctionExpression`, so a
// return written directly in a class method crosses one counted scope on its way out to the
// property node. Our `KindMethodDeclaration` fuses the two, so the same source crosses zero.
// Measured by tracing upstream's own ancestor walk on each of the three shapes the corpus writes:
//
//	class { render() { HERE } }            upstream depth 1, our walk 0
//	class { render = () => { HERE } }      upstream depth 1, our walk 1   (the arrow is a real node)
//	createReactClass({ render: fn() {} })  upstream depth 1, our walk 1   (the expression is real)
//
// So only the class-method row is short and it is short by exactly one, matching what
// `no-did-mount-set-state` documents for the same reason. The correction IS needed, and reasoning
// that it was not is how the first draft of this rule shipped three false negatives.
//
// The argument that failed is worth recording because it is the plausible one: the threshold is
// `depth <= 1` and a DIRECT return in a class method counts 0 here against upstream's 1, which is
// on the same side of the boundary, so the fusion looks harmless. It is not, because the boundary
// is crossed by returns that are NESTED. A return inside a callback in a class method counts 2
// upstream and 1 here, which is the difference between silent and marked, and it silently satisfies
// a `render` that upstream leaves unsatisfied. Three fixtures caught it, including upstream's own
// third failing case.
//
// The correction is applied at the property node rather than in the scope test, because that is
// the only place the missing node's identity is known: an object property's function IS a real
// separate node the walk has already counted, and only a class method or accessor is short.
//
// # The flag is per component, not per render method
//
// Upstream marks `hasReturnStatement` through `components.set`, which walks from the return node up
// to the nearest *registered* component and sets the flag there. Two consequences, both measured
// and neither visible from the corpus.
//
// A class with two `render` methods where only the second returns is **clean**, even though the
// finding would point at the first. Measured: `class H extends React.Component { render() {} render()
// { return <div/>; } }` is clean, and so is the same pair written the other way round, while making
// both empty reports. That is upstream being loose rather than upstream being right, and it is
// reproduced: the flag is a property of the component here too.
//
// The sharper case is a nested class that is **not** a component. `class Outer extends
// React.Component { render() { class Inner { render() { return <div/>; } } } }` is **clean**
// upstream, because `Inner` was never registered, so the walk from the inner return climbs past it
// and marks `Outer`. Write `class Inner extends React.Component` instead and `Outer` reports, since
// the flag then lands on `Inner`. This is a genuine false negative in upstream and it is
// reproduced rather than improved on, with fixtures pinning both halves. Established by running
// both shapes against the installed build rather than by reading `Components.set`.
//
// # What counts as a component, and it is narrower than the shelf
//
// The rule's own `Program:exit` filter re-tests `isES5Component || isES6Component`, which throws
// away everything else `Components.detect` registered. So the stateless function components, the
// `memo` and `forwardRef` wrappers, and the arrow components the detection layer collects can never
// reach a finding, and none of that machinery needs porting. Confirmed by measuring: a `memo`-wrapped
// object literal carrying an empty `render` is clean, as is a plain function component.
//
// The ES5 factory name is `createReactClass` alone, not the wider set `internal/utilities/react`
// accepts. `getCreateClassFromContext` defaults to the literal string `createReactClass` and only a
// `settings.react.createClass` entry changes it, which our config has no surface for. Measured on
// the installed build: `createClass({ render: function() {} })` and its `React.createClass` twin are
// both clean, while `createReactClass` and `React.createReactClass` both report. This calls
// `isEs5ComponentCallStrict` in `no_did_mount_set_state.go`, which is that same narrowing and
// already shared by three rules in this package.
//
// The ES6 base set is the shelf's and agrees with upstream here, including on the parenthesized
// receiver. `class H extends (React.Component) { render() {} }` **reports** upstream, measured,
// which is the opposite of what `no-direct-mutation-state` records for oxc. The shelf skips
// parentheses and is therefore right for this rule, and the divergence its doc comment warns about
// does not apply.
//
// Upstream also accepts a class declared a component by a JSDoc `@extends React.Component` tag,
// through `isExplicitComponent`. That is not ported. Measured: the tag is clean on the installed
// build in this repository's configuration both with and without a superclass, because
// `sourceCode.getJSDocComment` returns nothing for the shapes tried, so reproducing the branch
// would add findings upstream does not produce here rather than match it.
//
// # There is no file gate
//
// Upstream has none, and this rule does not add one. Three rules in this package gate on `.tsx` and
// `.jsx` because they were ported from oxc, which declares `should_run` on `source_type().is_jsx()`.
// That gate is oxc's own; `eslint-plugin-react` reports identically whatever the extension, and
// carrying it here would cost every finding in a `.ts` file. Pinned by a fixture running one source
// under four extensions.
//
// # Where the finding points
//
// At the property node that holds `render`, not at its function and not at the class. Measured
// against the installed build on each shape: `render() {}` underlines eleven characters including
// the empty body, `render: function() {}` underlines twenty-one, and a `render = () => {...}`
// property underlines the whole declaration across its lines. Upstream reports
// `node: findRenderMethod(component.node)`, and `findRenderMethod` returns the property rather than
// the value, so a port anchoring on the function would carry the right message and the wrong span.
// Every fixture below asserts the sliced text.
var RequireRenderReturn = rule.Rule{
	// No namespace prefix. The config writes `react/require-render-return` and the parity guard
	// strips the namespace on a `/` boundary, so `react-require-render-return` would match no
	// inventory entry and lint no files while passing every fixture in this package.
	Name: "react/require-render-return",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The judgment needs the whole file before it can be made: a return anywhere marks the
		// component that encloses it, and only at the end is it known which components were never
		// marked. There is no `rule.OnExit`, and the walk is pre-order, so `KindSourceFile` fires
		// before its children and the file is walked by hand from there.
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				collector := &renderReturnCollector{}
				collector.walk(node)
				for _, component := range collector.components {
					if component.hasReturnStatement {
						continue
					}
					if component.renderProperty == nil {
						continue
					}
					ctx.ReportNode(component.renderProperty, messageNoRenderReturn)
				}
			},
		}
	},
}

// renderReturnComponent is one component and what the file said about it.
//
// `renderProperty` is nil when the component has no `render` that is function-like, which is three
// of upstream's clean cases and is the reason it is a pointer rather than a bool beside a node.
type renderReturnComponent struct {
	node               *ast.Node
	renderProperty     *ast.Node
	hasReturnStatement bool
}

// renderReturnCollector gathers the components in source order and the returns that satisfy them.
//
// Source order matters for the report, not for the judgment: two broken components in one file
// report twice and upstream emits them in the order they were declared, which a map would not
// preserve. The index beside the slice is what makes the nested case cheap, since a return marks
// the nearest enclosing entry and that is a walk up the parent chain rather than a search.
type renderReturnCollector struct {
	components []*renderReturnComponent
	byNode     map[*ast.Node]*renderReturnComponent
}

// walk collects every component and then every return, in one pre-order pass.
//
// One pass is enough because a return is always inside the component it marks, so the component
// entry exists by the time the return is reached. Upstream gets the same ordering from ESLint's own
// traversal, where `ClassDeclaration` fires before the returns in its body and `Program:exit` runs
// after everything.
func (collector *renderReturnCollector) walk(node *ast.Node) {
	if node == nil {
		return
	}

	if isRenderReturnComponent(node) {
		component := &renderReturnComponent{
			node:           node,
			renderProperty: findRenderProperty(node),
		}
		collector.components = append(collector.components, component)
		if collector.byNode == nil {
			collector.byNode = map[*ast.Node]*renderReturnComponent{}
		}
		collector.byNode[node] = component
	}

	if node.Kind == ast.KindReturnStatement {
		collector.markEnclosingComponent(node.Parent)
	}

	// An arrow with an expression body returns without writing a `ReturnStatement`, so nothing in
	// the walk above can see it. Upstream carries a second listener for exactly this and it is not
	// a shortcut for the first: `render = () => <div/>` is clean and has no return node anywhere.
	//
	// Upstream's `ArrowFunctionExpression` listener tests only that the arrow has an expression
	// body and that its parent property is named `render`, then marks unconditionally, without
	// consulting the scope count at all. This arm needs no separate flag to reproduce that: the
	// climb starts at the arrow's own parent, which `isRenderReturnImplicitReturnArrow` has
	// already established IS the `render` property, so the shared count is zero when the property
	// test runs and the walk satisfies it on the first step every time.
	//
	// That was measured rather than argued. Seeding the flag true for this arm was in the first
	// draft, and a mutation forcing the seed false was a SURVIVOR: the seed is subsumed by the
	// property check and no input can separate the two, because the arm's own precondition is what
	// guarantees the property is the first node the climb sees. Removed rather than kept with a
	// fixture, since no fixture could have covered it.
	if node.Kind == ast.KindArrowFunction && isRenderReturnImplicitReturnArrow(node) {
		collector.markEnclosingComponent(node.Parent)
	}

	node.ForEachChild(func(child *ast.Node) bool {
		collector.walk(child)
		return false
	})
}

// markEnclosingComponent reproduces upstream's `components.set`, which is the whole subtlety.
//
// The walk climbs to the nearest registered component and marks it, so a return inside a nested
// class that is NOT a component marks the outer one. See the rule doc: that is a measured false
// negative in upstream and it is reproduced here rather than corrected.
//
// The scope count is applied on the way up. A return only counts if the property named `render` is
// reached having crossed at most one function-like scope, which is upstream's `depth <= 1`. The
// climb continues past the property either way, because the component still has to be found even
// when the return does not satisfy it, and because upstream's `forEach` over the ancestors likewise
// does not stop at the first match.
func (collector *renderReturnCollector) markEnclosingComponent(from *ast.Node) {
	functionScopeDepth := 0
	satisfiesRender := false

	for current := from; current != nil; current = current.Parent {
		if isRenderReturnFunctionScope(current) {
			functionScopeDepth++
		}
		if isRenderNamedProperty(current) {
			// A class method or accessor is ONE node here and TWO upstream, so the function scope
			// it owns is counted at the property rather than by the walk. See the rule doc: this
			// is the same correction the sibling rules make and it does move the verdict here.
			depthAtProperty := functionScopeDepth
			switch current.Kind {
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				depthAtProperty++
			}
			if depthAtProperty <= 1 {
				satisfiesRender = true
			}
		}
		if component, found := collector.byNode[current]; found {
			if satisfiesRender {
				component.hasReturnStatement = true
			}
			return
		}
	}
}

// isRenderReturnComponent reports whether a node is a component by either of the two definitions
// the rule's own filter accepts.
//
// This is upstream's `isES5Component(node) || isES6Component(node)` from the `Program:exit` filter,
// which is what narrows `Components.detect`'s much wider registration back down. The ES5 arm anchors
// on the object literal rather than on the call, matching upstream, whose `isES5Component` reads
// `node.parent.callee` off the object.
func isRenderReturnComponent(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if isRenderReturnEs6ComponentClass(node) {
		return true
	}
	// The ES5 component IS the object literal, and the factory call is its parent. Upstream keys
	// the component list on the object for the same reason: `getComponentProperties` reads
	// `node.properties`, which only an object literal has.
	if node.Kind == ast.KindObjectLiteralExpression {
		return isEs5ComponentCallStrict(node.Parent)
	}
	return false
}

// isRenderReturnEs6ComponentClass is the shelf's `IsEs6ComponentClass` with the parentheses
// stripped from the whole heritage expression first.
//
// The shelf helper is correct for the rules that call it and wrong for this one, in a way neither
// its doc comment nor any corpus can show. It skips parentheses on the RECEIVER, so
// `extends (React).Component` answers true, and it does not skip them around the WHOLE expression,
// so `extends (React.Component)` answers false. Measured by walking the parse: the outer form
// produces a `KindParenthesizedExpression` between the `KindExpressionWithTypeArguments` and the
// property access, and the helper's kind switch declines it.
//
// Upstream reports on both. The reason is a parser difference rather than a judgment: espree
// **does not produce a parenthesized-expression node at all**, so `node.superClass` is the bare
// `MemberExpression` whichever way the parens are written, and `isES6Component` never sees them.
// Measured on the installed build across four spellings, `(React.Component)`, `(React).Component`,
// `((React.Component))` and `(Component)`, and all four report.
//
// So this is fidelity to a parser difference, in the direction the brief warns about: the shelf
// doing something sensible that upstream does not do is still a defect in a port, and here the
// shelf is narrower than upstream rather than wider. Narrowed locally rather than changed on the
// shelf, because the three rules that call it were ported against oxc, whose tree DOES carry a
// parenthesized node, and the shelf's doc records that measurement correctly for them.
func isRenderReturnEs6ComponentClass(node *ast.Node) bool {
	if node == nil {
		return false
	}
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
		// `implements` is a `KindTypeReference` in this tree rather than a
		// `KindExpressionWithTypeArguments`, so the kind test inside the shelf helper already
		// declines it. Reproduced here by delegating the base-name decision back to the shelf.
		types := clause.AsHeritageClause().Types
		if types == nil {
			continue
		}
		for _, typeNode := range types.Nodes {
			if typeNode.Kind != ast.KindExpressionWithTypeArguments {
				continue
			}
			expression := ast.SkipParentheses(typeNode.AsExpressionWithTypeArguments().Expression)
			if expression == nil {
				continue
			}
			// The base-name decision itself is the shelf's, reached by handing it a synthetic
			// class whose heritage is the unwrapped expression. Rather than restate
			// `isComponentBase`, which is unexported there, the unwrapped expression is compared
			// through the shelf's own namespaced-member predicate plus the bare-name arm.
			if isRenderReturnComponentBase(expression) {
				return true
			}
		}
	}
	return false
}

// isRenderReturnComponentBase reports whether an unwrapped extends target names a React base.
//
// This is the shelf's unexported `isComponentBase` at the two spellings upstream accepts:
// `React.Component` and `React.PureComponent` through the namespace, and the bare `Component` and
// `PureComponent` a file importing them directly writes. Upstream's test is
// `/^(Pure)?Component$/` against the property or the identifier, with the object required to equal
// the pragma, which defaults to `React`.
//
// The namespaced arm goes through the shelf's `IsNamespacedMember` so the receiver's own
// parenthesis handling stays in one place rather than being restated a fifth time.
func isRenderReturnComponentBase(expression *ast.Node) bool {
	if expression == nil {
		return false
	}
	if expression.Kind == ast.KindIdentifier {
		return isRenderReturnComponentBaseName(expression.Text())
	}
	return utilsreact.IsNamespacedMember(expression, isRenderReturnComponentBaseName)
}

// isRenderReturnComponentBaseName is upstream's `/^(Pure)?Component$/`.
func isRenderReturnComponentBaseName(name string) bool {
	return name == "Component" || name == "PureComponent"
}

// findRenderProperty returns the property node holding a function-like `render`, or nil.
//
// Upstream's `findRenderMethod` filters the component's properties down to those whose name is
// `render` and whose `value` is a `FunctionExpression` or an `ArrowFunctionExpression`, then takes
// the FIRST such property. Two parts of that are load-bearing and each is measured:
//
// The value must be function-like. `render = 5`, `render = someFn`, a bare `render` class field
// with no initializer, and the shorthand `{ render }` are all clean upstream, and the last two are
// clean cases upstream ships. A port testing only the name reports all four.
//
// The first match wins, which decides where a class with two `render` methods points. Measured:
// with both empty, the finding is on the first.
//
// A method declaration is function-like by construction and has no separate value node, so it
// matches on its own. `get render()` and `set render()` are `MethodDefinition` upstream, where
// `value` is a `FunctionExpression`, so they match there too and are measured to report; our tree
// splits them into their own kinds and they are matched here explicitly for that reason.
func findRenderProperty(component *ast.Node) *ast.Node {
	var found *ast.Node
	membersOfRenderReturnComponent(component, func(member *ast.Node) {
		if found != nil {
			return
		}
		if !isRenderNamedProperty(member) {
			return
		}
		if !hasFunctionLikeRenderValue(member) {
			return
		}
		found = member
	})
	return found
}

// membersOfRenderReturnComponent visits the direct members of a component, and nothing deeper.
//
// This is upstream's `getComponentProperties`, which returns `node.body.body` for a class and
// `node.properties` for an object literal and an empty list for anything else. Depth matters: a
// `render` on an object nested inside the component is not the component's `render`, and measured,
// `createReactClass({ foo: { render: function() {} } })` is clean.
func membersOfRenderReturnComponent(component *ast.Node, visit func(member *ast.Node)) {
	if component == nil {
		return
	}
	switch component.Kind {
	case ast.KindClassDeclaration:
		for _, member := range component.AsClassDeclaration().Members.Nodes {
			visit(member)
		}
	case ast.KindClassExpression:
		for _, member := range component.AsClassExpression().Members.Nodes {
			visit(member)
		}
	case ast.KindObjectLiteralExpression:
		for _, property := range component.AsObjectLiteralExpression().Properties.Nodes {
			visit(property)
		}
	}
}

// isRenderNamedProperty reports whether a node is a property whose name is exactly `render`.
//
// The kind set is upstream's `/(MethodDefinition|Property|ClassProperty|PropertyDefinition)$/` at
// our spelling, plus the two accessor kinds our tree splits out of `MethodDefinition`. A shorthand
// property assignment is deliberately included: it is a `Property` upstream, so it reaches the name
// test there and is then rejected for having no function-like value, and reproducing that split
// keeps the two decisions where upstream puts them.
//
// The name is read by kind rather than through `Text()`, which panics outright on a computed
// property name. That matters for fidelity as well as for safety: upstream's `getPropertyName`
// returns `nameNode.name`, which is `undefined` for both a string-literal key and a computed one,
// so `'render'() {}` and `['render']() {}` are BOTH clean upstream. Measured on all four spellings
// across the class and object forms, and all four are clean. So this compares only the identifier
// spelling, and `ast.TryGetTextOfPropertyName` is deliberately NOT used here even though it is what
// the sibling rules reach for: it resolves the string-literal and computed-holding-a-literal forms,
// which would report on four inputs upstream is silent on.
func isRenderNamedProperty(node *ast.Node) bool {
	if node == nil {
		return false
	}
	var name *ast.Node
	switch node.Kind {
	case ast.KindMethodDeclaration:
		name = node.AsMethodDeclaration().Name()
	case ast.KindPropertyDeclaration:
		name = node.AsPropertyDeclaration().Name()
	case ast.KindPropertyAssignment:
		name = node.AsPropertyAssignment().Name()
	case ast.KindShorthandPropertyAssignment:
		name = node.AsShorthandPropertyAssignment().Name()
	case ast.KindGetAccessor:
		name = node.AsGetAccessorDeclaration().Name()
	case ast.KindSetAccessor:
		name = node.AsSetAccessorDeclaration().Name()
	default:
		return false
	}
	// Identifier only. A string-literal or computed key answers `undefined` at upstream's
	// `nameNode.name` and is clean there; see the doc above for the measurement.
	return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "render"
}

// hasFunctionLikeRenderValue reports whether a `render` property carries something callable.
//
// Upstream's test is `astUtil.isFunctionLikeExpression(property.value)`, which accepts exactly
// `FunctionExpression` and `ArrowFunctionExpression`. A `MethodDefinition` passes because its
// `value` IS a `FunctionExpression`; our `KindMethodDeclaration` fuses the two and so answers true
// on its own, and the accessors do the same for the same reason.
//
// A property whose initializer is missing, a number, or an identifier is rejected, which is two of
// upstream's clean cases and two more measured beside them.
func hasFunctionLikeRenderValue(property *ast.Node) bool {
	switch property.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		// The method IS the function here. Upstream reaches through `value` to a separate node.
		return true
	case ast.KindPropertyDeclaration:
		return isRenderReturnFunctionLikeExpression(property.AsPropertyDeclaration().Initializer)
	case ast.KindPropertyAssignment:
		return isRenderReturnFunctionLikeExpression(property.AsPropertyAssignment().Initializer)
	case ast.KindShorthandPropertyAssignment:
		// `{ render }` has no value node at all, so upstream's `property.value` is the shorthand's
		// own identifier, which is not function-like. Clean, and it is upstream's twelfth valid
		// case.
		return false
	}
	return false
}

// isRenderReturnFunctionLikeExpression is upstream's `isFunctionLikeExpression`, exactly.
//
// Two kinds and no more. A `FunctionDeclaration` is absent because it cannot be a property value,
// and parentheses are NOT skipped, matching upstream, whose test is a bare `node.type` comparison
// on the value as written.
func isRenderReturnFunctionLikeExpression(value *ast.Node) bool {
	if value == nil {
		return false
	}
	return value.Kind == ast.KindFunctionExpression || value.Kind == ast.KindArrowFunction
}

// isRenderReturnImplicitReturnArrow reports whether an arrow returns without a return statement,
// as the value of a property named `render`.
//
// This is upstream's whole second listener: `if (node.expression === false ||
// astUtil.getPropertyName(node.parent) !== 'render') return;`. Both halves matter.
//
// `node.expression === false` means the arrow has a BLOCK body, and those are handled by the return
// walk instead. Our tree spells the same distinction as the body's kind: a block body is
// `KindBlock` and anything else is the expression. So `render = () => {}` falls through to the walk
// and reports, while `render = () => <div/>` is caught here and is clean, both measured.
//
// The parent test is `getPropertyName`, which is the same identifier-only comparison
// `isRenderNamedProperty` makes, so it is reused rather than restated. Note what upstream does NOT
// test here: it never asks whether the property is inside a component. A `render` arrow on a plain
// object literal marks whatever component encloses that object, which is measured and is the same
// looseness the nested-class case shows.
func isRenderReturnImplicitReturnArrow(node *ast.Node) bool {
	arrow := node.AsArrowFunction()
	if arrow == nil {
		return false
	}
	body := arrow.Body
	if body == nil || body.Kind == ast.KindBlock {
		return false
	}
	return isRenderNamedProperty(node.Parent)
}

// isRenderReturnFunctionScope reports whether an ancestor costs a scope on upstream's count.
//
// Upstream tests `/Function(Expression|Declaration)$/` against the node type, and the pattern is
// unanchored at the front, so it matches `FunctionExpression`, `FunctionDeclaration` AND
// `ArrowFunctionExpression`. That last one is the trap: an arrow reads as transparent and is not.
// Measured, `render() { var f = () => { return 1; }; }` reports.
//
// A method declaration is absent, and its absence is exact rather than an oversight. Upstream's
// `MethodDefinition` is not itself function-like and never matches the pattern; the
// `FunctionExpression` it wraps is what gets counted, and our fused node has no such inner scope
// to count. See the rule doc for why the resulting off-by-one cannot move a verdict at this
// threshold. The accessors are absent for the same reason.
func isRenderReturnFunctionScope(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindArrowFunction:
		return true
	}
	return false
}
