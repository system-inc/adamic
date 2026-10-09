package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// jsxNoConstructedContextValuesFile is where the fixtures pretend to live.
//
// A .tsx extension because every case holds JSX. The rule has no suffix gate, which the suffix
// cases below pin by writing the same reporting source to three extensions. There is no .ts row
// because a .ts file cannot hold JSX at all, which is a different zero from a rule declining to
// look, and self_closing_comp_test.go in this package records the same distinction.
const jsxNoConstructedContextValuesFile = "/repository/source/Providers.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the two tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/jsx-no-constructed-context-values.js, read
// by evaluating the two arrays in the tester and emitting Go raw strings from the decoded values,
// so no escape sequence was typed on the way here. Upstream carries 18 valid and 23 invalid cases,
// each invalid one naming exactly one messageId, which is 23 findings from 23 inputs.
//
// The rule declares the type checker, so every case runs through RunTyped. That harness trims each
// fixture to `strings.TrimSpace(source) + "\n"` before writing it, which is why the span
// assertions further down slice the transformed text rather than the Go literal.
//
// Upstream's `defaultMsgFunc` is reachable and this corpus never produces it: no failing case
// passes an inline function to a provider. The measured cases below cover it.

// TestJsxNoConstructedContextValuesFires runs the twenty-three failing cases from upstream.
func TestJsxNoConstructedContextValuesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"invalid 0 an object held in a const", `function Component() { const foo = {}; return (<Context.Provider value={foo}></Context.Provider>) }`, []string{"withIdentifierMsg"}},
		{"invalid 1 an array held in a const", `function Component() { const foo = []; return (<Context.Provider value={foo}></Context.Provider>) }`, []string{"withIdentifierMsg"}},
		{"invalid 2 an arrow held in a const", `function Component() { const foo = () => {}; return (<Context.Provider value={foo}></Context.Provider>)}`, []string{"withIdentifierMsgFunc"}},
		{"invalid 3 a named function expression held in a const", `function Component() { const foo = function bar(){}; return (<Context.Provider value={foo}></Context.Provider>)}`, []string{"withIdentifierMsgFunc"}},
		{"invalid 4 a class expression held in a const", `function Component() { const foo = class SomeClass{}; return (<Context.Provider value={foo}></Context.Provider>)}`, []string{"withIdentifierMsg"}},
		{"invalid 5 a new expression held in a const", `function Component() { const foo = new SomeClass(); return (<Context.Provider value={foo}></Context.Provider>)}`, []string{"withIdentifierMsg"}},
		{"invalid 6 a function declaration", `function Component() { function foo() {}; return (<Context.Provider value={foo}></Context.Provider>)}`, []string{"withIdentifierMsgFunc"}},
		{"invalid 7 a conditional whose consequent constructs", `function Component() { const foo = true ? {} : "fine"; return (<Context.Provider value={foo}></Context.Provider>)}`, []string{"withIdentifierMsg"}},
		{"invalid 8 a logical or whose right side constructs", `function Component() { const foo = bar || {}; return (<Context.Provider value={foo}></Context.Provider>)}`, []string{"withIdentifierMsg"}},
		{"invalid 9 a logical and whose right side constructs", `function Component() { const foo = bar && {}; return (<Context.Provider value={foo}></Context.Provider>)}`, []string{"withIdentifierMsg"}},
		{"invalid 10 a nested conditional", `function Component() { const foo = bar ? baz ? {} : null : null; return (<Context.Provider value={foo}></Context.Provider>)}`, []string{"withIdentifierMsg"}},
		{"invalid 11 an object held in a let", `function Component() { let foo = {}; return (<Context.Provider value={foo}></Context.Provider>) }`, []string{"withIdentifierMsg"}},
		{"invalid 12 an object held in a var", `function Component() { var foo = {}; return (<Context.Provider value={foo}></Context.Provider>)}`, []string{"withIdentifierMsg"}},
		{"invalid 13 a let that is reassigned to a number", `
        function Component() {
          let a = {};
          a = 10;
          return (<Context.Provider value={a}></Context.Provider>);
        }
      `, []string{"withIdentifierMsg"}},
		{"invalid 14 a const aliasing another const", `
        function Component() {
          const foo = {};
          const bar = foo;
          return (<Context.Provider value={bar}></Context.Provider>);
        }
      `, []string{"withIdentifierMsg"}},
		{"invalid 15 a conditional over a parameter and an object", `
        function Component(foo) {
          let bar = true ? foo : {};
          return (<Context.Provider value={bar}></Context.Provider>);
        }
      `, []string{"withIdentifierMsg"}},
		{"invalid 16 an object written inline", `function Component() { return (<Context.Provider value={{foo: "bar"}}></Context.Provider>);}`, []string{"defaultMsg"}},
		{"invalid 17 a jsx element held in a const", `function Component() { const Wrapper = (<SomeComp />); return (<Context.Provider value={Wrapper}></Context.Provider>);}`, []string{"withIdentifierMsg"}},
		{"invalid 18 a regular expression held in a const", `function Component() { const someRegex = /HelloWorld/; return (<Context.Provider value={someRegex}></Context.Provider>);}`, []string{"withIdentifierMsg"}},
		{"invalid 19 a chained assignment to an arrow", `
        function Component() {
          let foo = null;
          let bar = x = () => {};
          return (<Context.Provider value={bar}></Context.Provider>);
        }
      `, []string{"withIdentifierMsg"}},
		{"invalid 20 a function declaration under a namespaced context", `
        import React from 'react';

        const Context = React.createContext();
        function Component() {
          function foo() {};
          return (<Context value={foo}></Context>)
        }
      `, []string{"withIdentifierMsgFunc"}},
		{"invalid 21 an object held in a const under a bare context", `
        const MyContext = createContext();
        function Component() { const foo = {}; return (<MyContext value={foo}></MyContext>) }
      `, []string{"withIdentifierMsg"}},
		{"invalid 22 an object written inline under a bare context", `
        const MyContext = createContext();
        function Component() { return (<MyContext value={{foo: "bar"}}></MyContext>); }
      `, []string{"defaultMsg"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestJsxNoConstructedContextValuesStaysSilent runs the eighteen passing cases from upstream.
//
// Two of these carry upstream's own comment saying why they pass, and both are gaps rather than
// judgments: the rule does not look through a JSX spread attribute, and it cannot statically check
// a default parameter value. Reproduced as silence rather than improved on.
func TestJsxNoConstructedContextValuesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"valid 0 a props identifier", `const Component = () => <Context.Provider value={props}></Context.Provider>`},
		{"valid 1 a number literal", `const Component = () => <Context.Provider value={100}></Context.Provider>`},
		{"valid 2 a string attribute", `const Component = () => <Context.Provider value="Some string"></Context.Provider>`},
		{"valid 3 a memoized object", `function Component() { const foo = useMemo(() => { return {} }, []); return (<Context.Provider value={foo}></Context.Provider>)}`},
		{"valid 4 a destructured prop", `
        function Component({oneProp, twoProp, redProp, blueProp,}) {
          return (
            <NewContext.Provider value={twoProp}></NewContext.Provider>
          );
        }
      `},
		{"valid 5 an optional chain off a parameter", `
        function Foo(section) {
          const foo = section.section_components?.edges;

          return (
            <Context.Provider value={foo}></Context.Provider>
          )
        }
      `},
		{"valid 6 an imported binding member", `
        import foo from 'foo';
        function innerContext() {
          return (
            <Context.Provider value={foo.something}></Context.Provider>
          )
        }
      `},
		{"valid 7 a spread attribute rather than a value prop", `
        // Passes because the lint rule doesn't handle JSX spread attributes
        function innerContext() {
          const foo = {value: 'something'}
          return (
            <Context.Provider {...foo}></Context.Provider>
          )
        }
      `},
		{"valid 8 a spread attribute over a memoized value", `
        // Passes because the lint rule doesn't handle JSX spread attributes
        function innerContext() {
          const foo = useMemo(() => {
            return bar;
          })
          return (
            <Context.Provider value={foo}></Context.Provider>
          )
        }
      `},
		{"valid 9 a default parameter value", `
        // Passes because we can't statically check if it's using the default value
        function Component({ a = {} }) {
          return (<Context.Provider value={a}></Context.Provider>);
        }
      `},
		{"valid 10 a module level string", `
          import React from 'react';
          import MyContext from './MyContext';

          const value = '';

          function ContextProvider(props) {
              return (
                  <MyContext.Provider value={value as any}>
                      {props.children}
                  </MyContext.Provider>
              )
          }
        `},
		{"valid 11 a boolean literal", `
        import React from 'react';
        import BooleanContext from './BooleanContext';

        function ContextProvider(props) {
            return (
                <BooleanContext.Provider value>
                    {props.children}
                </BooleanContext.Provider>
            )
        }
      `},
		{"valid 12 a provider outside any component", `
        const root = ReactDOM.createRoot(document.getElementById('root'));
        root.render(
          <AppContext.Provider value={{}}>
            <AppView />
          </AppContext.Provider>
        );
      `},
		{"valid 13 a Consumer rather than a Provider", `
        // Passes because the context is not a provider
        function Component() {
          return <MyContext.Consumer value={{ foo: 'bar' }} />;
        }
      `},
		{"valid 14 a bare context with a props value", `
        import React from 'react';

        const MyContext = React.createContext();
        const Component = () => <MyContext value={props}></MyContext>;
      `},
		{"valid 15 a bare context with a memoized value", `
        import React from 'react';

        const MyContext = React.createContext();
        const Component = () => <MyContext value={100}></MyContext>;
      `},
		{"valid 16 a bare context with a string attribute", `
        const SomeContext = createContext();
        const Component = () => <SomeContext value="Some string"></SomeContext>;
      `},
		{"valid 17 a context that is a destructured prop", `
        // Passes because MyContext is not a variable declarator
        function Component({ MyContext }) {
          return <MyContext value={{ foo: "bar" }} />;
        }
      `},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestJsxNoConstructedContextValuesRequiresTheTypedHarness pins that the rule declares the checker.
//
// Every listener guards on a nil checker, so the plain harness makes the rule completely silent
// rather than crashing. That is the more dangerous of the two failure modes: a StaysSilent case
// would pass vacuously under a later revert of the declaration and nothing else here would notice.
func TestJsxNoConstructedContextValuesRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "const MyContext = createContext();\n" +
		"function Component() { return <MyContext value={{a: 1}} />; }\n"

	typed := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
		jsxNoConstructedContextValuesFile, source)
	rule_testing.ExpectFindings(t, typed, "defaultMsg")

	untyped := rule_testing.Run(t, JsxNoConstructedContextValues,
		jsxNoConstructedContextValuesFile, source)
	rule_testing.ExpectClean(t, untyped)
}

// TestJsxNoConstructedContextValuesHasNoFileSuffixGate pins the absence of a gate three rules in
// this package used to carry, as far as the typed harness can reach.
//
// Those three were oxc residue and were removed; upstream gates nothing. Measured on the installed
// build under .tsx, .jsx and .js: all three report.
//
// Only the .tsx row is expressible here, and the reason is the harness rather than the rule.
// `rule_testing.RunTyped` writes a tsconfig whose include list is `["**/*.ts", "**/*.tsx"]`
// (internal/rule_testing/program.go:30), so a fixture named .jsx or .js is not in the program at
// all and the run fails with TS18003 before any rule is offered a file. That is a fact about the
// typed harness, not about this rule, and weakening the rule to make a .js fixture pass would turn
// it into a fact about the rule.
//
// There is deliberately no .ts row either, and for a different reason: a .ts file cannot hold JSX,
// so the parser produces no element and the silence would say nothing. `self_closing_comp_test.go`
// in this package records that same distinction for a rule that can express it.
//
// What is left is one positive row confirming the rule does not decline a file for its name. The
// absence of a gate is pinned instead by there being no filename test anywhere in the rule, which a
// mutation over the classifier cannot fake.
func TestJsxNoConstructedContextValuesHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	source := "function Component() { return <Ctx.Provider value={{a: 1}} />; }\n"
	result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
		"/repository/source/Suffix.tsx", source)
	rule_testing.ExpectFindings(t, result, "defaultMsg")
}

// TestJsxNoConstructedContextValuesProviderShapes pins which tag names count, and the two paths are
// not symmetric.
//
// A qualified name only has to END in `Provider`; the object is never examined, so an arbitrary
// receiver qualifies. A bare identifier has to resolve to a `createContext` call. The corpus writes
// `Context.Provider` almost throughout and never tests an arbitrary receiver, so every row here was
// measured against the installed build.
func TestJsxNoConstructedContextValuesProviderShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		// The control. If this stops reporting the rest of the table measures nothing.
		{
			"a qualified Provider",
			"function Component() { return <Ctx.Provider value={{a: 1}} />; }\n",
			[]string{"defaultMsg"},
		},
		{
			"a qualified Provider on an arbitrary receiver",
			"function Component() { return <Whatever.Provider value={{a: 1}} />; }\n",
			[]string{"defaultMsg"},
		},
		{
			"a qualified member that is not Provider",
			"function Component() { return <Whatever.NotProvider value={{a: 1}} />; }\n",
			nil,
		},
		{
			"a bare identifier initialized by React.createContext",
			"declare const React: any;\nconst Ctx = React.createContext();\n" +
				"function Component() { return <Ctx value={{a: 1}} />; }\n",
			[]string{"defaultMsg"},
		},
		{
			"a bare identifier initialized by a bare createContext",
			"declare function createContext(): any;\nconst Ctx = createContext();\n" +
				"function Component() { return <Ctx value={{a: 1}} />; }\n",
			[]string{"defaultMsg"},
		},
		{
			"a bare identifier initialized by something else",
			"declare function somethingElse(): any;\nconst Ctx = somethingElse();\n" +
				"function Component() { return <Ctx value={{a: 1}} />; }\n",
			nil,
		},
		{
			"a bare identifier initialized by a foreign createContext",
			"declare const Other: any;\nconst Ctx = Other.createContext();\n" +
				"function Component() { return <Ctx value={{a: 1}} />; }\n",
			nil,
		},
		{
			"a bare identifier with no initializer at all",
			"declare const Ctx: any;\nfunction Component() { return <Ctx value={{a: 1}} />; }\n",
			nil,
		},
		{
			"an intrinsic element",
			"function Component() { return <div value={{a: 1}} />; }\n",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestJsxNoConstructedContextValuesValueAttributeShapes pins the three ways a value prop declines.
//
// All three measured silent on the installed build, and none is in the corpus as a distinct case.
func TestJsxNoConstructedContextValuesValueAttributeShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"a value expression container",
			"function Component() { return <Ctx.Provider value={{a: 1}} />; }\n",
			[]string{"defaultMsg"},
		},
		{"no value prop", "function Component() { return <Ctx.Provider />; }\n", nil},
		{"a boolean shorthand value", "function Component() { return <Ctx.Provider value />; }\n", nil},
		{"a string literal value", "function Component() { return <Ctx.Provider value=\"x\" />; }\n", nil},
		{
			"another prop constructing while value is fine",
			"function Component() { return <Ctx.Provider other={{a: 1}} value={props.v} />; }\n",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestJsxNoConstructedContextValuesConstructionKinds pins every arm of upstream's isConstruction.
//
// The corpus reaches most of these only through a const binding, so the inline forms below are what
// separate the arms from the identifier-following path. Every verdict measured against the
// installed build, including the four silent rows, which are the false positives a wider switch
// would ship.
func TestJsxNoConstructedContextValuesConstructionKinds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		value   string
		wantIds []string
	}{
		{"an object", "{a: 1}", []string{"defaultMsg"}},
		{"an array", "[1]", []string{"defaultMsg"}},
		{"an arrow", "() => {}", []string{"defaultMsgFunc"}},
		{"a function expression", "function () {}", []string{"defaultMsgFunc"}},
		{"a class expression", "class {}", []string{"defaultMsg"}},
		{"a new expression", "new Foo()", []string{"defaultMsg"}},
		{"a regular expression", "/x/", []string{"defaultMsg"}},
		{"a jsx element", "<div />", []string{"defaultMsg"}},
		{"a jsx fragment", "<></>", []string{"defaultMsg"}},
		{"a spread object", "{...rest}", []string{"defaultMsg"}},

		// The recursing arms. The left side is tried first, which the next two rows pin together.
		{"a conditional constructing on both sides", "c ? {a: 1} : {b: 2}", []string{"defaultMsg"}},
		{"a conditional constructing on the left", "c ? {a: 1} : x", []string{"defaultMsg"}},
		{"a conditional constructing on the right", "c ? x : {b: 2}", []string{"defaultMsg"}},
		{"a conditional constructing on neither", "c ? x : y", nil},
		{"a logical or constructing on the right", "x || {a: 1}", []string{"defaultMsg"}},
		{"a logical and constructing on the left", "{a: 1} && x", []string{"defaultMsg"}},
		{"a nullish coalescing constructing", "x ?? {a: 1}", []string{"defaultMsg"}},

		// A type assertion is transparent.
		{"an as expression over an object", "{a: 1} as any", []string{"defaultMsg"}},

		// Parentheses, which upstream's parser folds away and ours keeps. Nothing in either corpus
		// writes one in this position, so a port without the unwrap goes silent here and no
		// imported fixture could see it.
		{"a parenthesized object", "({a: 1})", []string{"defaultMsg"}},
		{"a doubly parenthesized object", "(({a: 1}))", []string{"defaultMsg"}},
		{"a parenthesized plain identifier", "(x)", nil},

		// The silent rows. A call result is the important one: this rule cannot judge what a
		// function returns, and a wider switch would report every `useMemo` call in the tree.
		{"a call result", "makeIt()", nil},
		{"a string", "\"s\"", nil},
		{"a number", "1", nil},
		{"a template literal", "`t`", nil},
		{"a plain identifier not in scope", "notDeclaredAnywhere", nil},
		{"an arithmetic binary expression", "a + b", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "declare const c: any;\ndeclare const x: any;\ndeclare const a: any;\n" +
				"declare const b: any;\ndeclare const rest: any;\ndeclare const Foo: any;\n" +
				"declare function makeIt(): any;\n" +
				"function Component() { return <Ctx.Provider value={" + testCase.value + "} />; }\n"
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestJsxNoConstructedContextValuesIdentifierFollowing pins the indirection path.
//
// Upstream resolves the identifier to its binding, takes the LAST definition, and bails out unless
// it is a variable or a function name. A parameter is what that bail-out is for, and it is the row
// that separates this from a rule that reports every identifier.
func TestJsxNoConstructedContextValuesIdentifierFollowing(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		body    string
		wantIds []string
	}{
		{"a const holding an object", "const v = {a: 1};", []string{"withIdentifierMsg"}},
		{"a const holding a number", "const v = 1;", nil},
		{"a const holding a call result", "const v = makeIt();", nil},
		{"a function declaration", "function v() {}", []string{"withIdentifierMsgFunc"}},
		{"a const holding an arrow", "const v = () => {};", []string{"withIdentifierMsgFunc"}},
		{"a const aliasing another const", "const inner = {a: 1}; const v = inner;", []string{"withIdentifierMsg"}},
		{"a const with no initializer", "let v;", nil},
		{"a member access off a constructed object", "const o = {a: {b: 1}}; const v = o.a;", []string{"withIdentifierMsg"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "declare function makeIt(): any;\n" +
				"function Component() {\n  " + testCase.body +
				"\n  return <Ctx.Provider value={v} />;\n}\n"
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestJsxNoConstructedContextValuesParameterIsSilent pins upstream's bail-out, in its own test
// because the source shape differs from the table above.
//
// A parameter is not a variable or a function name, so upstream declines it. Without that decline
// this rule would report every provider handed a prop, which is the single largest false-positive
// class available to it.
func TestJsxNoConstructedContextValuesParameterIsSilent(t *testing.T) {
	t.Parallel()

	source := "function Component(v: any) { return <Ctx.Provider value={v} />; }\n"
	result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
		jsxNoConstructedContextValuesFile, source)
	rule_testing.ExpectClean(t, result)
}

// TestJsxNoConstructedContextValuesEnclosingComponent pins the classifier, row for row.
//
// Upstream calls its 959-line component detector here. That detector is not in this tree and is not
// needed for this rule, because the node being classified is already a JSX element, so every
// question the detector would ask about whether a function returns JSX is settled before it is
// asked. What remains is the name and the class heritage.
//
// All twenty rows were measured against the installed build BEFORE the classifier was written, in a
// throwaway probe package that compared a candidate implementation against them and was deleted.
// Two rows are where the shelf disagrees and this rule cannot delegate: `scope.NameOf` prefers a
// function expression's own name over its binding, and it answers a method's name where the class
// is what decides.
func TestJsxNoConstructedContextValuesEnclosingComponent(t *testing.T) {
	t.Parallel()

	const provider = "<Ctx.Provider value={{a: 1}} />"
	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a capitalized function declaration", "function Component() { return " + provider + "; }\n", []string{"defaultMsg"}},
		{"a lowercase function declaration", "function component() { return " + provider + "; }\n", nil},

		// The walk stops at the FIRST function, so an inner lowercase helper hides a provider that
		// an outer component would otherwise expose, and an inner component exposes one an outer
		// helper would hide. Both measured.
		{
			"a lowercase helper inside a component",
			"function Component() { function helper() { return " + provider + "; } return helper(); }\n",
			nil,
		},
		{
			"a component inside a lowercase helper",
			"function helper() { function Component() { return " + provider + "; } return Component(); }\n",
			[]string{"defaultMsg"},
		},
		{
			"an anonymous arrow inside a component",
			"function Component() { const inner = () => " + provider + "; return inner(); }\n",
			nil,
		},

		// An anonymous function takes its binding's name.
		{"an arrow bound to a capitalized name", "const Component = () => " + provider + ";\n", []string{"defaultMsg"}},
		{"an arrow bound to a lowercase name", "const component = () => " + provider + ";\n", nil},

		// And the BINDING wins over the function expression's own name, in both directions. This is
		// the pair the shelf helper gets backwards.
		{
			"a capitalized function expression bound to a lowercase name",
			"const thing = function Component() { return " + provider + "; };\n",
			nil,
		},
		{
			"a lowercase function expression bound to a capitalized name",
			"const Component = function thing() { return " + provider + "; };\n",
			[]string{"defaultMsg"},
		},

		// Inside a class the heritage decides, not the member name and not the class name.
		{
			"a render method on a component class",
			"declare const React: any;\nclass C extends React.Component { render() { return " + provider + "; } }\n",
			[]string{"defaultMsg"},
		},
		{
			"any other method on a component class",
			"declare const React: any;\nclass C extends React.Component { other() { return " + provider + "; } }\n",
			[]string{"defaultMsg"},
		},
		{
			"a render method on a class with no heritage",
			"class C { render() { return " + provider + "; } }\n",
			nil,
		},
		{
			"a class extending something that is not a React base",
			"declare const Foo: any;\nclass C extends Foo.Bar { render() { return " + provider + "; } }\n",
			nil,
		},
		{
			"a lowercase class extending a React base",
			"declare const React: any;\nclass c extends React.Component { render() { return " + provider + "; } }\n",
			[]string{"defaultMsg"},
		},
		{
			"a class field holding an arrow",
			"declare const React: any;\nclass C extends React.Component { f = () => " + provider + "; }\n",
			[]string{"defaultMsg"},
		},
		{
			"a capitalized method on a plain class",
			"class Thing { Render() { return " + provider + "; } }\n",
			nil,
		},

		// No enclosing function at all.
		{"a provider at the top level", "const x = " + provider + ";\n", nil},

		// A call wrapper is transparent, so the binding one level out is what is read.
		{
			"an arrow wrapped in memo",
			"declare const React: any;\nconst Component = React.memo(() => " + provider + ");\n",
			[]string{"defaultMsg"},
		},
		{
			"an arrow wrapped in forwardRef",
			"declare const React: any;\nconst Component = React.forwardRef(() => " + provider + ");\n",
			[]string{"defaultMsg"},
		},
		{
			"an arrow wrapped in memo bound to a lowercase name",
			"declare const React: any;\nconst component = React.memo(() => " + provider + ");\n",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestJsxNoConstructedContextValuesSpans asserts where each finding points, and it is not where a
// reader would first guess.
//
// Every finding is anchored on the CONSTRUCTION rather than on the value expression, so an indirect
// case reports somewhere else entirely: a const declared two lines above the provider reports on
// its initializer, not on the identifier in the value prop. Measured by slicing the reported range
// on the installed build across twenty-six inputs. `ExpectFindings` cannot see any of this and the
// corpus asserts none of it.
//
// `RunTyped` writes `strings.TrimSpace(source) + "\n"` to disk, so the expectation is sliced from
// the same transform rather than from the Go literal, which would be one byte off.
func TestJsxNoConstructedContextValuesSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{
			"an inline object underlines the object",
			"function Component() { return <Ctx.Provider value={{a: 1}} />; }\n",
			"{a: 1}",
		},
		{
			"an inline arrow underlines the arrow",
			"function Component() { return <Ctx.Provider value={() => {}} />; }\n",
			"() => {}",
		},
		{
			"a held object underlines the initializer, not the identifier",
			"function Component() {\n  const v = {a: 1};\n  return <Ctx.Provider value={v} />;\n}\n",
			"{a: 1}",
		},
		{
			"a function declaration underlines the whole declaration",
			"function Component() {\n  function v() {}\n  return <Ctx.Provider value={v} />;\n}\n",
			"function v() {}",
		},
		{
			"an aliased const underlines the original construction",
			"function Component() {\n  const inner = {x: 1};\n  const v = inner;\n" +
				"  return <Ctx.Provider value={v} />;\n}\n",
			"{x: 1}",
		},
		{
			"a conditional underlines the branch that constructs",
			"declare const c: any;\ndeclare const x: any;\n" +
				"function Component() { return <Ctx.Provider value={c ? x : {b: 2}} />; }\n",
			"{b: 2}",
		},
		{
			// When BOTH branches construct, the left one wins. The message id is identical either
			// way, so only the span records which side was tried first: a mutant swapping the two
			// recursions survived every id fixture in this file. Measured on the installed build,
			// which underlines `{a: 1}`.
			"a conditional constructing on both sides underlines the left one",
			"declare const c: any;\n" +
				"function Component() { return <Ctx.Provider value={c ? {a: 1} : {b: 2}} />; }\n",
			"{a: 1}",
		},
		{
			// The same question for a logical expression, whose arm recurses in the same order.
			"a logical or constructing on both sides underlines the left one",
			"function Component() { return <Ctx.Provider value={{a: 1} || {b: 2}} />; }\n",
			"{a: 1}",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted exactly one finding, got %d", len(result.Diagnostics))
			}
			// The harness trims, so slice what it actually wrote.
			written := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]
			got := written[diagnostic.Range.Pos():diagnostic.Range.End()]
			if got != testCase.wantText {
				t.Errorf("span text = %q, wanted %q", got, testCase.wantText)
			}
		})
	}
}

// TestJsxNoConstructedContextValuesMessageText asserts the rendered message exactly.
//
// This rule builds its message with Sprintf, so the id assertion cannot see anything the format
// string does. Two line numbers and a variable name are interpolated, and a message id fixture is
// satisfied by all of them being wrong. The line numbers are the part most likely to be off, since
// the position helper is zero-based and upstream renders one-based.
//
// Asserted by equality against a literal typed here rather than against the rule's own constants,
// which would move on both sides under mutation and could not fail.
func TestJsxNoConstructedContextValuesMessageText(t *testing.T) {
	t.Parallel()

	t.Run("the default message names the construction line", func(t *testing.T) {
		t.Parallel()
		source := "function Component() {\n  return <Ctx.Provider value={{a: 1}} />;\n}\n"
		result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
			jsxNoConstructedContextValuesFile, source)
		rule_testing.ExpectFindings(t, result, "defaultMsg")
		want := "The object passed as the value prop to the Context provider (at line 2) changes " +
			"every render, so every consumer of this context re-renders on every render of this " +
			"component even when nothing it reads has changed. Wrap it in a useMemo hook so the " +
			"identity is stable between renders."
		if got := result.Diagnostics[0].Message.Description; got != want {
			t.Errorf("message =\n%q\nwanted\n%q", got, want)
		}
	})

	t.Run("the identifier message names both lines and the variable", func(t *testing.T) {
		t.Parallel()
		source := "function Component() {\n  const v = {a: 1};\n  return <Ctx.Provider value={v} />;\n}\n"
		result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
			jsxNoConstructedContextValuesFile, source)
		rule_testing.ExpectFindings(t, result, "withIdentifierMsg")
		want := "The \"v\" object (at line 2) passed as the value prop to the Context provider " +
			"(at line 3) changes every render, so every consumer of this context re-renders on " +
			"every render of this component even when nothing it reads has changed. Wrap it in a " +
			"useMemo hook so the identity is stable between renders."
		if got := result.Diagnostics[0].Message.Description; got != want {
			t.Errorf("message =\n%q\nwanted\n%q", got, want)
		}
	})

	t.Run("a function construction asks for useCallback", func(t *testing.T) {
		t.Parallel()
		source := "function Component() {\n  function v() {}\n  return <Ctx.Provider value={v} />;\n}\n"
		result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
			jsxNoConstructedContextValuesFile, source)
		rule_testing.ExpectFindings(t, result, "withIdentifierMsgFunc")
		want := "The \"v\" function declaration (at line 2) passed as the value prop to the " +
			"Context provider (at line 3) changes every render, so every consumer of this " +
			"context re-renders on every render of this component even when nothing it reads has " +
			"changed. Wrap it in a useCallback hook so the identity is stable between renders."
		if got := result.Diagnostics[0].Message.Description; got != want {
			t.Errorf("message =\n%q\nwanted\n%q", got, want)
		}
	})

	t.Run("an inline function asks for useCallback with no variable name", func(t *testing.T) {
		t.Parallel()
		// This is the fourth message id, which upstream's own corpus never exercises.
		source := "function Component() {\n  return <Ctx.Provider value={() => {}} />;\n}\n"
		result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
			jsxNoConstructedContextValuesFile, source)
		rule_testing.ExpectFindings(t, result, "defaultMsgFunc")
		want := "The function expression passed as the value prop to the Context provider (at " +
			"line 2) changes every render, so every consumer of this context re-renders on every " +
			"render of this component even when nothing it reads has changed. Wrap it in a " +
			"useCallback hook so the identity is stable between renders."
		if got := result.Diagnostics[0].Message.Description; got != want {
			t.Errorf("message =\n%q\nwanted\n%q", got, want)
		}
	})
}

// TestJsxNoConstructedContextValuesLineNumbersSkipTrivia pins the fix for a defect the span
// assertions structurally could not see.
//
// A node's `Pos()` includes its leading trivia, so a construction preceded by a comment or a blank
// line reports the line the TRIVIA starts on. `ctx.ReportNode` already routes through
// `rule.TokenRange`, so the finding pointed at the right place the whole time and every span
// fixture stayed green; only the line number interpolated into the message was wrong, which no
// message-id assertion and no span assertion can reach.
//
// Measured against the installed build: upstream renders line 4 for the construction here, and the
// port rendered line 2 before the line numbers were routed through TokenRange as well.
func TestJsxNoConstructedContextValuesLineNumbersSkipTrivia(t *testing.T) {
	t.Parallel()

	source := "function Component() {\n" +
		"  // a comment that occupies the line above the construction\n" +
		"\n" +
		"  const v = {a: 1};\n" +
		"  return <Ctx.Provider value={v} />;\n" +
		"}\n"
	result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
		jsxNoConstructedContextValuesFile, source)
	rule_testing.ExpectFindings(t, result, "withIdentifierMsg")

	want := "The \"v\" object (at line 4) passed as the value prop to the Context provider (at " +
		"line 5) changes every render, so every consumer of this context re-renders on every " +
		"render of this component even when nothing it reads has changed. Wrap it in a useMemo " +
		"hook so the identity is stable between renders."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Errorf("message =\n%q\nwanted\n%q", got, want)
	}
}

// TestJsxNoConstructedContextValuesTakesTheLastDeclaration pins which declaration decides when a
// name has more than one.
//
// Upstream reads `defs[defs.length - 1]`, the LAST definition rather than the first. Nothing in the
// corpus redeclares a name, so a mutation swapping the index survived the entire imported set. The
// first two rows below have opposite verdicts from declaration order alone, which is the cheapest
// possible distinguishing pair, and every row was measured against the installed build before it
// was written here.
//
// Whether our checker orders `symbol.Declarations` the same way as upstream's scope analysis is not
// assumed: these fixtures are the measurement. A `var` is used because it is the only binding form
// that legally redeclares.
func TestJsxNoConstructedContextValuesTakesTheLastDeclaration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		body    string
		wantIds []string
	}{
		// The control: one declaration, so ordering cannot be what decides.
		{"a single object declaration", "  var v = {a: 1};", []string{"withIdentifierMsg"}},

		// The pair. Same two declarations, opposite order, opposite verdict.
		{"an object then a number", "  var v = {a: 1};\n  var v = 1;", nil},
		{"a number then an object", "  var v = 1;\n  var v = {a: 1};", []string{"withIdentifierMsg"}},

		// The kind is read off the last declaration too, not just whether one exists.
		{"an object then an arrow", "  var v = {a: 1};\n  var v = () => {};", []string{"withIdentifierMsgFunc"}},
		{"an object then a function declaration", "  var v = {a: 1};\n  function v() {}", []string{"withIdentifierMsgFunc"}},

		// A measured divergence, and it is in the CHECKER rather than in the index.
		//
		// Upstream reports "object" here, taking the var as its last definition. Our checker
		// returns a declaration list of length ONE for a var-and-function pair, holding only the
		// function declaration, whichever order they are written in. Probed directly: both
		// orderings answer `count=1 [0=KindFunctionDeclaration]`, while two vars answer `count=2`
		// ordered by source position, which is the ordering upstream's scope analysis also
		// produces and which the pair above confirms.
		//
		// So there is no index choice that reproduces upstream on this shape, because the
		// declaration upstream would pick is not in the list at all. Recorded rather than worked
		// around: reaching for the function's own symbol to recover the var would diverge on the
		// far more common single-declaration case, and this shape requires redeclaring a name as
		// both a var and a function in one scope.
		{"a function declaration then an object", "  function v() {}\n  var v = {a: 1};", []string{"withIdentifierMsgFunc"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "function Component() {\n" + testCase.body +
				"\n  return <Ctx.Provider value={v} />;\n}\n"
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestJsxNoConstructedContextValuesMemberUsageIsTheObject pins which node a member access records
// as its usage.
//
// Upstream sets the usage to the member expression's OBJECT rather than to the whole member
// expression. Both choices produce the same message id and the same span, because the span comes
// from the construction and not from the usage, so a mutant swapping them survived every other
// fixture in this file. The usage is only observable through the message, and only through its
// usage LINE, since a member expression carries no name for the other slot.
//
// So the distinguishing input has to split the member access across two lines. Measured against the
// installed build: it renders line 4, the line the object sits on, not line 5.
func TestJsxNoConstructedContextValuesMemberUsageIsTheObject(t *testing.T) {
	t.Parallel()

	source := "function Component() {\n" +
		"  const o = {a: {b: 1}};\n" +
		"  return <Ctx.Provider value={\n" +
		"    o\n" +
		"    .a} />;\n" +
		"}\n"
	result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
		jsxNoConstructedContextValuesFile, source)
	rule_testing.ExpectFindings(t, result, "withIdentifierMsg")

	want := "The \"o\" object (at line 2) passed as the value prop to the Context provider (at " +
		"line 4) changes every render, so every consumer of this context re-renders on every " +
		"render of this component even when nothing it reads has changed. Wrap it in a useMemo " +
		"hook so the identity is stable between renders."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Errorf("message =\n%q\nwanted\n%q", got, want)
	}
}
