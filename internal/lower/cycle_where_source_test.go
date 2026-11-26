package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"
)

// cycleWhereAllowed is each function that may index the cycle finder's where directly, by file and name, with
// how many times and why: every other reader asks whereOf, which never hands a refusal a nil node (#ew207ss).
var cycleWhereAllowed = map[string]struct {
	indexes int
	why     string
}{
	"cycles.go use":     {2, "use tests whether a type was seen, by presence alone, and records where it was met, the one write"},
	"cycles.go whereOf": {1, "whereOf is the one reader that hands a node on, and falls back past a nil one"},
}

// A refusal's Where is read from the cycle finder's where only through whereOf, so a type met with no node
// is reported somewhere rather than handed to Program.Where as nil, which panicked on tsc (#ew207ss). The
// test reads every non-test file of the package and fails on an index of a field named where outside
// cycleWhereAllowed, on a range over one that binds its node, and on an allowance no longer used as given.
// It's syntactic: no other type here indexes a field named where, and one that comes to is listed with why.
// Mutant: one slotsOf refusal put back to l.program.Where(f.where[holder]); this goes red.
func TestCycleFinderWhereIsReadThroughWhereOf(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	counted := map[string]int{}
	var direct []string
	whereField := func(expression ast.Expr) bool {
		selector, isSelector := expression.(*ast.SelectorExpr)
		return isSelector && selector.Sel.Name == "where"
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			key := entry.Name() + " " + function.Name.Name
			ast.Inspect(function.Body, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.IndexExpr:
					if !whereField(node.X) {
						return true
					}
					counted[key]++
					if _, allowed := cycleWhereAllowed[key]; !allowed {
						direct = append(direct, files.Position(node.Pos()).String()+": "+function.Name.Name+" indexes where")
					}
				case *ast.RangeStmt:
					if whereField(node.X) && node.Value != nil {
						direct = append(direct, files.Position(node.Pos()).String()+": "+function.Name.Name+" ranges over where and binds its node")
					}
				}
				return true
			})
		}
	}
	if len(direct) > 0 {
		t.Errorf("the cycle finder's where is read directly, which may hand a refusal a nil node; ask f.whereOf (cycles.go):\n%s", strings.Join(direct, "\n"))
	}
	for _, key := range slices.Sorted(maps.Keys(cycleWhereAllowed)) {
		if counted[key] != cycleWhereAllowed[key].indexes {
			t.Errorf("%s indexes where %d times, but cycleWhereAllowed allows %d (%s); update the allowance with why", key, counted[key], cycleWhereAllowed[key].indexes, cycleWhereAllowed[key].why)
		}
	}
}
