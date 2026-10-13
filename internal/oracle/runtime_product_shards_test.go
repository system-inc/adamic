package oracle

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

// Inspect the actual selectable products, rather than a second registry that could drift.
// Mutants: an omitted product, a product calling the wrong shard, a dropped feature set,
// and complete archives combined into products that exceed the two-archive bound.
func TestRuntimeProductShardsCoverEveryArchive(t *testing.T) {
	t.Parallel()
	source, err := parser.ParseFile(token.NewFileSet(), "runtimes_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	families := map[string]string{"sanitized": "Sanitized", "release": "Release", "counted": "Counted", "wasm": "WASI"}
	products := map[string][]int{}
	for _, option := range runtimeOptions {
		products[option.name] = make([]int, runtimeProductGroups)
	}
	for _, declaration := range source.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestProduct_NativeRuntimes") {
			continue
		}
		calls := 0
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			callee, ok := call.Fun.(*ast.Ident)
			if !ok || callee.Name != "runtimeProducts" {
				return true
			}
			calls++
			if len(call.Args) != 3 {
				t.Errorf("%s has %d runtimeProducts arguments", function.Name.Name, len(call.Args))
				return true
			}
			name, nameOK := call.Args[1].(*ast.BasicLit)
			shard, shardOK := call.Args[2].(*ast.BasicLit)
			if !nameOK || !shardOK {
				t.Errorf("%s must name its option and shard explicitly", function.Name.Name)
				return true
			}
			option, nameErr := strconv.Unquote(name.Value)
			group, shardErr := strconv.Atoi(shard.Value)
			groups, declared := products[option]
			if nameErr != nil || shardErr != nil || !declared || group < 0 || group >= len(groups) {
				t.Errorf("%s names an undeclared option or shard: %s, %s", function.Name.Name, name.Value, shard.Value)
				return true
			}
			want := fmt.Sprintf("TestProduct_NativeRuntimes%s%d", families[option], group)
			if function.Name.Name != want {
				t.Errorf("%s builds %s", function.Name.Name, want)
			}
			groups[group]++
			return true
		})
		if calls != 1 {
			t.Errorf("%s has %d archive shard recipes, want one", function.Name.Name, calls)
		}
	}
	for option, groups := range products {
		seen := map[string]int{}
		for group, count := range groups {
			if count != 1 {
				t.Errorf("%s shard %d has %d products, want one", option, group, count)
			}
			sets := runtimeProductSets(group)
			if len(sets) == 0 || len(sets) > 2 {
				t.Errorf("%s shard %d has %d archives, want one or two", option, group, len(sets))
			}
			for _, set := range sets {
				seen[strings.Join(set.features, "+")] += count
			}
		}
		for _, set := range runtimeSets {
			features := strings.Join(set.features, "+")
			if seen[features] != 1 {
				t.Errorf("%s archives for [%s] built %d times, want once", option, features, seen[features])
			}
			delete(seen, features)
		}
		if len(seen) != 0 {
			t.Errorf("%s shards build undeclared archives: %v", option, seen)
		}
	}
}
