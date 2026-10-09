package react

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	reactUtilities "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// noTyposStaticClassProperties are the four names whose casing this rule polices.
//
// Upstream's `STATIC_CLASS_PROPERTIES`. A property whose lowercased spelling matches one of these
// and whose exact spelling does not is a typo: `PropTypes` on a component is not `propTypes` and
// React silently ignores it, so the declaration reads as configuration and does nothing.
var noTyposStaticClassProperties = []string{"propTypes", "contextTypes", "childContextTypes", "defaultProps"}

// noTyposInstanceLifecycleMethods is upstream's `lifecycleMethods.instance`, name for name.
var noTyposInstanceLifecycleMethods = []string{
	"getDefaultProps",
	"getInitialState",
	"getChildContext",
	"componentWillMount",
	"UNSAFE_componentWillMount",
	"componentDidMount",
	"componentWillReceiveProps",
	"UNSAFE_componentWillReceiveProps",
	"shouldComponentUpdate",
	"componentWillUpdate",
	"UNSAFE_componentWillUpdate",
	"getSnapshotBeforeUpdate",
	"componentDidUpdate",
	"componentDidCatch",
	"componentWillUnmount",
	"render",
}

// noTyposStaticLifecycleMethods is upstream's `lifecycleMethods.static`.
//
// One name, and it carries two separate judgments: written without `static` it reports
// `staticLifecycleMethod`, and written with the wrong casing it reports `typoLifecycleMethod` like
// any other. A method spelled `getderivedstatefromprops` without `static` reports BOTH.
var noTyposStaticLifecycleMethods = []string{"getDerivedStateFromProps"}

// noTyposPropTypeNames is `Object.keys(require('prop-types'))`, verbatim from version 15.8.1.
//
// Three of these are not prop types at all. `checkPropTypes` and `resetWarningCache` are module
// helpers and `PropTypes` is the package's own self-reference, and they are accepted here because
// upstream takes `Object.keys` of the whole module rather than a curated list. So
// `PropTypes.checkPropTypes` in a propTypes object is silent upstream, which is an accident rather
// than a decision, and it is reproduced because widening or narrowing the set would change verdicts
// upstream does not change. Measured against the installed build.
var noTyposPropTypeNames = []string{
	"array", "bigint", "bool", "func", "number", "object", "string", "symbol",
	"any", "arrayOf", "element", "elementType", "instanceOf", "node", "objectOf",
	"oneOf", "oneOfType", "shape", "exact",
	"checkPropTypes", "resetWarningCache", "PropTypes",
}

// NoTypos flags a React declaration whose casing means React will never read it.
//
//	valid:   class Foo extends React.Component { static propTypes = {}; }
//	valid:   class Foo extends React.Component { componentDidMount() {} }
//	invalid: class Foo extends React.Component { static PropTypes = {}; }
//	invalid: class Foo extends React.Component { ComponentDidMount() {} }
//	invalid: Foo.propTypes = { a: PropTypes.strng };
//
// Ported from `react/no-typos` in `eslint-plugin-react`, the implementation that defined it. Its 43
// clean and 53 reporting cases were extracted mechanically from the clone by stubbing its
// `RuleTester`, then all 96 were run against the installed build, 7.37.5, through the ESLint Linter
// API with the TypeScript parser. The two authorities agreed on 95; the one exception is a clone
// case the installed build cannot reproduce, recorded in the test.
//
// # Eight message ids, and each is a different judgment
//
//	typoStaticClassProp    `static PropTypes` on a class
//	typoPropDeclaration    the same casing error in a createReactClass object
//	typoLifecycleMethod    `ComponentDidMount` instead of `componentDidMount`
//	staticLifecycleMethod  `getDerivedStateFromProps` written without `static`
//	typoPropType           `PropTypes.strng`, a name prop-types does not export
//	typoPropTypeChain      `PropTypes.string.isRequird`, a qualifier that is not `isRequired`
//	noPropTypesBinding     `import 'prop-types'` with no local name
//	noReactBinding         `import 'react'` with no local name
//
// # The prop-type arms are inert until an import names the package
//
// `checkValidProp` returns immediately when neither package name is known, so a file that never
// imports `prop-types` or `react` produces no `typoPropType` or `typoPropTypeChain` finding however
// misspelled its propTypes are. That is upstream's guard, and it is why this rule tracks the two
// import bindings as state across the file rather than deciding per node.
//
// # What decides that something is a component, and why the shelf is enough here
//
// Three of the four arms gate on `componentUtil.isES6Component` or `isES5Component`, both of which
// the shelf answers. The fourth, the member-assignment arm, also accepts a non-class whose body
// returns JSX, which is upstream's `utils.isReturningJSX`. That predicate is narrow and is
// implemented directly here rather than reached for: measured against the installed build, it asks
// only whether a `return` in the function's OWN body returns JSX. There is no name gate, `null` does
// not count, and JSX inside a nested callback does not count. Every row is pinned by fixture.
//
// This is the difference from `default-props-match-prop-types`, which I stopped on in the same
// batch: that rule needs the registry's resolved `declaredPropTypes`, and this one needs a
// three-line predicate.
var NoTypos = rule.Rule{
	Name: "react/no-typos",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The two import bindings, tracked across the whole file. Upstream keeps them as closure
		// state and its listeners run in source order, so an import below a usage does not name the
		// package for that usage. Reproduced by walking the file once from the source node, in
		// order, rather than by registering per-kind listeners whose relative order is not stated.
		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				state := &noTyposState{ctx: ctx}
				state.walk(sourceFile)
			},
		}
	},
}

// noTyposState carries the two import bindings while the file is walked.
type noTyposState struct {
	ctx rule.Context

	// propTypesPackageName is the local name `prop-types` was imported under, or the name a
	// `PropTypes` specifier from `react` was bound to.
	propTypesPackageName string

	// reactPackageName is the local name `react` was imported under.
	reactPackageName string
}

// walk visits the file in source order, updating the bindings and checking each node.
//
// Source order matters and is not incidental: upstream's `ImportDeclaration` listener sets the
// binding as ESLint reaches it, so a `propTypes` object written ABOVE the import sees no binding and
// its prop-type arms stay silent. Pinned by fixture.
func (s *noTyposState) walk(node *ast.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindImportDeclaration:
		s.recordImport(node)

	case ast.KindPropertyDeclaration:
		s.checkStaticClassProperty(node)

	case ast.KindMethodDeclaration:
		s.checkClassMethod(node)

	case ast.KindPropertyAccessExpression:
		s.checkMemberAssignment(node)

	case ast.KindObjectLiteralExpression:
		s.checkEs5Specification(node)
	}

	node.ForEachChild(func(child *ast.Node) bool {
		s.walk(child)
		return false
	})
}

// recordImport updates the package bindings from one import declaration.
//
// Upstream reads `node.specifiers[0].local.name` for both packages, which is the FIRST specifier
// whatever its kind, and then separately looks for a named `PropTypes` specifier in a react import.
// A side-effect import with no specifiers is the reporting case for both packages.
func (s *noTyposState) recordImport(node *ast.Node) {
	declaration := node.AsImportDeclaration()
	moduleSpecifier := declaration.ModuleSpecifier
	if moduleSpecifier == nil || moduleSpecifier.Kind != ast.KindStringLiteral {
		return
	}

	specifiers := noTyposImportSpecifiers(node)

	switch moduleSpecifier.Text() {
	case "prop-types":
		if len(specifiers) == 0 {
			s.ctx.ReportNode(node, rule.Message{
				Id: "noPropTypesBinding",
				Description: "This imports `prop-types` for its side effects and never names it, " +
					"so nothing in the file can reference a prop type. The package has no side " +
					"effects worth importing for; write `import PropTypes from 'prop-types'`.",
			})
			return
		}
		s.propTypesPackageName = specifiers[0].localName

	case "react":
		if len(specifiers) == 0 {
			s.ctx.ReportNode(node, rule.Message{
				Id: "noReactBinding",
				Description: "This imports `react` for its side effects and never names it, so " +
					"nothing in the file can reference React. Write " +
					"`import React from 'react'`, or drop the import if it is genuinely unused.",
			})
			return
		}
		s.reactPackageName = specifiers[0].localName
		// A `PropTypes` named specifier from react binds the prop-types package too, which is the
		// pre-15.5 spelling `import { PropTypes } from 'react'`.
		for _, specifier := range specifiers {
			if specifier.importedName == "PropTypes" {
				s.propTypesPackageName = specifier.localName
			}
		}
	}
}

// noTyposImportSpecifier is one imported binding, with the name it was imported and bound under.
type noTyposImportSpecifier struct {
	// importedName is the name on the package's side, empty for a default or namespace import.
	importedName string

	// localName is the name the file uses.
	localName string
}

// noTyposImportSpecifiers flattens an import declaration into upstream's flat specifier list.
//
// Upstream's ESTree gives one array holding the default specifier, the namespace specifier and every
// named specifier together, and this rule reads `specifiers[0]` and searches the rest, so the ORDER
// matters and the flattening is this rule's business rather than the shelf's. Our parser splits the
// three shapes across separate fields, which `imports.BindingsOf` reads; this reassembles them in
// upstream's order, default first and then namespace or named.
//
// The three shapes come from the shelf rather than from a fourth inline copy of the same walk. That
// helper was lifted precisely because two rule packages held near-identical copies and three
// research passes each reported it absent, and this rule nearly became the third copy: it was
// written inline first and `TestRulePackagesDoNotReachPastWrappedAccessors` is what found it.
func noTyposImportSpecifiers(declaration *ast.Node) []noTyposImportSpecifier {
	bindings := imports.BindingsOf(declaration)

	specifiers := []noTyposImportSpecifier{}
	if bindings.Default != nil && bindings.Default.Kind == ast.KindIdentifier {
		specifiers = append(specifiers, noTyposImportSpecifier{localName: bindings.Default.Text()})
	}
	if bindings.Namespace != nil {
		if name := bindings.Namespace.Name(); name != nil && name.Kind == ast.KindIdentifier {
			specifiers = append(specifiers, noTyposImportSpecifier{localName: name.Text()})
		}
	}
	for _, element := range bindings.Named {
		if element.Kind != ast.KindImportSpecifier {
			continue
		}
		specifier := element.AsImportSpecifier()
		local := specifier.Name()
		if local == nil || local.Kind != ast.KindIdentifier {
			continue
		}
		imported := local.Text()
		if specifier.PropertyName != nil && specifier.PropertyName.Kind == ast.KindIdentifier {
			imported = specifier.PropertyName.Text()
		}
		specifiers = append(specifiers,
			noTyposImportSpecifier{importedName: imported, localName: local.Text()})
	}
	return specifiers
}

// checkStaticClassProperty answers upstream's `ClassProperty, PropertyDefinition` listener.
//
// Static only, and only on a class the shelf calls a component.
func (s *noTyposState) checkStaticClassProperty(node *ast.Node) {
	if !noTyposIsStatic(node) {
		return
	}
	class := noTyposEnclosingClass(node)
	if class == nil || !reactUtilities.IsEs6ComponentClass(class) {
		return
	}
	s.reportCasingTypo(node.AsPropertyDeclaration().Initializer, node.Name(), true)
}

// checkClassMethod answers upstream's `MethodDefinition` listener.
func (s *noTyposState) checkClassMethod(node *ast.Node) {
	class := noTyposEnclosingClass(node)
	if class == nil || !reactUtilities.IsEs6ComponentClass(class) {
		return
	}
	s.reportLifecycleCasingTypo(node, node.Name(), noTyposIsStatic(node))
}

// checkMemberAssignment answers upstream's `MemberExpression` listener.
//
// This is the widest arm and the only one that accepts a non-class. It fires on `Foo.PropTypes = {}`
// where `Foo` is a class component OR a function whose body returns JSX.
func (s *noTyposState) checkMemberAssignment(node *ast.Node) {
	access := node.AsPropertyAccessExpression()
	property := access.Name()
	if property == nil || property.Kind != ast.KindIdentifier {
		return
	}
	if !noTyposMatchesStaticClassPropertyIgnoringCase(property.Text()) {
		return
	}

	// Upstream tests `node.parent.type === 'AssignmentExpression' && node.parent.right`, and that is
	// ALL it tests. There is no left-side check and no operator check, which reads as an oversight
	// and is what the rule does. Measured on the installed build, and this port had both extra
	// checks at first with a surviving mutant for each:
	//
	//	Foo.PropTypes = {}        reports, the ordinary case
	//	target = Foo.PropTypes    REPORTS, the member is on the RIGHT
	//	Foo.PropTypes ||= {}      REPORTS, a compound operator
	//	Foo.PropTypes             silent, no assignment at all
	//	doThing(Foo.PropTypes)    silent, likewise
	//	if (Foo.PropTypes === 1)  silent, a comparison is not an assignment
	//
	// The second row is the surprising one: reading a mis-cased property reports as if it had been
	// written. Reproduced rather than narrowed, because narrowing would silence inputs upstream
	// reports and the two mutants prove no fixture would have noticed.
	//
	// The `Right` value passed on is upstream's `node.parent.right` whichever side the member sits
	// on, so for the right-hand row it is the member itself, and `checkPropObject` declines it for
	// not being an object literal.
	//
	// The nil test on `Right` is EQUIVALENT rather than discriminating, measured rather than argued
	// after a mutant dropping it survived: our parser's error recovery synthesizes a missing
	// identifier rather than leaving the slot empty, probed over five shapes including
	// `Foo.PropTypes =`, `Foo.PropTypes = ;` and `Foo.PropTypes ||=`, all of which came back with a
	// KindIdentifier. It is kept because upstream tests it, because it costs nothing, and because
	// what it guards is a nil dereference one call later, which the walk would turn into a lost file
	// rather than a lost finding.
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return
	}
	binary := parent.AsBinaryExpression()
	if !ast.IsAssignmentOperator(binary.OperatorToken.Kind) || binary.Right == nil {
		return
	}

	target := access.Expression
	if target == nil || target.Kind != ast.KindIdentifier {
		return
	}
	if !s.namesAComponent(node, target.Text()) {
		return
	}

	s.reportCasingTypo(binary.Right, property, true)
}

// namesAComponent reports whether an identifier names something this rule treats as a component.
//
// Upstream calls `utils.getRelatedComponent`, which resolves the name through its registry, and then
// accepts it when it is an ES6 component class OR a non-class whose body returns JSX. Resolution is
// done here by searching the file for a declaration of that name, which is what the registry is
// doing for this shape: the assignment and the declaration are both in the file by construction,
// since a name declared elsewhere has no `Foo.propTypes = ...` here to check.
//
// The JSX test is upstream's `isReturningJSX` and it is narrower than it sounds. Measured against
// the installed build:
//
//	function Foo() { return <div/>; }         counts
//	const Foo = () => <div/>;                 counts, an implicit return
//	function Foo() { return null; }           does NOT count
//	function Foo() { return 'x'; }            does NOT count
//	function Foo() { }                        does NOT count
//	function foo() { return <div/>; }         counts, there is no name gate
//	function Foo() { items.map(() => <div/>); return null; }   does NOT count
//
// The last row is the one a port is most likely to get wrong: JSX inside a nested function belongs
// to that function, not to this one.
func (s *noTyposState) namesAComponent(from *ast.Node, name string) bool {
	declaration := noTyposFindDeclaration(from, name)
	if declaration == nil {
		return false
	}
	if declaration.Kind == ast.KindClassDeclaration {
		return reactUtilities.IsEs6ComponentClass(declaration)
	}
	return noTyposReturnsJsx(declaration)
}

// noTyposFindDeclaration searches the file for a declaration binding this name.
//
// A class, a function, or a variable whose initializer is a function. Written as a walk rather than
// through the checker because the question is syntactic: upstream resolves through its own registry
// of components it has seen in this file, never through a type graph, and a name imported from
// elsewhere has no declaration here for either implementation to find.
func noTyposFindDeclaration(from *ast.Node, name string) *ast.Node {
	sourceFile := from
	for sourceFile != nil && sourceFile.Kind != ast.KindSourceFile {
		sourceFile = sourceFile.Parent
	}
	if sourceFile == nil {
		return nil
	}

	var found *ast.Node
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || found != nil {
			return
		}
		switch node.Kind {
		case ast.KindClassDeclaration:
			if declarationName := node.AsClassDeclaration().Name(); declarationName != nil &&
				declarationName.Kind == ast.KindIdentifier && declarationName.Text() == name {
				found = node
				return
			}
		case ast.KindFunctionDeclaration:
			if declarationName := node.AsFunctionDeclaration().Name(); declarationName != nil &&
				declarationName.Kind == ast.KindIdentifier && declarationName.Text() == name {
				found = node
				return
			}
		case ast.KindVariableDeclaration:
			declaration := node.AsVariableDeclaration()
			declarationName := declaration.Name()
			if declarationName != nil && declarationName.Kind == ast.KindIdentifier &&
				declarationName.Text() == name && declaration.Initializer != nil {
				found = declaration.Initializer
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(sourceFile)
	return found
}

// noTyposReturnsJsx reports whether a function's OWN body returns JSX.
//
// A concise arrow body counts when it is JSX directly. Otherwise every `return` in the body counts,
// except those inside a nested function, which belong to that function instead. Reproduces
// upstream's `isReturningJSX` on every shape measured; the table is at `namesAComponent`.
func noTyposReturnsJsx(node *ast.Node) bool {
	var body *ast.Node
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		body = node.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		body = node.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		body = node.AsArrowFunction().Body
	default:
		return false
	}
	if body == nil {
		return false
	}

	// A concise arrow body is the returned expression itself.
	if body.Kind != ast.KindBlock {
		return noTyposIsJsxExpression(body)
	}

	found := false
	var walk func(inner *ast.Node)
	walk = func(inner *ast.Node) {
		if inner == nil || found {
			return
		}
		// A nested function's returns are its own, not this function's.
		switch inner.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
			return
		case ast.KindReturnStatement:
			if noTyposIsJsxExpression(inner.AsReturnStatement().Expression) {
				found = true
				return
			}
		}
		inner.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	body.ForEachChild(func(child *ast.Node) bool {
		walk(child)
		return false
	})
	return found
}

// noTyposIsJsxExpression reports whether an expression is a JSX element or fragment.
//
// Parentheses are unwrapped first, and they nest. Upstream's parser folds them away entirely, so
// `return (<div/>)` reaches its JSX test as a bare element, and without this loop eight of the
// corpus's own reporting cases went silent here: wrapping a returned element in parentheses is the
// ordinary way to write a multi-line component. `((<div/>))` measured too, which is why this is a
// loop rather than one step.
//
// Written as a loop rather than with `ast.SkipParentheses`, which dereferences its argument, and the
// thing being unwrapped here is a return argument that is legitimately absent for a bare `return;`.
func noTyposIsJsxExpression(expression *ast.Node) bool {
	for expression != nil && expression.Kind == ast.KindParenthesizedExpression {
		expression = expression.AsParenthesizedExpression().Expression
	}
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return true
	}
	return false
}

// checkEs5Specification answers upstream's `ObjectExpression` listener.
//
// Every property of a `createReactClass` object is checked for both a casing typo and a lifecycle
// typo, and a spread contributes neither.
func (s *noTyposState) checkEs5Specification(node *ast.Node) {
	if !noTyposIsCreateReactClassSpecification(node) {
		return
	}
	properties := node.AsObjectLiteralExpression().Properties
	if properties == nil {
		return
	}
	for _, property := range properties.Nodes {
		// Upstream filters spreads out here. The skip is SUBSUMED by `noTyposKeyName`, which
		// declines any key that is not an identifier, a string or a number, and a spread has no key
		// node at all: a mutant removing this line survived the whole suite, and so did one removing
		// it while ALSO teaching the key reader to answer for a spread, because the name it answers
		// matches no canonical name either way. Kept because upstream writes it and because it says
		// what the loop is doing; the equivalence rests on a second function's behavior rather than
		// on anything visible here.
		if property.Kind == ast.KindSpreadAssignment {
			continue
		}
		s.reportCasingTypo(noTyposPropertyValue(property), property.Name(), false)
		s.reportLifecycleCasingTypo(property, property.Name(), false)
	}
}

// noTyposIsCreateReactClassSpecification reports whether an object is a createReactClass body.
//
// Written inline rather than through the shelf's `react.IsEs5ComponentCall`, which accepts
// `createClass` in both spellings. Upstream's `isES5Component` compares the callee against the
// createClass PRAGMA, which is `createReactClass` with no React settings read here to change it, so
// `createClass` fails in either spelling. Measured on the installed build, all five:
//
//	createReactClass({...})          counts
//	React.createReactClass({...})    counts, the namespaced form of the same pragma
//	createClass({...})               SILENT, not the pragma
//	React.createClass({...})         SILENT, likewise
//	Other.createReactClass({...})    SILENT, the namespace must be React
//
// The namespaced row is the one this port got wrong first, and eleven corpus cases caught it: they
// are written `React.createReactClass` and the rule was silent on every one.
func noTyposIsCreateReactClassSpecification(object *ast.Node) bool {
	parent := object.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	callee := parent.AsCallExpression().Expression
	if callee == nil {
		return false
	}

	if callee.Kind == ast.KindIdentifier {
		return callee.Text() == "createReactClass"
	}
	if callee.Kind == ast.KindPropertyAccessExpression {
		access := callee.AsPropertyAccessExpression()
		object := access.Expression
		if object == nil || object.Kind != ast.KindIdentifier || object.Text() != reactPragmaName {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "createReactClass"
	}
	return false
}

// noTyposPropertyValue returns an object member's value expression, or nil.
func noTyposPropertyValue(property *ast.Node) *ast.Node {
	switch property.Kind {
	case ast.KindPropertyAssignment:
		return property.AsPropertyAssignment().Initializer
	case ast.KindMethodDeclaration:
		// A method shorthand has no separate value node. Upstream sees a FunctionExpression here,
		// which `checkValidPropObject` declines because it is not an ObjectExpression, so the value
		// arm contributes nothing either way.
		return nil
	}
	return nil
}

// reportCasingTypo answers upstream's `reportErrorIfPropertyCasingTypo`.
//
// Two things happen here and they are independent. A property NAMED exactly `propTypes`,
// `contextTypes` or `childContextTypes` has its object value walked for prop-type typos. And any
// property whose lowercased name matches one of the four static names while differing in case is
// itself a casing typo.
func (s *noTyposState) reportCasingTypo(value *ast.Node, key *ast.Node, isClassProperty bool) {
	name, ok := noTyposKeyName(key)
	if !ok {
		return
	}

	if name == "propTypes" || name == "contextTypes" || name == "childContextTypes" {
		s.checkPropObject(value)
	}

	for _, canonical := range noTyposStaticClassProperties {
		if strings.EqualFold(canonical, name) && canonical != name {
			if isClassProperty {
				s.ctx.ReportNode(key, rule.Message{
					Id: "typoStaticClassProp",
					Description: "This static property is spelled with the wrong casing, so React " +
						"never reads it. React matches these names exactly, which means the " +
						"declaration sits there looking like configuration and does nothing.",
				})
				return
			}
			s.ctx.ReportNode(key, rule.Message{
				Id: "typoPropDeclaration",
				Description: "This property is spelled with the wrong casing, so React never " +
					"reads it. React matches these names exactly, which means the declaration " +
					"sits there looking like configuration and does nothing.",
			})
			return
		}
	}
}

// reportLifecycleCasingTypo answers upstream's `reportErrorIfLifecycleMethodCasingTypo`.
//
// Two independent judgments again. A name matching the ONE static lifecycle method case-insensitively
// and written without `static` reports `staticLifecycleMethod`. And a name matching any lifecycle
// method case-insensitively while differing in case reports `typoLifecycleMethod`. A name can trip
// both, which upstream reports as two findings in that order.
func (s *noTyposState) reportLifecycleCasingTypo(node *ast.Node, key *ast.Node, isStatic bool) {
	name, ok := noTyposKeyName(key)
	if !ok {
		return
	}

	for _, method := range noTyposStaticLifecycleMethods {
		if !isStatic && strings.EqualFold(name, method) {
			s.ctx.ReportNode(node, rule.Message{
				Id: "staticLifecycleMethod",
				Description: fmt.Sprintf(
					"`%s` is a static lifecycle method and this one is not declared `static`, so "+
						"React never calls it. Add `static`.", name),
			})
		}
	}

	for _, method := range append(append([]string{}, noTyposInstanceLifecycleMethods...),
		noTyposStaticLifecycleMethods...) {
		if strings.EqualFold(method, name) && method != name {
			s.ctx.ReportNode(node, rule.Message{
				Id: "typoLifecycleMethod",
				Description: fmt.Sprintf(
					"`%s` is spelled with the wrong casing and React calls `%s`, so this method "+
						"never runs. React matches lifecycle names exactly.", name, method),
			})
		}
	}
}

// checkPropObject walks a propTypes object looking for prop-type typos.
func (s *noTyposState) checkPropObject(node *ast.Node) {
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return
	}
	properties := node.AsObjectLiteralExpression().Properties
	if properties == nil {
		return
	}
	for _, property := range properties.Nodes {
		s.checkProp(noTyposPropertyValue(property))
	}
}

// checkProp answers upstream's `checkValidProp`, including its recursion through shape and oneOfType.
//
// The leading guard is upstream's and it is load-bearing: with NEITHER package name known, every
// prop-type arm is silent, so a file that imports nothing produces no typo findings however
// misspelled it is.
func (s *noTyposState) checkProp(node *ast.Node) {
	if s.propTypesPackageName == "" && s.reactPackageName == "" {
		return
	}
	if node == nil {
		return
	}

	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		object := access.Expression
		property := access.Name()
		if object == nil || property == nil {
			return
		}

		if object.Kind == ast.KindPropertyAccessExpression {
			inner := object.AsPropertyAccessExpression()
			if s.isPropTypesPackage(inner.Expression) {
				// `PropTypes.myProp.isRequired`
				s.checkPropTypeName(inner.Name())
				s.checkPropTypeQualifier(property)
				return
			}
		}

		if s.isPropTypesPackage(object) {
			// `PropTypes.myProp`, unless the property IS the qualifier.
			if property.Kind == ast.KindIdentifier && property.Text() == "isRequired" {
				return
			}
			s.checkPropTypeName(property)
			return
		}

		if object.Kind == ast.KindCallExpression {
			s.checkPropTypeQualifier(property)
			s.checkPropCall(object)
		}
		return
	}

	if node.Kind == ast.KindCallExpression {
		s.checkPropCall(node)
	}
}

// checkPropCall answers upstream's `checkValidCallExpression`.
//
// `shape(...)` recurses into the object literal, `oneOfType([...])` recurses into each element, and
// every other call is left alone.
func (s *noTyposState) checkPropCall(node *ast.Node) {
	call := node.AsCallExpression()
	callee := call.Expression
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return
	}
	name := callee.AsPropertyAccessExpression().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return
	}
	first := call.Arguments.Nodes[0]

	switch name.Text() {
	case "shape":
		s.checkPropObject(first)

	case "oneOfType":
		if first.Kind != ast.KindArrayLiteralExpression {
			return
		}
		elements := first.AsArrayLiteralExpression().Elements
		if elements == nil {
			return
		}
		for _, element := range elements.Nodes {
			s.checkProp(element)
		}
	}
}

// isPropTypesPackage reports whether an expression names the prop-types package.
//
// Either the bare local binding, or `<react binding>.PropTypes`, which is the pre-15.5 spelling.
func (s *noTyposState) isPropTypesPackage(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier {
		return s.propTypesPackageName != "" && node.Text() == s.propTypesPackageName
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		name := access.Name()
		object := access.Expression
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "PropTypes" &&
			object != nil && object.Kind == ast.KindIdentifier &&
			s.reactPackageName != "" && object.Text() == s.reactPackageName
	}
	return false
}

// checkPropTypeName reports a prop type the package does not export.
func (s *noTyposState) checkPropTypeName(node *ast.Node) {
	if node == nil || node.Kind != ast.KindIdentifier {
		return
	}
	name := node.Text()
	if name == "" {
		return
	}
	for _, known := range noTyposPropTypeNames {
		if known == name {
			return
		}
	}
	s.ctx.ReportNode(node, rule.Message{
		Id: "typoPropType",
		Description: fmt.Sprintf(
			"`%s` is not a prop type the `prop-types` package exports, so this validator is "+
				"`undefined` and React validates nothing for the prop.", name),
	})
}

// checkPropTypeQualifier reports a chain qualifier that is not `isRequired`.
func (s *noTyposState) checkPropTypeQualifier(node *ast.Node) {
	if node == nil || node.Kind != ast.KindIdentifier {
		return
	}
	if node.Text() == "isRequired" {
		return
	}
	s.ctx.ReportNode(node, rule.Message{
		Id: "typoPropTypeChain",
		Description: fmt.Sprintf(
			"`isRequired` is the only qualifier a prop type chain accepts, and `%s` is not it, so "+
				"this reads as `undefined` and validates nothing.", node.Text()),
	})
}

// noTyposKeyName renders a property key when it has a static name.
//
// Upstream reads `key.name`, falling back to `key.value` for a literal, and declines a private name
// or a computed key whose value is not a string.
func noTyposKeyName(key *ast.Node) (string, bool) {
	if key == nil {
		return "", false
	}
	switch key.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral:
		return key.Text(), true
	case ast.KindComputedPropertyName:
		inner := key.AsComputedPropertyName().Expression
		if inner != nil && inner.Kind == ast.KindStringLiteral {
			return inner.Text(), true
		}
	}
	return "", false
}

// noTyposMatchesStaticClassPropertyIgnoringCase is upstream's first filter on the member arm.
func noTyposMatchesStaticClassPropertyIgnoringCase(name string) bool {
	for _, canonical := range noTyposStaticClassProperties {
		if strings.EqualFold(canonical, name) {
			return true
		}
	}
	return false
}

// noTyposIsStatic reports whether a class member carries `static`.
func noTyposIsStatic(member *ast.Node) bool {
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

// noTyposEnclosingClass returns the class a member belongs to.
//
// Upstream reaches `node.parent.parent`, which is the member's class body's class. Ours puts members
// directly on the class, so this is one step, guarded because the parser can hand back a partial
// chain on malformed input.
func noTyposEnclosingClass(member *ast.Node) *ast.Node {
	parent := member.Parent
	if parent == nil {
		return nil
	}
	if parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindClassExpression {
		return parent
	}
	return nil
}
