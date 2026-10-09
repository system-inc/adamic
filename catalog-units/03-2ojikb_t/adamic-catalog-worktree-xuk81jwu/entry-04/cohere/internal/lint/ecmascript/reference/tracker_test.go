package reference_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// trackedIn runs the tracker over a typed program and renders what it found as "Usage path text", in
// the order it returned them.
func trackedIn(t *testing.T, code string, traceMap map[string]*reference.TraceMap) []string {
	t.Helper()
	var rendered []string
	rule_testing.RunTyped(t, rule.Rule{
		Name:             "tracker-probe",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					tracker := reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)
					source := ctx.SourceFile.Text()
					for _, tracked := range tracker.GlobalReferences(traceMap) {
						span := rule.TokenRange(ctx.SourceFile, tracked.Node)
						usage := map[reference.Usage]string{reference.Read: "Read", reference.Call: "Call", reference.Construct: "Construct"}[tracked.Usage]
						rendered = append(rendered, usage+" "+strings.Join(tracked.Path, ".")+" "+source[span.Pos():span.End()])
					}
				},
			}
		},
	}, "input.ts", code)
	return rendered
}

// decodeTraceMap reads a trace map written as JSON, with `$read`, `$call` and `$construct` standing
// for eslint-utils' READ, CALL and CONSTRUCT symbols.
func decodeTraceMap(t *testing.T, raw string) map[string]*reference.TraceMap {
	t.Helper()
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("trace map %s: %v", raw, err)
	}
	members := map[string]*reference.TraceMap{}
	for key, value := range decoded {
		members[key] = decodeTraceNode(t, value)
	}
	return members
}

func decodeTraceNode(t *testing.T, raw json.RawMessage) *reference.TraceMap {
	t.Helper()
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("trace node %s: %v", raw, err)
	}
	node := &reference.TraceMap{Members: map[string]*reference.TraceMap{}}
	for key, value := range decoded {
		switch key {
		case "$read":
			node.Read = true
		case "$call":
			node.Call = true
		case "$construct":
			node.Construct = true
		default:
			node.Members[key] = decodeTraceNode(t, value)
		}
	}
	return node
}

/*
 * eslint-utils 4.10.1's own tests for ReferenceTracker.iterateGlobalReferences, every case in
 * test/reference-tracker.mjs, run through the installed ReferenceTracker under @typescript-eslint/parser
 * with the test file's own configuration (ecmaVersion 2022, a module, `Reflect` declared global), and
 * each answer written here with the text of the node it names. Under that parser every case still
 * answers exactly what the test file expects.
 *
 * The answers are listed in the order this tracker returns them, by where each node starts, rather
 * than eslint-utils' yield order, which no rule depends on: ESLint sorts what a rule reports by
 * position before anyone reads it.
 *
 * `/*global window *\/` and its siblings declare a global to ESLint. Here the harness's lib has no DOM
 * or Node types, so the name resolves to nothing, which reads as the global for the reason the
 * tracker's own doc gives.
 */
func TestTrackerAgreesWithEslintUtils(t *testing.T) {
	t.Parallel()

	cases := []struct {
		description string
		code        string
		traceMap    string
		want        []string
	}{
		{
			description: "should iterate the references of a given global variable.",
			code:        "var x = Object; { let Object; var y = Object }",
			traceMap:    "{\"Object\":{\"foo\":{\"$call\":true},\"Foo\":{\"$construct\":true},\"$read\":true}}",
			want:        []string{"Read Object Object"},
		},
		{
			description: "should iterate the member references of a given global variable, with MemberExpression",
			code:        "Object.a; Object.a(); new Object.a();\nObject.b; Object.b(); new Object.b();\nObject.c; Object.c(); new Object.c();",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{"Read Object.a Object.a", "Read Object.a Object.a", "Read Object.a Object.a", "Call Object.b Object.b()", "Construct Object.c new Object.c()"},
		},
		{
			description: "should iterate the member references of a given global variable, with VariableDeclarator",
			code:        "var x = Object;\nx.a; x.a(); new x.a();\nx.b; x.b(); new x.b();\nx.c; x.c(); new x.c();",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{"Read Object.a x.a", "Read Object.a x.a", "Read Object.a x.a", "Call Object.b x.b()", "Construct Object.c new x.c()"},
		},
		{
			description: "should iterate the member references of a given global variable, with VariableDeclarator 2",
			code:        "var x = Object, a = x.a, b = x.b, c = x.c;\na; a(); new a();\nb; b(); new b();\nc; c(); new c();",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{"Read Object.a x.a", "Call Object.b b()", "Construct Object.c new c()"},
		},
		{
			description: "should iterate the member references of a given global variable, with AssignmentExpression",
			code:        "var x, a, b, c;\na = (x = Object).a; b = x.b; c = x.c;\na; a(); new a();\nb; b(); new b();\nc; c(); new c();",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{"Read Object.a (x = Object).a", "Call Object.b b()", "Construct Object.c new c()"},
		},
		{
			description: "should iterate the member references of a given global variable, with destructuring",
			code:        "var {a, b, c} = Object;\na; a(); new a();\nb; b(); new b();\nc; c(); new c();",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{"Read Object.a a", "Call Object.b b()", "Construct Object.c new c()"},
		},
		{
			description: "should iterate the member references of a given global variable, with AssignmentPattern",
			code:        "var {x: {a, b, c} = Object} = {};\na; a(); new a();\nb; b(); new b();\nc; c(); new c();",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{"Read Object.a a", "Call Object.b b()", "Construct Object.c new c()"},
		},
		{
			description: "should iterate the member references of a given global variable, with 'window'.",
			code:        "/*global window */\nvar {Object: {a, b, c}} = window;\na; a(); new a();\nb; b(); new b();\nc; c(); new c();",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{"Read Object.a a", "Call Object.b b()", "Construct Object.c new c()"},
		},
		{
			description: "should iterate the member references of a given global variable, with 'global'.",
			code:        "/*global global */\nglobal.Object.a;\nglobal.Object.b; global.Object.b(); new global.Object.b();\nglobal.Object.c; global.Object.c(); new global.Object.c();",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{"Read Object.a global.Object.a", "Call Object.b global.Object.b()", "Construct Object.c new global.Object.c()"},
		},
		{
			description: "should iterate the member references of a given global variable, with 'globalThis'.",
			code:        "/*global globalThis */\nglobalThis.Object.a;\nglobalThis.Object.b; globalThis.Object.b(); new globalThis.Object.b();\nglobalThis.Object.c; globalThis.Object.c(); new globalThis.Object.c();",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{"Read Object.a globalThis.Object.a", "Call Object.b globalThis.Object.b()", "Construct Object.c new globalThis.Object.c()"},
		},
		{
			description: "should iterate the member references of a given global variable, with 'self'.",
			code:        "/*global self */\nself.Object.a;\nself.Object.b; self.Object.b(); new self.Object.b();\nself.Object.c; self.Object.c(); new self.Object.c();",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{"Read Object.a self.Object.a", "Call Object.b self.Object.b()", "Construct Object.c new self.Object.c()"},
		},
		{
			description: "should iterate the member references of a given global variable, with 'window'.",
			code:        "/*global window */\nwindow.Object.a;\nwindow.Object.b; window.Object.b(); new window.Object.b();\nwindow.Object.c; window.Object.c(); new window.Object.c();",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{"Read Object.a window.Object.a", "Call Object.b window.Object.b()", "Construct Object.c new window.Object.c()"},
		},
		{
			description: "should not iterate the references of a given global variable if it's modified.",
			code:        "Object = {}\nObject.a\nObject.b()\nnew Object.c()",
			traceMap:    "{\"Object\":{\"a\":{\"$read\":true},\"b\":{\"$call\":true},\"c\":{\"$construct\":true}}}",
			want:        []string{},
		},
		{
			description: "should not iterate the references through unary/binary expressions.",
			code:        "var construct = typeof Reflect !== \"undefined\" ? Reflect.construct : undefined\nconstruct()",
			traceMap:    "{\"Reflect\":{\"$call\":true}}",
			want:        []string{},
		},
		{
			description: "should not mix up public and private identifiers.",
			code:        "class C { #value; wrap() { var value = MyObj.#value; } }",
			traceMap:    "{\"MyObj\":{\"value\":{\"$read\":true}}}",
			want:        []string{},
		},
	}

	for _, testCase := range cases {
		got := trackedIn(t, testCase.code, decodeTraceMap(t, testCase.traceMap))
		if strings.Join(got, "\n") != strings.Join(testCase.want, "\n") {
			t.Errorf("%s\n%s\n got: %q\nwant: %q", testCase.description, testCase.code, got, testCase.want)
		}
	}
}

// TestTrackerFollowsOnlyAssignmentsThatKeepTheValue pins the one place the tracker departs from
// eslint-utils. `x += JSON` leaves x a string, so x is not JSON; eslint-utils follows it anyway, and
// the installed no-obj-calls reports `x()` as a call of JSON. The logical assignments do keep the
// value, and so does a plain one.
func TestTrackerFollowsOnlyAssignmentsThatKeepTheValue(t *testing.T) {
	t.Parallel()

	traceMap := map[string]*reference.TraceMap{"JSON": {Call: true}}
	for _, code := range []string{"let x; x += JSON; x();", "let x = 1; x *= JSON; x();"} {
		if got := trackedIn(t, code, traceMap); len(got) != 0 {
			t.Errorf("%s: an arithmetic assignment does not make x hold JSON, got %q", code, got)
		}
	}
	for _, code := range []string{"let x; x = JSON; x();", "let x; x ||= JSON; x();", "let x; x ??= JSON; x();", "let x; x &&= JSON; x();"} {
		if got := trackedIn(t, code, traceMap); len(got) != 1 {
			t.Errorf("%s: x may hold JSON, so its call is one, got %q", code, got)
		}
	}
}

// TestTrackerTerminatesOnACycle pins the guard that keeps a binding's reads from being followed into
// themselves, which is eslint-utils' variableStack. Without it each of these recurses until the stack
// runs out.
func TestTrackerTerminatesOnACycle(t *testing.T) {
	t.Parallel()

	traceMap := map[string]*reference.TraceMap{"JSON": {Call: true}}
	cases := map[string][]string{
		"let a = JSON; a = a; a();":                        {"Call JSON a()"},
		"let a = JSON, b = a; a = b; b();":                 {"Call JSON b()"},
		"export const getConfig = getConfig; getConfig();": nil,
	}
	for code, want := range cases {
		if got := trackedIn(t, code, traceMap); strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s: got %q, want %q", code, got, want)
		}
	}
}

// TestTrackerDoesNotFollowAWrittenGlobal pins isModifiedGlobal: a global this file writes to is not
// followed anywhere in it, and neither is a global object it writes to. Writing a member of the global
// object is not writing the global object.
func TestTrackerDoesNotFollowAWrittenGlobal(t *testing.T) {
	t.Parallel()

	traceMap := map[string]*reference.TraceMap{"JSON": {Call: true}}
	cases := map[string]int{
		"JSON = 1; JSON();":                       0,
		"JSON(); [JSON] = [1];":                   0,
		"globalThis = 1; globalThis.JSON();":      0,
		"globalThis.JSON = 1; globalThis.JSON();": 1,
		"const JSON = 1; JSON();":                 0,
		"function f(JSON) { JSON(); } JSON();":    1,
	}
	for code, want := range cases {
		if got := trackedIn(t, code, traceMap); len(got) != want {
			t.Errorf("%s: got %q, want %d", code, got, want)
		}
	}
}

// TestConstantString pins the folding a computed key gets, which is getStringIfConstant's without a
// scope.
func TestConstantString(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`x["JSON"]`:        "JSON",
		"x[`JSON`]":        "JSON",
		`x["JS" + "ON"]`:   "JSON",
		"x[`J${\"S\"}ON`]": "JSON",
		`x[1 + "x"]`:       "1x",
		`x[0x10]`:          "16",
		`x[true]`:          "true",
		`x[null]`:          "null",
		`x[1 + 2]`:         "",
		`x[y]`:             "",
		"x[`J${y}`]":       "",
	}
	for code, want := range cases {
		var got string
		var settled bool
		rule_testing.Run(t, rule.Rule{
			Name: "constant-probe",
			Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{
					ast.KindElementAccessExpression: func(node *ast.Node) {
						got, settled = reference.ConstantString(node.AsElementAccessExpression().ArgumentExpression)
					},
				}
			},
		}, "input.ts", code+";")
		if want == "" && settled {
			t.Errorf("%s folded to %q, and it is not constant without a scope", code, got)
		}
		if want != "" && (!settled || got != want) {
			t.Errorf("%s folded to %q (%v), want %q", code, got, settled, want)
		}
	}
}

// TestConstantStringIn pins getStringIfConstant's reading with a scope: a binding counts when it has one
// declaration naming it alone, and is `const` or never written after its initializer. Each case
// evaluates the argument of `probe(...)`, and every answer is the installed eslint-utils 4.10.1's,
// read by running its getStringIfConstant with the scope over the same code under @typescript-eslint/parser.
func TestConstantStringIn(t *testing.T) {
	t.Parallel()

	cases := []struct {
		code     string
		want     string
		constant bool
	}{
		{`const a = 'x'; probe(a);`, "x", true},
		{`let a = 'x'; probe(a);`, "x", true},
		{`var a = 'x'; probe(a);`, "x", true},
		{`const a = 'x'; function f() { probe(a); }`, "x", true},
		{`const a = 'x', b = a + 'y'; probe(b);`, "xy", true},
		{"const a = 'x'; probe(`${a}y`);", "xy", true},
		{`probe(/a\1/g);`, `/a\1/g`, true},
		{`const a = /x/; probe(a + '');`, "/x/", true},
		{`let a = 'x'; a = 'y'; probe(a);`, "", false},
		{`let a = 'x'; function g() { a = 'z'; } probe(a);`, "", false},
		{`let a = 'x'; [a] = ['y']; probe(a);`, "", false},
		{`let a = 'x'; function g() { let a = 'q'; a = 'z'; } probe(a);`, "x", true},
		{`let a; a = 'x'; probe(a);`, "", false},
		{`const {a} = {a: 'x'}; probe(a);`, "", false},
		{`function f(a = 'x') { probe(a); }`, "", false},
		{`const a = b, b = a; probe(a);`, "", false},
		{`probe(undeclared);`, "", false},
	}
	for _, testCase := range cases {
		var got string
		var constant bool
		rule_testing.RunTyped(t, rule.Rule{
			Name:             "constant-in-probe",
			NeedsTypeChecker: true,
			Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{
					ast.KindCallExpression: func(node *ast.Node) {
						call := node.AsCallExpression()
						if call.Expression.Text() != "probe" {
							return
						}
						got, constant = reference.ConstantStringIn(ctx, call.Arguments.Nodes[0])
					},
				}
			},
		}, "input.ts", testCase.code)
		if constant != testCase.constant || got != testCase.want {
			t.Errorf("%s evaluated to %q (%v), want %q (%v)", testCase.code, got, constant, testCase.want, testCase.constant)
		}
	}
}

// TestIsConstantRegExpIn pins which arguments hold a RegExp object rather than a string: a regex
// literal, or a constant binding initialized with one, through any number of such bindings.
func TestIsConstantRegExpIn(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{
		`probe(/x/);`:                           true,
		`probe((/x/));`:                         true,
		`const r = /x/; probe(r);`:              true,
		`const r = /x/; const s = r; probe(s);`: true,
		`let r = /x/; probe(r);`:                true,
		`let r = /x/; r = /y/; probe(r);`:       false,
		`const r = "x"; probe(r);`:              false,
		`const r = /x/; probe(r + "");`:         false,
		`probe(new RegExp("x"));`:               false,
		`const a = b, b = a; probe(a);`:         false,
		`probe(undeclared);`:                    false,
	}
	for code, want := range cases {
		var got bool
		rule_testing.RunTyped(t, rule.Rule{
			Name:             "constant-regexp-probe",
			NeedsTypeChecker: true,
			Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{
					ast.KindCallExpression: func(node *ast.Node) {
						call := node.AsCallExpression()
						if call.Expression.Text() == "probe" {
							got = reference.IsConstantRegExpIn(ctx, call.Arguments.Nodes[0])
						}
					},
				}
			},
		}, "input.ts", code)
		if got != want {
			t.Errorf("%s: IsConstantRegExpIn is %v, want %v", code, got, want)
		}
	}
}
