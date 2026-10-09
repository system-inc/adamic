package core

import (
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// oneVarFile is where the fixtures pretend to live.
//
// A `.ts` name, because that is what cohere lints and because the oracle replayed every case
// under the TypeScript parser with `sourceType: module` before any of this was written. Nothing
// in the corpus is JSX.
const oneVarFile = "/repository/source/OneVar.ts"

// oneVarCase is one row of upstream's corpus.
type oneVarCase struct {
	// name is the corpus list and index the row came from, so a failure names a case findable
	// in upstream's own file rather than a number local to this table.
	name string

	// source is upstream's `code`, byte for byte.
	source string

	// options is the RAW JSON of upstream's single option value, routed through the rule's own
	// exported decoder rather than built as a struct. That is what puts the union decoding, the
	// absent-versus-set distinction and the default under test; a struct built by hand reaches
	// the rule having skipped every line that could be wrong.
	options string

	// ids are the message ids upstream produced, in order.
	ids []string

	// messages are the RENDERED message texts upstream produced, in order.
	//
	// Six of the seven messages interpolate `{{type}}`, and a message id cannot see a rendering.
	// A rule that reports the right id having computed "var" where upstream computed "await
	// using" passes every id assertion, so the text is asserted alongside.
	messages []string
}

// oneVarFixCase is one of upstream's `output` fixtures.
type oneVarFixCase struct {
	name    string
	source  string
	options string
	wanted  string
}

// decodedOneVar routes a row's raw option JSON through the rule's own decoder.
//
// An empty string means the config carried no options at all, which is what a bare `"error"`
// delivers, and the decoder answers with upstream's `defaultOptions: ["always"]`.
func decodedOneVar(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeOneVarOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding one-var options %q: %v", raw, err)
	}
	return decoded
}

// oneVarFiresCases are the rows upstream reports on.
var oneVarFiresCases = []oneVarCase{
	{name: "invalid-152", source: "var bar = true, baz = false;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-153", source: "function foo() { var bar = true, baz = false; }", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-154", source: "if (foo) { var bar = true, baz = false; }", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-155", source: "switch (foo) { case bar: var baz = true, quux = false; }", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-156", source: "switch (foo) { default: var baz = true, quux = false; }", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-157", source: "function foo() { var bar = true; var baz = false; }", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-158", source: "var a = 1; for (var b = 2;;) {}", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-159", source: "function foo() { var foo = true, bar = false; }", options: "{\"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'var' declarations into multiple statements."}},
	{name: "invalid-160", source: "function foo() { var foo, bar; }", options: "{\"uninitialized\": \"never\"}",
		ids:      []string{"splitUninitialized"},
		messages: []string{"Split uninitialized 'var' declarations into multiple statements."}},
	{name: "invalid-161", source: "function foo() { var bar, baz; var a = true; var b = false; var c, d;}", options: "{\"uninitialized\": \"always\", \"initialized\": \"never\"}",
		ids:      []string{"combineUninitialized"},
		messages: []string{"Combine this with the previous 'var' statement with uninitialized variables."}},
	{name: "invalid-162", source: "function foo() { var bar = true, baz = false; var a; var b; var c = true, d = false; }", options: "{\"uninitialized\": \"never\", \"initialized\": \"always\"}",
		ids:      []string{"combineInitialized"},
		messages: []string{"Combine this with the previous 'var' statement with initialized variables."}},
	{name: "invalid-163", source: "function foo() { var bar = true, baz = false; var a, b;}", options: "{\"uninitialized\": \"never\", \"initialized\": \"never\"}",
		ids:      []string{"split", "split"},
		messages: []string{"Split 'var' declarations into multiple statements.", "Split 'var' declarations into multiple statements."}},
	{name: "invalid-164", source: "function foo() { var bar = true; var baz = false; var a; var b;}", options: "{\"uninitialized\": \"always\", \"initialized\": \"always\"}",
		ids:      []string{"combine", "combine", "combine"},
		messages: []string{"Combine this with the previous 'var' statement.", "Combine this with the previous 'var' statement.", "Combine this with the previous 'var' statement."}},
	{name: "invalid-165", source: "function foo() { var a = [1, 2, 3]; var [b, c, d] = a; }", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-166", source: "function foo() { let a = 1; let b = 2; }", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "invalid-167", source: "function foo() { const a = 1; const b = 2; }", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'const' statement."}},
	{name: "invalid-168", source: "function foo() { let a = 1; let b = 2; }", options: "{\"let\": \"always\"}",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "invalid-169", source: "function foo() { const a = 1; const b = 2; }", options: "{\"const\": \"always\"}",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'const' statement."}},
	{name: "invalid-170", source: "function foo() { let a = 1, b = 2; }", options: "{\"let\": \"never\"}",
		ids:      []string{"split"},
		messages: []string{"Split 'let' declarations into multiple statements."}},
	{name: "invalid-171", source: "function foo() { let a = 1, b = 2; }", options: "{\"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'let' declarations into multiple statements."}},
	{name: "invalid-172", source: "function foo() { let a, b; }", options: "{\"uninitialized\": \"never\"}",
		ids:      []string{"splitUninitialized"},
		messages: []string{"Split uninitialized 'let' declarations into multiple statements."}},
	{name: "invalid-173", source: "function foo() { const a = 1, b = 2; }", options: "{\"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'const' declarations into multiple statements."}},
	{name: "invalid-174", source: "function foo() { const a = 1, b = 2; }", options: "{\"const\": \"never\"}",
		ids:      []string{"split"},
		messages: []string{"Split 'const' declarations into multiple statements."}},
	{name: "invalid-175", source: "let foo = true; switch(foo) { case true: let bar = 2; break; case false: let baz = 3; break; }", options: "{\"var\": \"always\", \"let\": \"always\", \"const\": \"never\"}",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "invalid-176", source: "var one = 1, two = 2;\nvar three;", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-177", source: "var i = [0], j;", options: "{\"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'var' declarations into multiple statements."}},
	{name: "invalid-178", source: "var i = [0], j;", options: "{\"uninitialized\": \"never\"}",
		ids:      []string{"splitUninitialized"},
		messages: []string{"Split uninitialized 'var' declarations into multiple statements."}},
	{name: "invalid-179", source: "for (var x of foo) {}; for (var y of foo) {}", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-180", source: "for (var x in foo) {}; for (var y in foo) {}", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-181", source: "var foo = function() { var bar = true; var baz = false; }", options: "",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-182", source: "function foo() { var bar = true; if (qux) { var baz = false; } else { var quxx = 42; } }", options: "",
		ids:      []string{"combine", "combine"},
		messages: []string{"Combine this with the previous 'var' statement.", "Combine this with the previous 'var' statement."}},
	{name: "invalid-183", source: "var foo = () => { var bar = true; var baz = false; }", options: "",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-184", source: "var foo = function() { var bar = true; if (qux) { var baz = false; } }", options: "",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-185", source: "var foo; var bar;", options: "",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-186", source: "var x = 1, y = 2; for (var z in foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'var' declarations into multiple statements."}},
	{name: "invalid-187", source: "var x = 1, y = 2; for (var z of foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'var' declarations into multiple statements."}},
	{name: "invalid-188", source: "var x; var y; for (var z in foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		ids:      []string{"combineUninitialized"},
		messages: []string{"Combine this with the previous 'var' statement with uninitialized variables."}},
	{name: "invalid-189", source: "var x; var y; for (var z of foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		ids:      []string{"combineUninitialized"},
		messages: []string{"Combine this with the previous 'var' statement with uninitialized variables."}},
	{name: "invalid-190", source: "var x; for (var y in foo) {var bar = y; var a; for (var z of bar) {}}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		ids:      []string{"combineUninitialized"},
		messages: []string{"Combine this with the previous 'var' statement with uninitialized variables."}},
	{name: "invalid-191", source: "var a = 1; var b = 2; var x, y; for (var z of foo) {var c = 3, baz = z; for (var d in baz) {}}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'var' declarations into multiple statements."}},
	{name: "invalid-192", source: "var {foo} = 1, [bar] = 2;", options: "{\"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'var' declarations into multiple statements."}},
	{name: "invalid-193", source: "const foo = 1,\n    bar = 2;", options: "{\"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'const' declarations into multiple statements."}},
	{name: "invalid-194", source: "var foo = 1,\n    bar = 2;", options: "{\"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'var' declarations into multiple statements."}},
	{name: "invalid-195", source: "var foo = 1, // comment\n    bar = 2;", options: "{\"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'var' declarations into multiple statements."}},
	{name: "invalid-196", source: "var f, k /* test */, l;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-197", source: "var f,          /* test */ l;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-198", source: "var f, k /* test \n some more comment \n even more */, l = 1, P;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-199", source: "var a = 1, b = 2", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-200", source: "var foo = require('foo'), bar;", options: "{\"separateRequires\": true, \"var\": \"always\"}",
		ids:      []string{"splitRequires"},
		messages: []string{"Split requires to be separated into a single block."}},
	{name: "invalid-201", source: "var foo, bar = require('bar');", options: "{\"separateRequires\": true, \"var\": \"always\"}",
		ids:      []string{"splitRequires"},
		messages: []string{"Split requires to be separated into a single block."}},
	{name: "invalid-202", source: "let foo, bar = require('bar');", options: "{\"separateRequires\": true, \"let\": \"always\"}",
		ids:      []string{"splitRequires"},
		messages: []string{"Split requires to be separated into a single block."}},
	{name: "invalid-203", source: "const foo = 0, bar = require('bar');", options: "{\"separateRequires\": true, \"const\": \"always\"}",
		ids:      []string{"splitRequires"},
		messages: []string{"Split requires to be separated into a single block."}},
	{name: "invalid-204", source: "const foo = require('foo'); const bar = require('bar');", options: "{\"separateRequires\": true, \"const\": \"always\"}",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'const' statement."}},
	{name: "invalid-205", source: "var a = 1, b; var c;", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-206", source: "var a = 0, b = 1; var c = 2;", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-207", source: "let a = 1, b; let c;", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "invalid-208", source: "let a = 0, b = 1; let c = 2;", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "invalid-209", source: "const a = 0, b = 1; const c = 2;", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'const' statement."}},
	{name: "invalid-210", source: "const a = 0; var b = 1; var c = 2; const d = 3;", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-211", source: "var a = true; var b = false;", options: "{\"separateRequires\": true, \"var\": \"always\"}",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-212", source: "const a = 0; let b = 1; let c = 2; const d = 3;", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "invalid-213", source: "let a = 0; const b = 1; const c = 1; var d = 2;", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'const' statement."}},
	{name: "invalid-214", source: "var a = 0; var b; var c; var d = 1", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		ids:      []string{"combineUninitialized"},
		messages: []string{"Combine this with the previous 'var' statement with uninitialized variables."}},
	{name: "invalid-215", source: "var a = 0; var b = 1; var c; var d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		ids:      []string{"combineInitialized", "combineUninitialized"},
		messages: []string{"Combine this with the previous 'var' statement with initialized variables.", "Combine this with the previous 'var' statement with uninitialized variables."}},
	{name: "invalid-216", source: "let a = 0; let b; let c; let d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		ids:      []string{"combineUninitialized"},
		messages: []string{"Combine this with the previous 'let' statement with uninitialized variables."}},
	{name: "invalid-217", source: "let a = 0; let b = 1; let c; let d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		ids:      []string{"combineInitialized", "combineUninitialized"},
		messages: []string{"Combine this with the previous 'let' statement with initialized variables.", "Combine this with the previous 'let' statement with uninitialized variables."}},
	{name: "invalid-218", source: "const a = 0; let b; let c; const d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		ids:      []string{"combineUninitialized"},
		messages: []string{"Combine this with the previous 'let' statement with uninitialized variables."}},
	{name: "invalid-219", source: "const a = 0; const b = 1; let c; let d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		ids:      []string{"combineInitialized", "combineUninitialized"},
		messages: []string{"Combine this with the previous 'const' statement with initialized variables.", "Combine this with the previous 'let' statement with uninitialized variables."}},
	{name: "invalid-220", source: "var a = 0; var b = 1; var c, d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		ids:      []string{"combineInitialized", "splitUninitialized"},
		messages: []string{"Combine this with the previous 'var' statement with initialized variables.", "Split uninitialized 'var' declarations into multiple statements."}},
	{name: "invalid-221", source: "var a = 0; var b, c; var d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		ids:      []string{"splitUninitialized"},
		messages: []string{"Split uninitialized 'var' declarations into multiple statements."}},
	{name: "invalid-222", source: "let a = 0; let b = 1; let c, d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		ids:      []string{"combineInitialized", "splitUninitialized"},
		messages: []string{"Combine this with the previous 'let' statement with initialized variables.", "Split uninitialized 'let' declarations into multiple statements."}},
	{name: "invalid-223", source: "let a = 0; let b, c; let d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		ids:      []string{"splitUninitialized"},
		messages: []string{"Split uninitialized 'let' declarations into multiple statements."}},
	{name: "invalid-224", source: "const a = 0; const b = 1; let c, d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		ids:      []string{"combineInitialized", "splitUninitialized"},
		messages: []string{"Combine this with the previous 'const' statement with initialized variables.", "Split uninitialized 'let' declarations into multiple statements."}},
	{name: "invalid-225", source: "const a = 0; let b, c; const d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		ids:      []string{"splitUninitialized"},
		messages: []string{"Split uninitialized 'let' declarations into multiple statements."}},
	{name: "invalid-226", source: "var a; var b; var c = 0; var d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		ids:      []string{"combineUninitialized", "combineInitialized"},
		messages: []string{"Combine this with the previous 'var' statement with uninitialized variables.", "Combine this with the previous 'var' statement with initialized variables."}},
	{name: "invalid-227", source: "var a; var b = 0; var c = 1; var d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		ids:      []string{"combineInitialized"},
		messages: []string{"Combine this with the previous 'var' statement with initialized variables."}},
	{name: "invalid-228", source: "let a; let b; let c = 0; let d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		ids:      []string{"combineUninitialized", "combineInitialized"},
		messages: []string{"Combine this with the previous 'let' statement with uninitialized variables.", "Combine this with the previous 'let' statement with initialized variables."}},
	{name: "invalid-229", source: "let a; let b = 0; let c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		ids:      []string{"combineInitialized"},
		messages: []string{"Combine this with the previous 'let' statement with initialized variables."}},
	{name: "invalid-230", source: "let a; let b; const c = 0; const d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		ids:      []string{"combineUninitialized", "combineInitialized"},
		messages: []string{"Combine this with the previous 'let' statement with uninitialized variables.", "Combine this with the previous 'const' statement with initialized variables."}},
	{name: "invalid-231", source: "let a; const b = 0; const c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		ids:      []string{"combineInitialized"},
		messages: []string{"Combine this with the previous 'const' statement with initialized variables."}},
	{name: "invalid-232", source: "var a; var b; var c = 0, d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		ids:      []string{"combineUninitialized", "splitInitialized"},
		messages: []string{"Combine this with the previous 'var' statement with uninitialized variables.", "Split initialized 'var' declarations into multiple statements."}},
	{name: "invalid-233", source: "var a; var b = 0, c = 1; var d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'var' declarations into multiple statements."}},
	{name: "invalid-234", source: "let a; let b; let c = 0, d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		ids:      []string{"combineUninitialized", "splitInitialized"},
		messages: []string{"Combine this with the previous 'let' statement with uninitialized variables.", "Split initialized 'let' declarations into multiple statements."}},
	{name: "invalid-235", source: "let a; let b = 0, c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'let' declarations into multiple statements."}},
	{name: "invalid-236", source: "let a; let b; const c = 0, d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		ids:      []string{"combineUninitialized", "splitInitialized"},
		messages: []string{"Combine this with the previous 'let' statement with uninitialized variables.", "Split initialized 'const' declarations into multiple statements."}},
	{name: "invalid-237", source: "let a; const b = 0, c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'const' declarations into multiple statements."}},
	{name: "invalid-238", source: "var a = 0; var b = 1;", options: "{\"var\": \"consecutive\"}",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-239", source: "let a = 0; let b = 1;", options: "{\"let\": \"consecutive\"}",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "invalid-240", source: "const a = 0; const b = 1;", options: "{\"const\": \"consecutive\"}",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'const' statement."}},
	{name: "invalid-241", source: "let a; let b; const c = 0; const d = 1;", options: "{\"let\": \"consecutive\", \"const\": \"always\"}",
		ids:      []string{"combine", "combine"},
		messages: []string{"Combine this with the previous 'let' statement.", "Combine this with the previous 'const' statement."}},
	{name: "invalid-242", source: "let a; const b = 0; const c = 1; let d;", options: "{\"let\": \"consecutive\", \"const\": \"always\"}",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'const' statement."}},
	{name: "invalid-243", source: "let a; let b; const c = 0, d = 1;", options: "{\"let\": \"consecutive\", \"const\": \"never\"}",
		ids:      []string{"combine", "split"},
		messages: []string{"Combine this with the previous 'let' statement.", "Split 'const' declarations into multiple statements."}},
	{name: "invalid-244", source: "let a; const b = 0, c = 1; let d;", options: "{\"let\": \"consecutive\", \"const\": \"never\"}",
		ids:      []string{"split"},
		messages: []string{"Split 'const' declarations into multiple statements."}},
	{name: "invalid-245", source: "const a = 0; const b = 1; let c; let d;", options: "{\"const\": \"consecutive\", \"let\": \"always\"}",
		ids:      []string{"combine", "combine"},
		messages: []string{"Combine this with the previous 'const' statement.", "Combine this with the previous 'let' statement."}},
	{name: "invalid-246", source: "const a = 0; let b; let c; const d = 1;", options: "{\"const\": \"consecutive\", \"let\": \"always\"}",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "invalid-247", source: "const a = 0; const b = 1; let c, d;", options: "{\"const\": \"consecutive\", \"let\": \"never\"}",
		ids:      []string{"combine", "split"},
		messages: []string{"Combine this with the previous 'const' statement.", "Split 'let' declarations into multiple statements."}},
	{name: "invalid-248", source: "const a = 0; let b, c; const d = 1;", options: "{\"const\": \"consecutive\", \"let\": \"never\"}",
		ids:      []string{"split"},
		messages: []string{"Split 'let' declarations into multiple statements."}},
	{name: "invalid-249", source: "var bar; var baz;", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-250", source: "var bar = 1; var baz = 2; qux(); var qux = 3; var quux;", options: "\"consecutive\"",
		ids:      []string{"combine", "combine"},
		messages: []string{"Combine this with the previous 'var' statement.", "Combine this with the previous 'var' statement."}},
	{name: "invalid-251", source: "let a, b; let c; var d, e;", options: "{\"var\": \"never\", \"let\": \"consecutive\", \"const\": \"consecutive\"}",
		ids:      []string{"combine", "split"},
		messages: []string{"Combine this with the previous 'let' statement.", "Split 'var' declarations into multiple statements."}},
	{name: "invalid-252", source: "var a; var b;", options: "{\"var\": \"consecutive\"}",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-253", source: "var a = 1; var b = 2; var c, d; var e = 3; var f = 4;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		ids:      []string{"combineInitialized", "splitUninitialized", "combineInitialized"},
		messages: []string{"Combine this with the previous 'var' statement with initialized variables.", "Split uninitialized 'var' declarations into multiple statements.", "Combine this with the previous 'var' statement with initialized variables."}},
	{name: "invalid-254", source: "var a = 1; var b = 2; foo(); var c = 3; var d = 4;", options: "{\"initialized\": \"consecutive\"}",
		ids:      []string{"combineInitialized", "combineInitialized"},
		messages: []string{"Combine this with the previous 'var' statement with initialized variables.", "Combine this with the previous 'var' statement with initialized variables."}},
	{name: "invalid-255", source: "var a\nvar b", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-256", source: "export const foo=1, bar=2;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'const' declarations into multiple statements."}},
	{name: "invalid-257", source: "const foo=1,\n bar=2;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'const' declarations into multiple statements."}},
	{name: "invalid-258", source: "export const foo=1,\n bar=2;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'const' declarations into multiple statements."}},
	{name: "invalid-259", source: "export const foo=1\n, bar=2;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'const' declarations into multiple statements."}},
	{name: "invalid-260", source: "export const foo= a, bar=2;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'const' declarations into multiple statements."}},
	{name: "invalid-261", source: "export const foo=() => a, bar=2;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'const' declarations into multiple statements."}},
	{name: "invalid-262", source: "export const foo= a, bar=2, bar2=2;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'const' declarations into multiple statements."}},
	{name: "invalid-263", source: "export const foo = 1,bar = 2;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'const' declarations into multiple statements."}},
	{name: "invalid-264", source: "if (foo) var x, y;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-265", source: "if (foo) var x, y;", options: "{\"var\": \"never\"}",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-266", source: "if (foo) var x, y;", options: "{\"uninitialized\": \"never\"}",
		ids:      []string{"splitUninitialized"},
		messages: []string{"Split uninitialized 'var' declarations into multiple statements."}},
	{name: "invalid-267", source: "if (foo) var x = 1, y = 1;", options: "{\"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'var' declarations into multiple statements."}},
	{name: "invalid-268", source: "if (foo) {} else var x, y;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-269", source: "while (foo) var x, y;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-270", source: "do var x, y; while (foo);", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-271", source: "do var x = f(), y = b(); while (x < y);", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-272", source: "for (;;) var x, y;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-273", source: "for (foo in bar) var x, y;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-274", source: "for (foo of bar) var x, y;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-275", source: "with (foo) var x, y;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-276", source: "label: var x, y;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-277", source: "class C { static { let x, y; } }", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'let' declarations into multiple statements."}},
	{name: "invalid-278", source: "class C { static { var x, y; } }", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	{name: "invalid-279", source: "class C { static { let x; let y; } }", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "invalid-280", source: "class C { static { var x; var y; } }", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-281", source: "class C { static { let x; foo; let y; } }", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "invalid-282", source: "class C { static { var x; foo; var y; } }", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-283", source: "class C { static { var x; if (foo) { var y; } } }", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-284", source: "class C { static { let x; let y; } }", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "invalid-285", source: "class C { static { var x; var y; } }", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "invalid-286", source: "class C { static { let a = 0; let b = 1; } }", options: "{\"initialized\": \"consecutive\"}",
		ids:      []string{"combineInitialized"},
		messages: []string{"Combine this with the previous 'let' statement with initialized variables."}},
	{name: "invalid-287", source: "class C { static { var a = 0; var b = 1; } }", options: "{\"initialized\": \"consecutive\"}",
		ids:      []string{"combineInitialized"},
		messages: []string{"Combine this with the previous 'var' statement with initialized variables."}},
	{name: "invalid-288", source: "using a = 0; using b = 1;", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'using' statement."}},
	{name: "invalid-289", source: "await using a = 0; await using b = 1;", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'await using' statement."}},
	{name: "invalid-290", source: "using a = 0, b = 1;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'using' declarations into multiple statements."}},
	{name: "invalid-291", source: "await using a = 0, b = 1;", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'await using' declarations into multiple statements."}},
	{name: "invalid-292", source: "using a = 0; using b = 1;", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'using' statement."}},
	{name: "invalid-293", source: "await using a = 0; await using b = 1;", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'await using' statement."}},
	{name: "invalid-294", source: "using a = 0, b = 1;", options: "{\"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'using' declarations into multiple statements."}},
	{name: "invalid-295", source: "await using a = 0, b = 1;", options: "{\"initialized\": \"never\"}",
		ids:      []string{"splitInitialized"},
		messages: []string{"Split initialized 'await using' declarations into multiple statements."}},
}

// oneVarSilentCases are the rows upstream leaves alone.
//
// These are the false positives upstream already thought about, and every one skipped is a class
// of false positive this rule would ship.
var oneVarSilentCases = []oneVarCase{
	{name: "valid-0", source: "function foo() { var bar = true; }", options: ""},
	{name: "valid-1", source: "function foo() { var bar = true, baz = 1; if (qux) { bar = false; } }", options: ""},
	{name: "valid-2", source: "var foo = function() { var bar = true; baz(); }", options: ""},
	{name: "valid-3", source: "function foo() { var bar = true, baz = false; }", options: "\"always\""},
	{name: "valid-4", source: "function foo() { var bar = true; var baz = false; }", options: "\"never\""},
	{name: "valid-5", source: "for (var i = 0, len = arr.length; i < len; i++) {}", options: "\"never\""},
	{name: "valid-6", source: "var bar = true; var baz = false;", options: "{\"initialized\": \"never\"}"},
	{name: "valid-7", source: "var bar = true, baz = false;", options: "{\"initialized\": \"always\"}"},
	{name: "valid-8", source: "var bar, baz;", options: "{\"initialized\": \"never\"}"},
	{name: "valid-9", source: "var bar; var baz;", options: "{\"uninitialized\": \"never\"}"},
	{name: "valid-10", source: "var bar, baz;", options: "{\"uninitialized\": \"always\"}"},
	{name: "valid-11", source: "var bar = true, baz = false;", options: "{\"uninitialized\": \"never\"}"},
	{name: "valid-12", source: "var bar = true, baz = false, a, b;", options: "{\"uninitialized\": \"always\", \"initialized\": \"always\"}"},
	{name: "valid-13", source: "var bar = true; var baz = false; var a; var b;", options: "{\"uninitialized\": \"never\", \"initialized\": \"never\"}"},
	{name: "valid-14", source: "var bar, baz; var a = true; var b = false;", options: "{\"uninitialized\": \"always\", \"initialized\": \"never\"}"},
	{name: "valid-15", source: "var bar = true, baz = false; var a; var b;", options: "{\"uninitialized\": \"never\", \"initialized\": \"always\"}"},
	{name: "valid-16", source: "var bar; var baz; var a = true, b = false;", options: "{\"uninitialized\": \"never\", \"initialized\": \"always\"}"},
	{name: "valid-17", source: "function foo() { var a = [1, 2, 3]; var [b, c, d] = a; }", options: "\"never\""},
	{name: "valid-18", source: "function foo() { let a = 1; var c = true; if (a) {let c = true; } }", options: "\"always\""},
	{name: "valid-19", source: "function foo() { const a = 1; var c = true; if (a) {const c = true; } }", options: "\"always\""},
	{name: "valid-20", source: "function foo() { if (true) { const a = 1; }; if (true) {const a = true; } }", options: "\"always\""},
	{name: "valid-21", source: "function foo() { let a = 1; let b = true; }", options: "\"never\""},
	{name: "valid-22", source: "function foo() { const a = 1; const b = true; }", options: "\"never\""},
	{name: "valid-23", source: "function foo() { let a = 1; const b = false; var c = true; }", options: "\"always\""},
	{name: "valid-24", source: "function foo() { let a = 1, b = false; var c = true; }", options: "\"always\""},
	{name: "valid-25", source: "function foo() { let a = 1; let b = 2; const c = false; const d = true; var e = true, f = false; }", options: "{\"var\": \"always\", \"let\": \"never\", \"const\": \"never\"}"},
	{name: "valid-26", source: "let foo = true; for (let i = 0; i < 1; i++) { let foo = false; }", options: "{\"var\": \"always\", \"let\": \"always\", \"const\": \"never\"}"},
	{name: "valid-27", source: "let foo = true; for (let i = 0; i < 1; i++) { let foo = false; }", options: "{\"var\": \"always\"}"},
	{name: "valid-28", source: "let foo = true, bar = false;", options: "{\"var\": \"never\"}"},
	{name: "valid-29", source: "let foo = true, bar = false;", options: "{\"const\": \"never\"}"},
	{name: "valid-30", source: "let foo = true, bar = false;", options: "{\"uninitialized\": \"never\"}"},
	{name: "valid-31", source: "let foo, bar", options: "{\"initialized\": \"never\"}"},
	{name: "valid-32", source: "let foo = true, bar = false; let a; let b;", options: "{\"uninitialized\": \"never\"}"},
	{name: "valid-33", source: "let foo, bar; let a = true; let b = true;", options: "{\"initialized\": \"never\"}"},
	{name: "valid-34", source: "var foo, bar; const a=1; const b=2; let c, d", options: "{\"var\": \"always\", \"let\": \"always\"}"},
	{name: "valid-35", source: "var foo; var bar; const a=1, b=2; let c; let d", options: "{\"const\": \"always\"}"},
	{name: "valid-36", source: "for (let x of foo) {}; for (let y of foo) {}", options: "{\"uninitialized\": \"always\"}"},
	{name: "valid-37", source: "for (let x in foo) {}; for (let y in foo) {}", options: "{\"uninitialized\": \"always\"}"},
	{name: "valid-38", source: "var x; for (var y in foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}"},
	{name: "valid-39", source: "var x, y; for (y in foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}"},
	{name: "valid-40", source: "var x, y; for (var z in foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}"},
	{name: "valid-41", source: "var x; for (var y in foo) {var bar = y; for (var z in bar) {}}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}"},
	{name: "valid-42", source: "var a = 1; var b = 2; var x, y; for (var z in foo) {var baz = z; for (var d in baz) {}}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}"},
	{name: "valid-43", source: "var x; for (var y of foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}"},
	{name: "valid-44", source: "var x, y; for (y of foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}"},
	{name: "valid-45", source: "var x, y; for (var z of foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}"},
	{name: "valid-46", source: "var x; for (var y of foo) {var bar = y; for (var z of bar) {}}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}"},
	{name: "valid-47", source: "var a = 1; var b = 2; var x, y; for (var z of foo) {var baz = z; for (var d of baz) {}}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}"},
	{name: "valid-48", source: "var foo = require('foo'), bar;", options: "{\"separateRequires\": false, \"var\": \"always\"}"},
	{name: "valid-49", source: "var foo = require('foo'), bar = require('bar');", options: "{\"separateRequires\": true, \"var\": \"always\"}"},
	{name: "valid-50", source: "var bar = 'bar'; var foo = require('foo');", options: "{\"separateRequires\": true, \"var\": \"always\"}"},
	{name: "valid-51", source: "var foo = require('foo'); var bar = 'bar';", options: "{\"separateRequires\": true, \"var\": \"always\"}"},
	{name: "valid-52", source: "var a = 0, b, c;", options: "\"consecutive\""},
	{name: "valid-53", source: "var a = 0, b = 1, c = 2;", options: "\"consecutive\""},
	{name: "valid-54", source: "var a = 0, b = 1; foo(); var c = 2;", options: "\"consecutive\""},
	{name: "valid-55", source: "let a = 0, b, c;", options: "\"consecutive\""},
	{name: "valid-56", source: "let a = 0, b = 1, c = 2;", options: "\"consecutive\""},
	{name: "valid-57", source: "let a = 0, b = 1; foo(); let c = 2;", options: "\"consecutive\""},
	{name: "valid-58", source: "const a = 0, b = 1; foo(); const c = 2;", options: "\"consecutive\""},
	{name: "valid-59", source: "const a = 0; var b = 1;", options: "\"consecutive\""},
	{name: "valid-60", source: "const a = 0; let b = 1;", options: "\"consecutive\""},
	{name: "valid-61", source: "let a = 0; const b = 1; var c = 2;", options: "\"consecutive\""},
	{name: "valid-62", source: "const foo = require('foo'); const bar = 'bar';", options: "{\"const\": \"consecutive\", \"separateRequires\": true}"},
	{name: "valid-63", source: "var a = 0, b = 1; var c, d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}"},
	{name: "valid-64", source: "var a = 0; var b, c; var d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}"},
	{name: "valid-65", source: "let a = 0, b = 1; let c, d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}"},
	{name: "valid-66", source: "let a = 0; let b, c; let d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}"},
	{name: "valid-67", source: "const a = 0, b = 1; let c, d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}"},
	{name: "valid-68", source: "const a = 0; let b, c; const d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}"},
	{name: "valid-69", source: "var a = 0, b = 1; var c; var d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}"},
	{name: "valid-70", source: "var a = 0; var b; var c; var d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}"},
	{name: "valid-71", source: "let a = 0, b = 1; let c; let d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}"},
	{name: "valid-72", source: "let a = 0; let b; let c; let d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}"},
	{name: "valid-73", source: "const a = 0, b = 1; let c; let d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}"},
	{name: "valid-74", source: "const a = 0; let b; let c; const d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}"},
	{name: "valid-75", source: "var a, b; var c = 0, d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}"},
	{name: "valid-76", source: "var a; var b = 0, c = 1; var d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}"},
	{name: "valid-77", source: "let a, b; let c = 0, d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}"},
	{name: "valid-78", source: "let a; let b = 0, c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}"},
	{name: "valid-79", source: "let a, b; const c = 0, d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}"},
	{name: "valid-80", source: "let a; const b = 0, c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}"},
	{name: "valid-81", source: "var a, b; var c = 0; var d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}"},
	{name: "valid-82", source: "var a; var b = 0; var c = 1; var d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}"},
	{name: "valid-83", source: "let a, b; let c = 0; let d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}"},
	{name: "valid-84", source: "let a; let b = 0; let c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}"},
	{name: "valid-85", source: "let a, b; const c = 0; const d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}"},
	{name: "valid-86", source: "let a; const b = 0; const c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}"},
	{name: "valid-87", source: "var a = 0, b = 1;", options: "{\"var\": \"consecutive\"}"},
	{name: "valid-88", source: "var a = 0; foo; var b = 1;", options: "{\"var\": \"consecutive\"}"},
	{name: "valid-89", source: "let a = 0, b = 1;", options: "{\"let\": \"consecutive\"}"},
	{name: "valid-90", source: "let a = 0; foo; let b = 1;", options: "{\"let\": \"consecutive\"}"},
	{name: "valid-91", source: "const a = 0, b = 1;", options: "{\"const\": \"consecutive\"}"},
	{name: "valid-92", source: "const a = 0; foo; const b = 1;", options: "{\"const\": \"consecutive\"}"},
	{name: "valid-93", source: "let a, b; const c = 0, d = 1;", options: "{\"let\": \"consecutive\", \"const\": \"always\"}"},
	{name: "valid-94", source: "let a; const b = 0, c = 1; let d;", options: "{\"let\": \"consecutive\", \"const\": \"always\"}"},
	{name: "valid-95", source: "let a, b; const c = 0; const d = 1;", options: "{\"let\": \"consecutive\", \"const\": \"never\"}"},
	{name: "valid-96", source: "let a; const b = 0; const c = 1; let d;", options: "{\"let\": \"consecutive\", \"const\": \"never\"}"},
	{name: "valid-97", source: "const a = 0, b = 1; let c, d;", options: "{\"const\": \"consecutive\", \"let\": \"always\"}"},
	{name: "valid-98", source: "const a = 0; let b, c; const d = 1;", options: "{\"const\": \"consecutive\", \"let\": \"always\"}"},
	{name: "valid-99", source: "const a = 0, b = 1; let c; let d;", options: "{\"const\": \"consecutive\", \"let\": \"never\"}"},
	{name: "valid-100", source: "const a = 0; let b; let c; const d = 1;", options: "{\"const\": \"consecutive\", \"let\": \"never\"}"},
	{name: "valid-101", source: "var a = 1, b = 2; foo(); var c = 3, d = 4;", options: "{\"initialized\": \"consecutive\"}"},
	{name: "valid-102", source: "var bar, baz;", options: "\"consecutive\""},
	{name: "valid-103", source: "var bar = 1, baz = 2; qux(); var qux = 3, quux;", options: "\"consecutive\""},
	{name: "valid-104", source: "let a, b; var c; var d; let e;", options: "{\"var\": \"never\", \"let\": \"consecutive\", \"const\": \"consecutive\"}"},
	{name: "valid-105", source: "const a = 1, b = 2; var d; var e; const f = 3;", options: "{\"var\": \"never\", \"let\": \"consecutive\", \"const\": \"consecutive\"}"},
	{name: "valid-106", source: "var a, b; const c = 1; const d = 2; let e; let f; ", options: "{\"var\": \"consecutive\"}"},
	{name: "valid-107", source: "var a = 1, b = 2; var c; var d; var e = 3, f = 4;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}"},
	{name: "valid-108", source: "var a; somethingElse(); var b;", options: "{\"var\": \"never\"}"},
	{name: "valid-109", source: "var foo = 1;\nlet bar = function() { var x; };\nvar baz = 2;", options: "{\"var\": \"never\"}"},
	{name: "valid-110", source: "class C { static { var a; let b; const c = 0; } }", options: "\"always\""},
	{name: "valid-111", source: "const a = 0; class C { static { const b = 0; } }", options: "\"always\""},
	{name: "valid-112", source: "class C { static { const b = 0; } } const a = 0; ", options: "\"always\""},
	{name: "valid-113", source: "let a; class C { static { let b; } }", options: "\"always\""},
	{name: "valid-114", source: "class C { static { let b; } } let a;", options: "\"always\""},
	{name: "valid-115", source: "var a; class C { static { var b; } }", options: "\"always\""},
	{name: "valid-116", source: "class C { static { var b; } } var a; ", options: "\"always\""},
	{name: "valid-117", source: "var a; class C { static { if (foo) { var b; } } }", options: "\"always\""},
	{name: "valid-118", source: "class C { static { if (foo) { var b; } } } var a; ", options: "\"always\""},
	{name: "valid-119", source: "class C { static { const a = 0; if (foo) { const b = 0; } } }", options: "\"always\""},
	{name: "valid-120", source: "class C { static { let a; if (foo) { let b; } } }", options: "\"always\""},
	{name: "valid-121", source: "class C { static { const a = 0; const b = 0; } }", options: "\"never\""},
	{name: "valid-122", source: "class C { static { let a; let b; } }", options: "\"never\""},
	{name: "valid-123", source: "class C { static { var a; var b; } }", options: "\"never\""},
	{name: "valid-124", source: "class C { static { let a; foo; let b; } }", options: "\"consecutive\""},
	{name: "valid-125", source: "class C { static { let a; const b = 0; let c; } }", options: "\"consecutive\""},
	{name: "valid-126", source: "class C { static { var a; foo; var b; } }", options: "\"consecutive\""},
	{name: "valid-127", source: "class C { static { var a; let b; var c; } }", options: "\"consecutive\""},
	{name: "valid-128", source: "class C { static { let a; if (foo) { let b; } } }", options: "\"consecutive\""},
	{name: "valid-129", source: "class C { static { if (foo) { let b; } let a;  } }", options: "\"consecutive\""},
	{name: "valid-130", source: "class C { static { const a = 0; if (foo) { const b = 0; } } }", options: "\"consecutive\""},
	{name: "valid-131", source: "class C { static { if (foo) { const b = 0; } const a = 0; } }", options: "\"consecutive\""},
	{name: "valid-132", source: "class C { static { var a; if (foo) var b; } }", options: "\"consecutive\""},
	{name: "valid-133", source: "class C { static { if (foo) var b; var a; } }", options: "\"consecutive\""},
	{name: "valid-134", source: "class C { static { if (foo) { var b; } var a; } }", options: "\"consecutive\""},
	{name: "valid-135", source: "class C { static { let a; let b = 0; } }", options: "{\"initialized\": \"consecutive\"}"},
	{name: "valid-136", source: "class C { static { var a; var b = 0; } }", options: "{\"initialized\": \"consecutive\"}"},
	{name: "valid-137", source: "using a = 0; let b = 1; const c = 2;", options: ""},
	{name: "valid-138", source: "await using a = 0; let b = 1; const c = 2;", options: ""},
	{name: "valid-139", source: "using a = 0, b = 1;", options: ""},
	{name: "valid-140", source: "await using a = 0, b = 1;", options: ""},
	{name: "valid-141", source: "function fn() { { using a = 0; } using b = 1; }", options: ""},
	{name: "valid-142", source: "using a = 0; using b = 1;", options: "\"never\""},
	{name: "valid-143", source: "await using a = 0; await using b = 1;", options: "\"never\""},
	{name: "valid-144", source: "using a = 0, b = 1;", options: "\"consecutive\""},
	{name: "valid-145", source: "await using a = 0, b = 1;", options: "\"consecutive\""},
	{name: "valid-146", source: "using a = 0, b = 1;", options: "{\"initialized\": \"always\"}"},
	{name: "valid-147", source: "await using a = 0, b = 1;", options: "{\"initialized\": \"always\"}"},
	{name: "valid-148", source: "using a = 0; using b = 1;", options: "{\"initialized\": \"never\"}"},
	{name: "valid-149", source: "await using a = 0; await using b = 1;", options: "{\"initialized\": \"never\"}"},
	{name: "valid-150", source: "using a = 0, b = 1; foo(); using c = 2, d = 3;", options: "{\"initialized\": \"consecutive\"}"},
	{name: "valid-151", source: "await using a = 0, b = 1; foo(); await using c = 2, d = 3;", options: "{\"initialized\": \"consecutive\"}"},
}

// oneVarRepairCases are the rows whose `output` upstream pins to a rewritten string.
//
// `output` is the SPECIFICATION for the repair, and it is the ONLY coverage a fixer has: a fixer
// that reports in the right place and rewrites wrongly is indistinguishable, at the message-id
// layer, from a correct one.
//
// These assert a SINGLE fix pass, which is what upstream's RuleTester records. The real engine
// loops to a fixed point, and three of these cases converge somewhere else because the second
// pass merges declarations the first one only just created. Asserting the fixed point here would
// disagree with upstream's own expectations on exactly those three.
var oneVarRepairCases = []oneVarFixCase{
	{name: "invalid-152", source: "var bar = true, baz = false;", options: "\"never\"",
		wanted: "var bar = true; var baz = false;"},
	{name: "invalid-153", source: "function foo() { var bar = true, baz = false; }", options: "\"never\"",
		wanted: "function foo() { var bar = true; var baz = false; }"},
	{name: "invalid-154", source: "if (foo) { var bar = true, baz = false; }", options: "\"never\"",
		wanted: "if (foo) { var bar = true; var baz = false; }"},
	{name: "invalid-155", source: "switch (foo) { case bar: var baz = true, quux = false; }", options: "\"never\"",
		wanted: "switch (foo) { case bar: var baz = true; var quux = false; }"},
	{name: "invalid-156", source: "switch (foo) { default: var baz = true, quux = false; }", options: "\"never\"",
		wanted: "switch (foo) { default: var baz = true; var quux = false; }"},
	{name: "invalid-157", source: "function foo() { var bar = true; var baz = false; }", options: "\"always\"",
		wanted: "function foo() { var bar = true,  baz = false; }"},
	{name: "invalid-159", source: "function foo() { var foo = true, bar = false; }", options: "{\"initialized\": \"never\"}",
		wanted: "function foo() { var foo = true; var bar = false; }"},
	{name: "invalid-160", source: "function foo() { var foo, bar; }", options: "{\"uninitialized\": \"never\"}",
		wanted: "function foo() { var foo; var bar; }"},
	{name: "invalid-161", source: "function foo() { var bar, baz; var a = true; var b = false; var c, d;}", options: "{\"uninitialized\": \"always\", \"initialized\": \"never\"}",
		wanted: "function foo() { var bar, baz; var a = true; var b = false,  c, d;}"},
	{name: "invalid-162", source: "function foo() { var bar = true, baz = false; var a; var b; var c = true, d = false; }", options: "{\"uninitialized\": \"never\", \"initialized\": \"always\"}",
		wanted: "function foo() { var bar = true, baz = false; var a; var b,  c = true, d = false; }"},
	{name: "invalid-163", source: "function foo() { var bar = true, baz = false; var a, b;}", options: "{\"uninitialized\": \"never\", \"initialized\": \"never\"}",
		wanted: "function foo() { var bar = true; var baz = false; var a; var b;}"},
	{name: "invalid-164", source: "function foo() { var bar = true; var baz = false; var a; var b;}", options: "{\"uninitialized\": \"always\", \"initialized\": \"always\"}",
		wanted: "function foo() { var bar = true,  baz = false,  a,  b;}"},
	{name: "invalid-165", source: "function foo() { var a = [1, 2, 3]; var [b, c, d] = a; }", options: "\"always\"",
		wanted: "function foo() { var a = [1, 2, 3],  [b, c, d] = a; }"},
	{name: "invalid-166", source: "function foo() { let a = 1; let b = 2; }", options: "\"always\"",
		wanted: "function foo() { let a = 1,  b = 2; }"},
	{name: "invalid-167", source: "function foo() { const a = 1; const b = 2; }", options: "\"always\"",
		wanted: "function foo() { const a = 1,  b = 2; }"},
	{name: "invalid-168", source: "function foo() { let a = 1; let b = 2; }", options: "{\"let\": \"always\"}",
		wanted: "function foo() { let a = 1,  b = 2; }"},
	{name: "invalid-169", source: "function foo() { const a = 1; const b = 2; }", options: "{\"const\": \"always\"}",
		wanted: "function foo() { const a = 1,  b = 2; }"},
	{name: "invalid-170", source: "function foo() { let a = 1, b = 2; }", options: "{\"let\": \"never\"}",
		wanted: "function foo() { let a = 1; let b = 2; }"},
	{name: "invalid-171", source: "function foo() { let a = 1, b = 2; }", options: "{\"initialized\": \"never\"}",
		wanted: "function foo() { let a = 1; let b = 2; }"},
	{name: "invalid-172", source: "function foo() { let a, b; }", options: "{\"uninitialized\": \"never\"}",
		wanted: "function foo() { let a; let b; }"},
	{name: "invalid-173", source: "function foo() { const a = 1, b = 2; }", options: "{\"initialized\": \"never\"}",
		wanted: "function foo() { const a = 1; const b = 2; }"},
	{name: "invalid-174", source: "function foo() { const a = 1, b = 2; }", options: "{\"const\": \"never\"}",
		wanted: "function foo() { const a = 1; const b = 2; }"},
	{name: "invalid-176", source: "var one = 1, two = 2;\nvar three;", options: "\"always\"",
		wanted: "var one = 1, two = 2,\n three;"},
	{name: "invalid-177", source: "var i = [0], j;", options: "{\"initialized\": \"never\"}",
		wanted: "var i = [0]; var j;"},
	{name: "invalid-178", source: "var i = [0], j;", options: "{\"uninitialized\": \"never\"}",
		wanted: "var i = [0]; var j;"},
	{name: "invalid-181", source: "var foo = function() { var bar = true; var baz = false; }", options: "",
		wanted: "var foo = function() { var bar = true,  baz = false; }"},
	{name: "invalid-183", source: "var foo = () => { var bar = true; var baz = false; }", options: "",
		wanted: "var foo = () => { var bar = true,  baz = false; }"},
	{name: "invalid-185", source: "var foo; var bar;", options: "",
		wanted: "var foo,  bar;"},
	{name: "invalid-186", source: "var x = 1, y = 2; for (var z in foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		wanted: "var x = 1; var y = 2; for (var z in foo) {}"},
	{name: "invalid-187", source: "var x = 1, y = 2; for (var z of foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		wanted: "var x = 1; var y = 2; for (var z of foo) {}"},
	{name: "invalid-188", source: "var x; var y; for (var z in foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		wanted: "var x,  y; for (var z in foo) {}"},
	{name: "invalid-189", source: "var x; var y; for (var z of foo) {}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		wanted: "var x,  y; for (var z of foo) {}"},
	{name: "invalid-190", source: "var x; for (var y in foo) {var bar = y; var a; for (var z of bar) {}}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		wanted: "var x; for (var y in foo) {var bar = y,  a; for (var z of bar) {}}"},
	{name: "invalid-191", source: "var a = 1; var b = 2; var x, y; for (var z of foo) {var c = 3, baz = z; for (var d in baz) {}}", options: "{\"initialized\": \"never\", \"uninitialized\": \"always\"}",
		wanted: "var a = 1; var b = 2; var x, y; for (var z of foo) {var c = 3; var baz = z; for (var d in baz) {}}"},
	{name: "invalid-192", source: "var {foo} = 1, [bar] = 2;", options: "{\"initialized\": \"never\"}",
		wanted: "var {foo} = 1; var [bar] = 2;"},
	{name: "invalid-193", source: "const foo = 1,\n    bar = 2;", options: "{\"initialized\": \"never\"}",
		wanted: "const foo = 1;\n    const bar = 2;"},
	{name: "invalid-194", source: "var foo = 1,\n    bar = 2;", options: "{\"initialized\": \"never\"}",
		wanted: "var foo = 1;\n    var bar = 2;"},
	{name: "invalid-195", source: "var foo = 1, // comment\n    bar = 2;", options: "{\"initialized\": \"never\"}",
		wanted: "var foo = 1; // comment\n    var bar = 2;"},
	{name: "invalid-196", source: "var f, k /* test */, l;", options: "\"never\"",
		wanted: "var f; var k /* test */; var l;"},
	{name: "invalid-197", source: "var f,          /* test */ l;", options: "\"never\"",
		wanted: "var f;          /* test */ var l;"},
	{name: "invalid-198", source: "var f, k /* test \n some more comment \n even more */, l = 1, P;", options: "\"never\"",
		wanted: "var f; var k /* test \n some more comment \n even more */; var l = 1; var P;"},
	{name: "invalid-199", source: "var a = 1, b = 2", options: "\"never\"",
		wanted: "var a = 1; var b = 2"},
	{name: "invalid-204", source: "const foo = require('foo'); const bar = require('bar');", options: "{\"separateRequires\": true, \"const\": \"always\"}",
		wanted: "const foo = require('foo'),  bar = require('bar');"},
	{name: "invalid-205", source: "var a = 1, b; var c;", options: "\"consecutive\"",
		wanted: "var a = 1, b,  c;"},
	{name: "invalid-206", source: "var a = 0, b = 1; var c = 2;", options: "\"consecutive\"",
		wanted: "var a = 0, b = 1,  c = 2;"},
	{name: "invalid-207", source: "let a = 1, b; let c;", options: "\"consecutive\"",
		wanted: "let a = 1, b,  c;"},
	{name: "invalid-208", source: "let a = 0, b = 1; let c = 2;", options: "\"consecutive\"",
		wanted: "let a = 0, b = 1,  c = 2;"},
	{name: "invalid-209", source: "const a = 0, b = 1; const c = 2;", options: "\"consecutive\"",
		wanted: "const a = 0, b = 1,  c = 2;"},
	{name: "invalid-210", source: "const a = 0; var b = 1; var c = 2; const d = 3;", options: "\"consecutive\"",
		wanted: "const a = 0; var b = 1,  c = 2; const d = 3;"},
	{name: "invalid-211", source: "var a = true; var b = false;", options: "{\"separateRequires\": true, \"var\": \"always\"}",
		wanted: "var a = true,  b = false;"},
	{name: "invalid-212", source: "const a = 0; let b = 1; let c = 2; const d = 3;", options: "\"consecutive\"",
		wanted: "const a = 0; let b = 1,  c = 2; const d = 3;"},
	{name: "invalid-213", source: "let a = 0; const b = 1; const c = 1; var d = 2;", options: "\"consecutive\"",
		wanted: "let a = 0; const b = 1,  c = 1; var d = 2;"},
	{name: "invalid-214", source: "var a = 0; var b; var c; var d = 1", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		wanted: "var a = 0; var b,  c; var d = 1"},
	{name: "invalid-215", source: "var a = 0; var b = 1; var c; var d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		wanted: "var a = 0,  b = 1; var c,  d;"},
	{name: "invalid-216", source: "let a = 0; let b; let c; let d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		wanted: "let a = 0; let b,  c; let d = 1;"},
	{name: "invalid-217", source: "let a = 0; let b = 1; let c; let d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		wanted: "let a = 0,  b = 1; let c,  d;"},
	{name: "invalid-218", source: "const a = 0; let b; let c; const d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		wanted: "const a = 0; let b,  c; const d = 1;"},
	{name: "invalid-219", source: "const a = 0; const b = 1; let c; let d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"always\"}",
		wanted: "const a = 0,  b = 1; let c,  d;"},
	{name: "invalid-220", source: "var a = 0; var b = 1; var c, d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		wanted: "var a = 0,  b = 1; var c; var d;"},
	{name: "invalid-221", source: "var a = 0; var b, c; var d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		wanted: "var a = 0; var b; var c; var d = 1;"},
	{name: "invalid-222", source: "let a = 0; let b = 1; let c, d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		wanted: "let a = 0,  b = 1; let c; let d;"},
	{name: "invalid-223", source: "let a = 0; let b, c; let d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		wanted: "let a = 0; let b; let c; let d = 1;"},
	{name: "invalid-224", source: "const a = 0; const b = 1; let c, d;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		wanted: "const a = 0,  b = 1; let c; let d;"},
	{name: "invalid-225", source: "const a = 0; let b, c; const d = 1;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		wanted: "const a = 0; let b; let c; const d = 1;"},
	{name: "invalid-226", source: "var a; var b; var c = 0; var d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		wanted: "var a,  b; var c = 0,  d = 1;"},
	{name: "invalid-227", source: "var a; var b = 0; var c = 1; var d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		wanted: "var a; var b = 0,  c = 1; var d;"},
	{name: "invalid-228", source: "let a; let b; let c = 0; let d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		wanted: "let a,  b; let c = 0,  d = 1;"},
	{name: "invalid-229", source: "let a; let b = 0; let c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		wanted: "let a; let b = 0,  c = 1; let d;"},
	{name: "invalid-230", source: "let a; let b; const c = 0; const d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		wanted: "let a,  b; const c = 0,  d = 1;"},
	{name: "invalid-231", source: "let a; const b = 0; const c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"always\"}",
		wanted: "let a; const b = 0,  c = 1; let d;"},
	{name: "invalid-232", source: "var a; var b; var c = 0, d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		wanted: "var a,  b; var c = 0; var d = 1;"},
	{name: "invalid-233", source: "var a; var b = 0, c = 1; var d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		wanted: "var a; var b = 0; var c = 1; var d;"},
	{name: "invalid-234", source: "let a; let b; let c = 0, d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		wanted: "let a,  b; let c = 0; let d = 1;"},
	{name: "invalid-235", source: "let a; let b = 0, c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		wanted: "let a; let b = 0; let c = 1; let d;"},
	{name: "invalid-236", source: "let a; let b; const c = 0, d = 1;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		wanted: "let a,  b; const c = 0; const d = 1;"},
	{name: "invalid-237", source: "let a; const b = 0, c = 1; let d;", options: "{\"uninitialized\": \"consecutive\", \"initialized\": \"never\"}",
		wanted: "let a; const b = 0; const c = 1; let d;"},
	{name: "invalid-238", source: "var a = 0; var b = 1;", options: "{\"var\": \"consecutive\"}",
		wanted: "var a = 0,  b = 1;"},
	{name: "invalid-239", source: "let a = 0; let b = 1;", options: "{\"let\": \"consecutive\"}",
		wanted: "let a = 0,  b = 1;"},
	{name: "invalid-240", source: "const a = 0; const b = 1;", options: "{\"const\": \"consecutive\"}",
		wanted: "const a = 0,  b = 1;"},
	{name: "invalid-241", source: "let a; let b; const c = 0; const d = 1;", options: "{\"let\": \"consecutive\", \"const\": \"always\"}",
		wanted: "let a,  b; const c = 0,  d = 1;"},
	{name: "invalid-242", source: "let a; const b = 0; const c = 1; let d;", options: "{\"let\": \"consecutive\", \"const\": \"always\"}",
		wanted: "let a; const b = 0,  c = 1; let d;"},
	{name: "invalid-243", source: "let a; let b; const c = 0, d = 1;", options: "{\"let\": \"consecutive\", \"const\": \"never\"}",
		wanted: "let a,  b; const c = 0; const d = 1;"},
	{name: "invalid-244", source: "let a; const b = 0, c = 1; let d;", options: "{\"let\": \"consecutive\", \"const\": \"never\"}",
		wanted: "let a; const b = 0; const c = 1; let d;"},
	{name: "invalid-245", source: "const a = 0; const b = 1; let c; let d;", options: "{\"const\": \"consecutive\", \"let\": \"always\"}",
		wanted: "const a = 0,  b = 1; let c,  d;"},
	{name: "invalid-246", source: "const a = 0; let b; let c; const d = 1;", options: "{\"const\": \"consecutive\", \"let\": \"always\"}",
		wanted: "const a = 0; let b,  c; const d = 1;"},
	{name: "invalid-247", source: "const a = 0; const b = 1; let c, d;", options: "{\"const\": \"consecutive\", \"let\": \"never\"}",
		wanted: "const a = 0,  b = 1; let c; let d;"},
	{name: "invalid-248", source: "const a = 0; let b, c; const d = 1;", options: "{\"const\": \"consecutive\", \"let\": \"never\"}",
		wanted: "const a = 0; let b; let c; const d = 1;"},
	{name: "invalid-249", source: "var bar; var baz;", options: "\"consecutive\"",
		wanted: "var bar,  baz;"},
	{name: "invalid-250", source: "var bar = 1; var baz = 2; qux(); var qux = 3; var quux;", options: "\"consecutive\"",
		wanted: "var bar = 1,  baz = 2; qux(); var qux = 3,  quux;"},
	{name: "invalid-251", source: "let a, b; let c; var d, e;", options: "{\"var\": \"never\", \"let\": \"consecutive\", \"const\": \"consecutive\"}",
		wanted: "let a, b,  c; var d; var e;"},
	{name: "invalid-252", source: "var a; var b;", options: "{\"var\": \"consecutive\"}",
		wanted: "var a,  b;"},
	{name: "invalid-253", source: "var a = 1; var b = 2; var c, d; var e = 3; var f = 4;", options: "{\"initialized\": \"consecutive\", \"uninitialized\": \"never\"}",
		wanted: "var a = 1,  b = 2; var c; var d; var e = 3,  f = 4;"},
	{name: "invalid-254", source: "var a = 1; var b = 2; foo(); var c = 3; var d = 4;", options: "{\"initialized\": \"consecutive\"}",
		wanted: "var a = 1,  b = 2; foo(); var c = 3,  d = 4;"},
	{name: "invalid-255", source: "var a\nvar b", options: "\"always\"",
		wanted: "var a,\n b"},
	{name: "invalid-256", source: "export const foo=1, bar=2;", options: "\"never\"",
		wanted: "export const foo=1; export const bar=2;"},
	{name: "invalid-257", source: "const foo=1,\n bar=2;", options: "\"never\"",
		wanted: "const foo=1;\n const bar=2;"},
	{name: "invalid-258", source: "export const foo=1,\n bar=2;", options: "\"never\"",
		wanted: "export const foo=1;\n export const bar=2;"},
	{name: "invalid-259", source: "export const foo=1\n, bar=2;", options: "\"never\"",
		wanted: "export const foo=1\n; export const bar=2;"},
	{name: "invalid-260", source: "export const foo= a, bar=2;", options: "\"never\"",
		wanted: "export const foo= a; export const bar=2;"},
	{name: "invalid-261", source: "export const foo=() => a, bar=2;", options: "\"never\"",
		wanted: "export const foo=() => a; export const bar=2;"},
	{name: "invalid-262", source: "export const foo= a, bar=2, bar2=2;", options: "\"never\"",
		wanted: "export const foo= a; export const bar=2; export const bar2=2;"},
	{name: "invalid-263", source: "export const foo = 1,bar = 2;", options: "\"never\"",
		wanted: "export const foo = 1; export const bar = 2;"},
	{name: "invalid-277", source: "class C { static { let x, y; } }", options: "\"never\"",
		wanted: "class C { static { let x; let y; } }"},
	{name: "invalid-278", source: "class C { static { var x, y; } }", options: "\"never\"",
		wanted: "class C { static { var x; var y; } }"},
	{name: "invalid-279", source: "class C { static { let x; let y; } }", options: "\"always\"",
		wanted: "class C { static { let x,  y; } }"},
	{name: "invalid-280", source: "class C { static { var x; var y; } }", options: "\"always\"",
		wanted: "class C { static { var x,  y; } }"},
	{name: "invalid-284", source: "class C { static { let x; let y; } }", options: "\"consecutive\"",
		wanted: "class C { static { let x,  y; } }"},
	{name: "invalid-285", source: "class C { static { var x; var y; } }", options: "\"consecutive\"",
		wanted: "class C { static { var x,  y; } }"},
	{name: "invalid-286", source: "class C { static { let a = 0; let b = 1; } }", options: "{\"initialized\": \"consecutive\"}",
		wanted: "class C { static { let a = 0,  b = 1; } }"},
	{name: "invalid-287", source: "class C { static { var a = 0; var b = 1; } }", options: "{\"initialized\": \"consecutive\"}",
		wanted: "class C { static { var a = 0,  b = 1; } }"},
	{name: "invalid-288", source: "using a = 0; using b = 1;", options: "\"always\"",
		wanted: "using a = 0,  b = 1;"},
	{name: "invalid-289", source: "await using a = 0; await using b = 1;", options: "\"always\"",
		wanted: "await using a = 0,   b = 1;"},
	{name: "invalid-290", source: "using a = 0, b = 1;", options: "\"never\"",
		wanted: "using a = 0; using b = 1;"},
	{name: "invalid-291", source: "await using a = 0, b = 1;", options: "\"never\"",
		wanted: "await using a = 0; await using b = 1;"},
	{name: "invalid-292", source: "using a = 0; using b = 1;", options: "\"consecutive\"",
		wanted: "using a = 0,  b = 1;"},
	{name: "invalid-293", source: "await using a = 0; await using b = 1;", options: "\"consecutive\"",
		wanted: "await using a = 0,   b = 1;"},
	{name: "invalid-294", source: "using a = 0, b = 1;", options: "{\"initialized\": \"never\"}",
		wanted: "using a = 0; using b = 1;"},
	{name: "invalid-295", source: "await using a = 0, b = 1;", options: "{\"initialized\": \"never\"}",
		wanted: "await using a = 0; await using b = 1;"},
}

// oneVarDeclineCases are the rows upstream reports and deliberately does NOT repair.
//
// `output: null` is a decision rather than an omission. Twenty-six of them here, more than any
// other rule in this batch, and they are asserted as the ABSENCE of a proposed fix -- never as
// "the output equals the input", because `ExpectFixedSource` fatals when no fix was proposed and
// would therefore agree with any expectation at all.
var oneVarDeclineCases = []oneVarCase{
	{name: "invalid-158", source: "var a = 1; for (var b = 2;;) {}", options: "\"always\""},
	{name: "invalid-175", source: "let foo = true; switch(foo) { case true: let bar = 2; break; case false: let baz = 3; break; }", options: "{\"var\": \"always\", \"let\": \"always\", \"const\": \"never\"}"},
	{name: "invalid-179", source: "for (var x of foo) {}; for (var y of foo) {}", options: "\"always\""},
	{name: "invalid-180", source: "for (var x in foo) {}; for (var y in foo) {}", options: "\"always\""},
	{name: "invalid-182", source: "function foo() { var bar = true; if (qux) { var baz = false; } else { var quxx = 42; } }", options: ""},
	{name: "invalid-184", source: "var foo = function() { var bar = true; if (qux) { var baz = false; } }", options: ""},
	{name: "invalid-200", source: "var foo = require('foo'), bar;", options: "{\"separateRequires\": true, \"var\": \"always\"}"},
	{name: "invalid-201", source: "var foo, bar = require('bar');", options: "{\"separateRequires\": true, \"var\": \"always\"}"},
	{name: "invalid-202", source: "let foo, bar = require('bar');", options: "{\"separateRequires\": true, \"let\": \"always\"}"},
	{name: "invalid-203", source: "const foo = 0, bar = require('bar');", options: "{\"separateRequires\": true, \"const\": \"always\"}"},
	{name: "invalid-264", source: "if (foo) var x, y;", options: "\"never\""},
	{name: "invalid-265", source: "if (foo) var x, y;", options: "{\"var\": \"never\"}"},
	{name: "invalid-266", source: "if (foo) var x, y;", options: "{\"uninitialized\": \"never\"}"},
	{name: "invalid-267", source: "if (foo) var x = 1, y = 1;", options: "{\"initialized\": \"never\"}"},
	{name: "invalid-268", source: "if (foo) {} else var x, y;", options: "\"never\""},
	{name: "invalid-269", source: "while (foo) var x, y;", options: "\"never\""},
	{name: "invalid-270", source: "do var x, y; while (foo);", options: "\"never\""},
	{name: "invalid-271", source: "do var x = f(), y = b(); while (x < y);", options: "\"never\""},
	{name: "invalid-272", source: "for (;;) var x, y;", options: "\"never\""},
	{name: "invalid-273", source: "for (foo in bar) var x, y;", options: "\"never\""},
	{name: "invalid-274", source: "for (foo of bar) var x, y;", options: "\"never\""},
	{name: "invalid-275", source: "with (foo) var x, y;", options: "\"never\""},
	{name: "invalid-276", source: "label: var x, y;", options: "\"never\""},
	{name: "invalid-281", source: "class C { static { let x; foo; let y; } }", options: "\"always\""},
	{name: "invalid-282", source: "class C { static { var x; foo; var y; } }", options: "\"always\""},
	{name: "invalid-283", source: "class C { static { var x; if (foo) { var y; } } }", options: "\"always\""},
}

func TestOneVarFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range oneVarFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

func TestOneVarStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range oneVarSilentCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestOneVarRendersTheTypeUpstreamRenders asserts the message TEXT, not only the id.
//
// Six of the seven messages interpolate `{{type}}`, whose value is the declaration spelling --
// and `await using` is the one most likely to be computed wrongly, because the flag named for it
// is a composite that is also true for every plain `const`. A rule that classified every `const`
// as `await using` would report the right id with the wrong word and pass TestOneVarFires
// completely.
//
// The comparison is on the rendered prefix rather than the whole Description, because this tree's
// messages carry an explanation upstream does not have. That is a weaker predicate than the whole
// string, so the prefix asserted is the entire sentence upstream renders, up to and including its
// terminating period.
func TestOneVarRendersTheTypeUpstreamRenders(t *testing.T) {
	t.Parallel()
	for _, testCase := range oneVarFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			if len(result.Diagnostics) != len(testCase.messages) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.messages))
			}
			for index, diagnostic := range result.Diagnostics {
				wanted := testCase.messages[index]
				if !strings.HasPrefix(diagnostic.Message.Description, wanted) {
					t.Errorf("finding %d rendered:\n  %q\nwant it to begin with upstream's:\n  %q",
						index, diagnostic.Message.Description, wanted)
				}
			}
		})
	}
}

// TestOneVarAnchorsOnTheDeclarationStatement asserts the SPAN of every finding.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule pointing at the wrong node
// passes a complete fixture pair. That matters twice over here because the findings carry fixes:
// a finding reported at the wrong line while carrying a repair means the edit lands somewhere the
// reader was never shown.
//
// Upstream reports on the whole declaration statement, so the reported text must begin with the
// declaration keyword -- or with `export`, for an exported one.
func TestOneVarAnchorsOnTheDeclarationStatement(t *testing.T) {
	t.Parallel()
	for _, testCase := range oneVarFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			source := result.SourceFile.Text()
			for index, diagnostic := range result.Diagnostics {
				start, end := diagnostic.Range.Pos(), diagnostic.Range.End()
				if start < 0 || end > len(source) || start >= end {
					t.Fatalf("finding %d has range [%d,%d) outside %q", index, start, end, source)
				}
				reported := source[start:end]
				if !oneVarBeginsADeclaration(reported) {
					t.Errorf("finding %d points at %q, which does not begin a declaration statement",
						index, reported)
				}
			}
		})
	}
}

// oneVarBeginsADeclaration answers whether a reported span starts where a declaration does.
func oneVarBeginsADeclaration(reported string) bool {
	for _, opening := range []string{"var", "let", "const", "using", "await", "export", "declare"} {
		if strings.HasPrefix(reported, opening) {
			return true
		}
	}
	return false
}

func TestOneVarRepairsWhatUpstreamRepairs(t *testing.T) {
	t.Parallel()
	for _, testCase := range oneVarRepairCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			rule_testing.ExpectFixedSource(t, result, testCase.wanted)
		})
	}
}

// TestOneVarDeclinesTheRepairsUpstreamDeclines asserts the absence of a fix.
//
// Not "the output equals the input": `ExpectFixedSource` fatals when no fix was proposed, so
// routing a decline through it would make the assertion agree with any expectation at all. The
// decline withholds the REPAIR, not the report, so the finding must still be there.
func TestOneVarDeclinesTheRepairsUpstreamDeclines(t *testing.T) {
	t.Parallel()
	for _, testCase := range oneVarDeclineCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			if len(result.Diagnostics) == 0 {
				t.Fatalf("expected a finding; a decline withholds the FIX, not the report")
			}
			for _, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("proposed %d fix(es); upstream declines this case", len(diagnostic.Fixes))
				}
			}
		})
	}
}

// oneVarTypeScriptGapCases are shapes upstream's corpus cannot contain, measured against the
// installed rule rather than reasoned about.
//
// Upstream's corpus is JavaScript of an era before class methods were common, and it writes ZERO
// class methods, constructors, getters or setters -- every "method-like" case in it is a static
// block. ESTree wraps a method body in a FunctionExpression, so upstream's single listener covers
// all of them; our parser has no wrapper, so each member kind is a function scope only because
// this rule names it. Deleting those four names is a mutation the whole 296-case corpus does not
// notice, which is what these rows exist to fix.
//
// The namespace rows are the other half of the same gap and fall the OTHER way: two sibling
// namespaces are not separate `var` scopes upstream, because a namespace body is not a function
// scope. So `ModuleBlock` is a statement list here and deliberately not a function scope.
//
// Every verdict below was produced by driving ESLint 10.10.0's own one-var over the source with
// the TypeScript parser, alongside controls that fire and controls that stay clean.
var oneVarTypeScriptGapCases = []oneVarCase{
	// Each member kind is its own `var` scope, so the second declaration is not a duplicate.
	{name: "gap-method-pair", source: "class C { a() { var x; } b() { var y; } }", options: "\"always\""},
	{name: "gap-ctor-pair", source: "class C { constructor() { var x; } m() { var y; } }", options: "\"always\""},
	{name: "gap-getter-pair", source: "class C { get a() { var x; return 1; } get b() { var y; return 2; } }", options: "\"always\""},
	{name: "gap-setter-pair", source: "class C { set a(v) { var x; } set b(v) { var y; } }", options: "\"always\""},
	{name: "gap-object-method-pair", source: "var o = { a() { var x; }, b() { var y; } };", options: "\"always\""},
	{name: "gap-method-let-pair", source: "class C { a() { let x; } b() { let y; } }", options: "\"always\""},
}

// oneVarTypeScriptGapFiringCases are the same gap from the reporting side.
//
// A clean row alone cannot tell a working scope from a rule that reports nothing at all, so each
// shape above has a partner here that must fire.
var oneVarTypeScriptGapFiringCases = []oneVarCase{
	{name: "gap-method-inner", source: "class C { a() { var x; var y; } }", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "gap-method-inner-let", source: "class C { a() { let x; let y; } }", options: "\"always\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'let' statement."}},
	{name: "gap-method-consecutive", source: "class C { a() { var x; var y; } }", options: "\"consecutive\"",
		ids:      []string{"combine"},
		messages: []string{"Combine this with the previous 'var' statement."}},
	{name: "gap-method-never", source: "class C { a() { var x, y; } }", options: "\"never\"",
		ids:      []string{"split"},
		messages: []string{"Split 'var' declarations into multiple statements."}},
	// A namespace body is a statement list but NOT a function scope, so these DO report.
	{name: "gap-namespace-pair", source: "namespace N { const a = 1; } namespace M { const b = 2; }",
		options: "\"always\"", ids: []string{"combine"},
		messages: []string{"Combine this with the previous 'const' statement."}},
	{name: "gap-namespace-inner", source: "namespace N { const a = 1; const b = 2; }",
		options: "\"always\"", ids: []string{"combine"},
		messages: []string{"Combine this with the previous 'const' statement."}},
}

func TestOneVarStaysSilentOnTypeScriptShapesUpstreamCannotWrite(t *testing.T) {
	t.Parallel()
	for _, testCase := range oneVarTypeScriptGapCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestOneVarFiresOnTypeScriptShapesUpstreamCannotWrite(t *testing.T) {
	t.Parallel()
	for _, testCase := range oneVarTypeScriptGapFiringCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.ids...)
			if len(result.Diagnostics) == len(testCase.messages) {
				for index, diagnostic := range result.Diagnostics {
					if !strings.HasPrefix(diagnostic.Message.Description, testCase.messages[index]) {
						t.Errorf("finding %d rendered:\n  %q\nwant it to begin with:\n  %q",
							index, diagnostic.Message.Description, testCase.messages[index])
					}
				}
			}
		})
	}
}

// TestOneVarSkipsDeclarationKindsTheConfigDidNotName pins the unconfigured-kind contract.
//
// Upstream's `if (!options[key]) return;` is what makes `{ "var": "never" }` govern `var` alone
// and leave `let` and `const` untouched. That guard is redundant in this port today, because every
// read of a group goes through `modeIs` and nil answers false there -- so a mutation removing the
// guard SURVIVES the whole 296-case corpus.
//
// The behaviour is still a contract, and this test asserts it where the guard cannot be seen: a
// change to `modeIs` that made nil compare equal would break these rows immediately, which is
// exactly the failure the guard exists to prevent.
func TestOneVarSkipsDeclarationKindsTheConfigDidNotName(t *testing.T) {
	t.Parallel()
	cases := []oneVarCase{
		// `never` on var alone: the let list has two declarators and must stay clean.
		{name: "var-never-leaves-let", source: "let foo = true, bar = false;", options: "{\"var\":\"never\"}"},
		// `never` on const alone: same shape, different unnamed kind.
		{name: "const-never-leaves-let", source: "let foo = true, bar = false;", options: "{\"const\":\"never\"}"},
		// `always` on var alone: two separate const statements must stay clean.
		{name: "var-always-leaves-const", source: "const a = 1; const b = 2;", options: "{\"var\":\"always\"}"},
		// And the using pair, which upstream added last and which no other test here reaches.
		{name: "var-never-leaves-using", source: "using a = f(), b = g();", options: "{\"var\":\"never\"}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}

	// A control, so the four clean rows above cannot pass by the rule being silent everywhere.
	t.Run("control-the-named-kind-does-report", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, OneVar, oneVarFile,
			"var foo = true, bar = false;", decodedOneVar(t, "{\"var\":\"never\"}"))
		rule_testing.ExpectFindings(t, result, "split")
	})
}

// oneVarExportCase is one exported-declaration row.
type oneVarExportCase struct {
	name     string
	source   string
	options  string
	findings int
	// fixes is how many repairs upstream proposes. Zero means upstream reports and declines,
	// which for these rows is the whole point.
	fixes int
}

// TestOneVarDeclinesToRepairExportedDeclarations pins a divergence the corpus cannot reach.
//
// Upstream reaches an exported declaration's siblings through `declaration.parent.parent.body`,
// and for an export `declaration.parent` is the ExportNamedDeclaration, which carries no `body`
// array. So the consecutive check sees index 0 and the join fixer sees an empty list: upstream
// neither reports a consecutive finding on an exported declaration nor repairs a combine on one,
// and an exported PREVIOUS sibling fails `previousNode.kind === type` because a wrapper node has
// no `kind`.
//
// Our parser has no wrapper. The exported statement sits directly in the enclosing statement list
// with real siblings and a real index, so without the collapse this rule reported where upstream
// is silent and repaired where upstream declines. Measured before the fix, the repair produced
//
//	export const a = 1; export const b = 2;   ->   export const a = 1; export,  b = 2;
//
// which is not valid syntax, and
//
//	export const a = 1; const b = 2;          ->   export const a = 1,  b = 2;
//
// which parses and moves a binding into the module's public surface.
//
// Every expectation below was produced by driving ESLint 10.10.0's own one-var, applying its
// fixes in a single pass exactly as RuleTester records `output`. The corpus is blind to all of it:
// all seven of its exported cases are under "never", which reaches the split fixer and never the
// join.
func TestOneVarDeclinesToRepairExportedDeclarations(t *testing.T) {
	t.Parallel()
	cases := []oneVarExportCase{
		{name: "export-join", source: "export const a = 1; export const b = 2;",
			options: "\"always\"", findings: 1, fixes: 0},
		{name: "export-join-consecutive", source: "export const a = 1; export const b = 2;",
			options: "\"consecutive\"", findings: 0, fixes: 0},
		{name: "export-then-plain", source: "export const a = 1; const b = 2;",
			options: "\"always\"", findings: 1, fixes: 0},
		{name: "plain-then-export", source: "const a = 1; export const b = 2;",
			options: "\"always\"", findings: 1, fixes: 0},
		// Controls, so the zero-fix rows above cannot pass by the fixer being dead everywhere.
		{name: "control-plain-join-repairs", source: "const a = 1; const b = 2;",
			options: "\"always\"", findings: 1, fixes: 2},
		{name: "control-different-kind-repairs", source: "let a = 1; const b = 2; const c = 3;",
			options: "\"always\"", findings: 1, fixes: 2},
		// A non-declaration previous sibling also declines, which is the same mechanism reached
		// through `previousNode.type !== "VariableDeclaration"` rather than through the wrapper.
		{name: "previous-not-a-declaration", source: "const a = 1; foo(); const b = 2;",
			options: "\"always\"", findings: 1, fixes: 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			if len(result.Diagnostics) != testCase.findings {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), testCase.findings)
			}
			proposed := 0
			for _, diagnostic := range result.Diagnostics {
				proposed += len(diagnostic.Fixes)
			}
			if proposed != testCase.fixes {
				t.Errorf("proposed %d fix(es), want %d", proposed, testCase.fixes)
			}
		})
	}
}

// TestOneVarSplitStillRepairsExports is the other half, and it is why the collapse above is scoped
// to the JOIN rather than applied to exports generally.
//
// Upstream's split fixer reads `declaration.parent.type === "ExportNamedDeclaration"` to decide
// whether to write an `export ` prefix, so it very much does repair an exported declaration under
// "never". Seven corpus cases pin it. A collapse that turned off every repair for exports would
// pass every join test above and break all seven.
func TestOneVarSplitStillRepairsExports(t *testing.T) {
	t.Parallel()
	result := rule_testing.RunWithOptions(t, OneVar, oneVarFile,
		"export const foo=1, bar=2;", decodedOneVar(t, "\"never\""))
	rule_testing.ExpectFindings(t, result, "split")
	rule_testing.ExpectFixedSource(t, result, "export const foo=1; export const bar=2;")
}

// TestOneVarCarriesDeclareThroughTheSplit pins the `declare` modifier on every statement a split
// creates.
//
// Both rows were broken, and both the same way upstream's installed build is broken: the split
// repeated `export` and dropped `declare`, so `declare let a: number, b: number;` became
// `declare let a: number; let b: number;`, a binding that now exists at runtime, and the exported
// const form became `export const b: number;`, which does not compile. The third row is the
// control that the plain split is untouched.
func TestOneVarCarriesDeclareThroughTheSplit(t *testing.T) {
	t.Parallel()
	for _, testCase := range []oneVarFixCase{
		{name: "declare", source: "declare let a: number, b: number;",
			options: "\"never\"", wanted: "declare let a: number; declare let b: number;"},
		{name: "export-declare", source: "export declare const a: number, b: number;",
			options: "\"never\"", wanted: "export declare const a: number; export declare const b: number;"},
		{name: "control-plain", source: "let a: number, b: number;",
			options: "\"never\"", wanted: "let a: number; let b: number;"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			rule_testing.ExpectFindings(t, result, "split")
			rule_testing.ExpectFixedSource(t, result, testCase.wanted)
		})
	}
}

// TestOneVarDeclinesToJoinAmbientDeclarations pins the join's decline on `declare`.
//
// The join turned `declare let a: number; declare let b: number;` into `declare let a: number;
// declare,  b: number;`, because the token before the second keyword is that statement's own
// `declare`, not the first statement's semicolon. It is still reported; it is no longer repaired.
// A mixed pair declines too, since joining an ambient binding into a real one changes whether it
// exists at runtime. The control shows the join is otherwise alive.
func TestOneVarDeclinesToJoinAmbientDeclarations(t *testing.T) {
	t.Parallel()
	cases := []oneVarExportCase{
		{name: "both-ambient", source: "declare let a: number;\ndeclare let b: number;",
			options: "\"always\"", findings: 1, fixes: 0},
		{name: "ambient-then-plain", source: "declare let a: number;\nlet b: number;",
			options: "\"always\"", findings: 1, fixes: 0},
		{name: "plain-then-ambient", source: "let a: number;\ndeclare let b: number;",
			options: "\"always\"", findings: 1, fixes: 0},
		{name: "control-plain-join-repairs", source: "let a: number;\nlet b: number;",
			options: "\"always\"", findings: 1, fixes: 2},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			if len(result.Diagnostics) != testCase.findings {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), testCase.findings)
			}
			proposed := 0
			for _, diagnostic := range result.Diagnostics {
				proposed += len(diagnostic.Fixes)
			}
			if proposed != testCase.fixes {
				t.Errorf("proposed %d fix(es), want %d", proposed, testCase.fixes)
			}
		})
	}
}

// TestOneVarJoinDeclinesAgainstADifferentKindPredecessor pins the fixer's own same-kind check.
//
// The check appears twice in upstream: once in the CONSECUTIVE arm, which tests
// `previousNode.kind === type` before reporting, and once inside `joinDeclarationsFixer`, which
// tests it again before yielding anything. The second looks redundant beside the first and is not,
// because the `always` arm calls the same fixer WITHOUT any such test -- it reports on what the
// enclosing scope has accumulated, which need not be the immediately preceding statement.
//
// The isolating shape is a reporting statement whose predecessor is a different declaration kind:
//
//	function f() { var a = 1; let b = 2; var c = 3; }
//
// The third statement reports a combine, because the function scope already holds a `var`. Its
// immediate predecessor is a `let`, so upstream's fixer declines and records `output: null`.
// Measured against ESLint 10.10.0 alongside three controls that do repair.
//
// A mutation dropping the fixer's own check survives the entire 296-case corpus, because no corpus
// case puts a different-kind statement between two same-kind ones under `always`.
func TestOneVarJoinDeclinesAgainstADifferentKindPredecessor(t *testing.T) {
	t.Parallel()

	t.Run("declines", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, OneVar, oneVarFile,
			"function f() { var a = 1; let b = 2; var c = 3; }", decodedOneVar(t, "\"always\""))
		rule_testing.ExpectFindings(t, result, "combine")
		for _, diagnostic := range result.Diagnostics {
			if len(diagnostic.Fixes) != 0 {
				t.Errorf("proposed %d fix(es); upstream declines when the predecessor is another kind",
					len(diagnostic.Fixes))
			}
		}
	})

	// Controls, so the decline above cannot pass by the join fixer being dead. Each has a
	// same-kind predecessor and must repair.
	for _, control := range []oneVarFixCase{
		{name: "control-let-then-var", source: "function f() { let a = 1; var b = 2; var c = 3; }",
			options: "\"always\"", wanted: "function f() { let a = 1; var b = 2,  c = 3; }"},
		{name: "control-const-then-let", source: "function f() { const a = 1; let b; let c; }",
			options: "\"always\"", wanted: "function f() { const a = 1; let b,  c; }"},
		{name: "control-same-kind", source: "function f() { var a = 1; var b = 2; }",
			options: "\"always\"", wanted: "function f() { var a = 1,  b = 2; }"},
	} {
		t.Run(control.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, control.source,
				decodedOneVar(t, control.options))
			rule_testing.ExpectFixedSource(t, result, control.wanted)
		})
	}
}

// TestOneVarRequireIsASpellingNotAResolution pins what `isRequire` actually asks.
//
// Upstream reads `decl.init.callee.name === "require"`, which is a NAME test on the callee rather
// than a resolution. Two consequences follow, and no corpus case can see either, because all ten
// of its `separateRequires` cases call `require` and nothing else:
//
//   - any other call is not a require, so a mixed list containing one is ordinary;
//   - a LOCALLY declared function named `require` still counts, because nothing resolves it.
//
// The callee-kind guard is the other half. `callee.name` is `undefined` for a member access, and
// `undefined === "require"` is false, so `a.require('foo')` is not a require either. Reproducing
// that without dereferencing something that has no name is why the port tests the callee's kind
// before its text.
//
// A mutation making `isRequire` answer true for every call survives the whole corpus. These rows
// are what kill it.
func TestOneVarRequireIsASpellingNotAResolution(t *testing.T) {
	t.Parallel()
	separateRequires := "{\"separateRequires\":true,\"var\":\"always\"}"
	cases := []oneVarCase{
		{name: "require-mixed", source: "var foo = require('foo'), bar;",
			options: separateRequires, ids: []string{"splitRequires"}},
		{name: "other-call-mixed", source: "var foo = notRequire('foo'), bar;",
			options: separateRequires, ids: nil},
		{name: "member-call-mixed", source: "var foo = a.require('foo'), bar;",
			options: separateRequires, ids: nil},
		{name: "local-require-still-counts",
			source:  "function require(x) { return x; } var foo = require('foo'), bar;",
			options: separateRequires, ids: []string{"splitRequires"}},
		{name: "both-requires", source: "var foo = require('foo'), bar = require('bar');",
			options: separateRequires, ids: nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			if len(testCase.ids) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

// TestOneVarConsecutiveNeedsUpstreamsParentBody pins where the sibling index exists.
//
// Upstream's consecutive arm reads `(parent.body && parent.body.length > 0 &&
// parent.body.indexOf(node)) || 0`, so it fires only where the container has a `body` ARRAY. An
// ESTree SwitchCase keeps its statements in `.consequent`, so `parent.body` is undefined and the
// consecutive arm never fires inside a switch clause -- while `always` and `never`, which read the
// scope stack rather than the index, still do.
//
// Our parser calls a switch clause a statement list, because it is one, so this rule reported
// `combine` there until the divergence was measured. The corpus writes three switch cases and none
// under `consecutive`, so nothing in it could see the false positive.
//
// A TypeScript namespace body falls the other way: upstream's TSModuleBlock does carry `body`, so
// the consecutive arm DOES fire there. Both directions are pinned, because a fix that turned the
// arm off for every unusual container would pass the switch rows and break the namespace ones.
//
// Every expectation measured against ESLint 10.10.0.
func TestOneVarConsecutiveNeedsUpstreamsParentBody(t *testing.T) {
	t.Parallel()
	silent := []oneVarCase{
		{name: "case-consecutive", source: "switch (x) { case 1: let a = 1; let b = 2; }",
			options: "\"consecutive\""},
		{name: "default-consecutive", source: "switch (x) { default: let a = 1; let b = 2; }",
			options: "\"consecutive\""},
		{name: "case-nonconsecutive", source: "switch (x) { case 1: let a = 1; foo(); let b = 2; }",
			options: "\"consecutive\""},
		{name: "namespace-nonconsecutive", source: "namespace N { const a = 1; foo(); const b = 2; }",
			options: "\"consecutive\""},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}

	firing := []oneVarCase{
		// A namespace body DOES carry `body` upstream, so consecutive fires there.
		{name: "namespace-consecutive", source: "namespace N { const a = 1; const b = 2; }",
			options: "\"consecutive\"", ids: []string{"combine"}},
		// The same switch clause still reports under the two arms that do not read the index,
		// which is what separates "this container has no sibling list" from "the rule went quiet".
		{name: "case-always", source: "switch (x) { case 1: let a = 1; let b = 2; }",
			options: "\"always\"", ids: []string{"combine"}},
		{name: "case-never", source: "switch (x) { case 1: let a = 1, b = 2; }",
			options: "\"never\"", ids: []string{"split"}},
		// And the block control, so the silent rows cannot pass by consecutive being dead.
		{name: "control-block-consecutive", source: "function f() { let a = 1; let b = 2; }",
			options: "\"consecutive\"", ids: []string{"combine"}},
	}
	for _, testCase := range firing {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OneVar, oneVarFile, testCase.source,
				decodedOneVar(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}
