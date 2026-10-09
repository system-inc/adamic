package core

// Generated from ESLint's own no-unmodified-loop-condition corpus and verified byte for byte
// against it.
//
// The cases were loaded by running upstream's tester file with its RuleTester stubbed, so they are
// the file's own values rather than retyped, and each was driven through the installed eslint
// 10.8.1 build to record what it reports.
//
// # Three cases are not here, and the reason is a version gap rather than a judgment
//
// The clone is at 10.9.1 and adds a `checkConditionalExpressions` option. The installed build this
// repository runs is 10.8.1, whose `schema` is empty: configuring that option throws at config
// validation with "should NOT have more than 0 items". So those three cases have no oracle verdict
// to import, and they are listed in unmodifiedLoopOptionOnlyCases with what upstream's own tester
// states rather than folded into either table.

// unmodifiedLoopFiringCases are the corpus inputs upstream reports on, with the variable names it
// names, in the order it names them. A case reporting twice lists two.
var unmodifiedLoopFiringCases = []struct {
	source    string
	variables []string
}{
	{"var foo = 0; while (foo) { } foo = 1;", []string{"foo"}},
	{"var foo = 0; while (!foo) { } foo = 1;", []string{"foo"}},
	{"var foo = 0; while (foo != null) { } foo = 1;", []string{"foo"}},
	{"var foo = 0, bar = 9; while (foo < bar) { } foo = 1;", []string{"foo", "bar"}},
	{"var foo = 0, bar = 0; while (foo && bar) { ++bar; } foo = 1;", []string{"foo"}},
	{"var foo = 0, bar = 0; while (foo && bar) { ++foo; } foo = 1;", []string{"bar"}},
	{"var a, b, c; while (a < c && b < c) { ++a; } foo = 1;", []string{"b", "c"}},
	{"var foo = 0; while (foo ? 1 : 0) { } foo = 1;", []string{"foo"}},
	{"var foo = 0; while (foo) { update(); } function update(foo) { ++foo; }", []string{"foo"}},
	{"var foo; do { } while (foo);", []string{"foo"}},
	{"for (var foo = 0; foo < 10; ) { } foo = 1;", []string{"foo"}},
}

// unmodifiedLoopCleanCases are the corpus inputs upstream stays silent on. These carry the whole
// discrimination: a condition that modifies its own variable, a call or member access that could do
// anything, a group where one member changes, and a write reached through a called function.
var unmodifiedLoopCleanCases = []string{
	"var foo = 0; while (foo) { ++foo; }",
	"let foo = 0; while (foo) { ++foo; }",
	"var foo = 0; while (foo) { foo += 1; }",
	"var foo = 0; while (foo++) { }",
	"var foo = 0; while (foo = next()) { }",
	"var foo = 0; while (ok(foo)) { }",
	"var foo = 0, bar = 0; while (++foo < bar) { }",
	"var foo = 0, obj = {}; while (foo === obj.bar) { }",
	"var foo = 0, f = {}, bar = {}; while (foo === f(bar)) { }",
	"var foo = 0, f = {}; while (foo === f()) { }",
	"var foo = 0, tag = 0; while (foo === tag`abc`) { }",
	"function* foo() { var foo = 0; while (yield foo) { } }",
	"function* foo() { var foo = 0; while (foo === (yield)) { } }",
	"var foo = 0; while (foo.ok) { }",
	"var foo = 0; while (foo) { update(); } function update() { ++foo; }",
	"var foo = 0, bar = 9; while (foo < bar) { foo += 1; }",
	"var foo = 0, bar = 1, baz = 2; while (foo ? bar : baz) { foo += 1; }",
	"var foo = 0, bar = 0; while (foo && bar) { ++foo; ++bar; }",
	"var foo = 0, bar = 0; while (foo || bar) { ++foo; ++bar; }",
	"var foo = 0; do { ++foo; } while (foo);",
	"var foo = 0; do { } while (foo++);",
	"for (var foo = 0; foo; ++foo) { }",
	"for (var foo = 0; foo;) { ++foo }",
	"var foo = 0, bar = 0; for (bar; foo;) { ++foo }",
	"var foo; if (foo) { }",
	"var a = [1, 2, 3]; var len = a.length; for (var i = 0; i < len - 1; i++) {}",
}

// unmodifiedLoopOptionOnlyCases are the three cases requiring an option the installed build does not
// have. `statedFindings` is what upstream's own tester asserts under 10.9.1, recorded so the gap is
// visible rather than silent. Nothing asserts these today.
var unmodifiedLoopOptionOnlyCases = []struct {
	source         string
	statedFindings int
}{
	{"let foo = 0, bar = 1, baz = 2; while (foo ? bar : baz) { foo += 1; bar += 1; baz += 1; }", 0},
	{"let foo = 0, bar = 1, baz = 2; while (foo ? bar : baz) { foo += 1; }", 2},
	{"let chunk = true, done = false; while (chunk ? !done : false) { chunk = nextOrNull(); }", 1},
}
