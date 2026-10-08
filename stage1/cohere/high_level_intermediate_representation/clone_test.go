package high_level_intermediate_representation

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Native is blocked by GAPS.md gap 2, so this is explicitly a Node-only certificate.
func TestCloneFunctionNodeOracle(t *testing.T) {
	manifest := os.Getenv("HIR_CLONE_CENSUS")
	if manifest == "" {
		t.Skip("set HIR_CLONE_CENSUS to the complete construction manifest; native clone is pending gap 2")
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	got := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(lane, "clone_main.ts"), manifest)
	matched, total := compareConstructionCensus(t, got, manifest, false)
	t.Logf("Node-only CloneFunction: %d/%d including probes; native blocked by gap 2", matched, total)
	dir, err := os.MkdirTemp(filepath.Dir(lane), "hir-clone-mutant-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	files, err := filepath.Glob(filepath.Join(lane, "*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Base(file) == "clone.ts" {
			const anchor = "dst.blockOrder = src.blockOrder.map((id) => dst.blockAt(id.slot + 1));"
			if strings.Count(string(data), anchor) != 1 {
				t.Fatal("clone mutant anchor moved")
			}
			data = []byte(strings.Replace(string(data), anchor, "dst.blockOrder = src.blockOrder;", 1))
		}
		if err := os.WriteFile(filepath.Join(dir, filepath.Base(file)), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	c := exec.Command("node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), filepath.Join(dir, "clone_main.ts"), manifest)
	c.Dir = root
	out, err := c.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "BlockIndex belongs to another arena") {
		t.Fatalf("clone alias mutant survived: %v %s", err, out)
	}
	t.Log("Node catches clone storage alias at the checked block arena read")
}
