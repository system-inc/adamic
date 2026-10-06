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

var coverageRuleNames = []string{
	"no-use-before-define", "adamic/no-unchecked-cast", "@typescript-eslint/method-signature-style", "no-implicit-coercion",
	"nexus/correctness-no-caller-data-mutation", "no-param-reassign", "adamic/invariant-mutable", "adamic/no-optional-widening",
	"nexus/consistency-no-property-alias", "@typescript-eslint/no-unused-vars",
}

func coverageControls() []string {
	return []string{
		"export const before=later;const later=1; function f(p:number){p=2;return p;} export {f};",
		"declare const unknownValue:unknown;export const unchecked=unknownValue as {p:number};",
		"declare const lit:0|1|null;export const test=!!lit;",
		"declare const anyValue:any;export const anyCast=anyValue as number;export const unknownCast=anyValue as unknown;",
		"declare const shape:{kind:'a';value:number}|{kind:'b';value:string};export const narrow=shape as {kind:'a';value:number};",
		"class A {a=1};class B {b=1};declare const ab:A|B;export const a=ab as A;",
		"export interface Methods { m(x:number):string; overloaded(x:string):number;overloaded(x:number):string; }",
		"export interface ThisMethods {m():this;readonly p:number} declare module 'external' {interface C {m():number}}",
		"export function coerce(n:number,s:string|number){return [!!n,!!s,+s,s*1,~s.indexOf('x')];}",
		"export function coercions(s:string,n:number){return [''+n,n+'',s-0,-(-s)];}",
		"function Boolean(x:unknown){return x};export function shadow(n:number){return !!n;}",
		"export function mutation(p:{x:number;rows:number[]}){p.x++;p.rows.push(1);p.rows[0]=2;return p;}",
		"export function copied(p:{x:number}){p={...p};p.x=2;return p;} export function callbacks(rows:{x:number}[]){rows.forEach(p=>{p.x=2;});}",
		"export function setters(p:Host){p.x=2;}class Host {set x(v:number){console.log(v)}}",
		"/** @processState shared counter */ export interface State {x:number};export function update(p:State){p.x++;}",
		"/** @processState */ export interface Untagged {x:number};export function update(p:Untagged){p.x++;}",
		"interface RecordData{x:number};interface Fill {/** @mutates entity filled by contract */ run(entity:RecordData):void};export class Writer implements Fill {run(renamed:RecordData){renamed.x=1;}}",
		"export class Writer {/** @mutates entity */ run(entity:{x:number}){entity.x=1;}/** @mutates stale because */ bad(entity:{x:number}){entity.x=2;}}",
		"interface Animal {name:string};interface Dog extends Animal {bark():void};declare const dogs:Dog[];export const animals:Animal[]=dogs;export const readonlyAnimals:readonly Animal[]=dogs;",
		"interface A {x:number};interface D extends A {y:number};declare const map:Map<string,D>;export const wide:Map<string,A>=map;export const read:ReadonlyMap<string,A>=map;",
		"interface A{x:number};interface D extends A{y:number};declare const tuple:[D];export const widened:[A]=tuple;export const read:readonly[A]=tuple;",
		"interface A{x:number};interface D extends A{y:number};declare const f:()=>D;export const object:{f:()=>A}={f};declare const holder:{f:()=>D};export const widen:{f:()=>A}=holder;",
		"declare const narrow:{x:number};export const wide:{x:number;y?:number}=narrow;const fresh={x:1};export const exact:{x:number;y?:number}=fresh;",
		"declare const n:never;export const impossible:{a:number;first?:string}|{a:number;second?:number}=n;",
		"export function alias(o:{x:number}){const x=o.x;return x;}export function cached(o:{get:()=>{x:number}}){const x=o.get().x;return x;}",
		"export function snapshot(o:{x:number}){const x=o.x;o.x++;return x;}export function reassigned(o:{x:number}){let x=o.x;x++;return x;}",
		"class Getter{get x(){return 1}};export function cached(o:Getter){const x=o.x;return x;}",
		"export function narrowed(o:{x:number|undefined}){if(o.x){const x=o.x;return ()=>x;}}",
		"export function optional(o:{x:number}|undefined){const x=o?.x;return x;}export function annotation(o:{x:any}){const x:unknown=o.x;return x;}",
		"const unused=1;let written=1;written=2;export function params(a:number,b:number){return a;}export function identity(x:number){return x;} const arrow=(x:number)=>x;export {arrow};",
		"function recursive(){return recursive();}type Self={next:Self};export const pass=(x:number)=>x;export function defaults(a:number,b=1){return a;}",
		"const onlyType={};export type Shape=typeof onlyType; export function pred(x:unknown):x is string{return false;}",
		"import {Used,Unused,AlsoUnused} from './helper.js';export {Used};",
		"interface A{x:number};interface D extends A{y:number};declare const rows:D[];export class Holder{#untyped=1;['tag']=1;'plain'=0;123=1;#wide:A[]=rows;readonly #narrow:D[]=rows;}",
		"/* 世界 🌍 */\r\nexport function unicode(é:number){é=2;const 漢={x:1};const x=漢.x;return x;}\r\n",
	}
}

func coverageSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.ts"))
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
	return h.build(stage0, name, filepath.Join(directory, "coverage_suite.ts"), archive, false)
}

// All output, including failures and mutants, goes to files. This suite is serial
// because sanitizer archives and corpus runs share the limited scratch volume.
func TestCoverageAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_COVERAGE_ARTIFACTS"); path != "" {
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/coverage_suite.ts")
	binary := h.build(stage0, "coverage", entry, archive, false)
	oracle := volumeOracle(h, "coverage-oracle", "oracle_coverage.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	paths := []string{}
	for i, source := range coverageControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("helper.ts", "export const Used=1,Unused=2,AlsoUnused=3;\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range coverageRuleNames {
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
		{"before", "before.ts", "if(!above && !initializing)", "if(!initializing)"},
		{"cast", "casts.ts", "(t.flags & 3) !== 0", "(t.flags & 3) === 0"},
		{"methods", "methods.ts", "this.visit(index);", "// Mutant drops method-signature judgments."},
		{"coercion", "coercion.ts", "const value = frames.field();", "const value = ['false'].join(''); frames.field();"},
		{"caller", "caller.ts", "(symbol.flags & 65536) === 0", "(symbol.flags & 32768) === 0"},
		{"parameter", "reassign.ts", "anchored = true;", "anchored = false;"},
		{"invariant", "flow.ts", "!this.assignable(pair.t.id, pair.s.id)", "!this.assignable(pair.s.id, pair.t.id)"},
		{"optional", "flow.ts", "this.missingName = pair.property;", "this.missingName = 'wrong';"},
		{"alias", "property_alias.ts", "this.rules.byte(node.end), '')", "this.rules.byte(node.end) + 1, '')"},
		{"unused", "unused.ts", "node.kind === 'ArrowFunction' ? -1 : this.bindings.shape.name(current)", "this.bindings.shape.name(current)"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			mutant := coverageSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
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
	// New raw questions are also independently held to their direct checker tests.
	if manifest := os.Getenv("ADAMIC_COVERAGE_REPOSITORY_MANIFEST"); manifest != "" {
		h.compare("repository", oracle, binary, filepath.Join(repository, "tsconfig.json"), manifest)
		h.compare("repository-asan", oracle, asan, filepath.Join(repository, "tsconfig.json"), manifest)
	}
	if manifest := os.Getenv("ADAMIC_COVERAGE_COMPILER_MANIFEST"); manifest != "" {
		corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
		if corpus == "" {
			t.Fatal("compiler source is required")
		}
		h.compare("compiler", oracle, binary, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
		h.compare("compiler-asan", oracle, asan, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
	}
	released := h.write("released.ts", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','node-symbol-details'));
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
