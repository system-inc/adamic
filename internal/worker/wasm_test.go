package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func requireWasm(test *testing.T) {
	test.Helper()
	if err := native.ValidateOptions(native.Options{Target: "wasm32-wasi"}); err != nil {
		test.Fatalf("Workers require setup.sh --wasi-sdk: %v", err)
	}
}

func wasmCompiler(test *testing.T) string {
	test.Helper()
	requireWasm(test)
	return workerCompilerProduct(test, activeMutation(test), false)
}

func wasmOracle(test *testing.T, directory, entry string) string {
	test.Helper()
	oracle := test.TempDir()
	bridge, err := os.ReadFile(filepath.Join(directory, "worker.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	entry, _ = filepath.Abs(entry)
	runtime, _ := filepath.Abs("../../oracle/adamic.mjs")
	write(
		test,
		filepath.Join(oracle, "worker.mjs"),
		strings.Replace(string(bridge), `"./handler.mjs"`, `"`+filepath.ToSlash(entry)+`"`, 1),
	)
	write(test, filepath.Join(oracle, "adamic.mjs"), "export * from "+string(mustJSON(test, runtime))+";\n")
	// The exact same bridge drives source, with no Wasm module or crossing.
	write(test, filepath.Join(oracle, "wasm-state.mjs"), "export async function initializeCrossing() {}\n")
	return filepath.Join(oracle, "worker.mjs")
}

func mustJSON(test *testing.T, value any) []byte {
	test.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		test.Fatal(err)
	}
	return encoded
}

type wasmEvidence struct {
	Calls, Instances, InitialMemory, FinalMemory, BaselineMemory, MaximumMemory, Live int
}

func wasmObserve(
	test *testing.T,
	directory, requests, stress string,
	expectedCalls int,
) (string, string, wasmEvidence) {
	test.Helper()
	evidencePath := filepath.Join(test.TempDir(), "evidence.json")
	arguments := []string{
		"--disable-warning=ExperimentalWarning",
		"wasm-run.mjs",
		filepath.Join(directory, "worker.mjs"),
		requests,
		evidencePath,
	}
	arguments = append(arguments, stress, strconv.Itoa(expectedCalls))
	command := exec.Command("node", arguments...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		test.Fatalf("Wasm witness: %v\nstderr: %s", err, &stderr)
	}
	var evidence wasmEvidence
	contents, err := os.ReadFile(evidencePath)
	if err != nil || json.Unmarshal(contents, &evidence) != nil {
		test.Fatalf("evidence: %s %v", contents, err)
	}
	return stdout.String(), stderr.String(), evidence
}

func runWasmWorkersShard(test *testing.T, shard int) {
	requireWasm(test)
	for _, fixture := range wasmVariants {
		if fixture.counted || fixture.name == "multiple" {
			continue
		}
		if !familyOwns("TestWasmWorkers", shard, fixture.name) {
			continue
		}
		test.Run(fixture.name, func(test *testing.T) {
			test.Parallel()
			entry := filepath.Join("testdata", "wasm", fixture.name+".a")
			requests, _ := filepath.Abs(strings.TrimSuffix(entry, ".a") + ".jsonl")
			directory := workerWasmProduct(test, fixture.name, activeMutation(test), false)
			actual, stderr, evidence := wasmObserve(test, directory, requests, "", 0)
			expected, oracleStderr := node(test, wasmOracle(test, directory, entry), requests, true, entry)
			if actual != expected || stderr != oracleStderr {
				test.Fatalf(
					"Wasm/source oracle mismatch (%s)\ncompiled stderr: %s\nsource stderr: %s",
					fixture.name,
					stderr,
					oracleStderr,
				)
			}
			if evidence.Calls != len(responses(test, actual)) {
				test.Fatalf("Wasm did not run: %#v", evidence)
			}
			if fixture.name == "panic" {
				if evidence.Instances != 3 || stderr != "wasm exploded\nwasm exploded\n" {
					test.Fatalf("panicked instance was not replaced: %#v stderr=%q", evidence, stderr)
				}
			} else if evidence.FinalMemory <= evidence.InitialMemory {
				test.Fatalf("large corpus did not grow memory: %#v", evidence)
			}
			entries, err := os.ReadDir(directory)
			if err != nil || len(entries) != 11 {
				test.Fatalf("want eleven Wasm Worker files: %v %v", entries, err)
			}
			for _, file := range entries {
				if filepath.Ext(file.Name()) != ".mjs" {
					continue
				}
				contents, err := os.ReadFile(filepath.Join(directory, file.Name()))
				if err != nil {
					test.Fatal(err)
				}
				if forbidden := workersUnavailable(contents); forbidden != "" {
					test.Fatalf("%s uses Workers-unavailable %s", file.Name(), forbidden)
				}
			}
			configuration, _ := os.ReadFile(filepath.Join(directory, "wrangler.toml"))
			if !strings.Contains(string(configuration), "type = \"CompiledWasm\"") {
				test.Fatal("missing CompiledWasm rule")
			}
			test.Logf(
				"identical source responses; Wasm calls=%d instances=%d memory=%d -> %d",
				evidence.Calls,
				evidence.Instances,
				evidence.InitialMemory,
				evidence.FinalMemory,
			)
		})
	}
}

func runWasmWorkerMemoryShard(test *testing.T, shard int) {
	requireWasm(test)
	if !familyOwns("TestWasmWorkerMemory", shard, "transform") {
		return
	}
	test.Run("transform", func(test *testing.T) {
		entry := "testdata/wasm/transform.a"
		directory := workerWasmProduct(test, "counted", activeMutation(test), false)
		requests, _ := filepath.Abs("testdata/wasm/transform.jsonl")
		_, _, evidence := wasmObserve(test, directory, requests, entry, 0)
		if evidence.BaselineMemory != evidence.MaximumMemory || evidence.Live != 0 {
			test.Fatalf("10,000 requests leaked: %#v", evidence)
		}
		test.Logf(
			"10,000 requests: flat linear memory=%d, live=%d, real Wasm calls=%d",
			evidence.MaximumMemory,
			evidence.Live,
			evidence.Calls,
		)
	})
}

func TestWasmPurity(test *testing.T) {
	test.Parallel()
	binary := wasmCompiler(test)
	prefix := "import type { HttpRequest, HttpResponse } from 'adamic/http';\n"
	handler := "\nexport function handle(request: HttpRequest): HttpResponse { return {status: 200, headers: [], body: 'ok'}; }\n"
	cases := []struct{ name, source, selected, witness string }{
		{
			"console",
			"export function compute(value: number): number { console.log(String(value)); return value; }",
			"compute",
			"calls console",
		},
		{
			"transitive console",
			"function helper(): number { console.error('bad'); return 1; } export function compute(value: number): number { return helper() + value; }",
			"compute",
			"calls console",
		},
		{
			"global read",
			"let state = 1; export function compute(value: number): number { return state + value; }",
			"compute",
			"accesses module global state",
		},
		{
			"global write",
			"let state = 1; export function compute(value: number): number { state = value; return value; }",
			"compute",
			"accesses module global state",
		},
		{
			"io",
			"import { programArguments } from 'adamic'; export function compute(value: number): number { return programArguments().length + value; }",
			"compute",
			"calls adamic I/O",
		},
		{
			"missing",
			"export function other(value: number): number { return value; }",
			"compute",
			"not a top-level function",
		},
		{"arrow", "export const compute = (value: number): number => value;", "compute", "not a top-level function"},
		{
			"abi",
			"export function compute(value: number): Map<string, number> { return new Map<string, number>(); }",
			"compute",
			"compute.return",
		},
	}
	for _, fixture := range cases {
		test.Run(fixture.name, func(test *testing.T) {
			test.Parallel()
			entry := filepath.Join(test.TempDir(), "entry.a")
			write(test, entry, prefix+fixture.source+handler)
			directory := filepath.Join(test.TempDir(), "refused")
			if fixture.name == "console" {
				// Refusal is before the build seam, not a verdict obtained by
				// compiling C. In particular the accept-console mutant must
				// fail here without warming an unshipped runtime archive.
				builds := 0
				err := BuildWith(
					entry,
					directory,
					Options{Wasm: []string{fixture.selected}, BuildWasm: func(string, string, string, []string) error {
						builds++
						return fmt.Errorf("unexpected native build")
					}},
				)
				if builds != 0 ||
					err == nil ||
					!strings.Contains(err.Error(), fixture.witness) ||
					!strings.Contains(err.Error(), "fix:") {
					test.Fatalf("expected refusal before native build: builds=%d err=%v", builds, err)
				}
			}
			command := exec.Command(binary, "worker", entry, "--out", directory, "--wasm", fixture.selected)
			output, err := command.CombinedOutput()
			if err == nil ||
				!strings.Contains(string(output), fixture.witness) ||
				!strings.Contains(string(output), "fix:") {
				test.Fatalf("expected refusal %q with fix: err=%v\n%s", fixture.witness, err, output)
			}
			if _, err := os.Stat(directory); !os.IsNotExist(err) {
				test.Fatalf("refusal wrote output: %v", err)
			}
		})
	}
}

func runWasmMultipleFunctionsShard(test *testing.T, shard int) {
	requireWasm(test)
	if !familyOwns("TestWasmMultipleFunctions", shard, "multiple") {
		return
	}
	entry := "testdata/wasm/multiple.a"
	requests, _ := filepath.Abs("testdata/wasm/multiple.jsonl")
	directory := workerWasmProduct(test, "multiple", activeMutation(test), false)
	actual, stderr, evidence := wasmObserve(test, directory, requests, "", 5)
	expected, oracleStderr := node(test, wasmOracle(test, directory, entry), requests, true, entry)
	if actual != expected || stderr != oracleStderr || evidence.Instances != 1 {
		test.Fatalf(
			"multiple functions and startup mismatch: %#v\ncompiled: %s\nsource: %s",
			evidence,
			actual,
			expected,
		)
	}
	results := responses(test, actual)
	if results[0].Body != "39" || results[1].Body != "45" {
		test.Fatalf("module top level did not run exactly once: %v", results)
	}
	test.Logf(
		"two selected functions, private declaration, transitive call and startup: calls=%d instances=%d",
		evidence.Calls,
		evidence.Instances,
	)
}
