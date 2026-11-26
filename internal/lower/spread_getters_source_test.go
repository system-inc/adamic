package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// spreadGettersCallers are the functions that read spreadGetters' answer, each with what it does with it.
// SpreadGetters is read past lowering by native/reuse.go's touches and worker/purity.go's CheckPure, which
// take the set as lowering wrote it in finishAccessors.
var spreadGettersCallers = map[string]string{
	"finishAccessors":   "writes ir.ObjectLiteral's SpreadGetters, which reuse and purity read",
	"spreadCopyGetters": "spreadCopyRead's stop on a class getter's name read from a spread's copy",
}

// classHolderReaders are the functions that read a class's holder type, each with how it reads it.
var classHolderReaders = map[string]string{
	"spreadGetters": "relates it to the spread's view, through relatedShape",
	"slotWriteSite": "takes an object literal's own type as its layout, never relating it to a view",
}

// A fresh object literal's type fails the checker's relation by excess properties, so a holder related
// raw leaves a literal's getter out of the set a spread may run, and reuse lent a global past the getter
// that replaced it (#z6h9s7k). Every reader of spreadGetters' answer and every reader of a class's
// holder is listed here, and spreadGetters relates each holder through relatedShape, so a new reader is
// a red test until it says how it reads the holder. Mutant (spread_getters.mutants.md): relatedShape
// dropped from spreadGetters' relation.
func TestSpreadGettersHolderReadersAreListed(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	callers, readers := map[string]bool{}, map[string]bool{}
	relations, related := 0, 0
	files := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.CallExpr:
					if selector, isSelector := node.Fun.(*ast.SelectorExpr); isSelector {
						if selector.Sel.Name == "spreadGetters" {
							callers[function.Name.Name] = true
							if spreadGettersCallers[function.Name.Name] == "" {
								t.Errorf("%s: %s reads spreadGetters' answer and isn't in spreadGettersCallers", files.Position(node.Pos()), function.Name.Name)
							}
						}
						if selector.Sel.Name == "IsTypeAssignableTo" && function.Name.Name == "spreadGetters" && len(node.Args) == 2 {
							relations++
							if argument, isCall := node.Args[0].(*ast.CallExpr); isCall {
								if inner, isSelector := argument.Fun.(*ast.SelectorExpr); isSelector && inner.Sel.Name == "relatedShape" {
									related++
								}
							}
						}
					}
				case *ast.IndexExpr:
					if selector, isSelector := node.X.(*ast.SelectorExpr); isSelector && selector.Sel.Name == "classHolders" {
						readers[function.Name.Name] = true
						if classHolderReaders[function.Name.Name] == "" {
							t.Errorf("%s: %s reads a class's holder and isn't in classHolderReaders", files.Position(node.Pos()), function.Name.Name)
						}
					}
				}
				return true
			})
		}
	}
	for listed := range spreadGettersCallers {
		if !callers[listed] {
			t.Errorf("spreadGettersCallers lists %s, which no longer calls spreadGetters", listed)
		}
	}
	for listed := range classHolderReaders {
		if !readers[listed] {
			t.Errorf("classHolderReaders lists %s, which no longer reads a class's holder", listed)
		}
	}
	if relations != 1 || related != relations {
		t.Errorf("spreadGetters relates a holder %d times, %d of them through relatedShape; want its one relation, through relatedShape", relations, related)
	}
}
