package ir

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// FunctionLength is an inert numeric metadata read like Length. Keep explicit
// Length cases in compiler proofs and emitters from overlooking the new node.
func TestEverySwitchOverLengthNamesFunctionLength(t *testing.T) {
	t.Parallel()
	// Test files and testdata aside, every Go file of the compiler.
	// Reads: internal/**/*.go
	root := ".."
	files := token.NewFileSet()
	switches := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			var body *ast.BlockStmt
			switch statement := node.(type) {
			case *ast.SwitchStmt:
				body = statement.Body
			case *ast.TypeSwitchStmt:
				body = statement.Body
			default:
				return true
			}
			length, functionLength := false, false
			for _, clause := range body.List {
				for _, expression := range clause.(*ast.CaseClause).List {
					switch functionLengthNamed(expression) {
					case "Length":
						length = true
					case "FunctionLength":
						functionLength = true
					}
				}
			}
			if length {
				switches++
				if !functionLength {
					t.Errorf("%s: a switch with a case for Length has none for FunctionLength", files.Position(node.Pos()))
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// At least the emitters, freshness analysis, and explicit read proofs.
	if switches < 8 {
		t.Fatalf("found %d switches with a case for Length: the walk no longer finds the compiler's", switches)
	}
}

func functionLengthNamed(expression ast.Expr) string {
	switch expression := expression.(type) {
	case *ast.SelectorExpr:
		if name, ok := expression.X.(*ast.Ident); ok && name.Name == "ir" {
			return expression.Sel.Name
		}
	case *ast.Ident:
		return expression.Name
	}
	return ""
}
