package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUseSetStateCallback = rule.Message{
	Id: "useCallback",
	Description: "Use callback in setState when referencing the previous state. React batches " +
		"state updates, so `this.state` read while building the next state is whatever was there " +
		"before the pending updates applied, and two updates in one batch will read the same " +
		"stale value. The callback form is handed the state React is about to apply.",
}

// NoAccessStateInSetstate reports reading `this.state` while computing the argument to `setState`.
//
//	valid:   this.setState(state => ({value: state.value + 1}))
//	valid:   this.setState({}, () => console.log(this.state))
//	valid:   var v = this.state.value; this.multiplyValue({value: v})
//	invalid: this.setState({value: this.state.value + 1})
//	invalid: var v = this.state.value + 1; this.setState({value: v})
//	invalid: this.setState(this.state, () => console.log(this.state))
//
// # Three routes to a finding, and they are three separate visitors upstream
//
// Direct, through a variable, and through a method. Each accumulates into shared state as the walk
// proceeds, so the rule's answer depends on walk order, and the order is reproduced rather than
// improved on.
//
//	this.setState({value: this.state.value})       direct, the member expression is inside the argument
//	var v = this.state.x; this.setState({value: v}) a variable holding this.state, used in the argument
//	nextState() where nextState reads this.state    a method reading this.state, called in the argument
//
// # Only the FIRST argument counts, and this is the whole valid/invalid split
//
// `this.setState({}, () => console.log(this.state))` is clean and
// `this.setState(this.state, () => console.log(this.state))` reports once. The second argument is
// the completion callback, which runs after the update and is the correct place to read state.
// Upstream walks up from the finding to the `setState` call and asks whether the ancestor it came
// through IS `arguments[0]`, which is what makes the two cases differ. Both measured.
//
// A callback FIRST argument is not exempt: `this.setState(() => ({value: this.state.value + 1}))`
// reports, because the callback receives the pending state as a parameter and reading `this.state`
// inside it defeats the point. Measured against the installed build.
//
// # The scope comparison on the variable route, which is stricter than it reads
//
// Upstream keeps each recorded variable's SCOPE and reports only when the use site's scope is the
// same object. That is block-level identity rather than resolution, measured three ways on the
// installed build:
//
//	var v = this.state.x; this.setState({value: v})                        reports
//	var v = this.state.x; this.setState(function(){return {value: v}})     SILENT, different scope
//	{ var v = this.state.x; } this.setState({value: v})                    SILENT, different scope
//
// The third is the surprising one, because `var` hoists to the function and the name genuinely
// resolves. Upstream is comparing where each site SITS rather than what the name binds to, so a
// block moves the answer. Reproduced rather than corrected: the direction is a false negative,
// which is visible to nobody but also harms nobody, and correcting it would report code upstream
// calls clean.
//
// # Which components count
//
// A class extending a React base, or an object literal handed to the factory. The factory spelling
// is `createReactClass` alone. `React.createClass` is silent unless `settings.react.createClass` is
// configured to it, which upstream's own test harness does not do, and cohere reads no React
// settings. Measured: `React.createClass` is silent on the installed build and
// `createReactClass` reports on byte-identical bodies. Six of upstream's nine failing cases are
// written in the `React.createClass` spelling and are therefore dead against the version this
// repository runs; they are recorded below in the reachable spelling, with every verdict replayed.
//
// # The method route matches on a BARE name
//
// Upstream's corpus reports `this.setState(nextState())` for a component whose method is
// `nextState`, even though the call is unqualified and would be a reference error at runtime. The
// recorded method name is compared against the callee's own name with no receiver check, so
// `this.nextState()` and a free `nextState()` are the same input to this rule. Reproduced, because
// narrowing it would go silent on upstream's own case.
var NoAccessStateInSetstate = rule.Rule{
	Name: "react/no-access-state-in-setstate",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// Both lists accumulate as the walk proceeds and are read by later visits, which is
		// upstream's design. A method or variable declared AFTER the `setState` that would have
		// used it is therefore not seen, and that ordering is part of the behaviour.
		methods := []noAccessStateMethod{}
		variables := []noAccessStateVariable{}

		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				// One walk in source order, dispatching to the four arms upstream registers. A
				// single walk is used rather than four listeners because the arms read each
				// other's accumulated state and the order between them is load-bearing: ESLint
				// interleaves its visitors in one traversal, so a member expression inside a call
				// is visited after the call that contains it.
				var visit func(*ast.Node)
				visit = func(node *ast.Node) {
					switch node.Kind {
					case ast.KindCallExpression:
						noAccessStateVisitCall(ctx, node, &methods)

					case ast.KindPropertyAccessExpression:
						noAccessStateVisitStateAccess(ctx, node, &methods, &variables)

					case ast.KindIdentifier:
						noAccessStateVisitIdentifier(ctx, node, variables)

					case ast.KindObjectBindingPattern:
						noAccessStateVisitObjectPattern(node, &variables)
					}

					node.ForEachChild(func(child *ast.Node) bool {
						visit(child)
						return false
					})
				}
				visit(sourceFile)
			},
		}
	},
}

// noAccessStateMethod is a component method known to read `this.state`, with the node to report.
//
// The reported node is the `this.state` access itself rather than the method, so a method used from
// two `setState` calls produces two findings both pointing at the same line. That is upstream's
// bookkeeping and it is reproduced.
type noAccessStateMethod struct {
	name string
	node *ast.Node
}

// noAccessStateVariable is a local holding `this.state`, with the scope it was declared in.
type noAccessStateVariable struct {
	name  string
	node  *ast.Node
	scope *ast.Node
}

// noAccessStateVisitCall handles upstream's CallExpression arm.
//
// Two jobs. It propagates the "reads this.state" mark from a called method up to the method doing
// the calling, and it reports when a marked method is called inside a `setState` first argument.
func noAccessStateVisitCall(ctx rule.Context, node *ast.Node, methods *[]noAccessStateMethod) {
	if enclosingComponentOf(node) == nil {
		return
	}
	callee := node.AsCallExpression().Expression
	calleeName := noAccessStateCalleeName(callee)

	// Propagation. A method calling a marked method becomes marked itself, carrying the ORIGINAL
	// method's reported node forward, so a two-hop chain still points at the `this.state` access.
	//
	// Upstream appends to the same slice it is iterating, which in JavaScript means the new entries
	// are visited by the same forEach. Go's range evaluates the length once, so the appended
	// entries are not revisited in this pass. That difference can only matter for a chain of three
	// or more methods discovered in one visit, which no corpus case writes; it is recorded here
	// rather than silently diverged, and a deeper chain would need a fixed point.
	if calleeName != "" {
		existing := len(*methods)
		for index := 0; index < existing; index++ {
			if (*methods)[index].name != calleeName {
				continue
			}
			if enclosing := noAccessStateEnclosingMethodName(node); enclosing != "" {
				*methods = append(*methods, noAccessStateMethod{
					name: enclosing,
					node: (*methods)[index].node,
				})
			}
		}
	}

	// Reporting. Walk up looking for a `setState` call this node sits inside the first argument of.
	for current := node.Parent; current != nil; current = current.Parent {
		if noAccessStateIsFirstArgumentOfSetState(current, node) {
			for _, method := range *methods {
				if method.name == calleeName {
					ctx.ReportNode(method.node, messageUseSetStateCallback)
				}
			}
			break
		}
	}
}

// noAccessStateVisitStateAccess handles upstream's MemberExpression arm, the `this.state` anchor.
//
// One walk upward from the access, taking the first of four exits: report, record as a method,
// record as a variable, or fall off the top.
func noAccessStateVisitStateAccess(
	ctx rule.Context,
	node *ast.Node,
	methods *[]noAccessStateMethod,
	variables *[]noAccessStateVariable,
) {
	access := node.AsPropertyAccessExpression()
	if access.Expression == nil || access.Expression.Kind != ast.KindThisKeyword {
		return
	}
	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "state" {
		return
	}
	if enclosingComponentOf(node) == nil {
		return
	}

	for current := node; current != nil; current = current.Parent {
		if noAccessStateIsFirstArgumentOfSetState(current, node) {
			ctx.ReportNode(node, messageUseSetStateCallback)
			return
		}

		// A method declaration, or a function expression sitting under a named property, records
		// the enclosing method as one that reads state. Upstream splits these into two arms because
		// ESTree spells a class method and an object method differently; ours has a method
		// declaration kind for both, plus the property-assigned function expression.
		if methodName, isMethod := noAccessStateDeclaringMethodName(current); isMethod {
			*methods = append(*methods, noAccessStateMethod{name: methodName, node: node})
			return
		}

		if current.Kind == ast.KindVariableDeclaration {
			declaration := current.AsVariableDeclaration()
			if declaration.Name() != nil && declaration.Name().Kind == ast.KindIdentifier {
				*variables = append(*variables, noAccessStateVariable{
					name:  declaration.Name().Text(),
					node:  node,
					scope: noAccessStateScopeOf(node),
				})
			}
			return
		}
	}
}

// noAccessStateVisitIdentifier handles upstream's Identifier arm, the variable route's use site.
//
// Upstream first walks past any enclosing binary expressions, then requires the identifier to be
// either the VALUE of a property or the OBJECT of a member access. That pair is what keeps a
// property KEY and a bare argument from matching: `{value: v}` matches on `v` as the value, and
// `v.value` matches on `v` as the object.
func noAccessStateVisitIdentifier(ctx rule.Context, node *ast.Node, variables []noAccessStateVariable) {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindBinaryExpression {
		current = current.Parent
	}
	if !noAccessStateIsValueOrObjectPosition(current) {
		return
	}

	// Upstream does NOT break out of this loop after reporting, so an identifier nested inside two
	// `setState` calls reports once per enclosing call. Reproduced; no corpus case nests them.
	for walk := current; walk != nil; walk = walk.Parent {
		if !noAccessStateIsFirstArgumentOfSetState(walk, node) {
			continue
		}
		scope := noAccessStateScopeOf(node)
		for _, variable := range variables {
			if variable.scope == scope && variable.name == node.Text() {
				ctx.ReportNode(variable.node, messageUseSetStateCallback)
			}
		}
	}
}

// noAccessStateVisitObjectPattern handles upstream's ObjectPattern arm.
//
// `var {state, ...rest} = this` records `state` as a variable holding the component state, so a
// later `state.value` inside a `setState` argument reports. The initializer must be exactly `this`.
func noAccessStateVisitObjectPattern(node *ast.Node, variables *[]noAccessStateVariable) {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return
	}
	initializer := parent.AsVariableDeclaration().Initializer
	if initializer == nil || initializer.Kind != ast.KindThisKeyword {
		return
	}
	pattern := node.AsBindingPattern()
	if pattern == nil || pattern.Elements == nil {
		return
	}
	for _, element := range pattern.Elements.Nodes {
		if element.Kind != ast.KindBindingElement {
			continue
		}
		// Upstream reads the property's KEY, which for a shorthand `{state}` is the same identifier
		// as the local name. A renaming `{state: s}` records the key `state` while the usable local
		// is `s`, so the recorded name matches nothing later; that is upstream's behaviour and the
		// property name is what is read here for the same reason.
		binding := element.AsBindingElement()
		key := binding.PropertyName
		if key == nil {
			key = binding.Name()
		}
		if key == nil || key.Kind != ast.KindIdentifier || key.Text() != "state" {
			continue
		}
		*variables = append(*variables, noAccessStateVariable{
			name:  key.Text(),
			node:  key,
			scope: noAccessStateScopeOf(node),
		})
	}
}

// noAccessStateIsFirstArgumentOfSetState reports whether `node` sits inside `candidate`'s first
// argument, where `candidate` is a `this.setState(...)` call.
//
// Upstream walks the node's ancestors up to the candidate and asks whether the child it arrived
// through IS `arguments[0]`. That is the test which separates the first argument from the
// completion callback, and it is why the same `this.state` access is a finding in one position and
// clean in the other.
func noAccessStateIsFirstArgumentOfSetState(candidate *ast.Node, node *ast.Node) bool {
	if !noAccessStateIsSetStateCall(candidate) {
		return false
	}
	arguments := candidate.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return false
	}
	// Climb until the node whose parent is the call, which is the argument it belongs to.
	walk := node
	for walk != nil && walk.Parent != candidate {
		walk = walk.Parent
	}
	return walk != nil && arguments.Nodes[0] == walk
}

// noAccessStateIsSetStateCall reports whether a node is `this.setState(...)`.
//
// The receiver must be `this` specifically. Upstream reads `node.callee.object.type ===
// 'ThisExpression'`, so `that.setState(...)` and `component.setState(...)` are both declines.
func noAccessStateIsSetStateCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := node.AsCallExpression().Expression
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if access.Expression == nil || access.Expression.Kind != ast.KindThisKeyword {
		return false
	}
	name := access.Name()
	return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "setState"
}

// noAccessStateIsValueOrObjectPosition reports whether a node is a property's value or a member
// access's object, which is upstream's gate on the identifier arm.
func noAccessStateIsValueOrObjectPosition(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Expression == node
	case ast.KindElementAccessExpression:
		return parent.AsElementAccessExpression().Expression == node
	}
	return false
}

// noAccessStateScopeOf returns the node whose identity stands in for upstream's scope object.
//
// Upstream compares `getScope(context, node)` by reference, so two sites are the same scope when
// ESLint hands back the same object. The innermost scope-introducing ancestor has that property:
// two nodes share it exactly when no scope boundary sits between them.
//
// Blocks are included and this is not an embellishment. Measured on the installed build, moving a
// `var` declaration into a block makes the rule go silent even though the name still resolves, so
// upstream is genuinely comparing block-level scopes rather than function-level ones.
//
// The source file itself is the outermost answer, so two top-level sites compare equal.
func noAccessStateScopeOf(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindConstructor, ast.KindClassStaticBlockDeclaration,
			ast.KindBlock, ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindCaseBlock, ast.KindCatchClause, ast.KindModuleBlock, ast.KindSourceFile:
			return current
		}
	}
	return nil
}

// noAccessStateCalleeName reads the bare name a call expression invokes, if it has one.
//
// Upstream reads `node.callee.name`, which is present only for a plain identifier callee, so
// `this.nextState()` contributes no name to the propagation half. The REPORTING half compares that
// same possibly-empty name, which is why upstream's own corpus writes an unqualified `nextState()`.
func noAccessStateCalleeName(callee *ast.Node) string {
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return ""
	}
	return callee.Text()
}

// noAccessStateEnclosingMethodName returns the name of the method a node sits inside, for the
// PROPAGATION arm only.
//
// This deliberately accepts fewer shapes than `noAccessStateDeclaringMethodName`, and the asymmetry
// is upstream's rather than an oversight. Upstream's propagation walk tests `current.type ===
// 'MethodDefinition'` and nothing else, while its `this.state` walk tests MethodDefinition AND a
// FunctionExpression sitting under a property. So a method chain propagates inside an ES6 class and
// does NOT propagate inside a `createReactClass` object literal, where every method is a function
// expression under a property.
//
// Measured on the installed build, and this port reported the object form until the probe was run:
//
//	class:  inner reads state, outer calls inner, setState(outer())    REPORTS
//	object: inner reads state, outer calls inner, setState(outer())    SILENT
//	object: inner reads state, setState(inner())                       REPORTS, one hop needs no propagation
//
// The one-hop object case still reports because it never uses this function: the `this.state` walk
// records `inner` through its own wider test, and the reporting half matches it directly.
//
// Reproduced rather than corrected. Widening this to the object form would report a two-hop chain
// upstream calls clean, and the shape is common enough in legacy factory components that the
// difference would show up as new findings nobody asked for.
func noAccessStateEnclosingMethodName(node *ast.Node) string {
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindSourceFile {
			return ""
		}
		// A CLASS method specifically. Our parser gives an object literal's method shorthand the
		// same kind, while ESTree spells it as a Property holding a FunctionExpression and reserves
		// MethodDefinition for a class member. Without the parent check this arm would accept the
		// shorthand and propagate a chain upstream does not, measured: the object shorthand two-hop
		// case is silent upstream and this port reported it until the parent check was added.
		if current.Kind == ast.KindMethodDeclaration && current.Parent != nil &&
			(current.Parent.Kind == ast.KindClassDeclaration ||
				current.Parent.Kind == ast.KindClassExpression) {
			name := current.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				return name.Text()
			}
			return ""
		}
	}
	return ""
}

// noAccessStateDeclaringMethodName answers whether a node declares a named method, and its name.
//
// Two shapes, matching upstream's two arms. A method declaration covers both a class method and an
// object literal method shorthand, which ESTree spells as MethodDefinition and as a Property whose
// value is a FunctionExpression. A function expression assigned to a named property is the third
// spelling and is the one upstream's corpus actually writes.
func noAccessStateDeclaringMethodName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindMethodDeclaration:
		name := node.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.Text(), true
		}
		// A computed or string-literal method name records an empty name upstream too, because it
		// reads `current.key.name` and gets undefined. The arm still fires and stops the walk.
		return "", true

	case ast.KindFunctionExpression:
		parent := node.Parent
		if parent == nil {
			return "", false
		}
		switch parent.Kind {
		case ast.KindPropertyAssignment:
			name := parent.AsPropertyAssignment().Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				return name.Text(), true
			}
			return "", true
		case ast.KindPropertyDeclaration:
			name := parent.AsPropertyDeclaration().Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				return name.Text(), true
			}
			return "", true
		}
	}
	return "", false
}
