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

func wasiClang(t *testing.T) string {
	t.Helper()
	if err := native.ValidateOptions(native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	compiler, err := native.WASIClang()
	if err != nil {
		t.Fatal(err)
	}
	return compiler
}

// Compilation, fixture bytes and generated reference observations are inputs,
// built ahead. The harness and each planted mutant still run in every test.
func integrationInputs(t *testing.T, name string) buildcache.Inputs {
	t.Helper()
	inputs, err := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = name
	inputs.Files = append(append([]string{}, inputs.Files...), "internal/worker/wasm", "cmd/adamic/testdata/wasi", "internal/native/wasm")
	request, err := native.RuntimeRequestFor(native.Options{Target: "wasm32-wasi", Count: true})
	if err != nil {
		t.Fatal(err)
	}
	inputs.Flags = append(append([]string{}, inputs.Flags...), request.Portable...)
	inputs.Toolchain = append(append([]string{}, inputs.Toolchain...), buildcache.Tool(wasiClang(t), "--version"), buildcache.Tool(request.Archiver, "--version"), buildcache.Tool("node", "--version"))
	return inputs
}

func commandRunner(t *testing.T, root string) func(string, ...string) []byte {
	t.Helper()
	return func(name string, args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = root
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, output)
		}
		return output
	}
}
