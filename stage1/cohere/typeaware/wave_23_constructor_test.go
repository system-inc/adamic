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

func wave23ConstructorControls(t *testing.T, repository string) []string {
	t.Helper()
	seen := map[string]bool{}
	var sources []string
	add := func(source string) {
		if !seen[source] {
			seen[source] = true
			sources = append(sources, source)
		}
	}
	for _, name := range []string{"no_new_func", "no_new_native_nonconstructor", "no_new_wrappers"} {
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
		"new Function;Function();(Function)();new ((Function))('x');Function?.('x');",
		"Function.call(null,'x');Function.apply(null,['x']);Function.bind(null,'x')();new Function.call(null,'x');",
		"Function['call']();Function[`apply`]();Function[(('bind'))]();(Function).call();(Function?.call)?.();",
		"Function[0]();Function[method]();Function['cal'+'l']();Function.toString();Function.bind;globalThis.Function();",
		"class Function{}new Function;Function();Function.call();Function.bind()();Function.apply();",
		"function f(Function:any){new Function;Function();Function.call()}function g(){class Function{}new Function}new Function;",
		"new Symbol();new BigInt(1);new (Symbol)();new ((BigInt))(1);Symbol();BigInt(1);new globalThis.Symbol();",
		"new String('x');new ((Number))(1);new Boolean(false);String('x');Number(1);Boolean(false);new Object();",
		"function f(Symbol:any,BigInt:any,String:any,Number:any,Boolean:any){new Symbol;new BigInt;new String;new Number;new Boolean}",
		"import Function from './constructor-helper.a';import {Symbol,BigInt,String,Number,Boolean} from './constructor-helper.a';new Function;Function.call();new Symbol;new BigInt;new String;new Number;new Boolean;",
		"/* 世界 🌍 */\r\n new Function('é');\r\n new Symbol('é');new ((String))('é');\r\n",
		"new (Function as any)();new (Symbol as any)();new (String as any)();",
		"declare const Function:any;Function();new Function;Function.call();",
		"declare const Symbol:any;declare const String:any;new Symbol;new String;",
	} {
		add(source)
	}
	sort.Strings(sources)
	return sources
}

func wave23ConstructorMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_23_constructor_suite.a"), archive, false)
}

// Not parallel: builds, sanitizer archives and measured runs share one machine.
func TestWave23ConstructorAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE23_CONSTRUCTOR_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_23_constructor_suite.a")
	binary := h.build(stage0, "native", entry, archive, false)
	oracle := volumeOracle(h, "wave23-constructor-oracle", "oracle_wave_23_constructor.go")
	h.write("prelude.d.ts", "")
	config := h.write("controls-tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ESNext"],"noEmit":true},"files":["prelude.d.ts"]}`)
	var paths []string
	helper := h.write("constructor-helper.a", "export default class Function{}export class Symbol{}export class BigInt{}export class String{}export class Number{}export class Boolean{}\n")
	paths = append(paths, helper)
	for i, source := range wave23ConstructorControls(t, repository) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-new-func", "no-new-native-nonconstructor", "no-new-wrappers"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatal("no positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "native-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"func", "no_new_func.a", "this.references.global(receiver)", "true"},
		{"nonconstructor", "no_new_native_nonconstructor.a", "this.rules.byte(callee.end)", "this.rules.byte(node.end)"},
		{"wrappers", "no_new_wrappers.a", "this.rules.byte(node.end)", "this.rules.byte(callee.end)"},
	} {
		mutant := wave23ConstructorMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}

	// Custom declarations prove that "global" means declaration file, not default library.
	h.write("custom-globals.d.ts", "declare const Function:any;declare const Symbol:any;declare const BigInt:any;declare const String:any;declare const Number:any;declare const Boolean:any;\n")
	customConfig := h.write("custom-tsconfig.json", `{"compilerOptions":{"strict":true,"noLib":true,"noEmit":true},"files":["custom-globals.d.ts"]}`)
	customRoot := h.write("custom.a", "new Function;Function.call();new Symbol;new BigInt;new String;new Number;new Boolean;export {};\n")
	customManifest := h.write("custom.manifest", customRoot+"\n")
	customTruth := h.compare("custom-declarations", oracle, binary, customConfig, customManifest)
	for _, name := range []string{"no-new-func", "no-new-native-nonconstructor", "no-new-wrappers"} {
		if !bytes.Contains(customTruth.stdout, []byte("\t"+name+"\t")) {
			t.Fatal("missing custom declaration positive", name)
		}
	}
	h.compare("custom-declarations-asan", oracle, asan, customConfig, customManifest)

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
