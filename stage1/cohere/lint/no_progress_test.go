package lint

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoProgressFixRefusal(t *testing.T) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, _ := filepath.Abs("testdata/no_progress_oracle.go")
	virtual := filepath.Join(root, "adamic_no_progress_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(t.TempDir(), "oracle")
	execute(t, root, "go", "build", "-overlay="+path, "-o", oracle, virtual)
	want := execute(t, "", oracle).output
	directory := mutant(t, "removable ? 'fix' : '',\n                '',", "removable ? 'fix' : '',\n                this.context.source.slice(this.context.start(index), node.end),", "rules/no-debugger/rule.ts")
	source := filepath.Join(t.TempDir(), "noop.ts")
	if err := os.WriteFile(source, []byte("debugger; debugger;"), 0644); err != nil {
		t.Fatal(err)
	}
	input := manifest(t, []string{source + "\tno-debugger"})
	check := func(mutated bool) {
		sides := []struct {
			name string
			run  execution
		}{{"Node", node(t, directory, input, false)}, {"emitted JavaScript", emittedNode(t, directory, input, false)}, {"sanitized native", execute(t, "", buildPort(t, directory, true), "--manifest", input)}}
		for _, side := range sides {
			var lines []string
			for _, line := range strings.Split(string(side.run.output), "\n") {
				if strings.HasPrefix(line, "rejected ") || strings.HasPrefix(line, "fixed\t") {
					lines = append(lines, line)
				}
			}
			got := []byte(strings.Join(lines, "\n") + "\n")
			if mutated {
				if bytes.Equal(got, want) {
					t.Fatalf("refusal mutant survived %s", side.name)
				}
				t.Logf("refusal output mutant caught on %s", side.name)
			} else if !bytes.Equal(got, want) {
				t.Fatalf("%s differs from Go: %s", side.name, difference(got, want))
			}
		}
	}
	check(false)
	file := filepath.Join(directory, "lint.ts")
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	const from = "the fix replaces text with itself"
	if strings.Count(string(data), from) != 1 {
		t.Fatal("refusal mutant anchor drift")
	}
	if err := os.WriteFile(file, []byte(strings.Replace(string(data), from, "incorrect refusal", 1)), 0644); err != nil {
		t.Fatal(err)
	}
	check(true)
}
