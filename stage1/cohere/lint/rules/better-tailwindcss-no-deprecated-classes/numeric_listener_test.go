package cores

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	syntax "github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Not parallel: sequential sanitizer builds bound memory across declaration mutants.
func TestNumericListeners(t *testing.T) {
	slugs := []string{"typescript-consistent-type-assertions", "structure-tailwind-no-physical-direction", "eslint-comments-require-description", "next-google-font-display", "better-tailwindcss-no-deprecated-classes", "better-tailwindcss-no-duplicate-classes", "better-tailwindcss-no-unknown-classes"}
	kinds := map[string]int{}
	for kind := syntax.Kind(0); kind < syntax.KindCount; kind++ {
		kinds[strings.TrimPrefix(kind.String(), "Kind")] = int(kind)
	}
	var imports, output, expected strings.Builder
	declarations := make([]string, len(slugs))
	for index, slug := range slugs {
		directory := filepath.Join(root(t), "stage1/cohere/lint/rules", slug)
		raw, err := os.ReadFile(filepath.Join(directory, "rule.json"))
		if err != nil {
			t.Fatal(err)
		}
		var descriptor struct{ Kinds []string }
		if err := json.Unmarshal(raw, &descriptor); err != nil {
			t.Fatal(err)
		}
		numbers := make([]string, len(descriptor.Kinds))
		for i, name := range descriptor.Kinds {
			value, ok := kinds[name]
			if !ok {
				t.Fatalf("unknown pinned parser kind %q", name)
			}
			numbers[i] = strconv.Itoa(value)
		}
		fmt.Fprintf(&expected, "%s\t%s\n", slug, strings.Join(numbers, ","))
		source, err := os.ReadFile(filepath.Join(directory, "numeric_listener.a"))
		if err != nil {
			t.Fatal(err)
		}
		declarations[index] = file(t, fmt.Sprintf("listener%d.a", index), source)
		fmt.Fprintf(&imports, "import { syntaxKinds as kinds%d } from '%s';\n", index, filepath.ToSlash(declarations[index]))
		fmt.Fprintf(&output, "console.log('%s\\t' + kinds%d.join(','));\n", slug, index)
	}
	source := imports.String() + output.String()
	entry := file(t, "numeric-main.a", []byte(source))
	binary, emitted := compile(t, entry)
	want := []byte(expected.String())
	observe := func(entry, binary, emitted string) [][]byte {
		runner := filepath.Join(root(t), "oracle/node.mjs")
		return [][]byte{
			command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry),
			command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted),
			command(t, "", binary),
		}
	}
	for side, got := range observe(entry, binary, emitted) {
		equal(t, got, want)
		t.Logf("numeric SyntaxKind declarations: side %d, %d Go-identical bytes", side, len(got))
	}
	for index, slug := range slugs {
		original, err := os.ReadFile(declarations[index])
		if err != nil {
			t.Fatal(err)
		}
		changed := strings.Replace(string(original), " = [", " = [0, ", 1)
		mutant := file(t, "numeric-mutant.a", []byte(changed))
		mutatedSource := strings.Replace(source, filepath.ToSlash(declarations[index]), filepath.ToSlash(mutant), 1)
		entry := file(t, "mutant-main.a", []byte(mutatedSource))
		binary, emitted := compile(t, entry)
		for side, got := range observe(entry, binary, emitted) {
			if bytes.Equal(got, want) {
				t.Fatalf("%s subscription mutant survived side %d", slug, side)
			}
			t.Logf("%s numeric subscription mutant compiles/runs; comparison alone catches side %d", slug, side)
		}
	}
}
