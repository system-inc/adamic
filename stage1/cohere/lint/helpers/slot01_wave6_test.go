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

func slot01Wave6Fixture(t *testing.T, dependency string, consumers int, mode string) (string, []byte) {
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
	source, _ := filepath.Abs("testdata/slot01_wave6_oracle.go")
	virtual := filepath.Join(root, "adamic_slot01_wave6.go")
	bridge, _ := filepath.Abs("testdata/slot01_wave6_collapse.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot01_wave6.go"): bridge}})
	utilitySource := filepath.Join(root, "internal/lint/rules/tailwind/collapse/utility_nodes.go")
	utilityData, err := os.ReadFile(utilitySource)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "valueAst := ParseValue(node.Value)"
	if strings.Count(string(utilityData), anchor) != 1 {
		t.Fatal("Go trace anchor drift")
	}
	tracedPath := filepath.Join(directory, "utility_nodes.go")
	if err := os.WriteFile(tracedPath, []byte(strings.Replace(string(utilityData), anchor, "wave6Trace = append(wave6Trace, node.Value)\n\t\t"+anchor, 1)), 0644); err != nil {
		t.Fatal(err)
	}
	overlay, _ = json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot01_wave6.go"): bridge, utilitySource: tracedPath}})
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
func slot01Wave6Check(t *testing.T, dependency, mode, file, old, replacement string) {
	t.Helper()
	cases, want := slot01Wave6Fixture(t, dependency, 6, mode)
	entry, _ := filepath.Abs("slot01_wave6_main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases, mode), want)
	compare(t, run(t, "", slot01Build(t, entry), cases, mode), want)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, slot01Wave3JavaScript(t, entry), cases, mode), want)
	t.Logf("%d Go output lines matched source Node, native and emitted JavaScript", bytes.Count(want, []byte("\n")))
	dir := t.TempDir()
	for _, name := range []string{"slot01_wave6_main.a", "collapse_normalize_utility_definition.a", "collapse_normalize_value_function_arguments.a", "collapse_register_framework_variants.a", "options_json.ts"} {
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
		if name == "slot01_wave6_main.a" {
			data = []byte(strings.ReplaceAll(string(data), "./options_json.ts", "./options_json.a"))
		}
		if name == "options_json.ts" {
			name = "options_json.a"
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutant := run(t, "", slot01Build(t, filepath.Join(dir, "slot01_wave6_main.a")), cases, mode)
	if bytes.Equal(mutant, want) {
		t.Fatal("compiled mutant survived")
	}
	slot01MutantWitness(t, mutant, want, file)
}

// Not parallel: bounded external captures and sanitizer builds.
func TestSlot01Wave6DefinitionMatchesCohere(t *testing.T) {
	slot01Wave6Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.normalizeUtilityDefinition", "definition", "collapse_normalize_utility_definition.a", "if (!definition.bound) { return; }", "")
}
func TestSlot01Wave6DefinitionForwardMutant(t *testing.T) {
	slot01Wave6Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.normalizeUtilityDefinition", "definition", "collapse_normalize_utility_definition.a", "normalizeArguments(definition.nodes);", "")
}

// Not parallel: bounded external captures and sanitizer compilation.
func TestSlot01Wave6WalkerMatchesCohere(t *testing.T) {
	slot01Wave6Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.normalizeValueFunctionArguments", "walker", "collapse_normalize_value_function_arguments.a", "collapseNormalizeValueFunctionArguments(nodes, node.children, rewrite);", "")
}
func TestSlot01Wave6WalkerPresentMutant(t *testing.T) {
	slot01Wave6Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.normalizeValueFunctionArguments", "walker", "collapse_normalize_value_function_arguments.a", "!node.valuePresent", "false")
}
func TestSlot01Wave6WalkerKindMutant(t *testing.T) {
	slot01Wave6Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.normalizeValueFunctionArguments", "walker", "collapse_normalize_value_function_arguments.a", "node.kind !== 'declaration'", "false")
}
func TestSlot01Wave6WalkerBailMutant(t *testing.T) {
	slot01Wave6Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.normalizeValueFunctionArguments", "walker", "collapse_normalize_value_function_arguments.a", "if (!node.value.includes('--value(') && !node.value.includes('--modifier(')) { continue; }", "")
}

// Not parallel: bounded sanitizer builds and external Go observations.
func TestSlot01Wave6WalkerOrderMutant(t *testing.T) {
	slot01Wave6Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.normalizeValueFunctionArguments", "walker", "collapse_normalize_value_function_arguments.a", `collapseNormalizeValueFunctionArguments(nodes, node.children, rewrite);
        if (node.kind !== 'declaration' || !node.valuePresent || node.value === '') { continue; }
        if (!node.value.includes('--value(') && !node.value.includes('--modifier(')) { continue; }
        node.value = rewrite(node.value);`, `if (node.kind === 'declaration' && node.valuePresent && node.value !== '' && (node.value.includes('--value(') || node.value.includes('--modifier('))) { node.value = rewrite(node.value); }
        collapseNormalizeValueFunctionArguments(nodes, node.children, rewrite);`)
}

// Not parallel: bounded external Go captures and sanitizer compilation.
func TestSlot01Wave6FrameworkMatchesCohere(t *testing.T) {
	slot01Wave6Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*VariantRegistry.RegisterFrameworkVariants", "framework", "collapse_register_framework_variants.a", "order: existing.order", "order: registration.order")
}
func TestSlot01Wave6FrameworkLastOrderMutant(t *testing.T) {
	slot01Wave6Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*VariantRegistry.RegisterFrameworkVariants", "framework", "collapse_register_framework_variants.a", "if (registration.order > registry.lastOrder) { registry.lastOrder = registration.order; }", "")
}
func TestSlot01Wave6FrameworkCopyMutant(t *testing.T) {
	slot01Wave6Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*VariantRegistry.RegisterFrameworkVariants", "framework", "collapse_register_framework_variants.a", "{ name: registration.name, order: registration.order, kind: registration.kind }", "registration")
}

// Not parallel: bounded source/native/emitted-JavaScript refusal builds.
func TestSlot01Wave6FrameworkDomainRefusal(t *testing.T) {
	entry, _ := filepath.Abs("slot01_wave6_main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	dir := t.TempDir()
	cases := filepath.Join(dir, "unsafe.jsonl")
	if err := os.WriteFile(cases, []byte(`[3,[],[["unsafe",9007199254740993,"static"]],["unsafe"]]`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	observe := func(command string, args ...string) {
		t.Helper()
		cmd := exec.Command(command, args...)
		log, err := os.CreateTemp(t.TempDir(), "refusal-")
		if err != nil {
			t.Fatal(err)
		}
		defer log.Close()
		cmd.Stdout = log
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err == nil {
			t.Fatal("unsafe order accepted")
		}
		if !strings.Contains(stderr.String(), "NotYet: framework variant order outside exact integer range") {
			t.Fatalf("wrong refusal: %s", &stderr)
		}
	}
	observe("node", "--disable-warning=ExperimentalWarning", runner, entry, cases, "framework")
	observe(slot01Build(t, entry), cases, "framework")
	observe("node", "--disable-warning=ExperimentalWarning", runner, slot01Wave3JavaScript(t, entry), cases, "framework")
	for _, name := range []string{"slot01_wave6_main.a", "collapse_normalize_utility_definition.a", "collapse_normalize_value_function_arguments.a", "collapse_register_framework_variants.a", "options_json.ts"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if name == "collapse_register_framework_variants.a" {
			old := "if (!valid) { panic('NotYet: framework variant order outside exact integer range'); }"
			if strings.Count(string(data), old) != 1 {
				t.Fatal("domain mutant anchor drift")
			}
			data = []byte(strings.Replace(string(data), old, "", 1))
		}
		if name == "slot01_wave6_main.a" {
			data = []byte(strings.ReplaceAll(string(data), "./options_json.ts", "./options_json.a"))
		}
		if name == "options_json.ts" {
			name = "options_json.a"
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutant := run(t, "", slot01Build(t, filepath.Join(dir, "slot01_wave6_main.a")), cases, "framework")
	if !strings.HasPrefix(string(mutant), "9007199254740992 1\n") {
		t.Fatalf("unexpected mutant output: %s", mutant)
	}
	t.Log("compiled exact-integer guard mutant accepted 9007199254740993 as rounded 9007199254740992; source/native/emitted-JS refusal checks catch it")
}
