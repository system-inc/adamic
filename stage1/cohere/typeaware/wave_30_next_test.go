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

// Not parallel: native builds, sanitizers and measured subprocesses share a machine.
func TestWave30NextAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_30_NEXT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_30_next_suite.a")
	binary := h.build(stage0, "wave-30-next", entry, archive, false)
	oracle := volumeOracle(h, "wave-30-next-oracle", "oracle_wave_30_next.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	controls := []string{
		"declare const m:Map<string,number>;m?.['extra'];m?.size<0;Object?.keys(m);declare const s:string;s.trim?.();",
		"declare const a:string[];declare const key:'Admin'|'Member';'Admin' in a;key in a;'0' in a;'length' in a;'map' in a;'00' in a;'4294967294' in a;'4294967295' in a;",
		"declare const a:readonly [string,string];'third' in a;declare const wide:string;wide in a;0 in a;Symbol.iterator in a;",
		"declare const a:string[]|number[];declare const key:'missing'|'length';'absent' in a;key in a;declare const mixed:string[]|Record<string,number>;'absent' in mixed;",
		"declare const a:string[];a.length<0;0>a.length;a.length>=0;a.length<=-1;-1<a.length;a.length===-1;a.length!==-1;a.length==0;a.length>0;a.length<=0;a.length<1;",
		"declare const text:string;declare const bytes:Uint8Array;text.length===-1;bytes.length<=-1;declare const tuple:readonly [1,2];tuple.length==-1;",
		"export function generic<T extends string[]>(items:T){return items.length<0;}declare const items:string[]|undefined;items?.length<0;declare const like:ArrayLike<string>;like.length<0;",
		"declare const a:string[];a.length<-0;a.length<=-0;a.length<0x0;a.length===-0.5;a.length<+0;a.length<-(0);a.length<=-1e-4;",
		"declare const a:string[];a.length<1e999;a.length===-1e999;declare const size:{length:number};size.length<0;",
		"declare const m:Map<string,number>;declare const s:Set<string>;declare const r:ReadonlyMap<string,number>;m.size!==-1;s.size>=0;r.size>-1;m['total']=1;s[0];r[1];",
		"declare const m:Map<string,number>;m['size'];m['get'];m[Symbol.iterator];declare const key:string;m[key];m[true];m[1n];m[null];m[undefined];",
		"declare const m:Map<string,number>|Set<number>;m['extra'];Object.keys(m);Object.values(m);Object.entries(m);Object.getOwnPropertyNames(m);Object.freeze(m);",
		"declare const w:WeakMap<object,number>;declare const s:WeakSet<object>;w['x'];s[false];w.size<0;Object.keys(w);",
		"class Registry extends Map<string,number>{label='x'};declare const r:Registry;r[0];Object.keys(r);r.size<0;",
		"class Map{size=-1;[key:string]:unknown};const local=new Map();local['x'];local.size<0;Object.keys(local);",
		"declare const m:Map<string,number>&{extra:string};m['x'];Object.keys(m);m.size<0;",
		"declare const m:Map<string,number>;export function own(Object:ObjectConstructor){Object.keys(m)};const fake={keys:(x:unknown)=>[]};fake.keys(m);",
		"const Object={keys:(x:unknown)=>[]};declare const m:Map<string,number>;Object.keys(m);",
		"declare let text:string;declare const list:string[];text.trim();text.replace('old','new');list.concat(['a']);list.slice(1);text=text.trim();const result=list.slice(0);void text.trim();",
		"declare const s:string|undefined;s?.toLowerCase();declare const a:readonly string[];a.indexOf('x');declare const both:string|string[];both.includes('x');(a.join(','));",
		"declare const s:string;declare const a:string[];s.trim().padStart(4);s.replace(/x/g,()=>{a.push('x');return 'x'});declare const replacer:(s:string)=>string;s.replace(/x/,replacer);",
		"declare const s:string;declare const loose:any;declare const unknownValue:unknown;declare const callback:((s:string)=>string)|undefined;s.replace(/x/,loose);s.replace(/x/,unknownValue);s.replace(/x/,callback);",
		"declare const a:string[];a.push('x');a.sort();a.splice(0,1);a.map(x=>x);a.forEach(x=>console.log(x));declare const s:string;s.normalize('NFC');s.repeat(2);s.localeCompare('x');",
		"class Text{trim(){return this}}new Text().trim();interface Label{slice(x:number):Label}declare const l:Label;l.slice(1);interface Array<T>{slice(x:number):void;first:T}declare const a:Array<string>;a.slice(1);",
		"declare const s:string;declare const args:[string,string];s.concat(...args);declare const fn:((s:string)=>string)&{tag:number};s.replace(/x/,fn);",
		"declare const a:number[];a.toSorted((x,y)=>x-y);a.toSorted();a.toReversed();a.toSpliced(0,1);a.with(0,2);a.at(0);",
		"/* 世界 🌍 */\r\ndeclare const a:string[];a.length < 0;\r\n'存在' in a;\r\n'a'.trim();\r\n",
		"import {parse} from './nexus/source/structured-text/json/Json';parse();(parse());void parse();const result=parse();",
		"import {parseJsonc} from './nexus/source/structured-text/json/Jsonc';parseJsonc();",
		"import {read,write,writeAsync,success,optional} from './nexus/source/structured-text/json/JsonFile';read<number>();write();writeAsync();export async function f(){await (writeAsync());}success();optional?.();",
		"import {readPackage} from './nexus/source/system/PackageJson';readPackage();",
		"import {parse} from './nexus/source/structured-text/json/Json';declare function extra():ReturnType<typeof parse>|{outcome:'Extra'};extra();export function wrapper(){return parse()}wrapper();",
		"import {parse} from './nexus/source/structured-text/json/Json';declare function only():Extract<ReturnType<typeof parse>,{outcome:'Parsed'}>;only();",
		"type JsonParseOutcomeType={outcome:'Parsed'}|{outcome:'Failed'};declare function local():JsonParseOutcomeType;local();",
		"import {JsonParseOutcomeType} from './nexus/source/structured-text/json/Json';import {JsonFileWriteOutcomeType} from './nexus/source/structured-text/json/JsonFile';declare function both():JsonFileWriteOutcomeType|JsonParseOutcomeType;both();",
		"import {namespaceOnly,single} from './nexus/source/structured-text/json/Json';namespaceOnly();single();",
		"import {parse} from './nexus/source/structured-text/json/Json';export async function list(){await Promise.all([parse(),parse()]);}true && parse();true?parse():parse();",
	}
	paths := []string{}
	for at, source := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", at), source+"\nexport {};\n"))
	}

	// Content is authored as .a. Filename aliases reproduce the imported TypeScript
	// declarations' identity, which the production outcome rule deliberately checks.
	helpers := map[string]string{
		"nexus/source/structured-text/json/Json":     `export type JsonParseOutcomeType=(({outcome:'Parsed';value:number})|({outcome:'Failed';message:string}));export declare function parse():JsonParseOutcomeType;export namespace Nested{export type JsonParseOutcomeType={outcome:'Parsed'}|{outcome:'Failed'};}export declare function namespaceOnly():Nested.JsonParseOutcomeType;export type Single={outcome:'Only'};export declare function single():Single;`,
		"nexus/source/structured-text/json/Jsonc":    `export type JsoncParseOutcomeType={outcome:'Parsed'}|{outcome:'Failed'};export declare function parseJsonc():JsoncParseOutcomeType;`,
		"nexus/source/structured-text/json/JsonFile": `export type JsonFileReadOutcomeType<T>={outcome:'Read';value:T}|{outcome:'Unreadable';message:string};export type JsonFileWriteOutcomeType={outcome:'Written'}|{outcome:'Unwritable';message:string};export declare function read<T>():JsonFileReadOutcomeType<T>;export declare function write():JsonFileWriteOutcomeType;export declare function writeAsync():Promise<JsonFileWriteOutcomeType>;export declare function success():Extract<JsonFileWriteOutcomeType,{outcome:'Written'}>;export declare const optional:(()=>JsonFileWriteOutcomeType)|undefined;`,
		"nexus/source/system/PackageJson":            `export type PackageJsonReadOutcomeType={outcome:'Read';value:object}|{outcome:'Unreadable';message:string};export declare function readPackage():PackageJsonReadOutcomeType;`,
	}
	for name, source := range helpers {
		path := filepath.Join(directory, name+".a")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		alias := filepath.Join(directory, name+".ts")
		if err := os.Symlink(filepath.Base(path), alias); err != nil {
			target, readError := os.Readlink(alias)
			if readError != nil || target != filepath.Base(path) {
				t.Fatal(err)
			}
		}
	}
	config = h.write("controls-tsconfig.json", `{"files":["control-000.a"],"compilerOptions":{"strict":true,"target":"ESNext","module":"ESNext","moduleResolution":"Bundler","lib":["ESNext"],"preserveSymlinks":true}}`)

	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"nexus/correctness-no-collection-misuse", "nexus/correctness-no-discarded-outcome", "nexus/correctness-no-discarded-pure-result"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	for _, id := range []string{"inOnArray", "impossibleSizeComparison", "bracketAccessOnCollection", "objectMethodOnCollection", "JsonParseOutcomeType", "JsoncParseOutcomeType", "JsonFileReadOutcomeType", "JsonFileWriteOutcomeType", "PackageJsonReadOutcomeType"} {
		if !bytes.Contains(truth.stdout, []byte(id)) {
			t.Fatalf("no positive control for %s", id)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-30-next-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, mutation := range []struct{ name, file, from, to string }{
		{"collection-boundary", "correctness_no_collection_misuse.a", "return value <= 0", "return value < 0"},
		{"outcome-range", "correctness_no_discarded_outcome.a", "rules.byte(node.end), '')", "rules.byte(node.end) + 1, '')"},
		{"pure-range", "correctness_no_discarded_pure_result.a", "rules.byte(call.end), '')", "rules.byte(call.end) + 1, '')"},
	} {
		sourceDir := filepath.Join(directory, mutation.name+"-source")
		if err := os.MkdirAll(sourceDir, 0755); err != nil {
			t.Fatal(err)
		}
		for _, pattern := range []string{"*.a", "*.ts"} {
			files, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware", pattern))
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				source := string(data)
				if filepath.Base(file) == mutation.file {
					if strings.Count(source, mutation.from) != 1 {
						t.Fatal("nonunique mutant")
					}
					source = strings.Replace(source, mutation.from, mutation.to, 1)
				}
				source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
				source = strings.ReplaceAll(source, "../lint/", filepath.Join(repository, "stage1/cohere/lint")+"/")
				if err := os.WriteFile(filepath.Join(sourceDir, filepath.Base(file)), []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
		}
		mutant := h.build(stage0, mutation.name, filepath.Join(sourceDir, "wave_30_next_suite.a"), archive, false)
		got := h.must(mutation.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("mutant survived: %s", mutation.name)
		}
		t.Logf("%s exits 0; independent Go bytes catch byte %d", mutation.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE_30_NEXT_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE_30_NEXT_COMPILER_MANIFEST")},
	} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		goRun := h.must(corpus.name+"-timed-go", exec.Command(oracle, corpus.config, corpus.manifest))
		cmd := exec.Command(binary, corpus.config, corpus.manifest)
		cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		native := h.must(corpus.name+"-timed-native", cmd)
		if !bytes.Equal(goRun.stdout, native.stdout) {
			t.Fatal("timed bytes differ")
		}
		t.Logf("%s whole process native=%s Go=%s; %s", corpus.name, native.elapsed, goRun.elapsed, native.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier',args[2]??''));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)

	for _, question := range []string{"type-declaration-ancestry", "type-leaf-facts\n1"} {
		got := h.run("released-"+strings.Split(question, "\n")[0], exec.Command(stale, config, probe, question))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
		}
		t.Logf("released %s handle refused with panic 70", strings.Split(question, "\n")[0])
	}
}
