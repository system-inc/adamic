package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchiveLightestAndMissingInput(t *testing.T) {
	p := plan{Count: 3, Units: []unit{
		{Package: "p", Test: "TestHeavy", Shard: 0, Seconds: 100},
		{Package: "p", Test: "TestMedium", Shard: 1, Seconds: 50},
		{Package: "github.com/system-inc/adamic/internal/native", Test: "TestSplitTSGoAgrees", Seconds: 69.4},
	}}
	if err := assignArchive(&p, nil); err != nil {
		t.Fatal(err)
	}
	if p.Archive == nil || p.Archive.Shard != 2 || p.Units[2].Shard != 2 {
		t.Fatal("archive not on lightest shard", p.Archive)
	}
	t.Setenv(archiveVariable, "")
	if err := archiveReady(p, 2); err == nil || !strings.Contains(err.Error(), "TestSplitTSGoAgrees") || !strings.Contains(err.Error(), archiveVariable) {
		t.Fatal("missing archive not refused by name", err)
	}
	if err := archiveReady(p, 0); err != nil {
		t.Fatal("nonconsumer requires archive", err)
	}
	path := filepath.Join(t.TempDir(), "checker.a")
	if err := os.WriteFile(path, []byte("archive"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(archiveVariable, path)
	if err := archiveReady(p, 2); err != nil {
		t.Fatal(err)
	}
	// A plan mutant moves the test onto a box which received no archive.
	t.Setenv(archiveVariable, "")
	p.Units[2].Shard = 1
	if err := archiveReady(p, 1); err == nil || !strings.Contains(err.Error(), "TestSplitTSGoAgrees") {
		t.Fatal("misplaced archive consumer accepted", err)
	}
}

func TestArchiveCountsRepeatedParentSetup(t *testing.T) {
	p := plan{Count: 2, Units: []unit{{Package: "p", Test: "TestParent/a", Shard: 0, Seconds: 1}, {Package: "p", Test: "TestOther", Shard: 1, Seconds: 50}, {Package: "github.com/system-inc/adamic/internal/native", Test: "TestSplitTSGoAgrees"}}}
	w := map[string]float64{"p::TestParent/a": 1, "p::TestParent": 100}
	if err := assignArchive(&p, w); err != nil {
		t.Fatal(err)
	}
	if p.Archive.Shard != 1 {
		t.Fatal("ignored repeated parent setup", p.Archive)
	}
}

func TestFleetTwentyTwoArchiveBriefs(t *testing.T) {
	root := t.TempDir()
	commit := strings.Repeat("a", 40)
	p := plan{Commit: commit, Count: 22, Units: []unit{{Package: "github.com/system-inc/adamic/internal/native", Test: "TestSplitTSGoAgrees", Shard: 7}}, Archive: &archiveRequirement{Shard: 7, Unit: archiveUnit, Variable: archiveVariable}}
	path := filepath.Join(root, "plan.json")
	if err := saveJSON(path, p); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", "../../cloud/gate/fleet.sh", "briefs", commit, path)
	cmd.Env = append(os.Environ(), "ADAMIC_GATE_FLEET_DIRECTORY="+root)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	files, err := filepath.Glob(filepath.Join(root, commit[:12], "briefs", "shard-*.md"))
	if err != nil || len(files) != 22 {
		t.Fatalf("briefs: %v %v", files, err)
	}
	archive := 0
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), "--gate-inputs-no-archive") {
			t.Fatalf("full archive inputs in %s", file)
		}
		if strings.Contains(string(b), "--gate-archive") {
			archive++
			if filepath.Base(file) != "shard-7.md" {
				t.Fatal("wrong archive brief", file)
			}
		}
	}
	if archive != 1 {
		t.Fatalf("%d archive briefs, want exactly one", archive)
	}
}
