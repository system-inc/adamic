package parallelcensus

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// No lookup whose only build is a "not built ahead" refusal goes through buildcache.Get in any _test.go: Get audits a
// listed hit and a sampled fetch with the lookup's own build, so a find's refusal read as a failed rebuild of a
// product that was there under its key (#v08876, internal/worker's wasm mutants under Loom 0.8.876's rekey).
// buildcache.BuiltAhead is the lookup that never audits. Mutant: one site moved back onto Get, red here.
func TestNoLookupThatOnlyRefusesGoesThroughGet(t *testing.T) {
	t.Parallel()
	// Reads: *_test.go
	root := parallelCheckRoot(t)
	positions := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "vendor" || entry.Name() == "testdata" || path == filepath.Join(root, "cohere") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(source), "not built ahead") {
			return nil
		}
		file, err := parser.ParseFile(positions, path, source, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, isCall := node.(*ast.CallExpr)
			if !isCall || len(call.Args) != 2 || !namesBuildcacheGet(call.Fun) {
				return true
			}
			literal, isLiteral := call.Args[1].(*ast.FuncLit)
			if !isLiteral {
				return true
			}
			ast.Inspect(literal.Body, func(inner ast.Node) bool {
				if text, isText := inner.(*ast.BasicLit); isText && text.Kind == token.STRING {
					if value, err := strconv.Unquote(text.Value); err == nil && strings.Contains(value, "not built ahead") {
						t.Errorf("%s: a lookup whose only build refuses (%q) goes through buildcache.Get; use buildcache.BuiltAhead", positions.Position(call.Pos()), value)
					}
				}
				return true
			})
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// namesBuildcacheGet is whether fun is buildcache.Get, called from another package or as Get within buildcache.
func namesBuildcacheGet(fun ast.Expr) bool {
	switch fun := fun.(type) {
	case *ast.SelectorExpr:
		if packageName, isIdent := fun.X.(*ast.Ident); isIdent {
			return packageName.Name == "buildcache" && fun.Sel.Name == "Get"
		}
	case *ast.Ident:
		return fun.Name == "Get"
	}
	return false
}
