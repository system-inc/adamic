package react

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferEs6ClassMode is which of the two component styles the configuration wants.
type PreferEs6ClassMode string

const (
	// PreferEs6ClassAlways wants ES6 classes, so a `createReactClass` call reports. This is
	// upstream's default and the unconfigured answer.
	PreferEs6ClassAlways PreferEs6ClassMode = "Always"

	// PreferEs6ClassNever wants the ES5 factory, so an ES6 component class reports.
	PreferEs6ClassNever PreferEs6ClassMode = "Never"
)

// PreferEs6ClassOptions configures the rule.
//
// Upstream's option is a bare string, `"always"` or `"never"`, and the config takes it in exactly
// that form. It used to take `{"mode": "Always"}`, a shape no ESLint version accepts, which refused
// upstream's own spelling; that form is now refused instead (#d21war2).
type PreferEs6ClassOptions struct {
	// Mode is which style the rule enforces. Absent means Always, which is upstream's default.
	Mode PreferEs6ClassMode
}

// DefaultPreferEs6ClassOptions is the unconfigured answer.
//
// Upstream reads `context.options[0] || 'always'`, so an absent option and an explicitly configured
// `always` are the same rule. The corpus states that agreement directly: it ships the same
// `createReactClass` source twice, once with no option and once under `always`, both failing.
func DefaultPreferEs6ClassOptions() PreferEs6ClassOptions {
	return PreferEs6ClassOptions{Mode: PreferEs6ClassAlways}
}

// DecodePreferEs6ClassOptions reads this rule's configuration from the config layer.
//
// The option is a bare enum string rather than an object, so the generic helper, which unmarshals
// into a struct, cannot read it. Upstream's schema is `[{ enum: ["always", "never"] }]`, so anything
// else is refused, naming the value: the two modes report on disjoint inputs, and a value read as
// neither would silently select a third behaviour of reporting nothing at all.
func DecodePreferEs6ClassOptions(raw []byte) (any, error) {
	options := DefaultPreferEs6ClassOptions()
	if len(raw) == 0 {
		return options, nil
	}
	var mode string
	if err := json.Unmarshal(raw, &mode); err != nil {
		return options, fmt.Errorf(`expected a mode string, "always" or "never", got %s`, raw)
	}
	switch mode {
	case "always":
		return options, nil
	case "never":
		options.Mode = PreferEs6ClassNever
		return options, nil
	}
	return options, fmt.Errorf(`mode %q is not "always" or "never"`, mode)
}

var messagePreferEs6ClassShouldUseEs6Class = rule.Message{
	Id: "shouldUseES6Class",
	Description: "This component is built with the `createReactClass` factory, which React " +
		"removed from the library and now ships as a separate package nobody maintains. It " +
		"carries its own version of features the language has since grown, notably automatic " +
		"binding of methods to the instance, so a reader who knows classes cannot predict what " +
		"`this` is here. Write it as a class extending `React.Component` instead.",
}

var messagePreferEs6ClassShouldUseCreateClass = rule.Message{
	Id: "shouldUseCreateClass",
	Description: "This component is written as an ES6 class while the configuration asks for the " +
		"`createReactClass` factory. What costs the reader is a codebase that uses each style in " +
		"about half its components, because then every component has to be read before its shape " +
		"is known. Write it with the factory instead, or change the configuration.",
}

// PreferEs6Class enforces one of the two ways a React component can be declared, and reports the
// other.
//
//	Always (the default):
//	  valid:   class Hello extends React.Component { render() { return null; } }
//	  valid:   var Hello = createClass({ ... })          (not the factory this rule knows)
//	  invalid: var Hello = createReactClass({ render: function() { return null; } })
//
//	Never:
//	  valid:   var Hello = createReactClass({ ... })
//	  valid:   const Hello = class extends React.Component { ... }   (an expression, see below)
//	  invalid: class Hello extends React.Component { render() { return null; } }
//
// Ported from `react/prefer-es6-class` in `eslint-plugin-react`, read from the clone at
// `lib/rules/prefer-es6-class.js` together with `componentUtil.isES5Component`,
// `componentUtil.isES6Component`, `componentUtil.isExplicitComponent`, and
// `pragma.getCreateClassFromContext`. The corpus is five passing and three failing cases and is
// silent about almost everything interesting, so most of what follows was measured by driving the
// installed build (7.37.5) rather than read.
//
// # There is no file-suffix gate
//
// Three siblings in this package gate on the file being read as JSX, inherited from oxc. The
// authority here has no such gate anywhere.
//
// # The factory name is `createReactClass` and nothing else
//
// `getCreateClassFromContext` defaults to the single name `createReactClass`, configurable through
// a shared setting this tree does not set. So `createClass({...})` and `React.createClass({...})`
// are both CLEAN, which is the opposite of what the shelf's `IsEs5ComponentCall` answers: that
// helper deliberately accepts `createClass` too, on the reasoning that rules meet both halves of
// the ecosystem. Following the shelf here would report on code upstream leaves alone.
//
// Measured, all four spellings under the default mode:
//
//	createReactClass({...})           REPORTS
//	React.createReactClass({...})     REPORTS
//	createClass({...})                SILENT
//	React.createClass({...})          SILENT
//	Foo.createReactClass({...})       SILENT     the receiver must be the React pragma
//
// `isEs5ComponentCallStrict` in `no_did_mount_set_state.go` is exactly this narrowing, written for
// oxc's identical single-constant set, so it is reused rather than restated. That the two
// authorities agree on the name set is a coincidence worth naming: oxc's `CREATE_CLASS` and
// ESLint's `getCreateClassFromContext` default are the same string for different reasons.
//
// # The object need not be the first argument, and nesting does not matter
//
// Upstream anchors on the object expression and asks whether its PARENT is a matching call. So the
// question is "is this object an argument to the factory", not "is this the factory's first
// argument". Measured: `createReactClass(x, {...})` reports, and an object nested inside the
// factory's argument does not, because its parent is the outer object rather than the call. A
// factory call with a nested object reports exactly once.
//
// # The Never arm anchors on a class DECLARATION only
//
// Upstream registers `ClassDeclaration` and not `ClassExpression`, so
// `const Hello = class extends React.Component {}` is SILENT. Measured, and it is worth stating
// because the sibling `state-in-constructor` in this package goes the other way: its scope walk
// finds a class expression and reports there. Two rules in one package, one authority, opposite
// answers, and the difference is which node the rule registers.
//
// That is upstream being narrow rather than upstream being right, and it is reproduced. A port
// that helpfully added class expressions would report on code upstream leaves alone, and no
// imported fixture could see it because the corpus writes none.
//
// # The base-class test is a regular expression over the name
//
// `/^(Pure)?Component$/` against either a bare identifier or the property of a member expression
// whose object is the React pragma. So `Component` and `PureComponent` match and nothing else
// does: `ComponentFoo` and `FooComponent` are both silent, measured, which is what the anchors in
// the pattern buy. `Foo.Component` is silent because the receiver is not React.
//
// Parentheses report on both arms, measured over four spellings each. ESLint's AST does not
// materialize a parenthesized expression in either position, so upstream never has to think about
// them; ours does, so both the callee and the base expression are unwrapped here.
//
// # The JSDoc path is dead in the installed build, and the silence is reproduced
//
// `isES6Component` first calls `isExplicitComponent`, which reads a JSDoc comment through
// `sourceCode.getJSDocComment` and looks for an `@extends React.Component` or `@augments` tag. A
// class with no heritage and that tag should report under Never.
//
// It does not. Measured under both the default parser and the TypeScript one, all four tag
// spellings are silent. Confirmed as a plugin-wide dead path rather than something specific to this
// rule by running `react/require-render-return`, which routes through the same
// `isES6Component`, over the same JSDoc class: also silent, while the same rule over a class with a
// real `extends` clause reports. ESLint 10's `getJSDocComment` no longer returns what
// `isExplicitComponent` expects.
//
// So this rule implements no JSDoc path. That reproduces the behaviour of the build this tree
// runs, which is what the gate compares against. If a later ESLint revives the path, this becomes a
// real divergence and the note above is where to start.
//
// # Where the findings point
//
// The two arms report different node kinds and therefore different spans. The Always arm reports
// the OBJECT EXPRESSION, so the span starts at the brace and excludes `createReactClass(`. The
// Never arm reports the whole class declaration. Measured by reading the reported columns, and
// asserted by the span fixtures rather than left to a message id.
var PreferEs6Class = rule.Rule{
	// No namespace prefix beyond the plugin's own. The config writes `react/prefer-es6-class` and
	// the parity guard strips the namespace on a `/` boundary.
	Name: "react/prefer-es6-class",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[PreferEs6ClassOptions](options)
		if !ok {
			// A rule configured as bare `"error"` reaches here with nil options, which type-asserts
			// to the zero value whose Mode matches neither arm. Falling back to the default is what
			// keeps the rule from registering and then reporting nothing.
			settings = DefaultPreferEs6ClassOptions()
		}
		if settings.Mode == "" {
			settings.Mode = PreferEs6ClassAlways
		}

		if settings.Mode == PreferEs6ClassAlways {
			return rule.Listeners{
				ast.KindObjectLiteralExpression: func(node *ast.Node) {
					if !preferEs6ClassIsFactoryArgument(node) {
						return
					}
					ctx.ReportNode(node, messagePreferEs6ClassShouldUseEs6Class)
				},
			}
		}

		return rule.Listeners{
			// A declaration only, never an expression. See the rule doc: upstream registers
			// `ClassDeclaration` and `const H = class extends React.Component {}` is silent.
			ast.KindClassDeclaration: func(node *ast.Node) {
				if !preferEs6ClassExtendsAComponent(node) {
					return
				}
				ctx.ReportNode(node, messagePreferEs6ClassShouldUseCreateClass)
			},
		}
	},
}

// preferEs6ClassIsFactoryArgument reports whether an object literal is an argument to the
// `createReactClass` factory.
//
// Upstream asks whether the object's PARENT is a call whose callee matches, which is why the
// object's position among the arguments does not matter and why a nested object is clean: a nested
// object's parent is the enclosing object rather than the call.
//
// # The kind test here is subsumed, and is kept anyway
//
// `IsEs5ComponentCall`, which `isEs5ComponentCallStrict` calls first, opens with exactly this nil
// and call-expression test, so neither half below can change an answer and both survive mutation
// independently. That is a correct verdict rather than a fixture gap.
//
// It is kept because it states the question this function asks at the point it asks it, and because
// the alternative reads as though any parent kind were acceptable. The cost is two comparisons per
// object literal, and object literals are common enough that this was worth checking rather than
// assuming: the dry run measures the rule at 0.5ms over the whole tree.
//
// The verdict names the callee as it stands today. If `isEs5ComponentCallStrict` ever stops routing
// through the shelf helper, the guard becomes load-bearing again.
func preferEs6ClassIsFactoryArgument(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	// `isEs5ComponentCallStrict` lives in no_did_mount_set_state.go and is the shelf helper
	// narrowed to the single factory name. Both authorities happen to accept exactly
	// `createReactClass`, for unrelated reasons, so the narrowing is reused rather than restated.
	return isEs5ComponentCallStrict(parent)
}

// preferEs6ClassExtendsAComponent reports whether a class declaration extends a React component
// base.
//
// Written here rather than reaching for `react.IsEs6ComponentClass` for two reasons that pull in
// opposite directions, so neither alone would justify it. The shelf helper accepts a class
// EXPRESSION, which this rule must not, and it declines a parenthesized base expression, which this
// rule must accept. Wrapping it would need both a narrowing and a widening, at which point the
// direct test is the shorter and more readable of the two.
//
// The name test is upstream's `/^(Pure)?Component$/`, which is anchored at both ends, so
// `ComponentFoo` and `FooComponent` are both silent. Measured.
//
// # The entry guard is subsumed, and is kept anyway
//
// Both halves of the guard below survive mutation independently, and that is correct rather than a
// fixture gap. The only caller is the `KindClassDeclaration` listener in this file, enumerated by
// grep at the time of writing, and a listener key can deliver neither nil nor another kind. So no
// input can reach either half, and no fixture can distinguish the two versions.
//
// It is kept because it is a nil dereference away from a panic if a second caller ever appears, and
// a panic in a rule costs every other rule that whole file rather than just this one. The cost of
// keeping it is two comparisons per class declaration.
//
// **That verdict names the callers that existed when it was taken.** If this helper acquires a
// second caller, the verdict is void and the guard becomes load-bearing rather than merely cheap.
func preferEs6ClassExtendsAComponent(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindClassDeclaration {
		return false
	}
	heritage := node.AsClassDeclaration().HeritageClauses
	if heritage == nil {
		return false
	}

	for _, clause := range heritage.Nodes {
		if clause.Kind != ast.KindHeritageClause {
			continue
		}
		types := clause.AsHeritageClause().Types
		if types == nil {
			continue
		}
		for _, typeNode := range types.Nodes {
			// An `implements` clause is given `KindTypeReference` by this parser rather than
			// `KindExpressionWithTypeArguments`, so this kind test already declines it and a
			// separate token test would be inert. Measured on the sibling
			// `state-in-constructor`, where such a test survived every fixture and a parse probe
			// over six heritage shapes showed why.
			if typeNode.Kind != ast.KindExpressionWithTypeArguments {
				continue
			}
			base := ast.SkipParentheses(typeNode.AsExpressionWithTypeArguments().Expression)
			if preferEs6ClassIsComponentBase(base) {
				return true
			}
		}
	}
	return false
}

// preferEs6ClassIsComponentBase reports whether a base expression names a React component base.
//
// Upstream accepts a bare identifier matching `/^(Pure)?Component$/`, or a member expression whose
// object is the React pragma and whose property matches the same pattern. Parentheses are unwrapped
// on the receiver as well as on the base itself, because ESLint's AST does not materialize them and
// all four parenthesized spellings report upstream.
func preferEs6ClassIsComponentBase(expression *ast.Node) bool {
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindIdentifier:
		return preferEs6ClassIsComponentBaseName(expression.Text())

	case ast.KindPropertyAccessExpression:
		access := expression.AsPropertyAccessExpression()
		receiver := ast.SkipParentheses(access.Expression)
		if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.Text() != "React" {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier &&
			preferEs6ClassIsComponentBaseName(name.Text())
	}
	return false
}

// preferEs6ClassIsComponentBaseName is upstream's `/^(Pure)?Component$/` written as equality.
//
// The pattern is anchored at both ends and its only variable part is an optional literal prefix, so
// it matches exactly two strings and a regular expression would buy nothing.
func preferEs6ClassIsComponentBaseName(name string) bool {
	return name == "Component" || name == "PureComponent"
}
