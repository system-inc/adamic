package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	reactUtilities "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageRequireOptimizationNoShouldComponentUpdate = rule.Message{
	Id: "noShouldComponentUpdate",
	Description: "This class component re-renders whenever its parent does, even when nothing it " +
		"displays has changed. React cannot know that on its own, so the component has to say so: " +
		"add a `shouldComponentUpdate` method, or extend `React.PureComponent`, which compares " +
		"props and state shallowly for you.",
}

// RequireOptimizationOptions is the decoded option object.
//
// One key, matching `meta.schema` exactly. Its zero value is the empty list, which is upstream's
// default (`configuration.allowDecorators || []`), so no inversion is needed.
type RequireOptimizationOptions struct {
	// AllowDecorators names decorators that exempt a class from the rule.
	//
	// Matched against the decorator expression's own identifier text, so `@pure` exempts and
	// `@pure()` does NOT, because upstream reads `expression.name` and a call expression has none.
	// Measured on the installed build; both spellings are pinned by fixture.
	AllowDecorators []string `json:"allowDecorators"`
}

// DecodeRequireOptimizationOptions decodes the option object.
//
// Hand-written rather than `rule.DecodeOptionsInto` because the generic helper errors on EMPTY
// input, which is what a bare `"error"` configuration hands a rule, and upstream accepts that as
// the defaults. The one default here is the zero value, so nothing else needs translating.
func DecodeRequireOptimizationOptions(raw []byte) (any, error) {
	var options RequireOptimizationOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// RequireOptimization flags a class component that never declines a re-render.
//
//	valid:   class C extends React.Component { shouldComponentUpdate() {} }
//	valid:   class C extends React.PureComponent {}
//	valid:   function C(props) { return <div />; }
//	valid:   @reactMixin.decorate(PureRenderMixin) class C extends Component {}
//	invalid: class C extends React.Component {}
//	invalid: @reactMixin.decorate(SomeOtherMixin) class C extends Component {}
//
// Ported from `react/require-optimization` in `eslint-plugin-react`, the implementation that defined
// it. Its 15 clean and 8 reporting cases were extracted mechanically from the clone by stubbing its
// `RuleTester`, then all 23 were run against the installed build, 7.37.5, through the ESLint Linter
// API with the TypeScript parser. The two authorities agreed on all 23.
//
// # The rule file is much wider than what it can actually report
//
// Upstream builds on `Components.detect`, its 959-line component registry, and then marks nearly
// every shape as already optimized: every `FunctionDeclaration`, `FunctionExpression` and
// `ArrowFunctionExpression` outside a class calls `markSCUAsDeclared` with the comment "Stateless
// Functional Components cannot be optimized (yet)". So the registry's function-component detection,
// which is the expensive half and which this tree does not have, decides NOTHING here: a function
// component is exempt whether or not anything detects it, and failing to detect one produces exactly
// the silence upstream produces.
//
// That is why this rule is portable and `boolean-prop-naming` is not, and the difference is worth
// stating because the two rule files look equally entangled. Measured shape by shape against the
// installed build, the whole reporting surface is:
//
//	class C extends React.Component {}      reports, no `shouldComponentUpdate` method
//	var C = class extends React.Component {} SILENT, only a ClassDeclaration is listened for
//	createReactClass({})                     reports, on the OBJECT, see below
//	function C() { return <div/>; }          SILENT, exempt by design
//	const C = () => <div/>;                  SILENT, exempt by design
//
// # The object arm is an upstream accident and is reproduced as one
//
// `createReactClass({})` reports and `createReactClass({ foo: function () {} })` does NOT, and the
// difference has nothing to do with `shouldComponentUpdate`. The `FunctionExpression` and
// `ArrowFunctionExpression` listeners mark their enclosing component as optimized, and for a
// `createReactClass` call that component IS the object, so ANY function value anywhere in the
// object exempts the whole thing. Measured:
//
//	createReactClass({})                                  reports
//	createReactClass({ foo: 1 })                          reports
//	createReactClass({ mixins: [RandomMixin] })           reports
//	createReactClass({ foo: function () { return 1; } })  SILENT, a function is present
//	createReactClass({ render() { return null; } })       SILENT, same reason
//	createReactClass({ mixins: [PureRenderMixin] })       SILENT, the mixin arm
//	createReactClass({ shouldComponentUpdate() {} })      SILENT, the method arm
//
// Nothing in upstream's documentation describes this and no corpus case distinguishes it: both of
// its reporting object cases (`{}` and `{ mixins: [RandomMixin] }`) contain no function, so the
// accident is invisible from the corpus alone. It is reproduced rather than corrected, because a
// port that reported on every real component object would report on almost every `createReactClass`
// in a codebase, where upstream reports on almost none. The intuitive reading is that this arm
// means "an object with no shouldComponentUpdate"; the actual one is "an object with no function
// value at all, and no PureRenderMixin".
var RequireOptimization = rule.Rule{
	Name: "react/require-optimization",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[RequireOptimizationOptions](options)
		if !ok {
			settings = RequireOptimizationOptions{}
		}

		return rule.Listeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				if !reactUtilities.IsEs6ComponentClass(node) {
					return
				}
				if requireOptimizationClassIsExempt(node, settings) {
					return
				}
				ctx.ReportRange(requireOptimizationClassRange(ctx, node),
					messageRequireOptimizationNoShouldComponentUpdate)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				if !requireOptimizationIsCreateReactClassCall(node) {
					return
				}
				arguments := node.AsCallExpression().Arguments
				if arguments == nil || len(arguments.Nodes) == 0 {
					return
				}
				specification := arguments.Nodes[0]
				if specification.Kind != ast.KindObjectLiteralExpression {
					return
				}
				if requireOptimizationObjectIsExempt(specification) {
					return
				}
				ctx.ReportNode(specification, messageRequireOptimizationNoShouldComponentUpdate)
			},
		}
	},
}

// requireOptimizationClassIsExempt answers the three ways a class escapes the rule.
//
// A `PureComponent` base, a `reactMixin.decorate(PureRenderMixin)` decorator, or a decorator whose
// bare name is in `allowDecorators`. Upstream also exempts a class carrying a `shouldComponentUpdate`
// METHOD, through a separate `MethodDefinition` listener; that is folded in here because our walk
// gives the class its members directly and a second listener would have to resolve back to the
// enclosing class.
func requireOptimizationClassIsExempt(node *ast.Node, settings RequireOptimizationOptions) bool {
	if requireOptimizationExtendsPureComponent(node) {
		return true
	}
	if requireOptimizationHasExemptingDecorator(node, settings) {
		return true
	}
	return requireOptimizationDeclaresShouldComponentUpdate(node)
}

// requireOptimizationExtendsPureComponent reports whether the class extends a PureComponent base.
//
// Upstream renders the superclass to TEXT and tests it against `^(React\.)?PureComponent$`, so
// exactly two spellings pass and `Foo.PureComponent` does not. Measured: `Foo.PureComponent` reports
// and both accepted spellings are silent.
//
// The parentheses question was measured rather than assumed, because the shelf's
// `IsEs6ComponentClass` skips them and the port brief records that skip as a divergence on a
// different rule. Here it is not one: `class C extends (React.Component) {}` REPORTS upstream, which
// means upstream considers it a component, and the shelf agreeing is the correct answer. The text
// comparison below does not skip, which matches upstream's `getText` on the raw superclass node, so
// `extends (React.PureComponent)` is NOT exempt in either implementation.
func requireOptimizationExtendsPureComponent(node *ast.Node) bool {
	heritage := node.AsClassDeclaration().HeritageClauses
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
			if requireOptimizationIsPureComponentBase(typeNode.AsExpressionWithTypeArguments().Expression) {
				return true
			}
		}
	}
	return false
}

// requireOptimizationIsPureComponentBase matches upstream's two accepted spellings.
//
// The pragma test in the PropertyAccess arm is SUBSUMED by the component gate above this function,
// and that is measured rather than argued. A mutant dropping it survived every fixture, so the
// reachability was probed directly: `IsEs6ComponentClass` accepts only `React.Component`,
// `React.PureComponent`, `Component` and `PureComponent`, so a class extending `Foo.PureComponent`
// is not a component at all and never reaches here. Probed over three heritage shapes, the arm was
// entered with the object `React` and nothing else.
//
// The line stays because the subsumption rests on a helper this rule does not own: widening
// `IsEs6ComponentClass` to accept another namespace would make this the only thing standing between
// `Foo.PureComponent` and a wrong exemption. The INVERSE mutation, making the test always fail, is
// caught by six failing lines, which is what proves the line is reached rather than dead.
func requireOptimizationIsPureComponentBase(expression *ast.Node) bool {
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindIdentifier:
		return expression.Text() == "PureComponent"

	case ast.KindPropertyAccessExpression:
		access := expression.AsPropertyAccessExpression()
		object := access.Expression
		if object == nil || object.Kind != ast.KindIdentifier || object.Text() != reactPragmaName {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "PureComponent"
	}
	return false
}

// requireOptimizationDeclaresShouldComponentUpdate reports whether the class has the method.
//
// A METHOD only, which is upstream's `MethodDefinition` listener, and this is narrower than it
// looks. Measured on the installed build, all four of these are the rule's actual answers:
//
//	shouldComponentUpdate() {}                          exempt, a method
//	static shouldComponentUpdate() {}                   exempt, static is not excluded
//	shouldComponentUpdate = () => true;                 REPORTS, a property is not a method
//	['shouldComponentUpdate']() {}                       REPORTS, upstream reads `key.name`
//
// The property spellings are the surprising pair: a class field holding an arrow function is how
// this method is commonly written today, and upstream does not accept it. Reproduced rather than
// widened, because widening would silence a class upstream reports and no imported fixture could
// see the difference.
func requireOptimizationDeclaresShouldComponentUpdate(node *ast.Node) bool {
	members := node.AsClassDeclaration().Members
	if members == nil {
		return false
	}
	for _, member := range members.Nodes {
		if member.Kind != ast.KindMethodDeclaration {
			continue
		}
		name := member.Name()
		if name != nil && name.Kind == ast.KindIdentifier && name.Text() == "shouldComponentUpdate" {
			return true
		}
	}
	return false
}

// requireOptimizationHasExemptingDecorator answers both decorator arms.
//
// The built-in one is `@reactMixin.decorate(PureRenderMixin)` and every part of it is compared:
// upstream tests the callee's object name, the callee's property name, and the first argument's
// name, so `@other.decorate(PureRenderMixin)` and `@reactMixin.other(PureRenderMixin)` both report.
// Measured, both of them.
//
// The configured one compares the decorator expression's own `name`, which only an Identifier has.
// So `@pure` exempts when `pure` is listed and `@pure()` does NOT, because a call expression has no
// `name` property in upstream's AST and the comparison is against `undefined`. Measured on the
// installed build; both are pinned by fixture.
func requireOptimizationHasExemptingDecorator(node *ast.Node, settings RequireOptimizationOptions) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind != ast.KindDecorator {
			continue
		}
		expression := modifier.AsDecorator().Expression
		if expression == nil {
			continue
		}
		if requireOptimizationIsPureRenderDecoration(expression) {
			return true
		}
		if expression.Kind == ast.KindIdentifier {
			for _, allowed := range settings.AllowDecorators {
				if expression.Text() == allowed {
					return true
				}
			}
		}
	}
	return false
}

// requireOptimizationIsPureRenderDecoration matches `reactMixin.decorate(PureRenderMixin)` exactly.
func requireOptimizationIsPureRenderDecoration(expression *ast.Node) bool {
	if expression.Kind != ast.KindCallExpression {
		return false
	}
	call := expression.AsCallExpression()
	callee := call.Expression
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	object := access.Expression
	if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "reactMixin" {
		return false
	}
	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "decorate" {
		return false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	argument := call.Arguments.Nodes[0]
	return argument.Kind == ast.KindIdentifier && argument.Text() == "PureRenderMixin"
}

// requireOptimizationObjectIsExempt answers the object arm, accident included.
//
// Three ways out, and only two of them are about optimization. A `shouldComponentUpdate` property
// and a `mixins` array containing `PureRenderMixin` are upstream's two deliberate exemptions. The
// third, any function value anywhere in the object, is the accident described on the rule; it is
// reproduced deliberately and its reasoning lives there.
func requireOptimizationObjectIsExempt(object *ast.Node) bool {
	properties := object.AsObjectLiteralExpression().Properties
	if properties == nil {
		return false
	}
	for _, property := range properties.Nodes {
		if requireOptimizationPropertyIsNamed(property, "shouldComponentUpdate") {
			return true
		}
		if requireOptimizationIsPureRenderMixinsProperty(property) {
			return true
		}
		if requireOptimizationPropertyHoldsAFunction(property) {
			return true
		}
	}
	return false
}

// requireOptimizationPropertyIsNamed reports whether an object member has this identifier key.
//
// Upstream reads `property.key.name`, which only an Identifier has, so a string-literal or computed
// key is invisible and the object reports as if the property were absent.
func requireOptimizationPropertyIsNamed(property *ast.Node, wanted string) bool {
	name := property.Name()
	return name != nil && name.Kind == ast.KindIdentifier && name.Text() == wanted
}

// requireOptimizationIsPureRenderMixinsProperty matches `mixins: [..., PureRenderMixin, ...]`.
func requireOptimizationIsPureRenderMixinsProperty(property *ast.Node) bool {
	if !requireOptimizationPropertyIsNamed(property, "mixins") {
		return false
	}
	if property.Kind != ast.KindPropertyAssignment {
		return false
	}
	value := property.AsPropertyAssignment().Initializer
	if value == nil || value.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	elements := value.AsArrayLiteralExpression().Elements
	if elements == nil {
		return false
	}
	for _, element := range elements.Nodes {
		if element.Kind == ast.KindIdentifier && element.Text() == "PureRenderMixin" {
			return true
		}
	}
	return false
}

// requireOptimizationPropertyHoldsAFunction reports whether a member's value is a function.
//
// This is the accident arm. Upstream reaches it through its `FunctionExpression` and
// `ArrowFunctionExpression` listeners marking the enclosing component, not through any test on the
// object, so the three shapes that count are exactly the three those listeners see: a method
// shorthand, a function expression, and an arrow. A property holding a function DECLARATION is not
// expressible in an object literal, so there is no fourth.
func requireOptimizationPropertyHoldsAFunction(property *ast.Node) bool {
	if property.Kind == ast.KindMethodDeclaration {
		return true
	}
	if property.Kind != ast.KindPropertyAssignment {
		return false
	}
	value := property.AsPropertyAssignment().Initializer
	if value == nil {
		return false
	}
	return value.Kind == ast.KindFunctionExpression || value.Kind == ast.KindArrowFunction
}

// requireOptimizationClassRange is the span upstream reports, which excludes `export`.
//
// Upstream's parser puts `export` on an ExportNamedDeclaration that WRAPS the class, so the class
// node it reports on begins at the `class` keyword, or at the first decorator when there is one.
// Measured on the installed build:
//
//	export class C extends React.Component {}           reported at column 8, the `class` keyword
//	export default class C extends React.Component {}   reported at column 16, likewise
//	@reactMixin.decorate(SomeOtherMixin)\nclass C ...    reported from the DECORATOR, not the class
//
// Our parser makes `export` and `default` modifiers ON the class node, so `ctx.ReportNode` would
// include them and the finding would point at a wider span than upstream's. Decorators are also
// modifiers here, and upstream DOES include those, so the two cannot be handled the same way: this
// walks the modifier list and starts at the first decorator if one exists, otherwise at the first
// token that is not a modifier.
//
// A finding pointing at the wrong span is invisible to every message-id fixture, which is why
// `TestRequireOptimizationSpans` asserts both arrangements.
func requireOptimizationClassRange(ctx rule.Context, node *ast.Node) core.TextRange {
	full := rule.TokenRange(ctx.SourceFile, node)

	modifiers := node.Modifiers()
	if modifiers == nil {
		return full
	}

	// Our modifier list is in SOURCE ORDER, probed over all three arrangements, so the span is
	// decided by walking it once and keeping only the leading run of decorators.
	//
	// A decorator is part of the class upstream and a keyword is not, and ORDER decides which:
	// upstream's parser attaches a decorator written BEFORE `export` to the wrapping export
	// declaration rather than to the class, so it falls outside the reported node there. Measured on
	// the installed build, all four arrangements:
	//
	//	@bar class C ...          reported from `@bar`
	//	@a @b @c class C ...      reported from `@a`
	//	@bar export class C ...   reported from `class`, the decorator is NOT included
	//	export @bar class C ...   reported from `@bar`
	//
	// The span therefore begins at the first decorator that has no keyword AFTER it: a decorator
	// followed by `export` belongs to the export declaration upstream, while one written after every
	// keyword belongs to the class. Scanning from the end finds it in one pass, and stopping at the
	// last keyword handles the no-decorator case with the same walk.
	// Only `export` and `default` move upstream's start, because they are the only two its wrapping
	// ExportNamedDeclaration or ExportDefaultDeclaration consumes. Every other modifier stays on the
	// class node there and is inside the reported span. Measured:
	//
	//	@bar export abstract class C ...   reported from `abstract`, not from `class`
	//	export default @bar class C ...    reported from `@bar`
	//
	// So the walk skips the leading run of export keywords, and any decorator caught inside that
	// run, and starts at the first thing that is neither. Scanning from the END finds that boundary
	// in one pass; scanning forward returns at the first export keyword and lands in the middle of
	// the run, which differs on `export default @bar class`.
	for index := len(modifiers.Nodes) - 1; index >= 0; index-- {
		modifier := modifiers.Nodes[index]
		if modifier.Kind == ast.KindExportKeyword || modifier.Kind == ast.KindDefaultKeyword {
			// Everything at or before this belongs to upstream's export declaration, so the span
			// starts at the next token, which may be a decorator, another modifier, or `class`.
			if modifier.End() >= full.End() {
				return full
			}
			return core.NewTextRange(requireOptimizationScanPastTrivia(ctx.SourceFile, modifier.End()), full.End())
		}
	}

	// No export keyword at all, so every modifier is on the class upstream too and the node's own
	// start is already correct.
	return full
}

// requireOptimizationScanPastTrivia returns the first position at or after `from` that is not whitespace.
//
// A modifier's `End()` sits immediately after its keyword, so the class keyword begins after the
// space that follows. Written as a byte scan rather than through the scanner because the only thing
// between two modifiers, or between a modifier and the class keyword, is whitespace: a comment there
// is legal but upstream would include it too, since its span starts at the class node's own start
// and comments before that are leading trivia of the same node.
func requireOptimizationScanPastTrivia(sourceFile *ast.SourceFile, from int) int {
	text := sourceFile.Text()
	for from < len(text) {
		switch text[from] {
		case ' ', '\t', '\n', '\r':
			from++
		default:
			return from
		}
	}
	return from
}

// requireOptimizationIsCreateReactClassCall reports whether a call is `createReactClass(...)`.
//
// Written inline rather than calling the shelf's `react.IsEs5ComponentCall`, because that helper
// accepts THREE spellings this rule's upstream rejects. Measured on the installed build, all three
// are silent and only the bare `createReactClass` reports:
//
//	createReactClass({})       reports
//	createClass({})            SILENT on the shelf's helper this would be true
//	React.createClass({})      SILENT, likewise
//	Foo.createClass({})        SILENT, likewise for the namespace test
//	notCreateReactClass({})    SILENT, both agree
//
// The cause is upstream's `createClass` pragma, which defaults to `createReactClass` and is only
// `createClass` when a project configures `settings.react.createClass`. cohere reads no React
// settings, so the default is the only reachable value and the shelf's leniency is a divergence rather than a
// convenience. The shelf's own doc comment says a rule whose upstream reads its callee differently
// should read it inline here, and this is that rule.
//
// The namespaced form is not accepted at all, because upstream's `isES5Component` requires the
// property name to equal the createClass pragma and `React.createClass` fails that test for the
// same reason the bare `createClass` does.
func requireOptimizationIsCreateReactClassCall(node *ast.Node) bool {
	callee := node.AsCallExpression().Expression
	if callee == nil {
		return false
	}
	return callee.Kind == ast.KindIdentifier && callee.Text() == "createReactClass"
}
