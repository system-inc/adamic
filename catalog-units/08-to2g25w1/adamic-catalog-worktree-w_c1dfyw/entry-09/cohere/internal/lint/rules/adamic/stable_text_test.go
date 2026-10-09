package adamic

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// stableTextFiles are a file whose array literal makes the checker normalize `extra` and `note` first, and the
// case: its first column lacks `extra`, which only the second declares, so normalization adds `extra?: undefined`
// to it. TypeScript caches that member by name, per checker, with the declaration of whichever sibling it was first
// made from, and prints members in declaration order (#dq881as). Met first in Another.ts, which comes before the
// case in the program, the member prints first; met only in the case, it prints after the column's own members.
// The function hands back rows it holds, so invariant-mutable reports it: one returning the literal itself builds what
// it returns, which nobody else holds (#jpdf48x).
var stableTextFiles = map[string]string{
	"Another.ts": "export const others = [{ extra: 1 }, { note: 'other' }];\n",
	"Case.ts": `interface Column { id: string; extra?: number; note?: string; label?: string }
const rows = [{ id: 'a', note: 'first' }, { id: 'b', extra: 2 }];
export const columns: () => Column[] = function() {
	return rows;
};
`,
}

// relationMessages walks Case.ts of files with the two relation rules on a program checked by one checker, after
// asking it about Another.ts first when anotherFirst is set, and returns each rule's message there.
func relationMessages(t *testing.T, files map[string]string, anotherFirst bool) map[string]string {
	t.Helper()
	directory := t.TempDir()
	writeAdamicSupport(t)(directory)
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(directory, "tsconfig.json"), SingleThreaded: true})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	another := graph.Program.GetSourceFile(tspath.RootedFilePath(filepath.Join(directory, "Another.ts")))
	caseFile := graph.Program.GetSourceFile(tspath.RootedFilePath(filepath.Join(directory, "Case.ts")))
	if another == nil || caseFile == nil {
		t.Fatal("the fixture's files are not in the program")
	}
	if anotherFirst {
		graph.Program.GetSemanticDiagnostics(context.Background(), another)
	}
	result, err := graph.Walk(context.Background(), []*ast.SourceFile{caseFile}, []rule.Rule{InvariantMutable, NoOptionalWidening})
	if err != nil {
		t.Fatalf("walking: %v", err)
	}
	messages := map[string]string{}
	for _, diagnostic := range result.Diagnostics {
		messages[diagnostic.RuleName] = diagnostic.Message.Description
	}
	return messages
}

// TestTheRelationRulesPrintTheSameTextWhateverTheCheckerMetFirst: the two relation rules name the same types in
// the same words whether the checker normalized another file's literals first or not, which at GOMAXPROCS=16 is
// a matter of scheduling (#dq881as). TypeToString printed `extra?: never` first in one and last in the other.
// The fixture is held to that: each rule must report in both orders, or the comparison proves nothing.
func TestTheRelationRulesPrintTheSameTextWhateverTheCheckerMetFirst(t *testing.T) {
	t.Parallel()
	alone, afterAnother := relationMessages(t, stableTextFiles, false), relationMessages(t, stableTextFiles, true)
	for _, ruleName := range []string{InvariantMutable.Name, NoOptionalWidening.Name} {
		if alone[ruleName] == "" || afterAnother[ruleName] == "" {
			t.Fatalf("%s did not report in both orders (%q, %q), so this compares nothing", ruleName, alone[ruleName], afterAnother[ruleName])
		}
		if alone[ruleName] != afterAnother[ruleName] {
			t.Errorf("%s reads differently once the checker met Another.ts first:\n  alone:  %s\n  after:  %s", ruleName, alone[ruleName], afterAnother[ruleName])
		}
	}
}

// nestedTextFiles put the borrowed sibling inside the case's own literal (#ncz8caa). The first column's `meta` holds an
// array whose second literal lacks `extra`, so widening the column widens that array first and makes TypeScript's
// cached `extra?: undefined` from the `extra` nested there. The column then borrows it too: its declaration lies
// inside the column's source, ahead of `id`, but it isn't the column's member. Met first in Another.ts, it borrows
// that file's instead. Own by range read the nested one as the column's own and printed it second; own by parent
// reads neither as own, and prints it after the column's own members either way.
var nestedTextFiles = map[string]string{
	"Another.ts": stableTextFiles["Another.ts"],
	"Case.ts": `interface Column { meta?: unknown; id: string; extra?: number; note?: string; label?: string }
const rows = [{ meta: [{ extra: 1 }, { other: 2 }], id: 'a', note: 'first' }, { id: 'b', extra: 2 }];
export const columns: () => Column[] = function() {
	return rows;
};
`,
}

// TestOwnMembersAreTheContainersNotWhatItsRangeHolds: with the borrowed sibling nested inside the case's own literal,
// the two relation rules still read the same whether the checker met Another.ts first or not, and the column's own
// members keep their written order with the borrowed one after them (#ncz8caa). Own by range printed
// `meta; extra?; id; note` alone and `meta; id; note; extra?` after Another.ts; own by nothing would print them all
// by name, `extra?; id; meta; note`.
func TestOwnMembersAreTheContainersNotWhatItsRangeHolds(t *testing.T) {
	t.Parallel()
	alone, afterAnother := relationMessages(t, nestedTextFiles, false), relationMessages(t, nestedTextFiles, true)
	// Under Adamic's exactOptionalPropertyTypes a synthesized member is missing-only, and prints `never` (#sp4xwtj).
	const column = "id: string; note: string; extra?: never; }"
	for _, ruleName := range []string{InvariantMutable.Name, NoOptionalWidening.Name} {
		if alone[ruleName] == "" || afterAnother[ruleName] == "" {
			t.Fatalf("%s did not report in both orders (%q, %q), so this compares nothing", ruleName, alone[ruleName], afterAnother[ruleName])
		}
		if alone[ruleName] != afterAnother[ruleName] {
			t.Errorf("%s reads differently once the checker met Another.ts first:\n  alone:  %s\n  after:  %s", ruleName, alone[ruleName], afterAnother[ruleName])
		}
		if !strings.Contains(alone[ruleName], column) {
			t.Errorf("%s doesn't print the column's own members in their written order, then the borrowed one (%q):\n  %s", ruleName, column, alone[ruleName])
		}
	}
}

// TestAnOptionalMemberPrintsWithoutTheMissingTypeItsQuestionMarkSays: an optional member prints the type TypeScript's
// own printer starts from (#sp4xwtj). Under Adamic's options, exactOptionalPropertyTypes, `y?: number` prints as
// written, where 632001cb printed the missing marker as `y?: number | undefined`, and a written `| undefined` is kept,
// since there it is a different type. Without the option the two are one type, and it prints with `| undefined`.
func TestAnOptionalMemberPrintsWithoutTheMissingTypeItsQuestionMarkSays(t *testing.T) {
	t.Parallel()
	withoutExactOptional := strings.Replace(adamicConfiguration, `"exactOptionalPropertyTypes": true,`, "", 1)
	if withoutExactOptional == adamicConfiguration {
		t.Fatal("Adamic's configuration no longer sets exactOptionalPropertyTypes where this test removes it")
	}
	cases := []struct {
		name, configuration, written, printed string
	}{
		{"exact, written number", adamicConfiguration, "y?: number", "y?: number; }"},
		{"exact, written number or undefined", adamicConfiguration, "y?: number | undefined", "y?: number | undefined; }"},
		{"not exact, written number", withoutExactOptional, "y?: number", "y?: number | undefined; }"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			messages := relationMessages(t, map[string]string{
				"tsconfig.json": testCase.configuration,
				"Another.ts":    "export {};\n",
				"Case.ts":       "declare const narrow: { x: number }; function draw(point: { x: number; " + testCase.written + " }): void {} draw(narrow);\n",
			}, false)
			message := messages[NoOptionalWidening.Name]
			if want := "seen here as '{ x: number; " + testCase.printed + "'"; !strings.Contains(message, want) {
				t.Errorf("with %s written, want %q in:\n  %s", testCase.written, want, message)
			}
		})
	}
}

// TestASymbolKeyedMemberPrintsWithoutItsCreationId: a member keyed by a unique symbol reads `[brand]` in both relation
// rules' text, never typescript-go's internal `\xfe@brand@id`, whose id is the symbol's creation order (#z9jcxp1).
func TestASymbolKeyedMemberPrintsWithoutItsCreationId(t *testing.T) {
	t.Parallel()
	messages := relationMessages(t, map[string]string{
		"Another.ts": "export {};\n",
		"Case.ts":    "declare const brand: unique symbol; declare const narrow: { x: number }; function draw(point: { x: number; [brand]?: string }): void {} draw(narrow);\n",
	}, false)
	message := messages[NoOptionalWidening.Name]
	for _, want := range []string{"seen here as '{ x: number; [brand]?: string; }'", "the optional property `[brand]`"} {
		if !strings.Contains(message, want) {
			t.Errorf("want %q in:\n  %q", want, message)
		}
	}
	if strings.Contains(message, "\xfe") || strings.Contains(message, `\xfe`) {
		t.Errorf("the internal name reached the text:\n  %q", message)
	}
}
