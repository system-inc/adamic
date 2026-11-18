package worker

import (
	"os/exec"
	"path/filepath"
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
}
