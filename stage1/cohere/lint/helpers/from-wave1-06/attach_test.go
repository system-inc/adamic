package wave06

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func attachOracle(t *testing.T) string {
	root, _ := filepath.Abs("../../../../../cohere")
	adapter, _ := filepath.Abs("testdata/attach_export.go")
	main, _ := filepath.Abs("testdata/attach_oracle.go")
	original := filepath.Join(root, "internal/lint/rules/tailwind/collapse/variant.go")
	source, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "func (registry *VariantRegistry) AttachComparison(order int, comparison VariantComparison) {"
	if strings.Count(string(source), anchor) != 1 {
		t.Fatal("Go trace drift")
	}
	observation := filepath.Join(t.TempDir(), "variant.go")
	if err := os.WriteFile(observation, []byte(strings.Replace(string(source), anchor, anchor+"\n adamicAttachTrace = append(adamicAttachTrace, order)", 1)), 0644); err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_attach.go")
	virtualMain := filepath.Join(root, "adamic_attach.go")
	data, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: adapter, virtualMain: main, original: observation}})
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtualMain)
	return binary
}
func attachSides(t *testing.T, dir, cases string) [][]byte {
	entry, _ := filepath.Abs(filepath.Join(dir, "attach_main.a"))
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	emitted := filepath.Join(t.TempDir(), "emitted.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return [][]byte{run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases), run(t, "", binary, cases)}
}
func TestAttachComparisons(t *testing.T) {
	names := capturedNames(t)
	names = append(names, "small", "large", "missing")
	goOracle := attachOracle(t)
	groups := run(t, "", goOracle, "groups")
	data, err := json.Marshal(map[string]any{"Names": names, "Groups": json.RawMessage(groups)})
	if err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(input, data, 0644); err != nil {
		t.Fatal(err)
	}
	want := run(t, "", goOracle, input)
	prepared := run(t, "", goOracle, input, "prepare")
	cases := filepath.Join(t.TempDir(), "prepared.json")
	if err := os.WriteFile(cases, prepared, 0644); err != nil {
		t.Fatal(err)
	}
	for i, got := range attachSides(t, ".", cases) {
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d mismatched Go", i)
		}
	}
	t.Logf("%d consumer-derived names plus controls, %d identical bytes on Go, source Node, emitted JavaScript and sanitized native", len(names), len(want))
	for _, m := range []struct{ name, old, new string }{
		{"wrong-direction", "const ascending = group.Ascending", "const ascending = !group.Ascending"},
		{"live-group-reference", "right, ascending)", "right, group.Ascending)"},
		{"wrong-order", "attachComparison(group.Order,", "attachComparison(group.Order + 1,"},
		{"copy-theme", "const capturedTheme = theme", "const capturedTheme = { ...theme }"},
	} {
		t.Run(m.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range []string{"attach_main.a", "attach_variant_comparisons.a"} {
				content, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				text := string(content)
				if name == "attach_variant_comparisons.a" {
					if strings.Count(text, m.old) != 1 {
						t.Fatal("anchor drift")
					}
					text = strings.Replace(text, m.old, m.new, 1)
				} else {
					options, _ := filepath.Abs("../options_json.ts")
					text = strings.Replace(text, "../options_json.ts", filepath.ToSlash(options), 1)
				}
				if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			for i, got := range attachSides(t, dir, cases) {
				if bytes.Equal(got, want) {
					t.Fatalf("mutant survived backend %d", i)
				}
				t.Logf("%s caught only by byte comparison on backend %d after clean execution", m.name, i)
			}
		})
	}
}
