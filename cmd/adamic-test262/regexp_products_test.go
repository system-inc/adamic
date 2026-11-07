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
func regExpRunnerRuntime(test testing.TB, options native.Options) (string, native.RuntimeRequest) {
	test.Helper()
	request, err := native.RuntimeRequestFor(options)
	if err != nil {
		test.Fatal(err)
	}
	if err := runtimeproducts.Declared(request); err != nil {
		test.Fatal(err)
	}
	directory, err := buildcache.BuiltAhead(runtimeproducts.Inputs(request), fmt.Errorf(
		"RegExp runner runtime not built ahead for options %+v; run TestProduct_RegExpRunner tests first",
		options,
	))
	if err != nil {
		test.Fatal(err)
	}
	return directory, request
}

func TestProduct_RegExpRunnerNative(test *testing.T) {
	test.Parallel()
	runtimeproducts.Find(test, "sanitized", 0, "internal/oracle TestProduct_NativeRuntimesSanitized0")
}
func TestProduct_RegExpRunnerWASI(test *testing.T) {
	test.Parallel()
	runtimeproducts.Find(test, "wasm", 0, "internal/oracle TestProduct_NativeRuntimesWASI0")
}

func regExpCompilerEngine(test testing.TB, corpus string) *engine {
	test.Helper()
	tools, err := resolveTools()
	if err != nil {
		test.Fatal(err)
	}
	worker := &compilerWorker{}
	test.Cleanup(worker.close)
	return &engine{test262: corpus, work: test.TempDir(), adapt: true, jobs: 1,
		tools: tools, compiler: worker, log: io.Discard}
}

func regExpRunnerEngine(test testing.TB, corpus string) *engine {
	test.Helper()
	e := regExpCompilerEngine(test, corpus)
	directory, _ := regExpRunnerRuntime(test, native.Options{Sanitize: true})
	e.runtimeLibrary = func(source string, _ native.Options) (string, error) {
		if features := native.Features(
			source,
		); len(features) != 0 || strings.Contains(source, "#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n") {
			return "", fmt.Errorf("undeclared RegExp runner runtime features: %v", features)
		}
		return filepath.Join(directory, "runtime.a"), nil
	}
	e.include = directory
	return e
}

func regExpFixture(test testing.TB, corpus, name string) classified {
	test.Helper()
	path := "built-ins/RegExp/" + name + ".js"
	text, err := os.ReadFile(filepath.Join(corpus, "test", path))
	if err != nil {
		test.Fatal(err)
	}
	return classify(path, string(text), true)
}
