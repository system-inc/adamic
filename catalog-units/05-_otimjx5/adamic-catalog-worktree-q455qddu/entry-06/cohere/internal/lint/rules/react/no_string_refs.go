package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageThisRefsDeprecated = rule.Message{
	Id: "thisRefsDeprecated",
	Description: "`this.refs` is the string-ref registry and is deprecated. It only ever holds " +
		"refs that were attached by name, so it is empty in any component written the modern " +
		"way, and it is gone entirely in React 19. Attach the node to a field of your own from " +
		"a ref callback and read that field instead of going through `refs`.",
}

var messageStringInRefDeprecated = rule.Message{
	Id: "stringInRefDeprecated",
	Description: "A ref written as a string asks React to keep the node in a name-keyed registry " +
		"rather than handing it to you. The name is resolved against whichever component owns " +
		"the element, which breaks when the element is passed through as a child, and it cannot " +
		"be typed or renamed with any confidence. React 19 removes the behavior outright. Pass a " +
		"ref callback or a ref object instead.",
}

// NoStringRefsOptions is the decoded option object.
//
// `noTemplateLiterals` is upstream's whole option surface. `meta.schema` is the authority and it
// declares exactly `{noTemplateLiterals: boolean}` with `additionalProperties: false`, confirmed by
// handing the running rule a `{bogus: true}` option and watching ESLint refuse the configuration
// by name. A captured rule inventory recorded this rule as `"options": "no"`; the schema is the
// authority and the inventory was wrong, which is why that catalog is not consulted for options.
//
// `CheckThisRefs` has no upstream counterpart and exists because upstream's gate for that half is
// unreachable here. See the type doc on NoStringRefs.
type NoStringRefsOptions struct {
	// NoTemplateLiterals additionally reports a ref written as a template literal.
	//
	// Off by default upstream, and the default is the whole reason the option exists: a template
	// ref is silent unless it is asked for. Measured on the running rule at both settings and on
	// both template shapes, because those two shapes are the part of this rule most likely to be
	// got wrong quietly.
	NoTemplateLiterals bool

	// CheckThisRefs runs the `this.refs` half of the rule.
	//
	// Defaults to false, which is upstream's answer under this repository's configuration. See
	// the type doc on NoStringRefs for the measurement and for why the default is not simply
	// inherited from the option's name.
	CheckThisRefs bool
}

// noStringRefsWireOptions is the JSON shape, which is upstream's key set plus one.
//
// Separate from the decoded struct because `checkThisRefs` needs to be distinguishable as absent,
// and because a wire type keeps the rule's own field names free of the upstream spelling.
type noStringRefsWireOptions struct {
	NoTemplateLiterals bool  `json:"noTemplateLiterals"`
	CheckThisRefs      *bool `json:"checkThisRefs"`
}

// NoStringRefs flags the two halves of React's removed string-ref feature.
//
//	valid:   <div ref={c => { this.hello = c; }} />
//	valid:   var Hello = function() { return this.refs; };      (no enclosing component)
//	valid:   <div ref={`hello`} />                              (default; reported under the option)
//	valid:   class H extends React.Component { m() { return this.refs.hello; } }
//	                                                            (default; see the version gate below)
//	invalid: <div ref="hello" />
//	invalid: <div ref={'hello'} />
//	invalid: <div ref={`hello`} />                              (noTemplateLiterals: true)
//	invalid: class H extends React.Component { m() { return this.refs.hello; } }
//	                                                            (checkThisRefs: true)
//
// Ported from `react/no-string-refs` in `eslint-plugin-react`, which is the implementation that
// defined this rule. Every behavior recorded below was measured by driving the installed build,
// version 7.37.5, through the ESLint Linter API rather than read off the source, because the two
// disagreed about five separate inputs and reading was wrong about all five.
//
// # One rule name, two judgments, and they share almost nothing
//
// The name reads as a single judgment about ref attributes and it is not. Upstream carries two
// message ids, `thisRefsDeprecated` and `stringInRefDeprecated`, and the two halves answer to
// different gates: the `this.refs` half requires an enclosing component and is gated on the React
// version, while the attribute half reads the option, ignores the version, and requires no
// component at all. That asymmetry is measured. A bare `<div ref="hello" />` at the top level of a
// file, inside no component of any kind, reports at every React version, and `this.refs` in a plain
// function does not report at any of them.
//
// Reporting both under one id would pass a fixture set asserting counts, since every upstream input
// that reports twice reports once under each. The ids are what separate them.
//
// # The version gate, which decides whether half this rule runs at all
//
// `create` opens with `testReactVersion(context, '< 18.3.0')`, because `this.refs` became writable
// in React 18.3 and the deprecation no longer applies. When no version is configured,
// `getReactVersionFromContext` falls back to `ULTIMATE_LATEST_SEMVER`, the literal string
// `999.999.999` at `util/version.js:15`, so the predicate is false and the `this.refs` half never
// runs. Auto-detection from `node_modules/react` happens only when the configuration says
// `version: "detect"` explicitly; absence does not trigger it.
//
// Measured on the running rule, one class, five configurations: no settings reports nothing,
// `18.2.0` and `17.0.0` report, `18.3.0` and `19.0.0` report nothing.
//
// cohere reads no React settings. The loader reads `settings` for better-tailwindcss only and
// refuses `settings.react` by name, so no React version can reach a rule here by any route. The faithful reading of that is not "pick a version" but "reproduce the answer
// upstream gives under this repository's configuration", and that answer is silence: this tree
// configures no React version, and it runs React 19.2.8, where the deprecation is genuinely not in
// force.
//
// So the half is off by default and reachable by an option this rule adds. Shipping it always-on
// would report on every `this.refs` in the tree where upstream reports on none, and shipping it
// with no way to turn on would delete a judgment upstream still makes for anyone pinned below
// 18.3. The option is named for what it does rather than for a version, because a version string
// this rule cannot obtain would be a promise the config cannot keep.
//
// # There is no file gate, and the previous port had one
//
// This rule was ported once before from oxc, which declares `should_run` on
// `source_type().is_jsx()`. Upstream has no such gate: the same class reports identically under
// `probe.jsx`, `probe.js`, `probe.cjs` and `probe.mjs`. The gate is oxc's own and reproducing it
// costs findings in every `.ts` file, which is most of this tree.
//
// # `this.refs` is the dotted spelling and nothing wider
//
// `isRefsUsage` reads `node.property.name === 'refs'` off a `MemberExpression`, and a computed key
// has no `property.name` unless the key is itself an identifier. So the three computed spellings
// fall out in a way that reads as inconsistent and is exactly what upstream does:
//
//	this.refs        reports
//	this["refs"]     silent    a string literal key has no `.name`
//	this[`refs`]     silent    a template key has no `.name`
//	this[refs]       REPORTS   an identifier key has `.name === "refs"`, and `computed` is never
//	                           checked, so an unrelated variable named `refs` is read as the
//	                           registry
//
// All four measured. The last is an upstream defect and it is reproduced rather than corrected,
// because correcting it silently is the divergence nothing states. The previous oxc-based port had
// this exactly inverted: it reported the two literal-keyed spellings and was silent on the
// identifier-keyed one.
//
// The `refs` read is what reports, not the property taken off it: `this.refs.hello` underlines nine
// characters, so the finding points at `this.refs` and the outer access is left alone. A bare
// `this.refs` with nothing taken off it reports on its own.
//
// Parentheses are transparent here, and that is a parser difference rather than a decision.
// ESLint's ESTree has no parenthesized node, so `(this).refs` is a member expression whose object
// is a `ThisExpression` and it reports, with the finding underlining `(this).refs`. Our parser
// keeps the parenthesis as a real node, so reproducing upstream means skipping it deliberately.
// Measured; the previous port was silent on this input.
//
// # The enclosing component is found by SCOPE, not by walking parents
//
// `getParentES6Component` walks *scopes* to the first `class` scope and asks whether that one class
// is a component. It does not keep climbing. `getParentES5Component` walks scopes upward asking, at
// each one, whether the scope's block is the function argument of a factory call.
//
// The difference from a parent-chain walk is not cosmetic and it runs in both directions:
//
//	class H extends React.Component { m() { class Inner { k() { return this.refs.x; } } } }
//	                 silent upstream. The first class scope is `Inner`, which is not a component,
//	                 and the walk stops there rather than finding `H` above it.
//
//	class Outer { m() { class H extends React.Component { k() { return this.refs.x; } } } }
//	                 reports. The first class scope IS the component.
//
//	var H = createReactClass({ m: this.refs.x });
//	                 silent. The ES5 walk reads `scope.block.parent.parent` at each scope, and with
//	                 no function between the read and the module there is no scope whose block sits
//	                 inside the factory call.
//
// All three measured. `utilsreact.EnclosingComponent` is a parent-chain walk and answers the first
// two the wrong way, so the ES6 half is spelled here as "the nearest enclosing class, and is that
// one a component" rather than reached for from the shelf.
//
// A class field initializer and a static method both report, and an arrow function inside a method
// reports, because an arrow creates no class scope and the walk passes straight through it.
//
// # The ES5 factory is one name, and the shelf accepts three
//
// `isES5Component` keys on the `createClass` pragma, which defaults to `createReactClass`, and
// checks the callee is either that bare identifier or `<pragma>.createReactClass`. So:
//
//	createReactClass({...})          reports
//	React.createReactClass({...})    reports
//	React.createClass({...})         silent
//	createClass({...})               silent
//
// Measured. `utilsreact.IsEs5ComponentCall` accepts `createClass` and `React.createClass` as well,
// via `isCreateClassName`, so reaching for the shelf would report on two shapes upstream is silent
// on, and nothing in the imported corpus writes either. The predicate is spelled locally for that
// reason, which is the same resolution `no-unsafe` in this package reached for the same helper.
//
// `utilsreact.IsEs6ComponentClass` is used as-is, and that is measured rather than assumed. It
// skips parentheses on the heritage receiver, which is a divergence from oxc and is NOT one from
// this authority: `(React.Component)` and `(React).Component` both report upstream, because ESTree
// has no parenthesized node for the check to trip over. Its accepted base names, `Component` and
// `PureComponent` bare or under the React namespace, are `/^(Pure)?Component$/` at
// `componentUtil.js:113` exactly.
//
// The one thing the shelf cannot do is upstream's `isExplicitComponent`, which reads a
// `@extends React.Component` JSDoc tag through doctrine. Measured silent here rather than assumed:
// `/** @extends React.Component */ class H extends Foo {}` does not report on the running rule,
// because that path requires the `componentDetection` option most configurations leave off. Not
// implemented, and recorded here so the absence reads as measured rather than as forgotten.
//
// # The two template shapes, which is where this family of rules has gone wrong before
//
// A template ref has two node kinds in our AST and one in ESTree. “ref={`hello`}” parses as a
// `NoSubstitutionTemplateLiteral` and “ref={`hello${index}`}” as a `TemplateExpression`, where
// upstream sees `TemplateLiteral` for both. Upstream's corpus fails under the option on both
// shapes, so matching only the first would drop a fixture and matching only the second would drop
// two. Both are listed for that reason, and both were measured under the option and with it off.
//
// # What the attribute half will not touch
//
// `containsStringLiteral` requires `node.value.type === 'Literal'` with a string value, so a bare
// `<div ref />` has no value at all and is silent, and `ref={123}` is a Literal whose value is not
// a string. `ref={"a"+"b"}` is a `BinaryExpression` rather than a Literal and is silent even though
// its value is knowable. A spread supplying the ref is not a `JSXAttribute` and is never visited.
// The name comparison is exact, so `<div REF="hello" />` is a different attribute and is silent,
// and a namespaced `xlink:ref` is a `JSXNamespacedName` whose `.name` is an object rather than a
// string. All measured.
//
// The finding underlines the whole attribute, `ref="hello"`, which is the node handed to `report`,
// rather than the name span that `no-children-prop` in this same package reports.
var NoStringRefs = rule.Rule{
	// No namespace prefix. The config writes `react/no-string-refs` and the parity guard strips the
	// namespace on a `/` boundary, so `react-no-string-refs` would match no inventory entry and lint
	// no files while passing every fixture in this package.
	Name: "react/no-string-refs",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// An unconfigured rule is handed nil, and the zero value is upstream's answer under this
		// repository's configuration: template refs unreported, and the `this.refs` half off
		// because no configured React version means `999.999.999`.
		settings, _ := rule.OptionsAs[NoStringRefsOptions](options)

		listeners := rule.Listeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				// `isRefAttribute` reads `node.name.name === 'ref'`, which is a string only for a
				// plain identifier. A namespaced name is an object there and compares unequal, which
				// is what AttributeName's kind check reproduces. The comparison is exact rather than
				// case-insensitive.
				name, named := jsx.AttributeName(node)
				if !named || name != "ref" {
					return
				}
				if !isStringRefValue(node.AsJsxAttribute().Initializer, settings.NoTemplateLiterals) {
					return
				}
				// The whole attribute, not its name. Upstream passes `node` to `report`, and `node`
				// here is the JSXAttribute.
				ctx.ReportNode(node, messageStringInRefDeprecated)
			},
		}

		// The version gate is resolved once, before any node is visited, exactly as upstream
		// resolves `checkRefsUsage` once in `create`. A rule offered a file whose configuration
		// turns this half off should register nothing for it rather than register a listener that
		// declines every node.
		if !settings.CheckThisRefs {
			return listeners
		}

		listeners[ast.KindPropertyAccessExpression] = func(node *ast.Node) {
			access := node.AsPropertyAccessExpression()
			name := access.Name()
			if name == nil || name.Text() != "refs" {
				return
			}
			reportRefsUsage(ctx, node, access.Expression)
		}

		// A computed access is upstream's `MemberExpression` with `computed: true`, and it reaches
		// the same `node.property.name` read. Only an identifier key answers that read, so a string
		// or template key falls out here and reports nowhere. See the type doc: this asymmetry is
		// upstream's and the previous port had it inverted.
		listeners[ast.KindElementAccessExpression] = func(node *ast.Node) {
			access := node.AsElementAccessExpression()
			argument := access.ArgumentExpression
			if argument == nil || argument.Kind != ast.KindIdentifier || argument.Text() != "refs" {
				return
			}
			reportRefsUsage(ctx, node, access.Expression)
		}

		return listeners
	},
}

// reportRefsUsage is the tail of upstream's `isRefsUsage`, after the property name has matched.
//
// The two remaining questions are upstream's in upstream's order: is the object `this`, and is
// there an enclosing component. Both halves of the object check come from `node.object.type ===
// 'ThisExpression'`, where the parenthesis skip is what reproduces ESTree rather than what widens
// past it.
func reportRefsUsage(ctx rule.Context, node *ast.Node, object *ast.Node) {
	object = ast.SkipParentheses(object)
	if object == nil || object.Kind != ast.KindThisKeyword {
		return
	}
	if !hasEnclosingComponent(node) {
		return
	}
	ctx.ReportNode(node, messageThisRefsDeprecated)
}

// hasEnclosingComponent is upstream's `getParentES6Component(...) || getParentES5Component(...)`.
//
// Both halves are scope walks rather than parent-chain walks, and the difference decides real
// inputs in both directions. See the type doc on NoStringRefs for the three measured cases.
func hasEnclosingComponent(node *ast.Node) bool {
	return hasEnclosingEs6Component(node) || hasEnclosingEs5Component(node)
}

// hasEnclosingEs6Component reproduces `getParentES6Component`, whose walk STOPS at the first class.
//
// Upstream walks scopes to the nearest one whose type is `class` and then asks whether that single
// class is a component, returning null rather than continuing if it is not. So a plain class nested
// inside a component hides the component, and that silence is upstream's answer rather than a
// limitation of this port. A parent-chain walk gets it wrong, which is why this is not
// `utilsreact.EnclosingComponent`.
//
// Only a class declaration or class expression opens a class scope. A function or an arrow does
// not, so an arrow inside a component method still finds the component, and an ordinary method
// inside a nested plain class does not.
func hasEnclosingEs6Component(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return utilsreact.IsEs6ComponentClass(current)
		}
	}
	return false
}

// hasEnclosingEs5Component reproduces `getParentES5Component`, which walks FUNCTION scopes.
//
// Upstream's loop is four lines and every one of them matters:
//
//	node = scope.block && scope.block.parent && scope.block.parent.parent
//	if (node && isES5Component(node, context)) return node
//	scope = scope.upper
//
// `scope.block` is the function, `.parent` is the ESTree `Property` holding it, and `.parent.parent`
// is the `ObjectExpression`. That object is the node handed to `isES5Component`, which then reads
// `node.parent.callee` at `componentUtil.js:43`. So the question asked at each function scope is
// really "is this function a property of an object literal that is the argument of a factory call",
// and the chain from the function to the call is FOUR hops rather than two.
//
// Our parser spells the same chain `FunctionExpression -> PropertyAssignment ->
// ObjectLiteralExpression -> CallExpression`, read off the tree rather than assumed: an earlier
// draft of this walk counted two hops, and the three factory cases in the corpus went silent.
//
// The walk continues upward through every enclosing scope rather than stopping at the first, which
// is the opposite of the ES6 half and is why a function nested arbitrarily deep inside a factory
// argument still answers true. It answers on the enclosing function's own hops, not on the inner
// one's.
//
// The consequence worth naming is the shape with no function in it at all:
// `createReactClass({m: this.refs.x})` is silent, because the only scope enclosing that read is the
// module's and its block is the program rather than a function. Measured.
func hasEnclosingEs5Component(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if !opensFunctionScope(current.Kind) {
			continue
		}
		objectLiteral := enclosingObjectLiteralOfFunction(current)
		if objectLiteral == nil {
			continue
		}
		// `isES5Component` then reads `node.parent.callee`, which is the hop after the object.
		if isEs5FactoryCall(objectLiteral.Parent) {
			return true
		}
	}
	return false
}

// enclosingObjectLiteralOfFunction is upstream's `scope.block.parent.parent`, at our parser's shape.
//
// Upstream can count hops because ESTree spells every object member the same way: a `Property`
// whose `value` is a function, so the function's parent is always the `Property` and its
// grandparent is always the `ObjectExpression`. That uniformity is what makes `.parent.parent` a
// correct constant there.
//
// Our parser does not have it. A property holding a function expression or an arrow goes through a
// `PropertyAssignment`, which is two hops, but a shorthand method or an accessor IS the member, so
// the object literal is one hop up:
//
//	{ m: function() {} }    FunctionExpression -> PropertyAssignment -> ObjectLiteralExpression
//	{ m: () => {} }         ArrowFunction      -> PropertyAssignment -> ObjectLiteralExpression
//	{ m() {} }              MethodDeclaration  -> ObjectLiteralExpression
//	{ get m() {} }          GetAccessor        -> ObjectLiteralExpression
//
// Read off the tree rather than assumed, twice. A first draft counted two hops from the function to
// the object and the three function-expression factory cases in the corpus went silent; a second
// counted them correctly for `PropertyAssignment` and then lost the shorthand method and the getter,
// which is what the fixtures added for a surviving mutant caught.
//
// So the question is asked as "which object literal is this function a member of" rather than as a
// hop count, which is what upstream's read means and is stable across the two shapes our parser
// distinguishes and ESTree does not. Both spellings report on the installed build, measured.
func enclosingObjectLiteralOfFunction(function *ast.Node) *ast.Node {
	parent := function.Parent
	if parent == nil {
		return nil
	}
	// A shorthand method or an accessor is the object member itself.
	if parent.Kind == ast.KindObjectLiteralExpression {
		return parent
	}
	// A property whose value is a function expression or an arrow.
	if parent.Kind != ast.KindPropertyAssignment || parent.Parent == nil {
		return nil
	}
	if parent.Parent.Kind != ast.KindObjectLiteralExpression {
		return nil
	}
	return parent.Parent
}

// opensFunctionScope reports whether a node is one ESLint's scope analysis gives its own scope.
//
// Arrow functions are included because eslint-scope opens a scope for them, even though they do not
// rebind `this`. That matters only for the ES5 half: an arrow passed as a factory property is still
// a function scope whose block sits inside the call.
func opensFunctionScope(kind ast.Kind) bool {
	switch kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindConstructor:
		return true
	}
	return false
}

// isEs5FactoryCall is upstream's `isES5Component`, and its narrowness is the point.
//
// `componentUtil.js:39` reads the `createClass` pragma, which defaults to `createReactClass`, and
// accepts the bare identifier or `<pragma>.<createClass>`. Both `React.createClass` and a bare
// `createClass` are therefore silent, which is measured and is where `utilsreact.IsEs5ComponentCall`
// diverges: its `isCreateClassName` accepts `createClass` too, so the shelf would report on two
// shapes upstream declines. Spelled here rather than by narrowing the shelf, because other rules
// may want the wider set.
//
// Parentheses are skipped on the callee, which reproduces ESTree rather than widening past it:
// `(createReactClass)({...})` reports upstream, measured, because ESTree carries no parenthesized
// node for the callee check to see.
func isEs5FactoryCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == createReactClassPragma

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		receiver := ast.SkipParentheses(access.Expression)
		if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.Text() != "React" {
			return false
		}
		name := access.Name()
		return name != nil && name.Text() == createReactClassPragma
	}
	return false
}

// createReactClassPragma is upstream's default `createClass` pragma.
//
// Configurable upstream through `settings.react.createClass`, which cannot reach a rule here for
// the same reason the React version cannot. Fixed to the default, which is what an unconfigured
// upstream uses.
const createReactClassPragma = "createReactClass"

// isStringRefValue reports whether a ref attribute's value is a string this rule refuses.
//
// The bare-attribute case is the one worth naming: `<div ref />` has no initializer at all, and
// `containsStringLiteral` opens with `!!node.value` for exactly that shape. It is silent here for
// the same reason.
func isStringRefValue(initializer *ast.Node, noTemplateLiterals bool) bool {
	if initializer == nil {
		return false
	}
	switch initializer.Kind {
	// `ref="hello"`, the attribute written without braces. ESTree calls this a `Literal` sitting
	// directly in `node.value`, which is `containsStringLiteral`'s arm.
	case ast.KindStringLiteral:
		return true

	// `ref={...}`, upstream's `JSXExpressionContainer`, where only some expressions count.
	case ast.KindJsxExpression:
		inner := initializer.AsJsxExpression().Expression
		if inner == nil {
			return false
		}
		switch inner.Kind {
		// A Literal whose value is a string. A numeric literal is also a Literal upstream and its
		// value is not a string, so `ref={123}` falls out here rather than being excluded by kind.
		case ast.KindStringLiteral:
			return true
		// Both template shapes, and only under the option. See the type doc: our AST splits what
		// upstream matches as one `TemplateLiteral` into two kinds, and upstream's corpus fails on
		// both.
		case ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
			return noTemplateLiterals
		}
	}
	return false
}

// DecodeNoStringRefsOptions maps the wire keys onto the struct the rule reads.
//
// Hand-written rather than `rule.DecodeOptionsInto` for one key: `checkThisRefs` has to be
// distinguishable as absent, because absent means "resolve upstream's version gate", and that
// resolution is a decision rather than a bind. It happens to land on the same value as the zero
// value today, and it is written out anyway so that the reasoning is at the line if the config ever
// grows a settings surface.
func DecodeNoStringRefsOptions(raw []byte) (any, error) {
	var wire noStringRefsWireOptions
	if len(raw) == 0 {
		return NoStringRefsOptions{}, nil
	}
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return NoStringRefsOptions{}, err
	}

	options := NoStringRefsOptions{NoTemplateLiterals: wire.NoTemplateLiterals}

	// Absent means upstream's unconfigured answer. `getReactVersionFromContext` returns
	// `999.999.999` when no version is configured, so `< 18.3.0` is false and the half is off.
	if wire.CheckThisRefs != nil {
		options.CheckThisRefs = *wire.CheckThisRefs
	}

	return options, nil
}
