package core

import (
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// preferArrowCallbackFile is where the fixtures pretend to live.
const preferArrowCallbackFile = "/repository/source/PreferArrowCallback.ts"

// preferArrowCallbackOf builds the settings a config carrying this option would decode to.
//
// The argument is the option as JSON TEXT, exactly the bytes the config layer hands the decoder.
func preferArrowCallbackOf(optionJson string) any {
	settings, err := DecodePreferArrowCallbackOptions([]byte(optionJson))
	if err != nil {
		panic(err)
	}
	return settings
}

// The corpus is ESLint's own, taken from
// /tmp/lint-sources-fresh/eslint/tests/lib/rules/prefer-arrow-callback.js by RUNNING that file with RuleTester
// intercepted. 111 cases from TWO RuleTester.run calls, the second of which supplies @typescript-eslint/parser.
//
// Every expectation is what the INSTALLED ESLint 10.8.1 rule answered, driven through the Linter
// interface under `sourceType: module` with the typescript-eslint parser. The oracle reproduced all
// 111 declared verdicts AND all 72 declared outputs, which is the control saying the fixer half of
// the harness measures the fixer rather than nothing. Zero cases change verdict between upstream's
// configuration and cohere's.
//
// # The repair columns are the point of this table
//
// Section 3b of the standard is blunt about why: a fixer that reports in the right place and
// repairs wrongly is indistinguishable, at the message-id layer, from a correct one. So a corpus
// `output` is not optional coverage, it is the ONLY coverage the fixer has.
//
// Upstream's corpus states three different things and they need three different assertions:
//
//	no output declared    the corpus pinned nothing, so only the findings are asserted
//	output: null          upstream REPORTS and deliberately declines to repair, so the
//	                      assertion is that this rule proposes no fix either
//	output: "<source>"    ExpectFixedSource compares the whole rewritten file
//
// Collapsing the middle case into "the output equals the input" does NOT work and the harness is
// right to refuse it: `ExpectFixedSource` fatals when a rule proposed no fixes, because passing
// there would make it agree with any expectation at all. A decline has to be asserted as an
// absence of fixes, not as an unchanged string.
//
//	72 reporting cases (14 of them declines), 39 clean cases.
func TestPreferArrowCallbackFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		findings int
		// fixed is the source after one fix pass, which is what upstream's `output` records.
		fixed *string
		// declinesRepair marks upstream's `output: null`: reported, deliberately not fixed.
		declinesRepair bool
	}{
		{source: "foo(function bar() {});", settings: nil, findings: 1, fixed: pointerTo("foo(() => {});"), declinesRepair: false},
		{source: "foo(function() {});", settings: preferArrowCallbackOf("{\"allowNamedFunctions\": true}"), findings: 1, fixed: pointerTo("foo(() => {});"), declinesRepair: false},
		{source: "foo(function bar() {});", settings: preferArrowCallbackOf("{\"allowNamedFunctions\": false}"), findings: 1, fixed: pointerTo("foo(() => {});"), declinesRepair: false},
		{source: "foo(function() {});", settings: nil, findings: 1, fixed: pointerTo("foo(() => {});"), declinesRepair: false},
		{source: "foo(nativeCb || function() {});", settings: nil, findings: 1, fixed: pointerTo("foo(nativeCb || (() => {}));"), declinesRepair: false},
		{source: "foo(bar ? function() {} : function() {});", settings: nil, findings: 2, fixed: pointerTo("foo(bar ? () => {} : () => {});"), declinesRepair: false},
		{source: "foo(function() { (function() { this; }); });", settings: nil, findings: 1, fixed: pointerTo("foo(() => { (function() { this; }); });"), declinesRepair: false},
		{source: "foo(function() { this; }.bind(this));", settings: nil, findings: 1, fixed: pointerTo("foo(() => { this; });"), declinesRepair: false},
		{source: "foo(bar || function() { this; }.bind(this));", settings: nil, findings: 1, fixed: pointerTo("foo(bar || (() => { this; }));"), declinesRepair: false},
		{source: "foo(function() { (() => this); }.bind(this));", settings: nil, findings: 1, fixed: pointerTo("foo(() => { (() => this); });"), declinesRepair: false},
		{source: "foo(function bar(a) { a; });", settings: nil, findings: 1, fixed: pointerTo("foo((a) => { a; });"), declinesRepair: false},
		{source: "foo(function(a) { a; });", settings: nil, findings: 1, fixed: pointerTo("foo((a) => { a; });"), declinesRepair: false},
		{source: "foo(function(arguments) { arguments; });", settings: nil, findings: 1, fixed: pointerTo("foo((arguments) => { arguments; });"), declinesRepair: false},
		{source: "foo(function() { this; });", settings: preferArrowCallbackOf("{\"allowUnboundThis\": false}"), findings: 1, fixed: nil, declinesRepair: true},
		{source: "foo(function() { (() => this); });", settings: preferArrowCallbackOf("{\"allowUnboundThis\": false}"), findings: 1, fixed: nil, declinesRepair: true},
		{source: "qux(function(foo, bar, baz) { return foo * 2; })", settings: nil, findings: 1, fixed: pointerTo("qux((foo, bar, baz) => { return foo * 2; })"), declinesRepair: false},
		{source: "qux(function(foo, bar, baz) { return foo * bar; }.bind(this))", settings: nil, findings: 1, fixed: pointerTo("qux((foo, bar, baz) => { return foo * bar; })"), declinesRepair: false},
		{source: "qux(function(foo, bar, baz) { return foo * this.qux; }.bind(this))", settings: nil, findings: 1, fixed: pointerTo("qux((foo, bar, baz) => { return foo * this.qux; })"), declinesRepair: false},
		{source: "foo(function() {}.bind(this, somethingElse))", settings: nil, findings: 1, fixed: pointerTo("foo((() => {}).bind(this, somethingElse))"), declinesRepair: false},
		{source: "qux(function(foo = 1, [bar = 2] = [], {qux: baz = 3} = {foo: 'bar'}) { return foo + bar; });", settings: nil, findings: 1, fixed: pointerTo("qux((foo = 1, [bar = 2] = [], {qux: baz = 3} = {foo: 'bar'}) => { return foo + bar; });"), declinesRepair: false},
		{source: "qux(function(baz, baz) { })", settings: nil, findings: 1, fixed: nil, declinesRepair: true},
		{source: "qux(function( /* no params */ ) { })", settings: nil, findings: 1, fixed: pointerTo("qux(( /* no params */ ) => { })"), declinesRepair: false},
		{source: "qux(function( /* a */ foo /* b */ , /* c */ bar /* d */ , /* e */ baz /* f */ ) { return foo; })", settings: nil, findings: 1, fixed: pointerTo("qux(( /* a */ foo /* b */ , /* c */ bar /* d */ , /* e */ baz /* f */ ) => { return foo; })"), declinesRepair: false},
		{source: "qux(async function (foo = 1, bar = 2, baz = 3) { return baz; })", settings: nil, findings: 1, fixed: pointerTo("qux(async (foo = 1, bar = 2, baz = 3) => { return baz; })"), declinesRepair: false},
		{source: "qux(async function (foo = 1, bar = 2, baz = 3) { return this; }.bind(this))", settings: nil, findings: 1, fixed: pointerTo("qux(async (foo = 1, bar = 2, baz = 3) => { return this; })"), declinesRepair: false},
		{source: "foo(async function /*\n*/ () { return 1; });", settings: nil, findings: 1, fixed: nil, declinesRepair: true},
		{source: "foo(async function // c\n () { return 1; });", settings: nil, findings: 1, fixed: nil, declinesRepair: true},
		{source: "foo(async function\n () { return 1; });", settings: nil, findings: 1, fixed: nil, declinesRepair: true},
		{source: "foo(async function /* c */ () { return 1; });", settings: nil, findings: 1, fixed: pointerTo("foo(async  /* c */ () => { return 1; });"), declinesRepair: false},
		{source: "foo((bar || function() {}).bind(this))", settings: nil, findings: 1, fixed: nil, declinesRepair: true},
		{source: "foo(function() {}.bind(this).bind(obj))", settings: nil, findings: 1, fixed: pointerTo("foo((() => {}).bind(obj))"), declinesRepair: false},
		{source: "foo?.(function() {});", settings: nil, findings: 1, fixed: pointerTo("foo?.(() => {});"), declinesRepair: false},
		{source: "foo?.(function() { return this; }.bind(this));", settings: nil, findings: 1, fixed: pointerTo("foo?.(() => { return this; });"), declinesRepair: false},
		{source: "foo(function() { return this; }?.bind(this));", settings: nil, findings: 1, fixed: pointerTo("foo(() => { return this; });"), declinesRepair: false},
		{source: "foo((function() { return this; }?.bind)(this));", settings: nil, findings: 1, fixed: nil, declinesRepair: true},
		{source: "\n            test(\n                function ()\n                { }\n            );\n            ", settings: nil, findings: 1, fixed: pointerTo("\n            test(\n                () =>\n                { }\n            );\n            "), declinesRepair: false},
		{source: "\n            test(\n                function (\n                    ...args\n                ) /* Lorem ipsum\n                dolor sit amet. */ {\n                    return args;\n                }\n            );\n            ", settings: nil, findings: 1, fixed: pointerTo("\n            test(\n                (\n                    ...args\n                ) => /* Lorem ipsum\n                dolor sit amet. */ {\n                    return args;\n                }\n            );\n            "), declinesRepair: false},
		{source: "foo(function bar() {});", settings: nil, findings: 1, fixed: pointerTo("foo(() => {});"), declinesRepair: false},
		{source: "foo(function(a:string) {});", settings: preferArrowCallbackOf("{\"allowNamedFunctions\": true}"), findings: 1, fixed: pointerTo("foo((a:string) => {});"), declinesRepair: false},
		{source: "foo(function bar() {});", settings: preferArrowCallbackOf("{\"allowNamedFunctions\": false}"), findings: 1, fixed: pointerTo("foo(() => {});"), declinesRepair: false},
		{source: "foo(function() {});", settings: nil, findings: 1, fixed: pointerTo("foo(() => {});"), declinesRepair: false},
		{source: "foo(nativeCb || function() {});", settings: nil, findings: 1, fixed: pointerTo("foo(nativeCb || (() => {}));"), declinesRepair: false},
		{source: "foo(bar ? function() {} : function() {});", settings: nil, findings: 2, fixed: pointerTo("foo(bar ? () => {} : () => {});"), declinesRepair: false},
		{source: "foo(function() { (function() { this; }); });", settings: nil, findings: 1, fixed: pointerTo("foo(() => { (function() { this; }); });"), declinesRepair: false},
		{source: "foo(function() { this; }.bind(this));", settings: nil, findings: 1, fixed: pointerTo("foo(() => { this; });"), declinesRepair: false},
		{source: "foo(bar || function() { this; }.bind(this));", settings: nil, findings: 1, fixed: pointerTo("foo(bar || (() => { this; }));"), declinesRepair: false},
		{source: "foo(function() { (() => this); }.bind(this));", settings: nil, findings: 1, fixed: pointerTo("foo(() => { (() => this); });"), declinesRepair: false},
		{source: "foo(function bar(a:string) { a; });", settings: nil, findings: 1, fixed: pointerTo("foo((a:string) => { a; });"), declinesRepair: false},
		{source: "foo(function(a:any) { a; });", settings: nil, findings: 1, fixed: pointerTo("foo((a:any) => { a; });"), declinesRepair: false},
		{source: "foo(function(arguments:any) { arguments; });", settings: nil, findings: 1, fixed: pointerTo("foo((arguments:any) => { arguments; });"), declinesRepair: false},
		{source: "foo(function(a:string) { this; });", settings: preferArrowCallbackOf("{\"allowUnboundThis\": false}"), findings: 1, fixed: nil, declinesRepair: true},
		{source: "foo(function() { (() => this); });", settings: preferArrowCallbackOf("{\"allowUnboundThis\": false}"), findings: 1, fixed: nil, declinesRepair: true},
		{source: "qux(function(foo:string, bar:number, baz:string) { return foo * 2; })", settings: nil, findings: 1, fixed: pointerTo("qux((foo:string, bar:number, baz:string) => { return foo * 2; })"), declinesRepair: false},
		{source: "qux(function(foo:number, bar:number, baz:number) { return foo * bar; }.bind(this))", settings: nil, findings: 1, fixed: pointerTo("qux((foo:number, bar:number, baz:number) => { return foo * bar; })"), declinesRepair: false},
		{source: "qux(function(foo:any, bar:any, baz:any) { return foo * this.qux; }.bind(this))", settings: nil, findings: 1, fixed: pointerTo("qux((foo:any, bar:any, baz:any) => { return foo * this.qux; })"), declinesRepair: false},
		{source: "foo(function() {}.bind(this, somethingElse))", settings: nil, findings: 1, fixed: pointerTo("foo((() => {}).bind(this, somethingElse))"), declinesRepair: false},
		{source: "qux(function(foo = 1, [bar = 2] = [], {qux: baz = 3} = {foo: 'bar'}) { return foo + bar; });", settings: nil, findings: 1, fixed: pointerTo("qux((foo = 1, [bar = 2] = [], {qux: baz = 3} = {foo: 'bar'}) => { return foo + bar; });"), declinesRepair: false},
		{source: "qux(function(baz:string, baz:string) { })", settings: nil, findings: 1, fixed: nil, declinesRepair: true},
		{source: "qux(function( /* no params */ ) { })", settings: nil, findings: 1, fixed: pointerTo("qux(( /* no params */ ) => { })"), declinesRepair: false},
		{source: "qux(function( /* a */ foo:string /* b */ , /* c */ bar:string /* d */ , /* e */ baz:string /* f */ ) { return foo; })", settings: nil, findings: 1, fixed: pointerTo("qux(( /* a */ foo:string /* b */ , /* c */ bar:string /* d */ , /* e */ baz:string /* f */ ) => { return foo; })"), declinesRepair: false},
		{source: "qux(async function (foo:number = 1, bar:number = 2, baz:number = 3) { return baz; })", settings: nil, findings: 1, fixed: pointerTo("qux(async (foo:number = 1, bar:number = 2, baz:number = 3) => { return baz; })"), declinesRepair: false},
		{source: "qux(async function (foo:number = 1, bar:number = 2, baz:number = 3) { return this; }.bind(this))", settings: nil, findings: 1, fixed: pointerTo("qux(async (foo:number = 1, bar:number = 2, baz:number = 3) => { return this; })"), declinesRepair: false},
		{source: "foo((bar || function() {}).bind(this))", settings: nil, findings: 1, fixed: nil, declinesRepair: true},
		{source: "foo(function() {}.bind(this).bind(obj))", settings: nil, findings: 1, fixed: pointerTo("foo((() => {}).bind(obj))"), declinesRepair: false},
		{source: "foo?.(function() {});", settings: nil, findings: 1, fixed: pointerTo("foo?.(() => {});"), declinesRepair: false},
		{source: "foo?.(function() { return this; }.bind(this));", settings: nil, findings: 1, fixed: pointerTo("foo?.(() => { return this; });"), declinesRepair: false},
		{source: "foo(function() { return this; }?.bind(this));", settings: nil, findings: 1, fixed: pointerTo("foo(() => { return this; });"), declinesRepair: false},
		{source: "foo((function() { return this; }?.bind)(this));", settings: nil, findings: 1, fixed: nil, declinesRepair: true},
		{source: "\n            test(\n                function ()\n                { }\n            );\n            ", settings: nil, findings: 1, fixed: pointerTo("\n            test(\n                () =>\n                { }\n            );\n            "), declinesRepair: false},
		{source: "\n            test(\n                function (\n                    ...args\n                ) /* Lorem ipsum\n                dolor sit amet. */ {\n                    return args;\n                }\n            );\n            ", settings: nil, findings: 1, fixed: pointerTo("\n            test(\n                (\n                    ...args\n                ) => /* Lorem ipsum\n                dolor sit amet. */ {\n                    return args;\n                }\n            );\n            "), declinesRepair: false},
		{source: "foo(function():string { return 'foo' });", settings: nil, findings: 1, fixed: pointerTo("foo(():string => { return 'foo' });"), declinesRepair: false},
		{source: "test('foo', function (this: any) {});", settings: nil, findings: 1, fixed: nil, declinesRepair: true},
	}

	for _, testCase := range cases {
		result := rule_testing.RunTypedWithOptions(t, PreferArrowCallback, preferArrowCallbackFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != testCase.findings {
			t.Errorf("%q with %+v: expected %d findings, got %d %v",
				testCase.source, testCase.settings, testCase.findings,
				len(result.Diagnostics), result.MessageIds())
			continue
		}
		expected := make([]string, 0, testCase.findings)
		for range testCase.findings {
			expected = append(expected, messagePreferArrowCallbackPreferArrowCallback.Id)
		}
		rule_testing.ExpectFindings(t, result, expected...)

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
			// `RunTyped` writes each fixture as `TrimSpace(contents)+"\n"`, so the file on disk
			// carries a trailing newline the corpus string does not. The expectation is normalised
			// the same way rather than baking a "\n" into 72 literals, which would make them stop
			// matching the corpus and defeat the byte verification.
			rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(*testCase.fixed)+"\n")
		}
	}
}

// TestPreferArrowCallbackStaysSilent carries upstream's clean cases.
func TestPreferArrowCallbackStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
	}{
		{source: "foo(a => a);", settings: nil},
		{source: "foo(function*() {});", settings: nil},
		{source: "foo(function() { this; });", settings: nil},
		{source: "foo(function bar() {});", settings: preferArrowCallbackOf("{\"allowNamedFunctions\": true}")},
		{source: "foo(function() { (() => this); });", settings: nil},
		{source: "foo(function() { this; }.bind(obj));", settings: nil},
		{source: "foo(function() { this; }.call(this));", settings: nil},
		{source: "foo(a => { (function() {}); });", settings: nil},
		{source: "var foo = function foo() {};", settings: nil},
		{source: "(function foo() {})();", settings: nil},
		{source: "foo(function bar() { bar; });", settings: nil},
		{source: "foo(function bar() { arguments; });", settings: nil},
		{source: "foo(function bar() { arguments; }.bind(this));", settings: nil},
		{source: "foo(function bar() { new.target; });", settings: nil},
		{source: "foo(function bar() { new.target; }.bind(this));", settings: nil},
		{source: "foo(function bar() { this; }.bind(this, somethingElse));", settings: nil},
		{source: "foo((function() {}).bind.bar)", settings: nil},
		{source: "foo((function() { this.bar(); }).bind(obj).bind(this))", settings: nil},
		{source: "foo(a => a);", settings: nil},
		{source: "foo((a:string) => a);", settings: nil},
		{source: "foo(function*() {});", settings: nil},
		{source: "foo(function() { this; });", settings: nil},
		{source: "foo(function bar(a:string) {});", settings: preferArrowCallbackOf("{\"allowNamedFunctions\": true}")},
		{source: "foo(function() { (() => this); });", settings: nil},
		{source: "foo(function() { this; }.bind(obj));", settings: nil},
		{source: "foo(function() { this; }.call(this));", settings: nil},
		{source: "foo(a => { (function() {}); });", settings: nil},
		{source: "var foo = function foo() {};", settings: nil},
		{source: "(function foo() {})();", settings: nil},
		{source: "foo(function bar() { bar; });", settings: nil},
		{source: "foo(function bar() { arguments; });", settings: nil},
		{source: "foo(function bar() { arguments; }.bind(this));", settings: nil},
		{source: "foo(function bar() { new.target; });", settings: nil},
		{source: "foo(function bar() { new.target; }.bind(this));", settings: nil},
		{source: "foo(function bar() { this; }.bind(this, somethingElse));", settings: nil},
		{source: "foo((function() {}).bind.bar)", settings: nil},
		{source: "foo((function() { this.bar(); }).bind(obj).bind(this))", settings: nil},
		{source: "test('clean', function (this: any) { this.foo = 'Cleaned!';});", settings: nil},
		{source: "obj.test('clean', function (foo) { this.foo = 'Cleaned!'; });", settings: nil},
	}

	for _, testCase := range cases {
		result := rule_testing.RunTypedWithOptions(t, PreferArrowCallback, preferArrowCallbackFile, testCase.source, testCase.settings)
		rule_testing.ExpectClean(t, result)
	}
}

// TestPreferArrowCallbackSelfReferenceIsResolvedNotMatched covers what upstream's corpus cannot.
//
// A function expression that calls itself by name needs that binding, and an arrow does not provide
// one, so upstream exempts it. The question is whether the reference actually RESOLVES to the
// function's own name or merely spells it the same way, and a mutation replacing the resolution
// with a text comparison survived all 111 imported fixtures: the corpus contains no case where a
// nested declaration shadows the function's name, so nothing in it separates the two.
//
// Every verdict is what the installed ESLint 10.8.1 rule answered, and the probe that established
// the substrate gives the matching symbol counts:
//
//	foo(function bar() { bar(); });                     clean,  1 resolved self-reference
//	foo(function bar() { function bar() {} bar(); });   REPORTS, 0 -- the inner declaration shadows
//	foo(function bar() { const bar = 1; bar; });        REPORTS, 0 -- the const shadows
//	foo(function bar() { baz(); });                     REPORTS, 0 -- no self-reference at all
//
// The first row is the control: without it a rule that never exempted anything would pass the other
// three and prove nothing.
//
// The same class covers `arguments`, which is resolved differently because it has no declaration to
// resolve TO. A reference belongs to the nearest enclosing function, and a PARAMETER named
// `arguments` is a real binding that shadows the implicit one:
//
//	foo(function() { arguments; });                       clean,   the implicit binding is used
//	foo(function() { function inner() { arguments; } });  REPORTS, it belongs to `inner`
//	foo(function(arguments) { arguments; });              REPORTS, the parameter shadows it
func TestPreferArrowCallbackSelfReferenceIsResolvedNotMatched(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		findings int
	}{
		{source: "foo(function bar() { bar(); });", findings: 0},
		{source: "foo(function bar() { function bar() {} bar(); });", findings: 1},
		{source: "foo(function bar() { const bar = 1; bar; });", findings: 1},
		{source: "foo(function bar() { baz(); });", findings: 1},

		{source: "foo(function() { arguments; });", findings: 0},
		{source: "foo(function() { function inner() { arguments; } });", findings: 1},
		{source: "foo(function(arguments) { arguments; });", findings: 1},
	}

	for _, testCase := range cases {
		result := rule_testing.RunTypedWithOptions(t, PreferArrowCallback, preferArrowCallbackFile,
			testCase.source, nil)
		if len(result.Diagnostics) != testCase.findings {
			t.Errorf("%q: expected %d findings, got %d %v",
				testCase.source, testCase.findings, len(result.Diagnostics), result.MessageIds())
		}
	}
}

// TestPreferArrowCallbackHandlesNewExpressions is a regression test for a crash, and it is the one
// case in this file that a fixture could never have predicted.
//
// The callback walk has an arm for `KindCallExpression` and `KindNewExpression`. A first draft
// shared one body and called `parent.AsCallExpression()` before testing which kind it had, which is
// an interface conversion that PANICS on a NewExpression rather than returning nil. It crashed
// **71 files** on the real ahra tree, and the fixture suite was completely green:
//
//	interface conversion: ast.nodeData is *ast.NewExpression, not *ast.CallExpression
//
// Two things made it invisible. Upstream's corpus writes no `new Foo(function() {})` case at all,
// so nothing imported reaches the arm. And the walk recovers per FILE rather than per rule, so one
// panic silently costs every registered rule its verdict on that file: the same run reported
// 2,029,962 nodes visited where a clean one reports 2,184,766, and still printed a findings summary
// that looked ordinary.
//
// Only a dry run against real source finds this class. Every verdict below is what the installed
// ESLint 10.8.1 rule answered:
//
//	new Foo(function() {});                        1 finding, fixed to `new Foo(() => {});`
//	new Foo(function() { this; });                 clean, allowUnboundThis
//	new (function() {})();                         clean, the function is the CALLEE
//	new Promise(function(resolve) {...});          1 finding, fixed
//
// The third row is the one that pins the arm's actual judgment rather than just its safety: a
// function in callee position is not a callback, and getting that backwards would report every
// immediately-constructed function in the tree.
func TestPreferArrowCallbackHandlesNewExpressions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		findings int
		fixed    *string
	}{
		{source: "new Foo(function() {});", findings: 1, fixed: pointerTo("new Foo(() => {});")},
		{source: "new Foo(function() { this; });", findings: 0},
		{source: "new (function() {})();", findings: 0},
		{
			source:   "const x = new Promise(function(resolve) { resolve(); });",
			findings: 1,
			fixed:    pointerTo("const x = new Promise((resolve) => { resolve(); });"),
		},
	}

	for _, testCase := range cases {
		result := rule_testing.RunTypedWithOptions(t, PreferArrowCallback, preferArrowCallbackFile,
			testCase.source, nil)
		if len(result.Diagnostics) != testCase.findings {
			t.Errorf("%q: expected %d findings, got %d %v",
				testCase.source, testCase.findings, len(result.Diagnostics), result.MessageIds())
			continue
		}
		if testCase.fixed != nil {
			rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(*testCase.fixed)+"\n")
		}
	}
}
