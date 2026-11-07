package wasm

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/worker/testtools"
)

// Select tools before parallel tests; no builds or verdicts live in TestMain.
func TestMain(m *testing.M) { testtools.ConfigureWASI(); os.Exit(m.Run()) }

func wasiClang(test *testing.T) string {
	test.Helper()
	if validateOptionsError := native.ValidateOptions(
		native.Options{Target: "wasm32-wasi"},
	); validateOptionsError != nil {
		test.Fatal(validateOptionsError)
	}
	compiler, wasiClangError := native.WASIClang()
	if wasiClangError != nil {
		test.Fatal(wasiClangError)
	}
	return compiler
}

// Compilation, fixture bytes and generated reference observations are inputs,
// built ahead. The harness and each planted mutant still run in every test.
func integrationInputs(test *testing.T, name string) buildcache.Inputs {
	test.Helper()
	inputs, goBuildInInputsError := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, nil)
	if goBuildInInputsError != nil {
		test.Fatal(goBuildInInputsError)
	}
	inputs.Name = name
	inputs.Files = append(
		append([]string{}, inputs.Files...),
		"internal/worker/wasm",
		"cmd/adamic/testdata/wasi",
		"internal/native/wasm",
	)
	request, runtimeRequestForError := native.RuntimeRequestFor(native.Options{Target: "wasm32-wasi", Count: true})
	if runtimeRequestForError != nil {
		test.Fatal(runtimeRequestForError)
	}
	inputs.Flags = append(append([]string{}, inputs.Flags...), request.Portable...)
	inputs.Toolchain = append(
		append([]string{}, inputs.Toolchain...),
		buildcache.Tool(wasiClang(test), "--version"),
		buildcache.Tool(request.Archiver, "--version"),
		buildcache.Tool("node", "--version"),
	)
	return inputs
}

func commandRunner(test *testing.T, root string) func(string, ...string) []byte {
	test.Helper()
	return func(name string, args ...string) []byte {
		test.Helper()
		timeoutContext, cancel := context.WithTimeout(test.Context(), 3*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(timeoutContext, name, args...)
		cmd.Dir = root
		output, combinedOutputError := cmd.CombinedOutput()
		if combinedOutputError != nil {
			test.Fatalf("%s %v: %v\n%s", name, args, combinedOutputError, output)
		}
		return output
	}
}
