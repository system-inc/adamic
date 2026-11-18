package worker

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native_runtime"
)

func workerdEnvironmentWasmProduct(t *testing.T, build bool) string {
	t.Helper()
	requireWasm(t)
	inputs, err := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Files = slices.Clone(inputs.Files)
	inputs.Flags = slices.Clone(inputs.Flags)
	inputs.Toolchain = slices.Clone(inputs.Toolchain)
	inputs.Name = "workerd string environment with Wasm compute and independent twin"
	inputs.Files = append(inputs.Files, "internal/worker/environment_wasm_workerd_test.go", "workers/environment-wasm/handler.a", "workers/environment-wasm/twin/worker.ts")
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("node", "--version"))
	request, err := native_runtime.RuntimeRequestFor(native_runtime.Options{Release: true, Target: "wasm32-wasi"})
	if err != nil {
		t.Fatal(err)
	}
	inputs.Flags = append(inputs.Flags, request.Portable...)
	inputs.Flags = append(inputs.Flags, "selected compute")
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool(request.Compiler, "--target="+request.Options.Target, "--version"), buildcache.Tool(request.Archiver, "--version"))
	if !build {
		return findWorkerProduct(t, inputs)
	}
	compiler := workerCompilerProduct(t, nil, false)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		command := exec.CommandContext(t.Context(), compiler, "worker", "workers/environment-wasm/handler.a", "--out", filepath.Join(directory, "worker"), "--wasm", "compute")
		command.Dir = root
		if report, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("environment/Wasm Worker build: %w\n%s", err, report)
		}
		twin := filepath.Join(directory, "twin")
		if err := os.Mkdir(twin, 0755); err != nil {
			return err
		}
		command = exec.CommandContext(t.Context(), "node", "--disable-warning=ExperimentalWarning", "--input-type=module", "-e", `import {readFileSync} from 'node:fs'; import {stripTypeScriptTypes} from 'node:module'; process.stdout.write(stripTypeScriptTypes(readFileSync(process.argv[1],'utf8')));`, filepath.Join(root, "workers/environment-wasm/twin/worker.ts"))
		source, err := command.Output()
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(twin, "worker.mjs"), source, 0644)
	})
}
func TestProduct_WorkerdEnvironmentWasm(t *testing.T) {
	t.Parallel()
	requireWorkerd(t)
	workerdEnvironmentWasmProduct(t, true)
}
func TestWorkerdEnvironmentWasmPairing(t *testing.T) {
	t.Parallel()
	requireWorkerd(t)
	product := workerdEnvironmentWasmProduct(t, false)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"worker", "twin"} {
		server, err := startWorkerdWithBindings(t, filepath.Join(product, kind), "127.0.0.1:0", `(name="PREFIX",text="zone"),(name="EXTRA",text="SECRET_UNDECLARED_f391")`)
		if err != nil {
			t.Fatal(err)
		}
		command := exec.CommandContext(t.Context(), "node", filepath.Join(root, "workers/replay.mjs"), "--url", "http://"+server.address, filepath.Join(root, "workers/environment-wasm/corpus/requests.jsonl"), "--compare", filepath.Join(root, "workers/environment-wasm/corpus/responses.jsonl"))
		if report, err := command.CombinedOutput(); err != nil {
			t.Fatalf("environment/Wasm %s corpus: %v\n%s", kind, err, report)
		}
	}
}
func TestWorkerdEnvironmentWasmMissingBinding(t *testing.T) {
	t.Parallel()
	requireWorkerd(t)
	product := workerdEnvironmentWasmProduct(t, false)
	const message = `environment binding "PREFIX" is missing or not a string`
	for _, kind := range []string{"worker", "twin"} {
		server, err := startWorkerdWithBindings(t, filepath.Join(product, kind), "127.0.0.1:0", `(name="EXTRA",text="SECRET_UNDECLARED_f391")`)
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequestWithContext(t.Context(), "GET", "http://"+server.address+"/two", nil)
		if err != nil {
			t.Fatal(err)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != 500 || string(body) != "internal error" {
			t.Fatalf("environment/Wasm missing binding: status %d body %q", response.StatusCode, body)
		}
		var log []byte
		deadline := time.Now().Add(time.Second)
		for {
			log, err = os.ReadFile(server.logPath)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(log), message) || time.Now().After(deadline) {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if strings.TrimSpace(string(log)) != message {
			t.Fatalf("environment/Wasm missing binding log: %q", log)
		}
	}
}
