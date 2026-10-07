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

var continuationRuleNames = []string{"nexus/correctness-no-collection-misuse", "nexus/correctness-no-discarded-outcome", "nexus/correctness-no-discarded-pure-result"}

func continuationControls() []string {
	return []string{
		"declare const roles:('Admin'|'Member')[];declare const role:'Admin'|'Member';declare const tuple:readonly ['first','second'];'Admin' in roles;role in roles;'third' in tuple;('Admin') in (roles);",
		"declare const roles:string[];declare const index:number;declare const key:string;0 in roles;index in roles;'0' in roles;'length' in roles;'map' in roles;key in roles;Symbol.iterator in roles;",
		"declare const roles:string[];'4294967294' in roles;'4294967295' in roles;'00' in roles;'-1' in roles;'+1' in roles;'' in roles;",
		"declare const items:string[];declare const text:string;declare const bytes:Uint8Array;declare const map:Map<string,number>;declare const set:Set<string>;items.length<0;0>items.length;text.length===-1;bytes.length<=-1;map.size!==-1;set.size>=0;-1!=(items.length);",
		"declare const items:string[];items.length<(-0);items.length<-0x1;items.length<0b0;items.length>=0o0;items.length<0.0_0;items.length<1e-100;items.length<+0;",
		"declare const optional:string[]|undefined;declare const lengthy:{length:number};declare const arrayLike:ArrayLike<number>;declare const items:string[];items.length===0;items.length>0;items.length>=1;items.length<=0;optional?.length<0;lengthy.length<0;arrayLike.length<0;",
		"export function generic<T extends string[]>(items:T){items.length<0;'other' in items;} export function key<T extends string>(key:T,map:Map<string,number>){map[key];}",
		"declare const map:Map<string,number>;declare const readonlyMap:ReadonlyMap<string,number>;declare const set:Set<string>;declare const weak:WeakMap<object,number>;declare const index:number;map['total']=1;set[0];readonlyMap[index];Object.keys(map);Object.entries(readonlyMap);Object.values(set);Object.getOwnPropertyNames(weak);",
		"declare const map:Map<string,number>;declare const key:string;declare const record:Record<string,number>;map['size'];map['get'];map[Symbol.iterator];map[key];record['total'];Object.keys(record);Object.freeze(map);",
		"declare const map:Map<string,number>;map[true];map[1n];map[null];map[undefined];map['get' as 'get'|'missing'];map['missing' as 'other'|'missing'];map['key' as any];",
		"class Registry extends Map<string,number>{label='own';}declare const registry:Registry;Object.keys(registry);registry[0];function local(){class Map{size=-1;[key:string]:unknown;}const map=new Map();map.size<0;map['total'];Object.keys(map);}function parameter(Object:ObjectConstructor,map:Map<string,number>){Object.keys(map);}",
		"declare const text:string;declare const list:string[];declare const readonlyList:readonly string[];declare const optional:string|undefined;declare const union:string|string[];text.trim();(text.toUpperCase());list.concat([]);readonlyList.indexOf('x');union.includes('x');optional?.trim();text.trim().padStart(3);",
		"declare const text:string;declare const list:string[];declare function replacer(match:string):string;declare const pattern:RegExp;declare const loose:any;declare const unknownValue:unknown;text.replace(pattern,'');text.replace(pattern,replacer);text.replace(pattern,loose);text.replace(pattern,unknownValue);text.replace(pattern,(match)=>match);list.toSorted((left,right)=>left.localeCompare(right));void text.trim();text.normalize('NFC');text.repeat(2);list.push('x');list.map((item)=>item);",
		"class Label {trim(){return this;}}new Label().trim();interface Array<T>{slice(start:number):void;first:T;}declare const local:Array<string>;local.slice(1);declare const loose:any;loose.trim();",
		"declare const text:string;text.trim?.();text.slice<number>(1);text.replace<string>('x','y');text.replace(...(['x','y'] as [string,string]));",
		"/* 世界 🌍 */\r\n declare const é:string; (é.trim());\r\n declare const 漢:string[];' missing ' in 漢;\r\n",
	}
}

func continuationOutcomeFiles(h *harness) []string {
	h.t.Helper()
	for _, directory := range []string{"nexus/source/structured-text/json", "nexus/source/system", "lookalike"} {
		if err := os.MkdirAll(filepath.Join(h.directory, directory), 0755); err != nil {
			h.t.Fatal(err)
		}
	}
	const jsonFile = `export type JsonFileReadOutcomeType<T=unknown> = {outcome:'Parsed';value:T}|{outcome:'Unreadable';message:string}|{outcome:'Invalid';message:string};
export declare function read<T=unknown>():JsonFileReadOutcomeType<T>;
export type JsonFileWriteOutcomeType = (({outcome:'Written'}) | ({outcome:'Unwritable';message:string}));
export declare function write():JsonFileWriteOutcomeType;
export declare function asyncWrite():Promise<JsonFileWriteOutcomeType>;
export declare function success():Extract<JsonFileWriteOutcomeType,{outcome:'Written'}>;
export declare function extra():JsonFileWriteOutcomeType|{outcome:'Skipped'};
export declare const store:{save():JsonFileWriteOutcomeType}|undefined;
export declare function both():JsonFileWriteOutcomeType|JsonFileReadOutcomeType;
`
	return []string{
		h.write("nexus/source/structured-text/json/JsonFile.ts", jsonFile),
		h.write("nexus/source/structured-text/json/Json.ts", `export type JsonParseOutcomeType<T=unknown>={outcome:'Parsed';value:T}|{outcome:'Invalid';message:string};export declare function parse<T=unknown>():JsonParseOutcomeType<T>;`),
		h.write("nexus/source/structured-text/json/Jsonc.ts", `export type JsoncParseOutcomeType<T=unknown>={outcome:'Parsed';value:T}|{outcome:'Invalid';message:string};export declare function parse<T=unknown>():JsoncParseOutcomeType<T>;`),
		h.write("nexus/source/system/PackageJson.ts", `export type PackageJsonReadOutcomeType={outcome:'Read'}|{outcome:'Missing'}|{outcome:'Invalid';message:string};export declare function read():PackageJsonReadOutcomeType;`),
		h.write("lookalike/JsonFile.ts", jsonFile),
		h.write("outcome-controls.ts", `import {write,asyncWrite,success,extra,store,read,both} from './nexus/source/structured-text/json/JsonFile';
import {parse} from './nexus/source/structured-text/json/Json';import {parse as parseJsonc} from './nexus/source/structured-text/json/Jsonc';import {read as readPackage} from './nexus/source/system/PackageJson';import {write as own} from './lookalike/JsonFile';
write();(write());read<number>();parse<string>();parseJsonc();readPackage();extra();store?.save();both();
export async function f(){await asyncWrite();await (write());asyncWrite();void await asyncWrite();}
void write();const kept=write();success();own();export function wrapper(){return write();}wrapper();
`),
	}
}

func continuationSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.ts"))
	additional, _ := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.a"))
	files = append(files, additional...)
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
				h.t.Fatalf("nonunique %s mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_02_continuation_suite.a"), archive, false)
}

// Not parallel: all output, including failures and mutants, goes to files. This suite is serial
// because sanitizer archives and corpus runs share the limited scratch volume.
func TestWave02ContinuationAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE02_CONTINUATION_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_02_continuation_suite.a")
	binary := h.build(stage0, "coverage", entry, archive, false)
	oracle := volumeOracle(h, "coverage-oracle", "oracle_wave_02_continuation.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	paths := []string{}
	for i, source := range continuationControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}
	paths = append(paths, continuationOutcomeFiles(h)...)
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range continuationRuleNames {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "coverage-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Each rule mutant changes a real judgment or report range, finishes normally,
	// and is killed exclusively by the independent cohere diagnostic bytes.
	for _, change := range []struct{ name, file, from, to string }{
		{"collection", "no_collection_misuse.a", "return value <= 0;", "return value < 0;"},
		{"outcome", "no_discarded_outcome.a", "group.arms.length === group.count", "group.arms.length < group.count"},
		{"pure", "no_discarded_pure_result.a", "!declaration.defaultLibrary", "declaration.defaultLibrary"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			mutant := continuationSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
			got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
			if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
				t.Fatalf("%s mutant survived", change.name)
			}
			t.Logf("%s: exit 0, Go byte oracle catches byte %d; %s", change.name, firstDifference(got.stdout, truth.stdout), summary(got.stdout))
			if err := os.Remove(mutant); err != nil {
				t.Fatal(err)
			}
		})
	}
	// A raw union-size fact mutant compiles and exits normally; only Go finding bytes kill it.
	overlayAncestry := h.overlay("ancestry", "bridge/tsgo/checker/type_declaration_ancestry.go", "out.number(uint64(len(members)))", "out.number(uint64(len(members) + 1))")
	ancestryArchive := h.archive("ancestry", overlayAncestry, false)
	ancestry := h.build(stage0, "ancestry", entry, ancestryArchive, false)
	altered := h.must("ancestry-run", exec.Command(ancestry, config, manifest))
	if len(altered.stderr) != 0 || bytes.Equal(altered.stdout, truth.stdout) {
		t.Fatal("ancestry mutant survived")
	}
	t.Logf("ancestry mutant: exit 0, Go byte oracle catches byte %d; %s", firstDifference(altered.stdout, truth.stdout), summary(altered.stdout))
	os.Remove(ancestryArchive)
	os.Remove(ancestry)

	if manifest := os.Getenv("ADAMIC_WAVE02_CONTINUATION_REPOSITORY_MANIFEST"); manifest != "" {
		h.compare("repository", oracle, binary, filepath.Join(repository, "tsconfig.json"), manifest)
		h.compare("repository-asan", oracle, asan, filepath.Join(repository, "tsconfig.json"), manifest)
	}
	if manifest := os.Getenv("ADAMIC_WAVE02_CONTINUATION_COMPILER_MANIFEST"); manifest != "" {
		corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
		if corpus == "" {
			t.Fatal("compiler source is required")
		}
		h.compare("compiler", oracle, binary, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
		h.compare("compiler-asan", oracle, asan, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoInspect(program,file,0,1,'Identifier','raw-shape');tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','type-declaration-ancestry\n1'));
`)
	probe := h.write("probe.ts", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released program queried using a new question: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, caught by the required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
