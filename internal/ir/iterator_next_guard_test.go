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

// IteratorNext steps a generator as GeneratorNext does (#qk8rztp S2c-2b), so every switch in the
// compiler that has a case for GeneratorNext, by its type or by flow's generatorNextType, has one for
// IteratorNext too: a proof, an effect, a throw edge, an emitter. This is the second line behind the
// closed defaults (ir.Inert): a switch that gave GeneratorNext a case of its own did so because the
// default was wrong for a step, and it is wrong for IteratorNext for the same reason. Mutant: the
// IteratorNext case dropped from any one of them, such as internal/flow/infer.go's callsBack, and
// this names its file and line.
func TestEverySwitchOverGeneratorNextNamesIteratorNext(t *testing.T) {
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
			generator, iterator := false, false
			for _, clause := range body.List {
				for _, expression := range clause.(*ast.CaseClause).List {
					switch stepNamed(expression) {
					case "GeneratorNext":
						generator = true
					case "IteratorNext":
						iterator = true
					}
				}
			}
			if generator {
				switches++
				if !iterator {
					t.Errorf("%s: a switch with a case for GeneratorNext has none for IteratorNext", files.Position(node.Pos()))
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// The emitters, fresh, split, flow's effects and throws, lowering's proofs and the closed defaults'
	// own cases: fewer means the walk no longer finds the compiler.
	if switches < 14 {
		t.Fatalf("found %d switches with a case for GeneratorNext: the walk no longer finds the compiler's", switches)
	}
}

// stepNamed is the step a case expression names: GeneratorNext for ir.GeneratorNext or flow's
// generatorNextType, IteratorNext for their twins, and nothing for anything else.
func stepNamed(expression ast.Expr) string {
	switch expression := expression.(type) {
	case *ast.SelectorExpr:
		if name, isName := expression.X.(*ast.Ident); isName && name.Name == "ir" {
			return expression.Sel.Name
		}
	case *ast.Ident:
		switch expression.Name {
		case "GeneratorNext", "generatorNextType":
			return "GeneratorNext"
		case "IteratorNext", "iteratorNextType":
			return "IteratorNext"
		}
	}
	return ""
}
