package helpers

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	if mode == "dissect" {
		// Its defining rule is outside the frozen blocked cohort; exercise its direct fixture too.
		extra := filepath.Join(root, "internal/lint/rules/tailwind/no_deprecated_classes_test.go")
		if _, err := os.Stat(extra); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, extra)
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
	if mode == "table" {
		livePath := filepath.Join(root, "internal/lint/rules/tailwind/collapse/descriptor_live.go")
		data, err := os.ReadFile(livePath)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, anchor := range []string{"table.addThemeNamespaces(system.theme)", "table.addRepositoryStatics(system)", "table.addRepositoryFunctionalRoots(system)"} {
			if strings.Count(text, anchor) != 1 {
				t.Fatal("table trace anchor drift")
			}
			text = strings.Replace(text, anchor, "wave7TableTrace = append(wave7TableTrace, "+strconv.Quote(anchor)+")\n    "+anchor, 1)
		}
		anchor := "FrameworkStaticReading(name)"
		if strings.Count(text, anchor) != 1 {
			t.Fatal("static factory anchor drift")
		}
		text = strings.Replace(text, anchor, "wave7StaticReading(name)", 1)
		traced := filepath.Join(directory, "descriptor_live.go")
		if err := os.WriteFile(traced, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		overlay, _ = json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot01_wave7.go"): bridge, livePath: traced}})
	}
	var configuration struct{ Replace map[string]string }
	if err := json.Unmarshal(overlay, &configuration); err != nil {
		t.Fatal(err)
	}
	bridgeTailwind, _ := filepath.Abs("testdata/slot01_wave7_tailwind.go")
	configuration.Replace[filepath.Join(root, "internal/lint/rules/tailwind/adamic_slot01_wave7.go")] = bridgeTailwind
	overlay, _ = json.Marshal(configuration)
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
	count := 6
	if mode == "dissect" {
		count = 5
	}
	cases, want := slot01Wave7Fixture(t, dependency, count, mode)
	entry, _ := filepath.Abs("slot01_wave7_main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases, mode), want)
	compare(t, run(t, "", slot01Build(t, entry), cases, mode), want)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, slot01Wave3JavaScript(t, entry), cases, mode), want)
	t.Logf("%d Go output lines matched source Node, sanitized native and emitted JavaScript", bytes.Count(want, []byte("\n")))
	dir := t.TempDir()
	for _, name := range []string{"slot01_wave7_main.a", "tailwind_dissect_class.a", "collapse_ingest_theme_block.a", "collapse_new_table.a", "options_json.ts"} {
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

// Not parallel: Go global table mutation probes are restored after each observation.
func TestSlot01Wave7TableMatchesCohere(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.NewTable", "table", "collapse_new_table.a", "if (!system.bound) { return table; }", "")
}
func TestSlot01Wave7TableReadingCopyMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.NewTable", "table", "collapse_new_table.a", "{ order: { values: reading.order.values, length: reading.order.length, capacity: reading.order.capacity }, count: reading.count }", "reading")
}
func TestSlot01Wave7TablePropertyAliasMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.NewTable", "table", "collapse_new_table.a", "system.bound ? dependencies.propertyOrder : new Map<string, number>()", "system.bound ? new Map<string, number>(dependencies.propertyOrder) : new Map<string, number>()")
}
func TestSlot01Wave7TableOrderMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.NewTable", "table", "collapse_new_table.a", "dependencies.addThemeNamespaces(table);\n    dependencies.addRepositoryStatics(table);", "dependencies.addRepositoryStatics(table);\n    dependencies.addThemeNamespaces(table);")
}

func TestSlot01Wave7TableDescriptorMapMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.NewTable", "table", "collapse_new_table.a", "descriptors: new Map<string, number>()", "descriptors: system.bound ? dependencies.baseDescriptors : new Map<string, number>()")
}
func TestSlot01Wave7TableFrameworkMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.NewTable", "table", "collapse_new_table.a", "for (const name of dependencies.frameworkNames)", "for (const name of ([] as readonly string[]))")
}

func TestSlot01Wave7TableHeaderCopyMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.NewTable", "table", "collapse_new_table.a", "order: { values: reading.order.values, length: reading.order.length, capacity: reading.order.capacity }", "order: reading.order")
}
func TestSlot01Wave7TableBackingShareMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.NewTable", "table", "collapse_new_table.a", "values: reading.order.values", "values: reading.order.values.slice()")
}

// Not parallel: bounded consumer captures and sanitizer builds.
func TestSlot01Wave7DissectMatchesCohere(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.dissectClass", "dissect", "tailwind_dissect_class.a", "base.lastIndexOf(':')", "base.indexOf(':')")
}
func TestSlot01Wave7DissectSuffixMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.dissectClass", "dissect", "tailwind_dissect_class.a", "base.endsWith('!')", "base.startsWith('!')")
}
func TestSlot01Wave7DissectColonMutant(t *testing.T) {
	slot01Wave7Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.dissectClass", "dissect", "tailwind_dissect_class.a", "variants = base.slice(0, colon + 1)", "variants = base.slice(0, colon)")
}
