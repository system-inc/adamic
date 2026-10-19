package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
)

// Oracle owns these shared archives. The product units find them ahead;
// runner shards only read them, lower their program and link it.
func regExpRunnerRuntime(t testing.TB, options native.Options) (string, native.RuntimeRequest) {
	t.Helper()
	request, err := native.RuntimeRequestFor(options)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtimeproducts.Declared(request); err != nil {
		t.Fatal(err)
	}
	directory, err := buildcache.Get(runtimeproducts.Inputs(request), func(string) error {
		return fmt.Errorf("RegExp runner runtime not built ahead for options %+v; run TestProduct_RegExpRunner tests first", options)
	})
	if err != nil {
		t.Fatal(err)
	}
	return directory, request
}

func TestProduct_RegExpRunnerNative(t *testing.T) {
	t.Parallel()
	runtimeproducts.Find(t, "sanitized", 0, "internal/oracle TestProduct_NativeRuntimesSanitized0")
}
func TestProduct_RegExpRunnerWASI(t *testing.T) {
	t.Parallel()
	runtimeproducts.Find(t, "wasm", 0, "internal/oracle TestProduct_NativeRuntimesWASI0")
}

func regExpCompilerEngine(t testing.TB, corpus string) *engine {
	t.Helper()
	tools, err := resolveTools()
	if err != nil {
		t.Fatal(err)
	}
	worker := &compilerWorker{}
	t.Cleanup(worker.close)
	return &engine{test262: corpus, work: t.TempDir(), adapt: true, jobs: 1,
		tools: tools, compiler: worker, log: io.Discard}
}

func regExpRunnerEngine(t testing.TB, corpus string) *engine {
	t.Helper()
	e := regExpCompilerEngine(t, corpus)
	directory, _ := regExpRunnerRuntime(t, native.Options{Sanitize: true})
	e.runtimeLibrary = func(source string, _ native.Options) (string, error) {
		if features := native.Features(source); len(features) != 0 || strings.Contains(source, "#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n") {
			return "", fmt.Errorf("undeclared RegExp runner runtime features: %v", features)
		}
		return filepath.Join(directory, "runtime.a"), nil
	}
	e.include = directory
	return e
}

func regExpFixture(t testing.TB, corpus, name string) classified {
	t.Helper()
	path := "built-ins/RegExp/" + name + ".js"
	text, err := os.ReadFile(filepath.Join(corpus, "test", path))
	if err != nil {
		t.Fatal(err)
	}
	return classify(path, string(text), true)
}
