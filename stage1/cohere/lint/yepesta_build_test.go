package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Products are built once per content address, across both leaves and processes.
// Lowering emits C and JS together so native and JS use the same mutated program.
//
// The product names nothing outside itself: the mutated port sits at tree/stage1/cohere/lint with its imports as
// written, beside every file outside the package they reach, so the same product reads the same from Workshop's cache
// and a runner's, wherever either checkout is. Its key names the files that change it and never this checkout's path.
func jsxTextnodesLowered(test *testing.T) (string, buildcache.Inputs) {
	test.Helper()
	inputs, build := jsxTextnodesLoweredRecipe(test)
	return buildcache.Product(test, inputs, build), inputs
}

// jsxTextnodesLoweredRecipe is the lowered mutant's key and its build, apart, so a reader of the native can key the
// native without building the lowered port it's made from.
func jsxTextnodesLoweredRecipe(test *testing.T) (buildcache.Inputs, func(out string) error) {
	test.Helper()
	d := jsxTextnodesDescriptor(test)
	var change struct{ Name, File, From, To string }
	data, err := os.ReadFile(filepath.Join("rules", d.Slug, "mutant.json"))
	if err != nil {
		test.Fatal(err)
	}
	if err := json.Unmarshal(data, &change); err != nil {
		test.Fatal(err)
	}
	if change.File == "" {
		change.File = d.Module
	}
	files := []string{
		"go.mod",
		"cohere",
		"internal",
		"stage1/typescript",
		"oracle",
		"stage1/cohere/lint/rules",
		"stage1/cohere/lint/yepesta_build_test.go",
		"stage1/cohere/lint/lint_test.go",
	}
	for _, file := range portFiles(test) {
		files = append(files, filepath.ToSlash(filepath.Join(jsxTextnodesPackage, file)))
	}
	inputs := buildcache.Inputs{
		Name:      "jsx-textnodes-mutant-lowered",
		Files:     files,
		Flags:     []string{change.From, change.To, change.File},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
	return inputs, func(out string) error {
		port, err := portTree(
			test,
			filepath.Join(out, "tree"),
			change.From,
			change.To,
			filepath.Join("rules", d.Slug, change.File),
		)
		if err != nil {
			return err
		}
		prepareRegistry(test, port)
		started := time.Now()
		program, err := load.Load([]string{filepath.Join(port, "main.ts")})
		if err != nil {
			return err
		}
		test.Logf("load: %.3fs", time.Since(started).Seconds())
		started = time.Now()
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		test.Logf("lower: %.3fs", time.Since(started).Seconds())
		started = time.Now()
		if err := os.WriteFile(filepath.Join(out, "main.c"), []byte(native.C(lowered)), 0644); err != nil {
			return err
		}
		test.Logf("emit C: %.3fs", time.Since(started).Seconds())
		started = time.Now()
		err = os.WriteFile(filepath.Join(out, "lint.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
		test.Logf("emit JS: %.3fs", time.Since(started).Seconds())
		if err != nil {
			return err
		}
		// The port's imports name this checkout for the build's own lowering; the product names it <repository>
		// (#tqrqx60).
		return buildcache.RelativeFiles(out)
	}
}

// jsxTextnodesPackage is this package's path in the repository, where portTree places the copied port.
const jsxTextnodesPackage = "stage1/cohere/lint"

// treeImport is a module specifier a TypeScript file imports or re-exports from, on a line of its own.
var treeImport = regexp.MustCompile(`(?m)^\s*(?:(?:import|export)\b[^;'"]*?\bfrom\s+|import\s+)['"]([^'"]+)['"]`)

func keepPortImports(test *testing.T, file, source string) string { return source }

// portTree copies the port, mutated as copyPort mutates it, to tree/stage1/cohere/lint with its imports as written,
// and every file outside the package those imports reach (stage1/typescript's parser and scanner, today) to its own
// repository path under tree, so the copy resolves within itself. It returns the copied package's directory.
func portTree(test *testing.T, tree, from, to string, targets ...string) (string, error) {
	test.Helper()
	port := copyPortImporting(
		test,
		filepath.Join(tree, filepath.FromSlash(jsxTextnodesPackage)),
		keepPortImports,
		from,
		to,
		targets...,
	)
	root, err := filepath.Abs(repository)
	if err != nil {
		return "", err
	}
	var queue []string
	reach := func(file, source string) error {
		for _, match := range treeImport.FindAllStringSubmatch(source, -1) {
			if !strings.HasPrefix(match[1], ".") {
				continue
			}
			target := path.Clean(path.Join(path.Dir(file), match[1]))
			switch {
			case target == ".." || strings.HasPrefix(target, "../"):
				return fmt.Errorf("%s imports %s, outside the repository", file, match[1])
			case !strings.HasPrefix(target, jsxTextnodesPackage+"/"):
				queue = append(queue, target)
			}
		}
		return nil
	}
	for _, file := range portFiles(test) {
		if !strings.HasSuffix(file, ".ts") && !strings.HasSuffix(file, ".a") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(port, file))
		if err != nil {
			return "", err
		}
		if err := reach(path.Join(jsxTextnodesPackage, filepath.ToSlash(file)), string(data)); err != nil {
			return "", err
		}
	}
	copied := map[string]bool{}
	for len(queue) > 0 {
		file := queue[0]
		queue = queue[1:]
		if copied[file] {
			continue
		}
		copied[file] = true
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			return "", err
		}
		destination := filepath.Join(tree, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(destination, data, 0644); err != nil {
			return "", err
		}
		if err := reach(file, string(data)); err != nil {
			return "", err
		}
	}
	return port, nil
}

func jsxTextnodesNative(test *testing.T) string {
	test.Helper()
	lowered, inputs := jsxTextnodesLowered(test)
	return jsxTextnodesNativeFrom(test, lowered, inputs)
}

func jsxTextnodesNativeFrom(test *testing.T, lowered string, inputs buildcache.Inputs) string {
	test.Helper()
	built := buildcache.Product(test, jsxTextnodesNativeInputs(inputs), func(out string) error {
		return jsxTextnodesNativeBuild(lowered, out)
	})
	return filepath.Join(built, "scanner")
}

// jsxTextnodesNativeInputs is the sanitized native's key: the lowered mutant's, whose flags and files carry the
// mutation (mutant.json's from, to and file, and the rules directory that holds it), with the native build's own
// options.
func jsxTextnodesNativeInputs(inputs buildcache.Inputs) buildcache.Inputs {
	inputs.Name = "jsx-textnodes-mutant-native"
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 4})...)
	inputs.Flags = append(inputs.Flags, "Split=true", "Jobs=4", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	return inputs
}

// jsxTextnodesNativeBuild builds the sanitized native, the executable scanner, into out from the lowered mutant's C.
func jsxTextnodesNativeBuild(lowered, out string) error {
	source, err := os.ReadFile(filepath.Join(lowered, "main.c"))
	if err != nil {
		return err
	}
	return native.Build(
		string(source),
		filepath.Join(out, "scanner"),
		native.Options{Sanitize: true, Split: true, Jobs: 4},
	)
}

// jsxTextnodesNativeOwner is the one product test that builds the sanitized native. Every other product test that reads
// it finds it by name, after this owner in cloud/fast-gate/product-order.json, so the tree settles it once, under one
// key.
const jsxTextnodesNativeOwner = "stage1/cohere/lint TestProduct_JsxTextnodesMutantNative"

// jsxTextnodesFoundNative is the sanitized native's scanner for every reader but its owner: the native mutant canary
// reads it as well, so the same mutant isn't built a second time from a port of its own. Under the trace it is only
// found, and the lowered port is never built here; outside one it builds both, as the owner would.
func jsxTextnodesFoundNative(test *testing.T) string {
	test.Helper()
	inputs, _ := jsxTextnodesLoweredRecipe(test)
	built := buildcache.BuiltElsewhere(test, "the JSX textnode mutant's sanitized native", jsxTextnodesNativeInputs(
		inputs,
	), jsxTextnodesNativeOwner, func(out string) error {
		lowered, _ := jsxTextnodesLowered(test)
		return jsxTextnodesNativeBuild(lowered, out)
	})
	return filepath.Join(built, "scanner")
}

// jsxTextnodesBuilt holds the products' places once per process: every shard in one process reads the same products,
// and their keys hash the repository's largest trees, so each is looked up once.
var jsxTextnodesBuilt struct {
	once              sync.Once
	port, js, scanner string
}

// jsxTextnodesProducts are the mutated port, its emitted JavaScript and its sanitized native scanner.
func jsxTextnodesProducts(test *testing.T) (string, string, string) {
	test.Helper()
	jsxTextnodesBuilt.once.Do(func() {
		lowered, inputs := jsxTextnodesLowered(test)
		scanner := jsxTextnodesNativeFrom(test, lowered, inputs)
		jsxTextnodesBuilt.port = filepath.Join(lowered, "tree", filepath.FromSlash(jsxTextnodesPackage))
		jsxTextnodesBuilt.js = filepath.Join(lowered, "lint.mjs")
		jsxTextnodesBuilt.scanner = scanner
	})
	if jsxTextnodesBuilt.scanner == "" {
		test.Fatal("the JSX textnode mutant's products failed earlier in this process")
	}
	return jsxTextnodesBuilt.port, jsxTextnodesBuilt.js, jsxTextnodesBuilt.scanner
}
