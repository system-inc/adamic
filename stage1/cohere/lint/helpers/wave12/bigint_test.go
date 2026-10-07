package wave12

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func controlFlowOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../../cohere")
	main, _ := filepath.Abs("testdata/bigint_oracle.go.txt")
	exports, _ := filepath.Abs("testdata/control_flow_exports.go.txt")
	virtual := filepath.Join(root, "adamic_wave12_control_flow_oracle.go")
	mapping, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: main, filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/adamic_wave12_exports.go"): exports}})
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, mapping, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	execute(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func TestBigIntNormalizationMatchesGo(t *testing.T) {
	corpus, _ := filepath.Abs("testdata/bigint_cases.json")
	entry, _ := filepath.Abs("bigint_main.a")
	want := execute(t, "", controlFlowOracle(t), corpus)
	for i, got := range backends(t, entry, corpus) {
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d differs at row %d", i, difference(got, want))
		}
	}
	t.Logf("%d cases and %d UTF-16 observation bytes match actual Go, source Node, ASan/UBSan native and emitted JavaScript", bytes.Count(want, []byte{'\n'}), len(want))
}
func TestBigIntNormalizationMutants(t *testing.T) {
	corpus, _ := filepath.Abs("testdata/bigint_cases.json")
	want := execute(t, "", controlFlowOracle(t), corpus)
	for _, change := range []struct{ name, old, replacement string }{
		{"hex-digit-value", "value = character - 97 + 10;", "value = character - 97 + 11;"},
		{"negative-zero", "negative && result !== '0'", "negative"},
		{"invalid-fallback", "if(!consumedDigit || !previousDigit) { return digits; }", "if(!consumedDigit || !previousDigit) { return ''; }"},
	} {
		t.Run(change.name, func(t *testing.T) {
			parent := t.TempDir()
			directory := filepath.Join(parent, "wave12")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"bigint_main.a", "control_flow_normalize_bigint_literal.a", "../options_json.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(file, "literal.a") {
					if strings.Count(string(data), change.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), change.old, change.replacement, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			for i, got := range backends(t, filepath.Join(directory, "bigint_main.a"), corpus) {
				if bytes.Equal(got, want) {
					t.Fatalf("backend %d mutant survived", i)
				}
				t.Logf("backend %d compiles and runs cleanly; Go comparison catches mutant at row %d", i, difference(got, want))
			}
		})
	}
}
