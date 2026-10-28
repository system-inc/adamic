package lower

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

// Read the wrappers that go test actually discovers, rather than assuming that
// every shard number has a test. Check the same path selection the runners use:
// deleting a wrapper or changing the partition must not lose or repeat a path.
func TestCycleFinderEqualityPartition(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), "cycles_equality_shards_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := cycleComparisonInputs()
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 3500 {
		t.Fatalf("cycle comparison inventory shrank: %d", len(paths))
	}
	wrappers := make(map[uint64]int)
	covered := make(map[string]int)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestCycleFinderEquality_") {
			continue
		}
		arguments := cycleEqualityWrapperArguments(t, function, "runCycleFinderEqualityShard")
		literal, ok := arguments[1].(*ast.BasicLit)
		if !ok || literal.Kind != token.INT {
			t.Fatalf("%s: shard must be an integer literal", function.Name.Name)
		}
		shard, err := strconv.ParseUint(literal.Value, 10, 64)
		if err != nil || shard >= cycleEqualityShardCount || function.Name.Name != fmt.Sprintf("TestCycleFinderEquality_%03d", shard) {
			t.Fatalf("%s: invalid shard %s", function.Name.Name, literal.Value)
		}
		wrappers[shard]++
		for _, path := range cycleComparisonShardPaths(paths, shard) {
			covered[path]++
		}
	}
	for shard := uint64(0); shard < cycleEqualityShardCount; shard++ {
		if wrappers[shard] != 1 {
			t.Errorf("shard %d has %d wrappers, want exactly one", shard, wrappers[shard])
		}
	}
	if len(covered) != len(paths) {
		t.Errorf("partition covers %d distinct paths, inventory has %d", len(covered), len(paths))
	}
	for _, path := range paths {
		if covered[path] != 1 {
			t.Errorf("%s has %d comparisons, want exactly one", path, covered[path])
		}
	}
	t.Logf("all %d inventory paths occur exactly once across %d actual shard wrappers", len(paths), len(wrappers))
}

func cycleEqualityWrapperArguments(t *testing.T, function *ast.FuncDecl, helper string) []ast.Expr {
	t.Helper()
	if len(function.Body.List) == 2 && isParallelCall(function.Body.List[0]) {
		if expression, ok := function.Body.List[1].(*ast.ExprStmt); ok {
			if call, ok := expression.X.(*ast.CallExpr); ok && len(call.Args) == 2 {
				name, named := call.Fun.(*ast.Ident)
				receiver, received := call.Args[0].(*ast.Ident)
				if named && name.Name == helper && received && receiver.Name == "t" {
					return call.Args
				}
			}
		}
	}
	t.Fatalf("%s: want t.Parallel() then %s(t, literal) and nothing else", function.Name.Name, helper)
	return nil
}

func TestCycleEqualityProductsHaveBuilders(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), "cycles_equality_products_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := cycleComparisonInputs()
	if err != nil {
		t.Fatal(err)
	}
	inventory := make(map[string]bool)
	for _, path := range paths {
		inventory[path] = true
	}
	builders := make(map[string]int)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestProduct_CycleEquality") {
			continue
		}
		arguments := cycleEqualityWrapperArguments(t, function, "cycleEqualityProduct")
		literal, ok := arguments[1].(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			t.Fatalf("%s: path must be a string literal", function.Name.Name)
		}
		path, err := strconv.Unquote(literal.Value)
		if err != nil || !inventory[path] || function.Name.Name != "TestProduct_CycleEquality"+cycleEqualityProducts[path] {
			t.Fatalf("%s: invalid product path %s", function.Name.Name, literal.Value)
		}
		builders[path]++
	}
	for path := range cycleEqualityProducts {
		if builders[path] != 1 {
			t.Errorf("%s has %d product builders, want exactly one", path, builders[path])
		}
	}
}
