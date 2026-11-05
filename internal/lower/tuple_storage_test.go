package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// A destructuring read of a tuple's element is judged once, in tupleField, by how the element is stored
// (tupleElementStorage): a literal written into the pattern is stored as the pattern's slots, not as the
// checker narrows each element, and reading it by the checker's tuple printed a box's pointer bits
// as a number (#fsb6n82). The walk names every function that calls tupleField, so a new caller is
// seen here and is made to hand tupleField its lowered value, and that no caller reads the checker's
// element types to pass in instead.
func TestTupleFieldCallersReadStorage(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	callers := map[string]bool{}
	storageCallers := map[string]bool{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			readsElementTypes := false
			callsTupleField := false
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, isCall := node.(*ast.CallExpr)
				if !isCall {
					return true
				}
				selector, isSelector := call.Fun.(*ast.SelectorExpr)
				if !isSelector {
					return true
				}
				switch selector.Sel.Name {
				case "tupleField":
					callsTupleField = true
					callers[function.Name.Name] = true
					if len(call.Args) != 6 {
						t.Errorf("%s: %s calls tupleField without its value node and lowered value", path, function.Name.Name)
					}
				case "tupleElementStorage":
					storageCallers[function.Name.Name] = true
				case "GetTypeArguments":
					readsElementTypes = true
				}
				return true
			})
			if callsTupleField && readsElementTypes {
				t.Errorf("%s: %s reads a tuple's element types beside tupleField, which judges them by storage", path, function.Name.Name)
			}
		}
	}
	if got := sortedCallers(callers); strings.Join(got, ",") != "destructuringAssignment" {
		t.Errorf("tupleField's callers are %v; a new one passes the value it lowered, and is named here once it does", got)
	}
	if got := sortedCallers(storageCallers); strings.Join(got, ",") != "tupleField" {
		t.Errorf("tupleElementStorage's callers are %v, want tupleField alone, the one place the judgment is made", got)
	}
}

func sortedCallers(names map[string]bool) []string {
	sorted := []string{}
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	return sorted
}
