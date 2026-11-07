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
func findWorkerProduct(test testing.TB, inputs buildcache.Inputs) string {
	test.Helper()
	directory, builtAheadError := buildcache.BuiltAhead(
		inputs,
		fmt.Errorf("Worker product %s not built ahead; run TestProduct_Worker tests first", inputs.Name),
	)
	if builtAheadError != nil {
		test.Fatal(builtAheadError)
	}
	return directory
}

func activeMutation(test *testing.T) *workerMutation {
	test.Helper()
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
	test.Fatalf("unknown Worker mutation %q", name)
	return nil
}

func workerCompilerProduct(test *testing.T, mutant *workerMutation, build bool) string {
	test.Helper()
	if build {
		if mutant == nil {
			return buildcache.AdamicBuiltElsewhere(test)
		}
		return buildcache.GoBuildInMutated(test, "", "adamic", "./cmd/adamic", nil, mutationOf(mutant))
	}
	inputs, goBuildInInputsError := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, mutationOf(mutant))
	if goBuildInInputsError != nil {
		test.Fatal(goBuildInInputsError)
	}
	return filepath.Join(findWorkerProduct(test, inputs), "adamic")
}

func workerTestProduct(test *testing.T, mutant *workerMutation, build bool) string {
	test.Helper()
	if build {
		return buildcache.GoTestMutated(test, "", "worker.test", "./internal/worker", nil, mutationOf(mutant))
	}
	inputs, goTestInputsError := buildcache.GoTestInputs(
		"",
		"worker.test",
		"./internal/worker",
		nil,
		mutationOf(mutant),
	)
	if goTestInputsError != nil {
		test.Fatal(goTestInputsError)
	}
	return filepath.Join(findWorkerProduct(test, inputs), "worker.test")
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

func wasmVariantNamed(test *testing.T, key string) wasmVariant {
	test.Helper()
	for _, variant := range wasmVariants {
		if variant.name == key {
			return variant
		}
	}
	test.Fatalf("unknown Worker Wasm variant %s", key)
	return wasmVariant{}
}

func workerWasmInputs(test *testing.T, key string, mutant *workerMutation) buildcache.Inputs {
	test.Helper()
	variant := wasmVariantNamed(test, key)
	inputs, goBuildInInputsError := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, mutationOf(mutant))
	if goBuildInInputsError != nil {
		test.Fatal(goBuildInInputsError)
	}
	// Go input keys are memoized and share their slices across callers. Detach
	// before extending: concurrent variants must not overwrite each other's
	// fixture paths (or a compiler product's identity).
	inputs.Files = slices.Clone(inputs.Files)
	inputs.Flags = slices.Clone(inputs.Flags)
	inputs.Toolchain = slices.Clone(inputs.Toolchain)
	inputs.Name = "worker wasm " + key
	inputs.Files = append(
		inputs.Files,
		"internal/worker/products_test.go",
		"internal/worker/testdata/wasm/"+variant.entry+".a",
	)
	inputs.Flags = append(inputs.Flags, "selected "+variant.selected, fmt.Sprintf("counted %t", variant.counted))
	request, runtimeRequestForError := native.RuntimeRequestFor(
		native.Options{Release: true, Target: "wasm32-wasi", Count: variant.counted},
	)
	if runtimeRequestForError != nil {
		test.Fatal(runtimeRequestForError)
	}
	inputs.Flags = append(inputs.Flags, request.Portable...)
	inputs.Toolchain = append(
		inputs.Toolchain,
		buildcache.Tool(request.Compiler, "--version"),
		buildcache.Tool(request.Archiver, "--version"),
		buildcache.Tool("node", "--version"),
	)
	return inputs
}

func workerWasmProduct(test *testing.T, key string, mutant *workerMutation, build bool) string {
	test.Helper()
	requireWasm(test)
	inputs := workerWasmInputs(test, key, mutant)
	if !build {
		return filepath.Join(findWorkerProduct(test, inputs), "worker")
	}
	binary := workerCompilerProduct(test, mutant, false)
	variant := wasmVariantNamed(test, key)
	root, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	directory := buildcache.Product(test, inputs, func(directory string) error {
		output := filepath.Join(directory, "worker")
		entry := "internal/worker/testdata/wasm/" + variant.entry + ".a"
		run := func(arguments ...string) error {
			command := exec.CommandContext(test.Context(), binary, arguments...)
			command.Dir = root
			if report, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
				return fmt.Errorf("%v: %w\n%s", arguments, combinedOutputError, report)
			}
			return nil
		}
		if !variant.counted {
			return run("worker", entry, "--out", output, "--wasm", variant.selected)
		}
		// Reuse the ordinary Worker, then replace just the module with the counted
		// build. Do not compile the same ordinary module again in the memory unit.
		ordinary := workerWasmProduct(test, "transform", mutant, false)
		if copyFsError := os.CopyFS(output, os.DirFS(ordinary)); copyFsError != nil {
			return copyFsError
		}
		return run(
			"build",
			"--target",
			"wasm32-wasi",
			entry,
			"-o",
			filepath.Join(output, "handler.wasm"),
			"--export",
			"transform",
			"--abi-json",
			filepath.Join(output, "abi.json"),
			"--count",
		)
	})
	return filepath.Join(directory, "worker")
}

// Not parallel: seeds the compiler shared by the following products.
func TestProduct_WorkerCompiler(test *testing.T) { workerCompilerProduct(test, nil, true) }

// Not parallel: builds the shared test binary before mutant products.
func TestProduct_WorkerChecks(test *testing.T) { workerTestProduct(test, nil, true) }

// This seed precedes other Wasm products, so independent compiler processes
// never race to compile the identical ordinary runtime on a cold tree build.
// Not parallel: seeds the WASI runtime shared by the other variants.
func TestProduct_WorkerWasmTransform(test *testing.T) {
	requireWasm(test)
	workerWasmProduct(test, "transform", nil, true)
}

func runWorkerWasmProductsShard(test *testing.T, shard int) {
	requireWasm(test)
	for _, variant := range wasmVariants {
		if variant.name == "transform" {
			continue
		}
		if familyOwns("TestProduct_WorkerWasm", shard, variant.name) {
			test.Run(variant.name, func(test *testing.T) { workerWasmProduct(test, variant.name, nil, true) })
		}
	}
}

func buildWorkerMutant(test *testing.T, mutant workerMutation, wasm bool) {
	test.Helper()
	if strings.HasPrefix(mutant.check, "TestWorkerModulesMatchSourceResponses") ||
		strings.HasPrefix(mutant.check, "TestWasm") {
		workerCompilerProduct(test, &mutant, true)
	}
	workerTestProduct(test, &mutant, true)
	if wasm {
		if mutant.check == "TestWasmWorkers/transform" || mutant.check == "TestWasmWorkerMemory" {
			workerWasmProduct(test, "transform", &mutant, true)
			if mutant.check == "TestWasmWorkerMemory" {
				workerWasmProduct(test, "counted", &mutant, true)
			}
		}
	}
}

func runWorkerMutantProductsShard(test *testing.T, shard int) {
	requireMutants(test, false)
	for _, mutant := range workerMutants {
		if familyOwns("TestProduct_WorkerMutants", shard, mutant.name) {
			test.Run(mutant.name, func(test *testing.T) { buildWorkerMutant(test, mutant, false) })
		}
	}
}
func runWasmWorkerMutantProductsShard(test *testing.T, shard int) {
	requireMutants(test, true)
	for _, mutant := range wasmWorkerMutants {
		if familyOwns("TestProduct_WorkerWasmMutants", shard, mutant.name) {
			test.Run(mutant.name, func(test *testing.T) { buildWorkerMutant(test, mutant, true) })
		}
	}
}
