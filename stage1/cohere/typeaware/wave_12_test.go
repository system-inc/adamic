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

func wave12Controls() []string {
	return []string{
		"export {}; var a=1; var a=2; function f(){var x=1;if(true){var x=2;}}",
		"export {}; class A{} class A{} namespace A{}; interface I{} interface I{}; function f():number;function f(){return 1};",
		"export {}; type T=1;const T=2; enum E{A}enum E{B}; function F(){}function F(){}",
		"export {};function a(){let x=1;}function b(){let x=2;} namespace N{var x=1;var x=2;} class C{static{let x=1;let x=2;}}",
		"export {};var a=1;var {a,b:c=1}={a:2,b:3};var c=1;function f(a:number,a:number){return a};",
		"const re=/x/g;export function test(s:string){return re.test(s)}",
		"export function test(s:string){const re=/x/g;return re.test(s)}",
		"const re=/x/g;export function test(s:string){re.lastIndex=0;return re.test(s)}",
		"const re=/x/gy;export function test(s:string){re.lastIndex=2;return re.test(s)}",
		"const re=/x/g;export function test(s:string){while(re.test(s)){};do{}while(re.test(s));for(;re.test(s);){};}",
		"const re=new RegExp('x','g');export function test(s:string){return re.test(s)}",
		"const re=/x/g;export function test(s:string){for(let i=0;i<3;i++){re.test(s)}}",
		"export function test(s:string){for(let i=0;i<3;i++){const re=/x/g;re.test(s)}}",
		"export class C{readonly re=/x/g;test(s:string){return this.re.test(s)}}",
		"export class C{re=/x/g;test(s:string){return this.re.test(s)}}",
		"const fake={test(s:string){return true}};export function test(s:string){return fake.test(s)}",
		"export function f(){const a:number[]=[];a.push(1);const m=new Map<string,number>();m.set('x',1);const s=new Set<number>();s.add(1);}",
		"export function f(){const a:number[]=[];a[0]=1;const s=new Set<number>();s.clear();}",
		"export function f(){const a:number[]=[];a.push(1);return {a};}export function g(){const m=new Map<string,number>();m.set('x',1);return m.get('x')}",
		"export function f(){const a:number[]=[];const x=a.push(1);const b:number[]=[];const cb=()=>b.push(1);return [x,cb];}",
		"export function f(){const a:number[]=[];[1].forEach(x=>{a.push(x)});}",
		"export function f(){const a:number[]=[];function g(){const a:number[]=[];return a};a.push(1)}",
		"export function f(){const a:number[]=[];a.push(1);type T=typeof a;return 0}",
		"export {}; /* 世界 🌍 */\r\nvar é=1;var é=2;const re=/x/g;export function f(){return re.test('x')}\r\n",
		"export function f(){const a=([] as number[]) satisfies number[];(a.push(1));}",
		"export class C{static{const a:number[]=[];a.push(1)}; p=(()=>{const a:number[]=[];a.push(1);return 0})()}",
		"export {};const re=/x/g;for(const x of [re.test('x')]){};for(re.test('x');false;){};for(;false;re.test('x')){};",
	}
}

func wave12Mutant(h *harness, stage0, archive, file, from, to string) string {
	directory := filepath.Join(h.directory, file+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*"))
	if err != nil {
		h.t.Fatal(err)
	}
	for _, path := range files {
		if filepath.Ext(path) != ".a" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filepath.Base(path) == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatal("nonunique mutant", file)
			}
			source = strings.Replace(source, from, to, 1)
		}
		// Existing .ts modules remain in the repository; new mutant sources stay .a.
		for _, old := range []string{"bindings", "caller", "diagnostic", "frames", "rules", "unary_minus"} {
			source = strings.ReplaceAll(source, "'./"+old+".ts'", "'"+filepath.Join(h.repository, "stage1/cohere/typeaware", old+".ts")+"'")
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0644); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, file+"-mutant", filepath.Join(directory, "wave_12_suite.a"), archive, false)
}

// Not parallel: archive builds, sanitizers and timings share scratch resources.
func TestWave12AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE12_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_12_suite.a")
	binary := h.build(stage0, "wave12", entry, archive, false)
	oracle := volumeOracle(h, "wave12-oracle", "oracle_wave_12.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range append(wave12Controls(), wave12ReferenceControls(t, repository)...) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"@typescript-eslint/no-redeclare", "nexus/correctness-no-test-on-global-regex", "nexus/correctness-no-write-only-collection"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatal("missing positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave12-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, m := range []struct{ file, from, to string }{
		{"no_redeclare.a", "survivors.slice(1)", "survivors.slice(2)"},
		{"no_test_on_global_regex.a", "!this.flags(creation).includes('g')", "!this.flags(creation).includes('y')"},
		{"no_write_only_collection.a", "rules.byte(node.end), '')", "rules.byte(node.end) + 1, '')"},
	} {
		mutant := wave12Mutant(h, stage0, archive, m.file, m.from, m.to)
		got := h.must(m.file+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) > 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", m.file)
		}
		t.Logf("%s mutant exits 0, empty stderr, byte oracle catches byte %d", m.file, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
	}
	for _, corpus := range []string{"REPOSITORY", "COMPILER"} {
		roots := os.Getenv("ADAMIC_WAVE12_" + corpus + "_MANIFEST")
		if roots == "" {
			continue
		}
		cfg := filepath.Join(repository, "tsconfig.json")
		if corpus == "COMPILER" {
			cfg = filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")
		}
		h.compare(strings.ToLower(corpus), oracle, binary, cfg, roots)
		h.compare(strings.ToLower(corpus)+"-asan", oracle, asan, cfg, roots)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','global-symbol-details'));`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released handle panics 70; retaining registry mutant exits 0 and is caught")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
