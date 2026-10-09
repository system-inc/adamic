package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// requireRenderReturnFile is where the fixtures pretend to live.
//
// A `.tsx` name because the corpus is JSX and needs a parser that reads it. It is NOT load-bearing:
// this rule has no file gate, and `TestRequireRenderReturnHasNoFileGate` pins that by running one
// reporting source under four extensions.
const requireRenderReturnFile = "/repository/source/RequireRenderReturn.tsx"

// The corpus is `eslint-plugin-react`'s own, extracted rather than retyped.
//
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/require-render-return.js` holds 13 valid
// and 4 invalid cases, each invalid one carrying exactly one `noRenderReturn`. Every string below
// was pulled out of that file by loading it with a stubbed `RuleTester` and serializing the
// captured object, then compared byte against byte back into the source with a script rather than
// by eye. All 17 were then run against the installed build, version 7.37.5, and every verdict below
// is that run's.
//
// There are no `options` and no `output` columns because `meta.schema` is `[]` and `meta.fixable`
// is unset. Three valid cases carry a `features: ['class fields']` marker, which selects a parser in
// upstream's harness and has no counterpart here, since our parser reads class fields
// unconditionally.
func TestRequireRenderReturnFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"upstream invalid 0", "\n        var Hello = createReactClass({\n          displayName: 'Hello',\n          render: function() {}\n        });\n      ", []string{"noRenderReturn"}},
		{"upstream invalid 1", "\n        class Hello extends React.Component {\n          render() {}\n        }\n      ", []string{"noRenderReturn"}},
		{"upstream invalid 2", "\n        class Hello extends React.Component {\n          render() {\n            const names = this.props.names.map(function(name) {\n              return <div>{name}</div>\n            });\n          }\n        }\n      ", []string{"noRenderReturn"}},
		{"upstream invalid 3", "\n        class Hello extends React.Component {\n          render = () => {\n            <div>Hello {this.props.name}</div>\n          }\n        }\n      ", []string{"noRenderReturn"}},
		// --- cases upstream does not cover, each measured against the installed build ---

		// The ES5 factory name is `createReactClass` alone. `createClass` and `React.createClass`
		// are both clean, which the two silent fixtures pin; this is the spelling that reports.
		{"es5 namespaced createReactClass", "var Hello = React.createReactClass({\n  render: function() {}\n});\n", []string{"noRenderReturn"}},

		// Parentheses in the heritage clause, all four spellings, all reporting.
		//
		// espree produces no parenthesized-expression node at all, so `node.superClass` is the bare
		// member expression however the parens are written and upstream never sees them. Our parser
		// does produce the node, and the shelf's `IsEs6ComponentClass` skips it on the RECEIVER
		// only, so `(React).Component` answers true there while `(React.Component)` answers false.
		// The first and third of these fail against the shelf helper and pass against the local
		// one, which is the entire reason `isRenderReturnEs6ComponentClass` exists. Measured on
		// the installed build; nothing in the corpus writes a parenthesis.
		{"parens around the whole heritage expression", "class Hello extends (React.Component) {\n  render() {}\n}\n", []string{"noRenderReturn"}},
		{"parens around the heritage receiver", "class Hello extends (React).Component {\n  render() {}\n}\n", []string{"noRenderReturn"}},
		{"nested parens around the heritage expression", "class Hello extends ((React.Component)) {\n  render() {}\n}\n", []string{"noRenderReturn"}},
		{"parens around a bare base name", "class Hello extends (Component) {\n  render() {}\n}\n", []string{"noRenderReturn"}},

		// The ES5 factory callee in parentheses. `isEs5ComponentCallStrict` skips them and upstream
		// agrees, so this is the one paren position where the shelf and the authority already
		// match.
		{"parens around the es5 factory callee", "var Hello = (createReactClass)({\n  render: function() {}\n});\n", []string{"noRenderReturn"}},

		// The bare base names, and the pure variant. All four report.
		{"bare Component base", "class Hello extends Component {\n  render() {}\n}\n", []string{"noRenderReturn"}},
		{"bare PureComponent base", "class Hello extends PureComponent {\n  render() {}\n}\n", []string{"noRenderReturn"}},
		{"namespaced PureComponent base", "class Hello extends React.PureComponent {\n  render() {}\n}\n", []string{"noRenderReturn"}},

		// A class expression is a component too, which `Components.detect` registers alongside the
		// declaration and the rule's own filter keeps.
		{"class expression component", "var Hello = class extends React.Component {\n  render() {}\n};\n", []string{"noRenderReturn"}},

		// An arrow with a BLOCK body has no implicit return, so it falls through to the walk and
		// reports. Its expression-bodied twin is a silent fixture below, and the pair is what
		// separates the two listeners.
		{"arrow render with empty block body", "class Hello extends React.Component {\n  render = () => {}\n}\n", []string{"noRenderReturn"}},
		{"es5 arrow render with empty block body", "var Hello = createReactClass({\n  render: () => {}\n});\n", []string{"noRenderReturn"}},

		// A `render` property whose value is a function expression rather than an arrow.
		{"render property holding a function expression", "class Hello extends React.Component {\n  render = function() {}\n}\n", []string{"noRenderReturn"}},

		// Accessors and modifiers. Upstream sees each of these as a `MethodDefinition` whose
		// `value` is a `FunctionExpression`, so all five report; our tree splits the accessors into
		// their own kinds, which is why they are matched explicitly in the rule.
		{"static render", "class Hello extends React.Component {\n  static render() {}\n}\n", []string{"noRenderReturn"}},
		{"getter render", "class Hello extends React.Component {\n  get render() {}\n}\n", []string{"noRenderReturn"}},
		{"setter render", "class Hello extends React.Component {\n  set render(v) {}\n}\n", []string{"noRenderReturn"}},
		{"async render", "class Hello extends React.Component {\n  async render() {}\n}\n", []string{"noRenderReturn"}},
		{"generator render", "class Hello extends React.Component {\n  *render() {}\n}\n", []string{"noRenderReturn"}},

		// The scope count. An arrow costs a scope exactly as a function expression does, because
		// upstream's `/Function(Expression|Declaration)$/` is unanchored at the front and
		// `ArrowFunctionExpression` matches it. Reading arrows as transparent is the natural wrong
		// answer and these three are what catch it.
		{"return only inside a nested arrow", "class Hello extends React.Component {\n  render() {\n    var f = () => { return 1; };\n  }\n}\n", []string{"noRenderReturn"}},
		{"return only inside a nested function declaration", "class Hello extends React.Component {\n  render() {\n    function f() { return 1; }\n  }\n}\n", []string{"noRenderReturn"}},
		{"es5 return only inside a nested arrow", "var Hello = createReactClass({\n  render: function() { var f = () => { return 1; }; }\n});\n", []string{"noRenderReturn"}},

		// The per-component flag, in the direction that reports. Both renders empty reports once,
		// on the FIRST, which is `findRenderMethod` taking the first match.
		{"two empty render methods report once on the first", "class Hello extends React.Component {\n  render() {}\n  render() {}\n}\n", []string{"noRenderReturn"}},

		// A return in a sibling method does not satisfy `render`: the ancestor named `render` is
		// never reached on the climb.
		{"return only in a sibling method", "class Hello extends React.Component {\n  foo() { return 1; }\n  render() {}\n}\n", []string{"noRenderReturn"}},

		// Two broken components in one file report twice, in source order.
		{"two broken components", "class A extends React.Component { render() {} }\nclass B extends React.Component { render() {} }\n", []string{"noRenderReturn", "noRenderReturn"}},

		// Nested components are each judged on their own. The inner one is registered, so the
		// return inside it marks the inner and leaves the outer unmarked.
		{"inner component returns, outer does not", "class Outer extends React.Component {\n  render() {\n    class Inner extends React.Component { render() { return <div/>; } }\n  }\n}\n", []string{"noRenderReturn"}},

		// An implicit-return arrow whose parent property is NOT named `render` must not satisfy
		// anything. Upstream's second listener returns early on exactly this test, and a mutation
		// removing it survived every other fixture: the arm then marks the enclosing component for
		// any expression-bodied arrow anywhere inside it, which is three of the commonest shapes
		// in real code. All three measured as reporting on the installed build.
		{"implicit arrow bound to a plain variable", "class Hello extends React.Component {\n  render() {\n    var f = () => <div/>;\n  }\n}\n", []string{"noRenderReturn"}},
		{"implicit arrow on a property named something else", "class Hello extends React.Component {\n  render() {\n    var o = { other: () => <div/> };\n  }\n}\n", []string{"noRenderReturn"}},
		{"implicit arrow passed as a callback argument", "class Hello extends React.Component {\n  render() {\n    [].map(x => x + 1);\n  }\n}\n", []string{"noRenderReturn"}},

		// A `render` arrow on a plain object elsewhere in the file does not reach this component,
		// because the climb from that arrow stops at the nearest enclosing component and there is
		// none. Its nested twin, which DOES reach, is a silent fixture below.
		{"stray render arrow outside the component", "class Hello extends React.Component {\n  render() {}\n}\nvar o = { render: () => <div/> };\n", []string{"noRenderReturn"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, RequireRenderReturn, requireRenderReturnFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// The clean cases. Upstream's thirteen, plus the ones that pin a decision the corpus never writes.
func TestRequireRenderReturnStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"upstream valid 0", "\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      "},
		{"upstream valid 1", "\n        class Hello extends React.Component {\n          render = () => {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      "},
		{"upstream valid 2", "\n        class Hello extends React.Component {\n          render = () => (\n            <div>Hello {this.props.name}</div>\n          )\n        }\n      "},
		{"upstream valid 3", "\n        var Hello = createReactClass({\n          displayName: 'Hello',\n          render: function() {\n            return <div></div>\n          }\n        });\n      "},
		{"upstream valid 4", "\n        function Hello() {\n          return <div></div>;\n        }\n      "},
		{"upstream valid 5", "\n        var Hello = () => (\n          <div></div>\n        );\n      "},
		{"upstream valid 6", "\n        var Hello = createReactClass({\n          render: function() {\n            switch (this.props.name) {\n              case 'Foo':\n                return <div>Hello Foo</div>;\n              default:\n                return <div>Hello {this.props.name}</div>;\n            }\n          }\n        });\n      "},
		{"upstream valid 7", "\n        var Hello = createReactClass({\n          render: function() {\n            if (this.props.name === 'Foo') {\n              return <div>Hello Foo</div>;\n            } else {\n              return <div>Hello {this.props.name}</div>;\n            }\n          }\n        });\n      "},
		{"upstream valid 8", "\n        class Hello {\n          render() {}\n        }\n      "},
		{"upstream valid 9", "class Hello extends React.Component {}"},
		{"upstream valid 10", "var Hello = createReactClass({});"},
		{"upstream valid 11", "\n        var render = require('./render');\n        var Hello = createReactClass({\n          render\n        });\n      "},
		{"upstream valid 12", "\n        class Foo extends Component {\n          render\n        }\n      "},
		// --- cases upstream does not cover, each measured against the installed build ---

		// The factory name set is exactly `createReactClass`. Both `createClass` spellings are
		// clean, which is narrower than the shelf's `IsEs5ComponentCall` and is why the rule calls
		// `isEs5ComponentCallStrict`.
		{"bare createClass is not a component", "var Hello = createClass({\n  render: function() {}\n});\n"},
		{"namespaced React.createClass is not a component", "var Hello = React.createClass({\n  render: function() {}\n});\n"},

		// The parenthesized factory callee with a render that DOES return. Paired with the
		// reporting fixture above, this separates the parenthesis handling from the verdict.
		{"parenthesized factory callee with a returning render", "var Hello = (createReactClass)({\n  render: function() { return <div/>; }\n});\n"},

		// A base class that is not React's.
		{"unrelated namespace base", "class Hello extends Foo.Component {\n  render() {}\n}\n"},
		{"unrelated base name", "class Hello extends Whatever {\n  render() {}\n}\n"},

		// The key spelling. `getPropertyName` reads `nameNode.name`, which is undefined for a
		// string-literal key and for a computed one, so all four of these are clean upstream. This
		// is why the rule compares the identifier spelling directly instead of reaching for
		// `ast.TryGetTextOfPropertyName`, which resolves both forms and would report on all four.
		{"string-literal render key in a class", "class Hello extends React.Component {\n  'render'() {}\n}\n"},
		{"string-literal render key in an object", "var Hello = createReactClass({\n  'render': function() {}\n});\n"},
		{"computed render key in a class", "class Hello extends React.Component {\n  ['render']() {}\n}\n"},
		{"computed render key in an object", "var Hello = createReactClass({\n  ['render']: function() {}\n});\n"},

		// The value has to be function-like. Four ways it is not, two of which upstream ships as
		// clean cases and two measured beside them.
		{"render property holding a number", "class Hello extends React.Component {\n  render = 5\n}\n"},
		{"render property holding an identifier", "class Hello extends React.Component {\n  render = someFn\n}\n"},
		{"es5 render holding an identifier", "var Hello = createReactClass({\n  render: someFn\n});\n"},

		// The implicit-return arrow, which has no return statement anywhere and is caught by the
		// second listener alone. Three spellings of it.
		{"arrow render returning jsx implicitly", "class Hello extends React.Component {\n  render = () => <div/>\n}\n"},
		{"arrow render returning null implicitly", "class Hello extends React.Component {\n  render = () => null\n}\n"},
		{"es5 arrow render returning implicitly", "var Hello = createReactClass({\n  render: () => <div/>\n});\n"},

		// A bare `return;` with no argument still satisfies the rule. Upstream marks on the
		// statement's presence and never reads its argument.
		{"bare return with no value", "class Hello extends React.Component {\n  render() { return; }\n}\n"},

		// Blocks, conditionals and switches are transparent to the scope count, which is two of
		// upstream's clean cases and this third shape beside them.
		{"return inside an if block", "class Hello extends React.Component {\n  render() { if (a) { return <div/>; } }\n}\n"},

		// The per-component flag, in the direction that goes quiet. A return in EITHER render marks
		// the whole component, so both orderings are clean while the both-empty pair reports. This
		// is upstream being loose and it is reproduced.
		{"second render returns, first is empty", "class Hello extends React.Component {\n  render() {}\n  render() { return <div/>; }\n}\n"},
		{"first render returns, second is empty", "class Hello extends React.Component {\n  render() { return <div/>; }\n  render() {}\n}\n"},

		// The measured false negative, and the reason the flag climbs to the nearest REGISTERED
		// component rather than to the nearest class. `Inner` is not a component, so the return
		// inside it climbs past and marks `Outer`. Its twin where `Inner` IS a component is a
		// reporting fixture above, and the pair is the only thing that can see this decision.
		{"return inside a nested plain class marks the outer component", "class Outer extends React.Component {\n  render() {\n    class Inner { render() { return <div/>; } }\n  }\n}\n"},

		// The same looseness through the arrow listener: an implicit-return `render` arrow on a
		// plain object nested in the broken render marks the enclosing component.
		{"nested implicit render arrow marks the outer component", "class Hello extends React.Component {\n  render() {\n    var o = { render: () => <div/> };\n  }\n}\n"},

		// A `render` nested one object deeper is not the component's `render`.
		{"render on a nested object is not the component's", "var Hello = createReactClass({\n  foo: { render: function() {} }\n});\n"},

		// A component with no `render` at all, and a component whose only empty method is not
		// `render`. Both are `renderProperty == nil`.
		{"component with a non-render empty method", "class Hello extends React.Component {\n  foo() {}\n}\n"},

		// The wrappers `Components.detect` registers and the rule's own filter throws away.
		{"memo-wrapped object with an empty render", "var Hello = React.memo({ render: function() {} });\n"},
		{"plain object with an empty render", "var Hello = { render: function() {} };\n"},

		// A JSDoc `@extends React.Component` tag is upstream's `isExplicitComponent` branch and is
		// clean here on the installed build, which is why the branch is not ported.
		{"jsdoc extends tag", "/**\n * @extends React.Component\n */\nclass Hello extends Foo {\n  render() {}\n}\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, RequireRenderReturn, requireRenderReturnFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The span, which no `ExpectFindings` assertion can see.
//
// Upstream reports `node: findRenderMethod(component.node)`, and `findRenderMethod` returns the
// PROPERTY rather than its value, so the finding covers the whole `render() {}` or
// `render: function() {}` including the empty body. A port anchoring on the function would carry
// the right message id on every fixture above while pointing at the wrong node, and the four
// upstream cases would all still pass.
//
// Each expectation below is the exact text the installed build underlines, taken from its
// reported columns rather than from reading the rule. `rule_testing.Run` does not trim the fixture, so
// the slice out of the literal and the bytes on disk are the same string.
func TestRequireRenderReturnSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		reported   string
	}{
		{
			"class method covers the empty body",
			"class Hello extends React.Component {\n  render() {}\n}\n",
			"render() {}",
		},
		{
			"es5 property covers the key and the function",
			"var Hello = createReactClass({\n  render: function() {}\n});\n",
			"render: function() {}",
		},
		{
			"class property covers the whole declaration across lines",
			"class Hello extends React.Component {\n  render = () => {\n    <div/>\n  }\n}\n",
			"render = () => {\n    <div/>\n  }",
		},
		{
			"static modifier is inside the span",
			"class Hello extends React.Component {\n  static render() {}\n}\n",
			"static render() {}",
		},
		{
			"getter keyword is inside the span",
			"class Hello extends React.Component {\n  get render() {}\n}\n",
			"get render() {}",
		},
		{
			// Both must be empty. A return in EITHER marks the whole component, so a middle
			// `render` that returns makes all three clean, which is measured and is the
			// per-component flag rather than a per-method one. Writing this case with a returning
			// render was the first draft's mistake and the rule was right.
			//
			// The two renders must also read DIFFERENTLY, and this is the case's whole point.
			// `findRenderMethod` takes the FIRST match, and a mutation flipping that to the last
			// survived every fixture while two identical `render() {}` bodies were the probe: the
			// slices compared equal whichever one was reported. Giving the second a parameter is
			// what makes the choice visible. Measured on the installed build, which reports the
			// first in both the class and the object spellings.
			"the first of two empty renders is the one reported",
			"class Hello extends React.Component {\n  render() {}\n  render(a) {}\n}\n",
			"render() {}",
		},
		{
			"the first of two empty es5 renders is the one reported",
			"var Hello = createReactClass({\n  render: function() {},\n  render: function(a) {}\n});\n",
			"render: function() {}",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, RequireRenderReturn, requireRenderReturnFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.reported {
				t.Errorf("finding points at %q, want %q", reported, testCase.reported)
			}
		})
	}
}

// Both spans of a two-component file, in order.
//
// `ExpectFindings` asserts the count and the ids and cannot see that the second finding points at
// the second component. A rule collecting into a map rather than a slice passes that assertion and
// emits these in whatever order the map iterated.
func TestRequireRenderReturnSpansAreInSourceOrder(t *testing.T) {
	t.Parallel()

	sourceText := "class A extends React.Component { render() { } }\nvar B = createReactClass({ render: function() { } });\n"
	result := rule_testing.Run(t, RequireRenderReturn, requireRenderReturnFile, sourceText)
	if len(result.Diagnostics) != 2 {
		t.Fatalf("want 2 diagnostics, got %d", len(result.Diagnostics))
	}
	want := []string{"render() { }", "render: function() { }"}
	for i, expected := range want {
		diagnostic := result.Diagnostics[i]
		reported := sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != expected {
			t.Errorf("diagnostic %d points at %q, want %q", i, reported, expected)
		}
	}
}

// The message, asserted by equality rather than by containment.
//
// `rule.Message` is `{Id, Description}` with no interpolation layer, so there is nothing to render
// and nothing a format string can move. What this guards is the pair staying attached to each
// other: a rule reporting the right id with a description copied from a neighbour passes every
// assertion above.
func TestRequireRenderReturnMessage(t *testing.T) {
	t.Parallel()

	if messageNoRenderReturn.Id != "noRenderReturn" {
		t.Errorf("message id is %q, want %q", messageNoRenderReturn.Id, "noRenderReturn")
	}
	const wantPrefix = "A class component's `render` returns nothing, so React receives `undefined` where it expected an element."
	if !strings.HasPrefix(messageNoRenderReturn.Description, wantPrefix) {
		t.Errorf("description starts %q, want prefix %q", messageNoRenderReturn.Description, wantPrefix)
	}
}

// There is no file gate, and three rules in this package have one.
//
// `no-did-mount-set-state`, `no-direct-mutation-state` and `no-this-in-sfc` all decline any file
// that is not `.tsx` or `.jsx`, because they were ported from oxc, which declares `should_run` on
// `source_type().is_jsx()`. `eslint-plugin-react` has no such gate: measured, the same class
// reports identically under every extension. Carrying the gate here would silence the rule in every
// `.ts` file, which is most of this tree, and no imported fixture could see it because the whole
// corpus is JSX.
//
// The source below is deliberately free of JSX so a `.ts` parse is legal.
func TestRequireRenderReturnHasNoFileGate(t *testing.T) {
	t.Parallel()

	const sourceText = "class Hello extends React.Component {\n  render() {}\n}\n"
	for _, fileName := range []string{
		"/repository/source/Probe.tsx",
		"/repository/source/Probe.ts",
		"/repository/source/Probe.jsx",
		"/repository/source/Probe.js",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, RequireRenderReturn, fileName, sourceText)
			rule_testing.ExpectFindings(t, result, "noRenderReturn")
		})
	}
}

// The rule takes no options, and a nil options value must not change what it decides.
//
// A rule configured as bare `"error"` is handed nil rather than a zero struct, and a rule reading
// options without a fallback yields the zero value silently. This rule has no `Decode` at all and
// therefore no decoder to go wrong, so this pins that the absence is deliberate: passing nil
// explicitly reaches the rule the same way the live config does, and it still reports.
func TestRequireRenderReturnIgnoresOptions(t *testing.T) {
	t.Parallel()

	const sourceText = "class Hello extends React.Component {\n  render() {}\n}\n"
	result := rule_testing.RunWithOptions(t, RequireRenderReturn, requireRenderReturnFile, sourceText, nil)
	rule_testing.ExpectFindings(t, result, "noRenderReturn")
}
