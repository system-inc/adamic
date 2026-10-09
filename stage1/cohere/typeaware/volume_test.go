package typeaware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/corpusfiles"
)

var volumeRules = []string{"no_unsafe_type_assertion", "no_unsafe_member_access", "prefer_nullish_coalescing", "no_shadow", "no_unsafe_enum_comparison", "no_unsafe_assignment", "no_confusing_void_expression", "consistent_return", "switch_exhaustiveness_check", "unbound_method"}

func volumeOracle(h *harness, name, source string) string {
	virtual := filepath.Join(h.repository, "cohere/adamic_"+name+".go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(h.repository, "stage1/cohere/typeaware/testdata", source)}})
	if err != nil {
		h.t.Fatal(err)
	}
	binary := filepath.Join(h.directory, name)
	cmd := exec.Command("go", "build", "-overlay", h.write(name+"-overlay.json", string(data)), "-o", binary, virtual)
	cmd.Dir = filepath.Join(h.repository, "cohere")
	h.must(name+"-build", cmd)
	return binary
}

// These controls exercise decisions, report spans and repairs, not implementation structure.
func volumeControls() []string {
	return []string{
		"declare const x: unknown; const y=x as number;",
		"const x={a:1}; const y=x as {a:number;b:number};",
		"const o={foo:'hi',bar:1} as {foo:string}; const p={foo:'hi'} as {foo:string};",
		"const a=new Map /* comment */ ( ) as Map<string,string>; const b=new Map() as Map<string,string>;",
		"declare const x:any; const y=x as string; const z='x' as any;",
		"function f<T>(x:unknown) { return x as T; } function g<T extends {a:number}>(x:{a:number}) { return x as T; }",
		"declare const x:any; x.a.b; x['c']; declare const y:{[key:string]:number}; declare const key:any;y[key];",
		"declare const a:number|null; const b=a || 0; const c=a === null ? 0 : a;",
		"declare const a:string|undefined; if(a === undefined) { a = '🌍'; }",
		"declare const a:number|null; const b= a === null ? (true ? 0 : 1) : a;",
		"let a:string|undefined; if(a == null) { /* 世界 */ a = 'x'; // 🌍\n }",
		"enum E { A, B } enum F { A } declare const e:E; e===F.A; e===0; e===E.B; E.A===E.B;",
		"enum E {A='a',B='b'} declare const e:E; e==='a'; switch(e) { case 'a': break; }",
		"declare const x:any; const y:string=x; const a:[any,number]=[x,1];const [b,c]=a;",
		"declare const x:{a:any}; const {a}=x; declare const y:{p:{v:any}}; const {p:{v}}=y;",
		"declare const x:{a:{b:any}}; const {a:{b}}=x; declare const z:{a:number}; const {a:c}=z;",
		"declare const x:Set<any>; const a:{p:Set<string>}={p:x};",
		"declare const x:number; function f(x:number) { return x; }",
		"const x=1; (()=>{const x=2;return x;})();const g=(x:number)=>x;",
		"type T=number; class A<T> { m<T>(x:T) {return x;} static n<T>(){} } interface B<T>{p:T};",
		"enum E { E } const a=function a(){}; const b=wrap(function b(){}); declare function wrap(x:unknown):unknown;",
		"/* 世界 🌍 */\r\nconst é=1; function f(é:number) { return é; }",
		"export function f(){}; function g(){function f(){};return f;} export const x=1; function h(x:number){};",
		"declare function v():void; function f(x:boolean) { if(x) return v(); return; }",
		"declare function v():void; const f=()=>v(); function g(){return v();} function h(x:boolean){if(x)return v();console.log('x');}",
		"declare function v():void; const a= [v(), true?v():v()]; const f=()=>{v();};",
		"function f(x:boolean):number { if(x) return 1; } function g(x:boolean):number { if(x) return 1; return; }",
		"function f():boolean; function f(x:boolean):void; function f(x?:boolean) {if(x)return;return true;}",
		"async function f(x:boolean):Promise<void> { if(x) return;return Promise.resolve(); }",
		"async function f(x:boolean):Promise<Promise<void|number>> { if(x)return;return 1; }",
		"type P<T>={then(cb:(x:T)=>any):any}; async function f(x:boolean):P<void> {if(x)return;return {} as P<void>;}",
		"type P<T>={then(cb:number):any}; async function f(x:boolean):P<void> {if(x)return;return {} as P<void>;}",
		"type P<T>={then: number}; async function f(x:boolean):P<void> {if(x)return;return {} as P<void>;}",
		"function f(x:boolean) { while(true) { if(x)return 1; break; } }",
		"function f(x:boolean) {while(0xbn){if(x)return 1;}} function g(x:boolean){while(0x0n){if(x)return 1;}} function é(x:boolean){if(x)return 1;}",
		"function f(x:boolean) { outer: while(true) { switch(x){case true:break outer;default:return 1;} } }",
		"function f(x:boolean) { do { if(x)return 1;continue; } while(false); }",
		"function f(x:boolean) { for(;x;) { return 1; } } function g(){for(;;){return 1;}}",
		"function f(x:boolean) { for(;(x);) {return 1;} } function g() {try{return 1;}catch{}}",
		"function f(x:boolean) { switch(x){case true:return 1;default:return 2;} }",
		"function f(x:boolean) { try { return 1; } catch { throw 1; } finally { console.log('x'); } }",
		"declare const x:'a'|'b'|null; switch(x){case 'a':break;} declare const s:unique symbol; declare const y:typeof s|undefined;switch(y){case undefined:break;}",
		"declare const x:boolean; switch(x){case true:break;case false:break;} declare const y:number;switch(y){}",
		"class A { m(){}; f=function(){}; a=()=>{}; v(this:void){}; t(this:A){} };const a=new A;const m=a.m;const f=a.f;const b=a.a;const v=a.v;const t=a.t;a.m();if(a.m){};!a.m;",
		"interface A {m():void}; declare const a:A;const {m}=a;const {m:n}=a;const f=a['m'];",
		"interface A {m():void};declare const a:A;let m:()=>void;({m}=a);const f={m:function(){}}.m;",
		"declare const a:{m():void}|{m(this:void):void}; const f=a.m;",
		"const f=Math.floor;const {parseInt}=Number;const {map}=Array.prototype;class A extends Array<number>{};const a=new A;const m=a.map;",
		"const Math={floor(){}};const f=Math.floor;",
	}
}

// TestMain prepares immutable build inputs once before test units start.
// ADAMIC_TEST_SHARD=i/n (zero based) selects units; unset runs every unit.
func TestVolumeAgreementAndMutants(t *testing.T) {
	setupStarted := time.Now()
	var buildTime time.Duration
	built := func(build func() string) string {
		started := time.Now()
		path := build()
		buildTime += time.Since(started)
		return path
	}
	var shards []volumeShard
	var expected []string
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_VOLUME_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := built(func() string { return volumeStage0(t) })
	normal := built(func() string { return volumeArchive(h, "checker", "", false) })
	entry := filepath.Join(repository, "stage1/cohere/typeaware/volume_suite.ts")
	binary := built(func() string { return volumeNative(h, stage0, "volume", entry, normal, false, false) })
	oracle := built(func() string { return volumeProductOracle(h) })
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range volumeControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("native-globals.d.ts", "declare const console: {log():void};\n"))
	paths = append(paths, h.write("native-console.ts", "const detached=console.log;\nexport {};\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	controlInputs := volumeOracleData(t, "controls", config)
	truth := volumeOracleOutput(h, controlInputs, oracle, config, manifest)
	for _, name := range volumeRules {
		if !bytes.Contains(truth.stdout, []byte("@typescript-eslint/"+strings.ReplaceAll(name, "_", "-"))) {
			t.Fatalf("rule %s lacks a positive control", name)
		}
	}
	sanitized := built(func() string { return volumeArchive(h, "checker-asan", "", true) })
	asan := built(func() string { return volumeNative(h, stage0, "volume-asan", entry, sanitized, true, false) })

	expected = append(expected, volumeIDs("controls", paths)...)
	shards = append(shards, volumeShard{ids: volumeIDs("controls", paths), run: func(h *harness) {
		volumeCompareProducts(h, controlInputs, "controls", oracle, binary, asan, config, manifest)
	}})

	// Each new compiler question is proven consequential by replacing its answer
	// with another well-formed compiler fact. Every mutant must finish normally;
	// only the independent production-rule finding bytes may catch it.
	changes := volumeMutations()

	for _, change := range changes {

		var mutant string
		selected, err := volumeSelected(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
		if err != nil {
			t.Fatal(err)
		}
		if selected {
			overlay := h.overlay(change.name, change.path, change.from, change.to)
			archive := built(func() string { return volumeArchive(h, change.name+"-checker", overlay, false) })
			mutant = built(func() string { return volumeNative(h, stage0, change.name+"-native", entry, archive, false, true) })
		}
		id := "mutant:" + change.name
		expected = append(expected, id)
		shards = append(shards, volumeShard{ids: []string{id}, run: func(h *harness) {
			observed := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
			if len(observed.stderr) != 0 || bytes.Equal(observed.stdout, truth.stdout) {
				h.t.Fatalf("%s checker question mutant survived: %s", change.name, observed.stderr)
			}
			h.t.Logf("%s mutant: byte oracle catches byte %d; %s", change.name, firstDifference(observed.stdout, truth.stdout), summary(observed.stdout))
		}})
	}

	releasedSource := h.write("released-inspect.ts", volumeReleaseSource)
	stale := built(func() string {
		return volumeNative(h, stage0, "released-inspect", releasedSource, normal, false, false)
	})
	probe := h.write("probe.ts", "x;\n")

	var mutated string
	selected, err := volumeSelected(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if err != nil {
		t.Fatal(err)
	}
	if selected {
		overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant keeps released program live.")
		archive := built(func() string { return volumeArchive(h, "released-registry-checker", overlay, false) })
		mutated = built(func() string {
			return volumeNative(h, stage0, "released-registry-native", releasedSource, archive, false, true)
		})
	}

	expected = append(expected, "released-registry")
	shards = append(shards, volumeShard{ids: []string{"released-registry"}, run: func(h *harness) {
		observed := h.run("released-inspect-run", exec.Command(stale, config, probe))
		if code, ok := observed.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(observed.stderr) != "adamic: panic: invalid or released checker handle\n" {
			h.t.Fatalf("released program escaped: %v %s", observed.err, observed.stderr)
		}
		h.t.Log("released program queried with new question: panic 70, invalid or released checker handle")
		h.must("released-registry-run", exec.Command(mutated, config, probe))
		h.t.Log("released registry mutant: exit 0 caught by required panic 70")
	}})

	// Preserve the pre-port repository population, including declarations from config.
	if manifest := os.Getenv("ADAMIC_VOLUME_REPOSITORY_MANIFEST"); manifest != "" {
		config := filepath.Join(repository, "tsconfig.json")
		paths := volumeManifest(t, config, manifest)
		expected = append(expected, volumeIDs("repository", paths)...)
		shards = append(shards, volumeCorpusShards(t, "repository", paths, config, oracle, binary, asan)...)
	}
	if corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"); corpus != "" {
		paths := corpusfiles.Upstream(t, corpus, compilerCommit, []string{"src/compiler"}, []string{"*.ts"})
		expected = append(expected, volumeIDs("compiler", paths)...)
		shards = append(shards, volumeCorpusShards(t, "compiler", paths, filepath.Join(corpus, "src/compiler/tsconfig.json"), oracle, binary, asan)...)
	}
	outside := time.Since(setupStarted)
	t.Logf("volume-setup outside-shards=%.6fs build-inputs=%.6fs logic=%.6fs", outside.Seconds(), buildTime.Seconds(), (outside - buildTime).Seconds())
	volumeRunShards(t, repository, expected, shards)
}

// The strict-config limit is explicit. The six-rule runner still supports its
// previous configs; this expanded runner has not ported implicit-this messages.
// Not parallel: sanitizer archives consume the same limited scratch space as
// the main volume test; corpus timings must not compete with that test.
func TestVolumeConfigGuardAndMutant(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_VOLUME_GUARD_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/volume_suite.ts")
	binary := h.build(stage0, "volume", entry, archive, false)
	source := h.write("this.ts", "function f(){const x=this.m;const y:number=this;return x;} export {};\n")
	manifest := h.write("this.manifest", source+"\n")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"noImplicitThis":false,"target":"ES2022","lib":["ES2022"]}}`)
	observed := h.run("nonstrict-this", exec.Command(binary, config, manifest))
	message := []byte("adamic: panic: volume suite requires noImplicitThis; implicit-this messages are not yet ported\n")
	if code, ok := observed.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Equal(observed.stderr, message) || len(observed.stdout) != 0 {
		t.Fatalf("unsupported this mode escaped: %v %s", observed.err, observed.stderr)
	}
	overlay := h.overlay("strict-this", "bridge/tsgo/checker/facts.go", "option = p.Compiler.Options().NoImplicitThis", "option = p.Compiler.Options().StrictNullChecks")
	mutantArchive := h.archive("strict-this-checker", overlay, false)
	mutant := h.build(stage0, "strict-this-native", entry, mutantArchive, false)
	observed = h.must("strict-this-run", exec.Command(mutant, config, manifest))
	if len(observed.stderr) != 0 || !bytes.Contains(observed.stdout, []byte("findings ")) {
		t.Fatal("strict-this mutant did not finish normally")
	}
	t.Log("strict-this compiler-option mutant: refusal expectation catches exit 0 instead of 70")
	oracle := volumeOracle(h, "volume-oracle", "oracle_volume.go")
	truth := h.must("nonstrict-go", exec.Command(oracle, config, manifest))
	if bytes.Equal(observed.stdout, truth.stdout) {
		t.Fatal("implicit-this input did not distinguish the production message variants")
	}
	t.Logf("implicit-this mutant also differs from independent Go at byte %d", firstDifference(observed.stdout, truth.stdout))
	config = h.write("strict.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`)
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "volume-asan", entry, sanitized, true)
	h.compare("strict-this-control", oracle, asan, config, manifest)
	var controls []string
	for i, text := range volumeControls() {
		controls = append(controls, h.write(fmt.Sprintf("edge-control-%03d.ts", i), text+"\nexport {};\n"))
	}
	controls = append(controls, h.write("native-globals.d.ts", "declare const console: {log():void};\n"), h.write("native-console.ts", "const detached=console.log;\nexport {};\n"))
	controlsManifest := h.write("edge-controls.manifest", strings.Join(controls, "\n")+"\n")
	h.compare("edge-controls-asan", oracle, asan, config, controlsManifest)
	// Use the final scalar-option guard on both external corpora too.
	if manifest := os.Getenv("ADAMIC_VOLUME_REPOSITORY_MANIFEST"); manifest != "" {
		h.compare("repository-final-asan", oracle, asan, filepath.Join(repository, "tsconfig.json"), manifest)
	}
	if corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"); corpus != "" {
		paths := corpusfiles.Upstream(t, corpus, compilerCommit, []string{"src/compiler"}, []string{"*.ts"})
		manifest := h.write("compiler.manifest", strings.Join(paths, "\n")+"\n")
		h.compare("compiler-final-asan", oracle, asan, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
	}
}
