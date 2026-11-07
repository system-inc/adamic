package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
	"testing"
)

// instanceTables are the tables a lowered instance's body is found by: a generic function's, by key, in
// the program's table or a class-generic call's bucket swapped in for it, and a class's.
var instanceTables = map[string]bool{"genericInstances": true, "classGenericInstances": true, "instances": true}

// instanceTableWriters are the functions that write an instance table without lowering an instance's body
// there, each with why, so they need no judgment of their own.
var instanceTableWriters = map[string]string{
	"classIdentity":    "an identity record for a class's checks, with no body: constructor and initializer are -1",
	"classGenericCall": "a bucket for one checker type vector, swapped in as genericInstances; the instance itself is written by instantiateFunctionSignature, which judges it",
}

// Every request for an instance is judged before the instance table is read, hit or miss, so a body
// shared by requests that differ is judged for each (Adamic's ruling on #c7w6j0y, Nov 1, with Lower's
// door test). The test reads this package's source: a function that writes to an instance table by key
// calls brandedRequest before its first read or write of one, or is named in instanceTableWriters with
// why it lowers no body. A third door, a generic method if one is ever admitted, then can't skip the
// judgment by writing a table of its own. Mutant: class.go's brandedRequest call removed, and instantiate
// is named here.
func TestBrandedRequestHoldsEveryInstanceDoor(test *testing.T) {
	test.Parallel()
	entries, readDirectoryError := os.ReadDir(".")
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	files := token.NewFileSet()
	var unjudged []string
	writers := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, parseFileError := parser.ParseFile(files, entry.Name(), nil, 0)
		if parseFileError != nil {
			test.Fatal(parseFileError)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			// The first access of a table, the first write, and the first judgment, by position.
			firstAccess, firstWrite, judged := token.NoPos, token.NoPos, token.NoPos
			ast.Inspect(function.Body, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.AssignStmt:
					for _, left := range node.Lhs {
						if instanceTableIndex(left) && (firstWrite == token.NoPos || left.Pos() < firstWrite) {
							firstWrite = left.Pos()
						}
					}
				case *ast.IndexExpr:
					if instanceTableIndex(node) && (firstAccess == token.NoPos || node.Pos() < firstAccess) {
						firstAccess = node.Pos()
					}
				case *ast.CallExpr:
					if selector, isSelector := node.Fun.(*ast.SelectorExpr); isSelector &&
						selector.Sel.Name == "brandedRequest" &&
						(judged == token.NoPos || node.Pos() < judged) {
						judged = node.Pos()
					}
				}
				return true
			})
			if firstWrite == token.NoPos {
				continue
			}
			name := function.Name.Name
			writers[name] = true
			if instanceTableWriters[name] != "" {
				continue
			}
			if judged == token.NoPos || judged > firstAccess {
				unjudged = append(unjudged, files.Position(firstWrite).String()+" in "+name)
			}
		}
	}
	sort.Strings(unjudged)
	if len(unjudged) > 0 {
		test.Errorf(
			"a function writes an instance table without calling brandedRequest before it reads one, so a request can reach a lowered body unjudged; judge the request first, or name the function in instanceTableWriters with why it lowers no body:\n%s",
			strings.Join(unjudged, "\n"),
		)
	}
	for name := range instanceTableWriters {
		if !writers[name] {
			test.Errorf("instanceTableWriters names %s, which writes no instance table; take it out", name)
		}
	}
	// The walk has to see the two doors write, or it proved nothing about them.
	for _, name := range []string{"instantiateFunctionSignature", "instantiate"} {
		if !writers[name] {
			test.Errorf("the walk never saw %s write an instance table", name)
		}
	}
}

// instanceTableIndex is whether node indexes an instance table: l.genericInstances[key] and the like.
func instanceTableIndex(node ast.Node) bool {
	index, isIndex := node.(*ast.IndexExpr)
	if !isIndex {
		return false
	}
	selector, isSelector := index.X.(*ast.SelectorExpr)
	return isSelector && instanceTables[selector.Sel.Name]
}
