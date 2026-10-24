package worker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// Products cache builds, never test verdicts. Behavior tests can only find them:
// even a cold runner must not compile a CLI or runtime inside its unit's clock.
func findWorkerProduct(t *testing.T, inputs buildcache.Inputs) string {
	t.Helper()
	directory, err := buildcache.Get(inputs, func(string) error {
		return fmt.Errorf("Worker product %s not built ahead; run TestProduct_Worker tests first", inputs.Name)
	})
	if err != nil {
		t.Fatal(err)
	}
	return directory
}

func activeMutation(t *testing.T) *workerMutation {
	t.Helper()
	name := os.Getenv("ADAMIC_WORKER_TEST_MUTATION")
	if name == "" {
		return nil
	}
	for _, group := range [][]workerMutation{workerMutants, wasmWorkerMutants} {
		for _, mutant := range group {
			if mutant.name == name {
				return &mutant
			}
		}
	}
	t.Fatalf("unknown Worker mutation %q", name)
	return nil
}

func workerCompilerProduct(t *testing.T, mutant *workerMutation, build bool) string {
	t.Helper()
	if build {
		if mutant == nil {
			return buildcache.AdamicBuiltElsewhere(t)
		}
		return buildcache.GoBuildInMutated(t, "", "adamic", "./cmd/adamic", nil, mutationOf(mutant))
	}
	inputs, err := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, mutationOf(mutant))
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(findWorkerProduct(t, inputs), "adamic")
}

func workerTestProduct(t *testing.T, mutant *workerMutation, build bool) string {
	t.Helper()
	if build {
		return buildcache.GoTestMutated(t, "", "worker.test", "./internal/worker", nil, mutationOf(mutant))
	}
	inputs, err := buildcache.GoTestInputs("", "worker.test", "./internal/worker", nil, mutationOf(mutant))
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(findWorkerProduct(t, inputs), "worker.test")
}

type wasmVariant struct {
	name, entry, selected string
	counted               bool
}

var wasmVariants = []wasmVariant{
	{"sieve", "sieve", "sieve", false},
	{"stats", "stats", "stats", false},
	{"words", "words", "words", false},
	{"transform", "transform", "transform", false},
	{"panic", "panic", "danger", false},
	{"counted", "transform", "transform", true},
	{"multiple", "multiple", "offset,multiply", false},
}

func wasmVariantNamed(t *testing.T, key string) wasmVariant {
	t.Helper()
	for _, variant := range wasmVariants {
		if variant.name == key {
			return variant
		}
	}
	t.Fatalf("unknown Worker Wasm variant %s", key)
	return wasmVariant{}
}

func workerWasmInputs(t *testing.T, key string, mutant *workerMutation) buildcache.Inputs {
	t.Helper()
	variant := wasmVariantNamed(t, key)
	inputs, err := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, mutationOf(mutant))
	if err != nil {
		t.Fatal(err)
	}
	// Go input keys are memoized and share their slices across callers. Detach
	// before extending: concurrent variants must not overwrite each other's
	// fixture paths (or a compiler product's identity).
	inputs.Files = slices.Clone(inputs.Files)
	inputs.Flags = slices.Clone(inputs.Flags)
	inputs.Toolchain = slices.Clone(inputs.Toolchain)
	inputs.Name = "worker wasm " + key
	inputs.Files = append(inputs.Files, "internal/worker/products_test.go", "internal/worker/testdata/wasm/"+variant.entry+".a")
	inputs.Flags = append(inputs.Flags, "selected "+variant.selected, fmt.Sprintf("counted %t", variant.counted))
	request, err := native.RuntimeRequestFor(native.Options{Release: true, Target: "wasm32-wasi", Count: variant.counted})
	if err != nil {
		t.Fatal(err)
	}
	inputs.Flags = append(inputs.Flags, request.Portable...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version"), buildcache.Tool("node", "--version"))
	return inputs
}

func workerWasmProduct(t *testing.T, key string, mutant *workerMutation, build bool) string {
	t.Helper()
	requireWasm(t)
	inputs := workerWasmInputs(t, key, mutant)
	if !build {
		return filepath.Join(findWorkerProduct(t, inputs), "worker")
	}
	binary := workerCompilerProduct(t, mutant, false)
	variant := wasmVariantNamed(t, key)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		output := filepath.Join(directory, "worker")
		entry := "internal/worker/testdata/wasm/" + variant.entry + ".a"
		run := func(arguments ...string) error {
			command := exec.CommandContext(t.Context(), binary, arguments...)
			command.Dir = root
			if report, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("%v: %w\n%s", arguments, err, report)
			}
			return nil
		}
		if !variant.counted {
			return run("worker", entry, "--out", output, "--wasm", variant.selected)
		}
		// Reuse the ordinary Worker, then replace just the module with the counted
		// build. Do not compile the same ordinary module again in the memory unit.
		ordinary := workerWasmProduct(t, "transform", mutant, false)
		if err := os.CopyFS(output, os.DirFS(ordinary)); err != nil {
			return err
		}
		return run("build", "--target", "wasm32-wasi", entry, "-o", filepath.Join(output, "handler.wasm"), "--export", "transform", "--abi-json", filepath.Join(output, "abi.json"), "--count")
	})
	return filepath.Join(directory, "worker")
}

// Not parallel: seeds the compiler shared by the following products.
func TestProduct_WorkerCompiler(t *testing.T) { workerCompilerProduct(t, nil, true) }

// Not parallel: builds the shared test binary before mutant products.
func TestProduct_WorkerChecks(t *testing.T) { workerTestProduct(t, nil, true) }

// This seed precedes other Wasm products, so independent compiler processes
// never race to compile the identical ordinary runtime on a cold tree build.
// Not parallel: seeds the WASI runtime shared by the other variants.
func TestProduct_WorkerWasmTransform(t *testing.T) {
	requireWasm(t)
	workerWasmProduct(t, "transform", nil, true)
}

func runWorkerWasmProductsShard(t *testing.T, shard int) {
	requireWasm(t)
	for _, variant := range wasmVariants {
		if variant.name == "transform" {
			continue
		}
		if familyOwns("TestProduct_WorkerWasm", shard, variant.name) {
			t.Run(variant.name, func(t *testing.T) { workerWasmProduct(t, variant.name, nil, true) })
		}
	}
}

func buildWorkerMutant(t *testing.T, mutant workerMutation, wasm bool) {
	t.Helper()
	if strings.HasPrefix(mutant.check, "TestWorkerModulesMatchSourceResponses") || strings.HasPrefix(mutant.check, "TestWasm") {
		workerCompilerProduct(t, &mutant, true)
	}
	workerTestProduct(t, &mutant, true)
	if wasm {
		if mutant.check == "TestWasmWorkers/transform" || mutant.check == "TestWasmWorkerMemory" {
			workerWasmProduct(t, "transform", &mutant, true)
			if mutant.check == "TestWasmWorkerMemory" {
				workerWasmProduct(t, "counted", &mutant, true)
			}
		}
	}
}

func runWorkerMutantProductsShard(t *testing.T, shard int) {
	requireMutants(t, false)
	for _, mutant := range workerMutants {
		if familyOwns("TestProduct_WorkerMutants", shard, mutant.name) {
			t.Run(mutant.name, func(t *testing.T) { buildWorkerMutant(t, mutant, false) })
		}
	}
}
func runWasmWorkerMutantProductsShard(t *testing.T, shard int) {
	requireMutants(t, true)
	for _, mutant := range wasmWorkerMutants {
		if familyOwns("TestProduct_WorkerWasmMutants", shard, mutant.name) {
			t.Run(mutant.name, func(t *testing.T) { buildWorkerMutant(t, mutant, true) })
		}
	}
}
