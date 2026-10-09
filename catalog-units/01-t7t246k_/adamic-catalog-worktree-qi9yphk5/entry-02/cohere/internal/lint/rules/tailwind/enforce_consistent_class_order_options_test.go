package tailwind

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// runClassOrderFixtureWithOptions is runClassOrderFixture with the rule's options.
func runClassOrderFixtureWithOptions(t *testing.T, source string, options EnforceConsistentClassOrderOptions) rule_testing.Result {
	t.Helper()

	packageRoot := classOrderFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine, so the rule's design system cannot be " +
			"built and these fixtures would measure a decline rather than an order")
	}

	return rule_testing.RunTypedFilesWithSetupAndOptions(t, EnforceConsistentClassOrder, map[string]string{
		"Component.tsx":                 source,
		classOrderFixtureStylesheetPath: classOrderFixtureStylesheet,
	}, "Component.tsx", options, func(root string) {
		modules := filepath.Join(root, "node_modules")
		if err := os.MkdirAll(modules, 0o755); err != nil {
			t.Fatalf("creating the fixture node_modules: %v", err)
		}
		if err := os.Symlink(packageRoot, filepath.Join(modules, "tailwindcss")); err != nil {
			t.Fatalf("linking the installed tailwindcss into the fixture: %v", err)
		}
	})
}

// TestClassOrderOptions is upstream's own cases for `order`, `unknownClassOrder` and
// `unknownClassPosition` (enforce-consistent-class-order.test.ts at 4.7.0), each written and wanted
// exactly as upstream's jsx case is, plus the default each one changes, so every option is shown
// changing the verdict.
func TestClassOrderOptions(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		source  string
		options EnforceConsistentClassOrderOptions
		want    string
	}{
		{name: "asc", source: "b a", options: EnforceConsistentClassOrderOptions{Order: "asc"}, want: "a b"},
		{name: "desc", source: "a b", options: EnforceConsistentClassOrderOptions{Order: "desc"}, want: "b a"},
		{name: "official", source: "w-full absolute", options: EnforceConsistentClassOrderOptions{Order: "official"}, want: "absolute w-full"},
		// The engine's order puts px before py too, so asc is told apart by the pair desc reverses.
		{name: "asc is not locale order", source: "py-24 px-12", options: EnforceConsistentClassOrderOptions{Order: "asc"}, want: "px-12 py-24"},
		{name: "desc is not locale order", source: "px-12 py-24", options: EnforceConsistentClassOrderOptions{Order: "desc"}, want: "py-24 px-12"},
		{name: "desc ignores the engine", source: "absolute w-full", options: EnforceConsistentClassOrderOptions{Order: "desc"}, want: "w-full absolute"},
		{name: "unknown classes first by default", source: "flex unknown", want: "unknown flex"},
		{
			name:    "unknown classes sorted ascending at the start",
			source:  "flex cy-24 cx-12",
			options: EnforceConsistentClassOrderOptions{Order: "official", UnknownClassOrder: "asc", UnknownClassPosition: "start"},
			want:    "cx-12 cy-24 flex",
		},
		{
			name:    "unknown classes sorted descending at the end",
			source:  "flex cx-12 cy-24",
			options: EnforceConsistentClassOrderOptions{Order: "official", UnknownClassOrder: "desc", UnknownClassPosition: "end"},
			want:    "flex cy-24 cx-12",
		},
		// The defaults keep unknown classes in source order, so the same list stands as written.
		{name: "unknown classes keep their order by default", source: "cy-24 cx-12 flex", want: ""},
		{
			name:    "unknown classes at the end keep their order",
			source:  "cy-24 flex cx-12",
			options: EnforceConsistentClassOrderOptions{UnknownClassPosition: "end"},
			want:    "flex cy-24 cx-12",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := `export const element = <img className="` + testCase.source + `" />;`
			result := runClassOrderFixtureWithOptions(t, source, testCase.options)
			if testCase.want == "" {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFixedSource(t, result, `export const element = <img className="`+testCase.want+`" />;`+"\n")
		})
	}
}
