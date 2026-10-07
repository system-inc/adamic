package helpers

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: bound sanitizer compilation and large corpus comparisons.
func TestSlot01Wave2FactoryMatchesCohere(t *testing.T) {
	slot01Wave2Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.NewClassLiteralReader", 12, "factory", "tailwind_new_class_literal_reader.a", "return { attributeNames, calleeNames, variablePatterns, valuesBound: false };", "return { attributeNames, calleeNames, variablePatterns: variablePatterns.slice().reverse(), valuesBound: false };")
}
func TestSlot01Wave2MemoMatchesCohere(t *testing.T) {
	slot01Wave2Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.*ClassLiteralReader.classValuesIn", 11, "memo", "tailwind_class_values_in.a", "if (cache.has(node))", "if (false)")
}
func TestSlot01Wave2LiteralMatchesCohere(t *testing.T) {
	slot01Wave2Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.classLiteralFrom", 11, "literal", "tailwind_class_literal_from.a", "pos: tokenRange.pos + 1", "pos: tokenRange.pos")
}
func TestSlot01Wave2LiteralShortRangeMutant(t *testing.T) {
	slot01Wave2Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.classLiteralFrom", 11, "literal", "tailwind_class_literal_from.a", "tokenRange.end - tokenRange.pos >= 2", "true")
}
func TestSlot01Wave2FactoryInvalidPatternMutant(t *testing.T) {
	slot01Wave2Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.NewClassLiteralReader", 12, "factory", "tailwind_new_class_literal_reader.a", "if (result.valid)", "if (true)")
}
func TestSlot01Wave2FactoryAttributeMutant(t *testing.T) {
	slot01Wave2Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.NewClassLiteralReader", 12, "factory", "tailwind_new_class_literal_reader.a", "attributeNames.set(name, true)", "attributeNames.set(name, false)")
}
func TestSlot01Wave2MemoUnboundMutant(t *testing.T) {
	slot01Wave2Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.*ClassLiteralReader.classValuesIn", 11, "memo", "tailwind_class_values_in.a", "if (!bound)", "if (false)")
}
func TestSlot01Wave2MemoEmptyMutant(t *testing.T) {
	slot01Wave2Check(t, "github.com/system-inc/cohere/internal/lint/rules/tailwind.*ClassLiteralReader.classValuesIn", 11, "memo", "tailwind_class_values_in.a", "cache.set(node, values);", "if (values.literals.length > 0 || values.templates.length > 0) { cache.set(node, values); }")
}
func slot01Wave2Check(t *testing.T, dependency string, consumers int, mode, file, old, replacement string) {
	t.Helper()
	cases, want := slot01Fixture(t, dependency, consumers, mode)
	entry, _ := filepath.Abs("slot01_wave2_main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases, mode), want)
	compare(t, run(t, "", slot01Build(t, entry), cases, mode), want)
	t.Logf("%d Go output lines matched Node and sanitized native", bytes.Count(want, []byte("\n")))
	directory := t.TempDir()
	for _, name := range []string{"slot01_wave2_main.a", "tailwind_new_class_literal_reader.a", "tailwind_default_class_literal_settings.a", "tailwind_class_values_in.a", "tailwind_class_literal_from.a", "options_json.ts"} {
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
		if name == "slot01_wave2_main.a" {
			data = []byte(strings.ReplaceAll(string(data), "./options_json.ts", "./options_json.a"))
		}
		if name == "options_json.ts" {
			name = "options_json.a"
		}
		if err := os.WriteFile(filepath.Join(directory, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutant := run(t, "", slot01Build(t, filepath.Join(directory, "slot01_wave2_main.a")), cases, mode)
	if bytes.Equal(mutant, want) {
		t.Fatal("compiled mutant survived")
	}
	slot01MutantWitness(t, mutant, want, file)
}
