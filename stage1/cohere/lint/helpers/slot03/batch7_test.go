package slot03

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func batch7Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch7/testdata", []string{"/collapse.NewVariantRegistry", "/collapse.*VariantRegistry.Register", "/collapse.*VariantRegistry.AttachComparison"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch7/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch7.go")
	overlay := filepath.Join(directory, "overlay.json")

	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot03_batch7.go"): filepath.Join(here, "collapse_export.go")}})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "oracle")
	command(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	cases := filepath.Join(directory, "cases.json")
	want := command(t, "", binary, filepath.Join(here, "sources.jsonl.gz"), cases)
	return cases, want
}

// Not parallel: native sanitizer builds and large fixture observations bound memory.
func TestBatch7Helpers(t *testing.T) {
	cases, want := batch7Oracle(t)
	entry, _ := filepath.Abs("batch7/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch7JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch7Mutants(t *testing.T) {
	cases, want := batch7Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{{"new_variant_registry.a", "lastOrder: 0", "lastOrder: 1", ""}, {"new_variant_registry.a", "return {registrations: new Map<string, Registration>(), comparisons: new Map<number, Comparison>(), lastOrder: 0, groupPresent: false, groupOrder: 0};", "const existing = cache.get(0); if(existing !== undefined) { return existing; } const created: Registry = {registrations: new Map<string, Registration>(), comparisons: new Map<number, Comparison>(), lastOrder: 0, groupPresent: false, groupOrder: 0}; cache.set(0, created); return created;", "const cache = new Map<number, Registry>();\n"}, {"register.a", "order: existing.order", "order: state.lastOrder + 1", ""}, {"register.a", "state.groupPresent ? state.groupOrder : state.lastOrder + 1", "state.lastOrder + 1", ""}, {"register.a", "if(!state.groupPresent)", "if(true)", ""}, {"attach_comparison.a", "if(comparison === undefined) { return; }", "if(comparison === undefined) { state.comparisons.delete(order); return; }", ""}, {"attach_comparison.a", "state.comparisons.set(order, comparison);", "if(!state.comparisons.has(order)) { state.comparisons.set(order, comparison); }", ""}, {"attach_comparison.a", "state.comparisons.set(order, comparison);", "state.comparisons.set(order + 1, comparison);", ""}} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"new_variant_registry.a", "register.a", "attach_comparison.a", "registry.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch7", file))
				if err != nil {
					t.Fatal(err)
				}
				text := string(data)
				if file == mutant.file {
					if strings.Count(text, mutant.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					text = mutant.prefix + strings.Replace(text, mutant.old, mutant.new, 1)
				}
				if file == "main.a" {
					reader, _ := filepath.Abs("../options_json.ts")
					text = strings.ReplaceAll(text, "../../options_json.ts", filepath.ToSlash(reader))
				}
				if err = os.WriteFile(filepath.Join(scratch, file), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := command(t, "", build(t, filepath.Join(scratch, "main.a")), cases)
			if bytes.Equal(got, want) {
				t.Fatal("semantic mutant survived")
			}
			a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
			for i := 0; i < len(a) && i < len(b); i++ {
				if a[i] != b[i] {
					t.Logf("compiled semantic mutant caught at line %d: got %q Go %q", i+1, a[i], b[i])
					return
				}
			}
			t.Fatal("no changed output line")
		})
	}
}

func batch7JavaScript(t *testing.T, entry string) string {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "helper.mjs")
	if err = os.WriteFile(path, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
