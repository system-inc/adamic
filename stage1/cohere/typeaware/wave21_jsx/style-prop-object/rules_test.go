package wave21style

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

func TestStylePropObject(t *testing.T) {
	repository, e := filepath.Abs("../../../../..")
	if e != nil {
		t.Fatal(e)
	}
	directory := os.Getenv("ADAMIC_WAVE21_STYLE_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if e = os.MkdirAll(directory, 0755); e != nil {
		t.Fatal(e)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	modules := filepath.Join(repository, "stage1/cohere/typeaware/wave21_jsx/style-prop-object")
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	sanitized := h.archive("checker-asan", "", true)
	data, e := os.ReadFile(filepath.Join(modules, "oracle.go.txt"))
	if e != nil {
		t.Fatal(e)
	}
	virtual := filepath.Join(repository, "cohere/adamic_wave21_style.go")
	overlay, e := json.Marshal(map[string]any{"Replace": map[string]string{virtual: h.write("oracle.go", string(data))}})
	if e != nil {
		t.Fatal(e)
	}
	oracle := filepath.Join(directory, "oracle")
	build := exec.Command("go", "build", "-overlay", h.write("oracle-overlay.json", string(overlay)), "-o", oracle, virtual)
	build.Dir = filepath.Join(repository, "cohere")
	h.must("oracle-build", build)
	config := h.write("config.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","moduleDetection":"force","lib":["ES2022"],"jsx":"preserve","noEmit":true},"include":["*.d.ts"]}`)
	h.write("react.d.ts", `declare module 'react' {export function createElement(...args:any[]):any; const React:any;export default React;} declare const React:any;`)
	tree, e := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/react/style_prop_object_test.go"), nil, 0)
	if e != nil {
		t.Fatal(e)
	}
	sources := map[string]bool{}
	goast.Inspect(tree, func(n goast.Node) bool {
		literal, ok := n.(*goast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		text, e := strconv.Unquote(literal.Value)
		if e == nil && (strings.Contains(text, "<") || strings.Contains(text, "createElement(")) && !strings.HasPrefix(text, "/repository/") {
			sources[text] = true
		}
		return true
	})
	for _, source := range []string{
		`React.createElement('div',{style:'bad'});React.createElement('div',{style:null});React.createElement('div',{style:/x/});React.createElement('div',{style:1n});`,
		`const style='bad';React.createElement('div',{style});const empty={};React.createElement('div',{style:empty});`,
		`import {createElement} from 'react';createElement('div',{style:'bad'});`,
		`import {createElement as ce} from 'react';ce('div',{style:'clean alias'});`,
		`const {createElement}=React;createElement('div',{style:'bad'});`,
		`const {createElement}=require('react');createElement('div',{style:'bad'});`,
		`const createElement=require('react').anything;createElement('div',{style:'bad'});`,
		`const createElement=React.anything;createElement('div',{style:'bad'});`,
		`declare function createElement(...args:any[]):any;import {createElement} from 'react';createElement('div',{style:'bad'});`,
		`import {createElement} from 'react';declare function createElement(...args:any[]):any;createElement('div',{style:'bad'});`,
		`const createElement=(React);createElement('div',{style:'clean'});`,
		`(React.createElement)('div',{style:'bad'});(React).createElement('div',{style:'clean'});React['createElement']('div',{style:'clean'});`,
		`const styledé='bad';React.createElement('div',{style:styledé});const q=<Foo style={styledé}/>;`,
		`React.createElement('div',{style:unknown,style:'clean later'});React.createElement('div',{style:null,style:'clean later'});`,
		`<div style="bad" xlink:style="clean" />;const q=<Foo style={(false)}/>;`,
	} {
		sources[source] = true
	}
	keys := []string{}
	for s := range sources {
		keys = append(keys, s)
	}
	sort.Strings(keys)
	paths := []string{}
	for i, s := range keys {
		paths = append(paths, h.write(fmt.Sprintf("fixture-%03d.tsx", i), s+"\n"))
	}
	all := h.write("all.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid", exec.Command(oracle, config, all, "--valid-sources"))
	paths = strings.Split(strings.TrimSpace(string(valid.stdout)), "\n")
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	t.Logf("Go-valid controls %d/%d", len(paths), len(keys))
	prepared := h.must("prepare", exec.Command(oracle, config, manifest, "--prepare", filepath.Join(directory, "controls-graphs")))
	rawManifest := h.write("controls.raw.manifest", string(prepared.stdout))
	entry := filepath.Join(modules, "program.a")
	binary := h.build(stage0, "native", entry, archive, false)
	asan := h.build(stage0, "asan", entry, sanitized, true)
	var nativeSeconds, goSeconds float64
	emitted := h.must("emit-js", exec.Command(stage0, "js", entry))
	if len(emitted.stderr) != 0 {
		t.Fatal("JavaScript emission", emitted.stderr)
	}
	javascript := h.write("emitted.mjs", string(emitted.stdout))
	var defaultWant []byte
	for _, mode := range []struct {
		name  string
		flags []string
	}{{"default", nil}, {"allowed", []string{"--allow=Foo", "--allow=MyComponent", "--allow=div"}}, {"untyped", []string{"--untyped"}}} {
		want := h.must(mode.name+"-go", exec.Command(oracle, append([]string{config, manifest}, mode.flags...)...))
		got := h.must(mode.name+"-native", exec.Command(binary, append([]string{rawManifest}, mode.flags...)...))
		nativeSeconds += got.elapsed.Seconds()
		goSeconds += want.elapsed.Seconds()
		if !bytes.Equal(got.stdout, want.stdout) || len(got.stderr) != 0 {
			t.Fatalf("%s byte %d native %s Go %s errors %s", mode.name, firstDifference(got.stdout, want.stdout), got.stdout, want.stdout, got.stderr)
		}
		checked := h.must(mode.name+"-asan", exec.Command(asan, append([]string{rawManifest}, mode.flags...)...))
		if !bytes.Equal(checked.stdout, want.stdout) || len(checked.stderr) != 0 {
			t.Fatal("sanitizer comparison", mode.name, checked.stderr)
		}
		t.Logf("%s: %d identical findings/fixes/suggestions bytes, %s", mode.name, len(want.stdout), summary(want.stdout))
		if mode.name == "default" {
			defaultWant = want.stdout
		}
		node := h.must(mode.name+"-source-node", exec.Command("node", append([]string{"--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), entry, rawManifest}, mode.flags...)...))
		if !bytes.Equal(node.stdout, want.stdout) || len(node.stderr) != 0 {
			t.Fatal("source Node differs", node.stderr)
		}
		emittedNode := h.must(mode.name+"-emitted-node", exec.Command("node", append([]string{"--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), javascript, rawManifest}, mode.flags...)...))
		if !bytes.Equal(emittedNode.stdout, want.stdout) || len(emittedNode.stderr) != 0 {
			t.Fatal("emitted JavaScript differs", emittedNode.stderr)
		}
	}
	source, e := os.ReadFile(filepath.Join(modules, "rule.a"))
	if e != nil {
		t.Fatal(e)
	}
	anchor := "kind===K_StringLiteral||"
	if strings.Count(string(source), anchor) != 1 {
		t.Fatal("rule mutant anchor")
	}
	changed := strings.Replace(string(source), anchor, "", 1)
	changed = strings.ReplaceAll(changed, "'../", "'"+filepath.Dir(modules)+"/")
	changed = strings.ReplaceAll(changed, "'./", "'"+modules+"/")
	mutantModule := h.write("rule_mutant.a", changed)
	program, e := os.ReadFile(entry)
	if e != nil {
		t.Fatal(e)
	}
	mutated := strings.Replace(string(program), "'./rule.a'", strconv.Quote(mutantModule), 1)
	mutated = strings.ReplaceAll(mutated, "'../", "'"+filepath.Dir(modules)+"/")
	mutated = strings.ReplaceAll(mutated, "'../../../../", "'"+filepath.Join(repository, "stage1")+"/")
	mutant := h.build(stage0, "rule-mutant", h.write("mutant_main.a", mutated), archive, false)
	wrong := h.must("rule-mutant-run", exec.Command(mutant, rawManifest))
	if bytes.Equal(wrong.stdout, defaultWant) || len(wrong.stderr) != 0 {
		t.Fatal("rule mutant survived or failed outside comparison", wrong.stderr)
	}
	t.Logf("literal-set mutant exits 0 empty stderr; independent complete-byte oracle catches byte %d", firstDifference(wrong.stdout, defaultWant))
	t.Logf("control execution: native raw-graph load/decisions %.6fs, full Go %.6fs, raw Go preparation %.6fs; different source pipelines", nativeSeconds, goSeconds, prepared.elapsed.Seconds())
	for _, corpus := range []struct{ name, config, manifest string }{{"compiler", os.Getenv("ADAMIC_WAVE21_COMPILER_CONFIG"), os.Getenv("ADAMIC_WAVE21_COMPILER_MANIFEST")}, {"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE21_REPOSITORY_MANIFEST")}} {
		if corpus.manifest == "" {
			continue
		}
		raw := h.must(corpus.name+"-prepare", exec.Command(oracle, corpus.config, corpus.manifest, "--prepare", filepath.Join(directory, corpus.name+"-graphs")))
		graphManifest := h.write(corpus.name+".raw.manifest", string(raw.stdout))
		want := h.must(corpus.name+"-go", exec.Command(oracle, corpus.config, corpus.manifest))
		for _, kind := range []struct{ name, binary string }{{"native", binary}, {"asan", asan}} {
			got := h.must(corpus.name+"-"+kind.name, exec.Command(kind.binary, graphManifest))
			if !bytes.Equal(got.stdout, want.stdout) || len(got.stderr) != 0 {
				t.Fatal("corpus differs", corpus.name, kind.name, firstDifference(got.stdout, want.stdout), got.stderr)
			}
			t.Logf("%s %s: %d identical bytes; %s; native %.6fs Go %.6fs", corpus.name, kind.name, len(want.stdout), summary(want.stdout), got.elapsed.Seconds(), want.elapsed.Seconds())
		}
	}
	cfg := h.write("released-config.json", `{"compilerOptions":{"strict":true,"target":"ES2022"},"include":["*.d.ts"]}`)
	input := h.write("released-input.a", "-1;\n")
	released := filepath.Join(repository, "stage1/cohere/typeaware/testdata/released.ts")
	for _, kind := range []struct {
		name, archive string
		sanitize      bool
	}{{"released-native", archive, false}, {"released-asan", sanitized, true}} {
		probe := h.build(stage0, kind.name, released, kind.archive, kind.sanitize)
		got := h.run(kind.name+"-run", exec.Command(probe, cfg, input))
		code, ok := got.err.(*exec.ExitError)
		if !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatal("released handle", got.err, got.stderr)
		}
	}
	overlayPath := h.overlay("retained", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released program.")
	retained := h.archive("retained", overlayPath, false)
	probe := h.build(stage0, "retained-probe", released, retained, false)
	got := h.must("retained-run", exec.Command(probe, cfg, input))
	if len(got.stderr) != 0 {
		t.Fatal("retained mutant error", got.stderr)
	}
	t.Log("released program panic 70 normal/ASAN; retained registry mutant exits 0 and is caught")
}

func TestStyleListeners(t *testing.T) {
	data, e := os.ReadFile("rule.json")
	if e != nil {
		t.Fatal(e)
	}
	var descriptor struct {
		Name  string
		Kinds []string
	}
	if e = json.Unmarshal(data, &descriptor); e != nil {
		t.Fatal(e)
	}
	expected := []ast.Kind{ast.KindJsxAttribute, ast.KindCallExpression}
	if descriptor.Name != "react/style-prop-object" || len(descriptor.Kinds) != len(expected) {
		t.Fatal("descriptor")
	}
	for i, k := range expected {
		if descriptor.Kinds[i] != strings.TrimPrefix(k.String(), "Kind") {
			t.Fatal("named listener")
		}
	}
	for name, value := range map[string]ast.Kind{"NumericLiteral": ast.KindNumericLiteral, "BigIntLiteral": ast.KindBigIntLiteral, "TrueKeyword": ast.KindTrueKeyword, "FalseKeyword": ast.KindFalseKeyword} {
		s, e := os.ReadFile("kinds.a")
		if e != nil || !strings.Contains(string(s), fmt.Sprintf("export const %s=%d;", name, int(value))) {
			t.Fatal("raw tag", name)
		}
	}
}
