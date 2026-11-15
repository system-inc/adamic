package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestWorkerEnvironmentContract(t *testing.T) {
	t.Parallel()
	const prefix = "import type {HttpRequest,HttpResponse,HttpField} from 'adamic/http';\n"
	const fix = "; found an unsupported environment contract\n  fix: use a program-declared interface with required readonly string properties"
	cases := []struct{ name, decl, typ, field, reason string }{
		{"valid", "interface Environment { readonly ZONE: string; readonly GREETING: string }", "Environment", "", ""},
		{"empty", "interface Environment {}", "Environment", "", ""},
		{"own-alias", "interface Own { readonly ZONE: string } type Environment = Own;", "Environment", "", ""},
		{"merged", "interface Environment { readonly ZONE: string } interface Environment { readonly GREETING: string }", "Environment", "", ""},
		{"inherited", "interface Base { readonly ZONE: string } interface Environment extends Base { readonly GREETING: string }", "Environment", "", ""},
		{"mutable", "interface Environment { BAD: string }", "Environment", "BAD", "must be readonly"},
		{"number", "interface Environment { readonly BAD: number }", "Environment", "BAD", "must have type exactly string"},
		{"undefined", "interface Environment { readonly BAD: string | undefined }", "Environment", "BAD", "must have type exactly string"},
		{"object", "interface Environment { readonly BAD: {value: string} }", "Environment", "BAD", "must have type exactly string"},
		{"literal-string", "interface Environment { readonly BAD: 'fixed' }", "Environment", "BAD", "must have type exactly string"},
		{"optional", "interface Environment { readonly BAD?: string }", "Environment", "BAD", "must be required, not optional"},
		{"index", "interface Environment { readonly [key: string]: string }", "Environment", "[index]", "index signatures are not supported"},
		{"method", "interface Environment { BAD(): string }", "Environment", "BAD", "must be a data property, not a method or accessor"},
		{"any-field", "interface Environment { readonly BAD: any }", "Environment", "BAD", "must have type exactly string"},
		{"unknown-field", "interface Environment { readonly BAD: unknown }", "Environment", "BAD", "must have type exactly string"},
		{"any", "", "any", "environment", "type must be a program-declared interface"},
		{"unknown", "", "unknown", "environment", "type must be a program-declared interface"},
		{"inline", "", "{readonly ZONE: string}", "environment", "type must be a program-declared interface"},
		{"generic-interface", "interface Environment<T> { readonly ZONE: T }", "Environment<string>", "environment", "generic interfaces are not supported"},
		{"generic-alias", "interface Own {readonly ZONE:string} type Environment<T> = Own;", "Environment<string>", "environment", "generic interfaces are not supported"},
		{"prelude-alias", "type Environment = HttpField;", "Environment", "environment", "type must be a program-declared interface"},
		{"library-alias", "type Environment = String;", "Environment", "environment", "type must be a program-declared interface"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			for _, async := range []bool{false, true} {
				keyword, result := "", "HttpResponse"
				if async {
					keyword, result = "async ", "Promise<HttpResponse>"
				}
				path := filepath.Join(t.TempDir(), "entry.a")
				write(t, path, prefix+c.decl+"\nexport "+keyword+"function handle(request: HttpRequest, environment: "+c.typ+"): "+result+" { return {status:200,headers:[],body:request.path}; }")
				program, err := load.Load([]string{path})
				if err != nil {
					t.Fatal(err)
				}
				_, err = checkedWorkerEnvironment(context.Background(), program)
				if (err == nil) != (c.reason == "") {
					t.Fatalf("environment contract accepted=%t: %v", c.reason == "", err)
				}
				if c.reason != "" {
					want := fmt.Sprintf("worker: refused environment field %q: %s", c.field, c.reason) + fix
					if err.Error() != want {
						t.Fatalf("environment diagnostic: got %v; want %s", err, want)
					}
				}
				t.Logf("async=%t: %v", async, err)
			}
		})
	}
}

func TestWorkerEnvironmentParameterExclusions(t *testing.T) {
	t.Parallel()
	const prefix = "import type {HttpRequest,HttpResponse} from 'adamic/http'; interface Environment {readonly ZONE:string}\n"
	for _, source := range []string{
		"export function handle(request:HttpRequest, environment:Environment, third:string):HttpResponse {return {status:200,headers:[],body:request.path}}",
		"export function handle<T>(request:HttpRequest, environment:Environment):HttpResponse {return {status:200,headers:[],body:request.path}}",
		"export function handle(request:HttpRequest, environment?:Environment):HttpResponse {return {status:200,headers:[],body:request.path}}",
		"export function handle(request:HttpRequest, environment:Environment={ZONE:'x'}):HttpResponse {return {status:200,headers:[],body:request.path}}",
		"export function handle(request:HttpRequest, ...environment:Environment[]):HttpResponse {return {status:200,headers:[],body:request.path}}",
	} {
		p := filepath.Join(t.TempDir(), "entry.a")
		write(t, p, prefix+source)
		program, err := load.Load([]string{p})
		if err != nil {
			t.Fatal(err)
		}
		if err = CheckSignature(context.Background(), program); err == nil || !strings.Contains(err.Error(), "environment: Environment") {
			t.Fatalf("environment parameter exclusion: %v", err)
		}
		t.Log(err)
	}
}

func TestWorkerEnvironmentFieldOrder(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "entry.a")
	write(t, path, "import type {HttpRequest,HttpResponse} from 'adamic/http'; interface Environment {readonly ZONE:string;readonly GREETING:string} export function handle(request:HttpRequest, environment:Environment):HttpResponse{return {status:200,headers:[],body:request.path}}")
	p, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	fields, err := checkedWorkerEnvironment(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(fields, []string{"ZONE", "GREETING"}) {
		t.Fatalf("environment declaration order: %v", fields)
	}
	bridge, err := environmentBridge("./handler.mjs", fields, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(bridge, `const environmentFields = ["ZONE", "GREETING"];`) {
		t.Fatal("environment names not emitted in declaration order")
	}
}

func environmentBridgeNode(t *testing.T, async bool, names []string, handler, script string) {
	t.Helper()
	bridge, err := environmentBridge("./handler.mjs", names, async)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	write(t, filepath.Join(dir, "worker.mjs"), bridge)
	write(t, filepath.Join(dir, "adamic.mjs"), "export class AdamicPanic extends Error {}\n")
	write(t, filepath.Join(dir, "handler.mjs"), handler)
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script, filepath.Join(dir, "worker.mjs"))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("environment bridge: %v\n%s", err, out)
	}
}
func TestEnvironmentBridgeProjectsDeclaredNames(t *testing.T) {
	t.Parallel()
	for _, async := range []bool{false, true} {
		environmentBridgeNode(t, async, []string{"ZONE", "GREETING", "__proto__"}, `let previous;export function handle(input,environment){if(previous===environment)throw new Error('environment reused');previous=environment;return {status:200,headers:[],body:JSON.stringify({keys:Object.keys(environment),ordinary:Object.getPrototypeOf(environment)===Object.prototype,proto:environment.__proto__})};}`, `import assert from 'node:assert/strict';const {default:worker}=await import(process.argv[1]);const env={GREETING:'hi',EXTRA:'SECRET_BINDING_f1d9',ZONE:'z'};Object.defineProperty(env,'__proto__',{value:'safe',enumerable:true});for(let i=0;i<2;i++){const response=await worker.fetch(new Request('https://worker.test/'),env);assert.equal(response.status,200);assert.deepEqual(JSON.parse(await response.text()),{keys:['ZONE','GREETING','__proto__'],ordinary:true,proto:'safe'});}`)
	}
}
func TestEnvironmentBridgeRejectsBindingsPrivately(t *testing.T) {
	t.Parallel()
	for _, async := range []bool{false, true} {
		environmentBridgeNode(t, async, []string{"ZONE", "GREETING"}, `export let calls=0;export function handle(){calls++;return {status:200,headers:[],body:'handler reached'}}`, `import assert from 'node:assert/strict';const {default:worker}=await import(process.argv[1]);const handler=await import(new URL('./handler.mjs','file://'+process.argv[1]));const secret='SECRET_BINDING_72ad';for(const invalid of [undefined,827451,{token:secret}]){const logs=[];console.error=x=>logs.push(x);const env={ZONE:secret,EXTRA:secret};if(invalid!==undefined)env.GREETING=invalid;const response=await worker.fetch(new Request('https://worker.test/'),env);const body=await response.text();assert.equal(response.status,500);assert.equal(body,'internal error');assert.deepEqual(logs,['environment binding "GREETING" is missing or not a string']);assert.equal(handler.calls,0);assert(!body.includes(secret));assert(!logs.join('\n').includes(secret));}`)
	}
}

func TestOneParameterAsyncBridgeTextIsUnchanged(t *testing.T) {
	t.Parallel()
	before, err := os.ReadFile("testdata/sync-bridge.mjs")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(string(before), "output = handle(input);", "output = await handle(input);", 1)
	got, err := asyncBridge("./handler.mjs")
	if err != nil || got != want {
		t.Fatalf("one-parameter async bridge bytes changed: %v", err)
	}
}

func TestInitializeWasmBridgeRequiresExactlyOneExport(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, text string
		count      int
	}{
		{"missing", "// missing the export", 0},
		{"duplicate", "export default { }\nexport default { }", 2},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			text, err := initializeWasmBridge(c.text)
			want := fmt.Sprintf("worker: refused Wasm bridge: expected exactly one %q, found %d; fix: restore the generated bridge export", "export default {", c.count)
			if err == nil || err.Error() != want || text != "" {
				t.Fatalf("Wasm bridge needle must fail closed: got %q, %v; want %s", text, err, want)
			}
		})
	}
	if text, err := wasmBridge("./handler.mjs"); err != nil || strings.Count(text, "initializeCrossing();") != 1 {
		t.Fatalf("valid Wasm bridge: %q, %v", text, err)
	}
}
