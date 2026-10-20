package main

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
)

// Census the shared recipe's callers across the source tree, rather than assuming
// oracle is still its only producer. Build and Supply can build under a trace;
// Find cannot. A new producer or a private copy of the identity needs review here.
func runtimeBuilderPackages(t *testing.T) []string {
	t.Helper()
	// Reads: *.go
	return runtimeBuilderPackagesIn(t, filepath.Join("..", ".."))
}

// Loom assembles source without .git. Walk only this repository's sources, not
// fixtures, dependencies or ignored local output; never ask git for a listing.
func runtimeBuilderPackagesIn(t *testing.T, root string) []string {
	t.Helper()
	ignored, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	var directories []string
	for _, line := range strings.Split(string(ignored), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || !strings.HasSuffix(line, "/") {
			continue
		}
		// These are the root file's directory rules, not a second gitignore
		// implementation. Refuse new forms we cannot honor rather than miss a builder.
		if strings.HasPrefix(line, "!") || strings.Contains(line, "**") {
			t.Fatalf("unsupported census directory rule %q in .gitignore", line)
		}
		directories = append(directories, strings.TrimSuffix(line, "/"))
	}
	var submodules []string
	modules, err := os.ReadFile(filepath.Join(root, ".gitmodules"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(modules), "\n") {
		key, value, found := strings.Cut(line, "=")
		if found && strings.TrimSpace(key) == "path" {
			submodules = append(submodules, strings.TrimSpace(value))
		}
	}
	var files []string
	err = filepath.WalkDir(root, func(filename string, entry fs.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		name, err := filepath.Rel(root, filename)
		if err != nil {
			return err
		}
		name = filepath.ToSlash(name)
		if entry.IsDir() {
			if name == "." {
				return nil
			}
			switch entry.Name() {
			case "testdata", "vendor", "node_modules", ".git":
				return filepath.SkipDir
			}
			if slices.Contains(submodules, name) {
				return filepath.SkipDir
			}
			for _, rule := range directories {
				target := name
				if !strings.HasPrefix(rule, "/") && !strings.Contains(rule, "/") {
					target = entry.Name()
				}
				match, err := path.Match(strings.TrimPrefix(rule, "/"), target)
				if err != nil {
					return err
				}
				if match {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if entry.Type().IsRegular() && strings.HasSuffix(name, ".go") {
			files = append(files, name)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("census: %d Go files; skip testdata, vendor, node_modules, .git; ignored directories %v; submodules %v", len(files), directories, submodules)
	var packages []string
	for _, name := range files {
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		aliases := map[string]bool{}
		for _, imported := range file.Imports {
			path, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if path == "github.com/system-inc/adamic/internal/native/runtimeproducts" {
				alias := "runtimeproducts"
				if imported.Name != nil {
					alias = imported.Name.Name
				}
				if alias == "." {
					t.Fatalf("%s: a dot import hides runtime producers from the census", name)
				}
				aliases[alias] = true
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if field, ok := node.(*ast.KeyValueExpr); ok {
				key, named := field.Key.(*ast.Ident)
				value, literal := field.Value.(*ast.BasicLit)
				if named && literal && key.Name == "Name" && value.Value == `"native-runtime"` && name != "internal/native/runtimeproducts/runtimeproducts.go" {
					t.Errorf("%s declares a private native-runtime identity; use runtimeproducts.Inputs so ownership can be counted", name)
				}
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			qualifier, ok := selector.X.(*ast.Ident)
			if ok && aliases[qualifier.Name] && (selector.Sel.Name == "Build" || selector.Sel.Name == "Supply") {
				packageName := filepath.ToSlash(filepath.Dir(name))
				if !slices.Contains(packages, packageName) {
					packages = append(packages, packageName)
				}
			}
			return true
		})
	}
	slices.Sort(packages)
	return packages
}

// There is one builder package per native-runtime name key this package reads.
// Oracle's count products call Supply through identity(), so they belong to the
// same owner package as NativeRuntimes. Mutant: clear sanitized's owner, making
// commandRuntimeProduct build the already-owned sanitized key a second time.
func TestCommandRuntimeArchivesHaveOneBuilder(t *testing.T) {
	t.Parallel()
	packages := runtimeBuilderPackages(t)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, product := range commandRuntimes {
		if product.options.Target != "" && os.Getenv("WASI_SYSROOT") == "" {
			continue
		}
		request, err := native.RuntimeRequestFor(product.options, product.features...)
		if err != nil {
			t.Fatal(err)
		}
		key := buildcache.NameKey(root, commandRuntimeInputs(request))
		var builders []string
		if runtimeproducts.Declared(request) == nil {
			builders = append(builders, packages...)
		}
		if product.owner == "" {
			builders = append(builders, "cmd/adamic")
		}
		if len(builders) != 1 {
			t.Errorf("%s native-runtime name key %s has %d builder packages %v, want exactly one", product.name, key, len(builders), builders)
			continue
		}
		if product.owner != "" && strings.Fields(product.owner)[0] != builders[0] {
			t.Errorf("%s finds %s, but its builder is %s", product.name, product.owner, builders[0])
		}
		t.Logf("%s native-runtime name key %s builder %s", product.name, key, builders[0])
	}
}

// A shared product must already exist when its finder starts. Mutant: remove the
// sanitized finder's wait; BuiltElsewhere's missing-owner failure is also exercised
// under the real tracer, with an empty cache, in the unit's review evidence.
func TestCommandRuntimeProductsWaitForOwners(t *testing.T) {
	t.Parallel()
	contents, err := os.ReadFile("../../cloud/fast-gate/product-order.json")
	if err != nil {
		t.Fatal(err)
	}
	var order struct {
		Waits map[string][]string `json:"waits"`
	}
	if err := json.Unmarshal(contents, &order); err != nil {
		t.Fatal(err)
	}
	tests := map[string]string{}
	file, err := parser.ParseFile(token.NewFileSet(), "runtimes_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
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
			product, err := strconv.Unquote(argument.Value)
			if err != nil {
				t.Fatal(err)
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
func TestSharedCommandRuntimeCannotBuildItsOwnersArchive(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestProduct_CommandRuntimeSanitized$", "-test.timeout=40s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_BUILD_CACHE_DIR="+t.TempDir(), "ADAMIC_BUILD_TRACE="+t.TempDir(), "ADAMIC_BUILD_CACHE=", "ADAMIC_BUILD_STORE=off")
	output, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "internal/oracle TestProduct_NativeRuntimesSanitized0's to build") || !strings.Contains(string(output), "is this test in the tree's product order after it?") {
		t.Fatalf("a shared runtime product must fail for its missing owner, never build: %v\n%s", err, output)
	}
}

// Decoys must never be parsed, while ordinary sources and similarly named
// directories must remain visible. No Git checkout or executable is needed.
func TestCommandRuntimeCensusWalksSourceTree(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	files := map[string]string{
		".gitignore":  ".cache/\n.idea/\n/build/\nartifacts/cache/\nscratch-*/\n",
		".gitmodules": "[submodule \"cohere\"]\n path = cohere\n",
	}
	producer := "package producer\nimport rp \"github.com/system-inc/adamic/internal/native/runtimeproducts\"\nfunc produce() { rp.Build(nil, \"sanitized\", 0) }\n"
	want := []string{".cache-source", "nested/build", "ordinary"}
	for _, directory := range want {
		files[directory+"/producer.go"] = producer
	}
	files["reader/reader.go"] = "package reader\nimport rp \"github.com/system-inc/adamic/internal/native/runtimeproducts\"\nfunc read() { rp.Find(nil, \"sanitized\", 0, \"owner\") }\n"
	for _, directory := range []string{"testdata", "nested/testdata", "vendor", "node_modules", ".git", ".cache", "nested/.cache", ".idea", "build", "artifacts/cache", "scratch-one", "cohere"} {
		files[directory+"/decoy.go"] = "deliberately invalid Go, which must not be parsed"
	}
	for name, contents := range files {
		filename := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got := runtimeBuilderPackagesIn(t, root); !slices.Equal(got, want) {
		t.Fatalf("census builder packages %v, want %v", got, want)
	}
}
