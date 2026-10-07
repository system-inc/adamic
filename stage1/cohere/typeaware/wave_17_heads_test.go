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

// Not parallel: native and sanitizer builds share this machine.
func TestWave17HeadJudgmentsAndParserBoundary(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE17_HEAD_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	os.MkdirAll(directory, 0755)
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_17_projected_suite.a")
	native := h.build(stage0, "head-projected", entry, archive, false)
	oracle := volumeOracle(h, "head-oracle", "oracle_wave_17.go")
	h.write("root.d.ts", "export {};\n")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","jsx":"preserve","lib":["ES2022"]},"files":["root.d.ts"]}`)
	unique := map[string]bool{}
	for _, name := range []string{"no_duplicate_head", "no_script_component_in_head", "no_async_client_component"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/next", name+"_test.go"), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			literal, ok := n.(*goast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			s, err := strconv.Unquote(literal.Value)
			if err == nil && strings.Contains(s, ";") && strings.Contains(s, "<") && (strings.Contains(s, "import ") || strings.Contains(s, "const Head") || strings.Contains(s, "export default")) {
				unique[s] = true
			}
			return true
		})
	}
	for _, s := range []string{
		"import Head from 'next/head';export const x=<div><Head/><Head/><Head/></div>;",
		"import Head from 'widgets';export const x=<div><Head/><Head/></div>;",
		"import {Header as Head} from 'widgets';export const x=<><Head/><Head/></>;",
		"import {Head as H} from 'widgets';export const x=<><H/><H/></>;",
		"import * as Head from 'widgets';export const x=<><Head/><Head/></>;",
		"import type Head from 'widgets';export const x=<><Head/><Head/></>;",
		"import Head from 'next/head';function f(){const Head=()=>null;return <><Head/><Head/></>};",
		"import Head from 'next/head';export const x=<><Head.Sub/><Head.Sub/></>;",
		"import H from 'next/head';export const x=<H><Script/><Script/><div><Script/></div><S/><Foo.Script/></H>;",
		"import Head from 'other';export const x=<Head><Script/></Head>;",
		"import Head from 'next/head';export const x=<Head><><Script/></></Head>;",
		"/* 世界 🌍 */\r\nimport Head from 'next/head';export const x=<Head><Script/></Head>;\r\n",
		"'use client';export default async function MyComponent(){return <div/>}",
	} {
		unique[s] = true
	}
	keys := []string{}
	for s := range unique {
		keys = append(keys, s)
	}
	sort.Strings(keys)
	paths := []string{}
	for i, s := range keys {
		physical := h.write(fmt.Sprintf("control-%03d.a", i), s+"\n")
		path := strings.TrimSuffix(physical, ".a") + ".tsx"
		os.Remove(path)
		if err := os.Symlink(filepath.Base(physical), path); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	manifest := h.write("all.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid-sources", exec.Command(oracle, config, manifest, "--valid-sources"))
	paths = strings.Fields(string(valid.stdout))
	if len(paths) == 0 {
		t.Fatal("empty controls")
	}
	manifest = h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	for i, path := range paths {
		one := h.write("projection.manifest", path+"\n")
		r := h.must(fmt.Sprintf("tree-%03d", i), exec.Command(oracle, config, one, "--tree"))
		if err := os.WriteFile(path+".tree", r.stdout, 0644); err != nil {
			t.Fatal(err)
		}
	}
	truth := h.compare("projected-controls", oracle, native, config, manifest)
	for _, name := range []string{"no-duplicate-head", "no-script-component-in-head", "no-async-client-component"} {
		if !bytes.Contains(truth.stdout, []byte("\t@next/next/"+name+"\t")) {
			t.Fatal("no positive " + name)
		}
	}
	t.Logf("%d projected JSX controls; decisions are native, parser is explicitly bypassed", len(paths))
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "head-projected-asan", entry, sanitized, true)
	h.compare("projected-controls-asan", oracle, asan, config, manifest)
	for _, m := range []struct{ name, file, from, to string }{
		{"duplicate", "no_duplicate_head.a", "this.occurrences > 1", "this.occurrences > 2"},
		{"script", "no_script_component_in_head.a", "rules.parser.node(tag).text === 'Script'", "rules.parser.node(tag).text === 'S'"},
	} {
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware", m.file))
		if err != nil {
			t.Fatal(err)
		}
		s := strings.ReplaceAll(string(data), "'./", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
		if strings.Count(s, m.from) != 1 {
			t.Fatal("anchor changed " + m.name)
		}
		rule := h.write("mutant-"+m.name+".a", strings.Replace(s, m.from, m.to, 1))
		data, err = os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		s = strings.ReplaceAll(string(data), "'./", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
		s = strings.ReplaceAll(s, "'../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")
		s = strings.Replace(s, filepath.Join(repository, "stage1/cohere/typeaware", m.file), rule, 1)
		e := h.write("mutant-entry-"+m.name+".a", s)
		b := h.build(stage0, "mutant-"+m.name, e, sanitized, true)
		got := h.must("mutant-"+m.name+"-run", exec.Command(b, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived or sanitizer caught it")
		}
		t.Logf("%s mutant exits 0, empty stderr; Go bytes catch byte %d", m.name, firstDifference(got.stdout, truth.stdout))
	}
	real := h.build(stage0, "real-parser", filepath.Join(repository, "stage1/cohere/typeaware/wave_17_suite.a"), sanitized, true)
	one := h.write("boundary.manifest", paths[0]+"\n")
	got := h.run("parser-boundary", exec.Command(real, config, one))
	exit, ok := got.err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte("parser slice")) {
		t.Fatalf("unexpected parser boundary: %v %s", got.err, got.stderr)
	}
	t.Log("end-to-end JSX remains blocked: shared parser slice refusal, panic 70")
}
