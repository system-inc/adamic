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

// Not parallel: native archives, corpus comparisons and timings share this machine.
func TestWave17AgreementMutantAndNativeJSX(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE17_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_17_suite.a")
	binary := h.build(stage0, "wave17", entry, archive, false)
	oracle := volumeOracle(h, "wave17-oracle", "oracle_wave_17.go")
	h.write("config-root.d.ts", "export {};\n")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ES2022"],"jsx":"preserve","noEmit":true},"files":["config-root.d.ts"]}`)
	sources := map[string]bool{}
	tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/next/no_async_client_component_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	goast.Inspect(tree, func(n goast.Node) bool {
		row, ok := n.(*goast.CompositeLit)
		if !ok || len(row.Elts) < 2 {
			return true
		}
		value, ok := row.Elts[1].(*goast.BasicLit)
		if !ok || value.Kind != token.STRING {
			return true
		}
		text, err := strconv.Unquote(value.Value)
		if err == nil && strings.Contains(text, "export default") && !strings.Contains(text, "<") {
			sources[text] = true
		}
		return true
	})
	for _, source := range []string{
		"/* 世界 🌍 */\r\n'use client';export default async function Émile(){return null}\r\n",
		"'use client';export default async function 𐐀Name(){return null}",
		"'use client';export default async function ǅName(){return null}",
		"'use client';export default async function 𐐨Name(){return null}",
		"'use client';interface MyThing{};async function MyThing(){};export default MyThing",
		"'use client';async function MyThing(){};interface MyThing{};export default MyThing",
		"'use client';namespace MyThing{};async function MyThing(){};export default MyThing",
		"'use client';const MyThing=async()=>null;namespace MyThing{};export default MyThing",
		"'use client';const MyThing=async()=>null;export = MyThing;",
	} {
		sources[source] = true
	}
	keys := make([]string, 0, len(sources))
	for s := range sources {
		keys = append(keys, s)
	}
	sort.Strings(keys)
	paths := []string{}
	for i, s := range keys {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), s+"\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("\t@next/next/no-async-client-component\t")) {
		t.Fatal("no positive async control")
	}
	t.Logf("%d non-JSX upstream and edge controls", len(paths))
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave17-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Keep imports pointing to the original dependencies, change only the report end.
	original, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/no_async_client_component.a"))
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(original), "'./", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
	anchor := "this.rules.byte(node.end),"
	if strings.Count(source, anchor) != 1 {
		t.Fatal("mutant anchor changed")
	}
	mutantRule := h.write("mutant_async.a", strings.Replace(source, anchor, "this.rules.byte(node.end) + 1,", 1))
	original, err = os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	source = strings.ReplaceAll(string(original), "'./", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
	source = strings.ReplaceAll(source, "'../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")
	source = strings.Replace(source, filepath.Join(repository, "stage1/cohere/typeaware/no_async_client_component.a"), mutantRule, 1)
	mutantEntry := h.write("mutant_suite.a", source)
	mutant := h.build(stage0, "async-mutant", mutantEntry, archive, true)
	got := h.must("async-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("range mutant survived or sanitizer caught it")
	}
	t.Logf("async end+1 mutant: exit 0, empty stderr; Go byte oracle catches byte %d", firstDifference(got.stdout, truth.stdout))
	// The integrated shared parser now holds these positive JSX witnesses.
	// Require complete Go bytes rather than the historical parser refusal.
	for _, probe := range []struct{ name, rule, source string }{
		{"duplicate-head", "no-duplicate-head", "import Head from 'next/head';export const value=<div><Head/><Head/></div>;"},
		{"script-in-head", "no-script-component-in-head", "import Head from 'next/head';export const value=<Head><Script/></Head>;"},
		{"async-jsx", "no-async-client-component", "'use client';export default async function MyComponent(){return <></>}"},
	} {
		path := h.write(probe.name+".tsx", probe.source)
		list := h.write(probe.name+".manifest", path+"\n")
		want := h.must(probe.name+"-go", exec.Command(oracle, config, list))
		if !bytes.Contains(want.stdout, []byte("\t@next/next/"+probe.rule+"\t")) {
			t.Fatalf("%s witness has no Go finding", probe.name)
		}
		got := h.must(probe.name+"-native", exec.Command(asan, config, list))
		if len(got.stderr) != 0 || !bytes.Equal(got.stdout, want.stdout) {
			t.Fatalf("%s native JSX differs at byte %d: %s", probe.name, firstDifference(got.stdout, want.stdout), got.stderr)
		}
		t.Logf("%s: native JSX agrees on %d complete Go finding bytes; %s", probe.name, len(got.stdout), summary(want.stdout))
	}
	// The question used by this rule must reject a released program.
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const p=tsgoProgram(args[0]??'',[file]);tsgoRelease(p);
console.log(tsgoInspect(p,file,0,1,'Identifier','binding-declarations'));`)
	probe := h.write("released-input.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got = h.run("released-run", exec.Command(stale, config, probe))
	if exit, ok := got.err.(*exec.ExitError); !ok || exit.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released roots.")
	staleArchive := h.archive("released-checker", overlay, false)
	stale = h.build(stage0, "released-mutant", released, staleArchive, false)
	got = h.must("released-mutant-run", exec.Command(stale, config, probe))
	if len(got.stderr) != 0 {
		t.Fatal("released mutant did not finish cleanly")
	}
	t.Log("released handle: panic 70; registry mutant exits 0, caught by required panic")
	for _, population := range []struct{ name, root, config, manifest string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json"), "repository.manifest"},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), "compiler.manifest"},
	} {
		if population.root == "" {
			t.Fatal("ADAMIC_TYPESCRIPT_SOURCE is required for wave17")
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
		for _, b := range []string{oracle, binary} {
			cmd := exec.Command(b, population.config, list)
			cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			r := h.must("timing-"+population.name+"-"+filepath.Base(b), cmd)
			t.Logf("%s %s: whole process %.6fs; %s; %s", population.name, filepath.Base(b), r.elapsed.Seconds(), summary(r.stdout), strings.TrimSpace(string(r.stderr)))
		}
	}
}
