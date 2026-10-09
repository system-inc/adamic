package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestNoCaseDeclarationsReportsLexicalDeclarations(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   int
	}{
		{"let", `switch (a) { case 1: let x = 1; break; }`, 1},
		{"const", `switch (a) { case 1: const x = 1; break; }`, 1},
		{"function", `switch (a) { case 1: function f() {} break; }`, 1},
		{"class", `switch (a) { case 1: class C {} break; }`, 1},
		{"default clause", `switch (a) { default: let x = 1; break; }`, 1},
		{"using", `switch (a) { case 1: using resource = acquire(); break; }`, 1},
		{"await using", `async function g() { switch (a) { case 1: await using resource = acquire(); break; } }`, 1},
		{"two declarations in one clause", `switch (a) { case 1: let x = 1; const y = 2; break; }`, 2},
		{"declarations in separate clauses", `switch (a) { case 1: let x = 1; break; case 2: const y = 2; break; }`, 2},
		{"a block earlier in the clause does not scope a later declaration", `switch (a) { case 1: {} function f() {} break; }`, 1},
		{"declaration in a fallen-through arm", `switch (a) { case 1: case 2: let x = 1; }`, 1},
		{"nested switch", `switch (a) { case 1: switch (b) { case 2: let x = 1; break; } break; }`, 1},
		{"multiple declarators in one statement is one statement", `switch (a) { case 1: let x = 1, y = 2; break; }`, 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoCaseDeclarations, "file.ts", testCase.source)
			wantIds := make([]string, testCase.want)
			for index := range wantIds {
				wantIds[index] = "unexpected"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

func TestNoCaseDeclarationsAcceptsScopedAndHoistedDeclarations(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{"let inside braces", `switch (a) { case 1: { let x = 1; break; } }`},
		{"const inside braces", `switch (a) { case 1: { const x = 1; break; } }`},
		{"function inside braces", `switch (a) { case 1: { function f() {} break; } }`},
		{"class inside braces", `switch (a) { case 1: { class C {} break; } }`},
		{"var in a case clause", `switch (a) { case 1: var x = 1; break; }`},
		{"var in a default clause", `switch (a) { default: var x = 1; break; }`},
		{"no declaration at all", `switch (a) { case 1: break; }`},
		{"empty switch", `switch (a) {}`},
		{"clause with no statements", `switch (a) { case 1: case 2: {} }`},
		{"declaration nested in an if", `switch (a) { case 1: if (a) { const x = 1; } break; }`},
		{"declaration nested in a loop", `switch (a) { case 1: for (const x of xs) {} break; }`},
		{"declaration inside a function body in the clause", `switch (a) { case 1: (() => { const x = 1; })(); break; }`},
		{"type and interface are erased", `switch (a) { case 1: type T = string; interface I { value: T } break; }`},
		{"enum and namespace emit into hoisted bindings", `switch (a) { case 1: enum E { A } namespace N {} break; }`},
		{"declare const emits nothing", `switch (a) { case 1: declare const d: number; break; }`},
		{"declaration outside the switch", `const outside = 1; switch (a) { case 1: break; }`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoCaseDeclarations, "file.ts", testCase.source))
		})
	}
}

// The finding must land on the declaration's own text. Anchoring at the node's Pos would start the
// range inside the comment above it, which puts the finding on a line a disable comment for the
// declaration cannot reach.
func TestNoCaseDeclarationsReportsTheDeclarationWithoutTrivia(t *testing.T) {
	t.Parallel()

	const source = `switch (a) {
  case 1:
    // explain the binding
    const reported = 2;
    break;
}`
	result := rule_testing.Run(t, NoCaseDeclarations, "file.ts", source)
	rule_testing.ExpectFindings(t, result, "unexpected")

	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "const reported = 2;" {
		t.Fatalf("reported range = %q, want %q", reported, "const reported = 2;")
	}
}

// Both declarations in one clause must offer the same whole-clause repair, anchored at the first and
// last statements rather than at the reported declaration. Anchoring on the reported statement would
// propose two nested blocks that leave `break` outside the scope it was meant to sit in.
func TestNoCaseDeclarationsSuggestsWrappingTheWholeClause(t *testing.T) {
	t.Parallel()

	const source = `switch (a) {
  case 1:
    let first = 1;
    const second = 2;
    break;
}`
	result := rule_testing.Run(t, NoCaseDeclarations, "file.ts", source)
	rule_testing.ExpectFindings(t, result, "unexpected", "unexpected")

	for index, diagnostic := range result.Diagnostics {
		if len(diagnostic.Suggestions) != 1 {
			t.Fatalf("finding %d: suggestions = %d, want 1", index, len(diagnostic.Suggestions))
		}
		suggestion := diagnostic.Suggestions[0]
		if suggestion.Message.Id != "addBraces" {
			t.Fatalf("finding %d: suggestion id = %q, want %q", index, suggestion.Message.Id, "addBraces")
		}
		if len(suggestion.Fixes) != 2 {
			t.Fatalf("finding %d: fixes = %d, want 2", index, len(suggestion.Fixes))
		}

		openingBrace, closingBrace := suggestion.Fixes[0], suggestion.Fixes[1]
		if openingBrace.Text != "{ " || closingBrace.Text != " }" {
			t.Fatalf("finding %d: fix texts = %q, %q", index, openingBrace.Text, closingBrace.Text)
		}
		// Insertions are empty ranges, so a position is enough to say where each one lands.
		if got := source[openingBrace.Range.Pos():]; len(got) < 3 || got[:3] != "let" {
			t.Fatalf("finding %d: opening brace does not land before the first statement", index)
		}
		if got := source[:closingBrace.Range.Pos()]; len(got) < 6 || got[len(got)-6:] != "break;" {
			t.Fatalf("finding %d: closing brace does not land after the last statement", index)
		}
	}
}
