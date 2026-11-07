package lint

import (
	"context"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func ownedWitnessInputs(test *testing.T) ([]string, []string) {
	test.Helper()
	files := []string{"go.mod", "cohere/go.mod", "cohere/go.sum", "stage1/cohere/lint/owned_witness_products_test.go"}
	for _, root := range []string{
		"internal",
		"stage1/typescript",
		"cohere/rule_runner",
		"cohere/schema",
		"cohere/internal",
		"cohere/TypeScript-shim",
		"cohere/TypeScript/tsc",
	} {
		// Reads: cohere
		// Reads: stage1/*.a
		// Reads: stage1/*.ts
		// Reads: stage1/typescript/parser/grammar.ts
		// Reads: stage1/typescript/parser/jsx.ts
		// Reads: stage1/typescript/parser/lookahead.ts
		// Reads: stage1/typescript/parser/nodes.ts
		// Reads: stage1/typescript/parser/parser.ts
		// Reads: stage1/typescript/parser/statements.ts
		// Reads: stage1/typescript/parser/testdata/oracle.go
		// Reads: stage1/typescript/scanner/characters.ts
		// Reads: stage1/typescript/scanner/scanner.ts
		// Reads: stage1/typescript/scanner/tokens.ts
		// Reads: internal/load/prelude.d.ts
		// Reads: internal/native/runtime/*.[ch]
		walkError := filepath.WalkDir(filepath.Join(
			repository,
			root,
		), func(path string, entry fs.DirEntry, pathError error) error {
			if pathError != nil {
				return pathError
			}
			if entry.IsDir() {
				if entry.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := filepath.Ext(path)
			if (ext == ".go" && !strings.HasSuffix(path, "_test.go")) ||
				ext == ".c" ||
				ext == ".h" ||
				ext == ".ts" ||
				ext == ".a" ||
				entry.Name() == "go.mod" ||
				entry.Name() == "go.sum" {
				relative, relativePathError := filepath.Rel(repository, path)
				if relativePathError != nil {
					return relativePathError
				}
				files = append(files, filepath.ToSlash(relative))
			}
			return nil
		})
		if walkError != nil {
			test.Fatal(walkError)
		}
	}
	for _, file := range portFiles(test) {
		files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", file)))
	}
	tools := []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}
	return files, tools
}

var ownedWitnessLoweredOnce, ownedWitnessNativeOnce sync.Once
var ownedWitnessLoweredPath, ownedWitnessNativePath string

func ownedWitnessLowered(test *testing.T, directory string) string {
	test.Helper()
	ownedWitnessLoweredOnce.Do(func() { ownedWitnessLoweredPath = buildOwnedWitnessLowered(test, directory) })
	if ownedWitnessLoweredPath == "" {
		test.Fatal("owned witness lowered product failed earlier in this process")
	}
	return ownedWitnessLoweredPath
}
func ownedWitnessNative(test *testing.T, directory string) string {
	test.Helper()
	ownedWitnessNativeOnce.Do(func() { ownedWitnessNativePath = buildOwnedWitnessNative(test, directory) })
	if ownedWitnessNativePath == "" {
		test.Fatal("owned witness native product failed earlier in this process")
	}
	return ownedWitnessNativePath
}

func buildOwnedWitnessLowered(test *testing.T, directory string) string {
	test.Helper()
	files, tools := ownedWitnessInputs(test)
	lowered := buildcache.Product(test, buildcache.Inputs{
		Name:      "lint-owned-witnesses-lowered",
		Files:     files,
		Toolchain: tools,
	}, func(dir string) error {
		started := time.Now()
		program, loadError := load.Load([]string{filepath.Join(directory, "main.ts")})
		if loadError != nil {
			return loadError
		}
		ir, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		if writeError := os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(ir)), 0644); writeError != nil {
			return writeError
		}
		writeError := os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(ir)), 0644)
		test.Logf("cold build lowered=%s", time.Since(started))
		return writeError
	})
	return lowered
}

func buildOwnedWitnessNative(test *testing.T, directory string) string {
	test.Helper()
	files, tools := ownedWitnessInputs(test)
	lowered := ownedWitnessLowered(test, directory)
	flags := append(
		native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 4}),
		"Split=true",
		"Jobs=4",
		"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
	)
	tools = append(tools, buildcache.Tool("clang", "--version"))
	binary := buildcache.Product(test, buildcache.Inputs{
		Name:      "lint-owned-witnesses-sanitized",
		Files:     files,
		Flags:     flags,
		Toolchain: tools,
	}, func(dir string) error {
		source, readError := os.ReadFile(filepath.Join(lowered, "program.c"))
		if readError != nil {
			return readError
		}
		started := time.Now()
		buildError := native.Build(
			string(source),
			filepath.Join(dir, "native"),
			native.Options{Sanitize: true, Split: true, Jobs: 4},
		)
		test.Logf("cold build sanitized=%s", time.Since(started))
		return buildError
	})
	return filepath.Join(binary, "native")
}

var ownedWitnessProductsOnce sync.Once
var ownedWitnessBinary, ownedWitnessModule string

func ownedWitnessProducts(test *testing.T, directory string) (string, string) {
	test.Helper()
	ownedWitnessProductsOnce.Do(func() {
		ownedWitnessModule = filepath.Join(ownedWitnessLowered(test, directory), "program.mjs")
		ownedWitnessBinary = ownedWitnessNative(test, directory)
	})
	if ownedWitnessBinary == "" {
		test.Fatal("owned witness products failed earlier in this process")
	}
	return ownedWitnessBinary, ownedWitnessModule
}
func TestProduct_LintOwnedWitnessesLowered(test *testing.T) {
	test.Parallel()
	requireLintArtifacts(test, ownedWitnessLowered(test, packageDirectory), "program.c", "program.mjs")
}
func TestProduct_LintOwnedWitnessesSanitized(test *testing.T) {
	test.Parallel()
	requireLintExecutable(test, ownedWitnessNative(test, packageDirectory))
}
