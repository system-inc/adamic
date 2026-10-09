package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// undefInitFile is where the fixtures pretend to live.
const undefInitFile = "/repository/source/UndefInit.ts"

// The corpus is upstream's, extracted from its tester rather than retyped.
//
// `/tmp/lint-sources/eslint/tests/lib/rules/no-undef-init.js` carries 7 valid cases and 21 invalid
// ones, every invalid case naming exactly one `unnecessaryUndefinedInit` and every one of them
// carrying an `output`, of which nine are `null`. Rendered to Go literals by a script that reads
// that file through a stub RuleTester, so nothing was retyped.
//
// The clean list is small and every entry is a different reason. `var a;` has no initializer at all.
// The three constant bindings cannot omit one. `var undefined = 5;` shadows the name so the
// initializer is load-bearing. A class field is not a variable declaration. And the last is a
// conditional whose alternate is `undefined`, which is not the bare identifier the rule watches.
//
// This rule reads the checker, so its fixtures use `RunTyped`. Handed the plain harness the rule
// guards and goes completely silent, which would make every clean case below pass for the wrong
// reason.
func TestNoUndefInitStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"var a;"},
		{"const foo = undefined"},       // lang={"ecmaVersion":6}
		{"using foo = undefined"},       // lang={"ecmaVersion":2026}
		{"await using foo = undefined"}, // lang={"ecmaVersion":2026}
		{"var undefined = 5; var foo = undefined;"},
		{"class C { field = undefined; }"},                             // lang={"ecmaVersion":2022}
		{"using a = condition ? getDisposableResource() : undefined;"}, // lang={"ecmaVersion":2026,"sourceType":"module"}
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoUndefInit, undefInitFile, testCase.sourceText))
		})
	}
}

// Every reporting case names one `unnecessaryUndefinedInit`, which is upstream's own count.
func TestNoUndefInitFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"var a = undefined;"},               // errors=["unnecessaryUndefinedInit"] output=null
		{"var a = undefined, b = 1;"},        // errors=["unnecessaryUndefinedInit"] output=null
		{"var a = 1, b = undefined, c = 5;"}, // errors=["unnecessaryUndefinedInit"] output=null
		{"var [a] = undefined;"},             // errors=["unnecessaryUndefinedInit"] output=null lang={"ecmaVersion":6}
		{"var {a} = undefined;"},             // errors=["unnecessaryUndefinedInit"] output=null lang={"ecmaVersion":6}
		{"for(var i in [1,2,3]){var a = undefined; for(var j in [1,2,3]){}}"}, // errors=["unnecessaryUndefinedInit"] output=null
		{"let a = undefined;"},               // errors=["unnecessaryUndefinedInit"] output="let a;" lang={"ecmaVersion":6}
		{"let a = undefined, b = 1;"},        // errors=["unnecessaryUndefinedInit"] output="let a, b = 1;" lang={"ecmaVersion":6}
		{"let a = 1, b = undefined, c = 5;"}, // errors=["unnecessaryUndefinedInit"] output="let a = 1, b, c = 5;" lang={"ecmaVersion":6}
		{"let [a] = undefined;"},             // errors=["unnecessaryUndefinedInit"] output=null lang={"ecmaVersion":6}
		{"let {a} = undefined;"},             // errors=["unnecessaryUndefinedInit"] output=null lang={"ecmaVersion":6}
		{"for(var i in [1,2,3]){let a = undefined; for(var j in [1,2,3]){}}"}, // errors=["unnecessaryUndefinedInit"] output="for(var i in [1,2,3]){let a; for(var j in [1,2,3]){}}" lang={"ecmaVersion":6}
		{"let /* comment */a = undefined;"},                                   // errors=["unnecessaryUndefinedInit"] output="let /* comment */a;" lang={"ecmaVersion":6}
		{"let a/**/ = undefined;"},                                            // errors=["unnecessaryUndefinedInit"] output=null lang={"ecmaVersion":6}
		{"let a /**/ = undefined;"},                                           // errors=["unnecessaryUndefinedInit"] output=null lang={"ecmaVersion":6}
		{"let a//\n= undefined;"},                                             // errors=["unnecessaryUndefinedInit"] output=null lang={"ecmaVersion":6}
		{"let a = /**/undefined;"},                                            // errors=["unnecessaryUndefinedInit"] output=null lang={"ecmaVersion":6}
		{"let a = //\nundefined;"},                                            // errors=["unnecessaryUndefinedInit"] output=null lang={"ecmaVersion":6}
		{"let a = undefined/* comment */;"},                                   // errors=["unnecessaryUndefinedInit"] output="let a/* comment */;" lang={"ecmaVersion":6}
		{"let a = undefined/* comment */, b;"},                                // errors=["unnecessaryUndefinedInit"] output="let a/* comment */, b;" lang={"ecmaVersion":6}
		{"let a = undefined//comment\n, b;"},                                  // errors=["unnecessaryUndefinedInit"] output="let a//comment\n, b;" lang={"ecmaVersion":6}
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoUndefInit, undefInitFile, testCase.sourceText),
				"unnecessaryUndefinedInit")
		})
	}
}

// The annotation survives the repair.
//
// Upstream's corpus is JavaScript, so not one of its `output` cases carries a type annotation, and
// a port can pass all eight while destroying every annotated declaration in a TypeScript tree.
// This is what that looked like: `let x: Thing | undefined = undefined` fixed to `let x`, silently
// widening to `any`. It compiled, so nothing failed, and the only trace was the diff. Eight files
// in libraries/structure were rewritten this way before the range was corrected.
//
// The definite-assignment case is the same hazard one field over: `!` is a sibling of the name too,
// and dropping it changes what the compiler will accept.
func TestNoUndefInitKeepsTheTypeAnnotation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSource string
	}{
		{"let a: number | undefined = undefined;", "let a: number | undefined;"},
		{"let a: string = undefined;", "let a: string;"},
		{"let a: Array<{x: number}> | undefined = undefined;", "let a: Array<{x: number}> | undefined;"},
		{"let a: number | undefined = undefined, b = 1;", "let a: number | undefined, b = 1;"},
		{"let a = 1, b: string | undefined = undefined;", "let a = 1, b: string | undefined;"},
		{"let a!: number = undefined;", "let a!: number;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.RunTyped(t, NoUndefInit, undefInitFile, testCase.sourceText),
				strings.TrimSpace(testCase.wantSource)+"\n")
		})
	}
}

// The repair, asserted by applying it rather than by comparing the fix's text.
//
// A fix writing the right string over the wrong span passes a text comparison and is a real defect.
// These eight are every case upstream gives a non-null `output`, and the last four are the whole
// comment story: a comment before the binding name is outside the removal range and survives, and a
// comment after the initializer is also outside it and survives, while a comment between them
// blocks the fix entirely and appears in the declined list below.
func TestNoUndefInitFixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSource string
	}{
		{"let a = undefined;", "let a;"},
		{"let a = undefined, b = 1;", "let a, b = 1;"},
		{"let a = 1, b = undefined, c = 5;", "let a = 1, b, c = 5;"},
		{"for(var i in [1,2,3]){let a = undefined; for(var j in [1,2,3]){}}", "for(var i in [1,2,3]){let a; for(var j in [1,2,3]){}}"},
		{"let /* comment */a = undefined;", "let /* comment */a;"},
		{"let a = undefined/* comment */;", "let a/* comment */;"},
		{"let a = undefined/* comment */, b;", "let a/* comment */, b;"},
		{"let a = undefined//comment\n, b;", "let a//comment\n, b;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			// The harness writes each fixture as `TrimSpace(source)+"\n"`, so the file on disk
			// carries a trailing newline the upstream `output` string does not. Transforming the
			// expectation the same way the harness transforms the input is the honest fix; padding
			// the rule to emit one would be a rule change made to satisfy a harness detail.
			rule_testing.ExpectFixedSource(t,
				rule_testing.RunTyped(t, NoUndefInit, undefInitFile, testCase.sourceText),
				strings.TrimSpace(testCase.wantSource)+"\n")
		})
	}
}

// The thirteen cases upstream reports and deliberately declines to fix.
//
// `output: null` is a decision, not an omission, and a fixer that repairs a case upstream refuses to
// touch is a defect no message-id fixture can see. Three reasons are represented: a `var`
// declaration, where dropping the initializer changes what a re-reached loop body holds; a
// destructuring target, where `var [a] = undefined` throws and `var [a]` is a syntax error, so
// neither direction is a repair; and a comment sitting inside the range that would be removed.
//
// Asserted as "reports, and offers no fix", which is the pair. Asserting only the finding would pass
// for a rule that fixes all thirteen.
func TestNoUndefInitDeclinesToFix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"var a = undefined;"},
		{"var a = undefined, b = 1;"},
		{"var a = 1, b = undefined, c = 5;"},
		{"var [a] = undefined;"},
		{"var {a} = undefined;"},
		{"for(var i in [1,2,3]){var a = undefined; for(var j in [1,2,3]){}}"},
		{"let [a] = undefined;"},
		{"let {a} = undefined;"},
		{"let a/**/ = undefined;"},
		{"let a /**/ = undefined;"},
		{"let a//\n= undefined;"},
		{"let a = /**/undefined;"},
		{"let a = //\nundefined;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUndefInit, undefInitFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			if fixes := result.Diagnostics[0].Fixes; len(fixes) != 0 {
				t.Fatalf("upstream declines to fix this and we offered %d fix(es)", len(fixes))
			}
		})
	}
}

// The rendered message, which the id assertions above cannot see.
//
// The Description interpolates the binding name, and a `rule.Message` carries no rendering layer, so
// nothing between the format string and the reader checks it. A verb swapped into the name's slot
// produces a finding with the right id, the right span and a sentence naming the wrong thing.
func TestNoUndefInitNamesTheBinding(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoUndefInit, undefInitFile, "let someBinding = undefined;")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	description := result.Diagnostics[0].Message.Description
	if !strings.HasPrefix(description, "`someBinding` is initialized to `undefined`") {
		t.Fatalf("the message reads %q, which does not name the binding it reports", description)
	}
}

// The reported span, which no message-id and no fix fixture above can see.
//
// Upstream passes `node`, the whole declarator, so the finding covers `a = undefined` rather than
// just `a`. Measured on the installed rule: `let aLongName = undefined, b = 1;` reports columns 5
// through 26, which is exactly `aLongName = undefined`.
//
// A mutant reporting the name instead survived every one of the fifty-two fixtures above, because
// each of them asserts which message fired or what the repair wrote and none asserts where the
// finding points. A finding carrying a correct fix while pointing at the wrong span is the failure
// this closes: the edit lands somewhere the reader was never shown.
func TestNoUndefInitReportsTheWholeDeclarator(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpan   string
	}{
		{"let a = undefined;", "a = undefined"},
		{"let aLongName = undefined, b = 1;", "aLongName = undefined"},
		{"var a = 1, b = undefined, c = 5;", "b = undefined"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUndefInit, undefInitFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			// Sliced from the source the harness actually wrote rather than from the literal above,
			// because the harness trims and a slice of the literal is one byte off.
			written := result.SourceFile.Text()
			reported := written[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantSpan {
				t.Fatalf("reported %q, upstream spans %q", reported, testCase.wantSpan)
			}
		})
	}
}

// The shadow test, and why the shipped one in this package is wrong for this name.
//
// `no_new_native_nonconstructor.go` asks whether the symbol's first declaration lives in a
// declaration file, which is the right question for `Symbol` and `BigInt`. Measured with a probe:
// the real global `undefined` resolves to a symbol with ZERO declarations, so that predicate answers
// false for it and a rule built on it is silent on every reporting case. This asks the complement.
//
// Three shadow shapes, each a different binding kind, and all three confirmed silent on the
// installed ESLint build. The control is the same file shape with no shadow, which reports.
func TestNoUndefInitDeclinesAShadowedUndefined(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a parameter", "function f(undefined) { let foo = undefined; return foo; }"},
		{"a block-scoped let", "{ let undefined = 5; var foo = undefined; }"},
		{"a file-level var", "var undefined = 5; let foo = undefined;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUndefInit, undefInitFile, testCase.sourceText))
		})
	}

	// The control: the same shape with nothing shadowing the name, so the silence above is a
	// measurement about the shadow rather than about the surrounding code.
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUndefInit, undefInitFile,
		"function f() { let foo = undefined; return foo; }"), "unnecessaryUndefinedInit")
}

// An ambient declaration in a .d.ts is NOT a shadow, which only a multi-file fixture can see.
//
// The shadow test declines a declaration living in a declaration file, and that arm is invisible to
// every single-file fixture: measured with a probe, `undefined` gets ZERO declarations in ordinary
// source, and even `declare var undefined` written in a .ts file counts as a source declaration. A
// mutant dropping the `!IsDeclarationFile` half survived all fifty-five fixtures for exactly that
// reason, and it is a blind spot rather than an equivalence: with a real .d.ts in the program the
// symbol carries one declaration, in a declaration file, and the mutant then reads it as a shadow
// and goes silent.
//
// Reproducing upstream's decision is what settles which way it should fall. ESLint's scope analysis
// is per-file and cannot see another file's ambient declaration at all, so upstream reports
// `let a = undefined;` whatever a neighbouring .d.ts says. Driven on the installed rule to confirm.
// Our checker CAN see it, and fidelity is to the decision rather than to the mechanism, so an
// ambient declaration must not suppress the finding.
//
// The control is the same two-file shape with the declaration written in source instead, which is a
// real shadow and does suppress it.
func TestNoUndefInitIgnoresAnAmbientDeclaration(t *testing.T) {
	t.Parallel()

	reporting := map[string]string{
		"/repository/source/Ambient.d.ts": "declare var undefined: any;\n",
		"/repository/source/Use.ts":       "let a = undefined;\n",
	}
	rule_testing.ExpectFindings(t,
		rule_testing.RunTypedFiles(t, NoUndefInit, reporting, "/repository/source/Use.ts"),
		"unnecessaryUndefinedInit")

	// The control: a shadow written in source, in the same two-file program, is a real shadow.
	shadowed := map[string]string{
		"/repository/source/Ambient.d.ts": "declare var undefined: any;\n",
		"/repository/source/Use.ts":       "var undefined = 5;\nlet a = undefined;\n",
	}
	rule_testing.ExpectClean(t,
		rule_testing.RunTypedFiles(t, NoUndefInit, shadowed, "/repository/source/Use.ts"))
}

// The typed harness is required, so a later revert to `rule_testing.Run` fails loudly.
//
// The rule guards on a nil checker and returns, so under the plain harness it goes completely
// silent: every clean case passes vacuously and every reporting case fails in a way that reads as a
// rule defect rather than as a harness choice.
func TestNoUndefInitNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoUndefInit.NeedsTypeChecker {
		t.Fatal("this rule resolves `undefined` through the checker and must declare it")
	}
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUndefInit, undefInitFile, "let a = undefined;"))
}
