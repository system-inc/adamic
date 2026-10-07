package wave12

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestControlFlowBlockAllocation(t *testing.T) {
	root, _ := filepath.Abs("../../../../../cohere")
	main, _ := filepath.Abs("testdata/block_oracle.go.txt")
	exports, _ := filepath.Abs("testdata/block_exports.go.txt")
	virtual := filepath.Join(root, "adamic_wave12_block_oracle.go")
	mapping, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: main, filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/adamic_wave12_blocks.go"): exports}})
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, mapping, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	execute(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	corpus, _ := filepath.Abs("testdata/block_counts.json")
	want := execute(t, "", binary, corpus)
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("block_main.a")
	got := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, corpus)
	if !bytes.Equal(got, want) {
		t.Fatalf("source Node differs at row %d", difference(got, want))
	}
	t.Logf("actual Go/source Node %d allocation/identity/state observations, %d identical bytes", bytes.Count(want, []byte{'\n'}), len(want))
	parent := t.TempDir()
	directory := filepath.Join(parent, "wave12")
	if err := os.Mkdir(directory, 0755); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"block_main.a", "control_flow_new_block.a", "../options_json.ts"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(file, "new_block.a") {
			old := "chunk[index % 8]"
			if strings.Count(string(data), old) != 1 {
				t.Fatal("mutant anchor changed")
			}
			data = []byte(strings.Replace(string(data), old, "chunk[0]", 1))
		}
		if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutant := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "block_main.a"), corpus)
	if bytes.Equal(mutant, want) {
		t.Fatal("source Node alias mutant survived")
	}
	t.Logf("source Node alias mutant exits cleanly; only Go comparison catches row %d", difference(mutant, want))
	for i, got := range backends(t, entry, corpus) {
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d differs at row %d", i, difference(got, want))
		}
	}
	t.Logf("actual Go/source Node/emitted JavaScript/sanitized native identical")
	for i, got := range backends(t, filepath.Join(directory, "block_main.a"), corpus) {
		if bytes.Equal(got, want) {
			t.Fatalf("backend %d alias mutant survived", i)
		}
		t.Logf("backend %d compiles and exits cleanly; only Go comparison catches row %d", i, difference(got, want))
	}
}
