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

func wave10Controls() []string {
	return []string{
		`let foo=0; while(foo) {} foo=1;`,
		`let foo=0,bar=2; while(foo<bar) {} foo=1;`,
		`let foo=0,bar=2; while(foo<bar) {foo++;} while(foo&&bar){bar++;}`,
		`let foo=0,bar=2,baz=3; while(foo?bar:baz){foo++;} while(foo===read(bar)){} function read(n:number){return n;}`,
		`let foo=0;while(foo){update();} function update(){foo++;}`,
		`let foo=0;while(foo){update(0);} function update(foo:number){foo++;}`,
		`for(var foo=0;foo<2;){}foo=1;let bar=0;for(;bar<2;bar++){} do{}while(foo);`,
		`var foo=0;while(foo){var foo=1;} let bar=0;while(bar){let bar=1;}`,
		`let foo=0;async function f(){while(foo){await tick();}} function update(){foo=1;}update();declare function tick():Promise<void>;`,
		`let foo=0;async function f(){while(foo){await tick();}const update=()=>{foo=1;};}declare function tick():Promise<void>;`,
		`type Object=object|undefined;type Bottom=never;type Alias=Bottom|string;type T=string|any;type U=number|never;type V=string|unknown;type W=any&number;type X=never&string;type Y=unknown&string;`,
		`type A=string|'lit';type B=number|42;type C=boolean|false;type D=bigint|12n;type E=string&'x';type F=number&42;`,
		`type B=boolean;type F=false;type A=B|false;type C=F&boolean;type S='a'|'b';type T=S|string;type U=S&string;`,
		`type T=((string|any))|number;type U=(string)|(any);type X=(string|number)&'x';`,
		`declare function f():never|string;type F=()=>never|string;type T=never|string;type B=string|NotKnown;const arrow=():never|string=>"s";`,
		"type Template=`x${string}`;type T=Template|string;type I=Template&string;",
		`declare const a:string;declare const b:string;export const checks=[a.indexOf(b)!==-1,a.indexOf(b)===-1,a.indexOf(b)>=0,a.indexOf(b)<0,a.indexOf(b)>-1,a.indexOf(b)<=-1];`,
		`declare const a:number[];export const checks=[a.indexOf(1)!=-1.0,a.indexOf(2)==-0x1,a.indexOf(2)===0,a.indexOf(2)+1];`,
		`declare const a:string|undefined;export const checks=[a?.indexOf('a')!==-1,a?.indexOf('a')===-1];`,
		`declare const a:{indexOf(x:unknown):number;includes(x:any):boolean};export const b=a.indexOf(1)!==-1;`,
		`declare const a:{indexOf(x:unknown):number;includes(x:unknown):boolean};export const b=a.indexOf(1)!==-1;`,
		`declare const a:{indexOf(x:unknown):number;includes:boolean};export const b=a.indexOf(1)!==-1;`,
		`declare const a:string;export const checks=[/bar/.test(a),/b\.r/.test(a),/\x41/.test(a),/a\-b/.test(a),/ba[rz]/.test(a),/^bar/.test(a),/bar$/ .test(a),/bar/i.test(a),/bar/g.test(a),/foo|bar/.test(a)];`,
		`declare const text:string;declare const unknownPattern:RegExp;export const no=unknownPattern.test(text);const empty=/(?:)/;export const skip=empty.test(text);`,
		`declare const a:string;const pattern=/bar/;const ctor=new RegExp('baz');export const checks=[pattern.test(a),ctor.test(a),/\n\r\t\v\f\0/.test(a),/\u{1f30d}/u.test(a),/\u0041/.test(a)];`,
		`declare const a:string,b:string;export const checks=[/bar/.test((1+1,a)),/bar/.test((a)),/bar/.test(true?a:b),/bar/?.test(a),/bar/['test'](a)];`,
		`export function f<T extends string>(a:T){return /bar/.test(a);}declare const untyped:any;export const no=/bar/.test(untyped);`,
		"/* 世界 🌍 */\r\nlet é=0;while(é){}é=1;export type T='漢'|string;declare const text:string;export const yes=text.indexOf('🌍')===-1;\r\n",
	}
}

func wave10Mutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_10_suite.a"), archive, false)
}

// Not parallel: archives and native sanitizer binaries share bounded scratch storage.
func TestWave10AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE10_ARTIFACTS"); path != "" {
		directory = path
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_10_suite.a")
	binary := h.build(stage0, "wave10", entry, archive, false)
	oracle := volumeOracle(h, "wave10-oracle", "oracle_wave_10.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	paths := []string{}
	for i, source := range wave10Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-unmodified-loop-condition", "@typescript-eslint/no-redundant-type-constituents", "@typescript-eslint/prefer-includes"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave10-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"loop", "no_unmodified_loop_condition.a", "other.group === condition.group && other.modified", "other.group === condition.group && !other.modified"},
		{"redundant", "no_redundant_type_constituents.a", "part.flags === 262144 && !this.returnType(index)", "part.flags === 262144 && this.returnType(index)"},
		{"includes", "prefer_includes.a", "if(negative) {", "if(positive) {"},
	} {
		mutant := wave10Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE10_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE10_COMPILER_MANIFEST")},
	} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','member-parameters'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released member-parameters query: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, caught by required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
