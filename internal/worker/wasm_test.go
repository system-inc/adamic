package worker

import (
	"bytes"
	"encoding/json"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func requireWasm(t *testing.T) {
	t.Helper()
	if err := native.ValidateOptions(native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatalf("Workers require setup.sh --wasi-sdk: %v", err)
	}
}

func wasmCompiler(t *testing.T) string {
	t.Helper()
	requireWasm(t)
	return workerCompilerProduct(t, activeMutation(t), false)
}

func wasmOracle(t *testing.T, directory, entry string) string {
	t.Helper()
	oracle := t.TempDir()
	bridge, err := os.ReadFile(filepath.Join(directory, "worker.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	entry, _ = filepath.Abs(entry)
	runtime, _ := filepath.Abs("../../oracle/adamic.mjs")
	write(t, filepath.Join(oracle, "worker.mjs"), strings.Replace(string(bridge), `"./handler.mjs"`, `"`+filepath.ToSlash(entry)+`"`, 1))
	write(t, filepath.Join(oracle, "adamic.mjs"), "export * from "+string(mustJSON(t, runtime))+";\n")
	// The exact same bridge drives source, with no Wasm module or crossing.
	write(t, filepath.Join(oracle, "wasm-state.mjs"), "export async function initializeCrossing() {}\n")
	return filepath.Join(oracle, "worker.mjs")
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

type wasmEvidence struct {
	Calls, Instances, InitialMemory, FinalMemory, BaselineMemory, MaximumMemory, Live int
}

func wasmObserve(t *testing.T, directory, requests, stress string, expectedCalls int) (string, string, wasmEvidence) {
	t.Helper()
	evidencePath := filepath.Join(t.TempDir(), "evidence.json")
	arguments := []string{"--disable-warning=ExperimentalWarning", "wasm-run.mjs", filepath.Join(directory, "worker.mjs"), requests, evidencePath}
	arguments = append(arguments, stress, strconv.Itoa(expectedCalls))
	command := exec.Command("node", arguments...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("Wasm witness: %v\nstderr: %s", err, &stderr)
	}
	var evidence wasmEvidence
	contents, err := os.ReadFile(evidencePath)
	if err != nil || json.Unmarshal(contents, &evidence) != nil {
		t.Fatalf("evidence: %s %v", contents, err)
	}
	return stdout.String(), stderr.String(), evidence
}

func runWasmWorkersShard(t *testing.T, shard int) {
	requireWasm(t)
	for _, fixture := range wasmVariants {
		if fixture.counted || fixture.name == "multiple" {
			continue
		}
		if !familyOwns("TestWasmWorkers", shard, fixture.name) {
			continue
		}
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			entry := filepath.Join("testdata", "wasm", fixture.name+".a")
			requests, _ := filepath.Abs(strings.TrimSuffix(entry, ".a") + ".jsonl")
			directory := workerWasmProduct(t, fixture.name, activeMutation(t), false)
			actual, stderr, evidence := wasmObserve(t, directory, requests, "", 0)
			expected, oracleStderr := node(t, wasmOracle(t, directory, entry), requests, true, entry)
			if actual != expected || stderr != oracleStderr {
				t.Fatalf("Wasm/source oracle mismatch (%s)\ncompiled stderr: %s\nsource stderr: %s", fixture.name, stderr, oracleStderr)
			}
			if evidence.Calls != len(responses(t, actual)) {
				t.Fatalf("Wasm did not run: %#v", evidence)
			}
			if fixture.name == "panic" {
				if evidence.Instances != 3 || stderr != "wasm exploded\nwasm exploded\n" {
					t.Fatalf("panicked instance was not replaced: %#v stderr=%q", evidence, stderr)
				}
			} else if evidence.FinalMemory <= evidence.InitialMemory {
				t.Fatalf("large corpus did not grow memory: %#v", evidence)
			}
			entries, err := os.ReadDir(directory)
			if err != nil || len(entries) != 11 {
				t.Fatalf("want eleven Wasm Worker files: %v %v", entries, err)
			}
			for _, file := range entries {
				if filepath.Ext(file.Name()) != ".mjs" {
					continue
				}
				contents, err := os.ReadFile(filepath.Join(directory, file.Name()))
				if err != nil {
					t.Fatal(err)
				}
				for _, forbidden := range []string{"node:", "process.", "Buffer.", "eval(", "new Function("} {
					if bytes.Contains(contents, []byte(forbidden)) {
						t.Fatalf("%s uses Workers-unavailable %s", file.Name(), forbidden)
					}
				}
			}
			configuration, _ := os.ReadFile(filepath.Join(directory, "wrangler.toml"))
			if !strings.Contains(string(configuration), "type = \"CompiledWasm\"") {
				t.Fatal("missing CompiledWasm rule")
			}
			t.Logf("identical source responses; Wasm calls=%d instances=%d memory=%d -> %d", evidence.Calls, evidence.Instances, evidence.InitialMemory, evidence.FinalMemory)
		})
	}
}

func runWasmWorkerMemoryShard(t *testing.T, shard int) {
	requireWasm(t)
	if !familyOwns("TestWasmWorkerMemory", shard, "transform") {
		return
	}
	t.Run("transform", func(t *testing.T) {
		entry := "testdata/wasm/transform.a"
		directory := workerWasmProduct(t, "counted", activeMutation(t), false)
		requests, _ := filepath.Abs("testdata/wasm/transform.jsonl")
		_, _, evidence := wasmObserve(t, directory, requests, entry, 0)
		if evidence.BaselineMemory != evidence.MaximumMemory || evidence.Live != 0 {
			t.Fatalf("10,000 requests leaked: %#v", evidence)
		}
		t.Logf("10,000 requests: flat linear memory=%d, live=%d, real Wasm calls=%d", evidence.MaximumMemory, evidence.Live, evidence.Calls)
	})
}

func TestWasmPurity(t *testing.T) {
	t.Parallel()
	binary := wasmCompiler(t)
	prefix := "import type { HttpRequest, HttpResponse } from 'adamic/http';\n"
	handler := "\nexport function handle(request: HttpRequest): HttpResponse { return {status: 200, headers: [], body: 'ok'}; }\n"
	cases := []struct{ name, source, selected, witness string }{
		{"console", "export function compute(value: number): number { console.log(String(value)); return value; }", "compute", "calls console"},
		{"transitive console", "function helper(): number { console.error('bad'); return 1; } export function compute(value: number): number { return helper() + value; }", "compute", "calls console"},
		{"global read", "let state = 1; export function compute(value: number): number { return state + value; }", "compute", "accesses module global state"},
		{"global write", "let state = 1; export function compute(value: number): number { state = value; return value; }", "compute", "accesses module global state"},
		{"io", "import { programArguments } from 'adamic'; export function compute(value: number): number { return programArguments().length + value; }", "compute", "calls adamic I/O"},
		{"missing", "export function other(value: number): number { return value; }", "compute", "not a top-level function"},
		{"arrow", "export const compute = (value: number): number => value;", "compute", "not a top-level function"},
		{"abi", "export function compute(value: number): Map<string, number> { return new Map<string, number>(); }", "compute", "compute.return"},
	}
	for _, fixture := range cases {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			entry := filepath.Join(t.TempDir(), "entry.a")
			write(t, entry, prefix+fixture.source+handler)
			directory := filepath.Join(t.TempDir(), "refused")
			command := exec.Command(binary, "worker", entry, "--out", directory, "--wasm", fixture.selected)
			output, err := command.CombinedOutput()
			if err == nil || !strings.Contains(string(output), fixture.witness) || !strings.Contains(string(output), "fix:") {
				t.Fatalf("expected refusal %q with fix: err=%v\n%s", fixture.witness, err, output)
			}
			if _, err := os.Stat(directory); !os.IsNotExist(err) {
				t.Fatalf("refusal wrote output: %v", err)
			}
		})
	}
}

func runWasmMultipleFunctionsShard(t *testing.T, shard int) {
	requireWasm(t)
	if !familyOwns("TestWasmMultipleFunctions", shard, "multiple") {
		return
	}
	entry := "testdata/wasm/multiple.a"
	requests, _ := filepath.Abs("testdata/wasm/multiple.jsonl")
	directory := workerWasmProduct(t, "multiple", activeMutation(t), false)
	actual, stderr, evidence := wasmObserve(t, directory, requests, "", 5)
	expected, oracleStderr := node(t, wasmOracle(t, directory, entry), requests, true, entry)
	if actual != expected || stderr != oracleStderr || evidence.Instances != 1 {
		t.Fatalf("multiple functions and startup mismatch: %#v\ncompiled: %s\nsource: %s", evidence, actual, expected)
	}
	results := responses(t, actual)
	if results[0].Body != "39" || results[1].Body != "45" {
		t.Fatalf("module top level did not run exactly once: %v", results)
	}
	t.Logf("two selected functions, private declaration, transitive call and startup: calls=%d instances=%d", evidence.Calls, evidence.Instances)
}
