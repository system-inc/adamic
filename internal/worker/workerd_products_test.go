package worker

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// The two products compile inputs only: corpus observations always run fresh.
func workerdComputeProduct(test *testing.T, wasm, build bool) string {
	test.Helper()
	inputs, err := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, nil)
	if err != nil {
		test.Fatal(err)
	}
	inputs.Files = slices.Clone(inputs.Files)
	inputs.Flags = slices.Clone(inputs.Flags)
	inputs.Toolchain = slices.Clone(inputs.Toolchain)
	inputs.Name = fmt.Sprintf("workerd compute wasm=%t", wasm)
	inputs.Files = append(inputs.Files, "internal/worker/workerd_products_test.go", "workers/compute/handler.a")
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("wasm=%t", wasm))
	if wasm {
		requireWasm(test)
		request, err := native.RuntimeRequestFor(native.Options{Release: true, Target: "wasm32-wasi"})
		if err != nil {
			test.Fatal(err)
		}
		inputs.Flags = append(inputs.Flags, request.Portable...)
		inputs.Toolchain = append(
			inputs.Toolchain,
			buildcache.Tool(request.Compiler, "--version"),
			buildcache.Tool(request.Archiver, "--version"),
		)
	}
	if !build {
		return filepath.Join(findWorkerProduct(test, inputs), "worker")
	}
	compiler := workerCompilerProduct(test, nil, false)
	root, err := filepath.Abs("../..")
	if err != nil {
		test.Fatal(err)
	}
	directory := buildcache.Product(test, inputs, func(directory string) error {
		args := []string{"worker", "workers/compute/handler.a", "--out", filepath.Join(directory, "worker")}
		if wasm {
			args = append(args, "--wasm", "primes,summarize")
		}
		command := exec.CommandContext(test.Context(), compiler, args...)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("compute build: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(directory, "worker")
}

// Not parallel: compiles a shared corpus input ahead of all replay units.
func TestProduct_WorkerdJavaScript(test *testing.T) {
	requireWorkerd(test)
	workerdComputeProduct(test, false, true)
}

// Not parallel: compiles the shared Wasm corpus input ahead of all replay units.
func TestProduct_WorkerdWebAssembly(test *testing.T) {
	requireWorkerd(test)
	workerdComputeProduct(test, true, true)
}
