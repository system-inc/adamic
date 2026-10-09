package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is upstream's, extracted by stubbing its RuleTester and capturing what the test file
// hands it, so no case was retyped and no escape was typed by hand. All 12 cases were replayed
// against the installed build first and every one reproduced, so the counts here are measured.
const noObjectTypeAsDefaultPropFile = "/repository/source/NoObjectTypeAsDefaultProp.tsx"

func TestNoObjectTypeAsDefaultPropStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{"upstream valid 0", "\n      function Foo({\n        bar = emptyFunction,\n      }) {\n        return null;\n      }\n    "},
		{"upstream valid 1", "\n      function Foo({\n        bar = emptyFunction,\n        ...rest\n      }) {\n        return null;\n      }\n    "},
		{"upstream valid 2", "\n      function Foo({\n        bar = 1,\n        baz = 'hello',\n      }) {\n        return null;\n      }\n    "},
		{"upstream valid 3", "\n      function Foo(props) {\n        return null;\n      }\n    "},
		{"upstream valid 4", "\n      function Foo(props) {\n        return null;\n      }\n\n      Foo.defaultProps = {\n        bar: () => {}\n      }\n    "},
		{"upstream valid 5", "\n      const Foo = () => {\n        return null;\n      };\n    "},
		{"upstream valid 6", "\n      const Foo = ({bar = 1}) => {\n        return null;\n      };\n    "},
		{"upstream valid 7", "\n      const Foo = ({bar = 1}, context) => {\n        return null;\n      };\n    "},
		{"upstream valid 8", "\n      export default function NotAComponent({foo = {}}) {}\n    "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoObjectTypeAsDefaultProp, noObjectTypeAsDefaultPropFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoObjectTypeAsDefaultPropFires asserts the count AND the rendered message text of every
// finding. The message interpolates the property name and the forbidden type, and names the type
// twice, so a message-id assertion could not see an interpolation defect at all: the id and the
// count stay right while the text goes wrong. Each expected string was captured from the
// installed build rather than composed here.
func TestNoObjectTypeAsDefaultPropFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		wantMessages     []string
	}{
		{"upstream invalid 0", "\n        function Foo({\n          a = {},\n          b = ['one', 'two'],\n          c = /regex/i,\n          d = () => {},\n          e = function() {},\n          f = class {},\n          g = new Thing(),\n          h = <Thing />,\n          i = Symbol('foo')\n        }) {\n          return null;\n        }\n      ", []string{
			"a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.",
			"b has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.",
			"c has a/an regex literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of regex literal.",
			"d has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function.",
			"e has a/an function expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of function expression.",
			"f has a/an class expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of class expression.",
			"g has a/an construction expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of construction expression.",
			"h has a/an JSX element as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of JSX element.",
			"i has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal.",
		}},
		{"upstream invalid 1", "\n        const Foo = ({\n          a = {},\n          b = ['one', 'two'],\n          c = /regex/i,\n          d = () => {},\n          e = function() {},\n          f = class {},\n          g = new Thing(),\n          h = <Thing />,\n          i = Symbol('foo')\n        }) => {\n          return null;\n        }\n      ", []string{
			"a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.",
			"b has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.",
			"c has a/an regex literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of regex literal.",
			"d has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function.",
			"e has a/an function expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of function expression.",
			"f has a/an class expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of class expression.",
			"g has a/an construction expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of construction expression.",
			"h has a/an JSX element as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of JSX element.",
			"i has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal.",
		}},
		{"upstream invalid 2", "\n        const Foo = ({\n          a = {},\n          b = ['one', 'two'],\n          c = /regex/i,\n          d = () => {},\n          e = function() {},\n          f = class {},\n          g = new Thing(),\n          h = <Thing />,\n          i = Symbol('foo')\n        }, context) => {\n          return null;\n        }\n      ", []string{
			"a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.",
			"b has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.",
			"c has a/an regex literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of regex literal.",
			"d has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function.",
			"e has a/an function expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of function expression.",
			"f has a/an class expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of class expression.",
			"g has a/an construction expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of construction expression.",
			"h has a/an JSX element as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of JSX element.",
			"i has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal.",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoObjectTypeAsDefaultProp, noObjectTypeAsDefaultPropFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantMessages) {
				t.Fatalf("want %d findings, got %d", len(testCase.wantMessages), len(result.Diagnostics))
			}
			for index, diagnostic := range result.Diagnostics {
				if diagnostic.Message.Description != testCase.wantMessages[index] {
					t.Errorf("finding %d renders\n  got  %q\n  want %q",
						index, diagnostic.Message.Description, testCase.wantMessages[index])
				}
			}
		})
	}
}

// TestNoObjectTypeAsDefaultPropMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite is a
// differential table.
//
// Every row was run against the installed `eslint-plugin-react` build on 2026-08-27, and the
// expectation is the exact message text that build produced, so a wrong count and a wrong
// interpolation both fail here. Upstream's corpus is only twelve cases and writes almost none of
// these, and the rows pin four things it cannot see:
//
//	the forbidden set boundary   nine shapes report and everything else is silent, including a
//	                             plain call, a template literal, a tagged template, a member
//	                             access, a conditional, and every primitive
//	the Symbol special case       the callee must be the bare identifier, so `Symbol.for(1)` is
//	                             silent while `Symbol(1)` reports
//	the first parameter only      a destructured SECOND parameter is silent, and a nested pattern
//	                             does not recurse. Both look like upstream oversights and both are
//	                             reproduced rather than corrected.
//	which name the message uses   an explicit `{key: local = {}}` names the KEY, not the binding
//
// The parenthesized rows exist because our parser keeps a node upstream's parser folds away, so
// they are a shape upstream's corpus structurally cannot express.
func TestNoObjectTypeAsDefaultPropMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantMessages []string
	}{
		{"default is a object", "function C({a = {}}) { return <div/>; }", []string{"a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal."}},
		{"default is a array", "function C({a = []}) { return <div/>; }", []string{"a has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal."}},
		{"default is a arrow", "function C({a = ()=>{}}) { return <div/>; }", []string{"a has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function."}},
		{"default is a function expression", "function C({a = function(){}}) { return <div/>; }", []string{"a has a/an function expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of function expression."}},
		{"default is a class expression", "function C({a = class {}}) { return <div/>; }", []string{"a has a/an class expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of class expression."}},
		{"default is a new expression", "function C({a = new Foo()}) { return <div/>; }", []string{"a has a/an construction expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of construction expression."}},
		{"default is a jsx element", "function C({a = <div/>}) { return <div/>; }", []string{"a has a/an JSX element as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of JSX element."}},
		{"default is a jsx fragment", "function C({a = <></>}) { return <div/>; }", []string{}},
		{"default is a jsx self closing", "function C({a = <Foo/>}) { return <div/>; }", []string{"a has a/an JSX element as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of JSX element."}},
		{"default is a regex", "function C({a = /x/}) { return <div/>; }", []string{"a has a/an regex literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of regex literal."}},
		{"default is a regex with flags", "function C({a = /x/gi}) { return <div/>; }", []string{"a has a/an regex literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of regex literal."}},
		{"default is a Symbol call", "function C({a = Symbol()}) { return <div/>; }", []string{"a has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal."}},
		{"default is a Symbol with argument", "function C({a = Symbol(1)}) { return <div/>; }", []string{"a has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal."}},
		{"default is a Symbol.for member callee", "function C({a = Symbol.for(1)}) { return <div/>; }", []string{}},
		{"default is a symbol lowercase", "function C({a = symbol()}) { return <div/>; }", []string{}},
		{"default is a string", "function C({a = \"s\"}) { return <div/>; }", []string{}},
		{"default is a number", "function C({a = 1}) { return <div/>; }", []string{}},
		{"default is a null", "function C({a = null}) { return <div/>; }", []string{}},
		{"default is a undefined", "function C({a = undefined}) { return <div/>; }", []string{}},
		{"default is a boolean", "function C({a = true}) { return <div/>; }", []string{}},
		{"default is a identifier", "function C({a = ref}) { return <div/>; }", []string{}},
		{"default is a plain call", "function C({a = foo()}) { return <div/>; }", []string{}},
		{"default is a template literal", "function C({a = `x`}) { return <div/>; }", []string{}},
		{"default is a tagged template", "function C({a = tag`x`}) { return <div/>; }", []string{}},
		{"default is a member access", "function C({a = a.b}) { return <div/>; }", []string{}},
		{"default is a parenthesized object", "function C({a = ({})}) { return <div/>; }", []string{"a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal."}},
		{"default is a double parenthesized object", "function C({a = (({}))}) { return <div/>; }", []string{"a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal."}},
		{"default is a parenthesized Symbol call", "function C({a = (Symbol())}) { return <div/>; }", []string{"a has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal."}},
		{"default is a conditional", "function C({a = a?{}:{}}) { return <div/>; }", []string{}},
		{"default is a negative number", "function C({a = -1}) { return <div/>; }", []string{}},
		{"default is a void expression", "function C({a = void 0}) { return <div/>; }", []string{}},
		{"component returning null still counts", "function C({a = {}}) { return null; }", []string{"a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal."}},
		{"lowercase name is not a component", "function c({a = {}}) { return <div/>; }", []string{}},
		{"helper returning a number", "function helper({a = {}}) { return 1; }", []string{}},
		{"arrow component", "const C = ({a = {}}) => <div/>;", []string{"a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal."}},
		{"second parameter destructured", "function C(p, {a = {}}) { return <div/>; }", []string{}},
		{"nested destructuring does not recurse", "function C({x: {a = {}}}) { return <div/>; }", []string{}},
		{"explicit key names the key not the binding", "function C({key: local = {}}) { return <div/>; }", []string{"key has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal."}},
		{"rest element", "function C({...rest}) { return <div/>; }", []string{}},
		{"no default", "function C({a}) { return <div/>; }", []string{}},
		{"two offending defaults", "function C({a = {}, b = []}) { return <div/>; }", []string{"a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.", "b has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal."}},
		{"offending and clean mixed", "function C({a = {}, b = 1}) { return <div/>; }", []string{"a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal."}},
		{"memo wrapped component", "import {memo} from 'react';\nconst C = memo(({a = {}}) => <div/>);", []string{}},
		{"forwardRef wrapped component", "import {forwardRef} from 'react';\nconst C = forwardRef(({a = {}}, ref) => <div/>);", []string{}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoObjectTypeAsDefaultProp,
				noObjectTypeAsDefaultPropFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantMessages) {
				t.Fatalf("installed build reports %d findings, this rule reports %d",
					len(testCase.wantMessages), len(result.Diagnostics))
			}
			for index, diagnostic := range result.Diagnostics {
				if diagnostic.Message.Description != testCase.wantMessages[index] {
					t.Errorf("finding %d renders\n  got  %q\n  want %q",
						index, diagnostic.Message.Description, testCase.wantMessages[index])
				}
			}
		})
	}
}

// TestNoObjectTypeAsDefaultPropAnchorsOnTheWholeAssignment pins where each finding points.
//
// Upstream reports on `prop.value`, which is the whole assignment pattern rather than just the
// offending value, so the span covers the name, the equals sign and the default. Measured by
// slicing the installed build's reported range out of the source on 2026-08-27.
func TestNoObjectTypeAsDefaultPropAnchorsOnTheWholeAssignment(t *testing.T) {
	t.Parallel()

	sourceText := "function C({a = {}, b = []}) { return <div/>; }"
	wantSpans := []string{"a = {}", "b = []"}

	result := rule_testing.RunTyped(t, NoObjectTypeAsDefaultProp, noObjectTypeAsDefaultPropFile,
		sourceText)
	if len(result.Diagnostics) != len(wantSpans) {
		t.Fatalf("want %d findings, got %d", len(wantSpans), len(result.Diagnostics))
	}
	for index, diagnostic := range result.Diagnostics {
		reported := sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != wantSpans[index] {
			t.Errorf("finding %d points at %q, want %q", index, reported, wantSpans[index])
		}
	}
}

// TestNoObjectTypeAsDefaultPropRequiresTheTypedHarness asserts the rule declines rather than
// reports when it has no checker.
//
// The rule declares NeedsTypeChecker because its component detector resolves a bare `createElement`
// through the checker. A typed rule run on the plain harness receives a nil checker and, without
// the guard, would answer a narrower question while every quiet fixture passed vacuously.
func TestNoObjectTypeAsDefaultPropRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	sourceText := "function C({a = {}}) { return <div/>; }"

	// The control. On the typed harness this reports, so the zero below means the guard fired
	// rather than that the input was uninteresting.
	typed := rule_testing.RunTyped(t, NoObjectTypeAsDefaultProp, noObjectTypeAsDefaultPropFile,
		sourceText)
	rule_testing.ExpectFindings(t, typed, "forbiddenTypeDefaultParam")

	untyped := rule_testing.Run(t, NoObjectTypeAsDefaultProp, noObjectTypeAsDefaultPropFile,
		sourceText)
	rule_testing.ExpectClean(t, untyped)
}

// TestNoObjectTypeAsDefaultPropHasNoFileSuffixGate pins the absence of the gate three siblings in
// this package carry.
//
// Those siblings were ported from oxc, which gates on the file being read as JSX. This rule's
// authority has no filename condition anywhere. The source returns null rather than JSX so a `.ts`
// file parses cleanly and the parser's opinion cannot be mistaken for the rule's; the `.ts` row is
// exactly the file a `.tsx` gate would have silenced.
func TestNoObjectTypeAsDefaultPropHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	sourceText := "function C({a = {}}) { return null; }\n"
	for _, fileName := range []string{
		"/repository/source/Suffix.tsx",
		"/repository/source/Suffix.ts",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoObjectTypeAsDefaultProp, fileName, sourceText)
			rule_testing.ExpectFindings(t, result, "forbiddenTypeDefaultParam")
		})
	}
}

// TestNoObjectTypeAsDefaultPropRendersRealKeyNamesWhereUpstreamSaysUndefined pins the one place
// this port's output differs from the installed build, and pins that the difference is confined to
// the message text.
//
// Measured on 2026-08-27, all four rows against the installed build. A string key, a numeric key
// and a computed key all REPORT, exactly as a plain identifier key does, so which inputs produce a
// finding is reproduced with no divergence at all.
//
// What differs is the rendered name. Upstream reads `prop.key.name`, which a string or numeric key
// node does not have, so its message says literally "undefined has a/an object literal as default
// prop". Our node carries the real key, and rendering the word "undefined" into text a person reads
// is a defect this port can see, so these say "str" and "1" instead. The computed key renders as
// "k" in both, because upstream's key node for that shape IS the inner identifier.
//
// The divergence is deliberate, one-directional, and confined to the human-readable half of the
// finding. If this is ever judged wrong, the fix is in `destructuredPropertyName` and this test is
// the record of what upstream actually did.
func TestNoObjectTypeAsDefaultPropRendersRealKeyNamesWhereUpstreamSaysUndefined(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		sourceText       string
		wantMessageStart string
		upstreamRenders  string
	}{
		{
			name:             "identifier key agrees with upstream",
			sourceText:       "function C({a = {}}) { return <div/>; }",
			wantMessageStart: "a has a/an object literal",
			upstreamRenders:  "a has a/an object literal",
		},
		{
			name:             "computed key agrees with upstream",
			sourceText:       "function C({[k]: a = {}}) { return <div/>; }",
			wantMessageStart: "k has a/an object literal",
			upstreamRenders:  "k has a/an object literal",
		},
		{
			name:             "computed call key reports, where upstream renders undefined",
			sourceText:       "function C({[k()]: a = {}}) { return <div/>; }",
			wantMessageStart: "k() has a/an object literal",
			upstreamRenders:  "undefined has a/an object literal",
		},
		{
			name:             "computed string key reports, where upstream renders undefined",
			sourceText:       "function C({['lit']: a = {}}) { return <div/>; }",
			wantMessageStart: "'lit' has a/an object literal",
			upstreamRenders:  "undefined has a/an object literal",
		},
		{
			name:             "computed arithmetic key reports, where upstream renders undefined",
			sourceText:       "function C({[1 + 1]: a = {}}) { return <div/>; }",
			wantMessageStart: "1 + 1 has a/an object literal",
			upstreamRenders:  "undefined has a/an object literal",
		},
		{
			name:             "computed member key reports, where upstream renders undefined",
			sourceText:       "function C({[o.p]: a = {}}) { return <div/>; }",
			wantMessageStart: "o.p has a/an object literal",
			upstreamRenders:  "undefined has a/an object literal",
		},
		{
			name:             "string key renders the key where upstream renders undefined",
			sourceText:       "function C({'str': a = {}}) { return <div/>; }",
			wantMessageStart: "str has a/an object literal",
			upstreamRenders:  "undefined has a/an object literal",
		},
		{
			name:             "numeric key renders the key where upstream renders undefined",
			sourceText:       "function C({1: a = {}}) { return <div/>; }",
			wantMessageStart: "1 has a/an object literal",
			upstreamRenders:  "undefined has a/an object literal",
		},
	}

	const messageTail = " as default prop. This could lead to potential infinite render loop in " +
		"React. Use a variable reference instead of object literal."

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoObjectTypeAsDefaultProp,
				noObjectTypeAsDefaultPropFile, testCase.sourceText)

			// Every row reports, including the two that diverge. That is the half with no
			// divergence and it is asserted first, because it is the half that matters.
			rule_testing.ExpectFindings(t, result, "forbiddenTypeDefaultParam")

			want := testCase.wantMessageStart + messageTail
			if result.Diagnostics[0].Message.Description != want {
				t.Errorf("renders\n  got  %q\n  want %q",
					result.Diagnostics[0].Message.Description, want)
			}
		})
	}
}
