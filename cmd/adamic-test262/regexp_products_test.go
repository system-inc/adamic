package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// Each runner shard fetches this recipe. The product unit owns compilation;
// consumers only lower their program, link it and compare its observations.
func regExpRunnerRuntime(t testing.TB, options native.Options) (string, native.RuntimeRequest) {
	t.Helper()
	request, err := native.RuntimeRequestFor(options)
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.Product(t, buildcache.Inputs{
		Name:      "native-runtime",
		Files:     []string{"internal/native/runtime", "internal/native/library.go", "internal/native/native.go", "internal/native/target.go"},
		Flags:     request.Portable,
		Toolchain: []string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version")},
	}, func(directory string) error { return request.Build(directory, 4) })
	return directory, request
}

func TestProduct_RegExpRunnerNative(t *testing.T) {
	t.Parallel()
	regExpRunnerRuntime(t, native.Options{Sanitize: true})
}
func TestProduct_RegExpRunnerWASI(t *testing.T) {
	t.Parallel()
	regExpRunnerRuntime(t, native.Options{Target: "wasm32-wasi"})
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
	e.runtime = native.RuntimeLinkFlags(filepath.Join(directory, "runtime.a"))
	e.include = directory
	e.flags = native.Flags(native.Options{Sanitize: true})
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
