package lint

import (
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Products are built once per content address, across both leaves and processes.
// Lowering emits C and JS together so native and JS use the same mutated program.
func jsxTextnodesProducts(t *testing.T) (string, string, string) {
	t.Helper()
	d := jsxTextnodesDescriptor(t)
	var change struct{ Name, File, From, To string }
	data, err := os.ReadFile(filepath.Join("rules", d.Slug, "mutant.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &change); err != nil {
		t.Fatal(err)
	}
	if change.File == "" {
		change.File = d.Module
	}
	files := []string{"go.mod", "cohere", "internal", "stage1/typescript", "oracle", "stage1/cohere/lint/rules", "stage1/cohere/lint/mutants_jsx_textnodes_build_test.go"}
	for _, file := range portFiles(t) {
		files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", file)))
	}
	inputs := buildcache.Inputs{Name: "jsx-textnodes-mutant-lowered", Files: files,
		Flags: []string{change.From, change.To, change.File, packageDirectory}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	lowered := buildcache.Product(t, inputs, func(out string) error {
		port := copyPort(t, filepath.Join(out, "port"), change.From, change.To, filepath.Join("rules", d.Slug, change.File))
		prepareRegistry(t, port)
		started := time.Now()
		program, err := load.Load([]string{filepath.Join(port, "main.ts")})
		if err != nil {
			return err
		}
		t.Logf("load: %.3fs", time.Since(started).Seconds())
		started = time.Now()
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		t.Logf("lower: %.3fs", time.Since(started).Seconds())
		started = time.Now()
		if err := os.WriteFile(filepath.Join(out, "main.c"), []byte(native.C(lowered)), 0644); err != nil {
			return err
		}
		t.Logf("emit C: %.3fs", time.Since(started).Seconds())
		started = time.Now()
		err = os.WriteFile(filepath.Join(out, "lint.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
		t.Logf("emit JS: %.3fs", time.Since(started).Seconds())
		return err
	})
	inputs.Name = "jsx-textnodes-mutant-native"
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 4})...)
	inputs.Flags = append(inputs.Flags, "Split=true", "Jobs=4", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	built := buildcache.Product(t, inputs, func(out string) error {
		source, err := os.ReadFile(filepath.Join(lowered, "main.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(out, "scanner"), native.Options{Sanitize: true, Split: true, Jobs: 4})
	})
	return filepath.Join(lowered, "port"), filepath.Join(lowered, "lint.mjs"), filepath.Join(built, "scanner")
}
