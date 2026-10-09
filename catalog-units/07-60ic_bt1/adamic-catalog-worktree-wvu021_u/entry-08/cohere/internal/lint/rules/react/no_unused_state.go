package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageUnusedStateField builds the finding for one state key nothing reads.
//
// The key name is interpolated, so this needs a rendered-text assertion: a message-id assertion
// cannot see anything the interpolation does.
func messageUnusedStateField(name string) rule.Message {
	return rule.Message{
		Id: "unusedStateField",
		Description: "Unused state field: '" + name + "'. Nothing in this component reads it, so " +
			"every write to it re-renders for a value no one looks at, and a reader has to " +
			"search the whole class before they can be sure it is dead.",
	}
}

// NoUnusedState reports a state field the component never reads.
//
//	valid:   this.state = {a: 1}; ... return this.state.a;
//	valid:   this.state = {a: 1}; ... const {a} = this.state;
//	valid:   this.state = {a: 1}; ... return this.state[key];   gives up, see below
//	invalid: this.state = {a: 1}; ... render() { return null; }
//	invalid: state = {a: 1};      ... render() { return null; }
//	invalid: this.setState({a: 1}); with nothing reading `a`
//
// # Two passes over one class, and the whole difficulty is the second
//
// Writes are easy: an object handed to `this.state =` inside the constructor, a `state = {}` class
// property, an object handed to `this.setState`, and the object returned by `getInitialState` in a
// factory component. Reads are where a port goes wrong, because there are six of them and one is a
// parser shape ESTree does not have.
//
// # The four give-up conditions, and why they are the point of the rule
//
// When the rule cannot see through an access it abandons the WHOLE class rather than reporting.
// Each was measured against the installed build on 2026-08-27:
//
//	this.state[key]           computed with a non-literal property   gives up
//	f(this.state)             the state object handed to a call      gives up
//	<D {...this.state} />     spread as JSX attributes               gives up
//	({...this.state})         spread into an object                  gives up
//
// A computed access with a LITERAL property is not a give-up, it is an ordinary read of that key:
// `this.state["a"]` reads `a`, and `this.state["b"]` still reports `a`. Both measured.
//
// Giving up is the difference between a rule people keep and one they turn off. Reporting a key
// that is genuinely read through a computed access is a false positive nobody can act on, because
// the read is right there and the rule cannot see it.
//
// # Aliases, and why they are scoped per method
//
// `const s = this.state` makes `s` an alias, so `s.a` counts as a read of `a`. Upstream keeps one
// alias set per method and clears it on exit, saying in its own comment that tracking properly
// would need full scope and shadowing analysis and that it accepts false negatives instead. That is
// reproduced rather than improved: widening it would report differently from the tool this is
// compared against.
//
// The alias set is nil outside a method, and every alias-adding path tests it first. That nil is
// load-bearing rather than defensive: it is what stops a `const s = this.state` written at class
// property scope from leaking into every method.
//
// # The setState updater
//
// `this.setState(prev => ({b: prev.a}))` reads `a` through the parameter. Upstream tests for an
// ARROW function specifically, so `this.setState(function(prev) {...})` does NOT alias and its
// reads are lost. Measured: the arrow form is silent and the function-expression form reports.
// Reproduced, because it is a decision the corpus records rather than an oversight this port can
// see is wrong.
//
// # The lifecycle state parameter
//
// The second parameter of `shouldComponentUpdate`, `componentWillUpdate`,
// `UNSAFE_componentWillUpdate`, `getSnapshotBeforeUpdate`, `componentDidUpdate` and static
// `getDerivedStateFromProps` receives state, so reads through it count. The list is exact: a method
// with the same shape and a different name does NOT alias its second parameter, measured.
//
// Upstream reaches this by walking the scope chain from each identifier. This walks the enclosing
// function chain from the identifier instead, which answers the same question without a scope
// table, and a fixture pins the negative case.
//
// # What is deliberately not ported
//
// Upstream's `uncast` unwraps Flow's TypeCastExpression, `(expr: any)`. Our parser cannot produce
// that node at all: two of upstream's own cases fail to parse as TypeScript, measured, and both are
// tagged `flow` in the corpus. There is nothing to unwrap here, so the helper has no counterpart.
//
// # No fix
//
// `meta.fixable` is unset upstream. Deleting a state field is not a spelling change: the write that
// set it may have a side effect, and the right repair is usually to delete the code that computes
// the value too, which only the author can decide.
var NoUnusedState = rule.Rule{
	Name: "react/no-unused-state",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				// One walk, driven by this rule rather than by the dispatcher, because the
				// algorithm is stateful across node kinds and depends on entering and LEAVING a
				// class. There is no exit listener in this tree, so the traversal is written here.
				noUnusedStateWalk(ctx, sourceFile)
			},
		}
	},
}

// noUnusedStateClassInfo is upstream's `classInfo`, the per-component accumulator.
//
// A nil pointer means "not inside a component, or we gave up on this one", which is exactly what
// upstream's `classInfo = null` means. Both readings matter: every listener body starts by
// declining when it is nil, and the give-up paths set it to nil deliberately.
type noUnusedStateClassInfo struct {
	// stateFields holds one entry per written key, in source order, carrying the node to report.
	// A slice rather than a map because upstream reports in insertion order and a map would make
	// the finding order depend on Go's randomisation.
	stateFields []noUnusedStateField

	// usedStateFields is the set of key names something read.
	usedStateFields map[string]bool

	// aliases holds local names currently standing in for `this.state`. Nil outside a method,
	// which is upstream's own distinction rather than a Go convenience: an alias recorded at class
	// scope would leak into every method.
	aliases map[string]bool
}

// noUnusedStateField is one written key and the node a finding points at.
type noUnusedStateField struct {
	name string
	node *ast.Node
}

// noUnusedStateWalk traverses the file, maintaining the per-class accumulator.
//
// Written as an explicit recursion rather than as listeners because the rule needs to act on
// LEAVING a class and a method, and the dispatcher offers no exit hook. The pre-order call
// corresponds to upstream's plain visitor and the post-order call to its `:exit` handler.
func noUnusedStateWalk(ctx rule.Context, sourceFile *ast.Node) {
	var classInfo *noUnusedStateClassInfo

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		// entering, and whether this node owns the class or the alias scope
		enteredClass := false
		var restoredAliases map[string]bool
		restoresAliases := false

		switch node.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if isComponentClass(node) {
				classInfo = &noUnusedStateClassInfo{usedStateFields: map[string]bool{}}
				enteredClass = true
			}

		case ast.KindObjectLiteralExpression:
			// The factory form. Upstream anchors the component on the object literal handed to
			// `createReactClass`, so the accumulator opens and closes on that object.
			if noUnusedStateIsFactoryObject(node) {
				classInfo = &noUnusedStateClassInfo{usedStateFields: map[string]bool{}}
				enteredClass = true
			}

		case ast.KindMethodDeclaration:
			if classInfo != nil {
				// The factory's `getInitialState` can be written as an object method SHORTHAND,
				// which our parser spells as a method declaration while ESTree spells it as a
				// property holding a function expression. Upstream's `FunctionExpression` arm
				// therefore covers a shape that never reaches it here, and without this the
				// shorthand form records no state fields at all. Measured: upstream reports on
				// `createReactClass({getInitialState() {return {foo: 0};}})` and this port was
				// silent on it until the shorthand was routed here too.
				noUnusedStateVisitFunctionExpression(ctx, classInfo, node)
				if !noUnusedStateIsGetInitialState(node) {
					restoredAliases, restoresAliases = classInfo.aliases, true
					classInfo.aliases = map[string]bool{}
				}
			}

		case ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			// Upstream's `MethodDefinition` covers a constructor and both accessors, which our
			// parser spells as three separate kinds.
			if classInfo != nil {
				restoredAliases, restoresAliases = classInfo.aliases, true
				classInfo.aliases = map[string]bool{}
			}

		case ast.KindPropertyDeclaration:
			if classInfo != nil {
				noUnusedStateVisitClassProperty(ctx, classInfo, node)
				// A property holding an arrow function gets its own alias scope, matching
				// upstream's ClassProperty handler. `getDerivedStateFromProps` written as a static
				// property is excluded there, and it is excluded here for the same reason: its own
				// arm has already recorded the reads through its state parameter.
				if noUnusedStatePropertyOpensAnAliasScope(node) {
					restoredAliases, restoresAliases = classInfo.aliases, true
					classInfo.aliases = map[string]bool{}
				}
			}

		case ast.KindFunctionExpression:
			if classInfo != nil {
				noUnusedStateVisitFunctionExpression(ctx, classInfo, node)
				restoredAliases, restoresAliases = classInfo.aliases, true
				if noUnusedStateIsGetInitialState(node) {
					// `getInitialState` does not open an alias scope upstream; its arm returns
					// before the `else` that would.
					classInfo.aliases = restoredAliases
					restoresAliases = false
				} else {
					classInfo.aliases = map[string]bool{}
				}
			}

		case ast.KindCallExpression:
			if classInfo != nil {
				noUnusedStateVisitCall(ctx, classInfo, node)
			}

		case ast.KindBinaryExpression:
			if classInfo != nil {
				noUnusedStateVisitAssignment(ctx, classInfo, node)
			}

		case ast.KindVariableDeclaration:
			if classInfo != nil {
				declaration := node.AsVariableDeclaration()
				if declaration.Initializer != nil {
					noUnusedStateHandleAssignment(ctx, classInfo, declaration.Name(),
						noUnusedStateUnwrap(declaration.Initializer))
				}
			}

		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			if classInfo != nil {
				if noUnusedStateVisitMemberAccess(ctx, classInfo, node) {
					// The access was a give-up. Upstream sets classInfo to null, which abandons
					// the class without reporting, and the accumulator stays nil for the rest of
					// the file until a new component opens one.
					classInfo = nil
				}
			}

		case ast.KindJsxSpreadAttribute:
			if classInfo != nil &&
				noUnusedStateIsStateReference(classInfo,
					noUnusedStateUnwrap(node.AsJsxSpreadAttribute().Expression)) {
				classInfo = nil
			}

		case ast.KindSpreadAssignment:
			if classInfo != nil &&
				noUnusedStateIsStateReference(classInfo,
					noUnusedStateUnwrap(node.AsSpreadAssignment().Expression)) {
				classInfo = nil
			}

		case ast.KindSpreadElement:
			if classInfo != nil &&
				noUnusedStateIsStateReference(classInfo,
					noUnusedStateUnwrap(node.AsSpreadElement().Expression)) {
				classInfo = nil
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})

		// leaving
		if restoresAliases && classInfo != nil {
			classInfo.aliases = restoredAliases
		}
		if enteredClass {
			if classInfo != nil {
				noUnusedStateReport(ctx, classInfo)
			}
			classInfo = nil
		}
	}
	visit(sourceFile)
}

// noUnusedStateReport emits one finding per written key nothing read.
func noUnusedStateReport(ctx rule.Context, classInfo *noUnusedStateClassInfo) {
	for _, field := range classInfo.stateFields {
		if classInfo.usedStateFields[field.name] {
			continue
		}
		ctx.ReportNode(field.node, messageUnusedStateField(field.name))
	}
}

// noUnusedStateIsFactoryObject reports whether an object literal is the argument to the factory.
//
// Upstream's `isES5Component` asks whether the object's PARENT is a `createReactClass` call, which
// is the same question asked from the other end.
func noUnusedStateIsFactoryObject(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	if !isCreateReactClassCall(parent) {
		return false
	}
	arguments := parent.AsCallExpression().Arguments
	return arguments != nil && len(arguments.Nodes) > 0 && arguments.Nodes[0] == node
}

// noUnusedStateVisitClassProperty handles `state = {}` and the lifecycle state parameter.
func noUnusedStateVisitClassProperty(
	ctx rule.Context,
	classInfo *noUnusedStateClassInfo,
	node *ast.Node,
) {
	property := node.AsPropertyDeclaration()
	name, named := noUnusedStateKeyName(node.Name())
	initializer := noUnusedStateUnwrap(property.Initializer)

	if named && name == "state" && !noUnusedStateIsStatic(node) && initializer != nil &&
		initializer.Kind == ast.KindObjectLiteralExpression {
		noUnusedStateAddStateFields(classInfo, initializer)
	}

	// `static getDerivedStateFromProps = (props, state) => ...` written as a property. Upstream has
	// a separate handler for it that reads the state parameter's references directly.
	if noUnusedStateIsDerivedStateFromProps(node) && initializer != nil {
		noUnusedStateRecordParameterReads(ctx, classInfo, initializer, 1)
	}
}

// noUnusedStatePropertyOpensAnAliasScope reports whether a class property gets its own alias set.
//
// Upstream opens one for any non-static property holding an arrow function, and its exit handler
// clears it only when the property is not `getDerivedStateFromProps`. The asymmetry does not change
// any verdict, because the entry it would preserve is written by the same property; opening the
// scope for the arrow and closing it on the way out is the behaviour, and this reproduces that.
func noUnusedStatePropertyOpensAnAliasScope(node *ast.Node) bool {
	if noUnusedStateIsStatic(node) {
		return false
	}
	initializer := node.AsPropertyDeclaration().Initializer
	return initializer != nil && initializer.Kind == ast.KindArrowFunction
}

// noUnusedStateIsDerivedStateFromProps reports whether a member is the static lifecycle hook that
// receives state as its second parameter.
func noUnusedStateIsDerivedStateFromProps(node *ast.Node) bool {
	if !noUnusedStateIsStatic(node) {
		return false
	}
	name, named := noUnusedStateKeyName(node.Name())
	if !named || name != "getDerivedStateFromProps" {
		return false
	}
	var function *ast.Node
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		function = node.AsPropertyDeclaration().Initializer
	case ast.KindMethodDeclaration:
		function = node
	}
	if function == nil {
		return false
	}
	// Upstream requires at least two parameters, so a one-parameter hook records nothing.
	//
	// This check is SUBSUMED and the sweep correctly scores its removal as surviving. The only
	// caller is the class-property arm, which passes index 1 to
	// `noUnusedStateRecordParameterReads`, and that function opens with `len(parameters) <= index`.
	// So any input reaching here with fewer than two parameters is declined one call later by the
	// same arithmetic, and no input can distinguish the two versions.
	//
	// Settled by enumeration rather than argued: every parameter count from zero to three, with and
	// without a read through the second parameter, measured against the installed build and
	// matching this port on all five. It is kept because upstream writes it, and because deleting a
	// guard whose redundancy depends on a constant at a distant call site is how the redundancy
	// stops being true.
	return len(noUnusedStateParametersOf(function)) >= 2
}

// noUnusedStateVisitFunctionExpression handles the factory's `getInitialState`.
func noUnusedStateVisitFunctionExpression(
	ctx rule.Context,
	classInfo *noUnusedStateClassInfo,
	node *ast.Node,
) {
	if !noUnusedStateIsGetInitialState(node) {
		return
	}
	body := noUnusedStateBodyOf(node)
	if body == nil || body.AsBlock() == nil || body.AsBlock().Statements == nil {
		return
	}
	statements := body.AsBlock().Statements.Nodes
	if len(statements) == 0 {
		return
	}
	// Upstream reads only the LAST statement of the body, so a return written earlier contributes
	// nothing. Reproduced rather than widened.
	last := statements[len(statements)-1]
	if last.Kind != ast.KindReturnStatement {
		return
	}
	returned := last.AsReturnStatement().Expression
	if returned != nil && returned.Kind == ast.KindObjectLiteralExpression {
		noUnusedStateAddStateFields(classInfo, returned)
	}
}

// noUnusedStateIsGetInitialState reports whether a node is the factory's state initializer.
//
// Two spellings reach here because our parser splits what ESTree writes as one shape. A function
// expression under a property assignment is `getInitialState: function() {...}`; a method
// declaration whose parent is the object is `getInitialState() {...}`. Upstream sees a
// FunctionExpression in both cases and its single arm covers them together.
func noUnusedStateIsGetInitialState(node *ast.Node) bool {
	var object *ast.Node
	var nameNode *ast.Node

	switch node.Kind {
	case ast.KindMethodDeclaration:
		object = node.Parent
		nameNode = node.Name()
	default:
		parent := node.Parent
		if parent == nil || parent.Kind != ast.KindPropertyAssignment {
			return false
		}
		object = parent.Parent
		nameNode = parent.AsPropertyAssignment().Name()
	}

	// Upstream additionally requires the enclosing object to be a factory component, which is what
	// keeps a `getInitialState` on a plain object from contributing.
	if object == nil || object.Kind != ast.KindObjectLiteralExpression ||
		!noUnusedStateIsFactoryObject(object) {
		return false
	}
	name, named := noUnusedStateKeyName(nameNode)
	return named && name == "getInitialState"
}

// noUnusedStateBodyOf returns a function-like node's block body, or nil.
func noUnusedStateBodyOf(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Body
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Body
	}
	return nil
}

// noUnusedStateVisitCall handles `this.setState(...)` in both of its argument shapes.
func noUnusedStateVisitCall(ctx rule.Context, classInfo *noUnusedStateClassInfo, node *ast.Node) {
	if !noUnusedStateIsSetStateCall(node) {
		return
	}
	arguments := node.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return
	}
	first := noUnusedStateUnwrap(arguments.Nodes[0])
	if first == nil {
		return
	}

	if first.Kind == ast.KindObjectLiteralExpression {
		noUnusedStateAddStateFields(classInfo, first)
		return
	}

	// The updater form. Upstream tests for an ARROW function specifically, so a function expression
	// handed to setState contributes neither fields nor an alias. Measured against the installed
	// build: the arrow form is silent and the function-expression form reports.
	if first.Kind != ast.KindArrowFunction {
		return
	}
	arrow := first.AsArrowFunction()
	if arrow.Body != nil {
		// `noUnusedStateUnwrap` rather than `skipParenthesesOptional`, because an updater body can
		// be assertion-wrapped as well as parenthesized: `s => ({a: 1} as unknown)` is one of
		// upstream's own failing cases and the parenthesis-only skip leaves the assertion in place.
		body := noUnusedStateUnwrap(arrow.Body)
		if body != nil && body.Kind == ast.KindObjectLiteralExpression {
			noUnusedStateAddStateFields(classInfo, body)
		}
	}
	parameters := noUnusedStateParametersOf(first)
	if len(parameters) == 0 || classInfo.aliases == nil {
		return
	}
	firstParameterName := parameters[0].AsParameterDeclaration().Name()
	if firstParameterName == nil {
		return
	}
	if firstParameterName.Kind == ast.KindObjectBindingPattern {
		noUnusedStateHandleStateDestructuring(classInfo, firstParameterName)
		return
	}
	// The kind guard is load-bearing rather than defensive: `Node.Text()` PANICS on a binding
	// pattern rather than returning a name, measured directly, and a panic costs every rule its
	// verdict on the whole file because the walk recovers per file rather than per rule. An array
	// pattern reaches this line and must not be read.
	if firstParameterName.Kind == ast.KindIdentifier {
		classInfo.aliases[firstParameterName.Text()] = true
	}
}

// noUnusedStateVisitAssignment handles `this.state = {}` and alias-creating assignments.
func noUnusedStateVisitAssignment(
	ctx rule.Context,
	classInfo *noUnusedStateClassInfo,
	node *ast.Node,
) {
	expression := node.AsBinaryExpression()
	if expression.OperatorToken == nil || expression.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}
	left, right := noUnusedStateUnwrap(expression.Left), noUnusedStateUnwrap(expression.Right)
	if left == nil || right == nil {
		return
	}

	if noUnusedStateIsThisStateAccess(left) && right.Kind == ast.KindObjectLiteralExpression {
		// Upstream walks up to the nearest FunctionExpression and records the fields only when
		// that function is the constructor. So `this.state = {}` written in any other method
		// contributes nothing, which is deliberate: assigning state outside the constructor is a
		// different mistake that a different rule reports.
		if noUnusedStateInConstructor(node) {
			noUnusedStateAddStateFields(classInfo, right)
		}
		return
	}
	noUnusedStateHandleAssignment(ctx, classInfo, left, right)
}

// noUnusedStateInConstructor reports whether a node sits inside a constructor body.
//
// Upstream climbs to the nearest FunctionExpression and asks whether its parent is a
// MethodDefinition of kind `constructor`. Our parser gives a constructor its own node kind, so the
// climb stops at the first function-like ancestor and asks whether that ancestor IS one.
func noUnusedStateInConstructor(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindConstructor:
			return true
		case ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindClassDeclaration, ast.KindClassExpression, ast.KindSourceFile:
			return false
		}
	}
	return false
}

// noUnusedStateHandleAssignment records aliases and destructured reads, for both an assignment
// expression and a variable declarator.
func noUnusedStateHandleAssignment(
	ctx rule.Context,
	classInfo *noUnusedStateClassInfo,
	left *ast.Node,
	right *ast.Node,
) {
	if left == nil || right == nil {
		return
	}
	switch left.Kind {
	case ast.KindIdentifier:
		// `const s = this.state` makes `s` an alias.
		if noUnusedStateIsStateReference(classInfo, right) && classInfo.aliases != nil {
			classInfo.aliases[left.Text()] = true
		}

	case ast.KindObjectBindingPattern:
		if noUnusedStateIsStateReference(classInfo, right) {
			noUnusedStateHandleStateDestructuring(classInfo, left)
			return
		}
		// `const {state} = this` and `const {state: {foo}} = this`.
		if right.Kind == ast.KindThisKeyword && classInfo.aliases != nil {
			noUnusedStateHandleThisDestructuring(classInfo, left)
		}
	}
}

// noUnusedStateHandleThisDestructuring records what `const {state: ...} = this` introduces.
func noUnusedStateHandleThisDestructuring(
	classInfo *noUnusedStateClassInfo,
	pattern *ast.Node,
) {
	elements := noUnusedStateBindingElements(pattern)
	for _, element := range elements {
		binding := element.AsBindingElement()
		if binding.DotDotDotToken != nil {
			continue
		}
		key, named := noUnusedStateKeyName(noUnusedStateBindingKey(element))
		if !named || key != "state" {
			continue
		}
		// The bound value is either a name, which becomes an alias, or a nested pattern, whose
		// keys are reads.
		value := binding.Name()
		if value == nil {
			continue
		}
		if value.Kind == ast.KindObjectBindingPattern {
			noUnusedStateHandleStateDestructuring(classInfo, value)
			continue
		}
		// Guarded because `Text()` panics on a pattern; an array pattern reaches here.
		if value.Kind == ast.KindIdentifier {
			classInfo.aliases[value.Text()] = true
		}
	}
}

// noUnusedStateHandleStateDestructuring records reads and rest aliases for a pattern over state.
func noUnusedStateHandleStateDestructuring(
	classInfo *noUnusedStateClassInfo,
	pattern *ast.Node,
) {
	for _, element := range noUnusedStateBindingElements(pattern) {
		binding := element.AsBindingElement()
		if binding.DotDotDotToken != nil {
			// `const {...rest} = this.state` makes `rest` an alias for the whole state object.
			if classInfo.aliases == nil {
				continue
			}
			name := binding.Name()
			// Guarded: `Text()` panics on a pattern, and `const {...{a}} = x` is parseable.
			if name != nil && name.Kind == ast.KindIdentifier {
				classInfo.aliases[name.Text()] = true
			}
			continue
		}
		if key, named := noUnusedStateKeyName(noUnusedStateBindingKey(element)); named {
			classInfo.usedStateFields[key] = true
		}
	}
}

// noUnusedStateBindingKey returns the node naming the PROPERTY a binding element reads.
//
// This is the shape difference that a port written from ESTree gets wrong. There, a destructuring
// property carries `key` and `value` as separate fields whether or not the source renames. Here a
// shorthand `{foo}` produces a binding element whose `PropertyName` is NIL and whose `Name` is the
// identifier, while a renaming `{foo: bar}` fills both. So the property being read is `PropertyName`
// when it is present and `Name` otherwise, and reading `Name` unconditionally would record the
// LOCAL name of a renamed binding as a state key.
func noUnusedStateBindingKey(element *ast.Node) *ast.Node {
	binding := element.AsBindingElement()
	if binding.PropertyName != nil {
		return binding.PropertyName
	}
	return binding.Name()
}

// noUnusedStateBindingElements returns a binding pattern's elements, or nothing.
func noUnusedStateBindingElements(pattern *ast.Node) []*ast.Node {
	if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return nil
	}
	bindingPattern := pattern.AsBindingPattern()
	if bindingPattern == nil || bindingPattern.Elements == nil {
		return nil
	}
	elements := []*ast.Node{}
	for _, element := range bindingPattern.Elements.Nodes {
		if element.Kind == ast.KindBindingElement {
			elements = append(elements, element)
		}
	}
	return elements
}

// noUnusedStateVisitMemberAccess handles every read, and answers whether the class must be
// abandoned.
//
// The boolean return is the give-up signal. Upstream writes `classInfo = null` in two places here
// and the caller reproduces that by dropping the accumulator.
func noUnusedStateVisitMemberAccess(
	ctx rule.Context,
	classInfo *noUnusedStateClassInfo,
	node *ast.Node,
) bool {
	var object, property *ast.Node
	computed := false
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		object, property = noUnusedStateUnwrap(access.Expression), access.Name()
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		object, property = noUnusedStateUnwrap(access.Expression), access.ArgumentExpression
		computed = true
	}

	if noUnusedStateIsStateReference(classInfo, object) {
		// `this.state[key]` with a non-literal property. Nothing can be concluded about which key
		// was read, so the whole class is abandoned rather than reported on.
		if computed && !noUnusedStateIsLiteralKey(property) {
			return true
		}
		if name, named := noUnusedStateKeyName(property); named {
			classInfo.usedStateFields[name] = true
		}
		return false
	}

	// `f(this.state)` hands the whole object somewhere this rule cannot follow.
	if noUnusedStateIsStateReference(classInfo, node) && node.Parent != nil &&
		node.Parent.Kind == ast.KindCallExpression &&
		node.Parent.AsCallExpression().Expression != node {
		return true
	}
	return false
}

// noUnusedStateIsLiteralKey reports whether a computed property is a literal upstream accepts.
//
// Upstream's give-up test is `node.property.type !== 'Literal'`, and its ESTree Literal covers a
// string and a number. A template literal is NOT a Literal there, so `this.state[`+"`a`"+`]` gives
// up even though its value is knowable; reproduced.
func noUnusedStateIsLiteralKey(property *ast.Node) bool {
	if property == nil {
		return false
	}
	switch property.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral:
		return true
	}
	return false
}

// noUnusedStateIsStateReference reports whether a node may denote `this.state`.
//
// Three routes, matching upstream: the direct `this.state`, a local alias, and the state parameter
// of a lifecycle method.
func noUnusedStateIsStateReference(classInfo *noUnusedStateClassInfo, node *ast.Node) bool {
	if node == nil {
		return false
	}
	if noUnusedStateIsThisStateAccess(node) {
		return true
	}
	if node.Kind == ast.KindIdentifier {
		if classInfo.aliases != nil && classInfo.aliases[node.Text()] {
			return true
		}
		return noUnusedStateIsLifecycleStateParameter(node)
	}
	return false
}

// noUnusedStateIsThisStateAccess reports whether a node is exactly `this.state`.
func noUnusedStateIsThisStateAccess(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := node.AsPropertyAccessExpression()
	receiver := noUnusedStateUnwrap(access.Expression)
	if receiver == nil || receiver.Kind != ast.KindThisKeyword {
		return false
	}
	name := access.Name()
	// Upstream reads `node.property.name` here rather than going through its own `getName`, so a
	// computed or literal spelling does not match. Our `Name()` on a property access is always an
	// identifier or private name, so the kind test carries the same restriction.
	return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "state"
}

// noUnusedStateLifecycleMethodsReceivingState is the exact list whose SECOND parameter is state.
//
// Exact rather than approximate: a method with the same shape and a different name does not alias
// its second parameter, measured against the installed build.
var noUnusedStateLifecycleMethodsReceivingState = map[string]bool{
	"shouldComponentUpdate":      true,
	"componentWillUpdate":        true,
	"UNSAFE_componentWillUpdate": true,
	"getSnapshotBeforeUpdate":    true,
	"componentDidUpdate":         true,
	"getDerivedStateFromProps":   true,
}

// noUnusedStateIsLifecycleStateParameter reports whether an identifier names the state parameter of
// a lifecycle method.
//
// Upstream walks the SCOPE chain outward asking, at each level, whether the enclosing block is one
// of these methods and whether the identifier's name matches its second parameter. This walks the
// enclosing function chain instead, which answers the same question without a scope table: at every
// function-like ancestor, ask whether that function is such a method and whether the name matches.
//
// The name comparison is upstream's own and it is looser than resolution: an identifier that merely
// SHARES the parameter's spelling matches, even where a nested declaration shadows it. Reproduced,
// because the alternative reports differently from the tool this is compared against.
func noUnusedStateIsLifecycleStateParameter(identifier *ast.Node) bool {
	name := identifier.Text()
	for current := identifier.Parent; current != nil; current = current.Parent {
		member := noUnusedStateEnclosingMember(current)
		if member == nil {
			continue
		}
		memberName, named := noUnusedStateKeyName(member.Name())
		if !named {
			continue
		}
		if memberName == "getDerivedStateFromProps" {
			if !noUnusedStateIsStatic(member) {
				continue
			}
		} else if !noUnusedStateLifecycleMethodsReceivingState[memberName] {
			continue
		}
		parameters := noUnusedStateParametersOf(current)
		if len(parameters) < 2 {
			continue
		}
		second := parameters[1].AsParameterDeclaration().Name()
		// Guarded: `Text()` panics on a destructured parameter, which is legal here.
		if second != nil && second.Kind == ast.KindIdentifier && second.Text() == name {
			return true
		}
	}
	return false
}

// noUnusedStateEnclosingMember returns the class member a function node belongs to, if any.
//
// A method declaration is its own member. A function or arrow assigned to a class property has the
// property as its member, which is the `static getDerivedStateFromProps = (p, s) => ...` shape.
func noUnusedStateEnclosingMember(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindMethodDeclaration:
		return node
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		parent := node.Parent
		if parent != nil && parent.Kind == ast.KindPropertyDeclaration {
			return parent
		}
	}
	return nil
}

// noUnusedStateRecordParameterReads records every `param.key` read inside a function.
//
// Upstream resolves the parameter's references through scope analysis and reads the property off
// each. This walks the function body for member accesses whose object is an identifier matching the
// parameter's name, which answers the same question for the shapes the corpus writes.
func noUnusedStateRecordParameterReads(
	ctx rule.Context,
	classInfo *noUnusedStateClassInfo,
	function *ast.Node,
	index int,
) {
	parameters := noUnusedStateParametersOf(function)
	if len(parameters) <= index {
		return
	}
	name := parameters[index].AsParameterDeclaration().Name()
	// Guarded: `Text()` panics on a destructured parameter.
	if name == nil || name.Kind != ast.KindIdentifier {
		return
	}
	parameterName := name.Text()

	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindPropertyAccessExpression {
			access := node.AsPropertyAccessExpression()
			if access.Expression != nil && access.Expression.Kind == ast.KindIdentifier &&
				access.Expression.Text() == parameterName {
				if key, named := noUnusedStateKeyName(access.Name()); named {
					classInfo.usedStateFields[key] = true
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(function)
}

// noUnusedStateAddStateFields records every statically-named key of an object literal as written.
//
// Upstream accepts a key that is a literal, a substitution-free template, or a non-computed
// identifier, and skips everything else. A spread inside the object is not a property and
// contributes nothing.
func noUnusedStateAddStateFields(classInfo *noUnusedStateClassInfo, object *ast.Node) {
	literal := object.AsObjectLiteralExpression()
	if literal == nil || literal.Properties == nil {
		return
	}
	for _, property := range literal.Properties.Nodes {
		switch property.Kind {
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		default:
			// A spread, a method, or an accessor. Upstream filters on `prop.type === 'Property'`,
			// which excludes all three.
			continue
		}
		name, named := noUnusedStateKeyName(property.Name())
		if !named {
			continue
		}
		classInfo.stateFields = append(classInfo.stateFields,
			noUnusedStateField{name: name, node: property})
	}
}

// noUnusedStateKeyName renders the name a key node denotes, or reports that it has none.
//
// This is upstream's `getName`, narrowed to what our parser can produce. An identifier gives its
// text, a string or numeric literal gives its value, and a template with no substitutions gives its
// text. A computed key whose expression is not one of those has no static name.
//
// The kind switch is what keeps this off the panic path: it never calls `Text()` on a node whose
// kind it has not first confirmed.
func noUnusedStateKeyName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return node.Text(), true

	case ast.KindComputedPropertyName:
		// Upstream's filter accepts a computed key only when it is a literal or a
		// substitution-free template, so the inner expression is what is read.
		inner := node.AsComputedPropertyName().Expression
		if inner == nil {
			return "", false
		}
		return noUnusedStateLiteralText(inner)
	}
	return noUnusedStateLiteralText(node)
}

// noUnusedStateLiteralText renders a node the way upstream's `getName` renders an estree Literal.
//
// Upstream writes `String(node.value)`, and an estree Literal covers more than a string and a
// number: `true`, `false` and `null` are Literals whose values stringify to their keywords. Our
// parser gives each of those its own keyword kind, so they have to be named individually, and a
// port that stopped at string and numeric goes silent on two of upstream's own failing cases.
//
// Measured against the installed build on 2026-08-27, one class per run:
//
//	state = {[true]: 0}   reports  Unused state field: 'true'
//	state = {[null]: 0}   reports  Unused state field: 'null'
//	state = {[`+"`a`"+`]: 0}    reports  Unused state field: 'a'
//	state = {[k]: 0}      SILENT, an identifier is not a literal and has no static name
//
// A substitution-free template is separate from the Literal case upstream, handled by its own
// branch, and it is accepted in both the write and the read position. Both measured.
//
// The kind switch is also what keeps this off the panic path: it never calls `Text()` on a node
// whose kind it has not first confirmed, and `Text()` panics on several kinds rather than
// returning empty.
func noUnusedStateLiteralText(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return node.Text(), true
	case ast.KindTrueKeyword:
		return "true", true
	case ast.KindFalseKeyword:
		return "false", true
	case ast.KindNullKeyword:
		return "null", true
	}
	return "", false
}

// noUnusedStateUnwrap strips TypeScript assertion wrappers and parentheses from an expression.
//
// This is upstream's `unwrapTSAsExpression`, plus a parenthesis skip upstream does not need. Two
// separate reasons, and conflating them would leave a gap:
//
// The ASSERTION half is upstream's. `this.state = {a: 1} as unknown` writes an object through an
// `as` expression, and without the strip the right-hand side is not an object literal and no field
// is recorded. Upstream strips `TSAsExpression` in eight places for exactly this.
//
// The PARENTHESIS half is ours. Our parser keeps `KindParenthesizedExpression` as a real node and
// the tree upstream walks folds it away, so `(this as unknown).state` arrives here as a property
// access whose object is a PARENTHESIZED node wrapping the assertion. Stripping only the assertion
// would still leave the parenthesis and the receiver would not read as `this`.
//
// The loop matters rather than a single step, because the two wrappers nest and repeat:
// `((this as unknown) as unknown)` is three nodes deep.
//
// Measured against the installed build on 2026-08-27, six positions where the strip changes the
// verdict: the right side of a state assignment, the receiver of the assignment target, the object
// of a read, a setState argument, a setState callee, and the right side of an alias assignment.
func noUnusedStateUnwrap(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		default:
			return node
		}
	}
	return nil
}

// noUnusedStateIsSetStateCall reports whether a call is `this.setState(...)`.
func noUnusedStateIsSetStateCall(node *ast.Node) bool {
	callee := noUnusedStateUnwrap(node.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	// The RECEIVER needs unwrapping too, not only the callee. `((this as unknown).setState as
	// unknown)({...})` unwraps at the callee to a property access whose object is still a
	// parenthesized assertion, and reading it raw fails the `this` test. Measured: upstream reports
	// on that shape and this port was silent on it until the receiver was unwrapped as well.
	receiver := noUnusedStateUnwrap(access.Expression)
	if receiver == nil || receiver.Kind != ast.KindThisKeyword {
		return false
	}
	name, named := noUnusedStateKeyName(access.Name())
	return named && name == "setState"
}

// noUnusedStateIsStatic reports whether a class member carries the static modifier.
func noUnusedStateIsStatic(node *ast.Node) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindStaticKeyword {
			return true
		}
	}
	return false
}

// noUnusedStateParametersOf returns a function-like node's parameters, or nothing.
func noUnusedStateParametersOf(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	list := node.Parameters()
	if list == nil {
		return nil
	}
	return list
}
