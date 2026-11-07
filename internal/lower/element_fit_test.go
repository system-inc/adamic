package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// A container's elements are judged against its slot once, in elementsHeldOtherwise, where expression
// hands every lowered value to the slot it goes into (#b8fwyfx): an object literal's field, a declaration,
// an assignment, an argument, a return and an element all pass there. The walk names every function that
// calls it, so a new writer that hands a value to a slot some other way is seen here and made to judge it
// the same way, and that only elementsHeldOtherwise walks the elements (elementsSeenAs), so no second
// judgment drifts from the first.
func TestElementsHeldOtherwiseIsJudgedOnce(test *testing.T) {
	test.Parallel()
	files, globError := filepath.Glob("*.go")
	if globError != nil {
		test.Fatal(globError)
	}
	callers := map[string]bool{}
	walkCallers := map[string]bool{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, parseFileError := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if parseFileError != nil {
			test.Fatal(parseFileError)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
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
				case "elementsHeldOtherwise":
					callers[function.Name.Name] = true
				case "elementsSeenAs":
					if function.Name.Name != "elementsSeenAs" {
						walkCallers[function.Name.Name] = true
					}
				}
				return true
			})
		}
	}
	if got := sortedCallers(callers); strings.Join(got, ",") != "expression" {
		test.Errorf(
			"elementsHeldOtherwise's callers are %v; a new one hands a value to its slot, and is named here once it judges it there",
			got,
		)
	}
	// fieldsSeenAs is the walk's own step through fields and dictionary values, which goes back into it.
	if got := sortedCallers(walkCallers); strings.Join(got, ",") != "elementsHeldOtherwise,fieldsSeenAs" {
		test.Errorf(
			"elementsSeenAs's callers are %v, want elementsHeldOtherwise and the walk's own fieldsSeenAs, the one place the judgment is made",
			got,
		)
	}
}
