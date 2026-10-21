package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}

func node(t *testing.T, worker, requests string, oracle bool) (string, string) {
	t.Helper()
	arguments := []string{"--disable-warning=ExperimentalWarning", "run.mjs", worker, requests}
	if oracle {
		arguments = append(arguments, "--oracle-runtime")
	}
	command := exec.Command("node", arguments...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("Node: %v\nstdout: %s\nstderr: %s", err, &stdout, &stderr)
	}
	return stdout.String(), stderr.String()
}

type response struct {
	Status  int         `json:"status"`
	Headers [][2]string `json:"headers"`
	Body    string      `json:"body"`
}

func responses(t *testing.T, text string) []response {
	t.Helper()
	var result []response
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		var value response
		if err := json.Unmarshal([]byte(line), &value); err != nil {
			t.Fatal(err)
		}
		result = append(result, value)
	}
	return result
}

func header(result response, name string) string {
	for _, field := range result.Headers {
		if field[0] == name {
			return field[1]
		}
	}
	return ""
}

// TestWorkers runs the public command, then the same bridge over untouched source.
// Explicit expectations hold the shared bridge itself to the platform contract.
func TestWorkers(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "adamic")
	command := exec.Command("go", "build", "-o", binary, "../../cmd/adamic")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	fixtures, err := filepath.Glob("testdata/*.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		name := strings.TrimSuffix(filepath.Base(fixture), ".a")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			directory := filepath.Join(t.TempDir(), "compiled")
			command := exec.Command(binary, "worker", fixture, "--out", directory)
			output, err := command.CombinedOutput()
			if name == "wrong" {
				if err == nil {
					t.Fatal("accepted handle returning string")
				}
				for _, part := range []string{"refused:", "found handle:", "=> string", "fix:", "import type { HttpRequest, HttpResponse }"} {
					if !strings.Contains(string(output), part) {
						t.Fatalf("missing refusal %q: %s", part, output)
					}
				}
				if _, err := os.Stat(directory); !os.IsNotExist(err) {
					t.Fatalf("refusal wrote output: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("adamic worker: %v\n%s", err, output)
			}
			configuration, err := os.ReadFile(filepath.Join(directory, "wrangler.toml"))
			if err != nil {
				t.Fatal(err)
			}
			wantConfiguration := fmt.Sprintf("name = %q\nmain = \"worker.mjs\"\ncompatibility_date = \"2026-10-07\"\n", name)
			if string(configuration) != wantConfiguration {
				t.Fatalf("configuration: %s", configuration)
			}
			entries, err := os.ReadDir(directory)
			if err != nil || len(entries) != 4 {
				t.Fatalf("want exactly four generated files: %v, %v", entries, err)
			}
			source, _ := filepath.Abs(fixture)
			requests, _ := filepath.Abs(strings.TrimSuffix(fixture, ".a") + ".jsonl")
			oracleDirectory := t.TempDir()
			// Only the handler import changes. The bridge bytes otherwise stay identical.
			bridge, err := os.ReadFile(filepath.Join(directory, "worker.mjs"))
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(oracleDirectory, "worker.mjs"), strings.Replace(string(bridge), `"./handler.mjs"`, fmt.Sprintf("%q", source), 1))
			runtime, _ := filepath.Abs("../../oracle/adamic.mjs")
			write(t, filepath.Join(oracleDirectory, "adamic.mjs"), fmt.Sprintf("export * from %q;\n", runtime))
			actual, stderr := node(t, filepath.Join(directory, "worker.mjs"), requests, false)
			expected, oracleStderr := node(t, filepath.Join(oracleDirectory, "worker.mjs"), requests, true)
			if actual != expected || stderr != oracleStderr {
				t.Fatalf("oracle mismatch\ncompiled: %s\nsource: %s\ncompiled stderr: %s\nsource stderr: %s", actual, expected, stderr, oracleStderr)
			}
			results := responses(t, actual)
			switch name {
			case "echo":
				if len(results) != 3 {
					t.Fatalf("response count: %d", len(results))
				}
				bodies := []string{
					"1\nGET\nhttps://worker.test/a%2Fb/%E4%B8%96?z=last&a=one+two&a=%E4%B8%96%F0%9F%8C%8D&empty=&bad=%FF\n/a%2Fb/%E4%B8%96\nquery:\nz=last\na=one two\na=世🌍\nempty=\nbad=�\nheaders:\na-first=first\nx-repeated=one, two\nbody:\n",
					"1\nPOST\nhttps://worker.test/echo?q=%252F\n/echo\nquery:\nq=%2F\nheaders:\ncontent-type=text/plain\nbody:\nhéllo 世界 🌍\n",
					"1\nPOST\nhttps://worker.test/empty\n/empty\nquery:\nheaders:\ncontent-type=text/plain;charset=UTF-8\nbody:\n",
				}
				for index, result := range results {
					if result.Status != 200 || result.Body != bodies[index] {
						t.Fatalf("echo %d: %#v\nwant body %q", index, result, bodies[index])
					}
					if header(result, "x-order") != "first, second" || header(result, "x-last") != "last" {
						t.Fatalf("header append order: %v", result.Headers)
					}
				}
			case "routing":
				if len(results) != 3 || results[0].Status != 404 || results[0].Body != "not found" || results[1].Status != 405 || header(results[1], "allow") != "GET" || results[2].Status != 200 || results[2].Body != "hello" {
					t.Fatalf("routing: %v", results)
				}
			case "panic":
				if len(results) != 4 || results[0].Status != 500 || results[0].Body != "internal error" || results[1].Status != 200 || results[1].Body != "7:228,184,150,240,159,140,141" || results[2].Status != 500 || results[3].Status != 200 || results[3].Body != "0:" {
					t.Fatalf("panic and recovery: %v", results)
				}
				if stderr != "route exploded\nRangeError: utf8At index -1 is not a byte of a text of 0 bytes\n" {
					t.Fatalf("panic stderr: %q", stderr)
				}
			case "decode_json":
				want := []string{"new:2:-", "one:0:hello", "at $: missing field count", "at $.name: expected string, found number", "invalid JSON at line 1 column 15: expected object key"}
				if len(results) != len(want) {
					t.Fatalf("decode response count: %d", len(results))
				}
				for i, body := range want {
					status := 400
					if i < 2 {
						status = 200
					}
					if results[i].Body != body || results[i].Status != status {
						t.Fatalf("decode %d: %v", i, results[i])
					}
				}
			case "compute":
				if len(results) != 2 || results[0].Status != 200 || !strings.HasPrefix(results[0].Body, "/compute:") || !strings.HasPrefix(results[1].Body, "/compute%2Fagain:") {
					t.Fatalf("compute: %v", results)
				}
			}
			t.Logf("source and compiled responses identical (%d requests)", len(results))
		})
	}
}

func TestSignature(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, source string
		accepted     bool
	}{
		{"valid", "import type { HttpRequest, HttpResponse } from 'adamic/http'; export function handle(request: HttpRequest): HttpResponse { return { status: 200, headers: [], body: request.path }; }", true},
		{"aliases", "import type { HttpRequest as Request, HttpResponse as Response } from 'adamic/http'; function route(request: Request): Response { return { status: 200, headers: [], body: request.path }; } export { route as handle };", true},
		{"string", "import type { HttpRequest } from 'adamic/http'; export function handle(request: HttpRequest): string { return request.path; }", false},
		{"missing", "export function other(): string { return ''; }", false},
		{"unexported", "import type { HttpRequest, HttpResponse } from 'adamic/http'; function handle(request: HttpRequest): HttpResponse { return { status: 200, headers: [], body: '' }; }", false},
		{"lookalike", "import type { HttpResponse } from 'adamic/http'; interface HttpRequest { readonly path: string } export function handle(request: HttpRequest): HttpResponse { return { status: 200, headers: [], body: request.path }; }", false},
		{"response lookalike", "import type { HttpRequest } from 'adamic/http'; interface HttpResponse { readonly status: number; readonly headers: readonly { readonly name: string; readonly value: string }[]; readonly body: string } export function handle(request: HttpRequest): HttpResponse { return { status: 200, headers: [], body: request.path }; }", false},
		{"optional", "import type { HttpRequest, HttpResponse } from 'adamic/http'; export function handle(request?: HttpRequest): HttpResponse { return { status: 200, headers: [], body: '' }; }", false},
		{"extra", "import type { HttpRequest, HttpResponse } from 'adamic/http'; export function handle(request: HttpRequest, other: string): HttpResponse { return { status: 200, headers: [], body: other }; }", false},
		{"generic", "import type { HttpRequest, HttpResponse } from 'adamic/http'; export function handle<T>(request: HttpRequest): HttpResponse { return { status: 200, headers: [], body: '' }; }", false},
		{"async", "import type { HttpRequest, HttpResponse } from 'adamic/http'; export async function handle(request: HttpRequest): Promise<HttpResponse> { return { status: 200, headers: [], body: '' }; }", false},
	}
	for _, fixture := range cases {
		t.Run(fixture.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "entry.a")
			write(t, path, fixture.source)
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			err = CheckSignature(context.Background(), program)
			if (err == nil) != fixture.accepted {
				t.Fatalf("accepted=%t: %v", fixture.accepted, err)
			}
			if err != nil && (!strings.Contains(err.Error(), "found ") || !strings.Contains(err.Error(), "fix:")) {
				t.Fatalf("unhelpful refusal: %v", err)
			}
		})
	}
}

func TestRefusalPreservesOutput(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	write(t, filepath.Join(directory, "handler.mjs"), "sentinel")
	if err := Build("testdata/wrong.a", directory); err == nil {
		t.Fatal("accepted string handler")
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 1 {
		t.Fatalf("refusal changed directory: %v %v", entries, err)
	}
	contents, _ := os.ReadFile(filepath.Join(directory, "handler.mjs"))
	if string(contents) != "sentinel" {
		t.Fatalf("refusal overwrote handler: %s", contents)
	}
}

func TestRuntime(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if err := Build("testdata/panic.a", directory); err != nil {
		t.Fatal(err)
	}
	runtime := filepath.Join(directory, "adamic.mjs")
	script := `
import assert from 'node:assert/strict';
const oldBuffer = Buffer;
const runtime = await import(process.argv[1]);
const texts = ['', 'ascii', 'héllo 世界 🌍', '\ud800', '\udfff', 'a\ud800b', '\ud800\udfff', ''];
// Buffer is the oracle's independent UTF-8 witness.
for (const text of texts) {
 const bytes = oldBuffer.from(text, 'utf8');
 assert.equal(runtime.utf8Length(text), bytes.length);
 for (let index = 0; index < bytes.length; index++) assert.equal(runtime.utf8At(text, index), bytes[index]);
 for (const index of [-1, bytes.length, 0.5, NaN, Infinity, -Infinity]) {
  assert.throws(() => runtime.utf8At(text, index), error => error instanceof runtime.AdamicPanic && error.message === ` + "`RangeError: utf8At index ${index} is not a byte of a text of ${bytes.length} bytes`" + `);
 }
}
for (const name of ['readTextFile', 'writeTextFile', 'readDirectory', 'fileStatus', 'programArguments']) {
 assert.throws(() => runtime[name](), error => error instanceof runtime.AdamicPanic && error.message === name + ': not available on Workers');
}
globalThis.Buffer = undefined;
globalThis.process = undefined;
assert.equal(runtime.utf8Length('世🌍'), 7);
assert.throws(() => runtime.panic('still alive'), runtime.AdamicPanic);
assert.equal(runtime.utf8At('abc', 0), 97);
`
	command := exec.Command("node", "--input-type=module", "-e", script, runtime)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("runtime: %v\n%s", err, output)
	}
	for _, name := range []string{"handler.mjs", "adamic.mjs", "worker.mjs"} {
		source, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"node:", "process.", "Buffer.", "eval(", "new Function("} {
			if bytes.Contains(source, []byte(forbidden)) {
				t.Fatalf("%s uses Workers-unavailable %s", name, forbidden)
			}
		}
	}
}

func TestBridgeRethrows(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	write(t, filepath.Join(directory, "worker.mjs"), Bridge("./handler.mjs"))
	write(t, filepath.Join(directory, "adamic.mjs"), "export class AdamicPanic extends Error {}")
	write(t, filepath.Join(directory, "handler.mjs"), "export const bug = new Error('compiler bug'); export function handle() { throw bug; }")
	script := `import assert from 'node:assert/strict'; const {default: worker} = await import(process.argv[1]); const {bug} = await import(process.argv[2]); await assert.rejects(worker.fetch(new Request('https://worker.test/')), error => error === bug);`
	command := exec.Command("node", "--input-type=module", "-e", script, filepath.Join(directory, "worker.mjs"), filepath.Join(directory, "handler.mjs"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("bridge bug propagation: %v\n%s", err, output)
	}
}

func TestEntryExports(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	write(t, filepath.Join(directory, "dependency.a"), `export function dependency(): string { return 'dependency'; }`)
	entry := filepath.Join(directory, "entry.a")
	write(t, entry, `
import { dependency } from './dependency.a';
import type { HttpRequest, HttpResponse } from 'adamic/http';
let starts = 0;
starts += 1;
export function handle(request: HttpRequest): HttpResponse { return { status: 200, headers: [], body: dependency() + request.path }; }
export function renamed(): string { return dependency(); }
export interface Hidden { readonly value: string }
export function café(): string { return 'unicode'; }
export const arrow = (value: number): number => value + starts;
`)
	output := filepath.Join(directory, "output")
	if err := Build(entry, output); err != nil {
		t.Fatal(err)
	}
	script := `import assert from 'node:assert/strict'; const module = await import(process.argv[1]); const again = await import(process.argv[1]); assert.equal(module, again); assert.deepEqual(Object.keys(module).sort(), ['arrow', 'café', 'handle', 'renamed']); assert.equal(module.arrow(4), 5); assert.equal(module.café(), 'unicode'); assert.equal(module.renamed(), 'dependency'); assert.equal(module.handle({path: '/path'}).body, 'dependency/path');`
	command := exec.Command("node", "--input-type=module", "-e", script, filepath.Join(output, "handler.mjs"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("entry exports: %v\n%s", err, output)
	}
}

func TestGenericExportRefused(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "entry.a")
	write(t, entry, `import type { HttpRequest, HttpResponse } from 'adamic/http';
export function handle(request: HttpRequest): HttpResponse { return {status: 200, headers: [], body: ''}; }
export function identity<T>(value: T): T { return value; }`)
	output := filepath.Join(directory, "output")
	err := Build(entry, output)
	if err == nil || !strings.Contains(err.Error(), "exported generic function identity") {
		t.Fatalf("generic export was silently lost: %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("refusal wrote output: %v", err)
	}
}
