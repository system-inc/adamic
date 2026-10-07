package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func repository(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, source := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}} {
		if output, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git: %v %s", err, output)
		}
	}
	return root
}

func TestPlanAndRerun(t *testing.T) {
	t.Parallel()
	original := "// import './dep.ts'\nconst text = './dep.ts';\nimport { value } from './dep.ts';\nexport { value } from './dep.ts';\nconst later = import('./dep.ts');\ntype T = import('./dep.ts').T;\n"
	root := repository(t, map[string]string{
		"examples/main.ts": original, "examples/dep.ts": "export const value = 1; export type T = number;\n",
		"examples/existing.a":                         "import { value } from './dep.t\\u0073';\n",
		"stage1/cohere/slice/testdata/fake_test.go":   "package fake\n// *.ts\n",
		"stage1/cohere/slice/main.ts":                 "const x = 1;",
		"stage1/cohere/slice/profile_test.go":         "package profile\n// source/slice/main.ts ../../slice/main.ts\n",
		"stage1/cohere/slice/README.md":               "stage1/cohere/slice/main.ts",
		"docs/example.md":                             "[main](../examples/main.ts)\n",
		"internal/flow/flow_test.go":                  "package flow\n// ../load/testdata/0.1/compile/*.ts\n",
		"internal/load/testdata/0.1/compile/hello.ts": "console.log('hello');\n",
		"internal/load/prelude.d.ts":                  "declare const console: unknown;\n",
		"cmd/adamic-meter/testdata/corpus/refused.ts": "async function later() {}\n",
		"tsconfig.json":                               "{\"include\":[\"examples/*.ts\"]}\n",
		"CohereSettings.json":                         "{\"ignorePatterns\":[\"examples/main.ts\"]}\n",
	})
	p, err := prepare(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.renames) != 4 {
		t.Fatalf("renames: %v", p.renames)
	}
	data, _ := os.ReadFile(filepath.Join(root, "examples/main.ts"))
	if string(data) != original {
		t.Fatal("dry run wrote source")
	}
	if err := p.apply(root); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(filepath.Join(root, "examples/main.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := strings.ReplaceAll(original, "from './dep.ts'", "from './dep.a'")
	want = strings.ReplaceAll(want, "import('./dep.ts')", "import('./dep.a')")
	if string(data) != want {
		t.Fatalf("source:\n%s\nwant:\n%s", data, want)
	}
	data, _ = os.ReadFile(filepath.Join(root, "examples/existing.a"))
	if !strings.Contains(string(data), `"./dep.a"`) {
		t.Fatalf("escaped import: %s", data)
	}
	for _, name := range []string{"stage1/cohere/slice/profile_test.go", "stage1/cohere/slice/README.md", "docs/example.md", "tsconfig.json", "CohereSettings.json", "internal/flow/flow_test.go"} {
		data, _ = os.ReadFile(filepath.Join(root, name))
		if strings.Contains(string(data), ".ts") {
			t.Fatalf("stale reference in %s: %s", name, data)
		}
	}
	data, _ = os.ReadFile(filepath.Join(root, "stage1/cohere/slice/testdata/fake_test.go"))
	if string(data) != "package fake\n// *.ts\n" {
		t.Fatalf("TypeScript data glob changed: %s", data)
	}
	again, err := prepare(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(again.renames) != 0 || len(again.changes) != 0 {
		t.Fatalf("rerun: %+v", again)
	}
}

func TestRefuseCollisionAndUnknownPurpose(t *testing.T) {
	t.Parallel()
	for _, files := range []map[string]string{
		{"examples/main.ts": "const x = 1;", "examples/main.a": "const x = 2;"},
		{"tool/main.ts": "const x = 1;"},
		{"examples/main.ts": "import { broken from './other.ts';"},
	} {
		root := repository(t, files)
		if _, err := prepare(root); err == nil {
			t.Fatal("unsafe plan accepted")
		}
		for name, want := range files {
			data, _ := os.ReadFile(filepath.Join(root, name))
			if string(data) != want {
				t.Fatal("refusal changed a file")
			}
		}
	}
}

func TestSourceCorpusFilters(t *testing.T) {
	t.Parallel()
	source := `package main
import("fmt"; "strings")
func accepts(path string,directory bool)bool{return !directory && strings.HasSuffix(path,".ts")}
func main(){fmt.Println(accepts("upstream.ts",false),accepts("port.a",false),accepts("tool.js",false),accepts("directory.a",true))}
`
	root := repository(t, map[string]string{"stage1/main.go": source})
	execute := func() string {
		output, err := exec.Command("go", "run", filepath.Join(root, "stage1/main.go")).CombinedOutput()
		if err != nil {
			t.Fatalf("source filter: %v %s", err, output)
		}
		return string(output)
	}
	if got := execute(); got != "true false false false\n" {
		t.Fatalf("before: %q", got)
	}
	p, err := prepare(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.apply(root); err != nil {
		t.Fatal(err)
	}
	if got := execute(); got != "true true false false\n" {
		t.Fatalf("after: %q", got)
	}
	again, err := prepare(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(again.changes) != 0 {
		t.Fatalf("source filter is not idempotent: %+v", again.changes)
	}
}

func TestStage3PurposeAndStatus(t *testing.T) {
	t.Parallel()
	root := repository(t, map[string]string{
		"docs/tsc-strictness.md":                "upstream main.ts parser.ts",
		"stage1/main.ts":                        "console.log(1);",
		"stage3/fixtures/paths_test.go":         "package fixtures\n// main.ts is a deliberately invalid path\n",
		"stage3/fixtures/enums/value.ts":        "enum Value { One }; console.log(Value.One);",
		"stage3/fixtures/enums/status.json":     `[{"file":"value.ts","stage0":{"what":"stage3/fixtures/enums/value.ts:1"},"node":{"stdout":"value.ts\\nmain.ts"},"tsc":["src/compiler/types.ts"]}]`,
		"stage3/drivers/tsc/corpus/input.ts":    "const source = 1;",
		"stage3/drivers/tsc/corpus/bad/input.a": "const invalid = foo<?>;",
		"bridge/tsgo/testdata/sample.ts":        "export const value = 1;",
	})
	p, err := prepare(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.renames) != 2 || len(p.changes) != 1 || len(p.changes[0].edits) != 2 {
		t.Fatalf("stage3 plan: %+v", p)
	}
	if err := p.apply(root); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "stage3/fixtures/enums/status.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"file":"value.a"`) || !strings.Contains(string(data), "value.a:1") || !strings.Contains(string(data), "src/compiler/types.ts") || !strings.Contains(string(data), "value.ts") {
		t.Fatalf("status: %s", data)
	}
	if _, err := os.Stat(filepath.Join(root, "bridge/tsgo/testdata/sample.ts")); err != nil {
		t.Fatal(err)
	}
	again, err := prepare(root)
	if err != nil || len(again.renames) != 0 || len(again.changes) != 0 {
		t.Fatalf("repeat: %+v %v", again, err)
	}
}

func TestGeneratedAdamicNames(t *testing.T) {
	t.Parallel()
	root := repository(t, map[string]string{"stage1/cohere/markdownblocks/data.ts": "export const data = 1;", "stage1/cohere/markdownblocks/tools/generate_data/main.go": `package main
const output = "stage1/cohere/markdownblocks/data.ts"
const suffix = ".ts"
const temporary = "adamic-data-*.ts"
`})
	p, err := prepare(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.changes) != 1 || len(p.changes[0].edits) != 2 {
		t.Fatalf("generator: %+v", p)
	}
	if err := p.apply(root); err != nil {
		t.Fatal(err)
	}
	again, err := prepare(root)
	if err != nil || len(again.changes) != 0 {
		t.Fatalf("generator repeat: %+v %v", again, err)
	}
}

func TestOverlappingEditsRefuseBeforeWriting(t *testing.T) {
	t.Parallel()
	root := repository(t, map[string]string{
		"stage1/main.ts": "console.log(1);",
		"stage1/path_test.go": `package paths
import "strings"
func matches() bool { return strings.HasSuffix("main.ts", ".ts") }
`,
	})
	if _, err := prepare(root); err == nil || !strings.Contains(err.Error(), "overlapping reference edits") {
		t.Fatalf("overlap: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "stage1/main.ts")); err != nil {
		t.Fatal(err)
	}
}

func TestMixedTypeScriptConfig(t *testing.T) {
	t.Parallel()
	root := repository(t, map[string]string{
		"stage1/cohere/typeaware/testdata/tsconfig.json": `{"compilerOptions":{"strict":true},"include":["*.ts"]}`,
		"stage1/cohere/typeaware/testdata/probe.ts":      "console.log(1);",
	})
	p, err := prepare(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.apply(root); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "stage1/cohere/typeaware/testdata/tsconfig.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"include":["*.ts","*.a"]`) || !strings.Contains(string(data), `"sourceExtensions":[".a"]`) || !strings.Contains(string(data), "prelude.d.ts") {
		t.Fatalf("mixed config: %s", data)
	}
	again, err := prepare(root)
	if err != nil || len(again.changes) != 0 {
		t.Fatalf("mixed repeat: %+v %v", again, err)
	}
}

func TestRegistryRenameWitness(t *testing.T) {
	t.Parallel()
	source := `package registry
func witness() {
 os.Rename(filepath.Join(directory, "rule.ts"), filepath.Join(directory, "rule.a"))
 os.WriteFile(filepath.Join(directory, "rule.ts"), []byte("stale rename"), 0644)
}
`
	root := repository(t, map[string]string{
		"stage1/cohere/lint/rules/no-debugger/rule.ts": "export const x = 1;",
		"stage1/cohere/lint/registry/registry_test.go": source,
	})
	plan, err := prepare(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.changes) != 1 || len(plan.changes[0].edits) != 1 {
		t.Fatalf("stale TypeScript witness changed: %+v", plan.changes)
	}
	if err := plan.apply(root); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/registry/registry_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `os.Rename(filepath.Join(directory, "rule.a")`) || !strings.Contains(string(data), `os.WriteFile(filepath.Join(directory, "rule.ts")`) {
		t.Fatalf("registry witness: %s", data)
	}
}

func TestSerializationTransition(t *testing.T) {
	t.Parallel()
	source := `package main
import("fmt"; "os"; "path/filepath")
func serializationPort(directory string) error {
 for _, name := range []string{"rule.a"} {
  if err := os.WriteFile(filepath.Join(directory,"rules/no-debugger",name), []byte("replacement"),0644); err != nil {return err}
 }
 if err := os.Remove(filepath.Join(directory,"rules/no-debugger/rule.ts")); err != nil {return err}
 return nil
}
func main(){ directory:=os.Args[1]; if err:=serializationPort(directory);err!=nil {panic(err)};data,err:=os.ReadFile(filepath.Join(directory,"rules/no-debugger/rule.a"));if err!=nil{panic(err)};fmt.Print(string(data)) }
`
	root := repository(t, map[string]string{
		"stage1/cohere/lint/rules/no-debugger/rule.ts": "export const x=1;",
		"stage1/cohere/lint/harness_test.go":           source,
	})
	p, err := prepare(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.changes) != 1 || len(p.changes[0].edits) != 1 {
		t.Fatalf("transition plan: %+v", p.changes)
	}
	if err := p.apply(root); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/harness_test.go"))
	if err != nil {
		t.Fatal(err)
	}

	program := filepath.Join(t.TempDir(), "transition.go")
	if err := os.WriteFile(program, data, 0644); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("go", "run", program, filepath.Join(root, "stage1/cohere/lint")).CombinedOutput()
	if err != nil || string(output) != "replacement" {
		t.Fatalf("replacement lost: %v %s", err, output)
	}
	repeat, err := prepare(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(repeat.renames) != 0 || len(repeat.changes) != 0 {
		t.Fatalf("transition repeat: %+v", repeat)
	}
}
