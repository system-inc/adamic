package core

import (
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// objectShorthandFile is where the fixtures pretend to live.
const objectShorthandFile = "/repository/source/ObjectShorthand.ts"

// objectShorthandOf builds the settings a config carrying this option would decode to.
//
// The argument is the option list as JSON TEXT, exactly the bytes the config layer hands the
// decoder, so the fixtures exercise the three-shape `anyOf` decoder rather than bypassing it.
func objectShorthandOf(optionJson string) any {
	settings, err := DecodeObjectShorthandOptions([]byte(optionJson))
	if err != nil {
		panic(err)
	}
	return settings
}

// The corpus is ESLint's own, taken from
// /tmp/lint-sources-fresh/eslint/tests/lib/rules/object-shorthand.js by RUNNING that file with
// RuleTester intercepted. 262 cases from one RuleTester.run call, across 20 distinct option shapes.
//
// Every expectation is what the INSTALLED ESLint 10.8.1 rule answered, driven through the Linter
// interface under `sourceType: module` with the typescript-eslint parser. The oracle reproduced all
// 262 declared verdicts AND all 122 declared outputs, 16 of which are `output: null` declines.
//
// This rule has SEVEN message ids, so the findings column carries the real id per finding rather
// than a repeated one: a rule reporting the right count under the wrong id would otherwise pass.
//
//	134 reporting cases (16 declines), 128 clean cases.
func TestObjectShorthandFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source         string
		settings       any
		findings       []string
		fixed          *string
		declinesRepair bool
	}{
		{source: "var x = {x: x}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {x}"), declinesRepair: false},
		{source: "var x = {'x': x}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {x}"), declinesRepair: false},
		{source: "var x = {y: y, x: x}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id, messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {y, x}"), declinesRepair: false},
		{source: "var x = {y: z, x: x, a: b}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {y: z, x, a: b}"), declinesRepair: false},
		{source: "var x = {y: z,\n x: x,\n a: b\n // comment \n}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {y: z,\n x,\n a: b\n // comment \n}"), declinesRepair: false},
		{source: "var x = {y: z,\n a: b,\n // comment \nf: function() {}}", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {y: z,\n a: b,\n // comment \nf() {}}"), declinesRepair: false},
		{source: "var x = {a: b,\n/* comment */\ny: y\n }", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {a: b,\n/* comment */\ny\n }"), declinesRepair: false},
		{source: "var x = {\n  a: b,\n  /* comment */\n  y: y\n}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {\n  a: b,\n  /* comment */\n  y\n}"), declinesRepair: false},
		{source: "var x = {\n  f: function() {\n    /* comment */\n    a(b);\n    }\n  }", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {\n  f() {\n    /* comment */\n    a(b);\n    }\n  }"), declinesRepair: false},
		{source: "var x = {\n  [f]: function() {\n    /* comment */\n    a(b);\n    }\n  }", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {\n  [f]() {\n    /* comment */\n    a(b);\n    }\n  }"), declinesRepair: false},
		{source: "var x = {\n  f: function*() {\n    /* comment */\n    a(b);\n    }\n  }", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {\n  *f() {\n    /* comment */\n    a(b);\n    }\n  }"), declinesRepair: false},
		{source: "var x = {\n  f: /* comment */ function() {\n  }\n  }", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {\n f /* comment */: function() {\n  }\n  }", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {a: /* comment */ a}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {a /* comment */: a}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {a: (a /* comment */)}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {'a': /* comment */ a}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {'a': (a /* comment */)}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {'a' /* comment */: a}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {y: function() {}}", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {y() {}}"), declinesRepair: false},
		{source: "var x = {y: function*() {}}", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {*y() {}}"), declinesRepair: false},
		{source: "var x = {x: y, y: z, a: a}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {x: y, y: z, a}"), declinesRepair: false},
		{source: "var x = {ConstructorFunction: function(){}, a: b}", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {ConstructorFunction(){}, a: b}"), declinesRepair: false},
		{source: "var x = {x: y, y: z, a: function(){}, b() {}}", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {x: y, y: z, a(){}, b() {}}"), declinesRepair: false},
		{source: "var x = {x: x, y: function() {}}", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {x, y() {}}"), declinesRepair: false},
		{source: "doSomething({x: x})", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("doSomething({x})"), declinesRepair: false},
		{source: "doSomething({'x': x})", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("doSomething({x})"), declinesRepair: false},
		{source: "doSomething({a: 'a', 'x': x})", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("doSomething({a: 'a', x})"), declinesRepair: false},
		{source: "doSomething({y: function() {}})", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("doSomething({y() {}})"), declinesRepair: false},
		{source: "doSomething({[y]: function() {}})", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("doSomething({[y]() {}})"), declinesRepair: false},
		{source: "doSomething({['y']: function() {}})", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("doSomething({['y']() {}})"), declinesRepair: false},
		{source: "({ foo: async function () {} })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ async foo () {} })"), declinesRepair: false},
		{source: "({ 'foo': async function() {} })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ async 'foo'() {} })"), declinesRepair: false},
		{source: "({ [foo]: async function() {} })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ async [foo]() {} })"), declinesRepair: false},
		{source: "({ [foo.bar]: function*() {} })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ *[foo.bar]() {} })"), declinesRepair: false},
		{source: "({ [foo   ]: function() {} })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ [foo   ]() {} })"), declinesRepair: false},
		{source: "({ [ foo ]: async function() {} })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ async [ foo ]() {} })"), declinesRepair: false},
		{source: "({ foo: function *() {} })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ *foo() {} })"), declinesRepair: false},
		{source: "({ [  foo   ]: function() {} })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ [  foo   ]() {} })"), declinesRepair: false},
		{source: "({ [  foo]: function() {} })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ [  foo]() {} })"), declinesRepair: false},
		{source: "var x = {y: function() {}}", settings: objectShorthandOf("[\"methods\"]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {y() {}}"), declinesRepair: false},
		{source: "var x = {x, y() {}, z: function() {}}", settings: objectShorthandOf("[\"methods\"]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {x, y() {}, z() {}}"), declinesRepair: false},
		{source: "var x = {ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"methods\"]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {ConstructorFunction(){}, a: b}"), declinesRepair: false},
		{source: "var x = {[y]: function() {}}", settings: objectShorthandOf("[\"methods\"]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {[y]() {}}"), declinesRepair: false},
		{source: "({ [(foo)]: function() { return; } })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ [(foo)]() { return; } })"), declinesRepair: false},
		{source: "({ [(foo)]: async function() { return; } })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ async [(foo)]() { return; } })"), declinesRepair: false},
		{source: "({ [(((((((foo)))))))]: function() { return; } })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ [(((((((foo)))))))]() { return; } })"), declinesRepair: false},
		{source: "({ [(foo)]() { return; } })", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("({ [(foo)]: function() { return; } })"), declinesRepair: false},
		{source: "({ async [(foo)]() { return; } })", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("({ [(foo)]: async function() { return; } })"), declinesRepair: false},
		{source: "({ *[((foo))]() { return; } })", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("({ [((foo))]: function*() { return; } })"), declinesRepair: false},
		{source: "({ [(((((((foo)))))))]() { return; } })", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("({ [(((((((foo)))))))]: function() { return; } })"), declinesRepair: false},
		{source: "({ 'foo bar'() { return; } })", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("({ 'foo bar': function() { return; } })"), declinesRepair: false},
		{source: "({ *foo() { return; } })", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("({ foo: function*() { return; } })"), declinesRepair: false},
		{source: "({ async foo() { return; } })", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("({ foo: async function() { return; } })"), declinesRepair: false},
		{source: "({ *['foo bar']() { return; } })", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("({ ['foo bar']: function*() { return; } })"), declinesRepair: false},
		{source: "var x = {x: x}", settings: objectShorthandOf("[\"properties\"]"), findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {x}"), declinesRepair: false},
		{source: "var x = {a, b, c(){}, x: x}", settings: objectShorthandOf("[\"properties\"]"), findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {a, b, c(){}, x}"), declinesRepair: false},
		{source: "var x = {y() {}}", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("var x = {y: function() {}}"), declinesRepair: false},
		{source: "var x = {*y() {}}", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("var x = {y: function*() {}}"), declinesRepair: false},
		{source: "var x = {y}", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedPropertyLongform.Id}, fixed: pointerTo("var x = {y: y}"), declinesRepair: false},
		{source: "var x = {y, a: b, *x(){}}", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedPropertyLongform.Id, messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("var x = {y: y, a: b, x: function*(){}}"), declinesRepair: false},
		{source: "var x = {y: {x}}", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedPropertyLongform.Id}, fixed: pointerTo("var x = {y: {x: x}}"), declinesRepair: false},
		{source: "var x = {ConstructorFunction(){}, a: b}", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("var x = {ConstructorFunction: function(){}, a: b}"), declinesRepair: false},
		{source: "var x = {notConstructorFunction(){}, b: c}", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("var x = {notConstructorFunction: function(){}, b: c}"), declinesRepair: false},
		{source: "var x = {foo: foo, bar: baz, ...qux}", settings: objectShorthandOf("[\"always\"]"), findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {foo, bar: baz, ...qux}"), declinesRepair: false},
		{source: "var x = {foo, bar: baz, ...qux}", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedPropertyLongform.Id}, fixed: pointerTo("var x = {foo: foo, bar: baz, ...qux}"), declinesRepair: false},
		{source: "var x = {y: function() {}}", settings: objectShorthandOf("[\"methods\", {\"ignoreConstructors\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {y() {}}"), declinesRepair: false},
		{source: "var x = {_y: function() {}}", settings: objectShorthandOf("[\"methods\", {\"ignoreConstructors\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {_y() {}}"), declinesRepair: false},
		{source: "var x = {$y: function() {}}", settings: objectShorthandOf("[\"methods\", {\"ignoreConstructors\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {$y() {}}"), declinesRepair: false},
		{source: "var x = {__y: function() {}}", settings: objectShorthandOf("[\"methods\", {\"ignoreConstructors\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {__y() {}}"), declinesRepair: false},
		{source: "var x = {_0y: function() {}}", settings: objectShorthandOf("[\"methods\", {\"ignoreConstructors\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {_0y() {}}"), declinesRepair: false},
		{source: "var x = { afoob: function() {} }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^foo$\"}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = { afoob() {} }"), declinesRepair: false},
		{source: "var x = { afoob: function() {} }", settings: objectShorthandOf("[\"methods\", {\"methodsIgnorePattern\": \"^foo$\"}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = { afoob() {} }"), declinesRepair: false},
		{source: "var x = { 'afoob': function() {} }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^foo$\"}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = { 'afoob'() {} }"), declinesRepair: false},
		{source: "var x = { 1234: function() {} }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^123$\"}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = { 1234() {} }"), declinesRepair: false},
		{source: "var x = { bar: function() {} }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"foo\"}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = { bar() {} }"), declinesRepair: false},
		{source: "var x = { [foo]: function() {} }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"foo\"}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = { [foo]() {} }"), declinesRepair: false},
		{source: "var x = { foo: foo }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^foo$\"}]"), findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = { foo }"), declinesRepair: false},
		{source: "var x = {a: a}", settings: objectShorthandOf("[\"always\", {\"avoidQuotes\": true}]"), findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: pointerTo("var x = {a}"), declinesRepair: false},
		{source: "var x = {a: function(){}}", settings: objectShorthandOf("[\"methods\", {\"avoidQuotes\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {a(){}}"), declinesRepair: false},
		{source: "var x = {[a]: function(){}}", settings: objectShorthandOf("[\"methods\", {\"avoidQuotes\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("var x = {[a](){}}"), declinesRepair: false},
		{source: "var x = {'a'(){}}", settings: objectShorthandOf("[\"always\", {\"avoidQuotes\": true}]"), findings: []string{messageObjectShorthandExpectedLiteralMethodLongform.Id}, fixed: pointerTo("var x = {'a': function(){}}"), declinesRepair: false},
		{source: "var x = {['a'](){}}", settings: objectShorthandOf("[\"methods\", {\"avoidQuotes\": true}]"), findings: []string{messageObjectShorthandExpectedLiteralMethodLongform.Id}, fixed: pointerTo("var x = {['a']: function(){}}"), declinesRepair: false},
		{source: "var x = {a: a, b}", settings: objectShorthandOf("[\"consistent\"]"), findings: []string{messageObjectShorthandUnexpectedMix.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {b, c: d, f: g}", settings: objectShorthandOf("[\"consistent\"]"), findings: []string{messageObjectShorthandUnexpectedMix.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {foo, bar: baz, ...qux}", settings: objectShorthandOf("[\"consistent\"]"), findings: []string{messageObjectShorthandUnexpectedMix.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {a: a, b: b}", settings: objectShorthandOf("[\"consistent-as-needed\"]"), findings: []string{messageObjectShorthandExpectedAllPropertiesShorthanded.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {a, z: function z(){}}", settings: objectShorthandOf("[\"consistent-as-needed\"]"), findings: []string{messageObjectShorthandUnexpectedMix.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {foo: function() {}}", settings: objectShorthandOf("[\"consistent-as-needed\"]"), findings: []string{messageObjectShorthandExpectedAllPropertiesShorthanded.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {a: a, b: b, ...baz}", settings: objectShorthandOf("[\"consistent-as-needed\"]"), findings: []string{messageObjectShorthandExpectedAllPropertiesShorthanded.Id}, fixed: nil, declinesRepair: true},
		{source: "var x = {foo, bar: bar, ...qux}", settings: objectShorthandOf("[\"consistent-as-needed\"]"), findings: []string{messageObjectShorthandUnexpectedMix.Id}, fixed: nil, declinesRepair: true},
		{source: "({ x: (arg => { return; }) })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ x(arg) { return; } })"), declinesRepair: false},
		{source: "({ x: () => { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ x() { return; } })"), declinesRepair: false},
		{source: "({ x() { return; }, y: () => { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ x() { return; }, y() { return; } })"), declinesRepair: false},
		{source: "({ x: () => { return; }, y: () => foo })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ x() { return; }, y: () => foo })"), declinesRepair: false},
		{source: "({ x: () => { return; }, y: () => { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ x() { return; }, y() { return; } })"), declinesRepair: false},
		{source: "({ x: foo => { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ x(foo) { return; } })"), declinesRepair: false},
		{source: "({ x: (foo = 1) => { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ x(foo = 1) { return; } })"), declinesRepair: false},
		{source: "({ x: ({ foo: bar = 1 } = {}) => { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ x({ foo: bar = 1 } = {}) { return; } })"), declinesRepair: false},
		{source: "({ x: () => { function foo() { this; } } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ x() { function foo() { this; } } })"), declinesRepair: false},
		{source: "({ x: () => { var foo = function() { arguments; } } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ x() { var foo = function() { arguments; } } })"), declinesRepair: false},
		{source: "({ x: () => { function foo() { arguments; } } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ x() { function foo() { arguments; } } })"), declinesRepair: false},
		{source: "\n                ({\n                    x: () => {\n                        class Foo extends Bar {\n                            constructor() {\n                                super();\n                            }\n                        }\n                    }\n                })\n            ", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("\n                ({\n                    x() {\n                        class Foo extends Bar {\n                            constructor() {\n                                super();\n                            }\n                        }\n                    }\n                })\n            "), declinesRepair: false},
		{source: "\n                ({\n                    x: () => {\n                        function foo() {\n                            new.target;\n                        }\n                    }\n                })\n            ", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("\n                ({\n                    x() {\n                        function foo() {\n                            new.target;\n                        }\n                    }\n                })\n            "), declinesRepair: false},
		{source: "({ 'foo bar': () => { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ 'foo bar'() { return; } })"), declinesRepair: false},
		{source: "({ [foo]: () => { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ [foo]() { return; } })"), declinesRepair: false},
		{source: "({ a: 1, foo: async (bar = 1) => { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ a: 1, async foo(bar = 1) { return; } })"), declinesRepair: false},
		{source: "({ [ foo ]: async bar => { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ async [ foo ](bar) { return; } })"), declinesRepair: false},
		{source: "({ key: (arg = () => {}) => {} })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ key(arg = () => {}) {} })"), declinesRepair: false},
		{source: "\n                function foo() {\n                    var x = {\n                        x: () => {\n                            this;\n                            return { y: () => { foo; } };\n                        }\n                    };\n                }\n            ", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("\n                function foo() {\n                    var x = {\n                        x: () => {\n                            this;\n                            return { y() { foo; } };\n                        }\n                    };\n                }\n            "), declinesRepair: false},
		{source: "\n                function foo() {\n                    var x = {\n                        x: () => {\n                            ({ y: () => { foo; } });\n                            this;\n                        }\n                    };\n                }\n            ", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("\n                function foo() {\n                    var x = {\n                        x: () => {\n                            ({ y() { foo; } });\n                            this;\n                        }\n                    };\n                }\n            "), declinesRepair: false},
		{source: "({ a: (function(){ return foo; }) })", settings: nil, findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ a(){ return foo; } })"), declinesRepair: false},
		{source: "({ a: (() => { return foo; }) })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ a() { return foo; } })"), declinesRepair: false},
		{source: "({ a: ((arg) => { return foo; }) })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ a(arg) { return foo; } })"), declinesRepair: false},
		{source: "({ a: ((arg, arg2) => { return foo; }) })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ a(arg, arg2) { return foo; } })"), declinesRepair: false},
		{source: "({ a: (async () => { return foo; }) })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ async a() { return foo; } })"), declinesRepair: false},
		{source: "({ a: (async (arg) => { return foo; }) })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ async a(arg) { return foo; } })"), declinesRepair: false},
		{source: "({ a: (async (arg, arg2) => { return foo; }) })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ async a(arg, arg2) { return foo; } })"), declinesRepair: false},
		{source: "({ a: async function*() {} })", settings: objectShorthandOf("[\"always\"]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("({ async *a() {} })"), declinesRepair: false},
		{source: "({ async* a() {} })", settings: objectShorthandOf("[\"never\"]"), findings: []string{messageObjectShorthandExpectedMethodLongform.Id}, fixed: pointerTo("({ a: async function*() {} })"), declinesRepair: false},
		{source: "const test = {\n    key: <T>(): void => { },\n    key: async <T>(): Promise<void> => { },\n\n    key: <T>(arg: T): T => { return arg },\n    key: async <T>(arg: T): Promise<T> => { return arg },\n}", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("const test = {\n    key<T>(): void { },\n    async key<T>(): Promise<void> { },\n\n    key<T>(arg: T): T { return arg },\n    async key<T>(arg: T): Promise<T> { return arg },\n}"), declinesRepair: false},
		{source: "const test = {\n    key: (): void => {x()},\n    key: ( (): void => {x()} ),\n    key: ( (): (void) => {x()} ),\n\n    key: (arg: t): void => {x()},\n    key: ( (arg: t): void => {x()} ),\n    key: ( (arg: t): (void) => {x()} ),\n\n    key: (arg: t, arg2: t): void => {x()},\n    key: ( (arg: t, arg2: t): void => {x()} ),\n    key: ( (arg: t, arg2: t): (void) => {x()} ),\n\n    key: async (): void => {x()},\n    key: ( async (): void => {x()} ),\n    key: ( async (): (void) => {x()} ),\n\n    key: async (arg: t): void => {x()},\n    key: ( async (arg: t): void => {x()} ),\n    key: ( async (arg: t): (void) => {x()} ),\n\n    key: async (arg: t, arg2: t): void => {x()},\n    key: ( async (arg: t, arg2: t): void => {x()} ),\n    key: ( async (arg: t, arg2: t): (void) => {x()} ),\n}", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]"), findings: []string{messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id, messageObjectShorthandExpectedMethodShorthand.Id}, fixed: pointerTo("const test = {\n    key(): void {x()},\n    key(): void {x()},\n    key(): (void) {x()},\n\n    key(arg: t): void {x()},\n    key(arg: t): void {x()},\n    key(arg: t): (void) {x()},\n\n    key(arg: t, arg2: t): void {x()},\n    key(arg: t, arg2: t): void {x()},\n    key(arg: t, arg2: t): (void) {x()},\n\n    async key(): void {x()},\n    async key(): void {x()},\n    async key(): (void) {x()},\n\n    async key(arg: t): void {x()},\n    async key(arg: t): void {x()},\n    async key(arg: t): (void) {x()},\n\n    async key(arg: t, arg2: t): void {x()},\n    async key(arg: t, arg2: t): void {x()},\n    async key(arg: t, arg2: t): (void) {x()},\n}"), declinesRepair: false},
		{source: "({ val: /** regular comment */ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
		{source: "({ val: /** @param {string} name */ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
		{source: "({ val: /** @returns {number} */ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
		{source: "({ val: /** @description some text */ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
		{source: "({ val: /**\n * @param {string} name\n */ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
		{source: "({ val: /**\n  * @returns {number}\n  */ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
		{source: "({ val: /**\n   * @description some text\n   */ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
		{source: "({ val: /**\n\t* @param {string} name\n\t*/ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
		{source: "({ val: /**\n\t * @returns {number}\n\t */ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
		{source: "({ val: /**\n  *  @param   {string}  name  \n  */ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
		{source: "({ val: /**\n *  @returns   {number} result\n */ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
		{source: "({ val: /**\n   *\n   * @param {string} name\n   * @returns {number}\n   */ (val) })", settings: nil, findings: []string{messageObjectShorthandExpectedPropertyShorthand.Id}, fixed: nil, declinesRepair: false},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, ObjectShorthand, objectShorthandFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != len(testCase.findings) {
			t.Errorf("%q with %+v: expected %d findings, got %d %v",
				testCase.source, testCase.settings, len(testCase.findings),
				len(result.Diagnostics), result.MessageIds())
			continue
		}
		rule_testing.ExpectFindings(t, result, testCase.findings...)

		proposed := 0
		for _, diagnostic := range result.Diagnostics {
			proposed += len(diagnostic.Fixes)
		}
		switch {
		case testCase.declinesRepair:
			if proposed != 0 {
				t.Errorf("%q: upstream declines to repair this, so the rule must propose no fix, got %d",
					testCase.source, proposed)
			}
		case testCase.fixed != nil:
			// `RunWithOptions` does NOT trim, unlike `RunTyped`: the untyped harness hands the
			// rule the source exactly as written, so the expectation is the corpus string as-is.
			// prefer-arrow-callback needs the opposite because it is a typed rule.
			rule_testing.ExpectFixedSource(t, result, *testCase.fixed)
		}
	}
}

// TestObjectShorthandStaysSilent carries upstream's clean cases.
func TestObjectShorthandStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
	}{
		{source: "var x = {y() {}}", settings: nil},
		{source: "var x = {y}", settings: nil},
		{source: "var x = {a: b}", settings: nil},
		{source: "var x = {a: 'a'}", settings: nil},
		{source: "var x = {'a': 'a'}", settings: nil},
		{source: "var x = {'a': b}", settings: nil},
		{source: "var x = {y(x) {}}", settings: nil},
		{source: "var {x,y,z} = x", settings: nil},
		{source: "var {x: {y}} = z", settings: nil},
		{source: "var x = {*x() {}}", settings: nil},
		{source: "var x = {x: y}", settings: nil},
		{source: "var x = {x: y, y: z}", settings: nil},
		{source: "var x = {x: y, y: z, z: 'z'}", settings: nil},
		{source: "var x = {x() {}, y: z, l(){}}", settings: nil},
		{source: "var x = {x: y, y: z, a: b}", settings: nil},
		{source: "var x = {x: y, y: z, 'a': b}", settings: nil},
		{source: "var x = {x: y, y() {}, z: a}", settings: nil},
		{source: "var x = {[y]: y}", settings: nil},
		{source: "doSomething({x: y})", settings: nil},
		{source: "doSomething({'x': y})", settings: nil},
		{source: "doSomething({x: 'x'})", settings: nil},
		{source: "doSomething({'x': 'x'})", settings: nil},
		{source: "doSomething({y() {}})", settings: nil},
		{source: "doSomething({x: y, y() {}})", settings: nil},
		{source: "doSomething({y() {}, z: a})", settings: nil},
		{source: "!{ a: function a(){} };", settings: nil},
		{source: "var x = {y: (x)=>x}", settings: nil},
		{source: "doSomething({y: (x)=>x})", settings: nil},
		{source: "var x = {y: (x)=>x, y: a}", settings: nil},
		{source: "doSomething({x, y: (x)=>x})", settings: nil},
		{source: "({ foo: x => { return; }})", settings: nil},
		{source: "({ foo: (x) => { return; }})", settings: nil},
		{source: "({ foo: () => { return; }})", settings: nil},
		{source: "var x = {get y() {}}", settings: nil},
		{source: "var x = {set y(z) {}}", settings: nil},
		{source: "var x = {get y() {}, set y(z) {}}", settings: nil},
		{source: "doSomething({get y() {}})", settings: nil},
		{source: "doSomething({set y(z) {}})", settings: nil},
		{source: "doSomething({get y() {}, set y(z) {}})", settings: nil},
		{source: "var x = {[y]: y}", settings: objectShorthandOf("[\"properties\"]")},
		{source: "var x = {['y']: 'y'}", settings: objectShorthandOf("[\"properties\"]")},
		{source: "var x = {['y']: y}", settings: objectShorthandOf("[\"properties\"]")},
		{source: "var x = {[y]() {}}", settings: objectShorthandOf("[\"methods\"]")},
		{source: "var x = {[y]: function x() {}}", settings: objectShorthandOf("[\"methods\"]")},
		{source: "var x = {[y]: y}", settings: objectShorthandOf("[\"methods\"]")},
		{source: "var x = {y() {}}", settings: objectShorthandOf("[\"methods\"]")},
		{source: "var x = {x, y() {}, a:b}", settings: objectShorthandOf("[\"methods\"]")},
		{source: "var x = {y}", settings: objectShorthandOf("[\"properties\"]")},
		{source: "var x = {y: {b}}", settings: objectShorthandOf("[\"properties\"]")},
		{source: "var x = {a: n, c: d, f: g}", settings: objectShorthandOf("[\"never\"]")},
		{source: "var x = {a: function(){}, b: {c: d}}", settings: objectShorthandOf("[\"never\"]")},
		{source: "var x = {ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"always\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {_ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"always\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {$ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"always\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {__ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"always\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {_0ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"always\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {notConstructorFunction(){}, b: c}", settings: objectShorthandOf("[\"always\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"methods\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {_ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"methods\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {$ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"methods\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {__ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"methods\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {_0ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"methods\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {notConstructorFunction(){}, b: c}", settings: objectShorthandOf("[\"methods\", {\"ignoreConstructors\": true}]")},
		{source: "var x = {ConstructorFunction: function(){}, a: b}", settings: objectShorthandOf("[\"never\"]")},
		{source: "var x = {notConstructorFunction: function(){}, b: c}", settings: objectShorthandOf("[\"never\"]")},
		{source: "var x = { foo: function() {}  }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^foo$\"}]")},
		{source: "var x = { foo: function() {}  }", settings: objectShorthandOf("[\"methods\", {\"methodsIgnorePattern\": \"^foo$\"}]")},
		{source: "var x = { foo: function*() {}  }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^foo$\"}]")},
		{source: "var x = { foo: async function() {}  }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^foo$\"}]")},
		{source: "var x = { foo: () => { return 5; }  }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^foo$\", \"avoidExplicitReturnArrows\": true}]")},
		{source: "var x = { 'foo': function() {}  }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^foo$\"}]")},
		{source: "var x = { ['foo']: function() {}  }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^foo$\"}]")},
		{source: "var x = { 123: function() {}  }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^123$\"}]")},
		{source: "var x = { afoob: function() {}  }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"foo\"}]")},
		{source: "var x = { afoob: function() {}  }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^.foo.$\"}]")},
		{source: "var x = { '\U0001f44dfoo\U0001f44d': function() {}  }", settings: objectShorthandOf("[\"always\", {\"methodsIgnorePattern\": \"^.foo.$\"}]")},
		{source: "var x = {'a': function(){}}", settings: objectShorthandOf("[\"always\", {\"avoidQuotes\": true}]")},
		{source: "var x = {['a']: function(){}}", settings: objectShorthandOf("[\"methods\", {\"avoidQuotes\": true}]")},
		{source: "var x = {'y': y}", settings: objectShorthandOf("[\"properties\", {\"avoidQuotes\": true}]")},
		{source: "let {a, b} = o;", settings: objectShorthandOf("[\"never\"]")},
		{source: "var x = {foo: foo, bar: bar, ...baz}", settings: objectShorthandOf("[\"never\"]")},
		{source: "var x = {a: a, b: b}", settings: objectShorthandOf("[\"consistent\"]")},
		{source: "var x = {a: b, c: d, f: g}", settings: objectShorthandOf("[\"consistent\"]")},
		{source: "var x = {a, b}", settings: objectShorthandOf("[\"consistent\"]")},
		{source: "var x = {a, b, get test() { return 1; }}", settings: objectShorthandOf("[\"consistent\"]")},
		{source: "var x = {...bar}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "var x = {foo, bar, ...baz}", settings: objectShorthandOf("[\"consistent\"]")},
		{source: "var x = {bar: baz, ...qux}", settings: objectShorthandOf("[\"consistent\"]")},
		{source: "var x = {...foo, bar: bar, baz: baz}", settings: objectShorthandOf("[\"consistent\"]")},
		{source: "var x = {a, b}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "var x = {a, b, get test(){return 1;}}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "var x = {0: 'foo'}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "var x = {'key': 'baz'}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "var x = {foo: 'foo'}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "var x = {[foo]: foo}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "var x = {foo: function foo() {}}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "var x = {[foo]: 'foo'}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "var x = {bar, ...baz}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "var x = {bar: baz, ...qux}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "var x = {...foo, bar, baz}", settings: objectShorthandOf("[\"consistent-as-needed\"]")},
		{source: "({ x: () => foo })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": false}]")},
		{source: "({ x: () => { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": false}]")},
		{source: "({ x: () => foo })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "({ x() { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "({ x() { return; }, y() { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "({ x() { return; }, y: () => foo })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "({ x: () => foo, y() { return; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "({ x: () => { this; } })", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "function foo() { ({ x: () => { arguments; } }) }", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "\n                class Foo extends Bar {\n                  constructor() {\n                      var foo = { x: () => { super(); } };\n                  }\n              }\n            ", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "\n                class Foo extends Bar {\n                    baz() {\n                        var foo = { x: () => { super.baz(); } };\n                    }\n                }\n            ", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "\n                function foo() {\n                    var x = { x: () => { new.target; } };\n                }\n            ", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "\n                function foo() {\n                    var x = {\n                        x: () => {\n                            var y = () => { this; };\n                        }\n                    };\n                }\n            ", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "\n                function foo() {\n                    var x = {\n                        x: () => {\n                            var y = () => { this; };\n                            function foo() { this; }\n                        }\n                    };\n                }\n            ", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "\n                function foo() {\n                    var x = {\n                        x: () => {\n                            return { y: () => { this; } };\n                        }\n                    };\n                }\n            ", settings: objectShorthandOf("[\"always\", {\"avoidExplicitReturnArrows\": true}]")},
		{source: "({ [foo.bar]: () => {} })", settings: objectShorthandOf("[\"always\", {\"ignoreConstructors\": true}]")},
		{source: "({ val: /** @type {number} */ (val) })", settings: nil},
		{source: "({ 'prop': /** @type {string} */ (prop) })", settings: nil},
		{source: "({ val: /**\n * @type {number}\n */ (val) })", settings: nil},
		{source: "({ val: /**\n  * @type {number}\n  */ (val) })", settings: nil},
		{source: "({ val: /**\n   * @type {number}\n   */ (val) })", settings: nil},
		{source: "({ val: /**\n\t* @type {number}\n\t*/ (val) })", settings: nil},
		{source: "({ val: /**\n\t * @type {number}\n\t */ (val) })", settings: nil},
		{source: "({ val: /**\n  *  @type   {number}  \n  */ (val) })", settings: nil},
		{source: "({ val: /**\n *  @type   {string} myParam\n */ (val) })", settings: nil},
		{source: "({ val: /**\n  *  @type   {Object} options\n  */ (val) })", settings: nil},
		{source: "({ val: /**\n\t *\t@type\t{Array}\n\t */ (val) })", settings: nil},
		{source: "({ val: /**\n   *\n   * @type {Function}\n   * @param {string} name\n   */ (val) })", settings: nil},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, ObjectShorthand, objectShorthandFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != 0 {
			t.Errorf("%q with %+v: expected no findings, got %d: %v",
				testCase.source, testCase.settings, len(result.Diagnostics), result.MessageIds())
			continue
		}
		rule_testing.ExpectClean(t, result)
	}
}

// TestObjectShorthandRepairKeepsSignatureAndSpacing covers repair details the corpus reaches only
// partly, and which no message id can see.
//
// This rule has seven message ids and 122 corpus `output` values, and the ids were right long
// before the repairs were: eleven defects in this port produced valid source that meant something
// else, and every one was caught by an `output` rather than by a count.
//
// The rows below are the ones the corpus under-covers. Its only generic case is written with
// ARROWS, so the function-expression branch's handling of `<T>` is unreached by it, and a mutation
// removing that handling survived all 262 imported fixtures. Every expectation is what the
// installed ESLint 10.8.1 rule produced.
func TestObjectShorthandRepairKeepsSignatureAndSpacing(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		fixed    string
	}{
		{
			// A generic FUNCTION EXPRESSION. Upstream slices from just after `function`, so the
			// type parameters and the space both survive.
			source:   "({ key: function <T>(a: T): T { return a; } });",
			settings: nil,
			fixed:    "({ key <T>(a: T): T { return a; } });",
		},
		{
			source:   "({ key: async function <T>(a: T): Promise<T> { return a; } });",
			settings: nil,
			fixed:    "({ async key <T>(a: T): Promise<T> { return a; } });",
		},
		{
			// A generic ARROW takes the other branch, where `<T>` sits before the parameter list
			// and the arrow token has to be dropped between them.
			source:   "({ key: <T>(a: T): T => { return a; } });",
			settings: objectShorthandOf(`["always", {"avoidExplicitReturnArrows": true}]`),
			fixed:    "({ key<T>(a: T): T { return a; } });",
		},
		{
			// A single parameter written without parentheses gains them, since a method's list
			// always has them.
			source:   "({ key: arg => { return arg; } });",
			settings: objectShorthandOf(`["always", {"avoidExplicitReturnArrows": true}]`),
			fixed:    "({ key(arg) { return arg; } });",
		},
		{
			// A parenthesised value is still the value, and the repair has to cover the
			// parenthesis or it leaves a stray `)`.
			source:   "({ a: (function(){ return foo; }) });",
			settings: nil,
			fixed:    "({ a(){ return foo; } });",
		},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, ObjectShorthand, objectShorthandFile,
			testCase.source, testCase.settings)
		if len(result.Diagnostics) == 0 {
			t.Errorf("%q: expected a finding, got none", testCase.source)
			continue
		}
		rule_testing.ExpectFixedSource(t, result, testCase.fixed)
	}
}

// TestObjectShorthandIgnoresAccessors covers the two accessor guards, both of which survived a
// mutation sweep against all 262 imported fixtures.
//
// Upstream drops getters and setters twice over: `Property:exit` returns early on
// `node.kind === "get" || "set"`, and `checkConsistency` filters them out of its count with
// `canHaveShorthand` before comparing. The corpus writes accessors, but never beside the
// configuration that would reveal either guard, so deleting them changes nothing it can see.
//
// Every verdict is what the installed ESLint 10.8.1 rule answered:
//
//	{ get a() {} }            never                 clean -- an accessor has no longform to want
//	{ set a(v) {} }           never                 clean
//	{ get a() {} }            always                clean
//	{ get a() {}, b: b }      consistent            clean -- the accessor is not counted, so the
//	                                                object is all-longform rather than mixed
//	{ get a() {}, b }         consistent            clean -- likewise all-shorthand
//	{ get a() {}, b: b }      consistent-as-needed  1 -- `b: b` could be shorthand and is not
//
// Rows four and five are the ones that fail without the consistency filter: counting the accessor
// as shorthand makes each object look mixed and reports `unexpectedMix`. Row six is the control,
// confirming the object is still being judged rather than skipped entirely.
func TestObjectShorthandIgnoresAccessors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		mode     string
		findings []string
	}{
		{source: "var o = { get a() {} };", mode: `["never"]`},
		{source: "var o = { set a(v) {} };", mode: `["never"]`},
		{source: "var o = { get a() {} };", mode: `["always"]`},
		{source: "var o = { get a() {}, b: b };", mode: `["consistent"]`},
		{source: "var o = { get a() {}, b };", mode: `["consistent"]`},
		{
			source: "var o = { get a() {}, b: b };", mode: `["consistent-as-needed"]`,
			findings: []string{messageObjectShorthandExpectedAllPropertiesShorthanded.Id},
		},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, ObjectShorthand, objectShorthandFile,
			testCase.source, objectShorthandOf(testCase.mode))
		if len(result.Diagnostics) != len(testCase.findings) {
			t.Errorf("%q under %s: expected %d findings, got %d %v",
				testCase.source, testCase.mode, len(testCase.findings),
				len(result.Diagnostics), result.MessageIds())
			continue
		}
		if len(testCase.findings) > 0 {
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		}
	}
}

// TestDecodeObjectShorthandOptions pins the list shapes the config layer delivers and the ones it
// must refuse, including upstream's mode-dependent flag pairing. Every fixture above passes a list
// this file wrote, so the refusals live here.
func TestDecodeObjectShorthandOptions(t *testing.T) {
	t.Parallel()

	decoded, err := DecodeObjectShorthandOptions([]byte(`["methods", {"ignoreConstructors": true, "avoidQuotes": true}]`))
	if err != nil {
		t.Fatalf("a legal mode and flags were refused: %v", err)
	}
	settings := decoded.(ObjectShorthandSettings)
	if settings.Mode != ObjectShorthandMethods || !settings.IgnoreConstructors || !settings.AvoidQuotes {
		t.Errorf("the second element was not read: %+v", settings)
	}

	for _, raw := range []string{
		// A bare string, which the config layer never delivers to a list rule.
		`"always"`,
		// The nested workaround spelling.
		`[["always", {"avoidQuotes": true}]]`,
		// Flags beside a mode upstream's schema gives none.
		`["never", {"avoidQuotes": true}]`,
		`["consistent-as-needed", {"ignoreConstructors": true}]`,
		// Beside properties, only avoidQuotes.
		`["properties", {"ignoreConstructors": true}]`,
		// A key outside the four, and a third element.
		`["always", {"avoidQuote": true}]`,
		`["always", {"avoidQuotes": true}, "never"]`,
		// A mode outside the six.
		`["sometimes"]`,
	} {
		if decoded, err := DecodeObjectShorthandOptions([]byte(raw)); err == nil {
			t.Errorf("%s decoded to %+v; it must be refused", raw, decoded)
		}
	}
}
