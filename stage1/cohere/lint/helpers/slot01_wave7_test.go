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

func slot01Wave7Fixture(t *testing.T, dependency string, consumers int, mode string) (string, []byte) {
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
	source, _ := filepath.Abs("testdata/slot01_wave7_oracle.go")
	virtual := filepath.Join(root, "adamic_slot01_wave7.go")
	bridge, _ := filepath.Abs("testdata/slot01_wave7_collapse.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot01_wave7.go"): bridge}})
	if mode == "theme" {
		walkPath := filepath.Join(root, "internal/lint/rules/tailwind/collapse/walk.go")
		data, err := os.ReadFile(walkPath)
		if err != nil {
			t.Fatal(err)
		}
		anchor := "switch visit(node) {"
		if strings.Count(string(data), anchor) != 1 {
			t.Fatal("walk trace anchor drift")
		}
		traced := filepath.Join(directory, "walk.go")
		if err := os.WriteFile(traced, []byte(strings.Replace(string(data), anchor, "action := visit(node)\n        wave7Actions = append(wave7Actions, int(action))\n        switch action {", 1)), 0644); err != nil {
			t.Fatal(err)
		}
		designPath := filepath.Join(root, "internal/lint/rules/tailwind/collapse/design_system.go")
		data, err = os.ReadFile(designPath)
		if err != nil {
			t.Fatal(err)
		}
		anchor = "collector.theme.Add("
		if strings.Count(string(data), anchor) != 1 {
			t.Fatal("add trace anchor drift")
		}
		designTraced := filepath.Join(directory, "design_system.go")
		if err := os.WriteFile(designTraced, []byte(strings.Replace(string(data), anchor, "wave7ThemeAdd(collector.theme, ", 1)), 0644); err != nil {
			t.Fatal(err)
		}
		overlay, _ = json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot01_wave7.go"): bridge, walkPath: traced, designPath: designTraced}})
	}
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
func slot01Wave7Check(t *testing.T, dependency, mode, file, old, replacement string) {
	t.Helper()
	cases, want := slot01Wave7Fixture(t, dependency, 6, mode)
	entry, _ := filepath.Abs("slot01_wave7_main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases, mode), want)
	compare(t, run(t, "", slot01Build(t, entry), cases, mode), want)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, slot01Wave3JavaScript(t, entry), cases, mode), want)
	t.Logf("%d Go output lines matched source Node, sanitized native and emitted JavaScript", bytes.Count(want, []byte("\n")))
	dir := t.TempDir()
	for _, name := range []string{"slot01_wave7_main.a", "collapse_ingest_utility_block.a", "collapse_ingest_theme_block.a", "options_json.ts"} {
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
		if name == "slot01_wave7_main.a" {
			data = []byte(strings.ReplaceAll(string(data), "./options_json.ts", "./options_json.a"))
		}
		if name == "options_json.ts" {
			name = "options_json.a"
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutant := run(t, "", slot01Build(t, filepath.Join(dir, "slot01_wave7_main.a")), cases, mode)
	if bytes.Equal(mutant, want) {
		t.Fatal("compiled mutant survived")
	}
	slot01MutantWitness(t, mutant, want, file)
}

// Not parallel: bounded consumer capture and sanitizer builds.
func TestSlot01Wave7UtilityMatchesCohere(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*stylesheetCollector.ingestUtilityBlock", "utility", "collapse_ingest_utility_block.a", "(collector.roots.get(root) ?? 0) | kind", "kind")
}
func TestSlot01Wave7UtilityBodyMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*stylesheetCollector.ingestUtilityBlock", "utility", "collapse_ingest_utility_block.a", "collector.statics.set(root, nodes);", "collector.statics.set(root, nodes.slice());")
}
func TestSlot01Wave7UtilityNameMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*stylesheetCollector.ingestUtilityBlock", "utility", "collapse_ingest_utility_block.a", "root.includes('*')", "false")
}

// Not parallel: actual Go visitor observations and sanitized compiler builds.
func TestSlot01Wave7ThemeMatchesCohere(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*stylesheetCollector.ingestThemeBlock", "theme", "collapse_ingest_theme_block.a", "child.name === '@keyframes'", "child.name === '@not-keyframes'")
}
func TestSlot01Wave7ThemePrefixMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*stylesheetCollector.ingestThemeBlock", "theme", "collapse_ingest_theme_block.a", "theme.prefix = parsed.prefix;", "")
}
func TestSlot01Wave7ThemePropertyMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*stylesheetCollector.ingestThemeBlock", "theme", "collapse_ingest_theme_block.a", "child.property.startsWith('--')", "true")
}

func TestSlot01Wave7ThemeUnescapeMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*stylesheetCollector.ingestThemeBlock", "theme", "collapse_ingest_theme_block.a", "dependencies.unescape(child.property)", "child.property")
}
func TestSlot01Wave7ThemeErrorStopMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*stylesheetCollector.ingestThemeBlock", "theme", "collapse_ingest_theme_block.a", "error = `${path}: ${addError}`; return 2;", "error = `${path}: ${addError}`; return 0;")
}
func TestSlot01Wave7ThemeInvalidPrefixMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*stylesheetCollector.ingestThemeBlock", "theme", "collapse_ingest_theme_block.a", "!dependencies.validPrefix(parsed.prefix)", "false")
}
