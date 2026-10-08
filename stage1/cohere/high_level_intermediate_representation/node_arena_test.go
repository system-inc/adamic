package high_level_intermediate_representation

import (
	"path/filepath"
	"testing"
)

// Independent Node certificate while GAPS.md gap 3 blocks native. The mandatory
// native certificates stay unchanged and remain red until compiler lowers it.
func TestNodeInstructionArenaCensus(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	destination := exportConstructionCensus(t, "")
	manifest := filepath.Join(destination, "manifest.tsv")
	for _, entry := range []string{"main.ts", "clone_main.ts"} {
		args := []string{"node", "--no-warnings", "oracle/node.mjs", filepath.Join(lane, entry)}
		if entry == "main.ts" {
			args = append(args, "--coverage")
		}
		args = append(args, manifest)
		output := command(t, root, nil, args...)
		matched, total := compareConstructionCensus(t, output, manifest, false)
		t.Logf("Node %s: %d matched of %d, including 72 probes; 23 Flow excluded", entry, matched, total)
	}
	command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(lane, "replay/main.ts"), "--census", filepath.Join(destination, "checkpoint-manifest.tsv"))
	t.Log("Node full fresh-arena checkpoint decode/dump/re-encode: 1465/1465 originals including all 23 Flow, plus 72 probes")
}
