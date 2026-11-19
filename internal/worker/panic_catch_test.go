package worker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

// panicCatchPrelude imports a built worker.mjs and defines get, which answers one request, and quiet,
// which runs requests with console.error held in logs, where the bridge writes a panic's message.
const panicCatchPrelude = `import assert from 'node:assert/strict';
const {default:worker}=await import(process.argv[1]);
const logs=[];
const get=async(path)=>{const response=await worker.fetch(new Request('https://worker.test'+path));return {status:response.status,body:await response.text()};};
const quiet=async(work)=>{const error=console.error;console.error=(message)=>logs.push(String(message));try{return await work();}finally{console.error=error;}};
`

// A panic, and every check the compiler inserts, is never caught (docs/rulings.md). A Worker's panic
// throws an AdamicPanic for the bridge to answer 500, so each case here builds a fixture under
// testdata/panic_catch and runs its worker.mjs in Node: no catch of the program's takes a panic, and
// no finally of the program's runs while one unwinds. Each assertion's message is the witness a
// mutant's check reads (panic-catch.mutants.md).
func TestWorkerPanicIsNeverCaught(t *testing.T) {
	t.Parallel()
	for _, pin := range []struct{ name, script string }{
		{"catch", `await quiet(async()=>{
const panicked=await get('/panic');
assert.equal(panicked.body,'internal error','catch: a panic reached the program catch');
assert.equal(panicked.status,500,'catch: a panic reached the program catch');
const bounds=await get('/bounds');
assert.equal(bounds.body,'internal error','catch: an inserted check reached the program catch');
assert.equal(bounds.status,500,'catch: an inserted check reached the program catch');
const thrown=await get('/throw');
assert.deepEqual(thrown,{status:200,body:'caught'},'catch: an ordinary error missed the program catch');
assert.deepEqual(await get('/ok'),{status:200,body:'no panic'});
});
assert.equal(logs.length,2,'catch: the bridge logged '+JSON.stringify(logs));
assert.equal(logs[0],'panic inside try');`},
		{"finally", `await quiet(async()=>{
assert.deepEqual(await get('/ok'),{status:200,body:'no panic'});
assert.deepEqual(await get('/count'),{status:200,body:'1'},'finally: a finally missed an ordinary return');
const panicked=await get('/panic');
assert.deepEqual(panicked,{status:500,body:'internal error'},'finally: a panic under a finally was not the bridge 500');
assert.deepEqual(await get('/count'),{status:200,body:'1'},'finally: a finally ran while a panic unwound');
});
assert.deepEqual(logs,['panic inside try with a finally']);`},
		{"catch_finally", `await quiet(async()=>{
assert.deepEqual(await get('/ok'),{status:200,body:'caught'});
assert.deepEqual(await get('/count'),{status:200,body:'1'},'catch_finally: a finally missed a return from the catch');
const panicked=await get('/panic');
assert.deepEqual(panicked,{status:500,body:'internal error'},'catch_finally: a panic in the catch was not the bridge 500');
assert.deepEqual(await get('/count'),{status:200,body:'1'},'catch_finally: a finally ran while a panic raised in the catch unwound');
});
assert.deepEqual(logs,['panic inside a catch with a finally']);`},
		{"finally_return", `await quiet(async()=>{
assert.deepEqual(await get('/ok'),{status:200,body:'from the finally'},'finally_return: the finally return did not override the try');
const panicked=await get('/panic');
assert.deepEqual(panicked,{status:500,body:'internal error'},'finally_return: a finally return swallowed a panic');
});
assert.deepEqual(logs,['panic under a finally that returns']);`},
		{"interleaved", `const paths=['/a','/b','/bx','/bxx','/bxxx','/bxxxx','/bxxxxx','/bxxxxxx','/bxxxxxxx'];
const responses=await quiet(()=>Promise.all(paths.map(get)));
assert.deepEqual(responses[0],{status:500,body:'internal error'},'interleaved: the panic of /a was caught');
for(let index=1;index<paths.length;index++) assert.deepEqual(responses[index],{status:200,body:'b-caught'},'interleaved: '+paths[index]+' lost its own catch');
assert.deepEqual(await get('/count'),{status:200,body:String(paths.length-1)},'interleaved: a finally ran for /a or was skipped for a /b');
assert.deepEqual(logs,['request a panicked']);`},
	} {
		t.Run(pin.name, func(t *testing.T) {
			t.Parallel()
			out := filepath.Join(t.TempDir(), "worker")
			if err := Build(filepath.Join("testdata/panic_catch", pin.name+".a"), out); err != nil {
				t.Fatalf("panic pin build: %v", err)
			}
			command := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", panicCatchPrelude+pin.script, filepath.Join(out, "worker.mjs"))
			if report, err := command.CombinedOutput(); err != nil {
				t.Fatalf("panic pin %s: %v\n%s", pin.name, err, report)
			}
		})
	}
	t.Run("wasm_trap", func(t *testing.T) {
		t.Parallel()
		wasmTrapPin(t)
	})
}

// wasmTrapHook runs before panicCatchPrelude imports the Worker. It loads .wasm the way Workers do
// (wasm-node.mjs), and wraps the crossing's invoke so that while wasmTrapNext is set, the next call of
// an export first calls a one-instruction module that executes unreachable: the trap is the engine's
// own `WebAssembly.RuntimeError`, thrown inside invoke where a trap in the export would be.
const wasmTrapHook = `import { registerHooks } from 'node:module';
import { readFileSync } from 'node:fs';
import { fileURLToPath, pathToFileURL } from 'node:url';
const {registerWasm}=await import(pathToFileURL(process.argv[2]));
registerWasm();
// A module of one function, exported as trap, whose body is unreachable.
const trap=new WebAssembly.Instance(new WebAssembly.Module(new Uint8Array([0x00,0x61,0x73,0x6d,0x01,0x00,0x00,0x00,0x01,0x04,0x01,0x60,0x00,0x00,0x03,0x02,0x01,0x00,0x07,0x08,0x01,0x04,0x74,0x72,0x61,0x70,0x00,0x00,0x0a,0x05,0x01,0x03,0x00,0x00,0x0b]))).exports.trap;
globalThis.wasmTrapNext=false;
globalThis.wasmTrap=()=>{if(globalThis.wasmTrapNext){globalThis.wasmTrapNext=false;trap();}};
registerHooks({load(url,context,nextLoad){
if(!url.endsWith('/crossing-runtime.mjs')) return nextLoad(url,context);
const source=readFileSync(fileURLToPath(url),'utf8');
const needle='try { return this.api[name](...args); }';
if(source.split(needle).length!==2) throw new Error('wasm_trap: crossing-runtime.mjs has no single invoke to wrap');
return {format:'module',source:source.replace(needle,"try { if (name.startsWith('adamic_export_')) globalThis.wasmTrap(); return this.api[name](...args); }"),shortCircuit:true};
}});
`

// A trap in a --wasm function is a panic (wasm.go's wrapper), so it's never the program's catch's either.
// Adamic's runtime turns every trap it can see into a panic first, so no .a program here traps; the hook
// raises a real engine trap at the crossing instead. The control asks the same Worker before and after.
// The module is the transform Worker's, built ahead unmutated: a mutant of wasm.go changes only the
// wasm-state.mjs this BuildWith writes, never the module, so the product is found without the mutation.
func wasmTrapPin(t *testing.T) {
	transform := workerWasmProduct(t, "transform", nil, false)
	out := filepath.Join(t.TempDir(), "worker")
	err := BuildWith("testdata/panic_catch/wasm_trap.a", out, Options{Wasm: []string{"transform"}, BuildWasm: func(_, module, abi string, names []string) error {
		if !slices.Equal(names, []string{"transform"}) {
			return fmt.Errorf("wasm_trap: asked to build %v, not transform", names)
		}
		for target, name := range map[string]string{module: "handler.wasm", abi: "abi.json"} {
			contents, err := os.ReadFile(filepath.Join(transform, name))
			if err != nil {
				return err
			}
			if err := os.WriteFile(target, contents, 0644); err != nil {
				return err
			}
		}
		return nil
	}})
	if err != nil {
		t.Fatalf("wasm_trap pin build: %v", err)
	}
	wasmNode, err := filepath.Abs("wasm-node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	script := wasmTrapHook + panicCatchPrelude + `await quiet(async()=>{
assert.deepEqual(await get('/ok'),{status:200,body:'>ko/'},'wasm_trap: the Wasm call failed before any trap');
globalThis.wasmTrapNext=true;
const trapped=await get('/trap');
assert.equal(trapped.body,'internal error','wasm_trap: a Wasm trap reached the program catch');
assert.equal(trapped.status,500,'wasm_trap: a Wasm trap reached the program catch');
assert.equal(globalThis.wasmTrapNext,false,'wasm_trap: the hook never trapped');
assert.deepEqual(await get('/ok'),{status:200,body:'>ko/'},'wasm_trap: the crossing did not answer after a trap');
});
assert.equal(logs.length,1,'wasm_trap: the bridge logged '+JSON.stringify(logs));
assert.match(logs[0],/^Wasm trap: /,'wasm_trap: the bridge log did not name the trap');`
	command := exec.CommandContext(t.Context(), "node", "--disable-warning=ExperimentalWarning", "--input-type=module", "-e", script, filepath.Join(out, "worker.mjs"), wasmNode)
	if report, err := command.CombinedOutput(); err != nil {
		t.Fatalf("panic pin wasm_trap: %v\n%s", err, report)
	}
}
