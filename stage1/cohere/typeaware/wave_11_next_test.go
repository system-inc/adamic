package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func wave11NextControls() []string {
	return []string{
		`declare const rows:string[]; 'Admin' in rows; '0' in rows; '01' in rows; 'length' in rows; 'map' in rows; '4294967294' in rows; '4294967295' in rows; '-1' in rows;`,
		`declare const rows:readonly number[];declare const k:'missing'|'absent';k in rows;declare const maybe:'missing'|'length';maybe in rows;declare const wide:string;wide in rows;0 in rows;`,
		`declare const x:number[]|[number];'missing' in x;declare const y:number[]|{value:number};'missing' in y;function generic<T extends number[]>(rows:T){return 'value' in rows;}`,
		`declare const a:number[];a.length<0;a.length>=0;a.length<=-1;a.length>-1;a.length===-1;a.length!==-1;a.length==-1;a.length!=-1;a.length===0;a.length>0;a.length<-0;`,
		`declare const m:Map<string,number>;0>m.size;0<=m.size;-1>=m.size;-1<m.size;-1===m.size;m.size<+0;m.size<(-0x0);m.size<=(-0x1);m.size<=-1_0;m.size<=-1e0;`,
		`declare const s:string;declare const t:Uint8Array;s.length<0;t.length>=0;declare const list:readonly number[]|undefined;list?.length<0;declare const value:{length:number};value.length<0;`,
		`declare const a:ReadonlyMap<string,number>|ReadonlySet<number>;a.size!==-1;declare const w:WeakMap<object,number>;w.size>=0;class Extra extends Map<string,number>{x=1};new Extra().size<0;`,
		`declare const map:Map<string,number>;map['key'];map[1];map[true];map[null];map[undefined];map[1n];map['size'];map['get'];map[Symbol.iterator];declare const wide:string;map[wide];`,
		`declare const a:Set<number>|WeakSet<object>;a['value'];a[0];declare const b:Map<string,number>&{extra:number};b['unknown'];class Map<K,V>{size=0};declare const local:Map<string,number>;local['key'];local.size<0;`,
		`Object.keys(new Map());Object.values(new Set());Object.entries(new WeakMap());Object.getOwnPropertyNames(new WeakSet());Object.keys({x:1});`,
		`function f(Object:{keys(x:unknown):string[]}){Object.keys(new Map())}Object['keys'](new Map());Object.keys();`,
		`declare const s:string;s.trim();s.trimStart();s.trimEnd();s.slice(1);s.replace('x','y');s.match(/x/);s.repeat(2);s.normalize();void s.trim();const saved=s.trim();`,
		`declare const s:string|null;s?.trim();(s?.trim());(s.trim)();(s.trim());declare const a:string[];a.slice(1);a.concat([]);a.join(',');a.map(x=>x);a.push('x');`,
		`declare const a:readonly number[];a.slice();a.toSorted();a.toSorted((x,y)=>x-y);a.toReversed();a.with(0,1);declare const s:string;s.replace(/x/g,m=>m);s.replace(/x/g,function(m){return m});`,
		`declare const s:string;declare const callback:(x:string)=>string;declare const unknown:unknown;declare const any:any;s.replace('x',callback);s.replace('x',unknown);s.replace('x',any);`,
		`declare const s:string;declare const replacement:string|((x:string)=>string);s.replace('x',replacement);declare const constrained:never;s.replace('x',constrained);`,
		`interface String{trim():string};declare const s:String;s.trim();interface ReadonlyArray<T>{slice():T[]};declare const a:ReadonlyArray<number>;a.slice();`,
		`declare const s:{trim():string};s.trim();class Text{trim(){return ''}}new Text().trim();declare const both:string|string[];both.slice();`,
		"/* 世界 🌍 */\r\ndeclare const s:string;\r\n(s.trim());\r\n",
		`import type {JsonFileWriteOutcomeType,JsonFileReadOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';declare function save():JsonFileWriteOutcomeType;save();(save());void save();const result=save();`,
		`import type {JsonFileWriteOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';declare function save():Promise<JsonFileWriteOutcomeType>;save();async function f(){await save();(await (save()));void await save();}`,
		`import type {JsonFileWriteOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';declare const api:{save():JsonFileWriteOutcomeType}|undefined;api?.save();`,
		`import type {JsonFileWriteOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';declare function save():JsonFileWriteOutcomeType|{extra:true};save();declare function partial():Extract<JsonFileWriteOutcomeType,{outcome:'Written'}>;partial();`,
		`import type {JsonFileReadOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';declare function read():JsonFileReadOutcomeType<{value:string}>;read();`,
		`type JsonFileWriteOutcomeType={outcome:'Written'}|{outcome:'Unwritable'};declare function own():JsonFileWriteOutcomeType;own();declare function arbitrary():{outcome:'Written'}|{outcome:'Unwritable'};arbitrary();`,
		`import type {JsonParseOutcomeType} from './nexus/source/structured-text/json/Json.js';import type {JsoncParseOutcomeType} from './nexus/source/structured-text/json/Jsonc.js';declare function both():JsonParseOutcomeType|JsoncParseOutcomeType;both();`,
		`import type {PackageJsonReadOutcomeType} from './nexus/source/system/PackageJson.js';declare function read():PackageJsonReadOutcomeType;read();`,
		"/* 世界 🌍 */\r\nimport type {JsonFileWriteOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';declare function save():JsonFileWriteOutcomeType;\r\n(save());\r\n",
	}
}

func wave11NextSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, pattern := range []string{"*.ts", "*.a"} {
		files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware", pattern))
		if err != nil {
			h.t.Fatal(err)
		}
		for _, path := range files {
			data, err := os.ReadFile(path)
			if err != nil {
				h.t.Fatal(err)
			}
			source := string(data)
			if filepath.Base(path) == file {
				if strings.Count(source, from) != 1 {
					h.t.Fatalf("nonunique mutant %s", name)
				}
				source = strings.Replace(source, from, to, 1)
			}
			source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
			source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
			if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0644); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_11_next_suite.a"), archive, false)
}

// Not parallel: archive builds, sanitizer subprocesses and measurements share a machine.
func TestWave11NextAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_11_NEXT_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_11_next_suite.a")
	binary := h.build(stage0, "wave-11-next", entry, archive, false)
	oracle := volumeOracle(h, "wave-11-next-oracle", "oracle_wave_11_next.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	for _, fixture := range []struct{ path, source string }{
		{"nexus/source/structured-text/json/JsonFile.ts", `export type JsonFileWriteOutcomeType=({outcome:'Written'})|(({outcome:'Unwritable';message:string}));export type JsonFileReadOutcomeType<T>=({outcome:'Read';value:T}|{outcome:'Unreadable';message:string});`},
		{"nexus/source/structured-text/json/Json.ts", `export type JsonParseOutcomeType={outcome:'Parsed';value:unknown}|{outcome:'Unparseable';message:string};`},
		{"nexus/source/structured-text/json/Jsonc.ts", `export type JsoncParseOutcomeType={outcome:'ParsedJsonc';value:unknown}|{outcome:'UnparseableJsonc';message:string};`},
		{"nexus/source/system/PackageJson.ts", `export type PackageJsonReadOutcomeType={outcome:'Read';value:unknown}|{outcome:'Unreadable';message:string};`},
	} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(directory, fixture.path)), 0755); err != nil {
			t.Fatal(err)
		}
		h.write(fixture.path, fixture.source)
	}
	var paths []string
	for i, source := range wave11NextControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"nexus/correctness-no-collection-misuse", "nexus/correctness-no-discarded-outcome", "nexus/correctness-no-discarded-pure-result"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control: %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-11-next-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"collection", "correctness_no_collection_misuse.a", "this.rules.byte(node.end), '')", "this.rules.byte(node.end) + 1, '')"},
		{"outcome", "correctness_no_discarded_outcome.a", "this.rules.byte(node.end), '')", "this.rules.byte(node.end) + 1, '')"},
		{"pure", "correctness_no_discarded_pure_result.a", "this.rules.byte(call.end), '')", "this.rules.byte(call.end) + 1, '')"},
		{"lineage", "declaration_lineage.a", "declarationFile, library, children, frames.field()", "declarationFile, false, children, frames.field()"},
		{"awaited", "awaited_type_shape.a", "types(rules.ask(index, 'awaited-type-shape'), 'awaited-type-shape')", "types(rules.ask(index, 'raw-shape'), 'raw-shape')"},
	} {
		mutant := wave11NextSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, independent Go bytes catch byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	for _, population := range []struct{ name, root, config, manifest string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json"), "repository.manifest"},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), "compiler.manifest"},
	} {
		if population.root == "" {
			t.Log("compiler corpus not supplied")
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.manifest))
		if err != nil {
			t.Fatal(err)
		}
		var roots []string
		for _, path := range strings.Split(string(data), "\n") {
			if path != "" {
				roots = append(roots, filepath.Join(population.root, path))
			}
		}
		rootsManifest := h.write(population.name+".manifest", strings.Join(roots, "\n")+"\n")
		h.compare(population.name, oracle, binary, population.config, rootsManifest)
		h.compare(population.name+"-asan", oracle, asan, population.config, rootsManifest)
		for _, impl := range []struct{ name, path string }{{"go", oracle}, {"native", binary}} {
			command := exec.Command(impl.path, population.config, rootsManifest)
			command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			got := h.must(population.name+"-timed-"+impl.name, command)
			t.Logf("%s %s process=%s %s", population.name, impl.name, got.elapsed, strings.TrimSpace(string(got.stderr)))
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','declaration-lineage'));
`)
	probe := h.write("probe.ts", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released query: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released program.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, caught by required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
