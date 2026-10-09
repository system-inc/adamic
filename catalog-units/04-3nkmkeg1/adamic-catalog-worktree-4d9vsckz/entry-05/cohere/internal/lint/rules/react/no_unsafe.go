package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The three findings, split by which replacement the help text names.
//
// Upstream renders one message shape with the offending name interpolated, and a help string that
// varies by which lifecycle method was found: `componentDidMount`, `getDerivedStateFromProps`, or
// `componentDidUpdate`. That is three distinct judgments wearing one identifier, and collapsing
// them here would cost a real guard, because `ExpectFindings` asserts identifiers and a single
// identifier cannot tell a fixture that the mount finding was rendered where the update finding
// belonged. Split, a fixture pins which of the three fired.
//
// Upstream carries a fourth branch, `_ => "alternative lifecycle methods"`, which nothing can
// reach: the match arms in `is_unsafe_method` admit exactly the six names the first three branches
// enumerate, so no name reaching the diagnostic can fall through. Reproduced as absent rather than
// as a fourth message, and stated here because the next reader will find it upstream and wonder.
var (
	messageUnsafeComponentWillMount = rule.Message{
		Id: "unsafeComponentWillMount",
		Description: "`componentWillMount` runs before the first render and is not guaranteed to " +
			"run exactly once, so work started there can be discarded or repeated when React " +
			"interrupts the render. Move the work to `componentDidMount`, which runs after the " +
			"tree is committed and cannot be thrown away.",
	}
	messageUnsafeComponentWillReceiveProps = rule.Message{
		Id: "unsafeComponentWillReceiveProps",
		Description: "`componentWillReceiveProps` runs on every parent render rather than only when " +
			"the props changed, so state derived inside it drifts out of step with the props it " +
			"was derived from. Compute the value in `getDerivedStateFromProps`, which is handed " +
			"both the incoming props and the current state and cannot read `this`.",
	}
	messageUnsafeComponentWillUpdate = rule.Message{
		Id: "unsafeComponentWillUpdate",
		Description: "`componentWillUpdate` runs while React is partway through preparing an " +
			"update it may still abandon, so side effects there can fire for a render that never " +
			"commits. Move them to `componentDidUpdate`, which runs only after the update reached " +
			"the screen.",
	}
)

// NoUnsafeOptions is the decoded option object.
//
// One field, matching oxc's `NoUnsafeConfig { check_aliases }` and the JSON schema it derives. The
// inventory records this rule as taking no options, which is wrong; it was checked against the
// `config = NoUnsafeConfig` line in the `declare_oxc_lint!` block and against all fourteen corpus
// cases carrying a second tuple element, rather than believed. That column has now been wrong on
// every rule in this lifecycle family.
type NoUnsafeOptions struct {
	// CheckAliases widens the rule to the three unprefixed spellings.
	//
	// False by default, which is upstream's default and the reason six of upstream's ten passing
	// cases pass: they write `componentWillMount` and friends in real components and are clean
	// only because nothing asked for the aliases. Upstream's own documentation recommends setting
	// it true, and upstream still ships the default false; reproduced rather than improved.
	CheckAliases bool `json:"checkAliases"`
}

// The six names this rule latches on, in two families.
//
// React 16.3 introduced the `UNSAFE_` prefixed spellings and kept the old ones working, so a file
// mid-migration carries both. The prefixed three are checked unconditionally; the unprefixed three
// only under `checkAliases`. That asymmetry is the whole option surface.
const (
	unsafeWillMountName             = "UNSAFE_componentWillMount"
	unsafeWillReceivePropsName      = "UNSAFE_componentWillReceiveProps"
	unsafeWillUpdateName            = "UNSAFE_componentWillUpdate"
	unsafeAliasWillMountName        = "componentWillMount"
	unsafeAliasWillReceivePropsName = "componentWillReceiveProps"
	unsafeAliasWillUpdateName       = "componentWillUpdate"
)

// NoUnsafe reports the deprecated React lifecycle methods on a component.
//
//	valid:   class Foo extends React.Component { componentDidUpdate() {} render() {} }
//	valid:   class Foo extends Bar { UNSAFE_componentWillMount() {} }             (no component)
//	valid:   const Foo = bar({ UNSAFE_componentWillMount: function() {} });       (no factory)
//	valid:   class Foo extends React.Component { componentWillMount() {} }        (alias, default off)
//	invalid: class Foo extends React.Component { UNSAFE_componentWillMount() {} }
//	invalid: const Foo = createReactClass({ UNSAFE_componentWillUpdate: function() {} });
//	invalid: the two alias cases above, under `checkAliases`
//
// Ported from `react/no-unsafe`, read against oxc's `no_unsafe.rs`, and measured on the release
// oxlint binary for every shape our abstract syntax tree spells differently than oxc's does.
//
// # The name is accurate but the default is narrower than it reads
//
// "Unsafe" reads as though the three deprecated lifecycle methods were the subject. By default they
// are not: only the three `UNSAFE_` prefixed spellings report, and `componentWillMount` written
// plainly in a real component is clean. The prefix is literal rather than descriptive. Measured on
// the release binary rather than inferred, and it is why six of upstream's ten passing cases pass.
//
// # Ownership is checked, and the two arms check different things
//
// This is the surprise, and it is the opposite of the surprise in three sibling rules where no
// ownership check exists at all. Here there are two, and they are not the same one.
//
// A class method asks `get_parent_component`, which accepts either era: the nearest ancestor that
// is a class extending a React component base, or a `createReactClass(...)` call. An object
// property asks only `is_es5_component`, walking every ancestor for the factory call and never
// considering a class. So the identical property is judged differently by where it sits:
//
//	class Foo extends React.Component { render() { const o = { UNSAFE_componentWillMount: f }; } }
//	const Foo = createReactClass({ render: function() { const o = { UNSAFE_componentWillMount: f }; } })
//
// The first is silent and the second reports, both measured. The second is upstream's ownership
// hole: the walk stops at the first factory call above the property, however far away and whatever
// lies between, so an unrelated object literal nested anywhere inside a `createReactClass` call is
// judged as though it were that component's own lifecycle property. Reproduced rather than
// corrected, because a silent improvement is a divergence nothing states.
//
// # A class property is not a method definition, and that narrowness is upstream's
//
// `UNSAFE_componentWillUpdate = () => {}` as a class field is silent. oxc matches
// `AstKind::MethodDefinition` and a class field is `PropertyDefinition`, a different kind it never
// looks at. This is a real difference from the sibling rule `no-will-update-set-state`, which does
// reach `PropertyDefinition`, so the two rules disagree about the same syntax and both are
// faithful. Measured on the release binary across the arrow-valued, function-expression-valued and
// bare spellings; all three silent. A rule matching class fields would be an improvement and would
// report where upstream does not, so it is not made.
//
// # The React version gate is upstream's and is not reachable here, so its default branch is all
//
// oxc reads `settings.react.version` and drops the prefixed spellings entirely when the configured
// version predates 16.3, where the prefix was introduced. cohere reads no React settings: the
// loader reads `settings` for better-tailwindcss only and refuses `settings.react` by name. So no
// version can reach a rule here, and oxc's own default, `is_none_or(supports_unsafe_lifecycle_prefix)`, answers true.
//
// Stating the measurement rather than the reasoning, because this is the shape of divergence claim
// that inoculates the next reader against checking: with no version configured, a class carrying
// all three prefixed methods produces three findings on the release binary; with
// `"settings": {"react": {"version": "16.2.0"}}` in the config, the identical source in the
// identical file produces zero, in the same session. Upstream's eighth and tenth pass cases are
// exactly those silent ones, and they are byte-identical to upstream's second and fourth fail
// cases, which differ only in carrying `"version": "16.3.0"`. Both are recorded in the fires test
// with the reasoning at the line, because the configuration that makes them clean cannot be
// produced here.
//
// # The ES5 factory predicate is spelled here rather than taken from the shelf
//
// `utilsreact.IsEs5ComponentCall` looks like exactly this question and answers a wider one: it
// accepts `createClass` and `React.createClass` as well as `createReactClass`, while
// `is_es5_component`, which is the helper this rule reaches, keys both of its arms on the single
// constant `createReactClass` at `oxc utils/react.rs:554`. Measured rather than read: both wider
// spellings are clean on the release binary, so reaching for the shelf would report where upstream
// is silent, and nothing in the imported corpus writes either spelling so no fixture would catch
// it. Handled locally rather than by narrowing the shelf, because other rules may want the wider
// set.
//
// The ES6 half is spelled here too, and for a second and unrelated reason: `IsEs6ComponentClass`
// skips parentheses on the heritage receiver where oxc does not, so `class Foo extends
// (React).Component` answers true there and is silent upstream. That divergence is invisible from
// the helper's doc comment, which advertises the paren skip as a correctness feature, and invisible
// from the imported corpus, which writes no parentheses at all. Measured; see
// `unsafeIsEs6ComponentClass`.
//
// # Parentheses, measured in both directions in the same rule
//
// `(createReactClass)({...})` reports, because oxc reaches the bare arm through
// `get_identifier_reference()`, which sees through parentheses. `class Foo extends (React.Component)`
// and `class Foo extends (React).Component` are both silent, because `heritage_expression()` hands
// back the parenthesized node and `as_member_expression()` declines it, and because
// `member_expr.object()` on the second is a parenthesized expression rather than an identifier. All
// three measured on the release binary in one invocation. So this rule skips parentheses on the
// factory callee and on nothing else, which is upstream's own inconsistency within one rule rather
// than a simplification. Our tree keeps the parenthesis as a real node in the heritage clause
// exactly where oxc's declines to look through one, so withholding the skip there is what
// reproduces upstream rather than what narrows it, and the shelf's helper does skip it. A fixture
// pair pins both spellings and it is what caught the shelf version.
//
// # Reading a property key cannot go through Text()
//
// A computed key is a `ComputedPropertyName` and `Node.Text()` panics on it outright, so keys are
// resolved with `ast.TryGetTextOfPropertyName`, which checks kind before reading. That is the
// shim's spelling of oxc's `static_name()` and resolves the same shapes: an identifier, a string
// literal, and a computed key holding a literal. All three measured reporting on the release
// binary, so a port matching only the identifier spelling loses two real findings and crashes the
// linter on a third input.
//
// # Nothing asks what kind of member carries the name
//
// A `static UNSAFE_componentWillUpdate` is not a lifecycle method at all and reports, and a getter
// named for one reports. oxc reaches both through the same `MethodDefinition` arm and never asks.
// Both measured; reproduced rather than corrected.
//
// # No source-type gate, and that is a deliberate difference rather than an omission
//
// oxc declares `should_run` returning `ctx.source_type().is_jsx()`, so it skips a plain `.ts` file
// entirely. Nothing in this rule reads a JSX node; the gate is oxc narrowing a react plugin to
// react files. Our harness has no per-rule source-type predicate and the config decides which files
// a rule is offered, so reproducing it would mean inventing a mechanism our config already answers
// with. The decision is unchanged either way: a `.ts` file holding `class Foo extends
// React.Component` is legal TypeScript and the finding is correct there. Stated rather than silent,
// because it is the one place this port is knowingly wider than oxc, and the only inputs it can
// differ on are ones oxc declines to look at rather than ones it judges differently.
var NoUnsafe = rule.Rule{
	// No namespace prefix. The config writes `react/no-unsafe` and the parity guard strips the
	// namespace on a `/` boundary.
	Name: "react/no-unsafe",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// An unconfigured rule gets the zero value, which is upstream's default of false. The
		// comma-ok form matters: the plain harness and a bare severity in the config both hand this
		// a nil, and an unchecked assertion would panic on every real file.
		settings, _ := rule.OptionsAs[NoUnsafeOptions](options)

		report := func(name *ast.Node, message rule.Message) {
			// The key, not the member. Upstream's span is `key.span()` on both arms, so
			// `UNSAFE_componentWillMount() {}` underlines twenty five characters and stops before
			// the parenthesis, which all twelve snapshot diagnostics confirm.
			//
			// A computed key needs the brackets stripped, and this is a case where our tree keeps a
			// node oxc's has already dropped. oxc's `PropertyKey` for `["name"]() {}` is the string
			// literal itself, with no wrapper, so `key.span()` covers twenty seven characters
			// starting at the quote. Ours is a `ComputedPropertyName` whose span covers the
			// brackets too, and reporting it would point at twenty nine characters starting at the
			// bracket. Measured on the release binary rather than reasoned about: the computed key
			// underlines from column 4 and the neighbouring string-literal key from column 3, in
			// the same file, so the quotes are inside upstream's span and the brackets are not.
			//
			// Found by asserting the span, which is the whole reason step 8 is not optional here:
			// every identifier assertion in this file was already green with the brackets included.
			if name != nil && name.Kind == ast.KindComputedPropertyName {
				if inner := name.AsComputedPropertyName().Expression; inner != nil {
					name = inner
				}
			}
			ctx.ReportNode(name, message)
		}

		judge := func(name *ast.Node) (rule.Message, bool) {
			if name == nil {
				return rule.Message{}, false
			}
			text, named := ast.TryGetTextOfPropertyName(name)
			if !named {
				return rule.Message{}, false
			}
			switch text {
			case unsafeWillMountName:
				return messageUnsafeComponentWillMount, true
			case unsafeWillReceivePropsName:
				return messageUnsafeComponentWillReceiveProps, true
			case unsafeWillUpdateName:
				return messageUnsafeComponentWillUpdate, true
			case unsafeAliasWillMountName:
				return messageUnsafeComponentWillMount, settings.CheckAliases
			case unsafeAliasWillReceivePropsName:
				return messageUnsafeComponentWillReceiveProps, settings.CheckAliases
			case unsafeAliasWillUpdateName:
				return messageUnsafeComponentWillUpdate, settings.CheckAliases
			}
			return rule.Message{}, false
		}

		return rule.Listeners{
			// The class half. A method declaration and both accessors are one `MethodDefinition`
			// upstream, which is why all three are listened for here; a class field is deliberately
			// absent, and the rule's doc carries that measurement.
			ast.KindMethodDeclaration: func(node *ast.Node) {
				message, unsafe := judge(node.AsMethodDeclaration().Name())
				if !unsafe {
					return
				}
				if unsafeEnclosingComponent(node) == nil {
					return
				}
				report(node.AsMethodDeclaration().Name(), message)
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				message, unsafe := judge(node.AsGetAccessorDeclaration().Name())
				if !unsafe {
					return
				}
				if unsafeEnclosingComponent(node) == nil {
					return
				}
				report(node.AsGetAccessorDeclaration().Name(), message)
			},
			ast.KindSetAccessor: func(node *ast.Node) {
				message, unsafe := judge(node.AsSetAccessorDeclaration().Name())
				if !unsafe {
					return
				}
				if unsafeEnclosingComponent(node) == nil {
					return
				}
				report(node.AsSetAccessorDeclaration().Name(), message)
			},

			// The object half, and it asks a different question. Upstream's `ObjectProperty` arm
			// walks for `is_es5_component` alone, so a class ancestor never satisfies it however
			// near, and a factory ancestor satisfies it however far. The asymmetry is measured and
			// the rule's doc carries both halves.
			ast.KindPropertyAssignment: func(node *ast.Node) {
				message, unsafe := judge(node.AsPropertyAssignment().Name())
				if !unsafe {
					return
				}
				if unsafeEnclosingCreateReactClassCall(node) == nil {
					return
				}
				report(node.AsPropertyAssignment().Name(), message)
			},
		}
	},
}

// unsafeEnclosingComponent returns the nearest ancestor that is a component by either era, or nil.
//
// This is `get_parent_component`, which is `is_es5_component || is_es6_component` over ancestors.
// The ES6 half is the shelf's; the ES5 half is local, and the rule's doc says why. The node itself
// is considered, matching oxc's `ancestors`, which for a `MethodDefinition` starts above it anyway
// and therefore cannot differ: a method is neither a class nor a call.
func unsafeEnclosingComponent(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		if unsafeIsEs6ComponentClass(current) || unsafeIsCreateReactClassCall(current) {
			return current
		}
	}
	return nil
}

// unsafeIsEs6ComponentClass reports whether a class extends a React component base.
//
// Spelled here rather than taken from `utilsreact.IsEs6ComponentClass`, and the reason is a
// parenthesis, which is the one thing the shelf's doc comment advertises as a correctness feature
// and which is a divergence for this rule.
//
// The shelf reads the extends target through `isIdentifierNamed`, which calls `ast.SkipParentheses`
// on the receiver, so `class Foo extends (React).Component` answers true there. oxc reads
// `member_expr.object()` directly and gets a parenthesized expression rather than an identifier, so
// it answers false. Measured on the release binary in one invocation: `extends (React).Component`
// and `extends (React.Component)` are both silent while the bare spelling reports. Our tree keeps
// the parenthesis as a real node exactly where oxc's has already dropped it in the factory-callee
// case, so this rule skips parentheses on the factory callee and withholds the skip here, which is
// upstream's own inconsistency reproduced rather than smoothed.
//
// A fixture pair covers both spellings, and it caught the shelf version: the port passed every
// imported case with the shelf helper in place, because nothing in the corpus writes a parenthesis.
//
// Only `extends` counts. A TypeScript `implements` clause is a heritage clause of a different token
// and oxc reads `heritage_expression()`, which is the superclass alone.
func unsafeIsEs6ComponentClass(node *ast.Node) bool {
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
		// No check on the clause's token, and its absence is measured rather than an oversight.
		//
		// An `extends` clause and an `implements` clause are both `KindHeritageClause`, so testing
		// the token reads as necessary, and a mutation removing that test survives every fixture in
		// this file. It survives because it is subsumed: the parser gives an `implements` clause
		// types of kind `KindTypeReference` and an `extends` clause types of kind
		// `KindExpressionWithTypeArguments`, so the kind guard below already declines every
		// `implements` target. Probed on four spellings, including `implements React.Component`,
		// `implements Component<T>`, and a class carrying both clauses at once; all three
		// `implements` targets came back `KindTypeReference` and both `extends` targets came back
		// `KindExpressionWithTypeArguments`.
		//
		// Written down because the deleted check looked load-bearing and a later reader will want
		// to add it back. The two `implements` fixtures stay, pinning the kind guard rather than a
		// token guard, so if the parser ever changes shape they fail.
		types := clause.AsHeritageClause().Types
		if types == nil {
			continue
		}
		for _, typeNode := range types.Nodes {
			if typeNode.Kind != ast.KindExpressionWithTypeArguments {
				continue
			}
			if unsafeIsComponentBase(typeNode.AsExpressionWithTypeArguments().Expression) {
				return true
			}
		}
	}
	return false
}

// unsafeIsComponentBase reports whether an extends target names a React component base class.
//
// `React.Component`, `React.PureComponent`, and the bare `Component` and `PureComponent` a file
// importing them directly writes. All four measured reporting on the release binary; upstream's
// corpus writes only the first.
//
// No parentheses are skipped anywhere in here, and neither the receiver nor the whole expression is
// unwrapped. The doc on `unsafeIsEs6ComponentClass` carries that measurement.
func unsafeIsComponentBase(expression *ast.Node) bool {
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindIdentifier:
		return unsafeIsComponentBaseName(expression.Text())

	case ast.KindPropertyAccessExpression:
		access := expression.AsPropertyAccessExpression()
		object := access.Expression
		if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "React" {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && unsafeIsComponentBaseName(name.Text())
	}
	return false
}

// unsafeIsComponentBaseName accepts the two base classes a component may extend.
func unsafeIsComponentBaseName(name string) bool {
	return name == "Component" || name == "PureComponent"
}

// unsafeEnclosingCreateReactClassCall returns the nearest factory-call ancestor, or nil.
//
// Deliberately not `unsafeEnclosingComponent`. Upstream's object arm never asks about a class, so a
// lifecycle-shaped property inside a plain object inside a real ES6 class component is silent, and
// that is measured rather than assumed. Sharing one walk between the two arms would report there.
func unsafeEnclosingCreateReactClassCall(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		if unsafeIsCreateReactClassCall(current) {
			return current
		}
	}
	return nil
}

// unsafeIsCreateReactClassCall reports whether a call is `createReactClass(...)` or `React.createReactClass(...)`.
//
// Only that one name, matching `is_es5_component`, which keys both of its arms on the single
// constant `CREATE_CLASS = "createReactClass"`. The shelf's `IsEs5ComponentCall` also accepts
// `createClass` and `React.createClass`, both of which are clean on the release binary, so it is
// not reached; the rule's doc carries the measurement.
//
// Parentheses are skipped on the bare arm and on the receiver of the namespaced arm, matching
// `get_identifier_reference()` and `isIdentifierNamed` respectively. `(createReactClass)({...})`
// reports on the release binary. The receiver skip is not separately measured and is inherited from
// the shelf's convention; nothing in either corpus writes `(React).createReactClass`.
func unsafeIsCreateReactClassCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
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
