package react

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	reactUtilities "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// noUnusedClassComponentMethodsLifecycle is the set every component may declare without using.
//
// Copied name for name from upstream's `LIFECYCLE_METHODS`. React calls each of these itself, so a
// declaration with no call site in the class is exactly what the framework expects.
var noUnusedClassComponentMethodsLifecycle = map[string]bool{
	"constructor":                      true,
	"componentDidCatch":                true,
	"componentDidMount":                true,
	"componentDidUpdate":               true,
	"componentWillMount":               true,
	"componentWillReceiveProps":        true,
	"componentWillUnmount":             true,
	"componentWillUpdate":              true,
	"getChildContext":                  true,
	"getSnapshotBeforeUpdate":          true,
	"render":                           true,
	"shouldComponentUpdate":            true,
	"UNSAFE_componentWillMount":        true,
	"UNSAFE_componentWillReceiveProps": true,
	"UNSAFE_componentWillUpdate":       true,
}

// noUnusedClassComponentMethodsEs6Lifecycle is exempt only in a class component.
//
// Upstream's `ES6_LIFECYCLE`. `state` is the class field React reads directly, and it is NOT exempt
// in a `createReactClass` object, where the equivalent is `getInitialState`.
var noUnusedClassComponentMethodsEs6Lifecycle = map[string]bool{
	"state": true,
}

// noUnusedClassComponentMethodsEs5Lifecycle is exempt only in a createReactClass object.
//
// Upstream's `ES5_LIFECYCLE`, and the mirror of the set above: `getInitialState` in a class
// component REPORTS, because a class has no such hook. Measured on the installed build.
var noUnusedClassComponentMethodsEs5Lifecycle = map[string]bool{
	"getInitialState": true,
	"getDefaultProps": true,
	"mixins":          true,
}

// NoUnusedClassComponentMethods flags a component member nothing in the component reads.
//
//	valid:   class Foo extends React.Component { render() { return null; } }
//	valid:   class Foo extends React.Component { handleClick() {} componentDidMount() { this.handleClick(); } }
//	valid:   class Foo extends React.Component { static helper() {} render() { return null; } }
//	invalid: class Foo extends React.Component { handleClick() {} render() { return null; } }
//	invalid: class Foo extends React.Component { componentDidMount() { this.foo = 1; } render() { return null; } }
//
// Ported from `react/no-unused-class-component-methods` in `eslint-plugin-react`, the implementation
// that defined it. Its 36 clean and 22 reporting cases were extracted mechanically from the clone by
// stubbing its `RuleTester`, then all 58 were run against the installed build, 7.37.5, through the
// ESLint Linter API with the TypeScript parser. The two authorities agreed on all 58, compared on
// the RENDERED message text rather than on message ids, because the corpus states raw text and both
// ids interpolate.
//
// # Two ids, and which one fires depends on whether the class has a name
//
// `unusedWithClass` names the class and `unused` does not, and upstream picks between them on
// `classNode.id && classNode.id.name`. A `createReactClass` object has no name, so the es5 arm
// always reports `unused`, and so does `export default class extends React.Component`. Both
// measured, and both are pinned by fixture, because the two ids carry different interpolations and
// an id assertion alone cannot see a message naming the wrong class.
//
// # The gather is a pre-pass, because our walk has no exit event
//
// Upstream collects on the way down and reports in `ClassDeclaration:exit`. There is no
// `rule.OnExit` here and the walk is pre-order, so this listens on `KindSourceFile`, walks each
// component subtree itself, and reports at the end of that walk. That is the shipped shape for a
// gather-then-judge rule in this tree.
//
// # What counts as a definition and what counts as a use, measured
//
// The distinction is narrower than it reads and several shapes fall on the surprising side. Every
// row below was measured on the installed build:
//
//	this.foo = 1            DEFINITION, so a lone assignment reports
//	this.foo += 1           DEFINITION too, it is an AssignmentExpression as well
//	this.foo++              USE, an update expression is not an assignment
//	delete this.foo         USE
//	this.foo.bar = 1        USE of `foo`, the assignment target is the outer member
//	const {foo} = this      USE
//	this.foo() inside a STATIC method   NOT a use, the static guard suppresses collection
//
// The static one is the least obvious and it is deliberate upstream: entering a static member sets
// a flag that makes every `this` member access inside it invisible, on the reasoning that `this` in
// a static method is the constructor rather than the instance.
//
// # A class EXPRESSION is silent
//
// Upstream listens on `ClassDeclaration` only, so `var Foo = class extends React.Component { ... }`
// reports nothing however unused its members are. Measured, and reproduced rather than widened.
var NoUnusedClassComponentMethods = rule.Rule{
	Name: "react/no-unused-class-component-methods",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				var walk func(node *ast.Node)
				walk = func(node *ast.Node) {
					if node == nil {
						return
					}
					switch {
					case node.Kind == ast.KindClassDeclaration && reactUtilities.IsEs6ComponentClass(node):
						noUnusedClassComponentMethodsCheckClass(ctx, node)
					case node.Kind == ast.KindObjectLiteralExpression &&
						noUnusedClassComponentMethodsIsEs5Specification(node):
						noUnusedClassComponentMethodsCheckObject(ctx, node)
					}
					node.ForEachChild(func(child *ast.Node) bool {
						walk(child)
						return false
					})
				}
				walk(sourceFile)
			},
		}
	},
}

// noUnusedClassComponentMethodsIsEs5Specification reports whether an object is a component body.
//
// Upstream's `isES5Component` looks at the object's PARENT call, requiring the callee to be the
// createClass pragma. With no React settings read here that pragma is `createReactClass`, and the
// namespaced `React.createClass` fails upstream's property test for the same reason. Written inline
// rather than through the shelf's `react.IsEs5ComponentCall`, which accepts `createClass` and the
// namespaced spelling and would report on objects upstream ignores.
func noUnusedClassComponentMethodsIsEs5Specification(object *ast.Node) bool {
	parent := object.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	callee := call.Expression
	// The ARGUMENT POSITION is deliberately not checked, and this port had it wrong at first.
	// Upstream's `isES5Component` looks only at `node.parent.callee`, so an object in ANY argument
	// slot counts. A mutant dropping a first-argument test survived every fixture and probing the
	// installed build settled it: `createReactClass(other, { handleClick() {} })` REPORTS. A nested
	// object is still not a component, because its parent is a property rather than the call.
	return callee != nil && callee.Kind == ast.KindIdentifier && callee.Text() == "createReactClass"
}

// noUnusedClassComponentMethodsCheckClass gathers and reports for one class component.
func noUnusedClassComponentMethodsCheckClass(ctx rule.Context, class *ast.Node) {
	className := ""
	if name := class.AsClassDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
		className = name.Text()
	}

	state := &noUnusedClassComponentMethodsState{used: map[string]bool{}}

	members := class.AsClassDeclaration().Members
	if members != nil {
		for _, member := range members.Nodes {
			// A static member contributes no definition and, crucially, no USES either: upstream
			// sets a flag on entry that makes every `this` access inside invisible.
			if noUnusedClassComponentMethodsIsStatic(member) {
				continue
			}
			if key, ok := noUnusedClassComponentMethodsKeyNode(member); ok {
				state.definitions = append(state.definitions, key)
			}
			state.collect(member)
		}
	}

	state.report(ctx, className, noUnusedClassComponentMethodsEs6Lifecycle)
}

// noUnusedClassComponentMethodsCheckObject gathers and reports for one createReactClass object.
//
// The object has no name, so every finding here uses the `unused` id.
func noUnusedClassComponentMethodsCheckObject(ctx rule.Context, object *ast.Node) {
	state := &noUnusedClassComponentMethodsState{used: map[string]bool{}}

	properties := object.AsObjectLiteralExpression().Properties
	if properties != nil {
		for _, property := range properties.Nodes {
			if key, ok := noUnusedClassComponentMethodsKeyNode(property); ok {
				state.definitions = append(state.definitions, key)
			}
			state.collect(property)
		}
	}

	state.report(ctx, "", noUnusedClassComponentMethodsEs5Lifecycle)
}

// noUnusedClassComponentMethodsState is one component's gathered definitions and uses.
type noUnusedClassComponentMethodsState struct {
	// definitions holds the KEY node of each declaration, which is also the reported span.
	definitions []*ast.Node
	used        map[string]bool
}

// collect walks one member's body recording uses and assignment definitions.
//
// A nested class is NOT skipped, which matches upstream: its `classInfo` is replaced when the inner
// class declaration is entered and never restored, so the outer class stops collecting at that
// point. That is a defect rather than a design, and reproducing it exactly would mean modelling a
// single mutable slot; measured, the observable difference is that an inner component's members are
// judged and the outer class's remaining members are not. This walk judges each component
// separately and therefore judges MORE than upstream on a class containing a nested component.
//
// The divergence is stated rather than hidden because it is the one place this port is deliberately
// wider. A class declaring a component inside one of its methods is rare, and the alternative is
// reproducing a bug that silently drops findings. `TestNoUnusedClassComponentMethodsNestedClass`
// records both verdicts.
func (s *noUnusedClassComponentMethodsState) collect(node *ast.Node) {
	if node == nil {
		return
	}

	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		if object, property := noUnusedClassComponentMethodsAccessParts(node); isThisExpression(object) {
			if name, ok := noUnusedClassComponentMethodsStaticName(node, property); ok {
				if noUnusedClassComponentMethodsIsAssignmentTarget(node) {
					s.definitions = append(s.definitions, property)
				} else {
					s.used[name] = true
				}
			}
		}
	}

	if node.Kind == ast.KindVariableDeclaration {
		declaration := node.AsVariableDeclaration()
		if isThisExpression(declaration.Initializer) && declaration.Name() != nil &&
			declaration.Name().Kind == ast.KindObjectBindingPattern {
			for _, element := range declaration.Name().AsBindingPattern().Elements.Nodes {
				if element.Kind != ast.KindBindingElement {
					continue
				}
				binding := element.AsBindingElement()
				key := binding.PropertyName
				if key == nil {
					key = binding.Name()
				}
				if name, ok := noUnusedClassComponentMethodsLiteralName(key); ok {
					s.used[name] = true
				}
			}
		}
	}

	node.ForEachChild(func(child *ast.Node) bool {
		s.collect(child)
		return false
	})
}

// report emits one finding per definition whose name nothing used.
func (s *noUnusedClassComponentMethodsState) report(ctx rule.Context, className string, lifecycle map[string]bool) {
	for _, key := range s.definitions {
		name, ok := noUnusedClassComponentMethodsLiteralName(key)
		if !ok {
			continue
		}
		if s.used[name] || noUnusedClassComponentMethodsLifecycle[name] || lifecycle[name] {
			continue
		}

		if className == "" {
			ctx.ReportNode(key, rule.Message{
				Id:          "unused",
				Description: noUnusedClassComponentMethodsDescription(name, ""),
			})
			continue
		}
		ctx.ReportNode(key, rule.Message{
			Id:          "unusedWithClass",
			Description: noUnusedClassComponentMethodsDescription(name, className),
		})
	}
}

// noUnusedClassComponentMethodsDescription renders both message texts.
//
// Upstream's two strings are `Unused method or property "{{name}}"` and the same with
// ` of class "{{className}}"` appended. That wording is kept verbatim as the opening sentence, with
// the reason appended after it, and the fixtures assert the opening against literals so a format
// slot that interpolated the wrong value fails loudly.
func noUnusedClassComponentMethodsDescription(name string, className string) string {
	subject := fmt.Sprintf("Unused method or property %q", name)
	if className != "" {
		subject = fmt.Sprintf("Unused method or property %q of class %q", name, className)
	}
	return subject + ". Nothing inside this component reads it, so it is either dead code or a " +
		"call site that was renamed and left this behind. React calls the lifecycle methods " +
		"itself and those are exempt; anything else has to be reached from within the component."
}

// noUnusedClassComponentMethodsIsStatic reports whether a class member carries `static`.
func noUnusedClassComponentMethodsIsStatic(member *ast.Node) bool {
	modifiers := member.Modifiers()
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

// noUnusedClassComponentMethodsKeyNode returns a declaration's key when the key is literal-like.
//
// Upstream's `isKeyLiteralLike` accepts a Literal, a template with no substitutions, and an
// Identifier that is NOT computed. A computed identifier (`[someVariable]() {}`) is a name only
// known at run time, so it contributes nothing. A private name (`#secret`) has its own node kind and
// is invisible to upstream, which is why it never reports; measured.
func noUnusedClassComponentMethodsKeyNode(member *ast.Node) (*ast.Node, bool) {
	switch member.Kind {
	case ast.KindMethodDeclaration, ast.KindPropertyDeclaration, ast.KindGetAccessor,
		ast.KindSetAccessor, ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
	default:
		return nil, false
	}

	name := member.Name()
	if name == nil {
		return nil, false
	}
	if name.Kind == ast.KindComputedPropertyName {
		inner := name.AsComputedPropertyName().Expression
		if inner == nil {
			return nil, false
		}
		// A computed key counts only when what is inside is a LITERAL. Upstream's
		// `isKeyLiteralLike` is three arms and the identifier arm requires `node.computed === false`,
		// so `['handleClick']() {}` counts and `[foo]() {}` does not: the second names something
		// known only at run time.
		//
		// This was written accepting an identifier here and upstream's own clean case
		// `class ClassComputedMemberTest ... { [foo]() {} }` caught it, which is exactly what the
		// imported clean cases are for. `[`x`]` counts too, through the template arm.
		switch inner.Kind {
		case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
			return inner, true
		}
		return nil, false
	}
	if _, ok := noUnusedClassComponentMethodsLiteralName(name); ok {
		return name, true
	}
	return nil, false
}

// noUnusedClassComponentMethodsLiteralName renders a key node's name, if it has a static one.
//
// Upstream's `getName`: an Identifier gives its name, a Literal gives `String(value)`, and a
// template with no substitutions gives its raw text. Anything else answers null and is skipped.
func noUnusedClassComponentMethodsLiteralName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return node.Text(), true
	}
	return "", false
}

// noUnusedClassComponentMethodsAccessParts splits a member access into its object and its key node.
func noUnusedClassComponentMethodsAccessParts(node *ast.Node) (object *ast.Node, property *ast.Node) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		return access.Expression, access.Name()
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		return access.Expression, access.ArgumentExpression
	}
	return nil, nil
}

// noUnusedClassComponentMethodsStaticName renders an access's key when upstream would read it.
//
// A dotted access always has an Identifier key. A subscript is read only when it is literal-like,
// which is upstream's same `isKeyLiteralLike` test with `computed` true, so a literal subscript
// counts and a variable one does not.
func noUnusedClassComponentMethodsStaticName(access *ast.Node, property *ast.Node) (string, bool) {
	if property == nil {
		return "", false
	}
	if access.Kind == ast.KindElementAccessExpression && property.Kind == ast.KindIdentifier {
		// `this[someVariable]` names nothing known statically.
		return "", false
	}
	return noUnusedClassComponentMethodsLiteralName(property)
}

// noUnusedClassComponentMethodsIsAssignmentTarget reports whether an access is being written to.
//
// Upstream tests `parent.type === 'AssignmentExpression' && parent.left === node`, which is true for
// `=` and for every compound operator, and false for an update expression. So `this.foo += 1` is a
// DEFINITION and `this.foo++` is a USE, which reads backwards and is what upstream does. Both
// measured on the installed build and both pinned by fixture.
func noUnusedClassComponentMethodsIsAssignmentTarget(access *ast.Node) bool {
	parent := access.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	if binary.Left != access {
		return false
	}
	return ast.IsAssignmentOperator(binary.OperatorToken.Kind)
}
