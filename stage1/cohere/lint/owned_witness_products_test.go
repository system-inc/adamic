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

func ownedWitnessInputs(t *testing.T) ([]string, []string) {
	t.Helper()
	files := []string{"go.mod", "cohere/go.mod", "cohere/go.sum", "stage1/cohere/lint/owned_witness_products_test.go"}
	for _, root := range []string{"internal", "stage1/typescript", "cohere/rule_runner", "cohere/schema", "cohere/internal", "cohere/TypeScript-shim", "cohere/TypeScript/tsc"} {
		err := filepath.WalkDir(filepath.Join(repository, root), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := filepath.Ext(path)
			if (ext == ".go" && !strings.HasSuffix(path, "_test.go")) || ext == ".c" || ext == ".h" || ext == ".ts" || ext == ".a" || entry.Name() == "go.mod" || entry.Name() == "go.sum" {
				relative, err := filepath.Rel(repository, path)
				if err != nil {
					return err
				}
				files = append(files, filepath.ToSlash(relative))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range portFiles(t) {
		files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", file)))
	}
	tools := []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}
	return files, tools
}

var ownedWitnessLoweredOnce, ownedWitnessNativeOnce sync.Once
var ownedWitnessLoweredPath, ownedWitnessNativePath string

func ownedWitnessLowered(t *testing.T, directory string) string {
	t.Helper()
	ownedWitnessLoweredOnce.Do(func() { ownedWitnessLoweredPath = buildOwnedWitnessLowered(t, directory) })
	if ownedWitnessLoweredPath == "" {
		t.Fatal("owned witness lowered product failed earlier in this process")
	}
	return ownedWitnessLoweredPath
}
func ownedWitnessNative(t *testing.T, directory string) string {
	t.Helper()
	ownedWitnessNativeOnce.Do(func() { ownedWitnessNativePath = buildOwnedWitnessNative(t, directory) })
	if ownedWitnessNativePath == "" {
		t.Fatal("owned witness native product failed earlier in this process")
	}
	return ownedWitnessNativePath
}

func buildOwnedWitnessLowered(t *testing.T, directory string) string {
	t.Helper()
	files, tools := ownedWitnessInputs(t)
	lowered := buildcache.Product(t, buildcache.Inputs{Name: "lint-owned-witnesses-lowered", Files: files, Toolchain: tools}, func(dir string) error {
		started := time.Now()
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(ir)), 0644); err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(ir)), 0644)
		t.Logf("cold build lowered=%s", time.Since(started))
		return err
	})
	return lowered
}

func buildOwnedWitnessNative(t *testing.T, directory string) string {
	t.Helper()
	files, tools := ownedWitnessInputs(t)
	lowered := ownedWitnessLowered(t, directory)
	flags := append(native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 4}), "Split=true", "Jobs=4", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED="+os.Getenv("ADAMIC_GATE_UNCACHED"))
	tools = append(tools, buildcache.Tool("clang", "--version"))
	binary := buildcache.Product(t, buildcache.Inputs{Name: "lint-owned-witnesses-sanitized", Files: files, Flags: flags, Toolchain: tools}, func(dir string) error {
		source, err := os.ReadFile(filepath.Join(lowered, "program.c"))
		if err != nil {
			return err
		}
		started := time.Now()
		err = native.Build(string(source), filepath.Join(dir, "native"), native.Options{Sanitize: true, Split: true, Jobs: 4})
		t.Logf("cold build sanitized=%s", time.Since(started))
		return err
	})
	return filepath.Join(binary, "native")
}

var ownedWitnessProductsOnce sync.Once
var ownedWitnessBinary, ownedWitnessModule string

func ownedWitnessProducts(t *testing.T, directory string) (string, string) {
	t.Helper()
	ownedWitnessProductsOnce.Do(func() {
		ownedWitnessModule = filepath.Join(ownedWitnessLowered(t, directory), "program.mjs")
		ownedWitnessBinary = ownedWitnessNative(t, directory)
	})
	if ownedWitnessBinary == "" {
		t.Fatal("owned witness products failed earlier in this process")
	}
	return ownedWitnessBinary, ownedWitnessModule
}
func TestProduct_LintOwnedWitnessesLowered(t *testing.T) {
	t.Parallel()
	ownedWitnessLowered(t, packageDirectory)
}
func TestProduct_LintOwnedWitnessesSanitized(t *testing.T) {
	t.Parallel()
	ownedWitnessNative(t, packageDirectory)
}
