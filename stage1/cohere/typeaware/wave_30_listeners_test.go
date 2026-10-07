package typeaware

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Declaration parity only: the current shared parser has no numeric kind field.
func TestWave30NumericListeners(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_30_LISTENERS_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err = os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_30_listeners_suite.a")
	oracle := volumeOracle(h, "listeners-oracle", "oracle_wave_30_listeners.go")
	rows := []struct{ file, from, to string }{
		{"consistency_no_iso_string_date_cut", "[214, 213, 261]", "[213, 213, 261]"},
		{"correctness_no_callback_in_parse_try", "[259]", "[258]"},
		{"correctness_no_collection_misuse", "[227, 213, 214]", "[226, 213, 214]"},
		{"correctness_no_discarded_outcome", "[245]", "[244]"},
		{"correctness_no_discarded_pure_result", "[245]", "[244]"},
		{"correctness_no_uncleared_race_timeout", "[214]", "[213]"},
		{"correctness_no_process_exit_after_output", "[307]", "[306]"},
		{"correctness_require_blocking_standard_streams", "[307]", "[306]"},
	}
	args := []string{filepath.Join(repository, "cohere")}
	for _, row := range rows {
		args = append(args, row.file)
	}
	truth := h.must("listeners-go", exec.Command(oracle, args...))
	for _, sanitize := range []bool{false, true} {
		name := "listeners"
		if sanitize {
			name += "-asan"
		}
		binary := h.build(stage0, name, entry, archive, sanitize)
		got := h.must(name+"-run", exec.Command(binary))
		if len(got.stderr) != 0 || !bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s differs at byte %d: %s", name, firstDifference(got.stdout, truth.stdout), got.stderr)
		}
	}
	for _, row := range rows {
		name := "listener-mutant-" + row.file
		binary := wave30ProcessMutant(h, stage0, archive, name, "wave_30_listeners_suite.a", row.file+".a", row.from, row.to)
		got := h.must(name+"-run", exec.Command(binary))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s escaped comparison: %s", name, got.stderr)
		}
		t.Logf("%s caught only by Go comparison at byte %d", name, firstDifference(got.stdout, truth.stdout))
	}
	t.Logf("eight listener declarations agree with %d Go bytes", len(truth.stdout))
}
