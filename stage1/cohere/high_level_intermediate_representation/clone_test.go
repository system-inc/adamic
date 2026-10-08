package high_level_intermediate_representation

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Gap 2 remains proven; the authorized explicit-record workaround restores native.
func TestCloneFunctionOracle(t *testing.T) {
	t.Parallel()
	manifest := os.Getenv("HIR_CLONE_CENSUS")
	if manifest == "" {
		manifest = filepath.Join(exportConstructionCensus(t, os.Getenv("HIR_CENSUS_EXPORT")), "manifest.tsv")
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	got := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(lane, "clone_main.ts"), manifest)
	matched, total := compareConstructionCensus(t, got, manifest, false)
	binary := filepath.Join(t.TempDir(), "clone")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", filepath.Join(lane, "clone_main.ts"), "-o", binary)
	native := command(t, root, nil, binary, manifest)
	if string(native) != string(got) {
		t.Fatal("native and Node clone differ")
	}
	t.Logf("CloneFunction native = Node: %d/%d including probes", matched, total)
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
	badBinary := filepath.Join(t.TempDir(), "bad-clone")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", filepath.Join(dir, "clone_main.ts"), "-o", badBinary)
	for _, args := range [][]string{{"node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), filepath.Join(dir, "clone_main.ts"), manifest}, {badBinary, manifest}} {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = root
		out, err := c.CombinedOutput()
		if err == nil || !strings.Contains(string(out), "BlockIndex belongs to another arena") {
			t.Fatalf("clone alias mutant survived: %v %s", err, out)
		}
	}
	t.Log("native and Node catch clone storage alias at the checked block arena read")
}
