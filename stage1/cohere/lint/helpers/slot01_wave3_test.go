package helpers

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: bounded native sanitizer builds and external Go captures.
func TestSlot01Wave3AttributeMatchesCohere(t *testing.T) {
	slot01Wave3Check(t, "*ClassLiteralReader.attributeValues", "attribute", "tailwind_attribute_values.a", "if (!(attributeNames.get(name.text) ?? false))", "if (false)")
}
func slot01Wave3Fixture(t *testing.T, dependency string, consumers int, mode string) (string, []byte) {
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
	source, _ := filepath.Abs("testdata/slot01_wave3_oracle.go")
	tailwindBridge, _ := filepath.Abs("testdata/slot01_wave3_tailwind.go")
	textBridge, _ := filepath.Abs("testdata/slot01_wave3_text.go")
	virtual := filepath.Join(root, "adamic_slot01_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "internal/lint/rules/tailwind/adamic_slot01.go"): tailwindBridge, virtual: source, filepath.Join(root, "internal/lint/ecmascript/text/adamic_slot01_wave3.go"): textBridge}})
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
func slot01Wave3Check(t *testing.T, symbol, mode, file, old, replacement string) {
	t.Helper()
	dependency := "github.com/system-inc/cohere/internal/lint/rules/tailwind." + symbol
	consumers := 11
	if mode == "entity" {
		dependency = "github.com/system-inc/cohere/internal/lint/ecmascript/text." + symbol
		consumers = 9
	}
	cases, want := slot01Wave3Fixture(t, dependency, consumers, mode)
	entry, _ := filepath.Abs("slot01_wave3_main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases, mode), want)
	compare(t, run(t, "", slot01Build(t, entry), cases, mode), want)
	t.Logf("%d Go output lines match Node and sanitized native", bytes.Count(want, []byte("\n")))
	directory := t.TempDir()
	for _, name := range []string{"slot01_wave3_main.a", "tailwind_attribute_values.a", "tailwind_class_values_in.a", "tailwind_class_literal_from.a", "text_decode_entity.a", "text_xhtml_entities.a", "options_json.ts"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if name == file {
			if strings.Count(string(data), old) != 1 {
				t.Fatal("mutant anchor drift")
			}
			data = []byte(strings.Replace(string(data), old, replacement, 1))
		}
		if name == "slot01_wave3_main.a" {
			data = []byte(strings.ReplaceAll(string(data), "./options_json.ts", "./options_json.a"))
		}
		if name == "options_json.ts" {
			name = "options_json.a"
		}
		if err := os.WriteFile(filepath.Join(directory, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutant := run(t, "", slot01Build(t, filepath.Join(directory, "slot01_wave3_main.a")), cases, mode)
	if bytes.Equal(mutant, want) {
		t.Fatal("compiled semantic mutant survived")
	}
	slot01MutantWitness(t, mutant, want, file)
}

// Not parallel: bounded external corpus and sanitizer builds.
func TestSlot01Wave3EntityMatchesCohere(t *testing.T) {
	slot01Wave3Check(t, "decodeEntity", "entity", "text_decode_entity.a", "value = 0x110000;", "value = 0;")
}
func TestSlot01Wave3EntitySurrogateMutant(t *testing.T) {
	slot01Wave3Check(t, "decodeEntity", "entity", "text_decode_entity.a", "value >= 0xd800 && value <= 0xdfff", "false")
}
