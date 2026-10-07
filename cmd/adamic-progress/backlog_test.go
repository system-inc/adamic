package main

import (
	"fmt"
	"github.com/system-inc/adamic/internal/stage1progress"
	"strings"
	"testing"
	"time"
)

func TestPatchBacklogRebasedAndSharedChanges(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("base", "base\n")
	git("add", ".")
	git("commit", "-qm", "base")
	base := git("rev-parse", "HEAD")
	git("checkout", "-qb", "landed")
	write("base", "landed\n")
	git("add", ".")
	git("commit", "-qm", "rebased landing")
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	git("checkout", "-qb", "original", base)
	write("base", "landed\n")
	git("add", ".")
	git("commit", "-qm", "original change")
	git("update-ref", "refs/remotes/origin/original", "HEAD")
	if !strings.HasPrefix(git("cherry", "origin/main", "origin/original"), "- ") {
		t.Fatal("git cherry oracle does not recognize landing")
	}
	git("checkout", "-qb", "worker-a", base)
	write("pending", "new\n")
	git("add", ".")
	git("commit", "-qm", "pending A")
	pending := git("rev-parse", "HEAD")
	git("update-ref", "refs/remotes/origin/a", "HEAD")
	git("checkout", "-qb", "worker-b", base)
	write("pending", "n e w\n")
	git("add", ".")
	git("commit", "-qm", "pending B with whitespace")
	git("commit", "--allow-empty", "-qm", "empty")
	git("update-ref", "refs/remotes/origin/b", "HEAD")
	v, err := speed(r, time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if v.Backlog != 4 || v.PatchBacklog == nil || *v.PatchBacklog != 1 {
		t.Fatalf("want 4 identities, 1 patch: %#v", v)
	}
	// The landing gets a new ID and leaves the original remote commits waiting.
	git("checkout", "landed")
	git("cherry-pick", pending)
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	r.cache = nil
	v, err = speed(r, time.Now().Add(time.Second))
	if err != nil || v.Backlog != 4 || v.PatchBacklog == nil || *v.PatchBacklog != 0 {
		t.Fatalf("landed patch still waiting: %#v, %v", v, err)
	}
}

func TestHistoricalBacklogUnitsStaySeparate(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("documentation/velocity/backlog.csv", "timestamp,count\n2026-10-07T12:00:00Z,9\n")
	r = snapshotPaths(t, r, git)
	moment := time.Date(2026, 10, 7, 13, 0, 0, 0, time.UTC)
	metrics, err := historicalVelocity(r, moment)
	if err != nil || !metrics[3].Known || metrics[3].Done != 9 || metrics[4].Known {
		t.Fatal("commit counts credited as patches", metrics, err)
	}
	claim, err := checkBacklog(r, moment.Add(-time.Hour), moment)
	if err != nil || claim.Known {
		t.Fatal("legacy commit archive credited to patch milestone", claim, err)
	}
	write("documentation/velocity/patch-backlog.csv", "timestamp,count\n2026-10-07T12:00:00Z,3\n2026-10-07T13:00:00Z,2\n")
	r = snapshotPaths(t, r, git)
	metrics, err = historicalVelocity(r, moment)
	if err != nil || metrics[3].Done != 9 || metrics[4].Done != 2 {
		t.Fatal("wrong archive units", metrics, err)
	}
	claim, err = checkBacklog(r, moment.Add(-time.Hour), moment)
	if err != nil || !claim.OnTrack {
		t.Fatal(claim, err)
	}
}

func TestHistoricalMissingInventoryIsUnknown(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("stage1/cohere/early/slice.a", "early")
	r = snapshotPaths(t, r, git)
	missing := func(string, string) (*stage1progress.Report, error) {
		return nil, fmt.Errorf("slice stage1/cohere/early needs GAPS.md: missing")
	}
	tracks, err := snapshot(r, git("rev-parse", "HEAD"), time.Now(), missing)
	if err != nil || tracks[1].Measures[0].Known {
		t.Fatal("missing historical record invented", tracks, err)
	}
	write("stage1/cohere/early/GAPS.md", "record exists")
	r = snapshotPaths(t, r, git)
	tracks, err = snapshot(r, git("rev-parse", "HEAD"), time.Now(), missing)
	if err != nil || tracks[1].Measures[0].Known || !strings.Contains(tracks[1].Measures[0].Note, "needs GAPS.md") {
		t.Fatal("unreadable inventory not labeled", tracks, err)
	}
}
