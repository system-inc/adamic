package wave12

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func breakableOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../../cohere")
	main, _ := filepath.Abs("testdata/breakable_oracle.go.txt")
	exports, _ := filepath.Abs("testdata/control_flow_exports.go.txt")
	virtual := filepath.Join(root, "adamic_wave12_breakable_oracle.go")
	mapping, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: main, filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/adamic_wave12_exports.go"): exports}})
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, mapping, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	execute(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func TestBreakableStatementMatchesGo(t *testing.T) {
	oracle := breakableOracle(t)
	fixtures, _ := filepath.Abs("testdata/cfg_consumers.json")
	adapted := execute(t, "", oracle, "--adapt", fixtures)
	corpus := filepath.Join(t.TempDir(), "facts.json")
	if err := os.WriteFile(corpus, adapted, 0644); err != nil {
		t.Fatal(err)
	}
	kinds := execute(t, "", oracle, "--kinds")
	t.Logf("actual Go numeric switch/while/do/for/for-in/for-of kinds: %s", bytes.TrimSpace(kinds))
	want := execute(t, "", oracle, corpus)
	entry, _ := filepath.Abs("breakable_main.a")
	for i, got := range backends(t, entry, corpus) {
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d differs at row %d", i, difference(got, want))
		}
	}
	t.Logf("%d consumer AST and kind/nil controls, %d identical Go/source Node/emitted JS/sanitized native bytes", bytes.Count(want, []byte{'\n'}), len(want))
	for _, change := range []struct{ name, old, replacement string }{{"missing-switch", "node.kind === 256", "node.kind === 0"}, {"nil-loop", "if(!node.present) { return false; }", "if(!node.present) { return true; }"}} {
		t.Run(change.name, func(t *testing.T) {
			parent := t.TempDir()
			directory := filepath.Join(parent, "wave12")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"breakable_main.a", "control_flow_is_breakable_statement.a", "../options_json.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(file, "statement.a") {
					if strings.Count(string(data), change.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), change.old, change.replacement, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			for i, got := range backends(t, filepath.Join(directory, "breakable_main.a"), corpus) {
				if bytes.Equal(got, want) {
					t.Fatalf("backend %d mutant survived", i)
				}
				t.Logf("backend %d compiles and exits cleanly; Go comparison catches mutant at row %d", i, difference(got, want))
			}
		})
	}
}
