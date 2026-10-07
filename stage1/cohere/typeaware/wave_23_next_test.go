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

func wave23NextControls() []string {
	return []string{
		"declare const a:string[]; 'Admin' in a; 'length' in a; 'map' in a; '0' in a; '00' in a; '4294967294' in a; '4294967295' in a; '-1' in a; '+' in a; '' in a;",
		"declare const a:readonly ['first','second'];declare const k:'third'|'fourth';k in a;'1' in a;declare const mixed:'third'|'length';mixed in a;",
		"declare const a:string[]|number[];declare const k:string;declare const sym:symbol;'missing' in a;k in a;sym in a;1 in a;",
		"declare const a:string[]&{extra:1};'extra' in a;'missing' in a;declare const mixed:string[]|Record<string,number>;'missing' in mixed;",
		"declare const a:string[];a.length<0;0>a.length;a.length>=0;a.length<=-1;a.length>-1;a.length===-1;a.length==-1;a.length!==-1;-1!=a.length;",
		"declare const a:string[];a.length===0;a.length>0;a.length>=1;a.length<=0;a.length<1;a.length===1;declare const count:number;a.length<count;a.length < +0;",
		"declare const a:string[];a.length < 0x0;a.length <= -0x1;a.length < 0b0;a.length >= 0o0;a.length < 0e2;a.length <= -(1_0);a.length >= -0;a.length < 1e999;",
		"declare const a:string|Uint8Array|readonly [1,2];a.length < 0;declare const maybe:string[]|undefined;maybe?.length < 0;declare const fake:{length:number};fake.length < 0;",
		"declare const m:Map<string,number>;declare const s:ReadonlySet<string>;m.size < 0;s.size >= 0;declare const w:WeakMap<object,number>;w.size < 0;",
		"function f<T extends string[]>(a:T){a.length<0;'missing' in a;}function g<T extends Map<string,number>>(m:T){m[0];m.size < 0;}",
		"declare const m:Map<string,number>;m['total']=1;m[0];m[true];m[1n];m[null];m[undefined];m['size'];m['get'];m[Symbol.iterator];declare const k:string;m[k];",
		"declare const m:Map<string,number>|Set<number>;m['missing'];m[0];declare const key:'missing'|'size';m[key];declare const k:'missing'|number;m[k];",
		"declare const w:WeakMap<object,number>;w[0];declare const s:WeakSet<object>;s['missing'];declare const m:Map<string,number>&{extra:1};m[0];",
		"declare const m:Map<string,number>;Object.keys(m);Object.values(m);Object.entries(m);Object.getOwnPropertyNames(m);Object.freeze(m);Object['keys'](m);(Object.keys)(m);",
		"declare const m:Map<string,number>;function f(Object:ObjectConstructor){Object.keys(m);}class Registry extends Map<string,number>{label='x';}Object.keys(new Registry());",
		"class Map{size=-1;}declare const m:Map;m.size<0;m[0];Object.keys(m);",
		"declare const text:string;text.trim();(text.slice(1));text.trim().padStart(4);text.replace(/x/g,'a');text.split('/');text.includes('a');text.toString();",
		"declare const maybe:string|undefined;maybe?.toLowerCase();declare const list:readonly string[];list.slice(1);list.includes('x');list.join(',');declare const either:string|string[];either.includes('a');",
		"declare let text:string;text=text.trim();void text.trim();const kept=text.trim();text.normalize('NFC');text.repeat(2);text.localeCompare('x');",
		"declare const list:string[];list.push('x');list.sort();list.splice(0,1);list.map(x=>x);list.forEach(x=>{});list.concat(['x']);list.slice();list.indexOf('x');list.entries();list.keys();list.values();",
		"declare const text:string;declare const f:(s:string)=>string;text.replace(/x/,f);text.replace(/x/,function(m){return m;});text.replace(/x/,(m)=>m);declare const anyValue:any;text.replace(/x/,anyValue);declare const unknownValue:unknown;text.replace(/x/,unknownValue);",
		"declare const text:string;declare const cb:string|((x:string)=>string);text.replace(/x/,cb);declare const rest:any[];text.concat(...rest);declare const restString:string[];text.concat(...restString);",
		"interface Label{slice(n:number):Label}declare const l:Label;l.slice(1);interface String{trim():void}declare const s:String;s.trim();declare const loose:any;loose.trim();",
		"/* 世界 🌍 */\r\ndeclare const é:string;\r\n (é.trim());\r\n declare const a:string[]; a.length < -1;\r\n",
		"import {parseJson} from './nexus/source/structured-text/json/Json.js';import {parseJsonc} from './nexus/source/structured-text/json/Jsonc.js';import {readJsonFile,writeJsonFile,type JsonFileWriteOutcomeType,type JsonFileReadOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';import {readPackageJson} from './nexus/source/system/PackageJson.js';parseJson('', 'x');parseJsonc('', 'x');readJsonFile<{x:1}>('', 'x');writeJsonFile('',{},'x');readPackageJson('');",
		"import {parseJson} from './nexus/source/structured-text/json/Json.js';import {parseJsonc} from './nexus/source/structured-text/json/Jsonc.js';import {readJsonFile,writeJsonFile,type JsonFileWriteOutcomeType,type JsonFileReadOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';import {readPackageJson} from './nexus/source/system/PackageJson.js';(writeJsonFile('',{},'x'));void writeJsonFile('',{},'x');const kept=writeJsonFile('',{},'x');function wrapper(){return writeJsonFile('',{},'x');}wrapper();",
		"import {parseJson} from './nexus/source/structured-text/json/Json.js';import {parseJsonc} from './nexus/source/structured-text/json/Jsonc.js';import {readJsonFile,writeJsonFile,type JsonFileWriteOutcomeType,type JsonFileReadOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';import {readPackageJson} from './nexus/source/system/PackageJson.js';async function asyncWrite():Promise<JsonFileWriteOutcomeType>{return writeJsonFile('',{},'x');}async function run(){await asyncWrite();(await (asyncWrite()));asyncWrite();}",
		"import {parseJson} from './nexus/source/structured-text/json/Json.js';import {parseJsonc} from './nexus/source/structured-text/json/Jsonc.js';import {readJsonFile,writeJsonFile,type JsonFileWriteOutcomeType,type JsonFileReadOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';import {readPackageJson} from './nexus/source/system/PackageJson.js';declare const store:{save():JsonFileWriteOutcomeType}|undefined;store?.save();declare const f:(()=>JsonFileWriteOutcomeType)|undefined;f?.();",
		"import {parseJson} from './nexus/source/structured-text/json/Json.js';import {parseJsonc} from './nexus/source/structured-text/json/Jsonc.js';import {readJsonFile,writeJsonFile,type JsonFileWriteOutcomeType,type JsonFileReadOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';import {readPackageJson} from './nexus/source/system/PackageJson.js';type Extra=JsonFileWriteOutcomeType|{outcome:'Skipped'};declare function extra():Extra;extra();declare function success():Extract<JsonFileWriteOutcomeType,{outcome:'Written'}>;success();",
		"import {parseJson} from './nexus/source/structured-text/json/Json.js';import {parseJsonc} from './nexus/source/structured-text/json/Jsonc.js';import {readJsonFile,writeJsonFile,type JsonFileWriteOutcomeType,type JsonFileReadOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';import {readPackageJson} from './nexus/source/system/PackageJson.js';declare function both():JsonFileWriteOutcomeType|JsonFileReadOutcomeType;both();declare function narrow():Extract<JsonFileReadOutcomeType,{outcome:'Parsed'}>;narrow();",
		"import {parseJson} from './nexus/source/structured-text/json/Json.js';import {parseJsonc} from './nexus/source/structured-text/json/Jsonc.js';import {readJsonFile,writeJsonFile,type JsonFileWriteOutcomeType,type JsonFileReadOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';import {readPackageJson} from './nexus/source/system/PackageJson.js';declare const yes:boolean;yes&&writeJsonFile('',{},'x');yes?writeJsonFile('',{},'x'):writeJsonFile('',{},'x');writeJsonFile('',{},'x').outcome;",
		"import {parseJson} from './nexus/source/structured-text/json/Json.js';import {parseJsonc} from './nexus/source/structured-text/json/Jsonc.js';import {readJsonFile,writeJsonFile,type JsonFileWriteOutcomeType,type JsonFileReadOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';import {readPackageJson} from './nexus/source/system/PackageJson.js';declare function wrapped():JsonFileReadOutcomeType<{x:number}>;wrapped();declare function promised():Promise<JsonFileReadOutcomeType<string>|undefined>;async function f(){await promised();promised();}",
		"import {parseJson} from './nexus/source/structured-text/json/Json.js';import {parseJsonc} from './nexus/source/structured-text/json/Jsonc.js';import {readJsonFile,writeJsonFile,type JsonFileWriteOutcomeType,type JsonFileReadOutcomeType} from './nexus/source/structured-text/json/JsonFile.js';import {readPackageJson} from './nexus/source/system/PackageJson.js';/* 世界 🌍 */\r\n(writeJsonFile('',{},'é'));\r\n",
		"type JsonFileWriteOutcomeType={outcome:'Written'}|{outcome:'Unwritable';message:string};declare function own():JsonFileWriteOutcomeType;own();",
		"declare const s:string;s.at(0);s.charAt(0);s.charCodeAt(0);s.codePointAt(0);s.concat('x');s.endsWith('x');s.indexOf('x');s.isWellFormed();s.lastIndexOf('x');s.match(/x/);s.matchAll(/x/g);s.padEnd(3);s.padStart(3);s.replaceAll('x','a');s.search(/x/);s.startsWith('x');s.substr(1);s.substring(1);s.toLocaleLowerCase();s.toLocaleUpperCase();s.toLowerCase();s.toUpperCase();s.toWellFormed();s.trimEnd();s.trimLeft();s.trimRight();s.trimStart();s.valueOf();",
		"declare const a:readonly number[];a.at(0);a.flat();a.lastIndexOf(1);a.toReversed();a.toSorted();a.toSpliced(0,1);a.with(0,1);a.toSorted((x,y)=>x-y);declare const compare:any;a.toSorted(compare);declare const maybe:((x:number,y:number)=>number)|undefined;a.toSorted(maybe);",
		"import {Hidden} from './nexus/source/structured-text/json/Json.js';Hidden.own();",
		"declare const m:Map<string,number>;declare const key:('absent'|null);m[key];declare const anyKey:any;m[anyKey];declare const unknownKey:unknown;m[unknownKey];declare const symbolKey:symbol;m[symbolKey];",
	}
}
func wave23NextNexusFiles(h *harness) []string {
	var paths []string
	for _, file := range []struct{ path, source string }{
		{"nexus/source/structured-text/json/Json.ts", "export type JsonParseOutcomeType<T=unknown>={outcome:'Parsed';value:T}|{outcome:'Invalid';message:string};export declare function parseJson<T=unknown>(text:string,context:string):JsonParseOutcomeType<T>;export namespace Hidden {export type JsonParseOutcomeType={outcome:'Parsed'}|{outcome:'Invalid'};export declare function own():JsonParseOutcomeType;}"},
		{"nexus/source/structured-text/json/Jsonc.ts", "export type JsoncParseOutcomeType<T=unknown>={outcome:'Parsed';value:T}|{outcome:'Invalid';message:string};export declare function parseJsonc<T=unknown>(text:string,context:string):JsoncParseOutcomeType<T>;"},
		{"nexus/source/structured-text/json/JsonFile.ts", "export type JsonFileReadOutcomeType<T=unknown>=({outcome:'Parsed';value:T}|({outcome:'Unreadable';message:string})|{outcome:'Invalid';message:string});export type JsonFileWriteOutcomeType=({outcome:'Written'}|{outcome:'Unwritable';message:string});export declare function readJsonFile<T=unknown>(path:string,context:string):JsonFileReadOutcomeType<T>;export declare function writeJsonFile(path:string,value:unknown,context:string):JsonFileWriteOutcomeType;"},
		{"nexus/source/system/PackageJson.ts", "export type PackageJsonReadOutcomeType<T=unknown>={outcome:'Read';value:T}|{outcome:'Missing'}|{outcome:'Invalid';message:string};export declare function readPackageJson(path:string):PackageJsonReadOutcomeType;"},
	} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(h.directory, file.path)), 0755); err != nil {
			h.t.Fatal(err)
		}
		paths = append(paths, h.write(file.path, file.source+"\n"))
	}
	return paths
}
func wave23NextMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(h.repository, "stage1/cohere/typeaware"))
	if err != nil {
		h.t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || (filepath.Ext(entry.Name()) != ".a" && filepath.Ext(entry.Name()) != ".ts") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(h.repository, "stage1/cohere/typeaware", entry.Name()))
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if entry.Name() == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatal("nonunique mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_23_next_suite.a"), archive, false)
}

// Not parallel: builds, sanitizer archives and measured runs share one machine.
func TestWave23NextAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE23_NEXT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_23_next_suite.a")
	binary := h.build(stage0, "native", entry, archive, false)
	oracle := volumeOracle(h, "wave23-next-oracle", "oracle_wave_23_next.go")
	h.write("prelude.d.ts", "")
	config := h.write("controls-tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ESNext"],"noEmit":true},"files":["prelude.d.ts"]}`)
	var paths []string
	for i, source := range wave23NextControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	// External TypeScript declaration fixtures keep the production rule's .ts path
	// identity. Adamic implementations and all executable controls remain .a.
	paths = append(paths, wave23NextNexusFiles(h)...)
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"correctness-no-collection-misuse", "correctness-no-discarded-outcome", "correctness-no-discarded-pure-result"} {
		if !bytes.Contains(truth.stdout, []byte("\tnexus/"+name+"\t")) {
			t.Fatal("no positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "native-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"collection", "no_collection_misuse.a", "value < 4294967295", "value < 4294967294"},
		{"outcome", "no_discarded_outcome.a", "collected.push(arm)", "collected.push(arm); collected.push(arm)"},
		{"pure", "no_discarded_pure_result.a", "const reported = this.rules.parser.node(callIndex);", "const reported = this.rules.parser.node(index);"},
	} {
		mutant := wave23NextMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}

	for _, change := range []struct{ name, file, from, to string }{
		{"ancestry-count", "bridge/tsgo/checker/declaration_ancestry.go", "out.number(uint64(count))", "out.number(uint64(count + 1))"},
		{"awaited", "bridge/tsgo/checker/awaited_shape.go", "t = checker.Checker_getAwaitedType(c, t)", "t = c.GetTypeAtLocation(node)"},
	} {
		overlay := h.overlay(change.name, change.file, change.from, change.to)
		mutantArchive := h.archive(change.name, overlay, false)
		mutant := h.build(stage0, change.name+"-native", entry, mutantArchive, false)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("fact mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, Go byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE23_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE23_COMPILER_MANIFEST")},
	} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		goRun := h.must(corpus.name+"-timing-go", exec.Command(oracle, corpus.config, corpus.manifest))
		command := exec.Command(binary, corpus.config, corpus.manifest)
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		nativeRun := h.must(corpus.name+"-timing-native", command)
		if !bytes.Equal(goRun.stdout, nativeRun.stdout) {
			t.Fatal("timed bytes differ")
		}
		t.Logf("%s timing: native %s, Go %s; native stderr %s; Go stderr %s", corpus.name, nativeRun.elapsed, goRun.elapsed, nativeRun.stderr, goRun.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoInspect(program,file,0,1,'Identifier','raw-shape');tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','declaration-ancestry\n1'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// mutant keeps released program")
	staleArchive := h.archive("released-registry", overlay, false)
	staleBinary := h.build(stage0, "released-registry-probe", released, staleArchive, false)
	staleGot := h.run("released-registry-run", exec.Command(staleBinary, config, probe))
	// The retained registry mutant answers normally; only the required refusal catches it.
	if staleGot.err != nil || len(staleGot.stderr) != 0 {
		t.Fatal("released registry mutant survived")
	}
	t.Logf("released handle: normal exit 70; registry mutant caught by required released-handle error")
}
