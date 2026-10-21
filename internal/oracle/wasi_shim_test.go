package oracle

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// The fixtures that read or write files, which the shim refuses by design (docs/wasm.md). They are
// left out by name rather than skipped, and their imports are still asserted, so the list cannot drift:
// TestWASIAgreesWithNode runs these fixtures under node:wasi.
var shimFileFixtures = map[string]bool{
	"internal/oracle/testdata/closure_convention_host24.a":  true,
	"internal/oracle/testdata/unknown_narrowing_host.a":     true,
	"internal/oracle/testdata/write_stdout_order.a":         true,
	"internal/oracle/testdata/write_stderr_order.a":         true,
	"internal/oracle/testdata/prompt_then_read.a":           true,
	"internal/oracle/testdata/node_fs_rm_options.a":         true,
	"internal/oracle/testdata/node_options_const_capture.a": true,
	"internal/oracle/testdata/node_options_widening.a":      true,
}

// TestWASIShimAgreesWithNode runs every fixture's wasm32 build through the shim, as a family (families_test.go):
// TestWASIShimAgreesWithNode_000 and on each run the fixtures their shard owns, since all of them in one test ran
// 579.6 s in Loom against its 60 s budget. Within a shard the fixtures run in parallel, as TestWASIAgreesWithNode's
// do, since one after another a shard of 15 ran 34 s on a loaded box, and each shard logs its own counts once they
// have all finished.
func runWASIShimShard(t *testing.T, shard int) {
	t.Helper()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	rows := familyCases(t, "TestWASIShimAgreesWithNode", shard, fixturePaths())
	if len(rows) == 0 {
		return
	}
	planted, probing := familyProbe("TestWASIShimAgreesWithNode")
	if err := native.ValidateOptions(native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	var ran, agreed, notYet, refused, excluded atomic.Int32
	// Cleanups run after the parallel subtests, so the counts are the shard's whole.
	t.Cleanup(func() {
		t.Logf("SHIM COUNTS shard=%d run=%d agreed=%d notyet=%d refused=%d excluded=%d", shard, ran.Load(), agreed.Load(), notYet.Load(), refused.Load(), excluded.Load())
	})
	for _, index := range rows {
		fixture := fixtures[index]
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			// Held to the refusal TestWASIAgreesWithNode holds them to (runtime_boundary_test.go).
			if runtimeBoundaryViewFixture(fixture.path) {
				assertRuntimeBoundaryViewRefusal(t, err)
				refused.Add(1)
				return
			}
			if !fixture.lowers {
				var refusal *lower.NotYet
				if !errors.As(err, &refusal) {
					t.Fatalf("want NotYet, got %v", err)
				}
				notYet.Add(1)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			// The witness comes from the program as lowered, before the probe's mutant can reach it.
			witness := func() run {
				if fixture.checked {
					return onJavaScriptBackend(t, program)
				}
				return onNode(t, path)
			}
			var expected run
			mutant := probing && fixture.path == planted
			if mutant {
				// TestFamiliesCatchAPlantedCase's mutant: one byte more in the program the shim and node:wasi run.
				expected = witness()
				program.Strings[0] += "!"
			}
			binary := filepath.Join(t.TempDir(), "program.wasm")
			if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
				t.Fatal(err)
			}
			inventory := execute(t, "node", "-e", `const fs=require('node:fs'); const m=new WebAssembly.Module(fs.readFileSync(process.argv[1])); console.log(JSON.stringify(WebAssembly.Module.imports(m).filter(i=>i.module==='wasi_snapshot_preview1').map(i=>i.name).sort()));`, binary)
			if inventory.exitCode != 0 {
				t.Fatalf("inventory: %s", inventory.stderr)
			}
			var imports []string
			if err := json.Unmarshal(inventory.stdout, &imports); err != nil {
				t.Fatal(err)
			}
			t.Logf("IMPORTS %s | %s", fixture.path, strings.Join(imports, ", "))
			fileAccess := ""
			for _, name := range imports {
				if strings.HasPrefix(name, "path_") || name == "fd_read" || name == "fd_readdir" || name == "fd_filestat_get" {
					fileAccess = name
					break
				}
			}
			if shimFileFixtures[fixture.path] {
				if fileAccess == "" {
					t.Fatalf("listed in shimFileFixtures but imports no file access: run it through the shim and drop it from the list")
				}
				excluded.Add(1)
				return
			}
			if fileAccess != "" {
				t.Fatalf("imports %s, file access the shim refuses: add it to shimFileFixtures with its reason in docs/wasm.md", fileAccess)
			}
			ran.Add(1)
			if os.Getenv("ADAMIC_SHIM_INVENTORY_ONLY") == "1" {
				return
			}
			reference := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "wasi.mjs"), binary)
			actual := execute(t, "node", filepath.Join(repository, "oracle", "wasi-shim.mjs"), binary)
			if difference := disagreement(reference, actual); difference != "" {
				t.Errorf("shim vs node:wasi: %s\nreference: exit %d stdout %q stderr %q\nshim: exit %d stdout %q stderr %q", difference, reference.exitCode, reference.stdout, reference.stderr, actual.exitCode, actual.stdout, actual.stderr)
			}
			if !mutant {
				expected = witness()
			}
			if difference := disagreement(expected, actual); difference != "" {
				t.Errorf("shim vs expected witness: %s\nwitness: exit %d stdout %q stderr %q\nshim: exit %d stdout %q stderr %q", difference, expected.exitCode, expected.stdout, expected.stderr, actual.exitCode, actual.stdout, actual.stderr)
			}
			if disagreement(reference, actual) == "" && disagreement(expected, actual) == "" {
				agreed.Add(1)
			}
		})
	}
}

// Comparing this checked fixture to source instead of the backend is a wrong-witness mutant.
// Not parallel: sets ADAMIC_GATE_UNCACHED for the whole process with t.Setenv.
func TestWASIShimCheckedWitnessControl(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	const fixturePath = "internal/oracle/testdata/writes_past_end.a"
	checked := false
	for _, fixture := range fixtures {
		if fixture.path == fixturePath {
			checked = fixture.checked
		}
	}
	if !checked {
		t.Fatal("control must remain a checked fixture")
	}
	path, err := filepath.Abs(filepath.Join(repository, fixturePath))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "program.wasm")
	if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	actual := execute(t, "node", filepath.Join(repository, "oracle", "wasi-shim.mjs"), binary)
	backend := onJavaScriptBackend(t, program)
	if difference := disagreement(backend, actual); difference != "" {
		t.Fatalf("checked backend control: %s", difference)
	}
	source := onNode(t, path)
	if difference := disagreement(source, actual); difference != "exit codes differ" || source.exitCode != 0 || actual.exitCode != 70 {
		t.Fatalf("wrong source witness must differ: source exit %d, shim exit %d, difference %q", source.exitCode, actual.exitCode, difference)
	}
	t.Log("wrong source witness caught: writes_past_end.a source exits 0, backend and shim exit 70")
}

func TestWASIShimContractsAndMutants(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	original, err := os.ReadFile(filepath.Join(repository, "internal/native/wasm/shim.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	checks := `import assert from 'node:assert/strict';
import { wasiShim } from './shim.mjs';
const stdout=[], stderr=[];
const shim=wasiShim({stdout:b=>stdout.push(...b),stderr:b=>stderr.push(...b)});
const memory=new WebAssembly.Memory({initial:1}); shim.attach({exports:{memory}});
const api=shim.imports.wasi_snapshot_preview1, view=new DataView(memory.buffer);
new Uint8Array(memory.buffer,100,5).set([65,0,255,66,67]);
view.setUint32(0,100,true); view.setUint32(4,3,true); view.setUint32(8,103,true); view.setUint32(12,2,true);
const check=process.argv[2];
if(check==='iovecs') {
 assert.equal(api.fd_write(1,0,2,20),0,'stdout write');
 assert.deepEqual(stdout,[65,0,255,66,67],'all iovecs, exact bytes'); assert.equal(view.getUint32(20,true),5,'byte count');
 assert.equal(api.fd_write(2,0,2,20),0,'stderr write'); assert.deepEqual(stderr,stdout,'stderr bytes');
 memory.grow(1); assert.equal(api.fd_write(1,0,0,20),0,'grown memory'); assert.equal(view.getUint32.call(new DataView(memory.buffer),20,true),0);
 assert.equal(api.fd_write(3,0,2,20),8,'invalid output fd');
} else if(check==='exit') {
 assert.throws(()=>api.proc_exit(23),e=>e.name==='AdamicExit'&&e.code===23,'exit must terminate');
} else if(check==='preopens') {
 for(let fd=0;fd<8;fd++) assert.equal(api.fd_prestat_get(fd,32),8,'no preopened directory');
 for(const name of ['fd_close','fd_seek','fd_read','fd_readdir','fd_prestat_dir_name','path_open','fd_fdstat_get']) assert.equal(api[name](3,32),8,name);
} else if(check==='empty') {
 view.setUint32(32,123,true); view.setUint32(36,123,true);
 for(const name of ['args_sizes_get','environ_sizes_get']) {
  assert.equal(api[name](32,36),0); assert.equal(view.getUint32(32,true),0); assert.equal(view.getUint32(36,true),0);
 }
 for(const name of ['args_get','environ_get']) assert.equal(api[name](0,0),0);
 for(const name of ['random_get','not_a_wasi_service']) assert.throws(()=>api[name](),e=>e.name==='AdamicUnsupportedWASI'&&e.message.includes(name));
} else if(check==='clock') {
 view.setBigUint64(40,0n,true); assert.equal(api.clock_time_get(0,1n,40),0,'realtime');
 const realtime=Number(view.getBigUint64(40,true)/1000000n); assert.ok(Math.abs(realtime-Date.now())<1000,'realtime is the host clock in nanoseconds');
 assert.equal(api.clock_time_get(1,1n,48),0,'monotonic'); const first=view.getBigUint64(48,true);
 assert.equal(api.clock_time_get(1,1n,56),0,'monotonic again'); assert.ok(first>0n&&view.getBigUint64(56,true)>=first,'monotonic never goes back');
 assert.equal(api.clock_time_get(9,1n,64),28,'unknown clock is EINVAL');
} else if(check==='monotonic') {
 assert.equal(api.clock_time_get(0,1n,40),0,'realtime'); assert.equal(api.clock_time_get(1,1n,48),0,'monotonic');
 assert.ok(view.getBigUint64(48,true)<view.getBigUint64(40,true)/2n,'monotonic is not the wall clock');
}
`
	for _, check := range []string{"iovecs", "exit", "preopens", "empty", "clock", "monotonic"} {
		t.Run(check, func(t *testing.T) {
			directory := t.TempDir()
			if err := os.WriteFile(filepath.Join(directory, "shim.mjs"), original, 0644); err != nil {
				t.Fatal(err)
			}
			script := filepath.Join(directory, "check.mjs")
			if err := os.WriteFile(script, []byte(checks), 0644); err != nil {
				t.Fatal(err)
			}
			actual := execute(t, "node", script, check)
			if actual.exitCode != 0 {
				t.Fatalf("control failed: %s", actual.stderr)
			}
			changes := map[string][2]string{
				"iovecs": {"index < count", "index < count - 1"},
				"exit":   {"throw new AdamicExit(code)", "return 0"},
				// Two clock mutants: one returns without writing the time, one serves monotonic from the wall clock.
				"clock":     {"new DataView(memory.buffer).setBigUint64(time, nanoseconds, true); return 0;", "return 0;"},
				"monotonic": {"id === 0 ? BigInt(Date.now()) * 1000000n : BigInt(Math.round(performance.now() * 1e6))", "BigInt(Date.now()) * 1000000n"},
				"preopens":  {"fd_prestat_get: () => 8", "fd_prestat_get: (fd, pointer) => { new DataView(memory.buffer).setUint32(pointer, 0, true); new DataView(memory.buffer).setUint32(pointer + 4, 1, true); return 0; }"},
			}
			if change, ok := changes[check]; ok {
				mutant := strings.Replace(string(original), change[0], change[1], 1)
				if mutant == string(original) {
					t.Fatal("mutant did not change shim")
				}
				if err := os.WriteFile(filepath.Join(directory, "shim.mjs"), []byte(mutant), 0644); err != nil {
					t.Fatal(err)
				}
				actual = execute(t, "node", script, check)
				if actual.exitCode == 0 || !strings.Contains(string(actual.stderr), "AssertionError") {
					t.Fatalf("mutant not caught by assertion: exit %d stderr %s", actual.exitCode, actual.stderr)
				}
				t.Logf("mutant caught by %s: %s", check, actual.stderr)
			}
		})
	}
}

func TestWASIShimRequest(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	directory := t.TempDir()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "cmd/adamic/testdata/wasi/request.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	handler := -1
	for index, function := range program.Functions {
		if function.Name == "handleRequest" {
			handler = index
		}
	}
	source, err := native.WASI(program, handler)
	if err != nil || handler == -1 {
		t.Fatalf("request emission: handler=%d error=%v", handler, err)
	}
	binary := filepath.Join(directory, "request.wasm")
	if err := native.Build(source, binary, native.Options{Target: "wasm32-wasi", Request: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(directory, "request.mjs")
	probe := `import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {pathToFileURL} from 'node:url';
const [mode,binary,root]=process.argv.slice(2);
const requests=['','世界 🌍','echo','reassign','x'.repeat(1024),'a\0b'];
if(mode==='source') {
 const source=await import(pathToFileURL(root+'/cmd/adamic/testdata/wasi/request.a'));
 for(const request of requests) console.log(JSON.stringify(source.handleRequest(request)));
} else {
 const module=await WebAssembly.compile(readFileSync(binary));
 console.error('IMPORTS request reactor | '+WebAssembly.Module.imports(module).filter(i=>i.module==='wasi_snapshot_preview1').map(i=>i.name).sort().join(', '));
 let instance;
 if(mode==='shim') {
  const {wasiShim}=await import(pathToFileURL(root+'/internal/native/wasm/shim.mjs'));
  const shim=wasiShim({stdout:b=>process.stdout.write(b),stderr:b=>process.stderr.write(b)});
  instance=await WebAssembly.instantiate(module,shim.imports); shim.attach(instance); instance.exports._initialize();
 } else {
  const {WASI}=await import('node:wasi'); const wasi=new WASI({version:'preview1',args:[],env:{},preopens:{},returnOnExit:true});
  instance=await WebAssembly.instantiate(module,wasi.getImportObject()); wasi.initialize(instance);
 }
 const api=instance.exports, encoder=new TextEncoder(), decoder=new TextDecoder();
 const baseline=api.adamic_live();
 for(const request of requests) {
  const bytes=encoder.encode(request), pointer=api.malloc(Math.max(1,bytes.length));
  new Uint8Array(api.memory.buffer,pointer,bytes.length).set(bytes);
  const response=api.adamic_request(pointer,bytes.length);
  console.log(JSON.stringify(decoder.decode(new Uint8Array(api.memory.buffer,api.adamic_response_bytes(response),api.adamic_response_length(response)))));
  api.adamic_release(response); api.free(pointer);
  assert.equal(api.adamic_live(),baseline,'request leaked counted values');
 }
}
`
	if err := os.WriteFile(script, []byte(probe), 0644); err != nil {
		t.Fatal(err)
	}
	shim := execute(t, "node", "--disable-warning=ExperimentalWarning", script, "shim", binary, root)
	reference := execute(t, "node", "--disable-warning=ExperimentalWarning", script, "wasi", binary, root)
	if difference := disagreement(reference, shim); difference != "" {
		t.Fatalf("request shim vs WASI: %s\nshim %+v\nreference %+v", difference, shim, reference)
	}
	t.Logf("%s", shim.stderr)
	oracle := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), script, "source", binary, root)
	// Inventory goes only to the two Wasm runs, not to the source oracle.
	shim.stderr = []byte(strings.TrimPrefix(string(shim.stderr), strings.SplitN(string(shim.stderr), "\n", 2)[0]+"\n"))
	if difference := disagreement(oracle, shim); difference != "" {
		t.Fatalf("request shim vs source: %s\nshim %+v\nsource %+v", difference, shim, oracle)
	}
}

// Every module reads both clocks as it starts (the runtime's clock origin), and no Adamic program reads
// time yet, so a probe in C holds the shim's clock to node:wasi's: realtime is the host's wall clock,
// and monotonic only moves forward. A shim whose clock_time_get returns without writing the time is the
// mutant, and the probe's own output catches it.
func TestWASIShimClockReadsTheHost(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	clang, err := native.WASIClang()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	source := filepath.Join(directory, "clock.c")
	probe := "#include <stdio.h>\n#include <time.h>\nint main(void) {\n" +
		"\tstruct timespec wall, first, second;\n" +
		"\tclock_gettime(CLOCK_REALTIME, &wall); clock_gettime(CLOCK_MONOTONIC, &first); clock_gettime(CLOCK_MONOTONIC, &second);\n" +
		"\tprintf(\"realtime after 2023: %d\\n\", wall.tv_sec > 1672531200 && wall.tv_sec < 4102444800);\n" +
		"\tprintf(\"monotonic started and forward: %d\\n\", (first.tv_sec > 0 || first.tv_nsec > 0) && (second.tv_sec > first.tv_sec || (second.tv_sec == first.tv_sec && second.tv_nsec >= first.tv_nsec)));\n" +
		"\tprintf(\"monotonic is not the wall clock: %d\\n\", first.tv_sec < wall.tv_sec / 2);\n" +
		"\treturn 0;\n}\n"
	if err := os.WriteFile(source, []byte(probe), 0o644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "clock.wasm")
	if output, err := toolchain.Command(clang, "--target=wasm32-wasi", "--sysroot="+os.Getenv("WASI_SYSROOT"), "-O2", source, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("clock probe: %v\n%s", err, output)
	}
	const agreed = "realtime after 2023: 1\nmonotonic started and forward: 1\nmonotonic is not the wall clock: 1\n"
	reference := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "wasi.mjs"), binary)
	if string(reference.stdout) != agreed || reference.exitCode != 0 {
		t.Fatalf("node:wasi control: exit %d stdout %q stderr %q", reference.exitCode, reference.stdout, reference.stderr)
	}
	actual := execute(t, "node", filepath.Join(repository, "oracle", "wasi-shim.mjs"), binary)
	if difference := disagreement(reference, actual); difference != "" {
		t.Fatalf("shim vs node:wasi: %s\nshim: exit %d stdout %q stderr %q", difference, actual.exitCode, actual.stdout, actual.stderr)
	}
	// The same runner over a shim whose clock returns without writing the time.
	original, err := os.ReadFile(filepath.Join(repository, "internal/native/wasm/shim.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	runner, err := os.ReadFile(filepath.Join(repository, "oracle", "wasi-shim.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	const write = "new DataView(memory.buffer).setBigUint64(time, nanoseconds, true); return 0;"
	mutant := strings.Replace(string(original), write, "return 0;", 1)
	if mutant == string(original) {
		t.Fatal("clock mutant did not change the shim")
	}
	for path, contents := range map[string]string{"internal/native/wasm/shim.mjs": mutant, "oracle/wasi-shim.mjs": string(runner)} {
		if err := os.MkdirAll(filepath.Join(directory, filepath.Dir(path)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, path), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mutated := execute(t, "node", filepath.Join(directory, "oracle", "wasi-shim.mjs"), binary)
	if disagreement(reference, mutated) != "stdout differs" {
		t.Fatalf("clock mutant escaped: exit %d stdout %q stderr %q", mutated.exitCode, mutated.stdout, mutated.stderr)
	}
	t.Logf("clock mutant caught by the probe: %q", mutated.stdout)
}
