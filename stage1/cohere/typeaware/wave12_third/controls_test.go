package wave12third

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"
)

// Replay literal source rows without copying their expected diagnostics.
func referenceControls(t *testing.T, repository string) []string {
	t.Helper()
	var sources []string
	for _, file := range []string{"no_new_func_test.go", "no_new_native_nonconstructor_test.go", "no_new_wrappers_test.go"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/core", file), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		goast.Inspect(tree, func(n goast.Node) bool {
			row, ok := n.(*goast.CompositeLit)
			if !ok || len(row.Elts) < 2 {
				return true
			}
			// The first field is either a descriptive name or the source in span rows.
			at := 1
			if _, numeric := row.Elts[1].(*goast.BasicLit); numeric {
				if lit := row.Elts[1].(*goast.BasicLit); lit.Kind != token.STRING {
					at = 0
				}
			}
			lit, ok := row.Elts[at].(*goast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Fatal(err)
			}
			sources = append(sources, value)
			count++
			return false
		})
		if count == 0 {
			t.Fatal("empty extraction", file)
		}
		t.Logf("%s: %d literal reference rows", file, count)
	}
	return sources
}
