package typeaware

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWave19ThirdStrictTypeof(t *testing.T) {
	directory := os.Getenv("ADAMIC_WAVE19_THIRD_ARTIFACTS")
	if directory == "" {
		t.Skip("run after TestWave19ThirdSimpleRules with persistent artifacts")
	}
	repository, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	oracle := volumeOracle(h, "strict-oracle", "oracle_wave_19_third.go")
	args := []string{filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "valid.manifest"), "--strict-typeof"}
	want := h.must("strict-go", exec.Command(oracle, args...))
	if !bytes.Contains(want.stdout, []byte("notString")) || !bytes.Contains(want.stdout, []byte("suggestString")) {
		t.Fatal("missing strict positives/suggestion")
	}
	for _, name := range []string{"third", "third-asan"} {
		got := h.must("strict-"+name, exec.Command(filepath.Join(directory, name), args...))
		if len(got.stderr) != 0 || !bytes.Equal(got.stdout, want.stdout) {
			t.Fatalf("strict differs at byte %d", firstDifference(got.stdout, want.stdout))
		}
	}
	t.Logf("strict typeof normal and sanitizer match %d complete Go bytes", len(want.stdout))
	root := cloneWave19Third(h, "strict-option", "main.a", "args.includes('--strict-typeof')", "false")
	mutant := h.build(filepath.Join(directory, "adamic"), "strict-option-mutant", filepath.Join(root, "main.a"), filepath.Join(directory, "ancestry.a"), false)
	got := h.must("strict-option-mutant-run", exec.Command(mutant, args...))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, want.stdout) {
		t.Fatal("strict option mutant survived")
	}
	t.Logf("ignored-strict-option mutant exits 0 with empty stderr; only Go bytes catch byte %d", firstDifference(got.stdout, want.stdout))

}
