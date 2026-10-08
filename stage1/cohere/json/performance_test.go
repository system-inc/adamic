package json

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// The profiled -O2 -g binaries must answer the full corpus, not just the profile sample.
func TestProfileSnapshotsAgree(t *testing.T) {
	snapshots := os.Getenv("ADAMIC_JSON_PROFILE_BINARIES")
	if snapshots == "" {
		// census: measurement Opt-in timing or profile artifact comparison; does not replace correctness verification.
		t.Skip("set ADAMIC_JSON_PROFILE_BINARIES to profile snapshot binaries")
	}
	cases := corpusCases(t)
	answers, _ := cohereAnswers(t, cases, false)
	input, expected := protocol(cases, answers)
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	for _, binary := range filepath.SplitList(snapshots) {
		compare(t, binary, execute(t, nil, binary, "--cases", path), expected, cases)
	}
	t.Logf("%d profiled snapshots agree on %d texts, %d answer bytes", len(filepath.SplitList(snapshots)), len(cases), len(expected))
}

func TestCachedWidthMutantIsCaught(t *testing.T) {
	cases := []textCase{{"probe.json", `["` + strings.Repeat("wide", 25) + `",1]`}}
	answers, _ := cohereAnswers(t, cases, false)
	input, expected := protocol(cases, answers)
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	directory := portDirectory(t, &printerMutation{"zero cached width", "doc.ts", "width: stringWidth(text)", "width: 0"})
	entry := filepath.Join(directory, "main.ts")
	binary := filepath.Join(t.TempDir(), "release")
	if err := native.Build(native.C(lowered(t, entry)), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native release", execute(t, nil, binary, "--cases", path)},
		{"Node", onNode(t, entry, "--cases", path)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
			t.Fatalf("%s mutant must execute successfully: %+v", side.name, side.result)
		}
		if string(side.result.stdout) == expected {
			t.Fatalf("%s failed to catch zero cached width", side.name)
		}
		t.Logf("%s byte comparison caught zero cached width", side.name)
	}
}
