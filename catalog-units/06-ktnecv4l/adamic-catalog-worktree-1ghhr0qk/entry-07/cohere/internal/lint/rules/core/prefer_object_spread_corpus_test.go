package core

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * ESLint 10.8.1's prefer-object-spread, run through the installed rule under @typescript-eslint/parser over
 * its whole test corpus and over the edges this port's own reading raised (#jjfa7qb), with every
 * finding written here where ESLint placed it and the file as ESLint's fix leaves it.
 *
 * The corpus rows are the registry's corpus file, verbatim, less the rows ESLint runs under configured
 * globals. The edge rows were written for the move onto the shelf's ReferenceTracker: assign reached
 * through a destructured, renamed or copied Object, globalThis, an optional call either way, a
 * parenthesized callee, a parameter named Object, and a file that writes Object.
 *
 * One edge is left out because the rule departs from ESLint there, as the tracker does for every rule
 * on it: `window.Object.assign({}, a)`, which ESLint reports only under browser globals, and which the
 * tracker follows, since an undeclared `window` is the runtime's global object here.
 *
 * Each finding is its text and byte offset. The fixed column is the whole file after ESLint's fixes,
 * applied as its verifyAndFix applies them, or empty where ESLint offers none.
 */
func TestPreferObjectSpreadAgreesWithESLint(t *testing.T) {
	t.Parallel()

	cases := []struct {
		origin  string
		options string
		code    string
		want    []string
		fixed   string
	}{
		{"corpus", "", "Object.assign()", []string{}, ""},
		{"corpus", "", "let a = Object.assign(a, b)", []string{}, ""},
		{"corpus", "", "Object.assign(a, b)", []string{}, ""},
		{"corpus", "", "let a = Object.assign(b, { c: 1 })", []string{}, ""},
		{"corpus", "", "const bar = { ...foo }", []string{}, ""},
		{"corpus", "", "Object.assign(...foo)", []string{}, ""},
		{"corpus", "", "Object.assign(foo, { bar: baz })", []string{}, ""},
		{"corpus", "", "Object.assign({}, ...objects)", []string{}, ""},
		{"corpus", "", "foo({ foo: 'bar' })", []string{}, ""},
		{"corpus", "", "\n        const Object = {};\n        Object.assign({}, foo);\n        ", []string{}, ""},
		{"corpus", "", "\n        Object = {};\n        Object.assign({}, foo);\n        ", []string{}, ""},
		{"corpus", "", "\n        const Object = {};\n        Object.assign({ foo: 'bar' });\n        ", []string{}, ""},
		{"corpus", "", "\n        Object = {};\n        Object.assign({ foo: 'bar' });\n        ", []string{}, ""},
		{"corpus", "", "\n        const Object = require('foo');\n        Object.assign({ foo: 'bar' });\n        ", []string{}, ""},
		{"corpus", "", "\n        import Object from 'foo';\n        Object.assign({ foo: 'bar' });\n        ", []string{}, ""},
		{"corpus", "", "\n        import { Something as Object } from 'foo';\n        Object.assign({ foo: 'bar' });\n        ", []string{}, ""},
		{"corpus", "", "\n        import { Object, Array } from 'globals';\n        Object.assign({ foo: 'bar' });\n        ", []string{}, ""},
		{"corpus", "", "globalThis.Object.assign({}, foo)", []string{"globalThis.Object.assign({}, foo)@0"}, "({ ...foo})"},
		{"corpus", "", "globalThis.Object.assign({}, { foo: 'bar' })", []string{"globalThis.Object.assign({}, { foo: 'bar' })@0"}, "({ foo: 'bar'})"},
		{"corpus", "", "globalThis.Object.assign({}, baz, { foo: 'bar' })", []string{"globalThis.Object.assign({}, baz, { foo: 'bar' })@0"}, "({ ...baz, foo: 'bar'})"},
		{"corpus", "", "\n                var globalThis = foo;\n                globalThis.Object.assign({}, foo)\n                ", []string{}, ""},
		{"corpus", "", "class C { #assign; foo() { Object.#assign({}, foo); } }", []string{}, ""},
		{"corpus", "", "Object.assign({ get a() {} }, {})", []string{}, ""},
		{"corpus", "", "Object.assign({ set a(val) {} }, {})", []string{}, ""},
		{"corpus", "", "Object.assign({ get a() {} }, foo)", []string{}, ""},
		{"corpus", "", "Object.assign({ set a(val) {} }, foo)", []string{}, ""},
		{"corpus", "", "Object.assign({ foo: 'bar', get a() {}, baz: 'quux' }, quuux)", []string{}, ""},
		{"corpus", "", "Object.assign({ foo: 'bar', set a(val) {} }, { baz: 'quux' })", []string{}, ""},
		{"corpus", "", "Object.assign({}, { get a() {} })", []string{}, ""},
		{"corpus", "", "Object.assign({}, { set a(val) {} })", []string{}, ""},
		{"corpus", "", "Object.assign({}, { foo: 'bar', get a() {} }, {})", []string{}, ""},
		{"corpus", "", "Object.assign({ foo }, bar, {}, { baz: 'quux', set a(val) {}, quuux }, {})", []string{}, ""},
		{"corpus", "", "Object.assign({}, foo)", []string{"Object.assign({}, foo)@0"}, "({ ...foo})"},
		{"corpus", "", "Object.assign  ({}, foo)", []string{"Object.assign  ({}, foo)@0"}, "({ ...foo})"},
		{"corpus", "", "Object.assign({}, { foo: 'bar' })", []string{"Object.assign({}, { foo: 'bar' })@0"}, "({ foo: 'bar'})"},
		{"corpus", "", "Object.assign({}, baz, { foo: 'bar' })", []string{"Object.assign({}, baz, { foo: 'bar' })@0"}, "({ ...baz, foo: 'bar'})"},
		{"corpus", "", "Object.assign({}, { foo: 'bar', baz: 'foo' })", []string{"Object.assign({}, { foo: 'bar', baz: 'foo' })@0"}, "({ foo: 'bar', baz: 'foo'})"},
		{"corpus", "", "Object.assign({ foo: 'bar' }, baz)", []string{"Object.assign({ foo: 'bar' }, baz)@0"}, "({foo: 'bar', ...baz})"},
		{"corpus", "", "Object.assign({ foo: 'bar' }, cats, dogs, trees, birds)", []string{"Object.assign({ foo: 'bar' }, cats, dogs, trees, birds)@0"}, "({foo: 'bar', ...cats, ...dogs, ...trees, ...birds})"},
		{"corpus", "", "Object.assign({ foo: 'bar' }, Object.assign({ bar: 'foo' }, baz))", []string{"Object.assign({ foo: 'bar' }, Object.assign({ bar: 'foo' }, baz))@0", "Object.assign({ bar: 'foo' }, baz)@30"}, "({foo: 'bar', ...({bar: 'foo', ...baz})})"},
		{"corpus", "", "Object.assign({ foo: 'bar' }, Object.assign({ bar: 'foo' }, Object.assign({}, { superNested: 'butwhy' })))", []string{"Object.assign({ foo: 'bar' }, Object.assign({ bar: 'foo' }, Object.assign({}, { superNested: 'butwhy' })))@0", "Object.assign({ bar: 'foo' }, Object.assign({}, { superNested: 'butwhy' }))@30", "Object.assign({}, { superNested: 'butwhy' })@60"}, "({foo: 'bar', ...({bar: 'foo', ...({ superNested: 'butwhy'})})})"},
		{"corpus", "", "Object.assign({foo: 'bar', ...bar}, baz)", []string{"Object.assign({foo: 'bar', ...bar}, baz)@0"}, "({foo: 'bar', ...bar, ...baz})"},
		{"corpus", "", "Object.assign({}, { foo, bar, baz })", []string{"Object.assign({}, { foo, bar, baz })@0"}, "({ foo, bar, baz})"},
		{"corpus", "", "Object.assign({}, { [bar]: 'foo' })", []string{"Object.assign({}, { [bar]: 'foo' })@0"}, "({ [bar]: 'foo'})"},
		{"corpus", "", "Object.assign({ ...bar }, { ...baz })", []string{"Object.assign({ ...bar }, { ...baz })@0"}, "({...bar, ...baz})"},
		{"corpus", "", "Object.assign({ ...bar }, {\n                // this is a bar\n                foo: 'bar',\n                baz: \"cats\"\n            })", []string{"Object.assign({ ...bar }, {\n                // this is a bar\n                foo: 'bar',\n                baz: \"cats\"\n            })@0"}, "({...bar, // this is a bar\n                foo: 'bar',\n                baz: \"cats\"})"},
		{"corpus", "", "Object.assign({\n                boo: \"lol\",\n                // I'm a comment\n                dog: \"cat\"\n             }, {\n                // this is a bar\n                foo: 'bar',\n                baz: \"cats\"\n            })", []string{"Object.assign({\n                boo: \"lol\",\n                // I'm a comment\n                dog: \"cat\"\n             }, {\n                // this is a bar\n                foo: 'bar',\n                baz: \"cats\"\n            })@0"}, "({boo: \"lol\",\n                // I'm a comment\n                dog: \"cat\", // this is a bar\n                foo: 'bar',\n                baz: \"cats\"})"},
		{"corpus", "", "const test = Object.assign({ ...bar }, {\n                foo: 'bar', // inline comment\n                baz: \"cats\"\n            })", []string{"Object.assign({ ...bar }, {\n                foo: 'bar', // inline comment\n                baz: \"cats\"\n            })@13"}, "const test = {...bar, foo: 'bar', // inline comment\n                baz: \"cats\"}"},
		{"corpus", "", "const test = Object.assign({ ...bar }, {\n                /**\n                 * foo\n                 */\n                foo: 'bar',\n                baz: \"cats\"\n            })", []string{"Object.assign({ ...bar }, {\n                /**\n                 * foo\n                 */\n                foo: 'bar',\n                baz: \"cats\"\n            })@13"}, "const test = {...bar, /**\n                 * foo\n                 */\n                foo: 'bar',\n                baz: \"cats\"}"},
		{"corpus", "", "const result = doSomething()\nObject.assign({}, myData)", []string{"Object.assign({}, myData)@29"}, "const result = doSomething()\n;({ ...myData})"},
		{"corpus", "", "let a = foo + Object.assign({}, bar)", []string{"Object.assign({}, bar)@14"}, "let a = foo + ({ ...bar})"},
		{"corpus", "", "let foo = function() {};\nfoo\nObject.assign({}, bar)", []string{"Object.assign({}, bar)@29"}, "let foo = function() {};\nfoo\n;({ ...bar})"},
		{"corpus", "", "foo\nObject.assign({ foo: bar })", []string{"Object.assign({ foo: bar })@4"}, "foo\n;({foo: bar})"},
		{"corpus", "", "foo();\nObject.assign({}, bar)", []string{"Object.assign({}, bar)@7"}, "foo();\n({ ...bar})"},
		{"corpus", "", "const x = [1]\nObject.assign({}, bar)", []string{"Object.assign({}, bar)@14"}, "const x = [1]\n;({ ...bar})"},
		{"corpus", "", "foo\nObject.assign({}, bar).doSomething()", []string{"Object.assign({}, bar)@4"}, "foo\n;({ ...bar}).doSomething()"},
		{"corpus", "", "foo\nObject.assign({}, bar), 2", []string{"Object.assign({}, bar)@4"}, "foo\n;({ ...bar}), 2"},
		{"corpus", "", "Object.assign({})", []string{"Object.assign({})@0"}, "({})"},
		{"corpus", "", "Object.assign({ foo: bar })", []string{"Object.assign({ foo: bar })@0"}, "({foo: bar})"},
		{"corpus", "", "\n                const foo = 'bar';\n                Object.assign({ foo: bar })\n            ", []string{"Object.assign({ foo: bar })@52"}, "\n                const foo = 'bar';\n                ({foo: bar})\n            "},
		{"corpus", "", "\n                foo = 'bar';\n                Object.assign({ foo: bar })\n            ", []string{"Object.assign({ foo: bar })@46"}, "\n                foo = 'bar';\n                ({foo: bar})\n            "},
		{"corpus", "", "let a = Object.assign({})", []string{"Object.assign({})@8"}, "let a = {}"},
		{"corpus", "", "let a = Object.assign({}, a)", []string{"Object.assign({}, a)@8"}, "let a = { ...a}"},
		{"corpus", "", "let a = Object.assign   ({}, a)", []string{"Object.assign   ({}, a)@8"}, "let a = { ...a}"},
		{"corpus", "", "let a = Object.assign({ a: 1 }, b)", []string{"Object.assign({ a: 1 }, b)@8"}, "let a = {a: 1, ...b}"},
		{"corpus", "", "Object.assign(  {},  a,      b,   )", []string{"Object.assign(  {},  a,      b,   )@0"}, "({    ...a,      ...b,   })"},
		{"corpus", "", "Object.assign({}, a ? b : {}, b => c, a = 2)", []string{"Object.assign({}, a ? b : {}, b => c, a = 2)@0"}, "({ ...(a ? b : {}), ...(b => c), ...(a = 2)})"},
		{"corpus", "", "\n                const someVar = 'foo';\n                Object.assign({}, a ? b : {}, b => c, a = 2)\n            ", []string{"Object.assign({}, a ? b : {}, b => c, a = 2)@56"}, "\n                const someVar = 'foo';\n                ({ ...(a ? b : {}), ...(b => c), ...(a = 2)})\n            "},
		{"corpus", "", "\n                someVar = 'foo';\n                Object.assign({}, a ? b : {}, b => c, a = 2)\n            ", []string{"Object.assign({}, a ? b : {}, b => c, a = 2)@50"}, "\n                someVar = 'foo';\n                ({ ...(a ? b : {}), ...(b => c), ...(a = 2)})\n            "},
		{"corpus", "", "[1, 2, Object.assign({}, a)]", []string{"Object.assign({}, a)@7"}, "[1, 2, { ...a}]"},
		{"corpus", "", "const foo = Object.assign({}, a)", []string{"Object.assign({}, a)@12"}, "const foo = { ...a}"},
		{"corpus", "", "function foo() { return Object.assign({}, a) }", []string{"Object.assign({}, a)@24"}, "function foo() { return { ...a} }"},
		{"corpus", "", "foo(Object.assign({}, a));", []string{"Object.assign({}, a)@4"}, "foo({ ...a});"},
		{"corpus", "", "const x = { foo: 'bar', baz: Object.assign({}, a) }", []string{"Object.assign({}, a)@29"}, "const x = { foo: 'bar', baz: { ...a} }"},
		{"corpus", "", "\n                import Foo from 'foo';\n                Object.assign({ foo: Foo });\n            ", []string{"Object.assign({ foo: Foo })@56"}, "\n                import Foo from 'foo';\n                ({foo: Foo});\n            "},
		{"corpus", "", "\n                import Foo from 'foo';\n                Object.assign({}, Foo);\n            ", []string{"Object.assign({}, Foo)@56"}, "\n                import Foo from 'foo';\n                ({ ...Foo});\n            "},
		{"corpus", "", "\n                const Foo = require('foo');\n                Object.assign({ foo: Foo });\n            ", []string{"Object.assign({ foo: Foo })@61"}, "\n                const Foo = require('foo');\n                ({foo: Foo});\n            "},
		{"corpus", "", "\n                import { Something as somethingelse } from 'foo';\n                Object.assign({}, somethingelse);\n            ", []string{"Object.assign({}, somethingelse)@83"}, "\n                import { Something as somethingelse } from 'foo';\n                ({ ...somethingelse});\n            "},
		{"corpus", "", "\n                import { foo } from 'foo';\n                Object.assign({ foo: Foo });\n            ", []string{"Object.assign({ foo: Foo })@60"}, "\n                import { foo } from 'foo';\n                ({foo: Foo});\n            "},
		{"corpus", "", "\n                const Foo = require('foo');\n                Object.assign({}, Foo);\n            ", []string{"Object.assign({}, Foo)@61"}, "\n                const Foo = require('foo');\n                ({ ...Foo});\n            "},
		{"corpus", "", "\n                const actions = Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput,\n                    },\n                    this.props.actions\n                );\n            ", []string{"Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput,\n                    },\n                    this.props.actions\n                )@33"}, "\n                const actions = {\n                    onChangeInput: this.handleChangeInput,\n                    ...this.props.actions\n                };\n            "},
		{"corpus", "", "\n                const actions = Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput, //\n                    },\n                    this.props.actions\n                );\n            ", []string{"Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput, //\n                    },\n                    this.props.actions\n                )@33"}, "\n                const actions = {\n                    onChangeInput: this.handleChangeInput, //\n                    \n                    ...this.props.actions\n                };\n            "},
		{"corpus", "", "\n                const actions = Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput //\n                    },\n                    this.props.actions\n                );\n            ", []string{"Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput //\n                    },\n                    this.props.actions\n                )@33"}, "\n                const actions = {\n                    onChangeInput: this.handleChangeInput //\n                    ,\n                    ...this.props.actions\n                };\n            "},
		{"corpus", "", "\n                const actions = Object.assign(\n                    (\n                        {\n                            onChangeInput: this.handleChangeInput\n                        }\n                    ),\n                    (\n                        this.props.actions\n                    )\n                );\n            ", []string{"Object.assign(\n                    (\n                        {\n                            onChangeInput: this.handleChangeInput\n                        }\n                    ),\n                    (\n                        this.props.actions\n                    )\n                )@33"}, "\n                const actions = {\n                    \n                            onChangeInput: this.handleChangeInput\n                        ,\n                    ...(\n                        this.props.actions\n                    )\n                };\n            "},
		{"corpus", "", "\n                eventData = Object.assign({}, eventData, { outsideLocality: `${originLocality} - ${destinationLocality}` })\n            ", []string{"Object.assign({}, eventData, { outsideLocality: `${originLocality} - ${destinationLocality}` })@29"}, "\n                eventData = { ...eventData, outsideLocality: `${originLocality} - ${destinationLocality}`}\n            "},
		{"corpus", "", "Object.assign({ });", []string{"Object.assign({ })@0"}, "({});"},
		{"corpus", "", "Object.assign({\n});", []string{"Object.assign({\n})@0"}, "({});"},
		{"corpus", "", "globalThis.Object.assign({ });", []string{"globalThis.Object.assign({ })@0"}, "({});"},
		{"corpus", "", "globalThis.Object.assign({\n});", []string{"globalThis.Object.assign({\n})@0"}, "({});"},
		{"corpus", "", "\n                function foo () { var globalThis = bar; }\n                globalThis.Object.assign({ });\n            ", []string{"globalThis.Object.assign({ })@75"}, "\n                function foo () { var globalThis = bar; }\n                ({});\n            "},
		{"corpus", "", "\n                const Foo = require('foo');\n                globalThis.Object.assign({ foo: Foo });\n            ", []string{"globalThis.Object.assign({ foo: Foo })@61"}, "\n                const Foo = require('foo');\n                ({foo: Foo});\n            "},
		{"corpus", "", "Object.assign({ get a() {}, set b(val) {} })", []string{"Object.assign({ get a() {}, set b(val) {} })@0"}, "({get a() {}, set b(val) {}})"},
		{"edge", "", "const { assign } = Object; assign({}, a);", []string{"assign({}, a)@27"}, "const { assign } = Object; ({ ...a});"},
		{"edge", "", "const { assign: merge } = Object; merge({ a: 1 }, b);", []string{"merge({ a: 1 }, b)@34"}, "const { assign: merge } = Object; ({a: 1, ...b});"},
		{"edge", "", "const o = Object; o.assign({}, a);", []string{"o.assign({}, a)@18"}, "const o = Object; ({ ...a});"},
		{"edge", "", "globalThis.Object.assign({}, a);", []string{"globalThis.Object.assign({}, a)@0"}, "({ ...a});"},
		{"edge", "", "Object?.assign({}, a);", []string{"Object?.assign({}, a)@0"}, "({ ...a});"},
		{"edge", "", "Object.assign?.({}, a);", []string{"Object.assign?.({}, a)@0"}, "({ ...a});"},
		{"edge", "", "(Object.assign)({}, a);", []string{"(Object.assign)({}, a)@0"}, "({ ...a});"},
		{"edge", "", "function f(Object) { return Object.assign({}, a); }", []string{}, ""},
		{"edge", "", "Object = {}; Object.assign({}, a);", []string{}, ""},
		{"edge", "", "function g() { Object = 1; } Object.assign({}, a);", []string{}, ""},
		{"edge", "", "const x = Object.assign({}, a);", []string{"Object.assign({}, a)@10"}, "const x = { ...a};"},
		{"edge", "", "Object.assign(x, a);", []string{}, ""},
	}

	for index, testCase := range cases {
		result := rule_testing.RunTypedVerbatimWithOptions(t, PreferObjectSpread, "input.ts", testCase.code, nil)
		source := result.SourceFile.Text()
		got := []string{}
		fixes := 0
		for _, diagnostic := range result.Diagnostics {
			got = append(got, fmt.Sprintf("%s@%d", source[diagnostic.Range.Pos():diagnostic.Range.End()], diagnostic.Range.Pos()))
			fixes += len(diagnostic.Fixes)
		}
		want := append([]string{}, testCase.want...)
		sort.Strings(got)
		sort.Strings(want)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("case %d (%s, %s): %q\n got: %q\nwant: %q", index, testCase.origin, testCase.options, testCase.code, got, want)
			continue
		}

		fixed := ""
		if fixes > 0 {
			fixed = fixToConvergence(t, testCase.code)
		}
		if fixed != testCase.fixed {
			t.Errorf("case %d (%s, %s): %q\nfixed to: %q\n  ESLint: %q", index, testCase.origin, testCase.options, testCase.code, fixed, testCase.fixed)
		}
	}
}

// fixToConvergence applies the rule's fixes the way ESLint's verifyAndFix does: in each pass, every
// finding's fixes together, in source order, skipping a finding whose fixes start before the last one
// applied ends, then the rule runs again over the result, until a pass fixes nothing or ten have run.
// Nested calls fix outside in this way, as ESLint's own answer for them shows.
func fixToConvergence(t *testing.T, code string) string {
	t.Helper()
	for pass := 0; pass < 10; pass++ {
		result := rule_testing.RunTypedVerbatimWithOptions(t, PreferObjectSpread, "input.ts", code, nil)
		type finding struct {
			start, end int
			fixes      []rule.Fix
		}
		var findings []finding
		for _, diagnostic := range result.Diagnostics {
			if len(diagnostic.Fixes) == 0 {
				continue
			}
			current := finding{start: diagnostic.Fixes[0].Range.Pos(), end: diagnostic.Fixes[0].Range.End(), fixes: diagnostic.Fixes}
			for _, fix := range diagnostic.Fixes {
				current.start = min(current.start, fix.Range.Pos())
				current.end = max(current.end, fix.Range.End())
			}
			findings = append(findings, current)
		}
		if len(findings) == 0 {
			return code
		}
		sort.Slice(findings, func(first, second int) bool { return findings[first].start < findings[second].start })
		var applied []rule.Fix
		lastEnd := -1
		for _, current := range findings {
			if current.start < lastEnd {
				continue
			}
			applied = append(applied, current.fixes...)
			lastEnd = current.end
		}
		sort.Slice(applied, func(first, second int) bool { return applied[first].Range.Pos() > applied[second].Range.Pos() })
		for _, fix := range applied {
			code = code[:fix.Range.Pos()] + fix.Text + code[fix.Range.End():]
		}
	}
	return code
}
