package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoThisInSfc = rule.Message{
	Id: "noThisInSfc",
	Description: "A function component is called, not constructed, so `this` is not the component. " +
		"It is `undefined` under a module's strict mode and the global object otherwise, and " +
		"either way the props being read here belong to nothing. The read almost always survives " +
		"a class component that was rewritten as a function and the `this.` prefixes were left " +
		"behind. Take props and context as parameters instead.",
}

// NoThisInSfc flags a `this` read inside a stateless function component.
//
//	valid:   function Foo(props) { return <div>{props.bar}</div>; }
//	valid:   class Foo extends React.Component { render() { return <div>{this.props.bar}</div>; } }
//	valid:   function foo(bar) { this.bar = bar; }                    (not component-named)
//	valid:   function Foo() { function cb() { this.x = 1; } return <div />; }   (fn rebinds this)
//	valid:   const Foo = function () { return this.props.a; };        (no function id)
//	invalid: function Foo(props) { return <div>{this.props.foo}</div>; }
//	invalid: const Foo = (props) => this.props.foo ? <span /> : null;
//	invalid: function Foo() { const h = () => this.props.click(); return <button />; }
//
// Ported from `react/no-this-in-sfc`, read against oxc's `no_this_in_sfc.rs` and the three helpers
// it calls: `is_es5_component` and `is_es6_component` (`oxc_linter/src/utils/react.rs:556`, `:577`)
// and `is_react_component_name` (`:784`). ESLint's `eslint-plugin-react` source is in this tree and
// was read as a second opinion. It carries `schema: []`, which is the authoritative statement that
// this rule takes no options, and oxc deserializes nothing, so the two agree and this takes none.
//
// The two implementations differ in a way worth stating, and oxc wins because oxc is what the
// differential gate runs. ESLint anchors on `MemberExpression` and reports the WHOLE member
// expression, routing the component question through its `Components` detector, which tracks
// returned JSX and would recognize a `memo` or `forwardRef` wrapper. oxc anchors on the
// `ThisExpression` and reports just the four characters of `this`, and its component test is purely
// syntactic. The upstream snapshot underlines `this` alone on all fifteen findings, confirming the
// span, and the syntactic test is reproduced below rather than ESLint's semantic one.
//
// # What counts as a stateless function component, which is the whole rule
//
// This is where a port of this rule goes wrong, so it is measured on the release oxlint binary
// rather than reasoned about. The test is syntactic and much narrower than the name suggests.
// A function qualifies when, and only when, a NAME can be read off it whose first character is
// ASCII-uppercase, and the name is read from exactly two places:
//
//	function Foo(p) { ... }                REPORTS   a function's own id
//	const Foo = function Bar(p) { ... }    REPORTS   still the id; the variable is not consulted
//	const Foo = function (p) { ... }       SILENT    an anonymous function expression has no id
//	const Foo = function bar(p) { ... }    SILENT    the id is lowercase, the variable is ignored
//	const Foo = (p) => ...                 REPORTS   an arrow reads its variable declarator instead
//	let Foo = (p) => ...                   REPORTS   any binding kind
//	function foo(p) { ... }                SILENT    lowercase
//	function _Foo(p) { ... }               SILENT    underscore is not uppercase
//	const $Foo = (p) => ...                SILENT    nor is a dollar sign
//	function useFoo() { ... }              SILENT    a hook is not a component here
//
// **Returning JSX is not part of the test at all.** `function Foo(props) { return this.props.a; }`
// reports with no JSX anywhere in it, and upstream's own failing corpus contains such a case. A
// port that gated on "returns JSX" would pass most fixtures and silence this one.
//
// **No wrapper is recognized.** All three of these are SILENT on the release binary, because the
// arrow's parent is a call argument rather than a variable declarator, so no name can be read:
//
//	const Foo = memo((props) => this.props.a)
//	const Foo = React.memo((props) => this.props.a)
//	const Foo = React.forwardRef((props, ref) => this.props.a)
//
// The same reasoning silences an object property (`{ Foo: (p) => ... }`), a destructured binding,
// and a default export of an anonymous arrow. None of these is in the corpus; all were measured.
// That is upstream being crude rather than upstream being right, and it is reproduced rather than
// improved on, because a differential run against oxlint is what this has to survive.
//
// # The name test is ASCII, and the shelf helper is not
//
// `internal/utilities/react.IsLikelyComponentName` is exactly the shape this reaches for and is wrong
// for it. Its body is `unicode.IsUpper(runes[0])`, which is Unicode-wide; oxc's
// `is_react_component_name` is `c.is_ascii_uppercase()`. Measured on the release binary, all three
// of `function Фoo`, `function Λoo` and `function Éoo` reading `this.props.a` are SILENT upstream,
// while the shelf helper answers true for every one of them and would report all three. The
// predicate below is ASCII-only for that reason. Read the body, not the name.
//
// # Which construct rebinds `this`, measured rather than inferred from the language
//
// JavaScript semantics say a nested `function` rebinds `this` and a nested arrow does not, and
// upstream happens to agree, but it reaches that answer through two separate mechanisms that do not
// line up with the language and have to be reproduced separately.
//
// The first is the ancestor walk. It runs from the `this` outward and stops at the first function
// it meets. A `Function` ancestor ends the walk unconditionally, answering with itself only when it
// is component-named. An arrow ancestor that is NOT component-named is stepped over and the walk
// continues upward, which is what lets an anonymous callback inside a component still belong to
// that component. Measured:
//
//	function Foo(p) { const cb = () => this.props.a; }    REPORTS   arrow steps over
//	function Foo(p) { function cb() { this.props.a; } }   SILENT    fn stops the walk
//	function Foo(p) { const cb = function () { this.props.a; }; }  SILENT   anonymous fn, same
//	function Foo(p) { const Cb = () => this.props.a; }    REPORTS   the inner arrow is the component
//	function Foo(p) { return { m() { return this.props.a; } }; }   SILENT   a method is a Function
//
// The second is a separate exemption pass over the ancestors BETWEEN the `this` and the component
// that was found, which exempts a `Function`, a method, a property definition, and an accessor
// property whose value contains the `this`. The two overlap on nested functions and disagree
// nowhere the corpus reaches, but both are reproduced because each catches cases the other does
// not: a class body nested inside an arrow component is caught only by the second.
//
// # The accessor split, which our AST spells differently
//
// oxc has two node kinds where we have one. Its `PropertyDefinition` exempts unconditionally, and
// its `AccessorProperty` exempts only when the value's span contains the `this`. TypeScript's AST
// models both as `KindPropertyDeclaration`, distinguished by an accessor modifier. Measured, and
// the two rows genuinely differ:
//
//	function Foo() { class C { accessor h = () => this.value; } }    SILENT   this is in the value
//	function Foo() { class C { accessor [this.props.n] = 0; } }      REPORTS  a computed KEY is not
//	function Foo() { class C { h = () => this.value; } }             SILENT   plain, unconditional
//	function Foo() { class C { [this.props.n] = 0; } }               SILENT   plain, unconditional
//
// Both accessor rows are in upstream's corpus, one passing and one failing, so this split is the
// one part of the exemption the imported fixtures can actually see. `ast.HasAccessorModifier` is
// what picks the arm.
//
// # The class-component ancestor exemption, and the strict factory name
//
// After a component is found, upstream walks that component's own ancestors and exempts if any is
// an ES5 or ES6 React component. So an arrow component defined inside a class component's method
// is exempt, while the same arrow inside a plain class reports. Measured on all seven heritage
// spellings: `React.Component`, `React.PureComponent`, bare `Component` and bare `PureComponent`
// exempt, and `Foo.Component`, `Base` and a class with no extends clause do not.
//
// The ES5 half is where the shelf is wrong in the other direction.
// `internal/utilities/react.IsEs5ComponentCall` accepts `createClass` and `React.createClass` as well
// as `createReactClass`, while oxc keys both arms of `is_es5_component` on the single constant
// `CREATE_CLASS = "createReactClass"`. Measured on the release binary, an arrow component holding a
// `this` read inside `createReactClass({...})` is SILENT while the same thing inside
// `createClass({...})` and inside `React.createClass({...})` both REPORT. Following the shelf
// helper's name instead of its body would silence two cases upstream reports, and no imported
// fixture could see it because the corpus never nests a component inside an ES5 factory. This
// calls `isEs5ComponentCallStrict` in `no_did_mount_set_state.go`, which is that same narrowing
// measured there for the same reason.
//
// Upstream's own passing corpus contains `const Foo = React.createClass({ render: function () {
// return <div>{this.props.foo}</div>; } });`, which looks like it is passing because of this
// exemption and is not. It passes because `render: function ()` is an ANONYMOUS function
// expression, so no name can be read, the walk ends at the `Function` arm and answers nothing. The
// case would pass identically with the factory removed, which was measured. A port that read that
// fixture as evidence about the factory name would draw exactly the wrong conclusion, which is why
// the narrowing above rests on the probe rather than on the corpus.
//
// # Parentheses, measured because no imported fixture writes one
//
// The anchor requires the `this`'s immediate parent to be a member expression, with no
// paren-skipping anywhere. Measured on the release binary:
//
//	this.props.foo        REPORTS   the control
//	(this.props).foo      REPORTS   the parens are outside the member expression
//	(this).props.foo      SILENT    the parent is a parenthesized expression, not a member
//	((this)).props.foo    SILENT    the same
//	this["props"].foo     REPORTS   a computed member counts, and is in the corpus
//	this                  SILENT    a bare `this` with no member access at all
//	this()                SILENT    a call is not a member expression
//
// Our tree keeps parentheses as a real `KindParenthesizedExpression` node in exactly the places
// oxc does, confirmed by dumping the ancestor chain for both forms, so a plain parent-kind check
// with no `SkipParentheses` reproduces all seven rows. Adding the paren-skip that usually reads as
// a free correctness improvement would report two cases oxlint is silent on. All are pinned below.
var NoThisInSfc = rule.Rule{
	// No namespace prefix. The config writes `react/no-this-in-sfc` and the parity guard strips the
	// namespace on a `/` boundary, so `react-no-this-in-sfc` would match no inventory entry, lint
	// no files, and still pass every fixture in this package.
	Name: "react/no-this-in-sfc",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// No file gate. oxc gated this on `source_type().is_jsx()` and that gate came along with the
		// port from oxc, but the authority here is eslint-plugin-react, which does not gate on the
		// file name at all. React code in a `.ts` file is ordinary and legal, and the gate made this
		// rule silent across more `.ts` files than the `.tsx` files it could see.

		return rule.Listeners{
			ast.KindThisKeyword: func(node *ast.Node) {
				// The anchor. Upstream requires the immediate parent to be a member expression and
				// skips no parentheses, which is what makes `(this).props` silent while
				// `(this.props).foo` reports. See the rule doc for all seven measured forms.
				if node.Parent == nil {
					return
				}
				if node.Parent.Kind != ast.KindPropertyAccessExpression &&
					node.Parent.Kind != ast.KindElementAccessExpression {
					return
				}
				// A member expression's parent is also its object rather than its property name, so
				// `foo.this` cannot arise, but `a[this]` can: there the `this` is the ARGUMENT of an
				// element access and upstream's `is_member_expression_kind` on the parent answers
				// true for it just the same. Reproduced rather than narrowed, and pinned by a
				// fixture, because the check upstream makes is on the parent's kind alone.

				component := parentComponentOf(node)
				if component == nil {
					return
				}

				// The component's own ancestors decide the class-component exemption. Upstream
				// starts this walk at the component node itself.
				if hasReactComponentAncestor(component) {
					return
				}

				if isInNestedThisContext(node, component) {
					return
				}

				ctx.ReportNode(node, messageNoThisInSfc)
			},
		}
	},
}

// parentComponentOf walks outward from a `this` to the function component that owns it.
//
// This is oxc's `get_parent_component`, and its asymmetry between the two function kinds is the
// load-bearing part rather than an implementation detail. A `Function` ancestor RETURNS on sight,
// answering with itself only when it is component-named and answering nothing otherwise, which is
// how a nested function declaration or an anonymous function expression exempts its whole body. An
// arrow ancestor that is not component-named is stepped OVER and the walk continues, which is how
// an anonymous callback inside a component still belongs to that component.
//
// A class static block returns nothing, matching upstream's explicit `StaticBlock` arm. Without it
// the walk would escape a static block into the enclosing component, and upstream ships that case
// as passing.
//
// The walk is unbounded upward otherwise. It is not stopped by a class, which is why the separate
// exemption pass in `isInNestedThisContext` exists.
func parentComponentOf(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindArrowFunction:
			// Stepped over when it is not a component, so a callback keeps looking upward.
			if isPotentialReactComponent(current) {
				return current
			}
		case ast.KindClassStaticBlockDeclaration:
			return nil
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
			// Ends the walk either way. oxc's single `Function` node covers a declaration and an
			// expression alike; our tree spells that as two kinds.
			if isPotentialReactComponent(current) {
				return current
			}
			// This `return nil` is SUBSUMED rather than load-bearing, and the mutation sweep
			// proved it: rewriting it to `continue`, so the walk steps over a non-component
			// function the way it steps over a non-component arrow, survives the whole fixture
			// set. No input distinguishes the two, and the reason is one sentence: the `this` is
			// always a descendant of any function this arm would step over, so that same function
			// then sits strictly between the `this` and whatever component the widened walk finds,
			// and `isInNestedThisContext` exempts on a `Function` unconditionally. The second pass
			// answers every case this line would.
			//
			// Kept anyway, because it is upstream's structure rather than an optimization of it.
			// oxc writes `return is_potential_react_component(...).then_some(ancestor)`, which is
			// this exact early exit, and a port that dropped it would rest silently on the other
			// pass continuing to exempt functions. Deleting it would be safe today and would make
			// the next change to `isInNestedThisContext` able to break this rule without any
			// fixture noticing.
			return nil
		}
	}
	return nil
}

// isPotentialReactComponent reports whether a function carries a component-shaped name.
//
// oxc's `is_potential_react_component`, which is `get_function_name(...).is_some_and(...)`. The
// name is read from the function's own identifier for a declaration or expression, and from the
// enclosing variable declarator's binding identifier for an arrow. Nothing else is consulted: an
// anonymous function expression assigned to a capitalized variable answers false, and a named one
// answers on its own lowercase name rather than the variable's uppercase one. Both measured.
func isPotentialReactComponent(function *ast.Node) bool {
	name, ok := functionNameOf(function)
	return ok && isAsciiComponentName(name)
}

// functionNameOf reads the name upstream tests, or reports that there is none.
//
// The arrow case requires the declarator's name to be a plain binding identifier, so a destructured
// or otherwise patterned target answers nothing, matching oxc's `BindingPattern::BindingIdentifier`
// destructure. Measured: `const { Foo } = { Foo: (p) => this.props.a };` is silent upstream.
func functionNameOf(function *ast.Node) (string, bool) {
	switch function.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		name := function.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return "", false
		}
		return name.Text(), true
	case ast.KindArrowFunction:
		// Only the immediate parent is consulted. An arrow handed to `memo(...)` has a call
		// argument as its parent and therefore no name, which is why no wrapper is recognized.
		parent := function.Parent
		if parent == nil || parent.Kind != ast.KindVariableDeclaration {
			return "", false
		}
		name := parent.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return "", false
		}
		return name.Text(), true
	}
	return "", false
}

// isAsciiComponentName is oxc's `is_react_component_name`, which is ASCII-uppercase specifically.
//
// Deliberately not `internal/utilities/react.IsLikelyComponentName`, whose body is `unicode.IsUpper`
// and answers true for `Фoo`, `Λoo` and `Éoo`. All three are silent on the release binary. See the
// rule doc.
func isAsciiComponentName(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}

// hasReactComponentAncestor reports whether a class or ES5 React component encloses the node.
//
// oxc runs this over `ancestors(component_node.id())`, which excludes the component node itself.
// That distinction cannot change a verdict here, because the component is always a function and
// neither predicate answers true for one, but the walk starts at the parent to match upstream
// rather than relying on that.
//
// `isEs5ComponentCallStrict` rather than the shelf's `IsEs5ComponentCall`, because the shelf also
// accepts `createClass` and `React.createClass` while oxc keys on `createReactClass` alone, and the
// two disagree on inputs the release binary reports. See the rule doc.
func hasReactComponentAncestor(component *ast.Node) bool {
	for current := component.Parent; current != nil; current = current.Parent {
		if utilsreact.IsEs6ComponentClass(current) || isEs5ComponentCallStrict(current) {
			return true
		}
	}
	return false
}

// isInNestedThisContext reports whether something between the `this` and its component rebinds it.
//
// oxc's `is_in_nested_this_context`, walking the ancestors of the `this` and stopping when it
// reaches the component. Four kinds exempt, and the last one is conditional:
//
//	a function             a declaration or expression, which rebinds `this`
//	a method definition    a method, getter, setter or constructor
//	a property definition  unconditionally, wherever in it the `this` sits
//	an accessor property   only when the `this` falls inside the VALUE, not a computed key
//
// An arrow is deliberately absent, which is the point of the rule: an arrow inherits `this` from
// the component and so does not exempt.
//
// Our AST spells oxc's two property kinds as one `KindPropertyDeclaration` carrying an accessor
// modifier, so the modifier picks which arm applies. See the rule doc for the four measured rows.
func isInNestedThisContext(this *ast.Node, component *ast.Node) bool {
	for current := this.Parent; current != nil && current != component; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
			return true
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			return true
		case ast.KindPropertyDeclaration:
			if !ast.HasAccessorModifier(current) {
				// oxc's `PropertyDefinition` arm, which exempts with no condition at all.
				return true
			}
			// oxc's `AccessorProperty` arm, which asks whether the value's span contains the
			// `this`. Spelled here as containment in the tree rather than in the span, which
			// answers the same question without arithmetic: the value is a child node, so the
			// `this` is inside it exactly when walking up from the `this` reaches it. A computed
			// key is a sibling of the value rather than inside it, so it does not exempt.
			initializer := current.Initializer()
			if initializer != nil && isDescendantOf(this, initializer) {
				return true
			}
		}
	}
	return false
}

// isDescendantOf reports whether a node sits inside another, by walking parents.
//
// `ancestor` is included, so a node is a descendant of itself. That matches
// `span.contains_inclusive`, which is what upstream asks, and the `this` is never the initializer
// itself in any case because the initializer of an accessor property that IS a bare `this` would
// have no member parent and would have been declined at the anchor.
func isDescendantOf(node *ast.Node, ancestor *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}
