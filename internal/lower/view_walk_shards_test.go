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

// Not parallel: runViewWalkShard sets viewWalkCompare, which changes every view in the package.
func TestViewWalkIndexedAgreesWithWhole_0(t *testing.T) { runViewWalkShard(t, 0) }

// Not parallel: runViewWalkShard sets viewWalkCompare, which changes every view in the package.
func TestViewWalkIndexedAgreesWithWhole_1(t *testing.T) { runViewWalkShard(t, 1) }

// Not parallel: runViewWalkShard sets viewWalkCompare, which changes every view in the package.
func TestViewWalkIndexedAgreesWithWhole_2(t *testing.T) { runViewWalkShard(t, 2) }

// Not parallel: runViewWalkShard sets viewWalkCompare, which changes every view in the package.
func TestViewWalkIndexedAgreesWithWhole_3(t *testing.T) { runViewWalkShard(t, 3) }

// Every program the view walk comparison lowers falls in exactly one shard, and every shard has exactly one test,
// read from this file's own wrappers as TestCycleFinderEqualityPartition reads its, so deleting a wrapper or changing
// the partition can't drop a program or lower one twice. It lowers nothing: the shards' "N of M programs lowered with
// views" lines are what sum to the whole. Mutant: viewWalkShardPrograms leaving out shard 3's last program, and the
// program it leaves out falls in no shard.
func TestViewWalkShardsCoverEveryProgram(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), "view_walk_shards_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	names, err := viewWalkPrograms()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 4000 {
		t.Fatalf("the view walk comparison's programs shrank to %d", len(names))
	}
	wrappers := make(map[uint64]int)
	covered := make(map[string]int)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestViewWalkIndexedAgreesWithWhole_") {
			continue
		}
		shard := viewWalkWrapperShard(t, function)
		wrappers[shard]++
		for _, name := range viewWalkShardPrograms(names, shard) {
			covered[name]++
		}
	}
	for shard := uint64(0); shard < viewWalkShardCount; shard++ {
		if wrappers[shard] != 1 {
			t.Errorf("shard %d has %d wrappers, want exactly one", shard, wrappers[shard])
		}
	}
	if len(covered) != len(names) {
		t.Errorf("the shards cover %d distinct programs, the comparison has %d", len(covered), len(names))
	}
	for _, name := range names {
		if covered[name] != 1 {
			t.Errorf("%s falls in %d shards, want exactly one", name, covered[name])
		}
	}
	t.Logf("all %d programs fall in exactly one of %d shard wrappers", len(names), len(wrappers))
}

// viewWalkWrapperShard is the shard a wrapper lowers, refusing a wrapper that does anything but call
// runViewWalkShard with t and its own name's shard number.
func viewWalkWrapperShard(t *testing.T, function *ast.FuncDecl) uint64 {
	t.Helper()
	if len(function.Body.List) == 1 {
		if statement, ok := function.Body.List[0].(*ast.ExprStmt); ok {
			if call, ok := statement.X.(*ast.CallExpr); ok && len(call.Args) == 2 {
				name, named := call.Fun.(*ast.Ident)
				receiver, received := call.Args[0].(*ast.Ident)
				literal, numbered := call.Args[1].(*ast.BasicLit)
				if named && name.Name == "runViewWalkShard" && received && receiver.Name == "t" && numbered && literal.Kind == token.INT {
					shard, err := strconv.ParseUint(literal.Value, 10, 64)
					if err == nil && shard < viewWalkShardCount && function.Name.Name == fmt.Sprintf("TestViewWalkIndexedAgreesWithWhole_%d", shard) {
						return shard
					}
				}
			}
		}
	}
	t.Fatalf("%s: want runViewWalkShard(t, its own shard number) and nothing else", function.Name.Name)
	return 0
}
