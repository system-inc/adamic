package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestSynchronousBridgeTextIsUnchanged(t *testing.T) {
	t.Parallel()
	want, err := os.ReadFile("testdata/sync-bridge.mjs")
	if err != nil {
		t.Fatal(err)
	}
	if got := Bridge("./handler.mjs"); got != string(want) {
		t.Fatal("synchronous bridge bytes changed")
	}
}

func TestWorkerAsyncContract(t *testing.T) {
	t.Parallel()
	const prefix = "import type { HttpRequest, HttpResponse } from 'adamic/http';\n"
	const response = "{ status: 200, headers: [], body: request.path }"
	const shape = "interface Reply { readonly status: number; readonly headers: readonly { readonly name: string; readonly value: string }[]; readonly body: string }\n"
	cases := []struct {
		name, source, checker string
		accepted              bool
	}{
		{"async", "export async function handle(request: HttpRequest): Promise<HttpResponse> { return " + response + "; }", "", true},
		{"async-alias", "async function route(request: HttpRequest): Promise<HttpResponse> { return " + response + "; } export { route as handle };", "", true},
		{"non-async-promise", "export function handle(request: HttpRequest): Promise<HttpResponse> { return Promise.resolve(" + response + "); }", "", false},
		{"structural-response", shape + "export async function handle(request: HttpRequest): Promise<Reply> { return " + response + "; }", "", false},
		{"structural-request", "interface Request { readonly path: string } export async function handle(request: Request): Promise<HttpResponse> { return " + response + "; }", "", false},
		{"async-arrow", "export const handle = async (request: HttpRequest): Promise<HttpResponse> => (" + response + ");", "", false},
		{"outer-union", "export async function handle(request: HttpRequest): Promise<HttpResponse> | Promise<string> { return " + response + "; }", "TS1064", false},
		{"undefined-payload", "export async function handle(request: HttpRequest): Promise<HttpResponse | undefined> { return " + response + "; }", "", false},
		{"null-payload", "export async function handle(request: HttpRequest): Promise<HttpResponse | null> { return " + response + "; }", "", false},
		{"union-payload", "export async function handle(request: HttpRequest): Promise<HttpResponse | string> { return " + response + "; }", "", false},
		{"promise-like", "export async function handle(request: HttpRequest): PromiseLike<HttpResponse> { return " + response + "; }", "TS1064", false},
		{"thenable", "interface Thenable { then(callback: (value: HttpResponse) => void): void } export async function handle(request: HttpRequest): Thenable { return " + response + "; }", "TS1064", false},
		{"optional", "export async function handle(request?: HttpRequest): Promise<HttpResponse> { return { status: 200, headers: [], body: '' }; }", "", false},
		{"extra", "export async function handle(request: HttpRequest, environment: string): Promise<HttpResponse> { return " + response + "; }", "", false},
		{"rest", "export async function handle(...requests: HttpRequest[]): Promise<HttpResponse> { return { status: 200, headers: [], body: '' }; }", "", false},
		{"default", "export async function handle(request: HttpRequest = {method:'GET',url:'',path:'',query:[],headers:[],body:''}): Promise<HttpResponse> { return " + response + "; }", "", false},
		{"generic", "export async function handle<T>(request: HttpRequest): Promise<HttpResponse> { return " + response + "; }", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "entry.a")
			write(t, path, prefix+c.source)
			program, err := load.Load([]string{path})
			if c.checker != "" {
				if err == nil || !strings.Contains(err.Error(), c.checker) {
					t.Fatalf("want checker %s, got %v", c.checker, err)
				}
				t.Log(err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			err = CheckSignature(context.Background(), program)
			if (err == nil) != c.accepted {
				t.Fatalf("async contract accepted=%t: %v", c.accepted, err)
			}
			if err != nil && (!strings.Contains(err.Error(), "function handle(request: HttpRequest): HttpResponse or async function handle(request: HttpRequest): Promise<HttpResponse>") || !strings.Contains(err.Error(), "\n  fix: import type { HttpRequest, HttpResponse } from 'adamic/http'; export a synchronous handle returning HttpResponse, or an async handle returning Promise<HttpResponse>")) {
				t.Fatalf("async contract diagnostic: %v", err)
			}
			t.Logf("accepted=%t: %v", c.accepted, err)
		})
	}
}

func TestWorkerAsyncWasmRefusals(t *testing.T) {
	t.Parallel()
	const prefix = "import type { HttpRequest, HttpResponse } from 'adamic/http';\n"
	const sync = "export function handle(request: HttpRequest): HttpResponse { return {status:200,headers:[],body:request.path}; }\n"
	const async = "export async function handle(request: HttpRequest): Promise<HttpResponse> { await Promise.resolve(); return {status:200,headers:[],body:request.path}; }\n"
	const helper = "async function later(): Promise<number> { return 1; } function compute(value: number): number { return value + 1; }\n"
	for _, c := range []struct {
		name, source string
		names        []string
		want         string
	}{
		{"async-handle", async, []string{"handle"}, "worker: refused --wasm with async handle; fix: omit --wasm for an async handle"},
		{"async-handle-selected-helper", async + helper, []string{"compute"}, "worker: refused --wasm with async handle; fix: omit --wasm for an async handle"},
		{"async-handle-generic-export", async + "export function other<T>(value: T): T { return value; }", []string{"other"}, "worker: refused --wasm with async handle; fix: omit --wasm for an async handle"},
		{"async-handle-missing-name", async, []string{"missing"}, "worker: refused --wasm with async handle; fix: omit --wasm for an async handle"},
		{"async-selected", sync + helper, []string{"later"}, "worker: refused Wasm function \"later\": async functions cannot cross the Wasm boundary; fix: keep it in JavaScript and omit it from --wasm"},
	} {
		root := t.TempDir()
		entry := filepath.Join(root, "entry.a")
		write(t, entry, prefix+c.source)
		out := filepath.Join(root, "output")
		builds := 0
		err := BuildWith(entry, out, Options{Wasm: c.names, BuildWasm: func(string, string, string, []string) error { builds++; return fmt.Errorf("unexpected Wasm build") }})
		if err == nil || err.Error() != c.want || builds != 0 {
			t.Fatalf("async Wasm refusal %s: got %v, builds %d; want %s", c.name, err, builds, c.want)
		}
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Fatalf("async Wasm refusal wrote output: %v", err)
		}
	}
}

func TestWorkerAsyncRejectionForms(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ name, source, want string }{
		{"throw-error", `async function f(): Promise<string> { throw new Error("x"); } console.log(await f());`, ""},
		{"reject-error", `async function f(): Promise<string> { return await Promise.reject(new Error("x")); } console.log(await f());`, ""},
		{"throw-string", `async function f(): Promise<string> { throw "x"; } console.log(await f());`, "throwing a string that isn't an Error"},
		{"reject-number", `async function f(): Promise<string> { return await Promise.reject(1); } console.log(await f());`, "Promise rejection with a non-Error"},
		{"panic-after-await", `import { panic } from 'adamic'; async function f(): Promise<string> { await Promise.resolve(); panic('after await'); } console.log(await f());`, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "entry.a")
			write(t, path, c.source)
			p, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), p)
			if c.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			stop, ok := err.(*lower.NotYet)
			if !ok || stop.What != c.want {
				t.Fatalf("rejection stop: got %v, want NotYet %q", err, c.want)
			}
		})
	}
}

func asyncBridgeNode(t *testing.T, handler, script string) {
	t.Helper()
	directory := t.TempDir()
	bridge, err := asyncBridge("./handler.mjs")
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(directory, "worker.mjs"), bridge)
	write(t, filepath.Join(directory, "adamic.mjs"), "export class AdamicPanic extends Error {}\n")
	write(t, filepath.Join(directory, "handler.mjs"), handler)
	command := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script, filepath.Join(directory, "worker.mjs"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("async bridge: %v\n%s", err, output)
	}
}
func TestAsyncBridgeAwaits(t *testing.T) {
	t.Parallel()
	asyncBridgeNode(t, `export async function handle(input) { await Promise.resolve(); return {status:201,headers:[],body:input.path}; }`, `import assert from 'node:assert/strict'; const {default: worker}=await import(process.argv[1]); const response=await worker.fetch(new Request('https://worker.test/ready')); assert.equal(response.status,201); assert.equal(await response.text(),'/ready');`)
}
func TestAsyncBridgeCatchesRejectedPanic(t *testing.T) {
	t.Parallel()
	asyncBridgeNode(t, `import {AdamicPanic} from './adamic.mjs'; async function helper(){return await Promise.reject(new AdamicPanic('helper rejected'));} export async function handle(){return await helper();}`, `import assert from 'node:assert/strict';const logs=[];console.error=x=>logs.push(x); const {default:worker}=await import(process.argv[1]);const response=await worker.fetch(new Request('https://worker.test/reject'));assert.equal(response.status,500);assert.equal(await response.text(),'internal error');assert.deepEqual(logs,['helper rejected']);`)
}

func TestAsyncBridgePreservesErrorRejection(t *testing.T) {
	t.Parallel()
	asyncBridgeNode(t, `export const failure = new Error('ordinary error'); export async function handle() { return await Promise.reject(failure); }`, `import assert from 'node:assert/strict';const {default: worker}=await import(process.argv[1]);const {failure}=await import(new URL('./handler.mjs', 'file://'+process.argv[1]));await assert.rejects(worker.fetch(new Request('https://worker.test/error')),error=>error===failure);`)
}

func TestWorkerAsyncWasmCommand(t *testing.T) {
	t.Parallel()
	compiler := workerCompilerProduct(t, nil, false)
	for _, c := range []struct{ source, names, want string }{
		{`import type {HttpRequest,HttpResponse} from 'adamic/http'; export async function handle(request: HttpRequest): Promise<HttpResponse> { return {status:200,headers:[],body:request.path}; }`, "missing", "worker: refused --wasm with async handle; fix: omit --wasm for an async handle"},
		{`import type {HttpRequest,HttpResponse} from 'adamic/http'; async function later(): Promise<number> { return 1; } export function handle(request: HttpRequest): HttpResponse { return {status:200,headers:[],body:request.path}; }`, "later", `worker: refused Wasm function "later": async functions cannot cross the Wasm boundary; fix: keep it in JavaScript and omit it from --wasm`},
	} {
		root := t.TempDir()
		entry := filepath.Join(root, "entry.a")
		output := filepath.Join(root, "worker")
		write(t, entry, c.source)
		command := exec.CommandContext(t.Context(), compiler, "worker", entry, "--out", output, "--wasm", c.names)
		report, err := command.CombinedOutput()
		if err == nil || string(report) != "adamic: "+c.want+"\n" {
			t.Fatalf("async Wasm command: %v\n%s", err, report)
		}
		if _, err := os.Stat(output); !os.IsNotExist(err) {
			t.Fatalf("async Wasm command wrote output: %v", err)
		}
	}
}

func TestAsyncBridgeRequiresExactlyOneCall(t *testing.T) {
	t.Parallel()
	const needle = "output = handle(input);"
	original := Bridge("./handler.mjs")
	for _, c := range []struct {
		name, text string
		count      int
	}{
		{"missing", strings.Replace(original, needle, "output = handle (input);", 1), 0},
		{"duplicate", original + needle, 2},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			text, err := awaitHandlerCall(c.text)
			want := fmt.Sprintf("worker: refused async bridge: expected exactly one %q, found %d; fix: restore the generated handler call", needle, c.count)
			if err == nil || err.Error() != want || text != "" {
				t.Fatalf("async bridge needle must fail closed: got %q, %v; want %s", text, err, want)
			}
		})
	}
}

func TestWorkerWasmRefusesAsyncArrow(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	entry, output := filepath.Join(root, "entry.a"), filepath.Join(root, "worker")
	write(t, entry, `import type {HttpRequest,HttpResponse} from 'adamic/http';
const later = async (value: number): Promise<number> => value;
export function handle(request: HttpRequest): HttpResponse { return {status:200,headers:[],body:request.path}; }`)
	// Prove this is a well-typed source, not a refusal from a malformed fixture.
	if _, err := load.Load([]string{entry}); err != nil {
		t.Fatal(err)
	}
	builds := 0
	err := BuildWith(entry, output, Options{Wasm: []string{"later"}, BuildWasm: func(string, string, string, []string) error { builds++; return fmt.Errorf("unexpected Wasm compile") }})
	if err == nil || builds != 0 {
		t.Fatalf("async arrow reached Wasm compile: builds=%d, err=%v", builds, err)
	}
	if _, stat := os.Stat(output); !os.IsNotExist(stat) {
		t.Fatalf("async arrow refusal wrote output: %v", stat)
	}
	t.Log(err)
}
