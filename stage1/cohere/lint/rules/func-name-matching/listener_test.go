package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	syntax "github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var numericListenerSlugs = []string{"consistent-this", "func-name-matching", "eslint-comments-require-description", "next-google-font-display", "tailwind-no-physical-direction", "typescript-no-non-null-asserted-optional-chain", "typescript-no-non-null-assertion", "typescript-no-this-alias", "tailwind-important-position", "tailwind-variable-syntax", "tailwind-variant-order"}

func numericListenerOracle(t *testing.T) []byte {
	t.Helper()
	ids := map[string]int{}
	for kind := syntax.Kind(0); kind < syntax.KindCount; kind++ {
		ids[strings.TrimPrefix(kind.String(), "Kind")] = int(kind)
	}
	var output strings.Builder
	for _, slug := range numericListenerSlugs {
		data, err := os.ReadFile(filepath.Join(repo(t), "stage1/cohere/lint/rules", slug, "rule.json"))
		if err != nil {
			t.Fatal(err)
		}
		var descriptor struct{ Kinds []string }
		if err = json.Unmarshal(data, &descriptor); err != nil {
			t.Fatal(err)
		}
		var values []string
		for _, kind := range descriptor.Kinds {
			id, ok := ids[kind]
			if !ok {
				t.Fatal("unknown pinned parser kind", kind)
			}
			values = append(values, strconv.Itoa(id))
		}
		fmt.Fprintf(&output, "%s\t%s\n", slug, strings.Join(values, ","))
	}
	return []byte(output.String())
}
func numericListenerBuild(t *testing.T, mutant bool) (string, string, string) {
	t.Helper()
	folder := t.TempDir()
	var imports, body strings.Builder
	for i, slug := range numericListenerSlugs {
		source := filepath.Join(repo(t), "stage1/cohere/lint/rules", slug, "listener.a")
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		if mutant && slug == "consistent-this" {
			from := fmt.Sprintf("[%d]", syntax.KindSourceFile)
			to := fmt.Sprintf("[%d]", syntax.KindVariableDeclaration)
			if bytes.Count(data, []byte(from)) != 1 {
				t.Fatal("numeric mutant anchor changed")
			}
			data = bytes.Replace(data, []byte(from), []byte(to), 1)
		}
		name := fmt.Sprintf("listener%d.a", i)
		if err = os.WriteFile(filepath.Join(folder, name), data, 0644); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&imports, "import { listenerSyntaxKinds as kinds%d } from './%s';\n", i, name)
		fmt.Fprintf(&body, "console.log('%s\\t'+kinds%d.join(','));\n", slug, i)
	}
	entry := filepath.Join(folder, "main.a")
	if err := os.WriteFile(entry, []byte(imports.String()+body.String()), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(folder, "listeners")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	js := filepath.Join(folder, "listeners.mjs")
	if err = os.WriteFile(js, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return entry, js, binary
}
func numericListenerObservations(t *testing.T, mutant bool) map[string][]byte {
	t.Helper()
	entry, js, binary := numericListenerBuild(t, mutant)
	runner := filepath.Join(repo(t), "oracle/node.mjs")
	return map[string][]byte{
		"source Node":        clean(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry)),
		"emitted JavaScript": clean(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, js)),
		"ASan/UBSan native":  clean(t, run(t, "", binary)),
	}
}

// Not parallel: one owned declaration corpus shares bounded native compiler resources.
func TestNumericListenerDeclarations(t *testing.T) {
	want := numericListenerOracle(t)
	for side, got := range numericListenerObservations(t, false) {
		if !bytes.Equal(got, want) {
			t.Fatalf("%s numeric subscriptions differ from pinned Go: %s", side, difference(got, want))
		}
	}
	t.Logf("%d owned declarations, %d bytes match pinned Go numeric SyntaxKinds on all three execution paths", len(numericListenerSlugs), len(want))
}

// Not parallel: the wrong subscription must compile and finish without sanitizer diagnostics.
func TestNumericListenerDeclarationMutant(t *testing.T) {
	want := numericListenerOracle(t)
	for side, got := range numericListenerObservations(t, true) {
		if bytes.Equal(got, want) {
			t.Fatal("wrong numeric kind survived", side)
		}
		t.Logf("SourceFile -> VariableDeclaration compiled and finished cleanly; only Go subscription comparison caught it on %s: %s", side, difference(got, want))
	}
}
