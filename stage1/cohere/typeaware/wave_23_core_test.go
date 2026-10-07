package typeaware

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func wave23CoreControls(t *testing.T, repository string) []string {
	t.Helper()
	seen := map[string]bool{}
	var sources []string
	add := func(source string) {
		if !seen[source] {
			seen[source] = true
			sources = append(sources, source)
		}
	}
	for _, name := range []string{"no_eval", "no_extend_native", "no_func_assign"} {
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/core", name+"_test.go"), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			lit, ok := node.(*ast.CompositeLit)
			if !ok || lit.Type != nil || len(lit.Elts) < 2 {
				return true
			}
			value, ok := lit.Elts[0].(*ast.BasicLit)
			if !ok || value.Kind != token.STRING {
				return true
			}
			if second, ok := lit.Elts[1].(*ast.BasicLit); ok && second.Kind == token.STRING {
				value = second
			}
			text, err := strconv.Unquote(value.Value)
			if err != nil {
				t.Fatal(err)
			}
			add(text)
			return true
		})
	}
	for _, source := range []string{
		"eval(eval);eval?.(eval);(eval)('x');(eval as any)('x');(eval satisfies Function)('x');eval!('x');(<Function>eval)('x');",
		"const value={key:eval};enum E{member=eval}function f(p=eval){return p}class C{x=eval}const {x=eval}=Object;",
		"function eval(x:string){return x}eval('x');const alias=eval;const shorthand={eval};",
		"window.window.window.eval('x');window['window'][`eval`]('x');window.global.eval('x');this.eval('x');global.global.eval('x');globalThis.globalThis.eval('x');",
		"function f(window:any){window.eval('x')}const key='eval';window[key]('x');window[('eval')]('x');",
		"Array.prototype.x=0;Array.prototype.x+=1;Array.prototype.x++;delete Array.prototype.x;Array.prototype.x.y=1;Array.prototype={};",
		"Object.defineProperty(Array.prototype,'x',{});Object['defineProperties'](String['prototype'],{});Reflect.defineProperty(Array.prototype,'x',{});",
		"function f(Object:any,Array:any){Object.defineProperty(Array.prototype,'x',{});Array.prototype.x=1}",
		"declare const Array:any;Array.prototype.x=1;",
		"function foo(){}foo=1;foo++;({x:foo}={x:1});[foo]=[1];for(foo of [1]){}for(foo in {}){}",
		"function foo(){}function bar(foo:any){foo=1}const obj={foo};obj.foo=1;function baz(){let foo:any;foo=1}",
		"const f=function named(){named=1;named++};function foo():void;function foo(){}foo=1;",
		"function foo(){}({x=foo}={});({foo=1}={});[...foo]=[1];({x:(foo)}={x:1});",
		"/* 世界 🌍 */\r\nfunction é(){}é=1;Array.prototype.é=0;eval('é');\r\n",
	} {
		add(source)
	}

	for _, name := range []string{"AggregateError", "Array", "ArrayBuffer", "Atomics", "BigInt", "BigInt64Array", "BigUint64Array", "Boolean", "DataView", "Date", "Error", "EvalError", "FinalizationRegistry", "Float16Array", "Float32Array", "Float64Array", "Function", "Infinity", "Int16Array", "Int32Array", "Int8Array", "Intl", "Iterator", "JSON", "Map", "Math", "NaN", "Number", "Object", "Promise", "Proxy", "RangeError", "ReferenceError", "Reflect", "RegExp", "Set", "SharedArrayBuffer", "String", "Symbol", "SyntaxError", "TypeError", "URIError", "Uint16Array", "Uint32Array", "Uint8Array", "Uint8ClampedArray", "WeakMap", "WeakRef", "WeakSet", "parseFloat", "parseInt", "escape", "decodeURI", "isNaN", "undefined", "globalThis", "x", "Foo"} {
		add(name + ".prototype.p=0;")
	}
	for _, operator := range []string{"=", "+=", "-=", "*=", "/=", "%=", "**=", "<<=", ">>=", ">>>=", "&=", "|=", "^=", "&&=", "||=", "??="} {
		add("function foo(){}foo" + operator + "1;Array.prototype.x" + operator + "1;")
	}
	add("var foo:any;function foo(){}foo=1;")
	sort.Strings(sources)
	return sources
}

func wave23CoreMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_23_core_suite.a"), archive, false)
}

// Not parallel: builds, sanitizer archives and measured runs share one machine.
func TestWave23CoreAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE23_CORE_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_23_core_suite.a")
	binary := h.build(stage0, "native", entry, archive, false)
	oracle := volumeOracle(h, "wave23-core-oracle", "oracle_wave_23_core.go")
	h.write("prelude.d.ts", "")
	config := h.write("controls-tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ESNext"],"noEmit":true},"files":["prelude.d.ts"]}`)
	var paths []string
	for i, source := range wave23CoreControls(t, repository) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-eval", "no-extend-native", "no-func-assign"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatal("no positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "native-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for i, options := range [][]string{{"--allow-indirect-eval"}, {"--except-native=Array", "--except-native=Object"}, {"--except-native=unknown"}, {"--allow-indirect-eval", "--except-native=String", "--except-native=Array"}} {
		arguments := append([]string{config, manifest}, options...)
		expected := h.must(fmt.Sprintf("options-%d-go", i), exec.Command(oracle, arguments...))
		for _, entry := range []struct{ name, binary string }{{"native", binary}, {"asan", asan}} {
			got := h.must(fmt.Sprintf("options-%d-%s", i, entry.name), exec.Command(entry.binary, arguments...))
			if !bytes.Equal(got.stdout, expected.stdout) {
				t.Fatalf("options %v bytes differ at %d", options, firstDifference(got.stdout, expected.stdout))
			}
			t.Logf("options %v %s: %d identical bytes", options, entry.name, len(got.stdout))
		}
	}
	for _, change := range []struct{ name, file, from, to string }{
		{"eval", "no_eval.a", "this.report(callee);", "this.report(this.unwrap(callee));"},
		{"extend", "no_extend_native.a", "message.replace('%s', builtin)", "message.replace('%s', 'Object')"},
		{"func", "no_func_assign.a", "this.anchors.includes(declaration)", "true"},
	} {
		mutant := wave23CoreMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}

	// Custom declarations prove that "global" means declaration file, not default library.
	h.write("custom-globals.d.ts", "declare function eval(source:string):any;declare const Object:{prototype:any;defineProperty(...args:any[]):any;defineProperties(...args:any[]):any};declare const Array:{prototype:any};\n")
	customConfig := h.write("custom-tsconfig.json", `{"compilerOptions":{"strict":true,"noLib":true,"noEmit":true},"files":["custom-globals.d.ts"]}`)
	customRoot := h.write("custom.a", "const value=eval;eval('x');Array.prototype.x=1;Object.defineProperty(Array.prototype,'x',{});function foo(){}foo=1;export {};\n")
	customManifest := h.write("custom.manifest", customRoot+"\n")
	customTruth := h.compare("custom-declarations", oracle, binary, customConfig, customManifest)
	for _, name := range []string{"no-eval", "no-extend-native", "no-func-assign"} {
		if !bytes.Contains(customTruth.stdout, []byte("\t"+name+"\t")) {
			t.Fatal("missing custom declaration positive", name)
		}
	}
	h.compare("custom-declarations-asan", oracle, asan, customConfig, customManifest)
	for _, change := range []struct {
		name, file, from, to string
		options              []string
	}{
		{"eval-option", "no_eval.a", "if(this.allowIndirect && node.children.some", "if(false && node.children.some", []string{"--allow-indirect-eval"}},
		{"extend-option", "no_extend_native.a", "!this.exceptions.includes(node.text)", "true", []string{"--except-native=Array", "--except-native=Object"}},
	} {
		arguments := append([]string{config, manifest}, change.options...)
		expected := h.must(change.name+"-go", exec.Command(oracle, arguments...))
		mutant := wave23CoreMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, arguments...))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, expected.stdout) {
			t.Fatal("option mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, expected.stdout))
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
console.log(tsgoInspect(program,file,0,1,'Identifier','node-symbol-details'));
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
