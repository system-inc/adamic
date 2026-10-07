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

func wave23BehaviorControls(t *testing.T, repository string) []string {
	t.Helper()
	seen := map[string]bool{}
	var sources []string
	add := func(source string) {
		if !seen[source] {
			seen[source] = true
			sources = append(sources, source)
		}
	}
	for _, name := range []string{"no_throw_literal", "no_useless_backreference", "prefer_arrow_callback"} {
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/core", name+"_test.go"), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			lit, ok := node.(*ast.CompositeLit)
			if !ok || lit.Type != nil || len(lit.Elts) < 1 {
				return true
			}
			var value *ast.BasicLit
			if first, ok := lit.Elts[0].(*ast.BasicLit); ok && first.Kind == token.STRING {
				value = first
				if len(lit.Elts) > 1 {
					if second, ok := lit.Elts[1].(*ast.BasicLit); ok && second.Kind == token.STRING {
						value = second
					}
				}
			}
			for _, element := range lit.Elts {
				if pair, ok := element.(*ast.KeyValueExpr); ok {
					if key, ok := pair.Key.(*ast.Ident); ok && (key.Name == "source" || key.Name == "sourceText") {
						if text, ok := pair.Value.(*ast.BasicLit); ok && text.Kind == token.STRING {
							value = text
						}
					}
				}
			}
			if value == nil {
				return true
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
		"throw undefined;throw 1;throw new Error();throw (new Error());throw error||1;throw 1&&error;throw true?error:1;",
		"function f(undefined:any){throw undefined}function g(){throw undefined}",
		"RegExp('\\\\1(a)');new RegExp('\\\\1(a)');globalThis.RegExp('\\\\1(a)');window['RegExp']('\\\\1(a)');",
		"const R=RegExp;let Q=R;Q('\\\\1(a)');const {RegExp:E}=globalThis;new E('\\\\1(a)');",
		"let R:any;R=RegExp;R('\\\\1(a)');(true?RegExp:RegExp)('\\\\1(a)');",
		"const p='\\\\1'+'(a)';const flags='u';RegExp(p,flags);let q=p;RegExp(q);q='(a)';RegExp(q);",
		"const RegExp:any=()=>{};RegExp('\\\\1(a)');",
		"RegExp=()=>null;RegExp('\\\\1(a)');window=other;window.RegExp('\\\\1(a)');",
		"const re=/(a)\\1/;RegExp(re);RegExp('\\\\1(a)',unknownFlags);",
		"f(function(x){return x+1});f(function named(x){return x});f(function(){return arguments});f(function(){return this});f(function(){return this}.bind(this));",
		"f(function named(){return named()});f(function named(){function inner(named:any){return named()}return inner});",
		"const obj = {['世界']: function named(){}};f(obj);const holder={arguments:1};f(function(){return holder.arguments});",
		"const p='\\\\1(a)';const object={RegExp};export {RegExp as R};globalThis['Reg'+'Exp'](p);",
		"const {RegExp:R}=globalThis;let q:any;({RegExp:q}=window);R('\\\\1(a)');q('\\\\1(a)');",
		"const R=RegExp;let Q=R;Q=Q;Q('\\\\1(a)');f((R=RegExp)('\\\\1(a)'));",
		"function f(R=RegExp){R('\\\\1(a)')}const {R=RegExp}=obj;R('\\\\1(a)');",
		"const p=`${'\\\\1'}(a)`;const flags='u';const R=RegExp as any;R(p,flags);",
		"[...RegExp]=[];RegExp('\\\\1(a)');const {RegExp:R}=globalThis;new R('\\\\1(a)');",
		"RegExp(String.raw`\\1(a)`);RegExp('\\\\1(a){,}','u');RegExp('\\\\1(a){,}');",
		"/* 世界 🌍 */\r\n throw 'é';f(function /* é */ named(x){return x});f(async function(x){return x});\r\n",
	} {
		add(source)
	}

	sort.Strings(sources)
	return sources
}

func wave23BehaviorMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_23_behavior_suite.a"), archive, false)
}

// Not parallel: builds, sanitizer archives and measured runs share one machine.
func TestWave23BehaviorAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE23_BEHAVIOR_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_23_behavior_suite.a")
	binary := h.build(stage0, "native", entry, archive, false)
	oracle := volumeOracle(h, "wave23-behavior-oracle", "oracle_wave_23_behavior.go")
	h.write("prelude.d.ts", "")
	config := h.write("controls-tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ESNext"],"noEmit":true},"files":["prelude.d.ts"]}`)
	var paths []string
	for i, source := range wave23BehaviorControls(t, repository) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("source-filter", exec.Command(oracle, config, manifest, "--valid-sources"))
	filtered := strings.TrimSpace(string(valid.stdout))
	if filtered == "" {
		t.Fatal("empty filtered controls")
	}
	t.Logf("controls: %d submitted roots; %d Go-parse-valid roots", len(paths), len(strings.Split(filtered, "\n")))
	manifest = h.write("controls-valid.manifest", filtered+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-throw-literal", "no-useless-backreference", "prefer-arrow-callback"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatal("no positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "native-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"throw", "no_throw_literal.a", "message = undefMessage", "message = objectMessage"},
		{"backreference", "backreference_structure.a", "return 'backward'", "return 'forward'"},
		{"arrow", "prefer_arrow_callback.a", "this.rules.byte(insertion), ' =>'", "this.rules.byte(insertion), ' => '"},
	} {
		mutant := wave23BehaviorMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}

	overlayFacts := h.overlay("callback-symbol-origin", "bridge/tsgo/checker/callback_symbol_facts.go", "out.yes(source.IsDeclarationFile)", "out.yes(true)")
	factsArchive := h.archive("callback-symbol-origin", overlayFacts, false)
	factsBinary := h.build(stage0, "callback-symbol-origin-native", entry, factsArchive, false)
	factsGot := h.must("callback-symbol-origin-run", exec.Command(factsBinary, config, manifest))
	if len(factsGot.stderr) != 0 || bytes.Equal(factsGot.stdout, truth.stdout) {
		t.Fatal("callback symbol origin mutant survived")
	}
	t.Logf("callback symbol origin mutant: exit 0, empty stderr, Go bytes catch byte %d", firstDifference(factsGot.stdout, truth.stdout))

	for _, flags := range [][]string{{"--allow-named"}, {"--reject-unbound"}, {"--allow-named", "--reject-unbound"}} {
		name := "options" + strings.Join(flags, "")
		want := h.must(name+"-go", exec.Command(oracle, append([]string{config, manifest}, flags...)...))
		for label, executable := range map[string]string{"native": binary, "asan": asan} {
			got := h.must(name+"-"+label, exec.Command(executable, append([]string{config, manifest}, flags...)...))
			if len(got.stderr) != 0 || !bytes.Equal(got.stdout, want.stdout) {
				t.Fatalf("%s %s options mismatch byte %d", name, label, firstDifference(got.stdout, want.stdout))
			}
		}
		t.Logf("%s: %d identical finding bytes", name, len(want.stdout))
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
