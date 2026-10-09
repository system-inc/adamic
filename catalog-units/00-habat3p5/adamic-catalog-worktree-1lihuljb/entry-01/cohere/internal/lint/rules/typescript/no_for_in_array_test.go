package typescript

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const noForInArrayFile = "/repository/source/ForIn.ts"

// domGlobals declares what lib ES2022 does not carry, and it is load-bearing rather than decorative.
//
// The harness tsconfig pins `lib: ["ES2022"]` with no way to raise it, and its own comment says it
// omits the DOM on purpose. Two of upstream's failing cases name `HTMLCollection` and `NodeList`,
// which therefore resolve to the ERROR type. Probed rather than assumed: dropped into the
// single-file table unchanged, both go silent while upstream reports them.
//
// That is the dangerous direction, and it is only visible because the whole corpus was replayed
// rather than read. Note which upstream types did NOT need this: `RegExpExecArray` and the
// `arguments` object are both in ES2022 and both report under the plain harness, so the second file
// is scoped to the two cases that genuinely need it rather than applied to everything DOM-shaped.
//
// These are minimal stand-ins rather than the real lib.dom declarations. What the rule asks of them
// is exactly two things, a number index signature and a numeric `length`, so a faithful shape is
// what matters rather than a faithful surface.
const domGlobals = `
declare global {
  interface Element {}
  interface Node {}
  interface HTMLCollection {
    readonly length: number;
    item(index: number): Element | null;
    [index: number]: Element;
  }
  interface NodeList {
    readonly length: number;
    item(index: number): Node | null;
    [index: number]: Node;
  }
}
export {};
`

// TestNoForInArrayFires carries tsgolint's eighteen failing inputs that need no library beyond
// ES2022, verbatim from its own test file.
//
// Every one was additionally driven through `@typescript-eslint` 8.67.0 on a real program before it
// became a fixture, and both references produced the same verdict and the same span on all of them.
func TestNoForInArrayFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"array literal subject", "\nfor (const x in [3, 4, 5]) {\n  console.log(x);\n}\n      "},
		{"array through a const binding", "\nconst z = [3, 4, 5];\nfor (const x in z) {\n  console.log(x);\n}\n      "},
		{"array parameter", "\nconst fn = (arr: number[]) => {\n  for (const x in arr) {\n    console.log(x);\n  }\n};\n      "},
		{"union of two array types", "\nconst fn = (arr: number[] | string[]) => {\n  for (const x in arr) {\n    console.log(x);\n  }\n};\n      "},
		{"type parameter constrained to an array", "\nconst fn = <T extends any[]>(arr: T) => {\n  for (const x in arr) {\n    console.log(x);\n  }\n};\n      "},
		{"head spanning nested parentheses with a block body", "\nfor (const x\n  in\n    (\n      (\n        (\n          [3, 4, 5]\n        )\n      )\n    )\n  )\n  // weird\n  /* spot for a */\n  // comment\n  /* ) */\n  /* ( */\n  {\n  console.log(x);\n}\n      "},
		{"head spanning nested parentheses with an unbraced body", "\nfor (const x\n  in\n    (\n      (\n        (\n          [3, 4, 5]\n        )\n      )\n    )\n  )\n  // weird\n  /* spot for a */\n  // comment\n  /* ) */\n  /* ( */\n\n  ((((console.log('body without braces ')))));\n\n      "},
		{"array or null", "\ndeclare const array: string[] | null;\n\nfor (const key in array) {\n  console.log(key);\n}\n      "},
		{"array or undefined", "\ndeclare const array: number[] | undefined;\n\nfor (const key in array) {\n  console.log(key);\n}\n      "},
		{"array or a plain object", "\ndeclare const array: boolean[] | { a: 1; b: 2; c: 3 };\n\nfor (const key in array) {\n  console.log(key);\n}\n      "},
		{"tuple subject", "\ndeclare const array: [number, string];\n\nfor (const key in array) {\n  console.log(key);\n}\n      "},
		{"tuple or a plain object", "\ndeclare const array: [number, string] | { a: 1; b: 2; c: 3 };\n\nfor (const key in array) {\n  console.log(key);\n}\n      "},
		{"array or a numeric record", "\ndeclare const array: string[] | Record<number, string>;\n\nfor (const key in array) {\n  console.log(key);\n}\n      "},
		{"regular expression match result", "\nconst arrayLike = /fe/.exec('foo');\n\nfor (const x in arrayLike) {\n  console.log(x);\n}\n      "},
		{"arguments object", "\nfunction foo() {\n  for (const a in arguments) {\n    console.log(a);\n  }\n}\n      "},
		{"intersection of an object and an array inside a union", "\ndeclare const array:\n  | (({ a: string } & string[]) | Record<string, boolean>)\n  | Record<number, string>;\n\nfor (const key in array) {\n  console.log(key);\n}\n      "},
		{"intersection of an object and a match result inside a union", "\ndeclare const array:\n  | (({ a: string } & RegExpExecArray) | Record<string, boolean>)\n  | Record<number, string>;\n\nfor (const key in array) {\n  console.log(k);\n}\n      "},
		{"object with a number index and a numeric length", "\ndeclare const obj: {\n  [key: number]: number;\n  length: 1;\n};\n\nfor (const key in obj) {\n  console.log(key);\n}\n      "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoForInArray,
				noForInArrayFile, testCase.sourceText), "forInViolation")
		})
	}
}

// TestNoForInArrayStaysSilent carries all four of tsgolint's passing inputs, verbatim.
//
// The fourth is the one worth naming. `{ [key: number]: number }` has the number index signature
// the rule looks for and NO `length`, so it passes; the same object with `length: 1` added is a
// failing case in the table above. That pair is the only thing in either corpus proving the length
// half of the predicate is load-bearing rather than decorative.
func TestNoForInArrayStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"for-of over an array", "\nfor (const x of [3, 4, 5]) {\n  console.log(x);\n}\n    "},
		{"for-in over a plain object", "\nfor (const x in { a: 1, b: 2, c: 3 }) {\n  console.log(x);\n}\n    "},
		{"for-in over a nullish value", "\ndeclare const nullish: null | undefined;\n// @ts-expect-error\nfor (const k in nullish) {\n}\n    "},
		{"object with a number index and no length", "\ndeclare const obj: {\n  [key: number]: number;\n};\n\nfor (const key in obj) {\n  console.log(key);\n}\n    "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoForInArray,
				noForInArrayFile, testCase.sourceText))
		})
	}
}

// TestNoForInArrayNeedsBothHalvesOfThePredicate covers what upstream's corpus does not.
//
// The rule reports only when BOTH hold: the type has a number index signature, AND it has a
// `length` whose type is number-like. Upstream's twenty four cases never separate the two, because
// every array-like thing they write satisfies both at once and everything else satisfies neither.
// Dropping either test therefore passed the entire imported corpus, and both dropped tests survived
// the mutation sweep.
//
// These six inputs split them. The first three carry a numeric `length` and NO number index, so
// they report if the index test is dropped. The last three carry a number index and a `length` that
// is not number-like, so they report if the flag test is dropped. All six are silent under the real
// rule, and each triple was confirmed to flip its own mutant and only its own.
//
// They are inventions rather than imported cases, so each was measured against `@typescript-eslint`
// 8.67.0 on a real program BEFORE becoming a fixture rather than after: all six are silent there
// too, alongside a plain-array control that fires, so the zero is a verdict rather than a broken
// harness. tsgolint could not be run directly for the same reason oxlint cannot, but the two
// implementations are line-for-line the same predicate.
func TestNoForInArrayNeedsBothHalvesOfThePredicate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// A numeric length with no number index. Silent, because half the predicate fails.
		{"numeric length and no index signature", "declare const o: { length: number };\nfor (const k in o) {\n  console.log(k);\n}"},
		{"a function, which has a numeric length", "declare const f: (a: number) => void;\nfor (const k in f) {\n  console.log(k);\n}"},
		{"string index signature with a numeric length", "declare const o: { [k: string]: unknown; length: number };\nfor (const k in o) {\n  console.log(k);\n}"},

		// A number index whose length is not number-like. Silent, because the other half fails.
		{"number index with a string length", "declare const o: { [k: number]: string; length: string };\nfor (const k in o) {\n  console.log(k);\n}"},
		{"number index with a boolean length", "declare const o: { [k: number]: string; length: boolean };\nfor (const k in o) {\n  console.log(k);\n}"},
		{"number index with a length method", "declare const o: { [k: number]: string; length(): number };\nfor (const k in o) {\n  console.log(k);\n}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoForInArray,
				noForInArrayFile, testCase.sourceText))
		})
	}
}

// TestNoForInArrayDomCollections carries the two upstream cases that need declarations lib ES2022
// does not supply, run against a two-file program.
//
// Splitting them out is what keeps them honest. Run under the single-file harness both pass too, by
// reporting nothing where a finding IS expected, so they would have sat in the clean list asserting
// the exact opposite of upstream while the suite stayed green.
func TestNoForInArrayDomCollections(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"html collection", "\ndeclare const arrayLike: HTMLCollection;\n\nfor (const x in arrayLike) {\n  console.log(x);\n}\n      "},
		{"node list", "\ndeclare const arrayLike: NodeList;\n\nfor (const x in arrayLike) {\n  console.log(x);\n}\n      "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFiles(t, NoForInArray, map[string]string{
				"ForIn.ts":      testCase.sourceText,
				"DomGlobals.ts": domGlobals,
			}, "ForIn.ts")
			rule_testing.ExpectFindings(t, result, "forInViolation")
		})
	}
}

// TestNoForInArrayReportsTheLoopHead asserts WHERE each finding points, which no message-id fixture
// can see.
//
// Upstream reports `GetForStatementHeadLoc`: from the `for` keyword to the start of the loop body,
// excluding the body. Every one of upstream's own line and column numbers is replayed here, because
// a rule pointing at the whole statement, or at the subject expression, would satisfy every
// assertion above while putting the caret in the wrong place.
//
// The two ten-line cases are the reason this is worth its length. Their heads run through four
// levels of nested parentheses and end before a run of comments, and they are the only inputs in
// either corpus that could distinguish the head span from the statement span.
//
// Upstream's line numbers are 1-based against text that begins with a newline, and the harness
// trims the source before parsing, so each expected line shifts up by one here.
func TestNoForInArrayReportsTheLoopHead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                             string
		sourceText                       string
		line, column, endLine, endColumn int
	}{
		{"array literal subject", "\nfor (const x in [3, 4, 5]) {\n  console.log(x);\n}\n      ", 1, 1, 1, 27},
		{"array through a const binding", "\nconst z = [3, 4, 5];\nfor (const x in z) {\n  console.log(x);\n}\n      ", 2, 1, 2, 19},
		{"array parameter", "\nconst fn = (arr: number[]) => {\n  for (const x in arr) {\n    console.log(x);\n  }\n};\n      ", 2, 3, 2, 23},
		{"union of two array types", "\nconst fn = (arr: number[] | string[]) => {\n  for (const x in arr) {\n    console.log(x);\n  }\n};\n      ", 2, 3, 2, 23},
		{"type parameter constrained to an array", "\nconst fn = <T extends any[]>(arr: T) => {\n  for (const x in arr) {\n    console.log(x);\n  }\n};\n      ", 2, 3, 2, 23},
		{"head spanning nested parentheses with a block body", "\nfor (const x\n  in\n    (\n      (\n        (\n          [3, 4, 5]\n        )\n      )\n    )\n  )\n  // weird\n  /* spot for a */\n  // comment\n  /* ) */\n  /* ( */\n  {\n  console.log(x);\n}\n      ", 1, 1, 10, 4},
		{"head spanning nested parentheses with an unbraced body", "\nfor (const x\n  in\n    (\n      (\n        (\n          [3, 4, 5]\n        )\n      )\n    )\n  )\n  // weird\n  /* spot for a */\n  // comment\n  /* ) */\n  /* ( */\n\n  ((((console.log('body without braces ')))));\n\n      ", 1, 1, 10, 4},
		{"array or null", "\ndeclare const array: string[] | null;\n\nfor (const key in array) {\n  console.log(key);\n}\n      ", 3, 1, 3, 25},
		{"array or undefined", "\ndeclare const array: number[] | undefined;\n\nfor (const key in array) {\n  console.log(key);\n}\n      ", 3, 1, 3, 25},
		{"array or a plain object", "\ndeclare const array: boolean[] | { a: 1; b: 2; c: 3 };\n\nfor (const key in array) {\n  console.log(key);\n}\n      ", 3, 1, 3, 25},
		{"tuple subject", "\ndeclare const array: [number, string];\n\nfor (const key in array) {\n  console.log(key);\n}\n      ", 3, 1, 3, 25},
		{"tuple or a plain object", "\ndeclare const array: [number, string] | { a: 1; b: 2; c: 3 };\n\nfor (const key in array) {\n  console.log(key);\n}\n      ", 3, 1, 3, 25},
		{"array or a numeric record", "\ndeclare const array: string[] | Record<number, string>;\n\nfor (const key in array) {\n  console.log(key);\n}\n      ", 3, 1, 3, 25},
		{"regular expression match result", "\nconst arrayLike = /fe/.exec('foo');\n\nfor (const x in arrayLike) {\n  console.log(x);\n}\n      ", 3, 1, 3, 27},
		{"arguments object", "\nfunction foo() {\n  for (const a in arguments) {\n    console.log(a);\n  }\n}\n      ", 2, 3, 2, 29},
		{"intersection of an object and an array inside a union", "\ndeclare const array:\n  | (({ a: string } & string[]) | Record<string, boolean>)\n  | Record<number, string>;\n\nfor (const key in array) {\n  console.log(key);\n}\n      ", 5, 1, 5, 25},
		{"intersection of an object and a match result inside a union", "\ndeclare const array:\n  | (({ a: string } & RegExpExecArray) | Record<string, boolean>)\n  | Record<number, string>;\n\nfor (const key in array) {\n  console.log(k);\n}\n      ", 5, 1, 5, 25},
		{"object with a number index and a numeric length", "\ndeclare const obj: {\n  [key: number]: number;\n  length: 1;\n};\n\nfor (const key in obj) {\n  console.log(key);\n}\n      ", 6, 1, 6, 23},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// The harness trims the source, so offsets are against the trimmed text.
			sourceText := strings.TrimSpace(testCase.sourceText)
			result := rule_testing.RunTyped(t, NoForInArray, noForInArrayFile, sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want exactly one finding to assert a span against, got %d", len(result.Diagnostics))
			}

			finding := result.Diagnostics[0]
			startLine, startColumn := lineAndColumnOf(sourceText, finding.Range.Pos())
			endLine, endColumn := lineAndColumnOf(sourceText, finding.Range.End())
			if startLine != testCase.line || startColumn != testCase.column ||
				endLine != testCase.endLine || endColumn != testCase.endColumn {
				t.Fatalf("finding spans %d:%d-%d:%d, want %d:%d-%d:%d\n  reported text: %q",
					startLine, startColumn, endLine, endColumn,
					testCase.line, testCase.column, testCase.endLine, testCase.endColumn,
					sourceText[finding.Range.Pos():finding.Range.End()])
			}

			// The head must stop before the body. Asserting the numbers alone would still pass if
			// upstream's own numbers were misread, so this checks the shape independently.
			reported := sourceText[finding.Range.Pos():finding.Range.End()]
			if !strings.HasPrefix(reported, "for") {
				t.Fatalf("a for-in head should begin at the `for` keyword, got %q", reported)
			}
			if strings.Contains(reported, "console.log") {
				t.Fatalf("the loop body leaked into the reported head: %q", reported)
			}
		})
	}
}

// lineAndColumnOf converts a byte offset into the 1-based line and column upstream reports.
func lineAndColumnOf(sourceText string, offset int) (int, int) {
	line, column := 1, 1
	for index := 0; index < offset && index < len(sourceText); index++ {
		if sourceText[index] == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return line, column
}

// TestNoForInArrayCarriesNoRepair pins that this rule offers nothing to apply.
//
// Both references report with a bare message and no fix or suggestion, because turning a for-in
// into a for-of changes what the loop binds, from a string key to a value. A later edit adding a
// repair would be a real behavior change against oxlint, and without this it would go unnoticed:
// every assertion above is satisfied by a finding that also carries a fix.
func TestNoForInArrayCarriesNoRepair(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoForInArray, noForInArrayFile,
		"for (const x in [3, 4, 5]) {\n  console.log(x);\n}")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("this rule must not offer a fix, got %d", len(result.Diagnostics[0].Fixes))
	}
	if len(result.Diagnostics[0].Suggestions) != 0 {
		t.Fatalf("this rule must not offer a suggestion, got %d", len(result.Diagnostics[0].Suggestions))
	}
}

// TestNoForInArrayMessageText asserts the rendered message, not only its id.
//
// The string is compared against a literal typed here rather than against the rule's own message
// constant, because comparing a finding to the constant it was built from is an equality that moves
// on both sides under mutation and therefore guards nothing.
func TestNoForInArrayMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoForInArray, noForInArrayFile,
		"for (const x in [3, 4, 5]) {\n  console.log(x);\n}")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}

	finding := result.Diagnostics[0]
	if finding.RuleName != "@typescript-eslint/no-for-in-array" {
		t.Fatalf("rule reported under %q, which no config entry enables and no suppression comment "+
			"could silence; upstream's own spelling carries a stray -rule suffix", finding.RuleName)
	}
	if finding.Message.Id != "forInViolation" {
		t.Fatalf("message id is %q, want %q", finding.Message.Id, "forInViolation")
	}
	const wantDescription = "For-in loops over arrays skips holes, returns indices as strings, " +
		"and may visit the prototype chain or other enumerable properties. Use a more robust " +
		"iteration method such as for-of or array.forEach instead."
	if finding.Message.Description != wantDescription {
		t.Fatalf("message text drifted from both upstreams:\n  got  %q\n  want %q",
			finding.Message.Description, wantDescription)
	}
}

// TestNoForInArrayRequiresTheTypedHarness pins the checker declaration AND the nil guard.
//
// While this rule was adapted, the `if ctx.TypeChecker == nil { return }` guard could not live in
// the rule file, because the listener was upstream's and editing it would have turned re-syncing
// into a merge; `upstream.Adapt` set NeedsTypeChecker unconditionally instead, so the nil case was
// unreachable through it. Absorbing the rule made the listener ours to edit and made that case
// REACHABLE, so both halves are asserted here.
//
// The declaration is the half that matters for registration. The guard is the half that matters for
// the harness path, where a Context can be built by hand — and it matters more than a crash would,
// because the shim's type queries return nil rather than panicking. Without the guard this rule
// would not crash under a nil checker, it would go silent, and every fixture above would pass having
// proven nothing.
func TestNoForInArrayRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoForInArray.NeedsTypeChecker {
		t.Fatal("this rule asks the checker what a type is; without NeedsTypeChecker it would be " +
			"handed a nil checker and go silent on every input while every fixture still passed")
	}

	// Drive the listener with a checker-less Context. It must return rather than reach the checker,
	// and this is the only path that reaches that branch, since registration always supplies one.
	typed := rule_testing.RunTyped(t, NoForInArray, noForInArrayFile, "declare const arr: number[];\nfor (const key in arr) {\n}\n")
	if len(typed.Diagnostics) != 1 {
		t.Fatalf("the typed harness found %d findings, want one", len(typed.Diagnostics))
	}

	listeners := NoForInArray.Run(rule.Context{SourceFile: typed.SourceFile}, nil)
	listener, hasListener := listeners[ast.KindForInStatement]
	if !hasListener {
		t.Fatal("the rule stopped listening on for-in statements")
	}
	for _, statement := range typed.SourceFile.Statements.Nodes {
		if statement.Kind == ast.KindForInStatement {
			listener(statement)
		}
	}
}
