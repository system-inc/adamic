package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestQualifiedPredicateCallsLower(t *testing.T) {
	t.Parallel()
	for _, shape := range []string{"namespace", "type", "each", "defined", "reexport", "module", "nested", "false_read"} {
		for _, suffix := range []string{"_truth", "_lie"} {
			t.Run(shape+suffix, func(t *testing.T) {
				path := filepath.Join("..", "oracle", "testdata", "predicate_qualified", shape+suffix, "main.ts")
				program, err := load.Load([]string{path})
				if err != nil {
					t.Fatal(err)
				}
				if _, err = Lower(context.Background(), program); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestQualifiedPredicateValueAndWriteRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ path, reason string }{
		{"impostor.ts", "an indirect call of a checked predicate overload"},
		{"namespace_typed_impostor.ts", "an indirect call of a checked predicate overload"},
		{"mutable.ts", "an indirect call of a checked predicate"},
		{"mutable_alias.ts", "an indirect call of a checked predicate"},
		{"member_write.a", "replacing a namespace export; exported functions and constants have fixed identity"},
		{"namespace_import/main.a", "TS2540: Cannot assign to 'f' because it is a read-only property"},
		{"named_import/main.a", "TS2632: Cannot assign to 'f' because it is an import"},
	} {
		t.Run(probe.path, func(t *testing.T) {
			program, err := load.Load([]string{filepath.Join("testdata/predicate_qualified_refused", probe.path)})
			if err == nil {
				_, err = Lower(context.Background(), program)
			}
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %q, got %v", probe.reason, err)
			}
		})
	}
}

// This is the false-side fixture with the walk live, not merely a qualified call
// whose predicate has a body proof. The identifier-only valueCall mutant stops
// its n.y read as NotYet, although the static call checks both directions.
func TestQualifiedPredicateFalseSideReadLowers(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "oracle", "testdata", "predicate_qualified", "false_read_truth", "main.ts")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = lowerTypeScriptAssertionSource(t, string(source)); err != nil {
		t.Fatal(err)
	}
}

// A qualified call alone must not turn on the value-escape switch through the
// carrier's type walk after escapingReference has correctly called it direct.
func TestQualifiedPredicateCallsAreNotValueEscapes(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.ts")
	source := `namespace ts { export function isNumber(v: number | string): v is number { return true; } } function use(v: number | string): void { if (ts.isNumber(v)) console.log("yes"); } use(7);`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checker, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	lowering := &lowering{program: program, checker: checker, result: &ir.Program{}}
	walk := predicateFalseFlow{l: lowering, checked: map[*ast.Node]bool{}}
	if walk.escapes() {
		t.Fatal("a direct qualified call was counted as an escaped predicate value")
	}
}
