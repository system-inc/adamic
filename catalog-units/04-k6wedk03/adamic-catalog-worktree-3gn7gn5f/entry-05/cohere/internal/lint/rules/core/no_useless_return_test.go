package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is ESLint's own, 25 valid and 24 invalid cases carrying 25 findings and 21 fix
// vectors, extracted by loading its test file with a stubbed rule tester so nothing was retyped.
//
// Upstream runs the whole file at `ecmaVersion: 5` with `globalReturn`, so three of its cases are
// TOP-LEVEL returns. Our parser does not accept one, so those three are wrapped in a function here
// and marked at the line. The wrapping preserves what each case tests, since none of them is about
// being at the top level; they are about what follows the return.
//
// Three invalid cases carry `output: null`, and each is asserted as reporting AND offering no
// repair. Two of them are the comment guard, which is the most dangerous line in this rule: the
// fix deletes a statement, and without that guard it deletes a comment somebody wrote along with it.

// noUselessReturnDeclinesToFix marks a case upstream reports and deliberately does not repair.
const noUselessReturnDeclinesToFix = "\x00declines"

func TestNoUselessReturnStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{name: "valid0", source: "function foo() { return 5; }"},
		{name: "valid1", source: "function foo() { return null; }"},
		{name: "valid2", source: "function foo() { return doSomething(); }"},
		{name: "valid3", source: "\n          function foo() {\n            if (bar) {\n              doSomething();\n              return;\n            } else {\n              doSomethingElse();\n            }\n            qux();\n          }\n        "},
		{name: "valid4", source: "\n          function foo() {\n            switch (bar) {\n              case 1:\n                doSomething();\n                return;\n              default:\n                doSomethingElse();\n            }\n          }\n        "},
		{name: "valid5", source: "\n          function foo() {\n            switch (bar) {\n              default:\n                doSomething();\n                return;\n              case 1:\n                doSomethingElse();\n            }\n          }\n        "},
		{name: "valid6", source: "\n          function foo() {\n            switch (bar) {\n              case 1:\n                if (a) {\n                  doSomething();\n                  return;\n                } else {\n                  doSomething();\n                  return;\n                }\n              default:\n                doSomethingElse();\n            }\n          }\n        "},
		{name: "valid7", source: "\n          function foo() {\n            for (var foo = 0; foo < 10; foo++) {\n              return;\n            }\n          }\n        "},
		{name: "valid8", source: "\n          function foo() {\n            for (var foo in bar) {\n              return;\n            }\n          }\n        "},
		{name: "valid9", source: "\n          function foo() {\n            try {\n              return 5;\n            } finally {\n              return; // This is allowed because it can override the returned value of 5\n            }\n          }\n        "},
		{name: "valid10", source: "\n          function foo() {\n            try {\n              bar();\n              return;\n            } catch (err) {}\n            baz();\n          }\n        "},
		{name: "valid11", source: "\n          function foo() {\n              if (something) {\n                  try {\n                      bar();\n                      return;\n                  } catch (err) {}\n              }\n              baz();\n          }\n        "},
		{name: "valid12", source: "\n          function foo() {\n            return;\n            doSomething();\n          }\n        "},
		{name: "valid13", source: "\n              function foo() {\n                for (var foo of bar) return;\n              }\n            "},
		{name: "valid14", source: "() => { if (foo) return; bar(); }"},
		{name: "valid15", source: "() => 5"},
		{name: "valid16", source: "() => { return; doSomething(); }"},
		// Upstream writes this at the top level under ecmaVersion 5 with globalReturn.
		// Wrapped in a function here, which preserves the judgment: a return followed by
		// a statement is not useless.
		{name: "valid17", source: "function wrapper() { if (foo) { return; } doSomething(); }"},
		{name: "valid18", source: "\n          function foo() {\n            if (bar) return;\n            return baz;\n          }\n        "},
		{name: "valid19", source: "\n          function foo() {\n            if (bar) {\n              return;\n            }\n            return baz;\n          }\n        "},
		{name: "valid20", source: "\n          function foo() {\n            if (bar) baz();\n            else return;\n            return 5;\n          }\n        "},
		{name: "valid21", source: "\n          function foo() {\n            return;\n            while (foo) return;\n            foo;\n          }\n        "},
		{name: "valid22", source: "\n          try {\n            throw new Error('foo');\n            while (false);\n          } catch (err) {}\n        "},
		{name: "valid23", source: "\n          function foo(arg) {\n            throw new Error(\"Debugging...\");\n            if (!arg) {\n              return;\n            }\n            console.log(arg);\n          }\n        "},
		{name: "valid24", source: "\n        function foo() {\n          try {\n              bar();\n              return;\n          } finally {\n              baz();\n          }\n          qux();\n        }\n        "},
	}
	if len(cases) != 25 {
		t.Fatalf("REFUSING: the clean table holds %d cases, want 25", len(cases))
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUselessReturn, "file.ts", testCase.source))
		})
	}
}

func TestNoUselessReturnFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		source    string
		findings  int
		wantFixed string
	}{
		{name: "invalid0", source: "function foo() { return; }", findings: 1, wantFixed: "function foo() {  }"},
		{name: "invalid1", source: "function foo() { doSomething(); return; }", findings: 1, wantFixed: "function foo() { doSomething();  }"},
		{name: "invalid2", source: "function foo() { if (condition) { bar(); return; } else { baz(); } }", findings: 1, wantFixed: "function foo() { if (condition) { bar();  } else { baz(); } }"},
		{name: "invalid3", source: "function foo() { if (foo) return; }", findings: 1, wantFixed: noUselessReturnDeclinesToFix},
		{name: "invalid4", source: "function foo() { bar(); return/**/; }", findings: 1, wantFixed: noUselessReturnDeclinesToFix},
		{name: "invalid5", source: "function foo() { bar(); return//\n; }", findings: 1, wantFixed: noUselessReturnDeclinesToFix},
		// Upstream writes this at the top level under ecmaVersion 5 with globalReturn.
		// Wrapped in a function here; the judgment and the repair are unchanged.
		{name: "invalid6", source: "function wrapper() { foo(); return; }", findings: 1, wantFixed: "function wrapper() { foo();  }"},
		// Upstream writes this at the top level under ecmaVersion 5 with globalReturn.
		// Wrapped in a function here; the judgment and the repair are unchanged.
		{name: "invalid7", source: "function wrapper() { if (foo) { bar(); return; } else { baz(); } }", findings: 1, wantFixed: "function wrapper() { if (foo) { bar();  } else { baz(); } }"},
		// Upstream's recorded output keeps the second `return;`, and that is an artifact of its
		// FIX ENGINE rather than of the rule. Measured: it reports both returns and BOTH fixes
		// rewrite the whole function, range [0,47], because upstream wraps every repair in
		// FixTracker.retainEnclosingFunction to avoid conflicting with no-else-return. Two
		// overlapping fixes cannot both apply in one pass, so ESLint drops the second and its
		// corpus records the intermediate text. Run to a fixed point it produces exactly what
		// is asserted here, and so does our engine in one pass, since our fixes are the two
		// return statements themselves and do not overlap.
		{name: "invalid8", source: "\n              function foo() {\n                if (foo) {\n                  return;\n                }\n                return;\n              }\n            ", findings: 2, wantFixed: "\n              function foo() {\n                if (foo) {\n                  \n                }\n                \n              }\n            "},
		{name: "invalid9", source: "\n              function foo() {\n                switch (bar) {\n                  case 1:\n                    doSomething();\n                  default:\n                    doSomethingElse();\n                    return;\n                }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                switch (bar) {\n                  case 1:\n                    doSomething();\n                  default:\n                    doSomethingElse();\n                    \n                }\n              }\n            "},
		{name: "invalid10", source: "\n              function foo() {\n                switch (bar) {\n                  default:\n                    doSomething();\n                  case 1:\n                    doSomething();\n                    return;\n                }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                switch (bar) {\n                  default:\n                    doSomething();\n                  case 1:\n                    doSomething();\n                    \n                }\n              }\n            "},
		{name: "invalid11", source: "\n              function foo() {\n                switch (bar) {\n                  case 1:\n                    if (a) {\n                      doSomething();\n                      return;\n                    }\n                    break;\n                  default:\n                    doSomethingElse();\n                }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                switch (bar) {\n                  case 1:\n                    if (a) {\n                      doSomething();\n                      \n                    }\n                    break;\n                  default:\n                    doSomethingElse();\n                }\n              }\n            "},
		{name: "invalid12", source: "\n              function foo() {\n                switch (bar) {\n                  case 1:\n                    if (a) {\n                      doSomething();\n                      return;\n                    } else {\n                      doSomething();\n                    }\n                    break;\n                  default:\n                    doSomethingElse();\n                }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                switch (bar) {\n                  case 1:\n                    if (a) {\n                      doSomething();\n                      \n                    } else {\n                      doSomething();\n                    }\n                    break;\n                  default:\n                    doSomethingElse();\n                }\n              }\n            "},
		{name: "invalid13", source: "\n              function foo() {\n                switch (bar) {\n                  case 1:\n                    if (a) {\n                      doSomething();\n                      return;\n                    }\n                  default:\n                }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                switch (bar) {\n                  case 1:\n                    if (a) {\n                      doSomething();\n                      \n                    }\n                  default:\n                }\n              }\n            "},
		{name: "invalid14", source: "\n              function foo() {\n                try {} catch (err) { return; }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                try {} catch (err) {  }\n              }\n            "},
		{name: "invalid15", source: "\n              function foo() {\n                try {\n                  foo();\n                  return;\n                } catch (err) {\n                  return 5;\n                }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                try {\n                  foo();\n                  \n                } catch (err) {\n                  return 5;\n                }\n              }\n            "},
		{name: "invalid16", source: "\n              function foo() {\n                  if (something) {\n                      try {\n                          bar();\n                          return;\n                      } catch (err) {}\n                  }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                  if (something) {\n                      try {\n                          bar();\n                          \n                      } catch (err) {}\n                  }\n              }\n            "},
		{name: "invalid17", source: "\n              function foo() {\n                try {\n                  return;\n                } catch (err) {\n                  foo();\n                }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                try {\n                  \n                } catch (err) {\n                  foo();\n                }\n              }\n            "},
		{name: "invalid18", source: "\n              function foo() {\n                  try {\n                      return;\n                  } finally {\n                      bar();\n                  }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                  try {\n                      \n                  } finally {\n                      bar();\n                  }\n              }\n            "},
		{name: "invalid19", source: "\n              function foo() {\n                try {\n                  bar();\n                } catch (e) {\n                  try {\n                    baz();\n                    return;\n                  } catch (e) {\n                    qux();\n                  }\n                }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                try {\n                  bar();\n                } catch (e) {\n                  try {\n                    baz();\n                    \n                  } catch (e) {\n                    qux();\n                  }\n                }\n              }\n            "},
		{name: "invalid20", source: "\n              function foo() {\n                try {} finally {}\n                return;\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                try {} finally {}\n                \n              }\n            "},
		{name: "invalid21", source: "\n              function foo() {\n                try {\n                  return 5;\n                } finally {\n                  function bar() {\n                    return;\n                  }\n                }\n              }\n            ", findings: 1, wantFixed: "\n              function foo() {\n                try {\n                  return 5;\n                } finally {\n                  function bar() {\n                    \n                  }\n                }\n              }\n            "},
		{name: "invalid22", source: "() => { return; }", findings: 1, wantFixed: "() => {  }"},
		{name: "invalid23", source: "function foo() { return; return; }", findings: 1, wantFixed: "function foo() {  return; }"},
	}
	if len(cases) != 24 {
		t.Fatalf("REFUSING: the reporting table holds %d cases, want 24", len(cases))
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessReturn, "file.ts", testCase.source)
			wantIds := make([]string, testCase.findings)
			for index := range wantIds {
				wantIds[index] = "unnecessaryReturn"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			if testCase.wantFixed == noUselessReturnDeclinesToFix {
				for index, finding := range result.Diagnostics {
					if len(finding.Fixes) > 0 {
						t.Errorf("finding %d offered a repair upstream declines; for this rule that "+
							"means deleting a statement it refuses to delete", index)
					}
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// The span and the message, neither of which ExpectFindings can see. Upstream reports the whole
// return statement, which is also the span the repair deletes, so a wrong anchor here means an
// unattended edit lands somewhere the reader was never shown.
func TestNoUselessReturnSpanAndMessage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		wantText string
	}{
		{"onlyStatement", "function foo() { return; }", "return;"},
		{"afterAStatement", "function foo() { doSomething(); return; }", "return;"},
		{"insideAnIf", "function foo() { if (a) { bar(); return; } else { baz(); } }", "return;"},
		{"insideACase", "function foo() { switch (a) { case 1: bar(); return; } }", "return;"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessReturn, "file.ts", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected 1 finding, got %d", len(result.Diagnostics))
			}
			finding := result.Diagnostics[0]
			if got := testCase.source[finding.Range.Pos():finding.Range.End()]; got != testCase.wantText {
				t.Errorf("reported text: got %q, want %q", got, testCase.wantText)
			}
			if finding.Message.Id != "unnecessaryReturn" {
				t.Errorf("message id: got %q", finding.Message.Id)
			}
			wantPrefix := "This `return` ends the function at a point where the function was going " +
				"to end anyway, so it changes nothing"
			if got := finding.Message.Description; !strings.HasPrefix(got, wantPrefix) {
				t.Errorf("message:\n got %q\nwant prefix %q", got, wantPrefix)
			}
		})
	}
}

// The comment guard, written here rather than taken from the corpus.
//
// Upstream's corpus does exercise it, at `return/**/;` and `return//\n;`, so the omission is not
// invisible the way it would be for a guard nothing tested. What the corpus does NOT pin is the
// boundary: the guard is `getCommentsInside`, so a comment merely NEAR the return is still fixed.
// Measured against the installed build at 10.8.1:
//
//	function foo() { bar(); return/**/; }          reports, NO repair
//	function foo() { bar(); /*keep*/ return; }     reports, repair keeps the comment
//	function foo() { bar(); return; /*keep*/ }     reports, repair keeps the comment
//
// Both halves matter. Losing the first means the repair eats a developer's comment unattended;
// losing the second means the rule stops repairing ordinary code because a comment sits beside it.
func TestNoUselessReturnComments(t *testing.T) {
	t.Parallel()

	declines := []struct {
		name   string
		source string
	}{
		{"blockCommentInside", "function foo() { bar(); return/**/; }"},
		{"lineCommentInside", "function foo() { bar(); return//comment\n; }"},
		{"commentBetweenKeywordAndSemicolon", "function foo() { bar(); return /* why */ ; }"},
	}
	for _, testCase := range declines {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessReturn, "file.ts", testCase.source)
			rule_testing.ExpectFindings(t, result, "unnecessaryReturn")
			if len(result.Diagnostics[0].Fixes) != 0 {
				t.Errorf("offered a repair that would delete the comment inside the return")
			}
		})
	}

	// The other side of the boundary: a comment beside the return is not inside it, so the repair
	// stands and the comment survives it. Asserted through ExpectFixedSource so the surviving
	// comment is compared as text rather than inferred from a fix count.
	repairs := []struct {
		name      string
		source    string
		wantFixed string
	}{
		{"commentBefore", "function foo() { bar(); /*keep*/ return; }",
			"function foo() { bar(); /*keep*/  }"},
		{"commentAfter", "function foo() { bar(); return; /*keep*/ }",
			"function foo() { bar();  /*keep*/ }"},
	}
	for _, testCase := range repairs {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessReturn, "file.ts", testCase.source)
			rule_testing.ExpectFindings(t, result, "unnecessaryReturn")
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// The removability guard, upstream's third `output: null`. A return that is not a member of a
// statement list cannot be deleted, because removing it leaves a construct with no body.
func TestNoUselessReturnRemovability(t *testing.T) {
	t.Parallel()

	declines := []string{
		"function foo() { if (foo) return; }",
		"function foo() { if (foo) return; else bar(); }",
	}
	for _, source := range declines {
		result := rule_testing.Run(t, NoUselessReturn, "file.ts", source)
		if len(result.Diagnostics) == 0 {
			t.Fatalf("expected a finding on %q", source)
		}
		if len(result.Diagnostics[0].Fixes) != 0 {
			t.Errorf("offered a repair on %q, which would leave the `if` with no consequent", source)
		}
	}

	// A control: the same judgment inside a block IS repaired, so the decline above is about the
	// parent rather than about the rule declining to fix anything.
	rule_testing.ExpectFixedSource(t,
		rule_testing.Run(t, NoUselessReturn, "file.ts", "function foo() { if (foo) { return; } }"),
		"function foo() { if (foo) {  } }")
}

// TypeScript shapes upstream's corpus cannot contain, because its corpus is JavaScript.
//
// This rule deletes a whole statement rather than computing a span from a neighbouring node, so it
// does not have the failure that stranded type annotations elsewhere in this tree. These exist to
// pin that: a return inside a typed function, a generic one, and a method still reports and the
// repair leaves every annotation untouched.
func TestNoUselessReturnTypeScriptShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		source    string
		wantFixed string
	}{
		{"annotatedReturnType", "function foo(a: number): void { bar(a); return; }",
			"function foo(a: number): void { bar(a);  }"},
		{"generic", "function foo<T>(a: T): void { bar(a); return; }",
			"function foo<T>(a: T): void { bar(a);  }"},
		{"method", "class C { m(a: string): void { bar(a); return; } }",
			"class C { m(a: string): void { bar(a);  } }"},
		{"optionalParameter", "function foo(a?: number): void { bar(a); return; }",
			"function foo(a?: number): void { bar(a);  }"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Typed because an annotated function is exactly where the rule asks the checker
			// whether the compiler needs the return. See TestNoUselessReturnKeepsAReturnTheCompilerRequires.
			// The typed harness writes the fixture with a trailing newline, which the repair keeps.
			result := rule_testing.RunTyped(t, NoUselessReturn, "file.ts", testCase.source)
			rule_testing.ExpectFindings(t, result, "unnecessaryReturn")
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed+"\n")
		})
	}
}

// TestNoUselessReturnKeepsAReturnTheCompilerRequires pins the shapes where deleting the return stops
// the file compiling. The real site is api-phi-health's `GoogleAdsEnhancedConversionsService.ts`, a
// method declared `DictionaryType<unknown> | undefined` whose whole body was `return;`.
//
// Every expectation here was compiled with the return deleted rather than predicted: each silent row
// fails with TS2355 or TS2378, and each reporting row compiles.
func TestNoUselessReturnKeepsAReturnTheCompilerRequires(t *testing.T) {
	t.Parallel()

	const prelude = "type DictionaryType<T> = Record<string, T>;\ntype Nothing = void;\n" +
		"declare const condition: boolean;\ndeclare function bar(): void;\n"

	silent := []struct {
		name   string
		source string
	}{
		{"GoogleAdsEnhancedConversionsService.ts: an alias joined with undefined",
			"export class S {\n    generateRecurringPurchaseEvent(): DictionaryType<unknown> | undefined {\n" +
				"        return;\n    }\n}\n"},
		{"a union with undefined after other work", "export function f(): string | undefined { bar(); return; }\n"},
		{"unknown", "export function f(): unknown { bar(); return; }\n"},
		{"an async function's promised type", "export async function f(): Promise<string | undefined> { bar(); return; }\n"},
		{"a getter, whatever its type", "export class S { get g() { bar(); return; } }\n"},
		{"a nested function's return does not satisfy the outer one",
			"export function f(): string | undefined { const g = () => { return 'x'; }; g(); return; }\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUselessReturn, "file.ts", prelude+testCase.source))
		})
	}

	reporting := []struct {
		name   string
		source string
	}{
		{"exactly undefined", "export function f(): undefined { bar(); return; }\n"},
		{"void behind an alias", "export function f(): Nothing | string { bar(); return; }\n"},
		{"Promise of void", "export async function f(): Promise<void> { bar(); return; }\n"},
		{"any", "export function f(): any { bar(); return; }\n"},
		{"unannotated", "export function f() { bar(); return; }\n"},
		{"another return still satisfies the compiler",
			"export function f(): string | undefined { if (condition) { return 'x'; } bar(); return; }\n"},
	}
	for _, testCase := range reporting {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUselessReturn, "file.ts", prelude+testCase.source),
				"unnecessaryReturn")
		})
	}
}

// The try boundary, in all six directions. Found by a false positive on the real tree rather than
// by the corpus, which has no case for the first one.
//
// A return inside a try block is rescued by later statements in that SAME BLOCK, and not by the
// catch or the finally. Upstream's corpus writes the catch and finally halves and never writes an
// early exit followed by more work inside the try, which is the ordinary shape of guarded code:
// before this was corrected the rule reported 51 such sites across 33 files on this tree, and every
// one was a false positive. Each row measured against the installed build at 10.8.1.
func TestNoUselessReturnTryBoundary(t *testing.T) {
	t.Parallel()

	clean := []struct {
		name   string
		source string
	}{
		{"laterCodeInTheSameTryBlockRescues",
			"function f() { try { if (x) { return; } more(); } catch (e) {} }"},
		{"codeAfterTheWholeTryRescues",
			"function f() { try { bar(); return; } finally { baz(); } qux(); }"},
		{"codeAfterTryCatchRescues",
			"function f() { try { bar(); return; } catch (e) {} baz(); }"},
		// The shape this rule reported 51 times before the boundary was fixed: a guard clause
		// followed by the work it guards, all inside one try.
		{"guardClauseFollowedByItsWork",
			"function f() { try { const a = g(); if (a === b) { return; } write(a); } catch (e) { log(e); } }"},
	}
	for _, testCase := range clean {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUselessReturn, "file.ts", testCase.source))
		})
	}

	reports := []struct {
		name   string
		source string
	}{
		{"catchDoesNotRescue", "function f() { try { foo(); return; } catch (e) { qux(); } }"},
		{"finallyDoesNotRescue", "function f() { try { return; } finally { bar(); } }"},
		{"nothingAfterTheTry", "function f() { try { bar(); return; } catch (e) {} }"},
	}
	for _, testCase := range reports {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUselessReturn, "file.ts", testCase.source), "unnecessaryReturn")
		})
	}
}
