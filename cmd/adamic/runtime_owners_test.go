package main

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/cmd/adamic/treewalk"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
)

// There is one builder package per native-runtime name key this package reads.
// Oracle's count products call Supply through identity(), so they belong to the
// same owner package as NativeRuntimes. The packages that build runtimeproducts'
// archives are treewalk.RuntimeBuilders, which treewalk's own test holds to the
// source tree, so this package's key never reads every Go file (#hbq28nr).
// Mutant: clear sanitized's owner, making commandRuntimeProduct build the
// already-owned sanitized key a second time.
func TestCommandRuntimeArchivesHaveOneBuilder(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, product := range commandRuntimes {
		if product.options.Target != "" && os.Getenv("WASI_SYSROOT") == "" {
			continue
		}
		request, runtimeRequestForError := native.RuntimeRequestFor(product.options, product.features...)
		if runtimeRequestForError != nil {
			test.Fatal(runtimeRequestForError)
		}
		key := buildcache.NameKey(root, commandRuntimeInputs(request))
		var builders []string
		if productsBuildKey(test, root, key, product.features) {
			builders = append(builders, treewalk.RuntimeBuilders...)
		}
		if product.owner == "" {
			builders = append(builders, "cmd/adamic")
		}
		if len(builders) != 1 {
			test.Errorf(
				"%s native-runtime name key %s has %d builder packages %v, want exactly one",
				product.name,
				key,
				len(builders),
				builders,
			)
			continue
		}
		if product.owner != "" && strings.Fields(product.owner)[0] != builders[0] {
			test.Errorf("%s finds %s, but its builder is %s", product.name, product.owner, builders[0])
		}
		test.Logf("%s native-runtime name key %s builder %s", product.name, key, builders[0])
	}
}

// productsBuildKey says whether runtimeproducts builds the archive with this name key ahead, under any variant of
// the feature set. It compares keys, not options: commandRuntime drops Request, so the wasm-request variant over
// the empty set builds the key of cmd/adamic's wasm-counted row although their options differ, and comparing
// options let both packages build it (#eh4ym89). Mutant: clear wasm-counted's owner, and this finds two builders.
func productsBuildKey(test *testing.T, root, key string, features []string) bool {
	test.Helper()
	if !slices.ContainsFunc(
		runtimeproducts.Sets,
		func(set runtimeproducts.Set) bool { return slices.Equal(set.Features, features) },
	) {
		return false
	}
	for _, variant := range runtimeproducts.Variants {
		if variant.Options.Target != "" && os.Getenv("WASI_SYSROOT") == "" {
			continue
		}
		request, runtimeRequestForError := native.RuntimeRequestFor(variant.Options, features...)
		if runtimeRequestForError != nil {
			test.Fatal(runtimeRequestForError)
		}
		if buildcache.NameKey(root, runtimeproducts.Inputs(request)) == key {
			return true
		}
	}
	return false
}

// A shared product must already exist when its finder starts. Mutant: remove the
// sanitized finder's wait; BuiltElsewhere's missing-owner failure is also exercised
// under the real tracer, with an empty cache, in the unit's review evidence.
func TestCommandRuntimeProductsWaitForOwners(t *testing.T) {
	t.Parallel()
	contents, readError := os.ReadFile("../../cloud/fast-gate/product-order.json")
	if readError != nil {
		t.Fatal(readError)
	}
	var order struct {
		Waits map[string][]string `json:"waits"`
	}
	if decodeError := json.Unmarshal(contents, &order); decodeError != nil {
		t.Fatal(decodeError)
	}
	tests := map[string]string{}
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "runtimes_test.go", nil, 0)
	if parseFileError != nil {
		t.Fatal(parseFileError)
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestProduct_") {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "commandRuntimeProduct" || len(call.Args) != 2 {
				return true
			}
			argument, ok := call.Args[1].(*ast.BasicLit)
			if !ok {
				t.Fatal("runtime product name must be a literal")
			}
			product, unquoteError := strconv.Unquote(argument.Value)
			if unquoteError != nil {
				t.Fatal(unquoteError)
			}
			tests[product] = "cmd/adamic " + function.Name.Name
			return true
		})
	}
	for _, product := range commandRuntimes {
		test, exists := tests[product.name]
		if !exists {
			t.Errorf("no product test for %s", product.name)
			continue
		}
		if product.owner != "" && !slices.Contains(order.Waits[test], product.owner) {
			t.Errorf("%s must wait for %s in product-order.json", test, product.owner)
		}
	}
}

// Exercise the product itself with the tree builder's trace protocol and no owner
// product. It must fail by name, before compiling an archive. Mutant: replace
// BuiltElsewhere with Product, adding a second builder of Loom's a33c0b598f8a key.
func TestSharedCommandRuntimeCannotBuildItsOwnersArchive(test *testing.T) {
	test.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(
		ctx,
		os.Args[0],
		"-test.run=^TestProduct_CommandRuntimeSanitized$",
		"-test.timeout=40s",
		"-test.v",
	)
	command.Env = append(
		os.Environ(),
		"ADAMIC_BUILD_CACHE_DIR="+test.TempDir(),
		"ADAMIC_BUILD_TRACE="+test.TempDir(),
		"ADAMIC_BUILD_CACHE=",
		"ADAMIC_BUILD_STORE=off",
	)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError == nil ||
		!strings.Contains(string(output), "internal/oracle TestProduct_NativeRuntimesSanitized0's to build") ||
		!strings.Contains(string(output), "is this test in the tree's product order after it?") {
		test.Fatalf(
			"a shared runtime product must fail for its missing owner, never build: %v\n%s",
			combinedOutputError,
			output,
		)
	}
}
