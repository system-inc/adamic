package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const preferIncludesFile = "/repository/source/Thing.ts"

// preferIncludesCaseName numbers a row so a failure names which one, since many rows differ only in
// the comparison operator or in one element type.
func preferIncludesCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// preferIncludesOnDisk is what the harness actually writes for a fixture.
//
// `RunTyped` writes each file as `strings.TrimSpace(contents)+"\n"`, so every case copied from an
// upstream tester carries a leading newline the file on disk does not have. A span sliced from the
// Go literal is therefore off by one, and an expected fix output compared against the untrimmed
// literal fails on a trailing newline while the repair is byte correct. Transform the expectation
// the same way the harness transforms the input rather than padding the rule to match.
func preferIncludesOnDisk(sourceText string) string {
	return strings.TrimSpace(sourceText) + "\n"
}

// TestPreferIncludesStaysSilentOnUpstreamPassCases is the imported clean corpus.
//
// All thirteen of upstream's passing inputs, extracted from the clone's test file by parsing it with
// the TypeScript compiler rather than by reading it, so no escape sequence passed through a shell or
// a keyboard on the way here. Every one was additionally run through the installed 8.67.0 build
// driven over a real TypeScript program, which reported nothing on all thirteen.
//
// Four of them turn on the parameter comparison alone and are the reason that comparison is on
// source text rather than on types: a user type declaring `indexOf(x, fromIndex?)` beside
// `includes(x)` is clean, and so is one whose `includes` is a boolean property rather than a method.
func TestPreferIncludesStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\nfunction f(a: string): void {\n  a.indexOf(b);\n}\n    ",
		"\nfunction f(a: string): void {\n  a.indexOf(b) + 0;\n}\n    ",
		"\nfunction f(a: string | { value: string }): void {\n  a.indexOf(b) !== -1;\n}\n    ",
		"\ntype UserDefined = {\n  indexOf(x: any): number; // don't have 'includes'\n};\nfunction f(a: UserDefined): void {\n  a.indexOf(b) !== -1;\n}\n    ",
		"\ntype UserDefined = {\n  indexOf(x: any, fromIndex?: number): number;\n  includes(x: any): boolean; // different parameters\n};\nfunction f(a: UserDefined): void {\n  a.indexOf(b) !== -1;\n}\n    ",
		"\ntype UserDefined = {\n  indexOf(x: any, fromIndex?: number): number;\n  includes(x: any, fromIndex: number): boolean; // different parameters\n};\nfunction f(a: UserDefined): void {\n  a.indexOf(b) !== -1;\n}\n    ",
		"\ntype UserDefined = {\n  indexOf(x: any, fromIndex?: number): number;\n  includes: boolean; // different type\n};\nfunction f(a: UserDefined): void {\n  a.indexOf(b) !== -1;\n}\n    ",
		"\nfunction f(a: string): void {\n  /bar/i.test(a);\n}\n    ",
		"\nfunction f(a: string): void {\n  /ba[rz]/.test(a);\n}\n    ",
		"\nfunction f(a: string): void {\n  /foo|bar/.test(a);\n}\n    ",
		"\nfunction f(a: string): void {\n  /bar/.test();\n}\n    ",
		"\nfunction f(a: string): void {\n  something.test(a);\n}\n    ",
		"\nconst pattern = new RegExp('bar');\nfunction f(a) {\n  return pattern.test(a);\n}\n    ",
	}
	for index, sourceText := range cases {
		t.Run(preferIncludesCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreferIncludes,
				preferIncludesFile, sourceText))
		})
	}
}

// TestPreferIncludesFiresOnUpstreamFailCases is the imported reporting corpus.
//
// Twenty-nine inputs carrying twenty-nine findings, twenty-seven of which are fixable. The two that
// are not are optional chains, which upstream reports while deliberately withholding the repair
// because a nullish receiver makes the rewrite change meaning: `undefined !== -1` is true while
// `undefined.includes` throws. Those rows carry `output: null` upstream and are asserted here as a
// finding with zero fixes.
//
// Three things are asserted per row. The ids say which of the two halves ran. The spans say where
// the finding points, which for the comparison half is the whole comparison rather than the call.
// And the applied source says what the edit engine will write unattended, which no message-id
// assertion can see.
func TestPreferIncludesFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
		wantSpans  []string
		wantFixed  string
	}{
		{
			sourceText: "\nfunction f(a: string): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: string): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: string): void {\n  a.indexOf(b) != -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) != -1"},
			wantFixed:  "\nfunction f(a: string): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: string): void {\n  a.indexOf(b) > -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) > -1"},
			wantFixed:  "\nfunction f(a: string): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: string): void {\n  a.indexOf(b) >= 0;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) >= 0"},
			wantFixed:  "\nfunction f(a: string): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: string): void {\n  a.indexOf(b) === -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) === -1"},
			wantFixed:  "\nfunction f(a: string): void {\n  !a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: string): void {\n  a.indexOf(b) == -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) == -1"},
			wantFixed:  "\nfunction f(a: string): void {\n  !a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: string): void {\n  a.indexOf(b) <= -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) <= -1"},
			wantFixed:  "\nfunction f(a: string): void {\n  !a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: string): void {\n  a.indexOf(b) < 0;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) < 0"},
			wantFixed:  "\nfunction f(a: string): void {\n  !a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a?: string): void {\n  a?.indexOf(b) === -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a?.indexOf(b) === -1"},
			wantFixed:  "",
		},
		{
			sourceText: "\nfunction f(a?: string): void {\n  a?.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a?.indexOf(b) !== -1"},
			wantFixed:  "",
		},
		{
			sourceText: "\nfunction f(a: string): void {\n  /bar/.test(a);\n}\n      ",
			wantIds:    []string{"preferStringIncludes"},
			wantSpans:  []string{"/bar/.test(a)"},
			wantFixed:  "\nfunction f(a: string): void {\n  a.includes('bar');\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: string): void {\n  /bar/.test((1 + 1, a));\n}\n      ",
			wantIds:    []string{"preferStringIncludes"},
			wantSpans:  []string{"/bar/.test((1 + 1, a))"},
			wantFixed:  "\nfunction f(a: string): void {\n  (1 + 1, a).includes('bar');\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: string): void {\n  /\\0'\\\\\\n\\r\\v\\t\\f/.test(a);\n}\n      ",
			wantIds:    []string{"preferStringIncludes"},
			wantSpans:  []string{"/\\0'\\\\\\n\\r\\v\\t\\f/.test(a)"},
			wantFixed:  "\nfunction f(a: string): void {\n  a.includes('\\0\\'\\\\\\n\\r\\v\\t\\f');\n}\n      ",
		},
		{
			sourceText: "\nconst pattern = new RegExp('bar');\nfunction f(a: string): void {\n  pattern.test(a);\n}\n      ",
			wantIds:    []string{"preferStringIncludes"},
			wantSpans:  []string{"pattern.test(a)"},
			wantFixed:  "\nconst pattern = new RegExp('bar');\nfunction f(a: string): void {\n  a.includes('bar');\n}\n      ",
		},
		{
			sourceText: "\nconst pattern = /bar/;\nfunction f(a: string, b: string): void {\n  pattern.test(a + b);\n}\n      ",
			wantIds:    []string{"preferStringIncludes"},
			wantSpans:  []string{"pattern.test(a + b)"},
			wantFixed:  "\nconst pattern = /bar/;\nfunction f(a: string, b: string): void {\n  (a + b).includes('bar');\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: any[]): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: any[]): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: ReadonlyArray<any>): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: ReadonlyArray<any>): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: Int8Array): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: Int8Array): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: Int16Array): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: Int16Array): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: Int32Array): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: Int32Array): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: Uint8Array): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: Uint8Array): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: Uint16Array): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: Uint16Array): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: Uint32Array): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: Uint32Array): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: Float32Array): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: Float32Array): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: Float64Array): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: Float64Array): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f<T>(a: T[] | ReadonlyArray<T>): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f<T>(a: T[] | ReadonlyArray<T>): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f<\n  T,\n  U extends\n    | T[]\n    | ReadonlyArray<T>\n    | Int8Array\n    | Uint8Array\n    | Int16Array\n    | Uint16Array\n    | Int32Array\n    | Uint32Array\n    | Float32Array\n    | Float64Array,\n>(a: U): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f<\n  T,\n  U extends\n    | T[]\n    | ReadonlyArray<T>\n    | Int8Array\n    | Uint8Array\n    | Int16Array\n    | Uint16Array\n    | Int32Array\n    | Uint32Array\n    | Float32Array\n    | Float64Array,\n>(a: U): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\ntype UserDefined = {\n  indexOf(x: any): number;\n  includes(x: any): boolean;\n};\nfunction f(a: UserDefined): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\ntype UserDefined = {\n  indexOf(x: any): number;\n  includes(x: any): boolean;\n};\nfunction f(a: UserDefined): void {\n  a.includes(b);\n}\n      ",
		},
		{
			sourceText: "\nfunction f(a: Readonly<any[]>): void {\n  a.indexOf(b) !== -1;\n}\n      ",
			wantIds:    []string{"preferIncludes"},
			wantSpans:  []string{"a.indexOf(b) !== -1"},
			wantFixed:  "\nfunction f(a: Readonly<any[]>): void {\n  a.includes(b);\n}\n      ",
		},
	}
	for index, testCase := range cases {
		t.Run(preferIncludesCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferIncludes, preferIncludesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			onDisk := preferIncludesOnDisk(testCase.sourceText)
			for index, wantSpan := range testCase.wantSpans {
				reported := onDisk[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantSpan {
					t.Errorf("finding %d covers %q, want %q", index, reported, wantSpan)
				}
			}

			if testCase.wantFixed == "" {
				// Upstream's `output: null`: reported, deliberately not repaired.
				for index, diagnostic := range result.Diagnostics {
					if len(diagnostic.Fixes) != 0 {
						t.Errorf("finding %d carries %d fixes, want none because upstream declines to repair it",
							index, len(diagnostic.Fixes))
					}
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result, preferIncludesOnDisk(testCase.wantFixed))
		})
	}
}

// TestPreferIncludesReadsAPatternTheShelfCannot is the substrate fixture.
//
// The shelf's `regexpattern.Walk` is the obvious thing to build the regex half on and it cannot
// answer this rule's question. Measured by running it over these exact patterns: it renders `b.r`
// as `b.r`, renders `^bar`, `bar$`, `(bar)` and `b(?:a)r` all as `bar`, and renders `foo|bar` as
// `foobar`. A port built on it would rewrite `/^bar$/.test(a)` into `a.includes('bar')`, which is
// wrong, and no imported fixture would catch it because upstream's corpus writes only three of
// these shapes.
//
// So the reader is written on `regexsyntax` primitives instead, which is what the shelf's own doc
// says a caller wanting alternation should do. These rows are what proves it discriminates. Every
// verdict is the installed 8.67.0 build's, measured over a real program.
func TestPreferIncludesReadsAPatternTheShelfCannot(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantCount  int
		wantFixed  string
		reason     string
	}{
		{
			sourceText: "function f(a: string): void {\n  /bar/.test(a);\n}",
			wantCount:  1,
			wantFixed:  "function f(a: string): void {\n  a.includes('bar');\n}",
			reason:     "a fixed string reports",
		},
		{
			sourceText: "function f(a: string): void {\n  /b\\.r/.test(a);\n}",
			wantCount:  1,
			wantFixed:  "function f(a: string): void {\n  a.includes('b.r');\n}",
			reason:     "an escaped dot is a character, so this reports",
		},
		{
			sourceText: "function f(a: string): void {\n  /b.r/.test(a);\n}",
			wantCount:  0,
			wantFixed:  "",
			reason:     "an unescaped dot is the any-character metacharacter",
		},
		{
			sourceText: "function f(a: string): void {\n  /^bar/.test(a);\n}",
			wantCount:  0,
			wantFixed:  "",
			reason:     "an anchor is not a character, and Walk erases it",
		},
		{
			sourceText: "function f(a: string): void {\n  /bar$/.test(a);\n}",
			wantCount:  0,
			wantFixed:  "",
			reason:     "and so is the end anchor",
		},
		{
			sourceText: "function f(a: string): void {\n  /(bar)/.test(a);\n}",
			wantCount:  0,
			wantFixed:  "",
			reason:     "a group is not a character, and Walk erases it",
		},
		{
			sourceText: "function f(a: string): void {\n  /b(?:a)r/.test(a);\n}",
			wantCount:  0,
			wantFixed:  "",
			reason:     "nor is a non-capturing group",
		},
		{
			sourceText: "function f(a: string): void {\n  /foo|bar/.test(a);\n}",
			wantCount:  0,
			wantFixed:  "",
			reason:     "alternation, which Walk reports as one run of characters",
		},
		{
			sourceText: "function f(a: string): void {\n  /ba[rz]/.test(a);\n}",
			wantCount:  0,
			wantFixed:  "",
			reason:     "a character class",
		},
		{
			sourceText: "function f(a: string): void {\n  /ba+r/.test(a);\n}",
			wantCount:  0,
			wantFixed:  "",
			reason:     "a quantifier",
		},
		{
			sourceText: "function f(a: string): void {\n  /a{2}/.test(a);\n}",
			wantCount:  0,
			wantFixed:  "",
			reason:     "a counted quantifier",
		},
		{
			sourceText: "function f(a: string): void {\n  /\\d/.test(a);\n}",
			wantCount:  0,
			wantFixed:  "",
			reason:     "a set escape has no single character to answer",
		},
		{
			sourceText: "function f(a: string): void {\n  /\\x41/.test(a);\n}",
			wantCount:  1,
			wantFixed:  "function f(a: string): void {\n  a.includes('A');\n}",
			reason:     "a hex escape resolves to its character",
		},
		{
			sourceText: "function f(a: string): void {\n  /\\u0041/.test(a);\n}",
			wantCount:  1,
			wantFixed:  "function f(a: string): void {\n  a.includes('A');\n}",
			reason:     "and so does a unicode escape",
		},
		{
			sourceText: "function f(a: string): void {\n  /a\\-b/.test(a);\n}",
			wantCount:  1,
			wantFixed:  "function f(a: string): void {\n  a.includes('a-b');\n}",
			reason:     "an identity escape is the character itself",
		},
	}
	for index, testCase := range cases {
		t.Run(preferIncludesCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferIncludes, preferIncludesFile, testCase.sourceText)
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "preferStringIncludes"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
			if testCase.wantCount > 0 && testCase.wantFixed != "" {
				rule_testing.ExpectFixedSource(t, result, preferIncludesOnDisk(testCase.wantFixed))
			}
		})
	}
}

// TestPreferIncludesDiscriminatesOnCasesUpstreamDoesNotWrite covers the rest.
//
// Upstream's corpus writes no shadowed sentinel, no computed `test` access, and only one shape of
// receiver for the regex half. Each row was run through the installed build and carries the verdict
// that build produced.
func TestPreferIncludesDiscriminatesOnCasesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
		reason     string
	}{
		{
			sourceText: "function f(a: string): void {\n  a.indexOf(b) !== 0;\n}",
			wantIds:    nil,
			reason:     "a sentinel of zero under !== is not a presence test",
		},
		{
			sourceText: "function f(a: string): void {\n  a.indexOf(b) > 0;\n}",
			wantIds:    nil,
			reason:     "nor is greater-than-zero",
		},
		{
			sourceText: "function f(a: string): void {\n  -1 !== a.indexOf(b);\n}",
			wantIds:    nil,
			reason:     "the sentinel has to be on the right",
		},
		{
			sourceText: "function f(a: string): void {\n  a.indexOf(b) !== -1.0;\n}",
			wantIds:    []string{"preferIncludes"},
			reason:     "a fractional spelling still canonicalizes to 1",
		},
		{
			sourceText: "function f(a: string): void {\n  a.indexOf(b) !== -0x1;\n}",
			wantIds:    []string{"preferIncludes"},
			reason:     "and so does a hexadecimal one",
		},
		{
			sourceText: "declare const re: RegExp;\nfunction f(a: string): void {\n  re.test(a);\n}",
			wantIds:    nil,
			reason:     "an unresolvable receiver declines",
		},
		{
			sourceText: "function f(a: string): void {\n  /bar/[\"test\"](a);\n}",
			wantIds:    nil,
			reason:     "a computed test access is a different node kind",
		},
		{
			sourceText: "function f(a: string): void {\n  /bar/.test(a, b);\n}",
			wantIds:    nil,
			reason:     "the selector requires exactly one argument",
		},
		{
			sourceText: "function f(a: number): void {\n  /bar/.test(a as any);\n}",
			wantIds:    nil,
			reason:     "an argument whose type has no includes declines",
		},
		{
			// The function-like test is the ONLY thing declining this, and it took a survivor to
			// find an input that can see it. Upstream's own clean case pairs `includes: boolean`
			// with a two-parameter `indexOf`, so the parameter COUNT guard declines it first and a
			// mutant treating a property as a zero-parameter function survives the whole corpus.
			// Giving `indexOf` no parameters makes both counts zero, so only the function-like test
			// is left. Measured clean upstream against the control below, which reports.
			sourceText: "type U = { indexOf(): number; includes: boolean };\ndeclare const u: U;\nfunction f(): void {\n  u.indexOf(b) !== -1;\n}",
			wantIds:    nil,
			reason:     "a property is not function-like even when the parameter counts agree",
		},
		{
			sourceText: "type U = { indexOf(): number; includes(): boolean };\ndeclare const u: U;\nfunction f(): void {\n  u.indexOf(b) !== -1;\n}",
			wantIds:    []string{"preferIncludes"},
			reason:     "the control: two zero-parameter methods do pair",
		},
		{
			sourceText: "function f(a: string[]): void {\n  a.indexOf(b) !== -1;\n}",
			wantIds:    []string{"preferIncludes"},
			reason:     "an array reports through the same path as a string",
		},
	}
	for index, testCase := range cases {
		t.Run(preferIncludesCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferIncludes, preferIncludesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestPreferIncludesRequiresTheTypedHarness pins the checker declaration.
//
// A typed rule handed the plain harness gets a nil checker, and the guard at the top of each
// listener turns that into silence rather than a panic. Silence is the more dangerous failure: every
// clean case passes vacuously and every reporting case fails in a way that reads as a rule bug. This
// asserts the guard holds, so a later revert fails loudly here rather than going quiet everywhere.
func TestPreferIncludesRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "function f(a: string): void {\n  a.indexOf(b) !== -1;\n}"
	rule_testing.ExpectClean(t, rule_testing.Run(t, PreferIncludes, preferIncludesFile, source))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreferIncludes, preferIncludesFile, source),
		"preferIncludes")
}
