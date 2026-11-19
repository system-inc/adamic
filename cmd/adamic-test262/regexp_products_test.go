package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// Oracle owns these shared archives. The product units find them ahead;
// runner shards only read them, lower their program and link it.
func regExpRunnerRuntime(t testing.TB, options native_runtime.Options) (string, native_runtime.RuntimeRequest) {
	t.Helper()
	request, err := native_runtime.RuntimeRequestFor(options)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtimeproducts.Declared(request); err != nil {
		t.Fatal(err)
	}
	directory, err := buildcache.BuiltAhead(runtimeproducts.Inputs(request), fmt.Errorf("RegExp runner runtime not built ahead for options %+v; run TestProduct_RegExpRunner tests first", options))
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
	directory, _ := regExpRunnerRuntime(t, native_runtime.Options{Sanitize: true})
	e.runtimeLibrary = func(source string, _ native_runtime.Options) (string, error) {
		if features := native_runtime.Features(source); len(features) != 0 || strings.Contains(source, "#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n") {
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
