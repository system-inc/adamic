package helpers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func slot05Build(t *testing.T, dir string) string {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(dir, "main.a")})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "slot05")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}
func TestSlot05Identifier(t *testing.T) {
	slot05Verify(t, "github.com/system-inc/cohere/internal/lint/ecmascript/react.isIdentifierNamed", "react_identifier_named.a", "if(node.kind !== 'ParenthesizedExpression')", "if(true)")
}
func TestSlot05Imports(t *testing.T) {
	slot05Verify(t, "github.com/system-inc/cohere/internal/lint/ecmascript/imports.BindingsOf", "import_bindings.a", "namespace: clause.namedBindings", "namespace: binding.name")
}
func TestSlot05ComponentBase(t *testing.T) {
	slot05Verify(t, "github.com/system-inc/cohere/internal/lint/ecmascript/react.isComponentBase", "react_component_base.a", "if(!isIdentifierNamed(nodes, node.expression, 'React'))", "if(false)")
}
func slot05Verify(t *testing.T, symbol, target, old, replacement string) {
	// Not parallel: the bounded corpus and compiling mutants intentionally run in order.
	root, _ := filepath.Abs("../../../..")
	cohere := filepath.Join(root, "cohere")
	const pin = "715ba94f3608a6500086b1076ce5cb7e51b836db"
	if got := strings.TrimSpace(string(run(t, cohere, "git", "rev-parse", "HEAD"))); got != pin {
		t.Fatalf("Go cohere pin drift: %s", got)
	}
	virtual := filepath.Join(cohere, "adamic_slot05_oracle.go")
	side, _ := filepath.Abs("slot05/testdata/oracle.go")
	export, _ := filepath.Abs("slot05/testdata/react_export.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side, filepath.Join(cohere, "internal/lint/ecmascript/react/adamic_slot05.go"): export}})
	scratch := t.TempDir()
	overlayPath := filepath.Join(scratch, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(scratch, "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	t.Log(strings.TrimSpace(string(run(t, "", oracle, root, scratch, symbol))))
	cases := filepath.Join(scratch, "cases.json")
	coverage, err := os.ReadFile(filepath.Join(scratch, "coverage.json"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("consumer test-file literal counts: %s", coverage)
	caseData, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	if evidence := os.Getenv("ADAMIC_SLOT05_EVIDENCE"); evidence != "" {
		stem := strings.TrimSuffix(target, ".a")
		if err := os.WriteFile(filepath.Join(evidence, stem+"-coverage.json"), coverage, 0644); err != nil {
			t.Fatal(err)
		}
		hash := fmt.Sprintf("%x  cases.json\n", sha256.Sum256(caseData))
		if err := os.WriteFile(filepath.Join(evidence, stem+"-corpus.sha256"), []byte(hash), 0644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(filepath.Join(scratch, "want.txt"))
	if err != nil {
		t.Fatal(err)
	}
	dir, _ := filepath.Abs("slot05")
	runner := filepath.Join(root, "oracle/node.mjs")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(dir, "main.a"), cases), want)
	compare(t, run(t, "", slot05Build(t, dir), cases), want)
	mutant := t.TempDir()
	for _, name := range []string{"nodes.a", "react_identifier_named.a", "import_bindings.a", "react_component_base.a", "main.a"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if name == target {
			if strings.Count(string(data), old) != 1 {
				t.Fatal("mutant anchor drift")
			}
			data = []byte(strings.Replace(string(data), old, replacement, 1))
		}
		if name == "main.a" {
			data = []byte(strings.Replace(string(data), "'../options_json.ts'", strconvQuote(filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts")), 1))
		}
		if err = os.WriteFile(filepath.Join(mutant, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	got := run(t, "", slot05Build(t, mutant), cases)
	if bytes.Equal(got, want) {
		t.Fatal("compiled semantic mutant survived")
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := range a {
		if i < len(b) && a[i] != b[i] {
			t.Logf("%s mutant caught at verdict %d: mutant %s, Go %s", target, i+1, a[i], b[i])
			break
		}
	}
}
func strconvQuote(path string) string { data, _ := json.Marshal(path); return string(data) }
