package main

import (
	"fmt"
	"github.com/system-inc/adamic/internal/stage1progress"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMainRealTwoLineRecords(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	prefix := "stage3/meter/runs/20261007T053859Z.fVzNrN/"
	for _, line := range []string{"main", "area"} {
		b, err := os.ReadFile("testdata/main-71d7e491/" + line + ".json")
		if err != nil {
			t.Fatal(err)
		}
		write(prefix+line+"/report.json", string(b))
		if line == "area" {
			write(prefix+"report.json", string(b))
		}
	}
	b, err := os.ReadFile("testdata/main-71d7e491/landings.csv")
	if err != nil {
		t.Fatal(err)
	}
	write("documentation/velocity/landings.csv", string(b))
	r = snapshotPaths(t, r, git)
	now, _ := time.Parse(time.RFC3339, "2026-10-07T16:38:00Z")
	r.at = now
	s3, err := stage3(r)
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]float64{"compiler checker": 0, "compiler lowering": 0, "compiler own-file checker": 22, "area compiler checker": 1, "area compiler own-file checker": 25} {
		found := false
		for _, m := range s3.Measures {
			if m.Name == name {
				found = true
				if !m.Known || m.Done != want {
					t.Fatalf("%s: %+v", name, m)
				}
				if !strings.Contains(m.Source, "/main/") && !strings.HasPrefix(name, "area ") {
					t.Fatal("area credited to main", m)
				}
			}
		}
		if !found {
			t.Fatal("missing measure", name)
		}
	}
	got, err := readLandings(b, now)
	if err != nil || got["2026-10-07T05:00:00Z"] != 99 {
		t.Fatal("real CSV header or commits_landed ignored", got, err)
	}
	// A corrupt newer run must report its recorded and recomputed totals and leave the valid run visible.
	write("stage3/meter/runs/new/main/report.json", `{"timestamp_utc":"20261007T060000Z","files":[{"file":"src/compiler/x.ts","source":true}],"totals":{"source_files":2,"checker":0,"lowering":0}}`)
	r = snapshotPaths(t, r, git)
	r.at = now
	s3, err = stage3(r)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range s3.Measures {
		if strings.Contains(m.Note, "recorded source/checker/lowering=2/0/0; per-file=1/0/0") {
			found = true
		}
	}
	if !found || s3.Measures[0].Source != prefix+"main/report.json" {
		t.Fatal("bad run suppressed valid run or missing totals", s3)
	}
}

func TestSlowReadDoesNotCancelOtherReads(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("file", "record")
	r = snapshotPaths(t, r, git)
	executable, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	shim := filepath.Join(t.TempDir(), "git")
	script := "#!/bin/sh\nfor arg do\n if [ \"$arg\" = \"HEAD:slow\" ]; then exec sleep 2; fi\ndone\nexec '" + executable + "' \"$@\"\n"
	if err = os.WriteFile(shim, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	r.gitExecutable = shim
	r.readBudget = 100 * time.Millisecond
	if _, err = r.git("show", "HEAD:slow"); err == nil {
		t.Fatal("unbounded slow read")
	}
	if _, err = r.git("rev-parse", "HEAD"); err != nil {
		t.Fatal("slow read canceled other reads", err)
	}
	tracks, err := snapshot(r, git("rev-parse", "HEAD"), time.Now(), func(string, string) (*stage1progress.Report, error) {
		return nil, fmt.Errorf("one slow inventory read")
	})
	if err != nil || tracks[1].Measures[0].Known || !tracks[2].Measures[0].Known {
		t.Fatal("one track killed history", tracks, err)
	}
}

func TestMilestoneRecordFailurePreservesOtherClaims(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("stage1/progress.json", "{broken")
	r = snapshotPaths(t, r, git)
	goals, err := milestones(r, dashboard{Time: time.Now()})
	if err != nil || len(goals) < 10 {
		t.Fatal("one bad record killed milestones", goals, err)
	}
	found := false
	for _, g := range goals {
		for _, e := range g.Evidence {
			if strings.Contains(e.Note, "stage1/progress.json") {
				found = true
				if e.Known || g.Status == "done" {
					t.Fatal("invalid record earned credit", g)
				}
			}
		}
	}
	if !found {
		t.Fatal("record failure hidden")
	}
}
