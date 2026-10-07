package typeaware

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWave19ContractByteMutant(t *testing.T) {
	directory := os.Getenv("ADAMIC_WAVE19_AWAIT_ARTIFACTS")
	if directory == "" {
		t.Skip("run after TestWave19AwaitRule with persistent artifacts")
	}
	repository, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	root := cloneWave19Third(h, "union-mask", "require-await/contract.a", "134217728", "1048576")
	mutant := h.build(filepath.Join(directory, "adamic"), "union-mask-mutant", filepath.Join(root, "await_main.a"), filepath.Join(directory, "ancestry.a"), false)
	config := filepath.Join(directory, "tsconfig.json")
	manifest := filepath.Join(directory, "contract-0.manifest")
	want := h.must("union-mask-go", exec.Command(filepath.Join(directory, "await-oracle"), config, manifest))
	got := h.must("union-mask-native", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, want.stdout) {
		t.Fatal("contract mutant survived")
	}
	t.Logf("wrong-union-mask mutant exits 0 with empty stderr; only Go bytes catch byte %d", firstDifference(got.stdout, want.stdout))
}
