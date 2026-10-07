package typeaware

import (
	"bytes"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func wave06Sources(t *testing.T, repository string) []string {
	t.Helper()
	unique := map[string]bool{}
	for _, file := range []string{"core/no_useless_return_test.go", "typescript/no_duplicate_type_constituents_test.go"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules", file), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			pair, ok := node.(*goast.KeyValueExpr)
			if !ok {
				return true
			}
			key, ok := pair.Key.(*goast.Ident)
			if !ok || (key.Name != "source" && key.Name != "sourceText") {
				return true
			}
			value, ok := pair.Value.(*goast.BasicLit)
			if !ok || value.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(value.Value)
			if err == nil && !strings.ContainsRune(text, 0) {
				unique[text] = true
			}
			return true
		})
	}
	for _, text := range []string{
		"class A {x=1}; class B {x=1};const a:A=new B();const plain:A={x:1};",
		"class Box<T>{value:T;constructor(value:T){this.value=value} put(value:T){this.value=value}};declare const box:Box<string>;const wider:Box<string|number>=box;",
		"class Base{x=1};class Derived extends Base{y=1};const b:Base=new Derived();class Chain{next():Chain{return this;}}",
		"class A{x=1};class B{x=1};declare const holder:{readonly value:B};const wide:{readonly value:A}=holder;declare const rows:B[];const list:A[]=rows;",
		"class A{x=1};class B{x=1};declare const x:B&{tag:string};const a:A=x;declare const anyValue:any;const ok:A=anyValue;",
		"class A{x=1};class B{x=1};declare const x:A|B;const a:A=x;declare const y:B;const t:A|{x:number}=y;",
		"class A{x=1};class B{x=1};function f<T extends B>(x:T):A{return x};function g(x:A=new B()):A{return new B()};",
		"class A{x=1};class B{x=1};declare const x:{p:B};declare const f:(a:A)=>B;const h:(a:B)=>A=f;",
		"class Empty{};const x:Empty=42;class A{x=1};const fresh:A=true?{x:1}:{x:2};",
		"class A{x=1};class B{x=1};declare const rows:Map<string,B>;const wide:Map<string,A>=rows;",
		"type U=undefined;function optional(p?:string|U|undefined){};type T=number|number;",
		"type T={a:1}|{a:1};type A='a';type B=A;type U=A|B|('b'|A);type I=number&string&(number&string);",
		"/* 世界 🌍 */\r\ntype T='é'|'é';function f(){/*keep*/return/*why*/;}\r\n",
		"function required():unknown{return;} function allowed():void{return;} async function asyncRequired():Promise<string|undefined>{return;} async function asyncAllowed():Promise<void>{return;} async function asyncUndefined():Promise<undefined>{return;}",
		"class Getter{get value():number|undefined{return;}};function* generator():Generator<number,undefined>{return;}",
		"function promiseObject():Promise<void>{return;} function embeddedVoid():{f:()=>void}{return;} function voidUnion():string|void{return;}",
		"function branch(x:boolean){if(x){return;}else{console.log(x);}}function dead(){throw 1;return;}function double(){return;return;}",
		"function loop(){while(true){} return;} function skip(){while(false){return;} return;}",
		"function trial(){try{return;}finally{console.log('x')}}function rescued(){try{return;}finally{console.log('x')}console.log('after')}",
		"function f(){try{if(true)return;console.log('same try')}catch{console.log('catch')}}",
		"function f(x:boolean){label:{if(x)return;break label;}console.log('later')}",
	} {
		unique[text] = true
	}
	sources := make([]string, 0, len(unique))
	for source := range unique {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	return sources
}

func wave06Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, base := range []string{"wave_06_suite.a", "nominal_class.a", "no_duplicate_type_constituents.a", "no_useless_return.a", "promised_shape.a"} {
		data, err := os.ReadFile(filepath.Join(h.repository, "stage1/cohere/typeaware", base))
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if base == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique mutant %s", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		// New .a files travel together; existing shared .ts dependencies remain read-only.
		for _, entry := range []string{"rules", "types", "type_fact", "facts", "checker_facts", "frames", "bindings", "diagnostic", "unary_minus", "repair", "returns", "shadow"} {
			source = strings.ReplaceAll(source, "'./"+entry+".ts'", "'"+filepath.Join(h.repository, "stage1/cohere/typeaware", entry+".ts")+"'")
		}
		if err := os.WriteFile(filepath.Join(directory, base), []byte(source), 0644); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_06_suite.a"), archive, false)
}

// Not parallel: native artifacts, sanitizer archives and timings share scratch.
func TestWave06AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE06_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_06_suite.a")
	binary := h.build(stage0, "wave06", entry, archive, false)
	oracle := volumeOracle(h, "wave06-oracle", "oracle_wave_06.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	paths := []string{}
	for i, source := range wave06Sources(t, repository) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"adamic/nominal-class", "@typescript-eslint/no-duplicate-type-constituents", "no-useless-return"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("missing positive control %s", name)
		}
	}
	t.Logf("controls: %d files", len(paths))
	if os.Getenv("ADAMIC_WAVE06_INITIAL") == "1" {
		return
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave06-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"nominal", "nominal_class.a", "if(metadata.classSymbol() === targetSymbol)", "if(metadata.classSymbol() !== targetSymbol)"},
		{"duplicate", "no_duplicate_type_constituents.a", "this.rules.byte(node.end));", "this.rules.byte(node.end) + 1);"},
		{"return", "no_useless_return.a", "finding.fixEnd = finding.end;", "finding.fixEnd = finding.end + 1;"},
	} {
		mutant := wave06Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, only Go bytes catch byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	if os.Getenv("ADAMIC_WAVE06_CORPORA") == "1" {
		corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
		if corpus == "" {
			t.Fatal("missing compiler corpus")
		}
		pin := h.must("compiler-pin", exec.Command("git", "-C", corpus, "rev-parse", "HEAD"))
		if strings.TrimSpace(string(pin.stdout)) != compilerCommit {
			t.Fatal("wrong compiler pin")
		}
		for _, population := range []struct{ name, config, base, manifest string }{
			{"compiler", filepath.Join(corpus, "src/compiler/tsconfig.json"), corpus, "compiler.manifest"},
			{"repository", filepath.Join(repository, "tsconfig.json"), repository, "repository.manifest"},
		} {
			data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.manifest))
			if err != nil {
				t.Fatal(err)
			}
			roots := []string{}
			for _, path := range strings.Split(string(data), "\n") {
				if path != "" {
					roots = append(roots, filepath.Join(population.base, path))
				}
			}
			input := h.write(population.manifest, strings.Join(roots, "\n")+"\n")
			h.compare(population.name, oracle, binary, population.config, input)
			h.compare(population.name+"-asan", oracle, asan, population.config, input)
			// A single isolated observation with complete output; builds are outside it.
			command := exec.Command(binary, population.config, input)
			command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			native := h.must(population.name+"-timed-native", command)
			goResult := h.must(population.name+"-timed-go", exec.Command(oracle, population.config, input))
			if !bytes.Equal(native.stdout, goResult.stdout) {
				t.Fatal("timed diagnostic mismatch")
			}
			t.Logf("%s native %.6fs Go %.6fs ratio %.3f; native %s; Go %s", population.name, native.elapsed.Seconds(), goResult.elapsed.Seconds(), native.elapsed.Seconds()/goResult.elapsed.Seconds(), strings.TrimSpace(string(native.stderr)), strings.TrimSpace(string(goResult.stderr)))
		}
	}
	probeText := "function f():void{return;}"
	released := h.write("released.a", fmt.Sprintf(`import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,%d,'FunctionDeclaration','promised-shape'));
`, len(probeText)))
	probe := h.write("probe.a", probeText)
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped %v %s", got.err, got.stderr)
	}
	t.Log("released-handle query: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	if len(got.stderr) != 0 {
		t.Fatalf("released-registry mutant did not finish normally: %s", got.stderr)
	}
	t.Log("released-registry mutant exits 0, caught by required panic 70")

	questionOverlay := h.overlay("promised-question", "bridge/tsgo/checker/promised_shape.go",
		"subject = c.GetPromisedTypeOfPromise(subject)", "subject = c.GetTypeFromTypeNode(node.Type())")
	questionArchive := h.archive("promised-question", questionOverlay, false)
	questionMutant := h.build(stage0, "promised-question-native", entry, questionArchive, false)
	questionResult := h.must("promised-question-run", exec.Command(questionMutant, config, manifest))
	if len(questionResult.stderr) != 0 || bytes.Equal(questionResult.stdout, truth.stdout) {
		t.Fatal("promised-shape mutant survived")
	}
	t.Logf("promised-shape mutant: exit 0, only Go byte oracle catches byte %d", firstDifference(questionResult.stdout, truth.stdout))
}

func TestWave06PinnedFlags(t *testing.T) {
	if checker.TypeFlagsUndefined != 4 || checker.TypeFlagsVoid != 16 || checker.TypeFlagsNever != 262144 || checker.TypeFlagsTypeParameter != 524288 {
		t.Fatal("update wave 06 pinned flags and rerun the independent cohere byte oracle")
	}
}
