package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferStatelessFunctionOptions is the rule's option surface.
//
// Upstream's `meta.schema` declares one object with one boolean key defaulting to false, checked
// against the installed build rather than taken from the inventory column.
type PreferStatelessFunctionOptions struct {
	// IgnorePureComponents exempts a class extending PureComponent, on the reasoning that the
	// author chose that base for its shallow-compare and a function component cannot express it.
	IgnorePureComponents bool
}

type preferStatelessFunctionWireOptions struct {
	IgnorePureComponents bool `json:"ignorePureComponents"`
}

// DefaultPreferStatelessFunctionOptions is what an unconfigured rule uses.
//
// Upstream's default for `ignorePureComponents` is false, so Go's zero value happens to agree. That
// agreement is stated rather than relied on silently: a later upstream flip would make the generic
// decoder wrong here without anything failing.
func DefaultPreferStatelessFunctionOptions() PreferStatelessFunctionOptions {
	return PreferStatelessFunctionOptions{IgnorePureComponents: false}
}

// DecodePreferStatelessFunctionOptions is hand-rolled rather than `rule.DecodeOptionsInto`.
//
// The generic decoder ERRORS on empty input, and a rule configured as a bare `"error"` is handed
// nil options. Answering the documented default explicitly makes the unconfigured path a decision
// rather than a coincidence, which is the failure the port brief records costing a rule 3,407 inert
// registrations.
func DecodePreferStatelessFunctionOptions(raw []byte) (any, error) {
	options := DefaultPreferStatelessFunctionOptions()
	if len(raw) == 0 {
		return options, nil
	}
	var wire preferStatelessFunctionWireOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}
	options.IgnorePureComponents = wire.IgnorePureComponents
	return options, nil
}

var messagePreferStatelessFunction = rule.Message{
	Id: "componentShouldBePure",
	Description: "This class component uses nothing a class provides. It has a render method, and " +
		"beyond the members React reads as configuration it has no state, no lifecycle, no refs " +
		"and no other use of `this`, so the class body is ceremony around a single function of " +
		"its props. Written as a function it is shorter, it cannot accidentally grow instance " +
		"state, and React can treat it as a plain call. Rewrite it as a function component.",
}

// PreferStatelessFunction flags a class or factory component that could be a plain function.
//
//	valid:   const Foo = ({foo}) => <div>{foo}</div>;
//	valid:   class Foo extends React.Component { changeState() { this.setState({}); } render() {...} }
//	valid:   class Foo extends React.Component { render() { return <div>{this.bar}</div>; } }
//	invalid: class Foo extends React.Component { render() { return <div>{this.props.foo}</div>; } }
//	invalid: var Foo = createReactClass({ render: function() { return <div/>; } });
//
// Ported from `eslint-plugin-react`'s `prefer-stateless-function.js`. `meta.fixable` is absent, and
// there is nothing mechanical to write: converting a class to a function rewrites every `this.props`
// reference and picks a parameter shape, which is an authoring decision.
//
// # Upstream delegates to Components.detect, and almost none of that machinery is reachable here
//
// The rule's `create` is wrapped in `Components.detect`, which is a 959-line component registry
// whose expensive half is `getStatelessComponent`, 162 lines of positional analysis deciding
// whether a function is a component. **None of that half can produce a finding for this rule**,
// because the final filter is `isES5Component(node) || isES6Component(node)`. Measured on the
// installed build rather than argued from the source:
//
//	class Foo extends React.Component { render() { return <div/>; } }          REPORTS
//	var Foo = createReactClass({ render: function() { return <div/>; } });     REPORTS
//	function Foo() { return <div/>; }                                          SILENT
//	const Foo = () => <div/>;                                                  SILENT
//	class Foo { render() { return <div/>; } }                                   SILENT   no heritage
//	class Foo extends Bar { render() { return <div/>; } }                       SILENT   wrong base
//
// So the port anchors directly on the two shapes that can report and asks the questions upstream's
// registry would have answered. That is fidelity to the DECISION rather than to the mechanism, and
// it is why this rule is a few hundred lines instead of a substrate build. A sibling rule
// (`no-set-state`) genuinely needs the function half and is correctly not ported yet.
//
// # What upstream calls "other properties", which is an exemption list rather than a check
//
// `hasOtherProperties` walks every member and asks whether any is something OTHER than the five
// React reads as configuration. If one is, the class is doing something a function cannot, and the
// rule stays silent. So the five names below are the ALLOWED set and everything else exempts:
//
//	displayName          contextTypes         defaultProps
//	propTypes  (and a `props` member carrying a type annotation)
//	render
//	a constructor whose body is a redundant super call
//
// Measured, and the getter spellings matter because upstream reads a property NAME rather than a
// member kind, so `static get displayName()` and `static displayName =` are the same name:
//
//	static displayName = 'Foo';       REPORTS   allowed member
//	static get propTypes() {...}      REPORTS   allowed, read by name
//	static defaultProps = {...}       REPORTS   allowed
//	static contextTypes = {...}       REPORTS   allowed
//	other() {}                        SILENT    not in the allowed set
//	constructor() { super(); }        REPORTS   redundant super call is allowed
//	constructor() { doStuff(); }      SILENT    a real constructor body is not
//
// # The four independent disqualifiers, each measured
//
// Beyond the member list, upstream tracks four flags. Each one alone silences the class:
//
//	useThis            any `this.<x>` that is not props or context, or a `this` destructuring
//	                   binding a name other than props/context
//	useRef             a JSX `ref` attribute anywhere in the file
//	invalidReturn      a `render` returning something that is neither JSX nor null/false
//	useDecorators      any decorator on the class
//
// The `this` half is subtler than it reads and three shapes pin it:
//
//	return <div>{this.props.foo}</div>       REPORTS   props is exempt
//	return <div>{this['props'].foo}</div>    REPORTS   a string subscript reads as the same name
//	return <div>{this.bar}</div>             SILENT    any other member is a real use of this
//	return <div>{this['bar']}</div>          SILENT    same, through a subscript
//	return <div>{this[bar]}</div>            SILENT    a computed key has no readable name, and
//	                                                   upstream's `property.name || property.value`
//	                                                   yields undefined, which is not props
//	let {props:{foo}, bar} = this;           SILENT    `bar` is a name other than props/context
//	let {props:{foo}, context:{bar}} = this; REPORTS   both names are exempt
//
// # childContextTypes is asked about the OTHER side of a member access
//
// Upstream's `MemberExpression` handler has a second branch for a receiver that is NOT `this`: if
// the property is named `childContextTypes`, it resolves which component the receiver names and
// marks that component. This is upstream reaching for `Foo.childContextTypes = {}` written outside
// the class body, which is how the legacy child-context API was declared.
//
// A first draft declined this branch as needing a scope query, and an imported clean case caught it.
// The branch is narrower than `getRelatedComponent` suggests, and four measurements bound it:
//
//	class Foo ... }  Foo.childContextTypes = {};       SILENT    names this component
//	class Foo ... }  Bar.childContextTypes = {};       REPORTS   names something else
//	class Foo ... }  Foo.somethingElse = {};           REPORTS   wrong property
//	class Foo ... }  const x = y.childContextTypes;    REPORTS   wrong receiver
//
// So the question is whether the receiver names THIS component, which for a named class is a
// comparison against its own name rather than a resolution. Reproduced that way. The narrowing is
// stated because it is real: an aliased receiver (`const Alias = Foo; Alias.childContextTypes = {}`)
// resolves upstream and is silent, and is a false positive here. Measured zero occurrences of
// `childContextTypes` anywhere in this tree, against controls of 146 files for `useEffect` and 1,129
// for `className`, so nothing exercises even the faithful branch today.
//
// A `childContextTypes` member written INSIDE the class body is a separate path and is handled by
// the allowed-member list above, which does not contain it.
//
// # ignorePureComponents, and why it is not the same question as no-redundant-should-component-update
//
// The option exempts a class extending PureComponent. `react.IsEs6ComponentClass` accepts Component
// and PureComponent alike and is therefore the wrong predicate for the option, while being the RIGHT
// predicate for the outer gate. Both are used here, for their two different questions.
var PreferStatelessFunction = rule.Rule{
	Name: "react/prefer-stateless-function",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[PreferStatelessFunctionOptions](options)
		if !ok {
			settings = DefaultPreferStatelessFunctionOptions()
		}

		// A gather-then-judge shape, because the disqualifiers live anywhere in the file while the
		// finding anchors on the component. There is no rule.OnExit and the walk is pre-order, so
		// the whole judgment happens inside a KindSourceFile listener, which fires before its
		// children. `no_multi_comp.go` is the shipped example of this shape.
		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				// A JSX `ref` attribute anywhere in the file disqualifies every component in it.
				// That is upstream's own scope: `markRefAsUsed` sets the flag on the component the
				// registry currently holds, and with one component per file that is file-wide.
				fileUsesRef := preferStatelessFunctionFileUsesRef(sourceFile)

				var visit func(*ast.Node)
				visit = func(node *ast.Node) {
					if preferStatelessFunctionIsCandidate(node) {
						declaresChildContextTypes := preferStatelessFunctionDeclaresChildContextTypes(
							sourceFile, node,
						)
						if preferStatelessFunctionShouldReport(
							node, settings, fileUsesRef, declaresChildContextTypes,
						) {
							ctx.ReportNode(node, messagePreferStatelessFunction)
						}
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

// preferStatelessFunctionIsCandidate reports whether a node is one of the two shapes that can
// report.
//
// Upstream's final filter is `isES5Component || isES6Component`, and both predicates already live on
// the shelf. Measured that no other shape reports; see the rule's doc comment for the table.
func preferStatelessFunctionIsCandidate(node *ast.Node) bool {
	return react.IsEs6ComponentClass(node) || react.IsEs5ComponentCall(node)
}

// preferStatelessFunctionShouldReport applies the member list and the four disqualifiers.
func preferStatelessFunctionShouldReport(
	component *ast.Node,
	settings PreferStatelessFunctionOptions,
	fileUsesRef bool,
	declaresChildContextTypes bool,
) bool {
	if fileUsesRef {
		return false
	}
	if settings.IgnorePureComponents && preferStatelessFunctionExtendsPureComponent(component) {
		return false
	}
	if preferStatelessFunctionHasDecorator(component) {
		return false
	}
	if declaresChildContextTypes {
		return false
	}

	members := preferStatelessFunctionComponentMembers(component)
	if len(members) == 0 {
		// Upstream requires nothing here, but a component with no members has no render either,
		// and every reporting case in the corpus has one. Keeping the empty case silent matches
		// every measured input and avoids reporting on a shape nobody wrote.
		return false
	}
	for _, member := range members {
		if !preferStatelessFunctionIsAllowedMember(member) {
			return false
		}
	}

	if preferStatelessFunctionUsesThis(component) {
		return false
	}
	return !preferStatelessFunctionHasInvalidRender(component)
}

// preferStatelessFunctionComponentMembers returns the members upstream's `getComponentProperties`
// would return: a class body's members, or the properties of the object handed to the factory.
func preferStatelessFunctionComponentMembers(component *ast.Node) []*ast.Node {
	switch component.Kind {
	case ast.KindClassDeclaration:
		if list := component.AsClassDeclaration().Members; list != nil {
			return list.Nodes
		}
	case ast.KindClassExpression:
		if list := component.AsClassExpression().Members; list != nil {
			return list.Nodes
		}
	case ast.KindCallExpression:
		// `createReactClass({...})`. Upstream's `getComponentProperties` answers an ObjectExpression
		// with its properties, and the registry stored the object rather than the call, so the
		// members are the object literal's.
		arguments := component.AsCallExpression().Arguments
		if arguments == nil || len(arguments.Nodes) == 0 {
			return nil
		}
		first := preferStatelessFunctionUnwrapParentheses(arguments.Nodes[0])
		if first == nil || first.Kind != ast.KindObjectLiteralExpression {
			return nil
		}
		if list := first.AsObjectLiteralExpression().Properties; list != nil {
			return list.Nodes
		}
	}
	return nil
}

// preferStatelessFunctionUnwrapParentheses peels parenthesis nodes our parser keeps and upstream's
// folds away.
//
// Written as a loop rather than calling `ast.SkipParentheses`, because that helper dereferences its
// argument and the thing being unwrapped here is an argument that may be absent. That helper is how
// this project lost 167 files to a nil panic.
func preferStatelessFunctionUnwrapParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// preferStatelessFunctionMemberName reads a member's name the way upstream's `getPropertyName` does.
//
// That function returns `nameNode.name`, a field only an Identifier and a PrivateIdentifier carry,
// so a computed key yields the empty string and can never match an allowed name. Returning the empty
// string for those is the same decision reached the same way.
//
// A string-literal key is read too, because upstream's `this` handler falls back to
// `property.value`, and `this['props']` is measured REPORTING while `this['bar']` is silent. The
// member path and the `this` path share this reader so the two agree.
//
// **`Text()` is never called on an arbitrary name node.** It panics on a ComputedPropertyName, and
// the walk recovers per FILE rather than per rule, so one computed member would cost every rule in
// this package every finding in that file.
func preferStatelessFunctionMemberName(nameNode *ast.Node) string {
	if nameNode == nil {
		return ""
	}
	switch nameNode.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral:
		return nameNode.Text()
	}
	return ""
}

// preferStatelessFunctionIsAllowedMember reports whether a member is one React reads as
// configuration.
//
// This is upstream's `hasOtherProperties` inverted: it asks whether ANY member is something other
// than these, and stays silent if so. So this list is the allowed set and anything outside it
// exempts the component. Read by NAME rather than by member kind, which is why a getter and a static
// field with the same name are the same answer, measured on the installed build.
func preferStatelessFunctionIsAllowedMember(member *ast.Node) bool {
	if member == nil {
		return false
	}
	if member.Kind == ast.KindConstructor {
		return preferStatelessFunctionIsRedundantConstructor(member)
	}

	name := preferStatelessFunctionMemberName(member.Name())
	switch name {
	case "displayName", "contextTypes", "defaultProps", "render":
		return true
	case "propTypes":
		return true
	case "props":
		// Upstream allows a `props` member only when it carries a type annotation, which is the
		// Flow spelling `props: { name: string };`. A plain `props` member is a real field.
		return preferStatelessFunctionMemberHasTypeAnnotation(member)
	}
	return false
}

// preferStatelessFunctionMemberHasTypeAnnotation reports whether a class member declares a type.
func preferStatelessFunctionMemberHasTypeAnnotation(member *ast.Node) bool {
	if member.Kind != ast.KindPropertyDeclaration {
		return false
	}
	return member.AsPropertyDeclaration().Type != nil
}

// preferStatelessFunctionIsRedundantConstructor reproduces upstream's `isRedundantSuperCall`, which
// it borrowed from ESLint's `no-useless-constructor`.
//
// A constructor is allowed only when its whole body is one `super(...)` call that passes its own
// parameters straight through, because such a constructor is what a class writes when it has nothing
// to add. Anything else is real work a function component cannot do. Measured:
//
//	constructor(props) { super(props); }              REPORTS   passes through
//	constructor() { super(); }                        REPORTS   trivially passes through
//	constructor(props) { super(props); this.x = 1; }  SILENT    a second statement
//	constructor() { doSpecialStuffs(); }              SILENT    not a super call
//	constructor() {}                                  SILENT    an empty body is not a super call
func preferStatelessFunctionIsRedundantConstructor(member *ast.Node) bool {
	constructor := member.AsConstructorDeclaration()
	if constructor.Body == nil {
		return false
	}
	statements := constructor.Body.AsBlock().Statements
	if statements == nil || len(statements.Nodes) != 1 {
		return false
	}
	statement := statements.Nodes[0]
	if statement.Kind != ast.KindExpressionStatement {
		return false
	}
	call := preferStatelessFunctionUnwrapParentheses(statement.AsExpressionStatement().Expression)
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	callExpression := call.AsCallExpression()
	if callExpression.Expression == nil || callExpression.Expression.Kind != ast.KindSuperKeyword {
		return false
	}

	// Every parameter has to be a plain name or a rest element. Upstream's `isSimple` declines a
	// default and a destructuring pattern, because both can have side effects.
	var parameters []*ast.Node
	if constructor.Parameters != nil {
		parameters = constructor.Parameters.Nodes
	}
	for _, parameter := range parameters {
		declaration := parameter.AsParameterDeclaration()
		if declaration.Initializer != nil {
			return false
		}
		if declaration.Name() == nil || declaration.Name().Kind != ast.KindIdentifier {
			return false
		}
	}

	var superArguments []*ast.Node
	if callExpression.Arguments != nil {
		superArguments = callExpression.Arguments.Nodes
	}

	// `super(...arguments)` passes everything through whatever the parameters are.
	if len(superArguments) == 1 && superArguments[0].Kind == ast.KindSpreadElement {
		spread := superArguments[0].AsSpreadElement().Expression
		if spread != nil && spread.Kind == ast.KindIdentifier && spread.Text() == "arguments" {
			return true
		}
	}

	if len(parameters) != len(superArguments) {
		return false
	}
	for index, parameter := range parameters {
		declaration := parameter.AsParameterDeclaration()
		argument := superArguments[index]
		if declaration.DotDotDotToken != nil {
			if argument.Kind != ast.KindSpreadElement {
				return false
			}
			inner := argument.AsSpreadElement().Expression
			if inner == nil || inner.Kind != ast.KindIdentifier ||
				inner.Text() != declaration.Name().Text() {
				return false
			}
			continue
		}
		if argument.Kind != ast.KindIdentifier || argument.Text() != declaration.Name().Text() {
			return false
		}
	}
	return true
}

// preferStatelessFunctionExtendsPureComponent reports whether a class extends a PureComponent base.
//
// Deliberately NOT `react.IsEs6ComponentClass`, which accepts Component and PureComponent alike and
// would make `ignorePureComponents` exempt every class component. The two predicates answer
// different questions and both are used in this rule.
//
// Upstream tests the heritage's source text against `^(React\.)?PureComponent$`, so the accepted set
// is exactly two spellings.
func preferStatelessFunctionExtendsPureComponent(component *ast.Node) bool {
	var heritage *ast.NodeList
	switch component.Kind {
	case ast.KindClassDeclaration:
		heritage = component.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		heritage = component.AsClassExpression().HeritageClauses
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
			if isPureComponentBase(typeNode.AsExpressionWithTypeArguments().Expression) {
				return true
			}
		}
	}
	return false
}

// preferStatelessFunctionHasDecorator reports whether a class carries any decorator.
//
// Upstream reads `node.decorators && node.decorators.length` with no inspection of what the
// decorator is, so one bare `@foo` is enough. Three of upstream's valid cases are decorated classes.
func preferStatelessFunctionHasDecorator(component *ast.Node) bool {
	modifiers := component.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindDecorator {
			return true
		}
	}
	return false
}

// preferStatelessFunctionFileUsesRef reports whether any JSX `ref` attribute appears in the file.
//
// Upstream's handler sets the flag on whichever component the registry currently holds, which for a
// one-component file is the component being judged. File scope reproduces that for every input the
// corpus writes and for every shape in this tree.
//
// The name is compared exactly. Upstream renders the attribute name through `getText` and compares
// against `ref`, so a namespaced `a:ref` renders as `a:ref` and does not match.
func preferStatelessFunctionFileUsesRef(sourceFile *ast.Node) bool {
	found := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if found {
			return
		}
		if node.Kind == ast.KindJsxAttribute {
			if name := node.AsJsxAttribute().Name(); name != nil &&
				name.Kind == ast.KindIdentifier && name.Text() == "ref" {
				found = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(sourceFile)
	return found
}

// preferStatelessFunctionUsesThis reports whether the component uses `this` for anything beyond
// props and context.
//
// Two shapes, both of which upstream tracks separately and both of which are load-bearing.
//
// A member access on `this` is exempt only when the property reads as `props` or `context`.
// Upstream reads `node.property.name || node.property.value`, so an identifier and a string
// subscript both answer, while a computed key answers undefined and therefore is NOT exempt.
// Measured, and the last line is the one nobody would guess:
//
//	this.props.foo       exempt
//	this['props'].foo    exempt, the string subscript reads as the same name
//	this.bar             a real use of this
//	this['bar']          a real use of this
//	this[bar]            a real use of this, because the name cannot be read
//
// A destructuring of `this` is exempt only when EVERY bound name is props or context. Upstream's
// `some` asks whether any name is neither, so one stray name disqualifies the whole pattern:
//
//	let {props:{foo}, context:{bar}} = this;   exempt
//	let {props:{foo}, bar} = this;             a real use of this
func preferStatelessFunctionUsesThis(component *ast.Node) bool {
	found := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if found {
			return
		}
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			access := node.AsPropertyAccessExpression()
			if access.Expression != nil && access.Expression.Kind == ast.KindThisKeyword {
				if !preferStatelessFunctionIsPropsOrContext(preferStatelessFunctionMemberName(access.Name())) {
					found = true
					return
				}
			}
		case ast.KindElementAccessExpression:
			access := node.AsElementAccessExpression()
			if access.Expression != nil && access.Expression.Kind == ast.KindThisKeyword {
				// A computed key that is not a string literal reads as no name at all, which is
				// not props or context and therefore counts as a use of `this`.
				argument := preferStatelessFunctionUnwrapParentheses(access.ArgumentExpression)
				name := ""
				if argument != nil && argument.Kind == ast.KindStringLiteral {
					name = argument.Text()
				}
				if !preferStatelessFunctionIsPropsOrContext(name) {
					found = true
					return
				}
			}
		case ast.KindVariableDeclaration:
			declaration := node.AsVariableDeclaration()
			if declaration.Initializer == nil ||
				declaration.Initializer.Kind != ast.KindThisKeyword {
				break
			}
			name := declaration.Name()
			if name == nil || name.Kind != ast.KindObjectBindingPattern {
				break
			}
			for _, element := range name.AsBindingPattern().Elements.Nodes {
				if element.Kind != ast.KindBindingElement {
					continue
				}
				binding := element.AsBindingElement()
				// The PROPERTY name is what upstream reads, not the local binding, so
				// `{props: {foo}}` reads as `props` rather than `foo`.
				readable := binding.PropertyName
				if readable == nil {
					readable = binding.Name()
				}
				if !preferStatelessFunctionIsPropsOrContext(preferStatelessFunctionMemberName(readable)) {
					found = true
					return
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(component)
	return found
}

// preferStatelessFunctionIsPropsOrContext names the two members a function component can express.
func preferStatelessFunctionIsPropsOrContext(name string) bool {
	return name == "props" || name == "context"
}

// preferStatelessFunctionHasInvalidRender reports whether a render method returns something a
// function component could not return.
//
// Upstream's `ReturnStatement` handler walks scopes to find the enclosing method, checks its key is
// `render`, and then asks whether the returned value is JSX or null. React 15 and later allow a
// stateless component to return null, and this tree is on React 19, so the null branch is live.
//
// Measured on the installed build with the React version set to 19:
//
//	render() { return <div/>; }                REPORTS   JSX
//	render() { return null; }                  REPORTS   null is allowed since React 15
//	render() { return true ? <div/> : null; }  REPORTS   a conditional whose branches both qualify
//	render() { return 42; }                    SILENT    neither JSX nor null
//
// The conditional is upstream's `isReturningJSX` with `strict` false, which accepts a conditional
// when EITHER branch is JSX. The strict form is only used when null is disallowed, which this tree's
// React version never selects.
func preferStatelessFunctionHasInvalidRender(component *ast.Node) bool {
	invalid := false
	for _, member := range preferStatelessFunctionComponentMembers(component) {
		if preferStatelessFunctionMemberName(member.Name()) != "render" {
			continue
		}
		body := preferStatelessFunctionFunctionBody(member)
		if body == nil {
			continue
		}
		var visit func(*ast.Node)
		visit = func(node *ast.Node) {
			if invalid {
				return
			}
			// A nested function has its own returns and upstream's scope walk would not attribute
			// them to render, so the descent stops at one.
			if node != body && ast.IsFunctionLike(node) {
				return
			}
			if node.Kind == ast.KindReturnStatement {
				if !preferStatelessFunctionIsValidReturn(node.AsReturnStatement().Expression) {
					invalid = true
					return
				}
			}
			node.ForEachChild(func(child *ast.Node) bool {
				visit(child)
				return invalid
			})
		}
		visit(body)
	}
	return invalid
}

// preferStatelessFunctionFunctionBody returns the body of a render member, whichever way it is
// spelled.
func preferStatelessFunctionFunctionBody(member *ast.Node) *ast.Node {
	switch member.Kind {
	case ast.KindMethodDeclaration:
		return member.AsMethodDeclaration().Body
	case ast.KindPropertyAssignment:
		initializer := preferStatelessFunctionUnwrapParentheses(member.AsPropertyAssignment().Initializer)
		if initializer == nil {
			return nil
		}
		switch initializer.Kind {
		case ast.KindFunctionExpression:
			return initializer.AsFunctionExpression().Body
		case ast.KindArrowFunction:
			return initializer.AsArrowFunction().Body
		}
	case ast.KindMethodSignature:
		return nil
	}
	return nil
}

// preferStatelessFunctionIsValidReturn reports whether a returned expression is JSX or null.
//
// Upstream's `isReturningJSX` accepts a conditional or a logical when either side qualifies, and a
// sequence by its last expression. `isReturningNull` separately accepts the literals null and false.
// A bare `return;` has no argument and qualifies as neither, which upstream treats as invalid.
func preferStatelessFunctionIsValidReturn(expression *ast.Node) bool {
	expression = preferStatelessFunctionUnwrapParentheses(expression)
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return true
	case ast.KindNullKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindConditionalExpression:
		conditional := expression.AsConditionalExpression()
		return preferStatelessFunctionIsValidReturn(conditional.WhenTrue) ||
			preferStatelessFunctionIsValidReturn(conditional.WhenFalse)
	case ast.KindBinaryExpression:
		binary := expression.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return preferStatelessFunctionIsValidReturn(binary.Left) ||
				preferStatelessFunctionIsValidReturn(binary.Right)
		case ast.KindCommaToken:
			return preferStatelessFunctionIsValidReturn(binary.Right)
		}
	}
	return false
}

// preferStatelessFunctionDeclaresChildContextTypes reports whether the file assigns
// `childContextTypes` onto this component from outside its body.
//
// Upstream resolves the receiver through its component registry; this compares the receiver's name
// against the component's own name, which is the same answer for every shape measured. See the
// rule's doc comment for the four measurements and for the aliased-receiver narrowing.
//
// A class with no name cannot be matched by a receiver name, so it answers false, which is also
// what upstream does: `getRelatedComponent` has nothing to key on.
func preferStatelessFunctionDeclaresChildContextTypes(sourceFile *ast.Node, component *ast.Node) bool {
	name := component.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	componentName := name.Text()

	found := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if found {
			return
		}
		if node.Kind == ast.KindPropertyAccessExpression {
			access := node.AsPropertyAccessExpression()
			// The property must be exactly `childContextTypes` and the receiver must be an
			// identifier spelling this component's name. Upstream reads `node.property.name`, so a
			// computed or string subscript does not answer, and its receiver check is a registry
			// lookup that only a real component satisfies.
			if preferStatelessFunctionMemberName(access.Name()) == "childContextTypes" &&
				access.Expression != nil &&
				access.Expression.Kind == ast.KindIdentifier &&
				access.Expression.Text() == componentName {
				found = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(sourceFile)
	return found
}
