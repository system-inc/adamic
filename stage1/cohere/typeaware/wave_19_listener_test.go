package typeaware

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// The independent frontend enum, not a second list of numeric literals, is the oracle.
func TestWave19NumericListenerDeclarations(t *testing.T) {
	rows := []struct {
		file  string
		kinds []int
	}{
		{"consistent_generic_constructors.a", []int{int(ast.KindVariableDeclaration), int(ast.KindPropertyDeclaration), int(ast.KindParameter)}},
		{"dot_notation.a", []int{int(ast.KindElementAccessExpression)}},
		{"no_array_constructor.a", []int{int(ast.KindCallExpression), int(ast.KindNewExpression)}},
		{"no_uncleared_race_timeout.a", []int{int(ast.KindCallExpression)}},
		{"no_process_exit_after_output.a", []int{int(ast.KindCallExpression)}},
		{"require_blocking_standard_streams.a", []int{int(ast.KindSourceFile)}},
	}
	sourceRoot := os.Getenv("ADAMIC_WAVE19_LISTENER_SOURCE")
	if sourceRoot == "" {
		sourceRoot = "."
	}
	declaration := regexp.MustCompile(`readonly listenerKinds: readonly number\[\] = (\[[0-9, ]+\]);`)
	for _, row := range rows {
		t.Run(row.file, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join(sourceRoot, row.file))
			if err != nil {
				t.Fatal(err)
			}
			matches := declaration.FindAllSubmatch(source, -1)
			if len(matches) != 1 {
				t.Fatalf("expected one numeric listener declaration, got %d", len(matches))
			}
			var got []int
			if err := json.Unmarshal(matches[0][1], &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, row.kinds) {
				t.Fatalf("numeric listeners differ from pinned Go SyntaxKind: got %v, want %v", got, row.kinds)
			}
		})
	}
}
