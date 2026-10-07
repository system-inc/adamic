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

func labelsOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../../cohere")
	main, _ := filepath.Abs("testdata/labels_oracle.go.txt")
	exports, _ := filepath.Abs("testdata/labels_exports.go.txt")
	virtual := filepath.Join(root, "adamic_wave12_labels_oracle.go")
	mapping, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: main, filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/adamic_wave12_exports.go"): exports}})
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, mapping, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	execute(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func TestLabelsOfMatchesGo(t *testing.T) {
	oracle := labelsOracle(t)
	original, err := os.ReadFile("testdata/cfg_consumers.json")
	if err != nil {
		t.Fatal(err)
	}
	type fixture struct{ File, Source, Rule string }
	var inputs []fixture
	if err := json.Unmarshal(original, &inputs); err != nil {
		t.Fatal(err)
	}
	for _, labels := range []string{"outer", "outer: inner", "outer: middle: inner", "π: 𐐀: _label", "same: same", "a: b: c: d: e: f: g: h"} {
		for _, statement := range []string{"while (x) { continue outer; }", "for (;;) { break outer; }", "switch (x) { case 0: break outer; }", "{ let x = 1; x++; }", "if (x) { x(); }", "try { x(); } catch (e) { y(); }", "while (x) { nested: while (y) { break nested; } }"} {
			inputs = append(inputs, fixture{File: "controls.ts", Rule: "label-controls", Source: labels + ": " + statement})
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
	t.Logf("actual Go numeric labeled-statement kind: %s", bytes.TrimSpace(kinds))
	want := execute(t, "", oracle, "--want", fixtures)
	entry, _ := filepath.Abs("labels_main.a")
	for i, got := range backends(t, entry, corpus) {
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d differs at row %d", i, difference(got, want))
		}
	}
	t.Logf("%d consumer AST and nested-label/identity controls, %d identical Go/source Node/emitted JS/sanitized native bytes", bytes.Count(want, []byte{'\n'}), len(want))
	for _, change := range []struct{ name, old, replacement string }{{"omit-outer-labels", "current = parent;", "break;"}, {"ignore-child-identity", "if(!parent.statementMatches) { break; }", "if(false) { break; }"}, {"wrong-kind", "parent.kind !== 257", "parent.kind !== 0"}} {
		t.Run(change.name, func(t *testing.T) {
			parent := t.TempDir()
			directory := filepath.Join(parent, "wave12")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"labels_main.a", "control_flow_labels_of.a", "../options_json.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(file, "labels_of.a") {
					if strings.Count(string(data), change.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), change.old, change.replacement, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			for i, got := range backends(t, filepath.Join(directory, "labels_main.a"), corpus) {
				if bytes.Equal(got, want) {
					t.Fatalf("backend %d mutant survived", i)
				}
				t.Logf("backend %d compiles and exits cleanly; Go comparison catches mutant at row %d", i, difference(got, want))
			}
		})
	}
}

func TestLabelsNilRefused(t *testing.T) {
	oracle := labelsOracle(t)
	corpus := filepath.Join(t.TempDir(), "nil.json")
	if err := os.WriteFile(corpus, []byte(`[{"present":false,"kind":0,"label":"","statementMatches":false,"parents":[]}]`), 0644); err != nil {
		t.Fatal(err)
	}
	entry, _ := filepath.Abs("labels_main.a")
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
