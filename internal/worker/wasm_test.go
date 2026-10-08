package worker

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func wasmCompiler(t *testing.T) string {
	t.Helper()
	if os.Getenv("WASI_SYSROOT") == "" {
		t.Skip("Wasm Workers require WASI_SYSROOT; run bash cloud/setup.sh --wasi-sdk")
	}
	binary := filepath.Join(t.TempDir(), "adamic")
	command := exec.Command("go", "build", "-o", binary, "../../cmd/adamic")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	return binary
}

func wasmCommand(t *testing.T, binary, entry, directory, names string) {
	t.Helper()
	command := exec.Command(binary, "worker", entry, "--out", directory, "--wasm", names)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Wasm Worker build: %v\n%s", err, output)
	}
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

func TestWasmWorkers(t *testing.T) {
	t.Parallel()
	binary := wasmCompiler(t)
	for _, fixture := range []struct{ name, function string }{{"sieve", "sieve"}, {"stats", "stats"}, {"words", "words"}, {"transform", "transform"}, {"panic", "danger"}} {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			entry := filepath.Join("testdata", "wasm", fixture.name+".a")
			requests, _ := filepath.Abs(strings.TrimSuffix(entry, ".a") + ".jsonl")
			directory := t.TempDir()
			wasmCommand(t, binary, entry, directory, fixture.function)
			actual, stderr, evidence := wasmObserve(t, directory, requests, "", 0)
			expected, oracleStderr := node(t, wasmOracle(t, directory, entry), requests, true)
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

func TestWasmWorkerMemory(t *testing.T) {
	t.Parallel()
	binary := wasmCompiler(t)
	entry := "testdata/wasm/transform.a"
	directory := t.TempDir()
	wasmCommand(t, binary, entry, directory, "transform")
	// Rebuild only the module with counters; this is the same entry and ABI.
	command := exec.Command(binary, "build", "--target", "wasm32-wasi", entry, "-o", filepath.Join(directory, "handler.wasm"), "--export", "transform", "--abi-json", filepath.Join(directory, "abi.json"), "--count")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("counted Wasm: %v\n%s", err, output)
	}
	requests, _ := filepath.Abs("testdata/wasm/transform.jsonl")
	_, _, evidence := wasmObserve(t, directory, requests, entry, 0)
	if evidence.BaselineMemory != evidence.MaximumMemory || evidence.Live != 0 {
		t.Fatalf("10,000 requests leaked: %#v", evidence)
	}
	t.Logf("10,000 requests: flat linear memory=%d, live=%d, real Wasm calls=%d", evidence.MaximumMemory, evidence.Live, evidence.Calls)
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

func TestWasmMultipleFunctions(t *testing.T) {
	t.Parallel()
	binary := wasmCompiler(t)
	entry := filepath.Join(t.TempDir(), "multiple.a")
	write(t, entry, `
import type { HttpRequest, HttpResponse } from 'adamic/http';
function offset(value: number): number { return value + 7; }
export function multiply(left: number, right: number): number { return offset(left) * right; }
let starts = 0;
starts += 1;
const startup = offset(1);
export function handle(request: HttpRequest): HttpResponse {
 return {status: 200, headers: [], body: String(starts + startup + multiply(Number(request.body), 3) + offset(2))};
}
`)
	requests := filepath.Join(t.TempDir(), "requests.jsonl")
	write(t, requests, string(mustJSON(t, map[string]any{"method": "POST", "url": "https://worker.test/", "body": "0"}))+"\n"+string(mustJSON(t, map[string]any{"method": "POST", "url": "https://worker.test/", "body": "2"}))+"\n")
	directory := t.TempDir()
	wasmCommand(t, binary, entry, directory, "offset,multiply")
	actual, stderr, evidence := wasmObserve(t, directory, requests, "", 5)
	expected, oracleStderr := node(t, wasmOracle(t, directory, entry), requests, true)
	if actual != expected || stderr != oracleStderr || evidence.Instances != 1 {
		t.Fatalf("multiple functions and startup mismatch: %#v\ncompiled: %s\nsource: %s", evidence, actual, expected)
	}
	results := responses(t, actual)
	if results[0].Body != "39" || results[1].Body != "45" {
		t.Fatalf("module top level did not run exactly once: %v", results)
	}
	t.Logf("two selected functions, private declaration, transitive call and startup: calls=%d instances=%d", evidence.Calls, evidence.Instances)
}
