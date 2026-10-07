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
)

func wave17FifthControls(h *harness) []string {
	unique := map[string]bool{}
	for _, name := range []string{"unsupported_syntax", "use_memo"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/react", name+"_test.go"), nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			if lit, ok := n.(*goast.BasicLit); ok && lit.Kind == token.STRING {
				text, err := strconv.Unquote(lit.Value)
				if err == nil && (strings.Contains(text, "function") || strings.Contains(text, "=>")) {
					unique[text] = true
				}
			}
			return true
		})
	}
	for _, text := range []string{
		"/* 世界 🌍 */\r\nfunction Component(){useThing();eval('x');class Foo{};}",
		"function Component(){useThing();const f=eval;const o={eval(){}};o.eval();}",
		"function Component(eval){useThing();eval('x');}",
		"function lower(){function Component(){useThing();eval('x')}}",
		"function Outer(){function Component(){useThing();eval('x')}}",
		"function Component(){useMemo(()=>1,[f(),g(),this.a,p[0],p['x']]);}",
		"function Component(){useMemo();useCallback();useMemo(1,[f(),this.a,p[0],p['x'],...p]);}",
		"function Component(){let x;useMemo(async (p)=>{x=1;return p},[]);useCallback(async(p)=>p,[])}",
		"function Component(){let x;useMemo(()=>{let x;x=1;return x},[]);useMemo(()=>{function f(){x=1};return 1},[])}",
		"import {useMemo as um} from 'react';function Component(){um(1,[])}",
		"import React from 'preact/compat';function Component(){React.useMemo(1,[])}",
		"C.propTypes={enabled:PropTypes.bool,isEnabled:PropTypes.bool,'isGood':bool,[isFine]:bool,number:PropTypes.number};",
		"C.propTypes={nested:PropTypes.shape({enabled:PropTypes.bool}),required:PropTypes.bool.isRequired};",
		"class C {static propTypes={enabled:PropTypes.bool};}",
	} {
		unique[text] = true
	}
	keys := []string{}
	for text := range unique {
		keys = append(keys, text)
	}
	sort.Strings(keys)
	paths := []string{}
	for i, text := range keys {
		physical := h.write(fmt.Sprintf("control-%03d.a", i), text+"\nexport {};\n")
		path := strings.TrimSuffix(physical, ".a") + ".tsx"
		os.Remove(path)
		if err := os.Symlink(filepath.Base(physical), path); err != nil {
			h.t.Fatal(err)
		}
		paths = append(paths, path)
	}
	return paths
}

// Not parallel: native and sanitizer builds share the machine's compilation budget.
func TestWave17FifthAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE17_FIFTH_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_17_fifth_projected_suite.a")
	binary := h.build(stage0, "wave17-fifth", entry, archive, false)
	oracle := volumeOracle(h, "wave17-fifth-oracle", "oracle_wave_17_fifth.go")
	h.write("config-root.d.ts", "export {};\n")
	paths := wave17FifthControls(h)
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","jsx":"preserve","moduleDetection":"auto","allowJs":true,"lib":["ESNext"],"types":[],"noEmit":true},"files":["config-root.d.ts"]}`)
	manifest := h.write("all.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid-sources", exec.Command(oracle, config, manifest, "--valid-sources"))
	paths = strings.Fields(string(valid.stdout))
	manifest = h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	for i, path := range paths {
		one := h.write(fmt.Sprintf("tree-%03d.manifest", i), path+"\n")
		tree := h.must("raw-tree", exec.Command(oracle, config, one, "--tree"))
		if err := os.WriteFile(path+".tree", tree.stdout, 0644); err != nil {
			t.Fatal(err)
		}
	}
	truth := h.must("controls-go", exec.Command(oracle, config, manifest, "--boolean-props"))
	gotControls := h.must("controls-native", exec.Command(binary, config, manifest, "--boolean-props"))
	if len(gotControls.stderr) != 0 || !bytes.Equal(truth.stdout, gotControls.stdout) {
		t.Fatalf("control mismatch at byte %d", firstDifference(truth.stdout, gotControls.stdout))
	}
	t.Logf("controls: %d identical bytes", len(truth.stdout))
	for _, id := range []string{"unsupportedEval", "unsupportedWith", "unsupportedInlineClass", "useMemoMissingCallback", "useMemoCallbackNotInline", "useMemoDependencyListNotArrayLiteral", "useMemoDependencyNotSimple", "useMemoCallbackHasParameters", "useMemoCallbackAsyncOrGenerator", "useMemoCallbackReassignsOuterVariable"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+id+"\t")) {
			t.Fatal("missing positive " + id)
		}
	}
	t.Logf("%d upstream and edge control files", len(paths))
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave17-fifth-asan", entry, sanitized, true)
	gotAsan := h.must("controls-asan", exec.Command(asan, config, manifest, "--boolean-props"))
	if len(gotAsan.stderr) != 0 || !bytes.Equal(truth.stdout, gotAsan.stdout) {
		t.Fatal("sanitizer controls differ")
	}
	for _, m := range []struct{ name, file, from, to string }{
		{"syntax", "react_unsupported_syntax.a", "'unsupportedEval', unsupportedEval, i", "'unsupportedEval', unsupportedEval, i+1"},
		{"memo", "react_use_memo.a", "this.report('useMemoMissingCallback', useMemoMissingCallback, i)", "this.report('useMemoMissingCallback', useMemoMissingCallback, i+1)"},
		{"boolean", "react_boolean_prop_naming.a", "if(this.named(key) && this.matches(name))", "if(false && this.named(key) && this.matches(name))"},
	} {
		original, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware", m.file))
		if err != nil {
			t.Fatal(err)
		}
		s := strings.ReplaceAll(string(original), "'./", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
		if strings.Count(s, m.from) != 1 {
			t.Fatal("mutant anchor " + m.name)
		}
		rule := h.write("mutant-"+m.name+".a", strings.Replace(s, m.from, m.to, 1))
		data, err := os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		s = strings.ReplaceAll(string(data), "'./", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
		s = strings.ReplaceAll(s, "'../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")

		s = strings.Replace(s, filepath.Join(repository, "stage1/cohere/typeaware", m.file), rule, 1)
		e := h.write("mutant-suite-"+m.name+".a", s)
		b := h.build(stage0, "mutant-"+m.name, e, sanitized, true)
		got := h.must("mutant-"+m.name+"-run", exec.Command(b, config, manifest, "--boolean-props"))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived or sanitizer caught " + m.name)
		}
		t.Logf("%s mutant exits 0, empty stderr; Go byte oracle catches byte %d", m.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, flags := range [][]string{{"--boolean-props", "--nested"}} {
		commands := append([]string{config, manifest}, flags...)
		want := h.must("options-go", exec.Command(oracle, commands...))
		got := h.must("options-native", exec.Command(asan, commands...))
		if len(got.stderr) != 0 || !bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("options differ " + strings.Join(flags, " "))
		}
		t.Logf("options %v: %d identical bytes", flags, len(got.stdout))
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const p=tsgoProgram(args[0]??'',[file]);tsgoRelease(p);console.log(tsgoInspect(p,file,0,1,'Identifier','global-binding-facts'));`)
	probe := h.write("released-input.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if exit, ok := got.err.(*exec.ExitError); !ok || exit.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("binding question rejects released program: panic 70")
	registryOverlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released roots.")
	registryArchive := h.archive("released-registry-checker", registryOverlay, false)
	registryBinary := h.build(stage0, "released-registry-mutant", released, registryArchive, true)
	gotRegistry := h.must("released-registry-mutant-run", exec.Command(registryBinary, config, probe))
	if len(gotRegistry.stderr) != 0 {
		t.Fatal("released registry mutant did not finish cleanly")
	}
	t.Log("released registry mutant exits 0, empty stderr; required panic 70 catches it")

	sourceProbe := h.write("released-source.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const p=tsgoProgram(args[0]??'',[file]);tsgoRelease(p);console.log(tsgoInspect(p,file,0,1,'Identifier','symbol-identities'));`)
	sourceBinary := h.build(stage0, "released-source", sourceProbe, archive, false)
	sourceResult := h.run("released-source-run", exec.Command(sourceBinary, config, probe))
	if exit, ok := sourceResult.err.(*exec.ExitError); !ok || exit.ExitCode() != 70 || string(sourceResult.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatal("released source question escaped")
	}
	sourceMutant := h.build(stage0, "released-source-registry-mutant", sourceProbe, registryArchive, true)
	sourceGot := h.must("released-source-registry-run", exec.Command(sourceMutant, config, probe))
	if len(sourceGot.stderr) != 0 {
		t.Fatal("registry source mutant did not finish cleanly")
	}
	t.Log("symbol identity question rejects released handle; retained registry mutant finishes cleanly, required panic catches it")

	productionEntry := filepath.Join(repository, "stage1/cohere/typeaware/wave_17_fifth_suite.a")
	production := h.build(stage0, "wave17-fifth-production", productionEntry, archive, false)
	productionAsan := h.build(stage0, "wave17-fifth-production-asan", productionEntry, sanitized, true)
	nativePaths := []string{}
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(source, []byte("<")) {
			nativePaths = append(nativePaths, path)
		}
	}
	nativeManifest := h.write("native-controls.manifest", strings.Join(nativePaths, "\n")+"\n")
	h.compare("ordinary-controls", oracle, production, config, nativeManifest)
	h.compare("ordinary-controls-asan", oracle, productionAsan, config, nativeManifest)
	t.Logf("%d controls use the ordinary native parser", len(nativePaths))

	for _, population := range []struct{ name, root, config, manifest string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json"), "repository.manifest"},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), "compiler.manifest"},
	} {
		if population.root == "" {
			t.Fatal("ADAMIC_TYPESCRIPT_SOURCE required")
		}
		if population.name == "compiler" {
			pin := h.must("corpus-pin", exec.Command("git", "-C", population.root, "rev-parse", "HEAD"))
			if strings.TrimSpace(string(pin.stdout)) != compilerCommit {
				t.Fatal("wrong TypeScript pin")
			}
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.manifest))
		if err != nil {
			t.Fatal(err)
		}
		paths = nil
		for _, p := range strings.Split(string(data), "\n") {
			if p != "" {
				paths = append(paths, filepath.Join(population.root, p))
			}
		}
		list := h.write(population.name+".manifest", strings.Join(paths, "\n")+"\n")
		h.compare(population.name, oracle, production, population.config, list)
		h.compare(population.name+"-asan", oracle, productionAsan, population.config, list)
	}
}
