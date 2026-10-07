package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func wave09Controls() []string {
	return []string{
		"let a=1;let b=1;b++;let c; c=1; let d;{d=1;} export {a,b,c,d};",
		"let a=1; {let a=2;a++;} let b=2; function f(){b++;} export {a,b,f};",
		"let {a,b}={a:1,b:2};b++;let [c,d]=[1,2];export {a,b,c,d};",
		"let [,a,,b]=[1,2,3,4];export {a,b};",
		"for(let x of [1,2]){console.log(x);}for(let i=0,end=10;i<end;i++){}",
		"let a;console.log(a);a=1;let b;function f(){console.log(b);} b=2;export {f};",
		"let a,b;({a,b}={a:1,b:2});let c;[c]=[1];let rest;[...rest]=[1];",
		"let a=1;[...(a)]=[];let b=1;({files=b}={});export {a,b};",
		"parseInt();parseInt('10');parseInt('10',);Number.parseInt('10');Number['parseInt']('10');Number[`parseInt`]('10');",
		"parseInt('10',1);parseInt('10',36);parseInt('10',37);parseInt('10',10.5);parseInt('10',+37);parseInt('10',-10);parseInt('10',0x10);parseInt('10',1.6e1);parseInt('10',undefined);",
		"function f(parseInt:any,Number:any,undefined:any){parseInt();Number.parseInt();globalThis.parseInt('10',undefined);}export {f};",
		"parseInt('10',null);parseInt('10',true);parseInt('10','10');parseInt('10',10n);declare const args:any[];parseInt(...args);parseInt('10',...args);parseInt('10',1,...args);",
		"declare const x:string|null|undefined;const a=x as string;const b=(x) as string;const c=<string>x;const d=x as string|null;export {a,b,c,d};",
		"export function generic<T>(values:(T|undefined)[]){return values[0] as T;} export function constrained<T extends string>(values:(T|undefined)[]){return values[0] as T;}",
		"interface X{p:number};declare const x:X|null;const a=x as X;declare const y:number|string|null;const b=y as number;const c=y as number|string;export {a,b,c};",
		"declare const a:any;declare const u:unknown;const b=a as string;const c=u as string;const d=[] as const;export {b,c,d};",
		"declare const p:()=>string|null;declare const c:boolean;const a=(c?p():p()) as string;const b=p() as string;export {a,b};",
		"/* 世界 🌍 */\r\nlet é=1;parseInt('10',);declare const 漢:string|null;const a=漢 as string;export {é,a};\r\n",
	}
}

func wave09Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, extension := range []string{"*.a"} {
		paths, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware", extension))
		if err != nil {
			h.t.Fatal(err)
		}
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				h.t.Fatal(err)
			}
			source := string(data)
			if filepath.Base(path) == file {
				if strings.Count(source, from) != 1 {
					h.t.Fatal("nonunique mutant")
				}
				source = strings.Replace(source, from, to, 1)
			}
			source = regexp.MustCompile(`from '\./([^']+\.ts)'`).ReplaceAllString(source, "from '"+filepath.Join(h.repository, "stage1/cohere/typeaware")+"/$1'")
			source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
			source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
			if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0644); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_09_suite.a"), archive, false)
}

// Not parallel: sanitizer archives and corpus timings share the scratch disk and CPU quota.
func TestWave09AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE09_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_09_suite.a")
	binary := h.build(stage0, "wave09", entry, archive, false)
	oracle := volumeOracle(h, "wave09-oracle", "oracle_wave_09.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range wave09Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"prefer-const", "radix", "@typescript-eslint/non-nullable-type-assertion-style"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("missing positive control %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave09-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"prefer-const", "prefer_const.a", "writes.length === 0 ? name : -1", "writes.length >= 0 ? name : -1"},
		{"radix", "radix.a", "value <= 36", "value <= 35"},
		{"type-parameter-flag", "non_nullable_type_assertion_style.a", "t.flags & 524288", "t.flags & 262144"},
		{"non-nullable", "non_nullable_type_assertion_style.a", "finding.fixEnd = r.byte(node.end);", "finding.fixEnd = r.byte(node.end) + 1;"},
	} {
		mutant := wave09Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("mutant survived: %s", change.name)
		}
		t.Logf("%s mutant: exit 0, Go byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE09_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE09_COMPILER_MANIFEST")},
	} {
		if corpus.manifest == "" {
			t.Logf("%s corpus not configured", corpus.name)
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		got := h.must(corpus.name+"-timed-native", exec.Command(binary, corpus.config, corpus.manifest))
		want := h.must(corpus.name+"-timed-go", exec.Command(oracle, corpus.config, corpus.manifest))
		if !bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("timing outputs differ")
		}
		t.Logf("%s complete output: native %s, Go %s, %s", corpus.name, got.elapsed, want.elapsed, summary(got.stdout))
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','declaration-file-flags'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, caught by required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
