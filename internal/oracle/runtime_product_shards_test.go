package oracle

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// Inspect the actual selectable products, rather than a second registry that could drift.
// Mutants: an omitted product, a product calling the wrong shard, a dropped feature set,
// and complete archives combined into products that exceed the two-archive bound.
func TestRuntimeProductShardsCoverEveryArchive(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	var declarations []ast.Decl
	for _, path := range paths {
		source, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		declarations = append(declarations, source.Decls...)
	}
	families := map[string]string{
		"sanitized": "NativeRuntimesSanitized", "release": "NativeRuntimesRelease",
		"counted": "NativeRuntimesCounted", "wasm": "NativeRuntimesWASI",
		"parallel_tsan": "ParallelTSanRuntime",
	}
	products := map[string][]int{}
	for _, option := range runtimeOptions {
		if families[option.name] == "" {
			t.Errorf("option %s has no declared runtime product family", option.name)
		}
		products[option.name] = make([]int, runtimeProductGroupCount(option.name))
	}
	for _, declaration := range declarations {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestProduct_") {
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
			// These products skip before requesting an option on platforms without TSan.
			if !declared && option == "parallel_tsan" && runtime.GOOS != "linux" {
				return true
			}
			if nameErr != nil || shardErr != nil || !declared || group < 0 || group >= len(groups) {
				t.Errorf("%s names an undeclared option or shard: %s, %s", function.Name.Name, name.Value, shard.Value)
				return true
			}
			want := fmt.Sprintf("TestProduct_%s%d", families[option], group)
			if function.Name.Name != want {
				t.Errorf("%s builds %s", function.Name.Name, want)
			}
			groups[group]++
			return true
		})
		runtimeFamily := strings.HasPrefix(function.Name.Name, "TestProduct_NativeRuntimes") ||
			strings.HasPrefix(function.Name.Name, "TestProduct_ParallelTSanRuntime")
		if (calls != 0 || runtimeFamily) && calls != 1 {
			t.Errorf("%s has %d archive shard recipes, want one", function.Name.Name, calls)
		}
	}
	for option, groups := range products {
		seen := map[string]int{}
		for group, count := range groups {
			if count != 1 {
				t.Errorf("%s shard %d has %d products, want one", option, group, count)
			}
			sets := runtimeProductSets(group, len(groups))
			maximum := 2
			if option == "parallel_tsan" {
				maximum = 4
			}
			if len(sets) == 0 || len(sets) > maximum {
				t.Errorf("%s shard %d has %d archives, want between one and %d", option, group, len(sets), maximum)
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
