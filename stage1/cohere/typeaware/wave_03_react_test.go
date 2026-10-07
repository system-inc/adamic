package typeaware

import (
	"bytes"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func wave03ReactBodies(t *testing.T, repository string) []string {
	t.Helper()
	var bodies []string
	for _, name := range []string{"unsupported_syntax", "use_memo", "boolean_prop_naming"} {
		tree, err := goparser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/react", name+"_test.go"), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			literal, ok := node.(*goast.CompositeLit)
			if !ok {
				return true
			}
			for _, element := range literal.Elts {
				pair, ok := element.(*goast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := pair.Key.(*goast.Ident)
				if !ok || (key.Name != "source" && key.Name != "sourceText" && key.Name != "code") {
					continue
				}
				value, ok := pair.Value.(*goast.BasicLit)
				if !ok || value.Kind != token.STRING {
					continue
				}
				text, err := strconv.Unquote(value.Value)
				if err != nil {
					t.Fatal(err)
				}
				bodies = append(bodies, text)
				return false
			}
			at := 0
			if name == "boolean_prop_naming" {
				at = 1
			}
			if len(literal.Elts) > at {
				if value, ok := literal.Elts[at].(*goast.BasicLit); ok && value.Kind == token.STRING {
					text, err := strconv.Unquote(value.Value)
					if err != nil {
						t.Fatal(err)
					}
					if strings.Contains(text, "function") || strings.Contains(text, "useMemo") || strings.Contains(text, "eval") || strings.Contains(text, "propTypes") || strings.Contains(text, "<") {
						bodies = append(bodies, text)
						return false
					}
				}
			}
			return true
		})
	}
	return bodies
}

// Not parallel: checker archives, sanitizer subprocesses, and timings share a machine.
func TestWave03ReactAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE03_REACT_ARTIFACTS"); path != "" {
		directory = path
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	optionProbe := h.run("regex-option-probe", exec.Command(stage0, "build", filepath.Join(repository, "stage1/cohere/typeaware/wave_03_react/gaps/general_regex.a"), "-o", filepath.Join(directory, "regex-option-probe")))
	if optionProbe.err != nil {
		if strings.Contains(string(optionProbe.stderr), "stage 0 can't lower RegExp with a nonconstant pattern yet") {
			t.Fatalf("BLOCKED: required new RegExp(pattern, 'u') is not supported by native lowering: %s", optionProbe.stderr)
		}
		t.Fatalf("unexpected regex option probe failure: %v %s", optionProbe.err, optionProbe.stderr)
	}
	archive := h.archive("react-checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_03_react/main.a")
	binary := h.build(stage0, "react", entry, archive, false)
	oracle := volumeOracle(h, "react-oracle", "oracle_wave_03_react.go")
	types := h.write("react.d.ts", "export {}; declare global { interface GlobalProps { enabled:boolean; isReady:boolean } }\n")
	config := h.write("react-config.json", fmt.Sprintf(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022","DOM"],"moduleDetection":"auto","jsx":"preserve"},"files":[%q]}`, types))
	bodies := append(wave03ReactBodies(t, repository), "function Component(){eval('x');return <div/>}", "function Component(){class Inner{};return <div/>}", "function Component(){useMemo();return <div/>}", "function Component(){useMemo(value,[1]);return <div/>}")
	fixture, err := os.ReadFile(filepath.Join(repository, "cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.invalid-eval-unsupported.js"))
	if err != nil {
		t.Fatal(err)
	}
	bodies = append(bodies, string(fixture), "function Foreign(p:GlobalProps){return <div/>}", "function Component(){let outer=1;useMemo(()=>{outer+=1;return outer},[(outer)]);return <div/>}", "function Component(){with({}) { const x=1; } return <div/>}", "const C=(p:{'isReady':boolean; [key]:boolean})=><div/>;", "const C=(p:{ enabled:boolean | undefined; isReady:boolean})=><div/>;", "const C=(p:{enabled:boolean})=>{do{return <div/>}while(false)};", "const C=(p:{enabled:boolean})=>{while(true){return <div/>}};", "const C=(p:{enabled:boolean})=>{try{return <div/>}catch(e){}};", "const C=(p:{enabled:boolean})=>React.createElement('div');", "const C=(p:{enabled:boolean})=>createElement('div');", "import {createElement} from 'react';const C=(p:{enabled:boolean})=>createElement('div');", "function ß(p:{enabled:boolean}) {return null}", `X.propTypes={isA:bool,isx:bool,hasA:bool,has:bool,xisA:bool,shouldX:bool,withoutX:bool,disabled:bool,foorequiredbar:bool,defaultChecked:bool,'isA':bool,extra:mutuallyExclusiveTrueProps};`, "import {useMemo as cache} from 'react';const Component=()=>{cache();return <div/>};", "export default class extends React.Component {};const C:React.FC<{enabled:boolean}>=React.memo(()=>null);", "function Library(p:MediaTrackSettings){return null}", "function LibraryAlias(p:ConstrainBoolean){return null}", "function \u1f80(p:{enabled:boolean}){return null}", "function \u1f88(p:{enabled:boolean}){return null}", "function \u1fb3(p:{enabled:boolean}){return null}", "const C=(p:{enabled:boolean})=>{return <div/>;}; C.propTypes={enabled:bool};")
	var paths []string
	for index, source := range bodies {
		paths = append(paths, h.write(fmt.Sprintf("react-control-%03d.tsx", index), source+"\nexport {};\n"))
	}
	candidates := h.write("react-candidates.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("parse-filter", exec.Command(oracle, config, candidates, "--valid-sources"))
	manifest := h.write("react-controls.manifest", string(valid.stdout))
	t.Logf("controls: %d candidates %d parseable", len(paths), len(strings.Fields(string(valid.stdout))))
	truth := h.compare("react-controls", oracle, binary, config, manifest)
	for index, flags := range [][]string{{"--boolean-props"}, {"--boolean-props", "--nested"}, {"--boolean-props", "--is-only"}, {"--boolean-props", "--pattern=(is|has)[A-Z]([A-Za-z0-9]?)+"}, {"--boolean-props", "--pattern=(^(is|has|should|without)[A-Z]([A-Za-z0-9]?)+|disabled|required|checked|defaultChecked)"}, {"--boolean-props", "--nested", "--type-names=bool,mutuallyExclusiveTrueProps", "--message={{ propName }}:{{pattern}} / {{component}} / {{unknown}} / {{{{propName}} / {{\u00a0propName}} / {{\vpropName}}"}, {"--boolean-props", "--type-names=", "--pattern=^is[A-Z]"}} {
		args := append([]string{config, manifest}, flags...)
		want := h.must(fmt.Sprintf("react-options-%02d-go", index), exec.Command(oracle, args...))
		got := h.must(fmt.Sprintf("react-options-%02d-native", index), exec.Command(binary, args...))
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatalf("boolean flags %v differ byte %d: native %q Go %q", flags, firstDifference(got.stdout, want.stdout), got.stdout[max(0, firstDifference(got.stdout, want.stdout)-40):min(len(got.stdout), firstDifference(got.stdout, want.stdout)+300)], want.stdout[max(0, firstDifference(got.stdout, want.stdout)-40):min(len(want.stdout), firstDifference(got.stdout, want.stdout)+300)])
		}
		t.Logf("%v: %d identical bytes %s", flags, len(got.stdout), summary(got.stdout))
	}
	for _, rule := range []string{"react-hooks/unsupported-syntax", "react-hooks/use-memo"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+rule+"\t")) {
			t.Fatalf("no positive control for %s", rule)
		}
	}
	if os.Getenv("ADAMIC_WAVE03_REACT_QUICK") != "" {
		return
	}

	sanitized := h.archive("react-checker-asan", "", true)
	asan := h.build(stage0, "react-asan", entry, sanitized, true)
	h.compare("react-controls-asan", oracle, asan, config, manifest)
	wantBoolean := h.must("react-boolean-go", exec.Command(oracle, config, manifest, "--boolean-props", "--nested"))
	gotBoolean := h.must("react-boolean-asan", exec.Command(asan, config, manifest, "--boolean-props", "--nested"))
	if !bytes.Equal(wantBoolean.stdout, gotBoolean.stdout) || len(gotBoolean.stderr) != 0 {
		t.Fatal("boolean sanitizer mismatch")
	}
	for _, change := range []struct {
		name, file, from, to string
		flags                []string
	}{
		{"unsupported", "messages.a", "A class declared during render", "A class  declared during render", nil},
		{"memo", "messages.a", "Pass the calculation as an inline function.", "Pass  the calculation as an inline function.", nil},
		{"boolean", "boolean_prop_naming.a", `doesn\u2019t match rule`, `doesn\u0027t match rule`, []string{"--boolean-props", "--nested"}},
	} {
		mutant := wave03ReactMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		args := append([]string{config, manifest}, change.flags...)
		got := h.must(change.name+"-mutant-run", exec.Command(mutant, args...))
		want := truth.stdout
		if change.name == "boolean" {
			want = wantBoolean.stdout
		}
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, want) || summary(got.stdout) != summary(want) {
			t.Fatalf("%s byte-only mutant survived or changed findings count", change.name)
		}
		t.Logf("%s mutant builds, exits 0, empty stderr, unchanged %s, byte comparison catches byte %d", change.name, summary(got.stdout), firstDifference(got.stdout, want))
	}
	aliasMutant := wave03ReactMutant(h, stage0, archive, "memo-alias", "use_memo.a", "names.add(tree.text(node.name))", "names.add(tree.text(imported))")
	aliasResult := h.must("memo-alias-run", exec.Command(aliasMutant, config, manifest))
	if len(aliasResult.stderr) != 0 || bytes.Equal(aliasResult.stdout, truth.stdout) {
		t.Fatal("memo alias filter mutant survived")
	}
	t.Logf("memo alias filter mutant builds, exits 0, byte comparison catches byte %d", firstDifference(aliasResult.stdout, truth.stdout))
	optionArgs := []string{config, manifest, "--boolean-props", "--pattern=^(ready|enabled)$"}
	optionTruth := h.must("general-pattern-go", exec.Command(oracle, optionArgs...))
	optionNative := h.must("general-pattern-native", exec.Command(binary, optionArgs...))
	if !bytes.Equal(optionTruth.stdout, optionNative.stdout) || len(optionNative.stderr) != 0 {
		t.Fatal("configured JavaScript regex comparison differs")
	}
	patternMutant := wave03ReactMutant(h, stage0, archive, "pattern-option", "boolean_pattern.a", "new RegExp(pattern, 'u')", "new RegExp('', 'u')")
	patternResult := h.must("pattern-option-run", exec.Command(patternMutant, optionArgs...))
	if bytes.Equal(patternResult.stdout, optionTruth.stdout) || len(patternResult.stderr) != 0 {
		t.Fatal("configured regex mutation survived")
	}
	t.Log("configured regex mutant compiles, exits zero and is caught only by comparison")
	for _, population := range []struct{ name, config, manifest string }{{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE03_REPOSITORY_MANIFEST")}, {"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE03_COMPILER_MANIFEST")}} {
		if population.manifest == "" {
			continue
		}
		h.compare(population.name, oracle, binary, population.config, population.manifest)
		h.compare(population.name+"-asan", oracle, asan, population.config, population.manifest)
		want := h.must(population.name+"-timed-go", exec.Command(oracle, population.config, population.manifest))
		command := exec.Command(binary, population.config, population.manifest)
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(population.name+"-timed-native", command)
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatalf("timed %s differs", population.name)
		}
		t.Logf("%s native %s Go %s; native %s Go %s", population.name, got.elapsed, want.elapsed, got.stderr, want.stderr)
		want = h.must(population.name+"-boolean-go", exec.Command(oracle, population.config, population.manifest, "--boolean-props", "--nested"))
		got = h.must(population.name+"-boolean-native", exec.Command(binary, population.config, population.manifest, "--boolean-props", "--nested"))
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatalf("boolean %s differs byte %d", population.name, firstDifference(got.stdout, want.stdout))
		}
		got = h.must(population.name+"-boolean-asan", exec.Command(asan, population.config, population.manifest, "--boolean-props", "--nested"))
		if len(got.stderr) != 0 || !bytes.Equal(want.stdout, got.stdout) {
			t.Fatalf("boolean %s sanitizer differs", population.name)
		}
		t.Logf("%s boolean-options: %d identical bytes %s", population.name, len(want.stdout), summary(want.stdout))
	}
	probe := h.write("react-released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,2,'SourceFile','source-syntax-tree'));`)
	source := h.write("react-probe.ts", "x;")
	stale := h.build(stage0, "react-released", probe, archive, false)
	got := h.run("react-released-run", exec.Command(stale, config, source))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released syntax query: panic 70")
	overlay := h.overlay("react-released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains a released program.")
	mutantArchive := h.archive("react-released-registry", overlay, false)
	mutant := h.build(stage0, "react-released-mutant", probe, mutantArchive, false)
	h.must("react-released-mutant-run", exec.Command(mutant, config, source))
	t.Log("released registry mutant exits 0 and is caught by required panic 70")
}

func wave03ReactMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	paths, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/wave_03_react/*.a"))
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
				h.t.Fatalf("nonunique %s mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(source, "../../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "'../", "'"+filepath.Join(h.repository, "stage1/cohere/typeaware")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "main.a"), archive, false)
}
