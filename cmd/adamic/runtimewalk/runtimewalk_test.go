package runtimewalk

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Census the shared recipe's callers across the source tree, rather than assuming
// oracle is still its only producer. Build and Supply can build under a trace;
// Find cannot. A new producer or a private copy of the identity fails here until
// Builders names it. Mutant: drop internal/oracle from Builders.
func TestBuildersAreWhatTheWalkFinds(t *testing.T) {
	t.Parallel()
	// Reads: *.go
	packages := runtimeBuilderPackagesIn(t, filepath.Join("..", "..", ".."))
	if !slices.Equal(packages, Builders) {
		t.Fatalf("the packages that call runtimeproducts.Build or Supply are %v; Builders says %v", packages, Builders)
	}
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
