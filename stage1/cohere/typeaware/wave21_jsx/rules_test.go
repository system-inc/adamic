package wave21jsx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
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

// Not parallel: native/sanitizer builds and timing share the private archives.
func TestPreparedJSXRules(t *testing.T) {
	repository, e := filepath.Abs("../../../..")
	if e != nil {
		t.Fatal(e)
	}
	directory := os.Getenv("ADAMIC_WAVE21_JSX_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if e = os.MkdirAll(directory, 0755); e != nil {
		t.Fatal(e)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	modules := filepath.Join(repository, "stage1/cohere/typeaware/wave21_jsx")
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	sanitized := h.archive("checker-asan", "", true)
	data, e := os.ReadFile(filepath.Join(modules, "oracle.go.txt"))
	if e != nil {
		t.Fatal(e)
	}
	virtual := filepath.Join(repository, "cohere/adamic_wave21_jsx.go")
	overlay, e := json.Marshal(map[string]any{"Replace": map[string]string{virtual: h.write("oracle.go", string(data))}})
	if e != nil {
		t.Fatal(e)
	}
	oracle := filepath.Join(directory, "oracle")
	build := exec.Command("go", "build", "-overlay", h.write("oracle-overlay.json", string(overlay)), "-o", oracle, virtual)
	build.Dir = filepath.Join(repository, "cohere")
	h.must("oracle-build", build)
	h.write("react.d.ts", `declare module 'react' {export const Fragment:any;export function createContext(v:any):any;export function useMemo<T>(f:()=>T,deps?:unknown[]):T;export function useCallback<T>(f:T,deps?:unknown[]):T;export class Component {} export class PureComponent {} const React:any; export default React;}
declare function useMemo<T>(f:()=>T,deps?:unknown[]):T;declare function useCallback<T>(f:T,deps?:unknown[]):T;declare function createContext(v:any):any;`)
	config := h.write("config.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","moduleDetection":"force","lib":["ES2022"],"jsx":"preserve","noEmit":true},"include":["*.d.ts"]}`)
	texts := map[string]bool{}
	for _, name := range []string{"jsx_fragments", "jsx_no_undef", "jsx_no_constructed_context_values", "jsx_no_constructed_context_values_stability"} {
		tree, e := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/react", name+"_test.go"), nil, 0)
		if e != nil {
			t.Fatal(e)
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			lit, ok := n.(*goast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			text, e := strconv.Unquote(lit.Value)
			if e == nil && strings.Contains(text, "<") && (strings.Contains(text, "/>") || strings.Contains(text, "</") || strings.Contains(text, "<>")) && !strings.HasPrefix(text, "/repository/") {
				texts[text] = true
			}
			return true
		})
	}
	for _, source := range []string{
		`import React,{Fragment as F} from 'react';function Component(){const v={x:1};return <><F/><React.Fragment/><Ctx.Provider value={v}/><Missing/><app.Foo/><div/><Ctx.Provider value={useMemo(()=>({v}),[v])}/></>;}`,
		`function Component(){return <Ctx.Provider value={useMemo(()=>({x:1}))}/>;}`,
		`function ÉComponent(){return <Ctx.Provider value={()=>{}}/>;}`,
		`function Component(){const a\u200c={};return <Ctx.Provider value={useMemo(()=>({}),[a\u200c])}/>;}`,
		`function Component(){const fn=()=>{};return <Ctx.Provider value={fn}/>;}`,
		`function Component(){const {x}=other;return <Ctx.Provider value={useCallback(()=>x,[x])}/>;}`,
		`function Component(){return <Ctx.Provider value={useMemo(()=>({x:1}),[Promise.resolve(1)])}/>;}`,
		`<_foo/>;<$foo/>;<테스트/>;<Foo-bar/>;<this.foo/>;<a:b/>;<Map/>;`,
		`import {Fragment as F} from 'preact';<F/>;const {anything:Alias}=React;<Alias/>;const {Fragment}=require('react');<Fragment/>;`,
		`class C extends React.Component {render(){return <Ctx.Provider value={[]}/>;}}`,
		`function outer(){const v={};return function Component(){return <Ctx.Provider value={v}/>;}}`,
		`function Component(){let v;return <Ctx.Provider value={v={}}/>;}`,
		`function Component(){return <Ctx.Provider value={cond?{}:[]}/>;}`,
	} {
		texts[source] = true
	}
	keys := []string{}
	for s := range texts {
		keys = append(keys, s)
	}
	sort.Strings(keys)
	paths := []string{}
	for i, s := range keys {
		paths = append(paths, h.write(fmt.Sprintf("fixture-%03d.tsx", i), s+"\n"))
	}
	list := h.write("all.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("source-filter", exec.Command(oracle, config, list, "--valid-sources"))
	paths = strings.Split(strings.TrimSpace(string(valid.stdout)), "\n")
	t.Logf("valid raw production-fixture controls %d/%d", len(paths), len(keys))
	type batch struct{ generated, want []byte }
	batches := []batch{}
	var nativeSeconds, goSeconds, prepareSeconds float64
	for _, mode := range []struct {
		name  string
		flags []string
	}{{"default", nil}, {"element", []string{"--element"}}, {"globals", []string{"--allow-globals"}}} {
		for from := 0; from < len(paths); from += 10 {
			end := min(from+10, len(paths))
			name := fmt.Sprintf("%s-%02d", mode.name, from/10)
			manifest := h.write(name+".manifest", strings.Join(paths[from:end], "\n")+"\n")
			prepared := h.must(name+"-prepare", exec.Command(oracle, append([]string{config, manifest, "--prepare", modules}, mode.flags...)...))
			entry := h.write(name+".a", string(prepared.stdout))
			binary := h.build(stage0, name+"-native", entry, archive, false)
			want := h.must(name+"-go", exec.Command(oracle, append([]string{config, manifest}, mode.flags...)...))
			got := h.must(name+"-run", exec.Command(binary))
			if len(got.stderr) != 0 || !bytes.Equal(got.stdout, want.stdout) {
				t.Fatalf("%s byte %d\nnative %s\nGo %s\nstderr %s", name, firstDifference(got.stdout, want.stdout), got.stdout, want.stdout, got.stderr)
			}
			nativeSeconds += got.elapsed.Seconds()
			goSeconds += want.elapsed.Seconds()
			prepareSeconds += prepared.elapsed.Seconds()
			if mode.name == "default" {
				batches = append(batches, batch{prepared.stdout, want.stdout})
			}
			asan := h.build(stage0, name+"-asan", entry, sanitized, true)
			got = h.must(name+"-asan-run", exec.Command(asan))
			if len(got.stderr) != 0 || !bytes.Equal(got.stdout, want.stdout) {
				t.Fatalf("sanitized %s differs or reports %s", name, got.stderr)
			}
			t.Logf("%s complete findings/fixes/suggestions %d identical bytes; %s", name, len(want.stdout), summary(want.stdout))
		}
	}
	for _, change := range []struct{ name, file, from, to, marker string }{
		{"fragment", "jsx-fragments/rule.a", "if(!fragmentName(s,opening.tag))", "if(fragmentName(s,opening.tag))", "preferFragment"},
		{"undef", "jsx-no-undef/rule.a", "reference.symbol!==0", "reference.symbol===0", "jsxIdentifierNotDefined"},
		{"context", "jsx-no-constructed-context-values/rule.a", "if(!inComponent(s,node))", "if(inComponent(s,node))", "withIdentifierMsg"},
	} {
		var selected *batch
		for i := range batches {
			if bytes.Contains(batches[i].want, []byte("\t"+change.marker+"\t")) {
				selected = &batches[i]
				break
			}
		}
		if selected == nil {
			t.Fatal("missing positive", change.name)
		}
		original := filepath.Join(modules, change.file)
		data, e := os.ReadFile(original)
		if e != nil {
			t.Fatal(e)
		}
		if !strings.Contains(string(data), change.from) {
			t.Fatal("mutant anchor", change.name)
		}
		source := strings.Replace(string(data), change.from, change.to, 1)
		source = strings.ReplaceAll(source, "'./", "'"+filepath.Dir(original)+"/")
		source = strings.ReplaceAll(source, "'../", "'"+modules+"/")
		mutant := h.write(change.name+"_mutant.a", source)
		driver, e := os.ReadFile(filepath.Join(modules, "driver.a"))
		if e != nil {
			t.Fatal(e)
		}
		d := strings.Replace(string(driver), "'./"+change.file+"'", strconv.Quote(mutant), 1)
		d = strings.ReplaceAll(d, "'./", "'"+modules+"/")
		d = strings.ReplaceAll(d, "'../../../", "'"+filepath.Join(repository, "stage1")+"/")
		mutantDriver := h.write(change.name+"_driver.a", d)
		main := strings.Replace(string(selected.generated), strconv.Quote(filepath.Join(modules, "driver.a")), strconv.Quote(mutantDriver), 1)
		binary := h.build(stage0, change.name+"-mutant", h.write(change.name+"_main.a", main), archive, false)
		got := h.must(change.name+"-mutant-run", exec.Command(binary))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, selected.want) {
			t.Fatal("mutant survived or failed outside comparison", change.name)
		}
		t.Logf("%s mutant exit 0 empty stderr; complete-byte oracle catches byte %d", change.name, firstDifference(got.stdout, selected.want))
	}
	t.Logf("prepared-input native %.6fs, full Go %.6fs, raw Go preparation %.6fs; different pipelines, not end-to-end speed", nativeSeconds, goSeconds, prepareSeconds)
	for _, corpus := range []struct{ name, config, manifest string }{{"compiler", os.Getenv("ADAMIC_WAVE21_COMPILER_CONFIG"), os.Getenv("ADAMIC_WAVE21_COMPILER_MANIFEST")}, {"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE21_REPOSITORY_MANIFEST")}} {
		if corpus.manifest == "" {
			continue
		}
		p := h.must(corpus.name+"-prepare", exec.Command(oracle, corpus.config, corpus.manifest, "--prepare", modules))
		entry := h.write(corpus.name+".a", string(p.stdout))
		binary := h.build(stage0, corpus.name+"-native", entry, archive, false)
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		asan := h.build(stage0, corpus.name+"-asan", entry, sanitized, true)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
	}
}

func TestNumericListenerMetadata(t *testing.T) {
	for _, r := range []struct {
		name  string
		kinds []ast.Kind
	}{{"jsx-fragments", []ast.Kind{ast.KindJsxFragment, ast.KindJsxElement, ast.KindJsxSelfClosingElement}}, {"jsx-no-undef", []ast.Kind{ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement}}, {"jsx-no-constructed-context-values", []ast.Kind{ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement}}} {
		data, e := os.ReadFile(filepath.Join(r.name, "rule.json"))
		if e != nil {
			t.Fatal(e)
		}
		var v struct {
			Name  string
			Kinds []int
		}
		if e = json.Unmarshal(data, &v); e != nil {
			t.Fatal(e)
		}
		if v.Name != "react/"+r.name || len(v.Kinds) != len(r.kinds) {
			t.Fatal("listener metadata", r.name)
		}
		for i, k := range r.kinds {
			if v.Kinds[i] != int(k) {
				t.Fatal("numeric kind", r.name)
			}
		}
	}
}

func TestNativeSyntaxContracts(t *testing.T) {
	repository, e := filepath.Abs("../../../..")
	if e != nil {
		t.Fatal(e)
	}
	directory := t.TempDir()
	if requested := os.Getenv("ADAMIC_WAVE21_JSX_CONTRACT_ARTIFACTS"); requested != "" {
		directory = requested
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	modules := filepath.Join(repository, "stage1/cohere/typeaware/wave21_jsx")
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	sanitized := h.archive("checker-asan", "", true)
	var source, expected strings.Builder
	fmt.Fprintf(&source, "import {quote} from %s;\nimport {syntaxKinds as fragments} from %s;\nimport {syntaxKinds as context} from %s;\nimport {syntaxKinds as undef} from %s;\n", strconv.Quote(filepath.Join(modules, "syntax.a")), strconv.Quote(filepath.Join(modules, "jsx-fragments/rule.a")), strconv.Quote(filepath.Join(modules, "jsx-no-constructed-context-values/rule.a")), strconv.Quote(filepath.Join(modules, "jsx-no-undef/rule.a")))
	for _, text := range []string{"ordinary", "\a\b\f\n\r\t\v\x01\x7f", "\"\\", "É테스트😀", "\u00a0\u0085\u200b\u200c\u2028\u2029", "\U000e0001\U0001f600"} {
		encoded, err := json.Marshal(text)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&source, "console.log(quote(%s));\n", encoded)
		fmt.Fprintln(&expected, strconv.Quote(text))
	}
	source.WriteString("console.log(fragments.join(','));console.log(context.join(','));console.log(undef.join(','));\n")
	fmt.Fprintf(&expected, "%d,%d,%d\n%d,%d\n%d,%d\n", ast.KindJsxFragment, ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement, ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement)
	entry := h.write("contracts.a", source.String())
	for _, san := range []bool{false, true} {
		name := "normal"
		a := archive
		if san {
			name = "asan"
			a = sanitized
		}
		binary := h.build(stage0, name, entry, a, san)
		got := h.must(name+"-run", exec.Command(binary))
		if len(got.stderr) != 0 || string(got.stdout) != expected.String() {
			t.Fatalf("syntax contract: %s %s", got.stdout, got.stderr)
		}
	}
	config := h.write("config.json", `{"compilerOptions":{"strict":true,"target":"ES2022"}}`)
	h.write("declarations.d.ts", "declare const bridgeProbe:number;\n")
	input := h.write("input.a", "-1;\n")
	released := filepath.Join(repository, "stage1/cohere/typeaware/testdata/released.ts")
	for _, san := range []bool{false, true} {
		name := "released-normal"
		a := archive
		if san {
			name = "released-asan"
			a = sanitized
		}
		binary := h.build(stage0, name, released, a, san)
		got := h.run(name+"-run", exec.Command(binary, config, input))
		code, ok := got.err.(*exec.ExitError)
		if !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released handle %v %s", got.err, got.stderr)
		}
	}
	overlay := h.overlay("retained-handle", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released program.")
	retained := h.archive("retained-handle", overlay, false)
	binary := h.build(stage0, "retained-probe", released, retained, false)
	got := h.must("retained-run", exec.Command(binary, config, input))
	if len(got.stderr) != 0 {
		t.Fatal("retained mutant failed outside assertion", got.stderr)
	}
	t.Log("native Unicode quoted strings and numeric listener exports match Go under normal/ASAN; released bridge handle exits 70, retained-registry mutant exits 0 and is caught")
}
