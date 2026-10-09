package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// setterReturnFile is where the fixtures pretend to live.
//
// A `.js` extension on purpose: upstream calls `change_rule_path_extension("js")` on this rule's
// Tester, because half the corpus is object-literal and class-body syntax that TypeScript would
// also flag on its own. The rule itself reads no file extension, so this only keeps the fixtures
// honest about what upstream measured.
const setterReturnFile = "/repository/source/SetterReturn.js"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_setter_return.rs`,
// extracted by parsing that file rather than by retyping: 108 pass and 34 fail. The snapshot records
// 42 diagnostics against those 34 inputs, so one finding per input would be wrong here. Seven inputs
// report more than once and each is written with its own count below, taken from the snapshot rather
// than from what this rule happens to do.
//
// Copied because a fixture a porter invents encodes the same belief as the port, and the case that
// catches a bug is the one nobody would think to write. The pass list is where that pays: it holds
// forty cases about `Object.defineProperty` descriptors that exist only to pin a gap upstream
// deliberately leaves open (see the rule's doc comment).
func TestNoSetterReturnFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		findings   int
	}{
		{"({ set a(val){ return val + 1; } })", 1},
		{"({ set a(val) { return 1; } })", 1},
		{"class A { set a(val) { return 1; } }", 1},
		{"class A { static set a(val) { return 1; } }", 1},
		{"(class { set a(val) { return 1; } })", 1},
		{"({ set a(val) { return val; } })", 1},
		{"class A { set a(val) { return undefined; } }", 1},
		{"(class { set a(val) { return null; } })", 1},
		{"({ set a(val) { return x + y; } })", 1},
		{"class A { set a(val) { return foo(); } }", 1},
		{"(class { set a(val) { return this._a; } })", 1},
		{"({ set a(val) { return this.a; } })", 1},
		{"({ set a(val) { if (foo) { return 1; }; } })", 1},
		{"class A { set a(val) { try { return 1; } catch(e) {} } }", 1},
		{"(class { set a(val) { while (foo){ if (bar) break; else return 1; } } })", 1},
		{"({ set a(val) { return 1; }, set b(val) { return 1; } })", 2},
		{"class A { set a(val) { return 1; } set b(val) { return 1; } }", 2},
		{"(class { set a(val) { return 1; } static set b(val) { return 1; } })", 2},
		{"({ set a(val) { if(val) { return 1; } else { return 2 }; } })", 2},
		{"class A { set a(val) { switch(val) { case 1: return x; case 2: return y; default: return z } } }", 3},
		{"(class { static set a(val) { if (val > 0) { this._val = val; return val; } return false; } })", 2},
		{"({ set a(val) { if(val) { return 1; } else { return; }; } })", 1},
		{"class A { set a(val) { switch(val) { case 1: return x; case 2: return; default: return z } } }", 2},
		{"(class { static set a(val) { if (val > 0) { this._val = val; return; } return false; } })", 1},
		{"({ set a(val) { function b(){} return b(); } })", 1},
		{"class A { set a(val) { return () => {}; } }", 1},
		{"(class { set a(val) { function b(){ return 1; } return 2; } })", 1},
		{"({ set a(val) { function b(){ return; } return 1; } })", 1},
		{"class A { set a(val) { var x = function() { return 1; }; return 2; } }", 1},
		{"(class { set a(val) { var x = () => { return; }; return 2; } })", 1},
		{"function f(){}; ({ set a(val) { return 1; } });", 1},
		{"x = function f(){}; class A { set a(val) { return 1; } };", 1},
		{"x = () => {}; A = class { set a(val) { return 1; } };", 1},
		{"return; ({ set a(val) { return 1; } }); return 2;", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, testCase.findings)
			for index := range wantIds {
				wantIds[index] = "noSetterReturn"
			}
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoSetterReturn, setterReturnFile, testCase.sourceText), wantIds...)
		})
	}
}

// The clean cases are the whole discrimination, and they outnumber the firing ones three to one.
//
// They split into families, and each family is a different way to be wrong. A rule keying on the
// word `set` reports `function set(val) { return 1; }` and `class set { constructor(val) { return
// 1; } }`. A rule that forgets the function boundary reports every nested function inside a setter.
// A rule that walks past a getter reports `({ get foo() { return 1; } })`. A rule keying on
// property-name text reports `({ set: function(val) { return 1; } })`, which is a plain property
// whose key is spelled `set`. And a rule that reads computed keys as setter bodies reports
// `({ set [function() { return 1; }](val) {} })`.
//
// The largest family, forty cases, is the `Object.defineProperty` descriptor group. Those are
// setters in every sense a reader cares about, and upstream knowingly does not catch them; the
// corresponding fail cases are commented out in the Rust source. They are here as clean fixtures
// because that is what upstream asserts, and reproducing a stated gap is the instruction rather
// than improving on it silently.
func TestNoSetterReturnStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"function foo() { return 1; }",
		"function set(val) { return 1; }",
		"var foo = function() { return 1; };",
		"var foo = function set() { return 1; };",
		"var set = function() { return 1; };",
		"var set = function set(val) { return 1; };",
		"var set = val => { return 1; };",
		"var set = val => 1;",
		"({ set a(val) { }}); function foo() { return 1; }",
		"({ set a(val) { }}); (function () { return 1; });",
		"({ set a(val) { }}); (() => { return 1; });",
		"({ set a(val) { }}); (() => 1);",
		"return 1;",
		"return 1;",
		"return 1; function foo(){ return 1; } return 1;",
		"function foo(){} return 1; var bar = function*(){ return 1; }; return 1; var baz = () => {}; return 1;",
		"({ set foo(val) { return; } })",
		"({ set foo(val) { if (val) { return; } } })",
		"class A { set foo(val) { return; } }",
		"(class { set foo(val) { if (val) { return; } else { return; } return; } })",
		"class A { set foo(val) { try {} catch(e) { return; } } }",
		"({ get foo() { return 1; } })",
		"({ get set() { return 1; } })",
		"({ set(val) { return 1; } })",
		"({ set: function(val) { return 1; } })",
		"({ foo: function set(val) { return 1; } })",
		"({ set: function set(val) { return 1; } })",
		"({ set: (val) => { return 1; } })",
		"({ set: (val) => 1 })",
		"set = { foo(val) { return 1; } };",
		"class A { constructor(val) { return 1; } }",
		"class set { constructor(val) { return 1; } }",
		"class set { foo(val) { return 1; } }",
		"var set = class { foo(val) { return 1; } }",
		"(class set { foo(val) { return 1; } })",
		"class A { get foo() { return val; } }",
		"class A { get set() { return val; } }",
		"class A { set(val) { return 1; } }",
		"class A { static set(val) { return 1; } }",
		"({ set: set = function set(val) { return 1; } } = {})",
		"({ set: set = (val) => 1 } = {})",
		"class C { set; foo() { return 1; } }",
		"({ set foo(val) { function foo(val) { return 1; } } })",
		"({ set foo(val) { var foo = function(val) { return 1; } } })",
		"({ set foo(val) { var foo = (val) => { return 1; } } })",
		"({ set foo(val) { var foo = (val) => 1; } })",
		"({ set [function() { return 1; }](val) {} })",
		"({ set [() => { return 1; }](val) {} })",
		"({ set [() => 1](val) {} })",
		"({ set foo(val = function() { return 1; }) {} })",
		"({ set foo(val = v => 1) {} })",
		"(class { set foo(val) { function foo(val) { return 1; } } })",
		"(class { set foo(val) { var foo = function(val) { return 1; } } })",
		"(class { set foo(val) { var foo = (val) => { return 1; } } })",
		"(class { set foo(val) { var foo = (val) => 1; } })",
		"(class { set [function() { return 1; }](val) {} })",
		"(class { set [() => { return 1; }](val) {} })",
		"(class { set [() => 1](val) {} })",
		"(class { set foo(val = function() { return 1; }) {} })",
		"(class { set foo(val = (v) => 1) {} })",
		"Object.defineProperty(foo, 'bar', { set(val) { return; } })",
		"Reflect.defineProperty(foo, 'bar', { set(val) { if (val) { return; } } })",
		"Object.defineProperties(foo, { bar: { set(val) { try { return; } catch(e){} } } })",
		"Object.create(foo, { bar: { set: function(val) { return; } } })",
		"x = { set(val) { return 1; } }",
		"x = { foo: { set(val) { return 1; } } }",
		"Object.defineProperty(foo, 'bar', { value(val) { return 1; } })",
		"Reflect.defineProperty(foo, 'bar', { value: function set(val) { return 1; } })",
		"Object.defineProperties(foo, { bar: { [set](val) { return 1; } } })",
		"Object.create(foo, { bar: { 'set ': function(val) { return 1; } } })",
		"Object.defineProperty(foo, 'bar', { [`set `]: (val) => { return 1; } })",
		"Reflect.defineProperty(foo, 'bar', { Set(val) { return 1; } })",
		"Object.defineProperties(foo, { bar: { value: (val) => 1 } })",
		"Object.create(foo, { set: { value: function(val) { return 1; } } })",
		"Object.defineProperty(foo, 'bar', { baz(val) { return 1; } })",
		"Reflect.defineProperty(foo, 'bar', { get(val) { return 1; } })",
		"Object.create(foo, { set: function(val) { return 1; } })",
		"Object.defineProperty(foo, { set: (val) => 1 })",
		"Object.defineProperty(foo, 'bar', { set(val) { function foo() { return 1; } } })",
		"Reflect.defineProperty(foo, 'bar', { set(val) { var foo = function() { return 1; } } })",
		"Object.defineProperties(foo, { bar: { set(val) { () => { return 1 }; } } })",
		"Object.create(foo, { bar: { set: (val) => { (val) => 1; } } })",
		"Object.defineProperty(foo, 'bar', 'baz', { set(val) { return 1; } })",
		"Object.defineProperty(foo, { set(val) { return 1; } }, 'bar')",
		"Object.defineProperty({ set(val) { return 1; } }, foo, 'bar')",
		"Reflect.defineProperty(foo, 'bar', 'baz', { set(val) { return 1; } })",
		"Reflect.defineProperty(foo, { set(val) { return 1; } }, 'bar')",
		"Reflect.defineProperty({ set(val) { return 1; } }, foo, 'bar')",
		"Object.defineProperties(foo, bar, { baz: { set(val) { return 1; } } })",
		"Object.defineProperties({ bar: { set(val) { return 1; } } }, foo)",
		"Object.create(foo, bar, { baz: { set(val) { return 1; } } })",
		"Object.create({ bar: { set(val) { return 1; } } }, foo)",
		"Object.DefineProperty(foo, 'bar', { set(val) { return 1; } })",
		"Reflect.DefineProperty(foo, 'bar', { set(val) { if (val) { return 1; } } })",
		"Object.DefineProperties(foo, { bar: { set(val) { try { return 1; } catch(e){} } } })",
		"Object.Create(foo, { bar: { set: function(val) { return 1; } } })",
		"object.defineProperty(foo, 'bar', { set(val) { return 1; } })",
		"reflect.defineProperty(foo, 'bar', { set(val) { if (val) { return 1; } } })",
		"Reflect.defineProperties(foo, { bar: { set(val) { try { return 1; } catch(e){} } } })",
		"object.create(foo, { bar: { set: function(val) { return 1; } } })",
		"Reflect.defineProperty(foo, 'bar', { set(val) { if (val) { return 1; } } })",
		"/* globals Object:off */ Object.defineProperty(foo, 'bar', { set(val) { return 1; } })",
		"Object.defineProperties(foo, { bar: { set(val) { try { return 1; } catch(e){} } } })",
		"let Object; Object.defineProperty(foo, 'bar', { set(val) { return 1; } })",
		"function f() { Reflect.defineProperty(foo, 'bar', { set(val) { if (val) { return 1; } } }); var Reflect;}",
		"function f(Object) { Object.defineProperties(foo, { bar: { set(val) { try { return 1; } catch(e){} } } }) }",
		"if (x) { const Object = getObject(); Object.create(foo, { bar: { set: function(val) { return 1; } } }) }",
		"x = function Object() { Object.defineProperty(foo, 'bar', { set(val) { return 1; } }) }",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoSetterReturn, setterReturnFile, sourceText))
		})
	}
}

// The span, which the message-id fixtures above cannot see.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule pointing at the wrong bytes
// passes the whole corpus while being wrong about the only thing a reader looks at. That is not
// hypothetical here: a mutation reporting `statement.Expression` instead of the statement compiled
// and changed none of the 142 cases above.
//
// The expected text is lifted from upstream's snapshot rather than counted by hand. Each of oxc's
// diagnostics underlines a range, and decoding those underlines gives `return val + 1;` for the
// first case below, which is the whole ReturnStatement including its semicolon rather than the
// expression inside it. The third case is the one input that reports three times, and the last of
// its three has no semicolon in the source, so the statement ends at the expression and the span
// shortens with it. A rule spanning the expression would drop the leading `return ` from all six.
func TestNoSetterReturnSpansTheWholeStatement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpans  []string
	}{
		{"({ set a(val){ return val + 1; } })", []string{"return val + 1;"}},
		{"({ set a(val) { if(val) { return 1; } else { return 2 }; } })",
			[]string{"return 1;", "return 2"}},
		{"class A { set a(val) { switch(val) { case 1: return x; case 2: return y; default: return z } } }",
			[]string{"return x;", "return y;", "return z"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoSetterReturn, setterReturnFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d diagnostics, got %d",
					len(testCase.wantSpans), len(result.Diagnostics))
			}
			for index, wantSpan := range testCase.wantSpans {
				diagnostic := result.Diagnostics[index]
				gotSpan := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != wantSpan {
					t.Fatalf("finding %d spanned %q, wanted %q", index, gotSpan, wantSpan)
				}
			}
		})
	}
}
