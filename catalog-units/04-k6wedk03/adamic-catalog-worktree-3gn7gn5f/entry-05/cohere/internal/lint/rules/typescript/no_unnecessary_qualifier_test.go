package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noUnnecessaryQualifierFile = "/repository/source/Qualifiers.ts"

func noUnnecessaryQualifierCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoUnnecessaryQualifierStaysSilent is upstream's eight passing cases verbatim.
//
// Three of them are the ones that would be lost by a plausible simplification, and each was probed
// against our checker rather than assumed. The nested-shadow case passes because the scope lookup
// finds a DIFFERENT `T` and the export-symbol comparison answers false, not because nothing is in
// scope. `namespace X { const z = X.y }` passes because `y` does not exist, so the accessed symbol
// is nil. The two `Foo` cases where an enum and a namespace merge pass because the scope lookup for
// the accessed symbol's flags returns an empty scope.
func TestNoUnnecessaryQualifierStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"namespace X {\n  export type T = number;\n}\n\nnamespace Y {\n  export const x: X.T = 3;\n}\n",
		"namespace A {}\nnamespace A.B {\n  export type Z = 1;\n}\n",
		"enum A {\n  X,\n  Y,\n}\n\nenum B {\n  Z = A.X,\n}\n",
		"namespace X {\n  export type T = number;\n  namespace Y {\n    type T = string;\n    const x: X.T = 0;\n  }\n}\n",
		"namespace X {\n  const z = X.y;\n}\n",
		"enum Foo {\n  One,\n}\n\nnamespace Foo {\n  export function bar() {\n    return Foo.One;\n  }\n}\n",
		"namespace Foo {\n  export enum Foo {\n    One,\n  }\n}\n\nnamespace Foo {\n  export function bar() {\n    return Foo.One;\n  }\n}\n",
		"const x: A.B = 3;\n",

		// Measured boundary cases upstream does not write.

		// A computed access is a different node kind here and upstream excludes it with
		// `[computed=false]`. Without the kind split this would report and the fix would produce
		// `const y = ['x'];`, which is not the same program.
		"namespace A {\n  export const x = 3;\n  export const y = A['x'];\n}\n",

		// The qualifier is a namespace we are not inside, reached from a sibling namespace that
		// also declares the name. The scope lookup finds the local one, so the export comparison
		// separates them.
		"namespace A {\n  export type T = number;\n}\nnamespace B {\n  export type T = string;\n  const x: A.T = 3;\n}\n",

		// `globalThis` reaching a global declared by an augmentation. This is the case that pins
		// the namespaces-in-scope check, and it was found by a surviving mutant rather than by
		// reading: with that check removed the rule reports here and the fix would rewrite
		// `globalThis.g` to `g`, which is a different program under any shadowing. The other five
		// shapes tried first (a namespace merged with a class, with a function, an ordinary
		// outside access, an enum, a nested chain) do not separate the two versions, because the
		// export-symbol comparison already declines them. Measured on the installed 8.67.0 build:
		// silent there too.
		"declare global {\n  var g: number;\n}\nconst y = globalThis.g;\n",

		// A namespace merged with a class and one merged with a function, both accessed from
		// outside. Measured silent upstream and here. They are in this list because they were the
		// obvious guesses for the mutant above and are NOT the answer, which is worth pinning so
		// the next reader does not re-derive them.
		"class C {}\nnamespace C {\n  export const x = 3;\n}\nconst y = C.x;\n",
		"function F() {}\nnamespace F {\n  export const x = 3;\n}\nconst y = F.x;\n",
	}
	for index, sourceText := range cases {
		t.Run(noUnnecessaryQualifierCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnnecessaryQualifier,
				noUnnecessaryQualifierFile, sourceText))
		})
	}
}

// noUnnecessaryQualifierFinding is one expected finding with the two layers it can be wrong at.
//
// The rule has a single message id, so asserting the id proves almost nothing: what varies between
// a correct port and a broken one is WHERE the finding points and which name it says is in scope.
// On a nested chain the span is the whole qualifier prefix rather than its first segment, and a
// port reporting `A` where upstream reports `A.B.C` satisfies every id assertion.
type noUnnecessaryQualifierFinding struct {
	wantSpan string
	wantName string

	// wantNoFix marks the rows where this port reports WITHOUT the repair upstream ships, because
	// upstream's repair writes source that does not parse. Zero value is false, so every imported
	// row keeps its fix and only the deliberate declines say so.
	wantNoFix bool
}

// TestNoUnnecessaryQualifierFires is upstream's reporting cases verbatim, with their repairs.
//
// Eight of upstream's nine are here. The ninth is
//
//	import * as Foo from './foo';
//	declare module './foo' { const x: Foo.T = 3; }
//
// and it is silent in a single-file program. This was measured on the installed 8.67.0 build
// rather than assumed to be a harness limit, and the measurement corrected the assumption: driven
// through the real rule with a real program and no `./foo` on disk, UPSTREAM IS SILENT TOO. So the
// case is not something our harness cannot express, it is a case whose verdict depends on a module
// that has to exist. The alias branch in the rule is ported and no fixture here reaches it; a
// harness that can write two files should assert it.
//
// Every row asserts the span, the message text, and the repaired source. The span is the QUALIFIER
// rather than the whole access, which is upstream's `node: qualifier`, and on a nested chain like
// `A.B.C.D` it is `A.B.C` rather than `A`. Asserting only the message id would leave both the span
// and the nesting suppression free to be wrong.
func TestNoUnnecessaryQualifierFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []noUnnecessaryQualifierFinding
		wantOutput   string
	}{
		{
			sourceText: "namespace A {\n  export type B = number;\n  const x: A.B = 3;\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "B"},
			},
			wantOutput: "namespace A {\n  export type B = number;\n  const x: B = 3;\n}\n",
		},
		{
			sourceText: "namespace A {\n  export const x = 3;\n  export const y = A.x;\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "x"},
			},
			wantOutput: "namespace A {\n  export const x = 3;\n  export const y = x;\n}\n",
		},
		{
			sourceText: "namespace A {\n  export type T = number;\n  export namespace B {\n    const x: A.T = 3;\n  }\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "T"},
			},
			wantOutput: "namespace A {\n  export type T = number;\n  export namespace B {\n    const x: T = 3;\n  }\n}\n",
		},
		{
			sourceText: "namespace A {\n  export namespace B {\n    export type T = number;\n    const x: A.B.T = 3;\n  }\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A.B", wantName: "T"},
			},
			wantOutput: "namespace A {\n  export namespace B {\n    export type T = number;\n    const x: T = 3;\n  }\n}\n",
		},
		{
			sourceText: "namespace A {\n  export namespace B.C {\n    export type D = number;\n    const x: A.B.C.D = 3;\n  }\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A.B.C", wantName: "D"},
			},
			wantOutput: "namespace A {\n  export namespace B.C {\n    export type D = number;\n    const x: D = 3;\n  }\n}\n",
		},
		{
			sourceText: "namespace A {\n  export namespace B {\n    export const x = 3;\n    const y = A.B.x;\n  }\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A.B", wantName: "x"},
			},
			wantOutput: "namespace A {\n  export namespace B {\n    export const x = 3;\n    const y = x;\n  }\n}\n",
		},
		{
			sourceText: "enum A {\n  B,\n  C = A.B,\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "B"},
			},
			wantOutput: "enum A {\n  B,\n  C = B,\n}\n",
		},
		{
			sourceText: "namespace Foo {\n  export enum A {\n    B,\n    C = Foo.A.B,\n  }\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "Foo.A", wantName: "B"},
			},
			wantOutput: "namespace Foo {\n  export enum A {\n    B,\n    C = B,\n  }\n}\n",
		},

		{
			// Measured, not imported. Two reportable qualifiers where the second is a TYPE
			// ARGUMENT of the first, which is the shape that decides whether declining to recurse
			// into a reported node is the same suppression upstream gets from its exit visitor.
			// If the type argument were inside the reported node's subtree, upstream would suppress
			// it and this port would too, and both would be wrong. Driven on the installed 8.67.0
			// build: two findings at columns 70 and 74, and the fix rewrites both.
			sourceText: "namespace A { export type B<T> = T; export type C = number; const x: A.B<A.C> = 3; }\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "B"},
				{wantSpan: "A", wantName: "C"},
			},
			wantOutput: "namespace A { export type B<T> = T; export type C = number; const x: B<C> = 3; }\n",
		},

		{
			// Measured, not imported: this is the ALIAS branch, which the doc comment previously
			// called ported-and-untested. `import B = A` makes `B` an alias symbol whose own
			// declaration is the import rather than the namespace, so the identity loop misses and
			// the recursion through GetAliasedSymbol is the only thing that finds it. Without that
			// branch this case is silent. Driven on the installed 8.67.0 build: reports at line 6
			// column 13, and the repair drops the qualifier.
			sourceText: "namespace A {\n  export const x = 3;\n}\nimport B = A;\nnamespace A {\n  const y = B.x;\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "B", wantName: "x"},
			},
			wantOutput: "namespace A {\n  export const x = 3;\n}\nimport B = A;\nnamespace A {\n  const y = x;\n}\n",
		},
		{
			// Measured: the qualifier names an OUTER namespace from inside a nested one, so the
			// stack has to hold every enclosing declaration rather than only the innermost.
			sourceText: "namespace A {\n  export const x = 3;\n  export namespace B {\n    const y = A.x;\n  }\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "x"},
			},
			wantOutput: "namespace A {\n  export const x = 3;\n  export namespace B {\n    const y = x;\n  }\n}\n",
		},
		{
			// Measured: two declarations of the same namespace merge, and the access is in the
			// second while the declaration is in the first. The symbol carries both declarations,
			// so a port indexing Declarations[0] instead of looping would still pass here by luck;
			// what this pins is that the stack comparison is by NODE IDENTITY against whichever
			// declaration we are lexically inside.
			sourceText: "namespace A {\n  export const x = 3;\n}\nnamespace A {\n  const y = A.x;\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "x"},
			},
			wantOutput: "namespace A {\n  export const x = 3;\n}\nnamespace A {\n  const y = x;\n}\n",
		},

		{
			// Measured, not imported: the corpus writes no parenthesized form anywhere, so it has
			// no opinion and guessing costs a divergence either way. Driven on the installed
			// 8.67.0 build, all four parenthesized shapes REPORT there, and our parser declines
			// two of them without the unwrap because estree has no parenthesized node for
			// upstream's selector to trip over.
			//
			// The parens here enclose the whole edit, so upstream's fixer is safe and this row
			// asserts the repair.
			sourceText: "namespace A {\n  export namespace B {\n    export const x = 3;\n  }\n  const y = (A.B).x;\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "B"},
			},
			wantOutput: "namespace A {\n  export namespace B {\n    export const x = 3;\n  }\n  const y = (B).x;\n}\n",
		},
		{
			// A parenthesized type reference. The parens sit above the qualified name, so nothing
			// is unwrapped here and the repair is safe. Measured reporting upstream at column 13.
			sourceText: "namespace A {\n  export type T = number;\n  const x: (A.T) = 3;\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "T"},
			},
			wantOutput: "namespace A {\n  export type T = number;\n  const x: (T) = 3;\n}\n",
		},
		{
			// The two shapes where upstream's own fixer breaks the file. Both REPORT upstream and
			// report here; the repair is withheld. Measured outputs from the installed build:
			//
			//	(A).x     ->  `const y = (x;`     Parsing error: ')' expected
			//	((A).B).x ->  `const y = ((B).x;` Parsing error: ')' expected
			//
			// wantOutput is the source unchanged, which is what ExpectFixedSource compares when a
			// finding carries no fix, and is the assertion that would fail if the fix came back.
			sourceText: "namespace A {\n  export const x = 3;\n  const y = (A).x;\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "x", wantNoFix: true},
			},
			wantOutput: "namespace A {\n  export const x = 3;\n  const y = (A).x;\n}\n",
		},
		{
			sourceText: "namespace A {\n  export namespace B {\n    export const x = 3;\n  }\n  const y = ((A).B).x;\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "B", wantNoFix: true},
			},
			wantOutput: "namespace A {\n  export namespace B {\n    export const x = 3;\n  }\n  const y = ((A).B).x;\n}\n",
		},
		{
			// Nested parentheses, which is why the unwrap is a loop rather than one step.
			sourceText: "namespace A {\n  export const x = 3;\n  const y = ((A)).x;\n}\n",
			wantFindings: []noUnnecessaryQualifierFinding{
				{wantSpan: "A", wantName: "x", wantNoFix: true},
			},
			wantOutput: "namespace A {\n  export const x = 3;\n  const y = ((A)).x;\n}\n",
		},
	}
	for index, testCase := range cases {
		t.Run(noUnnecessaryQualifierCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnnecessaryQualifier, noUnnecessaryQualifierFile,
				testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position := range testCase.wantFindings {
				wantIds[position] = "unnecessaryQualifier"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// The harness writes the fixture trimmed, so both the span slice and the expected
			// repaired source are against that text rather than the Go literal above.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]
				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				wantMessage := "Qualifier is unnecessary since '" + want.wantName + "' is in scope."
				if diagnostic.Message.Description != wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, wantMessage,
						diagnostic.Message.Description)
				}

				// Whether the finding carries a repair is a judgment of its own, and the repaired
				// source alone cannot see it: a declined fix and a fix that happens to write the
				// original bytes produce the same output.
				wantFixes := 1
				if want.wantNoFix {
					wantFixes = 0
				}
				if len(diagnostic.Fixes) != wantFixes {
					t.Fatalf("finding %d fixes: expected %d, got %d", position, wantFixes,
						len(diagnostic.Fixes))
				}
			}
			// ExpectFixedSource refuses a result carrying no fixes, which is correct: there is no
			// rewrite to compare. The declined-fix rows are already pinned by the fix-count
			// assertion above and by wantOutput being the source unchanged, so they say so here
			// rather than being padded into a shape the helper accepts.
			anyFixExpected := false
			for _, want := range testCase.wantFindings {
				if !want.wantNoFix {
					anyFixExpected = true
				}
			}
			if anyFixExpected {
				rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(testCase.wantOutput)+"\n")
			} else if strings.TrimSpace(testCase.wantOutput)+"\n" != strings.TrimSpace(testCase.sourceText)+"\n" {
				t.Fatalf("a row expecting no fix must expect the source unchanged")
			}
		})
	}
}
