package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoSetStateInComponentDidUpdate = rule.Message{
	Id: "noSetStateInComponentDidUpdate",
	Description: "Updating state from `componentDidUpdate` triggers a second render pass before " +
		"the browser has painted the first one, so every update costs two renders and any layout " +
		"read between them measures a frame that is about to be thrown away. Derive the value " +
		"during render instead, or guard the update on a comparison of the previous props so it " +
		"cannot run on its own output.",
}

// NoDidUpdateSetState reports `this.setState` called from `componentDidUpdate`.
//
//	valid:   class H extends React.Component { componentDidMount() { this.setState({}); } }
//	valid:   class H extends React.Component { componentDidUpdate() { on(() => this.setState({})); } }
//	valid:   class H { componentDidUpdate() { this.setState({}); } }          (no component)
//	invalid: class H extends React.Component { componentDidUpdate() { this.setState({}); } }
//	invalid: var H = createReactClass({ componentDidUpdate: function() { this.setState({}); } });
//	invalid: the callback case above, under `disallow-in-func`
//
// Ported from `react/no-did-update-set-state`, read against oxc's `no_did_update_set_state.rs` and
// the shared `function_count_before_lifecycle_component` it calls, against `eslint-plugin-react`'s
// `makeNoMethodSetStateRule` factory, and measured on the release oxlint binary for every shape our
// abstract syntax tree spells differently than oxc's does.
//
// # What the rule actually decides, which is narrower than its name and wider than it reads
//
// Three conditions, all syntactic. There is no call-graph analysis anywhere in either upstream, and
// that is the thing most likely to be assumed: a helper method that calls `setState`, invoked from
// the lifecycle method, is silent at both tools and is silent here. Only ancestry counts.
//
//	the callee is `this.setState`, by member access of either spelling
//	walking outward reaches a property or method named `componentDidUpdate`
//	and continuing outward from there reaches a React component
//
// The nesting judgment sits on top: the walk counts function-like ancestors passed before reaching
// the lifecycle name, and by default declines when that count exceeds one. A call written directly
// in the method body passes one function, which is the method itself. A call in a callback passes
// two. `disallow-in-func` drops the count check and reports at any depth.
//
// # The two conditions being independent is upstream behavior, and it is the surprise here
//
// Nothing checks that the lifecycle-named property and the component are related. A plain object
// literal carrying a `componentDidUpdate` property, declared anywhere inside any component, reports
// as though it were that component's own lifecycle method. Measured on the release binary rather
// than inferred, because the loop reads as though it were checking a component's method and is not:
// it latches on a name match and then returns at the first component above it, however far away and
// whatever lies between.
//
// The inverse also holds and is the difference from ESLint. ESLint's factory matches the
// lifecycle-named property and stops, so it reports on a class extending nothing and on a bare
// object literal at the top level of a file. oxc requires the component and is silent on both.
// Reproduced oxc's answer, because oxc is what the differential gate compares against, and the
// three clean fixtures covering it are ours rather than upstream's.
//
// # Why the ES5 component predicate is spelled here rather than taken from the shelf
//
// `utilsreact.IsEs5ComponentCall` looks like exactly this question and answers a wider one. It
// accepts `createClass` and `React.createClass` as well as `createReactClass`, matching a different
// oxc helper, while `is_es5_component`, which is the one this rule reaches, accepts only the name
// `createReactClass`. Measured: both of the wider spellings are clean on the release binary, so
// reaching for the shelf would have reported where upstream is silent, and no imported fixture
// covers either spelling so nothing would have caught it.
//
// That is the third time in this package a react helper has been correct for the rule that
// motivated it and wrong for a neighbour that asks a question of the same shape. The ES6 half does
// match, and `utilsreact.IsEs6ComponentClass` is used for it rather than copied.
//
// # Reading a property key cannot go through Text()
//
// A computed key is a `ComputedPropertyName` and `Node.Text()` panics on it outright, so the key is
// resolved with `ast.TryGetTextOfPropertyName`, which checks kind before reading and answers a
// boolean for the keys that name nothing statically. That is the shim's spelling of oxc's
// `static_name()` and it resolves the same three shapes: an identifier, a string literal, and a
// computed key holding one. All three measured reporting on the release binary, so a port matching
// only the identifier spelling loses two real findings and crashes on a third input.
var NoDidUpdateSetState = rule.Rule{
	// No namespace prefix. The config writes `react/no-did-update-set-state` and the parity guard
	// strips the namespace on a `/` boundary.
	Name: "react/no-did-update-set-state",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// An unconfigured rule gets the zero value, which is upstream's permissive default.
		settings, _ := rule.OptionsAs[NoMethodSetStateOptions](options)
		reportInsideNestedFunctions := settings.DisallowInFunc

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// No parenthesis skipping, and that is measured rather than an omission.
				// `(this.setState)({...})` is silent on the release binary, because oxc reaches
				// `callee.as_member_expression()` which does not see through a parenthesized
				// expression. Skipping here would report where upstream does not, and no imported
				// fixture writes the shape.
				callee := node.AsCallExpression().Expression
				if !didUpdateIsThisSetStateCallee(callee) {
					return
				}

				functionCount, insideLifecycle := didUpdateFunctionCountBeforeLifecycle(node,
					"componentDidUpdate")
				if !insideLifecycle {
					return
				}
				if functionCount > 1 && !reportInsideNestedFunctions {
					return
				}

				// The callee, not the call. Upstream underlines thirteen characters for
				// `this.setState({...})`, stopping before the parenthesis, on all eighteen of its
				// snapshot diagnostics.
				ctx.ReportNode(callee, messageNoSetStateInComponentDidUpdate)
			},
		}
	},
}

// didUpdateIsThisSetStateCallee reports whether an expression is `this.setState` by either spelling.
//
// oxc reads one `MemberExpression` and asks two questions of it: that the object is a `this`
// expression, and that `static_property_name()` answers `setState`. Our tree splits that into a
// dotted `PropertyAccessExpression` and a bracketed `ElementAccessExpression`, so both are matched
// here and the bracketed one re-resolves the static name. `this["setState"]` reports upstream and
// was measured doing so.
//
// Parentheses are skipped nowhere in this function, and the doc on `didUpdateIsBareThisReceiver` carries the
// measurement: both `(this).setState({})` and `(this.setState)({})` are silent upstream.
//
// The optional spellings need no special handling and were measured rather than assumed:
// `this?.setState(...)` is a property access carrying a question-dot token and reaches this
// function unchanged, and `this.setState?.(...)` is an ordinary access under a call marked
// optional. Both report on the release binary and both report here.
func didUpdateIsThisSetStateCallee(callee *ast.Node) bool {
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if !didUpdateIsBareThisReceiver(access.Expression) {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "setState"

	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if !didUpdateIsBareThisReceiver(access.Expression) {
			return false
		}
		argument := access.ArgumentExpression
		// Narrower than the property-key resolver deliberately: oxc's
		// `ComputedMemberExpression::static_property_name` answers for a string literal and a
		// single-quasi template, and the result is compared against one fixed name, so a numeric
		// key could not change a verdict here.
		if argument == nil {
			return false
		}
		switch argument.Kind {
		case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
			return argument.Text() == "setState"
		}
	}
	return false
}

// didUpdateIsBareThisReceiver reports whether an expression is `this` written without parentheses.
//
// Parentheses are deliberately NOT skipped, which is the opposite of what the rest of this package
// does and is measured rather than reasoned about: `(this).setState({})` inside a real
// `componentDidUpdate` is silent on the release oxlint binary. oxc destructures
// `Expression::ThisExpression` directly, so a parenthesized receiver is a different variant and
// falls through. Skipping would have been the natural thing to write, it reads as an improvement,
// and it would have reported where upstream says nothing with no imported fixture able to see it.
//
// The same file's `didUpdateIsCreateReactClassCall` does skip parentheses on its callee, and that asymmetry
// is upstream's rather than an inconsistency here: `(createReactClass)({...})` reports on the same
// binary in the same run, because oxc reaches that one through `get_identifier_reference()`, which
// does look through parentheses. Both halves were measured in one invocation.
func didUpdateIsBareThisReceiver(expression *ast.Node) bool {
	return expression != nil && expression.Kind == ast.KindThisKeyword
}

// didUpdateFunctionCountBeforeLifecycle walks outward and answers how deeply a node is nested.
//
// This is oxc's shared `function_count_before_lifecycle_component` at our tree's spelling, and the
// two halves of its answer are what the rule decides on: how many function-like constructs were
// passed before a property named for the lifecycle method was reached, and whether a React
// component was reached after that. The boolean is false when either never happens, which is one
// return value in oxc's `Option<usize>` and two here because Go has no such shape without a pointer.
//
// # Three details that each look like a detail and are not
//
// **Counting stops at the first name match.** Once the lifecycle name is seen, further functions on
// the way out are not counted. A lifecycle-named property nested inside the real lifecycle method
// therefore has a count of one and reports in the default mode, which was measured rather than
// reasoned about because continuing to count would have made it two and silent.
//
// **The count includes the lifecycle method's own function.** A call written directly in the body
// passes exactly one, which is why the threshold is `> 1` rather than `> 0`. That holds across all
// three ways the method can be written: a `MethodDeclaration`, an object property whose value is a
// function expression, and a class property whose value is an arrow. In the last of those, the
// arrow is passed before the property that names it, which is what keeps the three consistent.
//
// **The name is read with the kind-checked resolver.** `ast.TryGetTextOfPropertyName` answers for an
// identifier, a string literal and a computed key holding a literal, and answers false rather than
// panicking on an interpolated computed key. Reading `Text()` here crashes the linter on the first
// `{[k]: v}` in the tree.
//
// Accessors are included because oxc's `is_lifecycle_component_method` matches its `MethodDefinition`
// without asking what kind of method it is, and a getter is one there. Measured: a getter named
// `componentDidUpdate` reports on the release binary.
func didUpdateFunctionCountBeforeLifecycle(node *ast.Node, lifecycleMethodName string) (int, bool) {
	functionCount := 0
	insideLifecycle := false

	for current := node.Parent; current != nil; current = current.Parent {
		if !insideLifecycle {
			if didUpdateIsLifecycleNamedMember(current, lifecycleMethodName) {
				insideLifecycle = true
				// The method's own function scope is counted here rather than by the walk, and
				// this is the one place our tree's shape forces a different spelling rather than a
				// different decision. oxc has two nodes where we have one: a `MethodDefinition`
				// carrying a `Function` value, so its walk passes the function first and latches on
				// the definition second, counting both. A `MethodDeclaration` here is a single node
				// that latches, so without this the count comes out one short for every class
				// component and the rule reports on every callback the corpus says is clean.
				//
				// Measured rather than reasoned about: this was written the tidy way first and four
				// imported clean cases went red, all of them arrow callbacks inside class methods.
				// An accessor is the same shape and is counted the same way.
				if didUpdateIsMethodLike(current) {
					functionCount++
				}
			} else if didUpdateIsCountedFunction(current) {
				functionCount++
			}
		}

		if insideLifecycle && didUpdateIsReactComponent(current) {
			return functionCount, true
		}
	}
	return 0, false
}

// didUpdateIsCountedFunction reports whether a node is one of the constructs oxc counts toward nesting depth.
//
// Exactly three kinds, matching oxc's `AstKind::Function(_) | AstKind::ArrowFunctionExpression(_)`,
// where its `Function` is one node for a declaration and an expression alike. A method and an
// accessor are deliberately absent here and counted at the latch instead, because in our tree they
// are the same node that carries the lifecycle name; see the walk above.
//
// This is narrower than `scope.EnclosingFunctionLike`, which is why that shelf helper is not used
// despite its name answering what looks like the same question. It includes methods, accessors and
// constructors unconditionally, and it returns the nearest enclosing construct rather than counting
// every one on the way out, so it answers neither half of what this needs. Read the body, not the
// name.
func didUpdateIsCountedFunction(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	}
	return false
}

// didUpdateIsMethodLike reports whether a node carries its own function scope as one node.
//
// These are the kinds that latch the lifecycle name AND introduce a function, which in oxc are two
// nodes and here are one. A class property holding an arrow is absent on purpose: there the arrow is
// a separate node passed by the walk before the property that names it, so counting the property
// would double it. The corpus covers that shape directly, with `componentDidUpdate = () => {}` as a
// fail case that must report at depth one.
func didUpdateIsMethodLike(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return true
	}
	return false
}

// didUpdateIsLifecycleNamedMember reports whether a node is a member whose key names the lifecycle method.
//
// The three kinds are oxc's `ObjectProperty`, `MethodDefinition` and `PropertyDefinition` at our
// spelling, plus the two accessors, which oxc reaches through `MethodDefinition` because it does not
// separate them by node kind. Nothing here asks whether the member is static or on a prototype:
// upstream does not, and a `static componentDidUpdate` reports on the release binary even though it
// is not a lifecycle method at all. Reproduced rather than corrected, because a silent improvement
// is a divergence nothing states.
func didUpdateIsLifecycleNamedMember(node *ast.Node, lifecycleMethodName string) bool {
	var name *ast.Node
	switch node.Kind {
	case ast.KindPropertyAssignment:
		name = node.AsPropertyAssignment().Name()
	case ast.KindMethodDeclaration:
		name = node.AsMethodDeclaration().Name()
	case ast.KindPropertyDeclaration:
		name = node.AsPropertyDeclaration().Name()
	case ast.KindGetAccessor:
		name = node.AsGetAccessorDeclaration().Name()
	case ast.KindSetAccessor:
		name = node.AsSetAccessorDeclaration().Name()
	default:
		return false
	}
	if name == nil {
		return false
	}
	text, named := ast.TryGetTextOfPropertyName(name)
	return named && text == lifecycleMethodName
}

// didUpdateIsReactComponent reports whether a node is a component by either era's definition.
//
// The ES6 half is `utilsreact.IsEs6ComponentClass`, which matches what this rule's upstream asks:
// a class expression or declaration extending `React.Component`, `React.PureComponent`, or the bare
// `Component` and `PureComponent` a file importing them directly writes.
//
// The ES5 half is spelled here rather than taken from `utilsreact.IsEs5ComponentCall`, and the type
// doc above says why: that helper also accepts `createClass` and `React.createClass`, which are
// clean at this rule's upstream. The difference is measured, not read.
func didUpdateIsReactComponent(node *ast.Node) bool {
	if utilsreact.IsEs6ComponentClass(node) {
		return true
	}
	return didUpdateIsCreateReactClassCall(node)
}

// didUpdateIsCreateReactClassCall reports whether a call is `createReactClass(...)` or the namespaced form.
//
// Only that one name, matching `is_es5_component`. `React.createReactClass(...)` counts because oxc
// accepts the namespaced spelling of the same name, though nothing in either corpus writes it.
func didUpdateIsCreateReactClassCall(node *ast.Node) bool {
	if node.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == "createReactClass"

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		object := ast.SkipParentheses(access.Expression)
		if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "React" {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "createReactClass"
	}
	return false
}
