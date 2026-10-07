package wave12

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func truthyOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../../cohere")
	main, _ := filepath.Abs("testdata/truthy_oracle.go.txt")
	exports, _ := filepath.Abs("testdata/truthy_exports.go.txt")
	virtual := filepath.Join(root, "adamic_wave12_truthy_oracle.go")
	mapping, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: main, filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/adamic_wave12_exports.go"): exports}})
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, mapping, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	execute(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func TestAlwaysTruthyTestMatchesGo(t *testing.T) {
	oracle := truthyOracle(t)
	original, err := os.ReadFile("testdata/cfg_consumers.json")
	if err != nil {
		t.Fatal(err)
	}
	type fixture struct{ File, Source, Rule string }
	var inputs []fixture
	if err := json.Unmarshal(original, &inputs); err != nil {
		t.Fatal(err)
	}
	for _, expression := range []string{"true", "false", "null", "undefined", "NaN", "Infinity", "0", "0x0", "0.0", "0e10", "0b0", "0o0", "0_0", "00", ".0", "1", "0x1", "1e2", "1_0", "1e400", "0n", "0x0n", "0b0n", "0o0n", "1n", "0x10000000000000000n", "999999999999999999999999999999999999999n", "''", "'x'", "'\\0'", "'😀'", "`x`", "/(?:)/", "/false/", "-1", "!false", "[]", "{}", "()=>true", "new Boolean(false)"} {
		for _, pair := range [][2]string{{"", ""}, {"(", ")"}, {"(((", ")))"}} {
			inputs = append(inputs, fixture{File: "controls.ts", Rule: "boundary-controls", Source: "while (" + pair[0] + expression + pair[1] + ") { break; }"})
		}
	}
	fixtureData, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	fixtures := filepath.Join(t.TempDir(), "fixtures.json")
	if err := os.WriteFile(fixtures, fixtureData, 0644); err != nil {
		t.Fatal(err)
	}
	adapted := execute(t, "", oracle, "--adapt", fixtures)
	corpus := filepath.Join(t.TempDir(), "facts.json")
	if err := os.WriteFile(corpus, adapted, 0644); err != nil {
		t.Fatal(err)
	}
	kinds := execute(t, "", oracle, "--kinds")
	t.Logf("actual Go numeric parenthesized/true/regex/numeric/bigint/string kinds: %s", bytes.TrimSpace(kinds))
	want := execute(t, "", oracle, "--want", fixtures)
	entry, _ := filepath.Abs("truthy_main.a")
	for i, got := range backends(t, entry, corpus) {
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d differs at row %d", i, difference(got, want))
		}
	}
	t.Logf("%d consumer AST and literal/parenthesis controls, %d identical Go/source Node/emitted JS/sanitized native bytes", bytes.Count(want, []byte{'\n'}), len(want))
	for _, change := range []struct{ name, old, replacement string }{{"zero-bigint", "normalizeBigIntLiteral(node.text) !== '0'", "node.text !== '0'"}, {"skip-parentheses", "node.kind === 218", "node.kind === 0"}, {"empty-string", "node.text !== ''", "node.text === ''"}} {
		t.Run(change.name, func(t *testing.T) {
			parent := t.TempDir()
			directory := filepath.Join(parent, "wave12")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"truthy_main.a", "control_flow_is_always_truthy_test.a", "control_flow_normalize_bigint_literal.a", "../options_json.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(file, "truthy_test.a") {
					if strings.Count(string(data), change.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), change.old, change.replacement, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			for i, got := range backends(t, filepath.Join(directory, "truthy_main.a"), corpus) {
				if bytes.Equal(got, want) {
					t.Fatalf("backend %d mutant survived", i)
				}
				t.Logf("backend %d compiles and exits cleanly; Go comparison catches mutant at row %d", i, difference(got, want))
			}
		})
	}
}

func TestAlwaysTruthyNilRefused(t *testing.T) {
	oracle := truthyOracle(t)
	corpus := filepath.Join(t.TempDir(), "nil.json")
	if err := os.WriteFile(corpus, []byte(`[{"present":false,"kind":0,"text":"","inner":[]}]`), 0644); err != nil {
		t.Fatal(err)
	}
	entry, _ := filepath.Abs("truthy_main.a")
	commands := append([][]string{{oracle, "--nil"}}, backendCommands(t, entry, corpus)...)
	for i, command := range commands {
		process := exec.Command(command[0], command[1:]...)
		var output, errors bytes.Buffer
		process.Stdout = &output
		process.Stderr = &errors
		if err := process.Run(); err == nil || output.Len() != 0 || errors.Len() == 0 {
			t.Fatalf("backend %d failed to refuse nil: stdout %q stderr %q", i, output.String(), errors.String())
		}
		t.Logf("backend %d rejects nil before any successful observation", i)
	}
}
