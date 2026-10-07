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

// Inputs are taken from production test sources; expected diagnostics are never imported.
func wave17ThirdControls(h *harness) []string {
	unique := map[string]bool{}
	for _, name := range []string{"no_global_assign", "no_implicit_globals", "no_implied_eval"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/core", name+"_test.go"), nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			if v, ok := n.(*goast.ValueSpec); ok && len(v.Names) == 1 && v.Names[0].Name == "impliedEvalAmbientGlobals" {
				lit := v.Values[0].(*goast.BasicLit)
				source, _ := strconv.Unquote(lit.Value)
				h.write("globals.d.ts", source)
			}
			if row, ok := n.(*goast.CompositeLit); ok {
				for _, entry := range row.Elts {
					if pair, ok := entry.(*goast.KeyValueExpr); ok {
						if key, ok := pair.Key.(*goast.Ident); ok && (key.Name == "source" || key.Name == "sourceText" || key.Name == "code") {
							if value, ok := pair.Value.(*goast.BasicLit); ok {
								source, err := strconv.Unquote(value.Value)
								if err == nil && source != "" {
									unique[source] = true
								}
							}
						}
					}
				}
				if len(row.Elts) >= 2 {
					if value, ok := row.Elts[1].(*goast.BasicLit); ok && value.Kind == token.STRING {
						source, err := strconv.Unquote(value.Value)
						if err == nil && (strings.Contains(source, ";") || strings.Contains(source, "=")) {
							unique[source] = true
						}
					}
				}
			}
			// Script corpus rows are often bare strings in an array.
			if lit, ok := n.(*goast.BasicLit); ok && lit.Kind == token.STRING {
				source, err := strconv.Unquote(lit.Value)
				if err == nil && strings.Contains(source, ";") && (strings.Contains(source, "var ") || strings.Contains(source, "setTimeout(") || strings.Contains(source, " = ")) {
					unique[source] = true
				}
			}
			return true
		})
	}
	keys := []string{}
	for source := range unique {
		keys = append(keys, source)
	}
	sort.Strings(keys)
	paths := []string{}
	for i, source := range keys {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	scripts := []string{
		"var wave17var=1;function wave17function(){};let wave17let=1;const wave17const=1;class Wave17Class{}",
		"var {wave17a,alias:wave17b,nested:{wave17c},...wave17d}={};var [wave17e,,...wave17f]=[];",
		"wave17leak=1;[wave17arrayLeak,...wave17restLeak]=[];({k:wave17objectLeak}=o);for(wave17forLeak in o){};for(wave17ofLeak of o){};",
		"'use strict';wave17strictLeak=1;",
		"(function(){'use strict';wave17functionStrictLeak=1;})();class StrictClass{method(){wave17classLeak=1;}}",
		"(function(){'use other';wave17nestedLeak=1;})();{let blockLocal=1;}",
		"/* 世界 🌍 */\r\nvar wave17unicode=1;\r\nwave17unicodeLeak=1;",
		"var wave17declared;wave17declared=1;window.foo=()=>{wave17insideLeak=1;};",
		"({wave17shorthandLeak,wave17defaultLeak=1,...wave17spreadLeak}=o);",
	}
	for i, source := range scripts {
		physical := h.write(fmt.Sprintf("script-%03d.a", i), source+"\n")
		paths = append(paths, physical)
		alias := strings.TrimSuffix(physical, ".a") + ".js"
		os.Remove(alias)
		if err := os.Symlink(filepath.Base(physical), alias); err != nil {
			h.t.Fatal(err)
		}
		paths = append(paths, alias)
	}
	paths = append(paths, h.write("positive.a", "/* 世界 🌍 */\r\nexport {};Object=1;String++;({Array=0,Number=0}={});function shadow(Object:number){Object=1;}\r\nwindow.window['setTimeout']('x'+unknownArg);globalThis.setInterval(`x${unknownArg}`);execScript('x');setTimeout<string>('x');setTimeout(()=>0);\n"))
	return paths
}

// Not parallel: native and sanitizer builds use this machine's compilation budget.
func TestWave17ThirdAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE17_THIRD_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_17_third_suite.a")
	binary := h.build(stage0, "wave17-third", entry, archive, false)
	oracle := volumeOracle(h, "wave17-third-oracle", "oracle_wave_17_third.go")
	h.write("config-root.d.ts", "export {};\n")
	paths := wave17ThirdControls(h)
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","moduleDetection":"auto","allowJs":true,"lib":["ESNext"],"types":[],"noEmit":true},"files":["config-root.d.ts","globals.d.ts"]}`)
	manifest := h.write("all.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid-sources", exec.Command(oracle, config, manifest, "--valid-sources"))
	paths = strings.Fields(string(valid.stdout))
	manifest = h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, id := range []string{"noGlobalAssign", "globalNonLexicalBinding", "globalVariableLeak", "impliedEval", "execScript"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+id+"\t")) {
			t.Fatal("missing positive " + id)
		}
	}
	t.Logf("%d upstream and edge control files", len(paths))
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave17-third-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, m := range []struct{ name, file, from, to string }{
		{"global", "no_global_assign.a", "!symbol.present || !(symbol.declarationFiles[0] ?? false)", "!symbol.present || false"},
		{"implicit", "no_implicit_globals.a", "if(this.source.module)", "if(false && this.source.module)"},
		{"eval", "no_implied_eval.a", "declarations.every((d) => d)", "true"},
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
		got := h.must("mutant-"+m.name+"-run", exec.Command(b, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived or sanitizer caught " + m.name)
		}
		t.Logf("%s mutant exits 0, empty stderr; Go byte oracle catches byte %d", m.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, m := range []struct{ name, file, from, to string }{
		{"source-js", "bridge/tsgo/checker/global_source_facts.go", "out.yes(ast.IsSourceFileJS(source))", "out.yes(false)"},
		{"shorthand", "bridge/tsgo/checker/global_binding_facts.go", "shorthand = c.GetShorthandAssignmentValueSymbol(node.Parent)", "shorthand = c.GetSymbolAtLocation(node)"},
	} {
		overlay := h.overlay(m.name, m.file, m.from, m.to)
		a := h.archive(m.name+"-checker", overlay, true)
		b := h.build(stage0, m.name+"-mutant", entry, a, true)
		got := h.must(m.name+"-run", exec.Command(b, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("fact mutant survived or sanitizer caught " + m.name)
		}
		t.Logf("%s fact mutant exits 0, empty stderr; Go bytes catch byte %d", m.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, flags := range [][]string{{"--lexical-bindings"}, {"--exception=Object", "--exception=String"}, {"--lexical-bindings", "--exception=Array"}} {
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
	t.Log("new binding question rejects released program: panic 70")
	registryOverlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released roots.")
	registryArchive := h.archive("released-registry-checker", registryOverlay, false)
	registryBinary := h.build(stage0, "released-registry-mutant", released, registryArchive, true)
	gotRegistry := h.must("released-registry-mutant-run", exec.Command(registryBinary, config, probe))
	if len(gotRegistry.stderr) != 0 {
		t.Fatal("released registry mutant did not finish cleanly")
	}
	t.Log("released registry mutant exits 0, empty stderr; required panic 70 catches it")

	sourceProbe := h.write("released-source.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const p=tsgoProgram(args[0]??'',[file]);tsgoRelease(p);console.log(tsgoInspect(p,file,0,3,'SourceFile','global-source-facts'));`)
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
	t.Log("source metadata question rejects released handle; retained registry mutant finishes cleanly, required panic catches it")

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
		h.compare(population.name, oracle, binary, population.config, list)
		h.compare(population.name+"-asan", oracle, asan, population.config, list)
	}
}
