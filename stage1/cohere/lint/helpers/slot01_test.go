package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const contextDependency = "github.com/system-inc/cohere/internal/lint/rules/structure.FileContextFor"

func slot01Fixture(t *testing.T, dependency string, consumers int, mode string) (string, []byte) {
	t.Helper()
	var readiness struct {
		Remaining []struct {
			Rule             string
			RemainingHelpers []string `json:"remaining_helpers"`
		}
	}
	data, err := os.ReadFile("readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &readiness); err != nil {
		t.Fatal(err)
	}
	root, _ := filepath.Abs("../../../../cohere")
	var paths []string
	for _, rule := range readiness.Remaining {
		if !slicesContain(rule.RemainingHelpers, dependency) {
			continue
		}
		namespace, name, has := strings.Cut(rule.Rule, "/")
		if !has {
			namespace = "core"
			name = rule.Rule
		}
		if namespace == "@next" {
			name = strings.TrimPrefix(name, "next/")
			namespace = "next"
		}
		if namespace == "better-tailwindcss" {
			namespace = "tailwind"
		}
		if namespace == "@typescript-eslint" {
			namespace = "typescript"
		}
		path := filepath.Join(root, "internal/lint/rules", namespace, strings.ReplaceAll(name, "-", "_")+"_test.go")
		if name == "react-element-no-anchor" || name == "react-element-no-horizontal-rule" {
			path = filepath.Join(root, "internal/lint/rules/structure/react_no_intrinsic_element_test.go")
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("consumer %s: %v", rule.Rule, err)
		}
		paths = append(paths, path)
	}
	if len(paths) != consumers {
		t.Fatalf("consumer ledger drift: %d", len(paths))
	}
	directory := t.TempDir()
	manifest, _ := json.Marshal(paths)
	manifestPath := filepath.Join(directory, "manifest.json")
	if err := os.WriteFile(manifestPath, manifest, 0644); err != nil {
		t.Fatal(err)
	}
	source, _ := filepath.Abs("testdata/slot01_oracle.go")
	bridge, _ := filepath.Abs("testdata/slot01_react.go")
	virtual := filepath.Join(root, "adamic_slot01_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/ecmascript/react/adamic_slot01.go"): bridge}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "go-oracle")
	run(t, root, "go", "build", "-overlay="+overlayPath, "-o", binary, virtual)
	cases := filepath.Join(directory, "cases.jsonl")
	expected := filepath.Join(directory, "expected.log")
	cmd := exec.Command(binary, manifestPath, cases, expected, mode)
	log, err := os.Create(filepath.Join(directory, "generation.log"))
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout = log
	cmd.Stderr = log
	err = cmd.Run()
	log.Close()
	generated, _ := os.ReadFile(log.Name())
	t.Logf("Go fixture observations:\n%s", generated)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(expected)
	if err != nil {
		t.Fatal(err)
	}
	if len(want) == 0 {
		t.Fatal("empty Go observations")
	}
	return cases, want
}
func slicesContain(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
func slot01Build(t *testing.T, entry string) string {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "slot01")
	if err := native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}

// Not parallel: keep sanitized native builds and corpus outputs bounded on cloud workers.
func TestSlot01FileContextMatchesCohere(t *testing.T) {
	slot01Check(t, contextDependency, 18, "context", "structure_file_context.a", "fileName.split('\\\\').join('/')", "fileName")
}

// Not parallel: keep sanitized native builds and corpus outputs bounded on cloud workers.
func TestSlot01Es6ComponentClassMatchesCohere(t *testing.T) {
	slot01Check(t, "github.com/system-inc/cohere/internal/lint/ecmascript/react.IsEs6ComponentClass", 14, "class", "react_es6_component_class.a", "node.kind !== 'ClassExpression'", "true")
}

// Not parallel: keep sanitized native builds and corpus outputs bounded on cloud workers.
func TestSlot01TailwindDefaultsMatchCohere(t *testing.T) {
	slot01Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.DefaultClassLiteralSettings", 12, "defaults", "tailwind_default_class_literal_settings.a", "attributeNames: ['class', 'className']", "attributeNames: ['className']")
}

func slot01Check(t *testing.T, dependency string, consumers int, mode, mutantFile, old, replacement string) {
	cases, want := slot01Fixture(t, dependency, consumers, mode)
	entry, _ := filepath.Abs("slot01_main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases, mode), want)
	compare(t, run(t, "", slot01Build(t, entry), cases, mode), want)
	t.Logf("%d Go answers matched Node and sanitized native", bytes.Count(want, []byte("\n")))
	directory := t.TempDir()
	for _, file := range []string{"slot01_main.a", "structure_file_context.a", "react_es6_component_class.a", "tailwind_default_class_literal_settings.a", "options_json.ts"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if file == mutantFile {
			if strings.Count(string(data), old) != 1 {
				t.Fatal("mutant anchor drift")
			}
			data = []byte(strings.Replace(string(data), old, replacement, 1))
		}
		if file == "slot01_main.a" {
			data = []byte(strings.ReplaceAll(string(data), "./options_json.ts", "./options_json.a"))
		}
		if file == "options_json.ts" {
			file = "options_json.a"
		}
		if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutant := run(t, "", slot01Build(t, filepath.Join(directory, "slot01_main.a")), cases, mode)
	if bytes.Equal(mutant, want) {
		t.Fatal("compiled semantic mutant survived")
	}
	slot01MutantWitness(t, mutant, want, mutantFile)
	if mode == "defaults" {
		for _, field := range []struct{ name, literal string }{
			{"attributeNames", "['class', 'className']"},
			{"calleeNames", "['mergeClassNames', 'createVariantClassNames']"},
			{"variablePatterns", "['.*[Cc]lassName$', '.*[Cc]lassNames$']"},
		} {
			data, err := os.ReadFile(mutantFile)
			if err != nil {
				t.Fatal(err)
			}
			anchor := field.name + ": " + field.literal
			if strings.Count(string(data), anchor) != 1 {
				t.Fatal("freshness anchor drift")
			}
			data = []byte("const sharedList: string[] = " + field.literal + ";\n" + strings.Replace(string(data), anchor, field.name+": sharedList", 1))
			if err := os.WriteFile(filepath.Join(directory, mutantFile), data, 0644); err != nil {
				t.Fatal(err)
			}
			mutated := run(t, "", slot01Build(t, filepath.Join(directory, "slot01_main.a")), cases, mode)
			if bytes.Equal(mutated, want) {
				t.Fatalf("shared %s mutant survived", field.name)
			}
			slot01MutantWitness(t, mutated, want, "shared "+field.name)
		}
	}
}

func slot01MutantWitness(t *testing.T, got, want []byte, name string) {
	t.Helper()
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Logf("compiled %s mutant caught at output line %d: got %q, Go %q", name, i+1, a[i], b[i])
			return
		}
	}
	t.Logf("compiled %s mutant caught by output length", name)
}
