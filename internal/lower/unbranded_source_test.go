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

// unbrandedCheckerArrayCalls are the checker's array questions that see a phantom brand as a non-array.
var unbrandedCheckerArrayCalls = []string{"IsArrayType", "GetElementTypeOfArrayType", "Checker_isArrayOrTupleType"}

// unbrandedAllowed is each file that may ask the checker's array questions directly, with how many times
// and why: every other file asks isArrayType and elementTypeOfArray, which see a phantom extension as its
// base (#gb0fqaa).
var unbrandedAllowed = map[string]struct {
	calls int
	why   string
}{
	"unbranded.go":  {3, "isArrayType, elementTypeOfArray and unbranded are the one place lowering asks the checker"},
	"cast_proof.go": {1, "arrayLikeType judges a cast, which sees the brand: a plain array never relates to a branded one by layout"},
}

// Lowering decides an array's layout through unbranded.go alone, so a phantom extension is held as its
// base everywhere it's held. The test reads every non-test file of the package and fails on a direct call
// of the checker's array questions outside unbrandedAllowed, and on an allowance no longer used as given.
// Mutant: one raw l.checker.IsArrayType put back in representation (expression.go); this goes red.
func TestArrayQuestionsGoThroughUnbranded(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	counted := map[string]int{}
	var direct []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, isCall := node.(*ast.CallExpr)
			if !isCall {
				return true
			}
			selector, isSelector := call.Fun.(*ast.SelectorExpr)
			if !isSelector || !slices.Contains(unbrandedCheckerArrayCalls, selector.Sel.Name) {
				return true
			}
			counted[entry.Name()]++
			if _, allowed := unbrandedAllowed[entry.Name()]; !allowed {
				direct = append(direct, files.Position(call.Pos()).String()+": "+selector.Sel.Name)
			}
			return true
		})
	}
	if len(direct) > 0 {
		t.Errorf("lowering asks the checker's array questions directly, which see a phantom brand as no array; ask l.isArrayType or l.elementTypeOfArray (unbranded.go):\n%s", strings.Join(direct, "\n"))
	}
	for _, name := range slices.Sorted(maps.Keys(unbrandedAllowed)) {
		if counted[name] != unbrandedAllowed[name].calls {
			t.Errorf("%s asks the checker's array questions %d times, but unbrandedAllowed allows %d (%s); update the allowance with why", name, counted[name], unbrandedAllowed[name].calls, unbrandedAllowed[name].why)
		}
	}
}
