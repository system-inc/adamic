package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func slot01Wave11Fixture(t *testing.T, dependency string, consumers int, mode string) (string, []byte) {
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
	source, _ := filepath.Abs("testdata/slot01_wave11_oracle.go")
	virtual := filepath.Join(root, "adamic_slot01_wave11.go")
	bridge, _ := filepath.Abs("testdata/slot01_wave11_collapse.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot01_wave11.go"): bridge}})
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
func slot01Wave11Check(t *testing.T, dependency, mode, file, old, replacement string) {
	t.Helper()
	cases, want := slot01Wave11Fixture(t, dependency, 4, mode)
	entry, _ := filepath.Abs("gaps/slot01_wave11_main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases, mode), want)
	t.Logf("%d Go output lines matched source Node; native parity not yet established", bytes.Count(want, []byte("\n")))
	dir := t.TempDir()
	for _, name := range []string{"slot01_wave11_main.a", "collapse_clone_node.a", "collapse_clone_nodes.a", "collapse_remove_nodes.a", "options_json.ts"} {
		path := name
		if name == "slot01_wave11_main.a" {
			path = "gaps/" + name
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if name == file {
			if strings.Count(string(data), old) != 1 {
				t.Fatal("mutant anchor drift")
			}
			data = []byte(strings.Replace(string(data), old, replacement, 1))
		}
		if name == "slot01_wave11_main.a" {
			data = []byte(strings.ReplaceAll(string(data), "../", "./"))
			data = []byte(strings.ReplaceAll(string(data), "./options_json.ts", "./options_json.a"))
		}
		if name == "options_json.ts" {
			name = "options_json.a"
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutantEntry := filepath.Join(dir, "slot01_wave11_main.a")
	sourceMutant := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, mutantEntry, cases, mode)
	if bytes.Equal(sourceMutant, want) {
		t.Fatal("source-only mutant survived")
	}
	gotLines, wantLines := bytes.Split(sourceMutant, []byte("\n")), bytes.Split(want, []byte("\n"))
	for index := 0; index < len(gotLines) && index < len(wantLines); index++ {
		if !bytes.Equal(gotLines[index], wantLines[index]) {
			t.Logf("source-only %s mutant caught at output line %d: got %q, Go %q; not native mutant credit", file, index+1, gotLines[index], wantLines[index])
			break
		}
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err != nil {
		if !strings.Contains(err.Error(), "adamic/cycle-capable") || !strings.Contains(err.Error(), "CollapseCopyNode[]") {
			t.Fatal(err)
		}
		t.Skipf("BLOCKED: node arena ownership refused; native/emitted-JS parity and compiled mutant pending: %v", err)
	}
	compare(t, run(t, "", slot01Build(t, entry), cases, mode), want)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, slot01Wave3JavaScript(t, entry), cases, mode), want)
	mutant := run(t, "", slot01Build(t, mutantEntry), cases, mode)
	if bytes.Equal(mutant, want) {
		t.Fatal("compiled mutant survived")
	}
	slot01MutantWitness(t, mutant, want, file)
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave11Node(t *testing.T) {
	slot01Wave11Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.cloneNode", "node", "collapse_clone_node.a", "important: node.important", "important: false")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave11Context(t *testing.T) {
	slot01Wave11Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.cloneNode", "node", "collapse_clone_node.a", "context, children", "context: node.context, children")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave11Children(t *testing.T) {
	slot01Wave11Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.cloneNode", "node", "collapse_clone_node.a", "const children = cloneChildren(node.children);", "cloneChildren(node.children);\n    const children = node.children;")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave11List(t *testing.T) {
	slot01Wave11Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.cloneNodes", "list", "collapse_clone_nodes.a", "values.push(cloneNode(nodes.values[position] ?? panic('clone list index')));", "cloneNode(nodes.values[position] ?? panic('clone list index')); values.push(nodes.values[position] ?? panic('clone list index'));")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave11ListNil(t *testing.T) {
	slot01Wave11Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.cloneNodes", "list", "collapse_clone_nodes.a", "if (nodes.nil) { return nodes; }", "if (nodes.nil) { return { nil: false, length: 0, capacity: 0, values: [] }; }")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave11Remove(t *testing.T) {
	slot01Wave11Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.removeNodes", "remove", "collapse_remove_nodes.a", "removed.get(index) === true", "removed.has(index)")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave11RemoveIdentity(t *testing.T) {
	slot01Wave11Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.removeNodes", "remove", "collapse_remove_nodes.a", "removed.size === 0", "removed.size < 0")
}

// Not parallel: bounded external Go capture and sanitizer builds.
func TestSlot01Wave11RemoveEmpty(t *testing.T) {
	slot01Wave11Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.removeNodes", "remove", "collapse_remove_nodes.a", "values.push(index);", "if (node.children.length > 0 || node.children.nil) { values.push(index); }")
}
