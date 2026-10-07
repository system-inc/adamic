package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Fast-import still creates real Git commits and branches, without 3,000
// separate commit subprocesses obscuring the backlog measurement.
func largeBacklogFixture(t *testing.T, ctx context.Context) repository {
	t.Helper()
	root := t.TempDir()
	cmd := exec.CommandContext(ctx, "git", "-C", root, "init", "-q")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init: %v %s", err, out)
	}
	var stream bytes.Buffer
	stamp := time.Now().Unix()
	emit := func(ref string, mark, parent int, file, body, message string) {
		fmt.Fprintf(&stream, "commit %s\nmark :%d\ncommitter fixture <fixture@example.invalid> %d +0000\ndata %d\n%s\n", ref, mark, stamp, len(message), message)
		if parent > 0 {
			fmt.Fprintf(&stream, "from :%d\n", parent)
		}
		fmt.Fprintf(&stream, "M 100644 inline %s\ndata %d\n%s\n\n", file, len(body), body)
	}
	emit("refs/remotes/origin/main", 1, 0, "base", "base\n", "base")
	body := strings.Repeat("fixture payload for bounded patch hashing\n", 200)
	emit("refs/remotes/origin/main", 2, 1, "file-0", body, "landed zero")
	emit("refs/remotes/origin/main", 3, 2, "file-1", body, "landed one")
	mark := 4
	for branch := 0; branch < 300; branch++ {
		parent := 1
		for step := 0; step < 10; step++ {
			emit(fmt.Sprintf("refs/remotes/origin/worker%03d", branch), mark, parent, fmt.Sprintf("file-%d", step), body, fmt.Sprintf("branch %d step %d", branch, step))
			parent = mark
			mark++
		}
	}
	cmd = exec.CommandContext(ctx, "git", "-C", root, "fast-import", "--quiet")
	cmd.Stdin = &stream
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fast-import: %v %s", err, out)
	}
	return repository{root: root, ref: "origin/main", ctx: ctx, backlogContext: ctx}
}

func TestBacklog300Branches3000Commits(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	creation := time.Now()
	r := largeBacklogFixture(t, ctx)
	creationTime := time.Since(creation)
	branches, err := r.git("for-each-ref", "--format=%(refname)", "refs/remotes/origin")
	if err != nil || len(strings.Fields(string(branches))) != 301 {
		t.Fatal("fixture branches", err)
	}
	total, err := r.git("rev-list", "--all", "--count")
	if err != nil || strings.TrimSpace(string(total)) != "3003" {
		t.Fatal("fixture commits", string(total), err)
	}
	start := time.Now()
	v, err := speed(r, time.Now())
	cold := time.Since(start)
	if err != nil || v.PatchError != "" || v.Backlog != 3000 || v.PatchBacklog == nil || *v.PatchBacklog != 8 {
		t.Fatalf("cold count: %#v, %v", v, err)
	}
	if v.PatchStats.Hashed != 3002 || v.PatchStats.Batches != (3002+patchBatchSize-1)/patchBatchSize {
		t.Fatalf("not bounded: %#v", v.PatchStats)
	}
	data, err := os.ReadFile(v.PatchStats.CacheFile)
	var cache patchCache
	if err != nil || json.Unmarshal(data, &cache) != nil || len(cache.Commits) != 3002 {
		t.Fatal("cache not persisted", err)
	}
	// Forbid hashing on the warm run: the cache must be used, not just written.
	r.gitExecutable = gitShim(t, "if [ \"$arg\" = \"patch-id\" ]; then exit 93; fi")
	start = time.Now()
	v, err = speed(r, time.Now())
	warm := time.Since(start)
	if err != nil || v.PatchError != "" || v.PatchBacklog == nil || *v.PatchBacklog != 8 || v.PatchStats.Hashed != 0 || v.PatchStats.Cached != 3002 {
		t.Fatalf("warm cache unused: %#v, %v", v, err)
	}
	// Run the complete report as well, under the same 60-second test deadline.
	start = time.Now()
	d, err := collect(r, time.Now(), lines)
	report := time.Since(start)
	if err != nil || d.Velocity.PatchError != "" || d.Velocity.PatchBacklog == nil || *d.Velocity.PatchBacklog != 8 {
		t.Fatalf("complete report failed: %#v %v", d.Velocity, err)
	}
	var text bytes.Buffer
	renderReport(&text, d, true)
	if !strings.Contains(text.String(), "Stage 3") || !strings.Contains(text.String(), "8 distinct patches waiting") || !strings.HasSuffix(text.String(), d.Slowest+"\n") {
		t.Fatal("incomplete report", text.String())
	}
	binary := buildProgressBinary(t, ctx)
	start = time.Now()
	command := exec.CommandContext(ctx, binary, "--json")
	command.Dir = r.root
	output, err := command.Output()
	cliWall := time.Since(start)
	var cli dashboard
	if err != nil || json.Unmarshal(output, &cli) != nil || cli.Velocity.PatchBacklog == nil || *cli.Velocity.PatchBacklog != 8 || cli.Velocity.PatchStats.Hashed != 0 || len(cli.Tracks) != 3 {
		t.Fatalf("full CLI failed: %v %s", err, output)
	}
	t.Logf("instrument=time.Now/time.Since monotonic wall clock; full CLI warm report=%.3fs (native executable)", cliWall.Seconds())
	t.Logf("instrument=time.Now/time.Since monotonic wall clock; branches=300; off-main commits=3000; creation=%.3fs; cold=%.3fs; warm=%.3fs; full warm report=%.3fs; total=%.3fs", creationTime.Seconds(), cold.Seconds(), warm.Seconds(), report.Seconds(), time.Since(creation).Seconds())
}

func gitShim(t *testing.T, action string) string {
	t.Helper()
	real, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "git")
	script := "#!/bin/sh\nfor arg do\n" + action + "\ndone\nexec '" + strings.ReplaceAll(real, "'", "'\\''") + "' \"$@\"\n"
	if err = os.WriteFile(file, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestBacklogFailureKeepsReport(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"killed", "deadline", "cache"} {
		t.Run(mode, func(t *testing.T) {
			r, write, git := fixture(t)
			write("base", "base")
			write("examples/apple/window.a", "window")
			git("add", ".")
			git("commit", "-qm", "main")
			git("update-ref", "refs/remotes/origin/main", "HEAD")
			write("pending", "new")
			git("add", ".")
			git("commit", "-qm", "pending")
			git("update-ref", "refs/remotes/origin/worker", "HEAD")
			switch mode {
			case "killed":
				r.gitExecutable = gitShim(t, "if [ \"$arg\" = \"patch-id\" ]; then kill -KILL $$; fi")
			case "deadline":
				r.backlogBudget = 40 * time.Millisecond
				r.gitExecutable = gitShim(t, "if [ \"$arg\" = \"patch-id\" ]; then exec sleep 2; fi")
			case "cache":
				write(".git/adamic-progress/patch-ids-v1.json", "invalid JSON")
			}
			start := time.Now()
			d, err := collect(r, time.Now(), lines)
			if err != nil || d.Velocity.PatchBacklog != nil || d.Velocity.PatchError == "" {
				t.Fatal("backlog failure killed report or fabricated count", d.Velocity, err)
			}
			if mode == "deadline" && (!strings.Contains(d.Velocity.PatchError, "deadline") || time.Since(start) > time.Second) {
				t.Fatal("deadline ignored", d.Velocity.PatchError, time.Since(start))
			}
			if d.Tracks[2].Measures[0].Done != 1 {
				t.Fatal("lost available Apple evidence")
			}
			var text bytes.Buffer
			renderReport(&text, d, true)
			if !strings.Contains(text.String(), "Backlog missing:") || !strings.Contains(text.String(), "Stage 1") || !strings.HasSuffix(text.String(), d.Slowest+"\n") {
				t.Fatal("missing rest of report", text.String())
			}
			data, err := json.Marshal(d)
			if err != nil || !bytes.Contains(data, []byte(`"distinct_patch_backlog":null`)) {
				t.Fatal("missing backlog not null", string(data), err)
			}
		})
	}
}

func TestPatchBatchBoundAndIncrementalCache(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("base", "base")
	git("add", ".")
	git("commit", "-qm", "main")
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	for i := 0; i < 17; i++ {
		write("file"+strconv.Itoa(i), "new")
		git("add", ".")
		git("commit", "-qm", fmt.Sprint(i))
	}
	git("update-ref", "refs/remotes/origin/worker", "HEAD")
	first, err := r.cachedPatchBacklog([]string{"origin/worker"})
	if err != nil || first.Hashed != 17 || first.Batches != 2 {
		t.Fatal("unbounded first run", first, err)
	}
	if _, err = r.patchIDs(make([]string, patchBatchSize+1)); err == nil {
		t.Fatal("oversized batch accepted")
	}
	write("another", "fresh")
	git("add", ".")
	git("commit", "-qm", "fresh")
	git("update-ref", "refs/remotes/origin/worker", "HEAD")
	second, err := r.cachedPatchBacklog([]string{"origin/worker"})
	if err != nil || second.Hashed != 1 || second.Cached != 17 || second.Batches != 1 || second.Count != 18 {
		t.Fatal("did not hash just new commit", second, err)
	}
}

func buildProgressBinary(t *testing.T, ctx context.Context) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "adamic-progress")
	command := exec.CommandContext(ctx, "go", "build", "-o", file, ".")
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v %s", err, out)
	}
	return file
}

func TestCLIBacklogFailureExitsZero(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	binary := buildProgressBinary(t, ctx)
	r, write, git := fixture(t)
	write("examples/apple/window.a", "window")
	git("add", ".")
	git("commit", "-qm", "main")
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	write("pending", "new")
	git("add", ".")
	git("commit", "-qm", "pending")
	git("update-ref", "refs/remotes/origin/worker", "HEAD")
	shim := gitShim(t, "if [ \"$arg\" = \"patch-id\" ]; then kill -KILL $$; fi")
	for _, args := range [][]string{{}, {"--json"}, {"--history"}} {
		command := exec.CommandContext(ctx, binary, args...)
		command.Dir = r.root
		command.Env = append(os.Environ(), "PATH="+filepath.Dir(shim)+string(os.PathListSeparator)+os.Getenv("PATH"))
		out, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("CLI %v exited nonzero: %v %s", args, err, out)
		}
		if len(args) > 0 && args[0] == "--json" {
			var d dashboard
			if err = json.Unmarshal(out, &d); err != nil || d.Velocity.PatchError == "" || d.Velocity.PatchBacklog != nil || len(d.Tracks) != 3 {
				t.Fatalf("wrong partial JSON: %v %s", err, out)
			}
		} else if !bytes.Contains(out, []byte("Backlog missing:")) || !bytes.Contains(out, []byte("Stage 1")) || !bytes.Contains(out, []byte("Slowest:")) {
			t.Fatalf("partial report missing: %s", out)
		}
	}
}

func TestBacklogBudgetIndependentOfReportReads(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("base", "base")
	git("add", ".")
	git("commit", "-qm", "main")
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	write("pending", "new")
	git("add", ".")
	git("commit", "-qm", "pending")
	git("update-ref", "refs/remotes/origin/worker", "HEAD")
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	r.ctx = ctx
	r.backlogContext = context.Background()
	r.backlogBudget = 2 * time.Second
	r.gitExecutable = gitShim(t, "if [ \"$arg\" = \"patch-id\" ]; then sleep 0.6; fi")
	d, err := collect(r, time.Now(), lines)
	if err != nil || ctx.Err() == nil || d.Velocity.PatchBacklog == nil || *d.Velocity.PatchBacklog != 1 || d.Velocity.PatchError != "" {
		t.Fatal("fast timeout killed backlog", d.Velocity, err)
	}
}

func TestFailedSectionKeepsOtherTracks(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("examples/apple/window.a", "window")
	write("stage3/meter/runs/bad/report.json", "not JSON")
	write("documentation/velocity/landings.csv", "timestamp\ninvalid\n")
	git("add", ".")
	git("commit", "-qm", "bad records")
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	d, err := collect(r, time.Now(), lines)
	if err != nil || len(d.SectionErrors) < 2 || d.Tracks[2].Measures[0].Done != 1 {
		t.Fatal("failed section killed other tracks", d, err)
	}
	var out bytes.Buffer
	renderReport(&out, d, true)
	if !strings.Contains(out.String(), "Section missing:") || !strings.Contains(out.String(), "Slowest:") {
		t.Fatal(out.String())
	}
}

func TestCompletedPatchBatchesSurviveFailure(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("base", "base")
	git("add", ".")
	git("commit", "-qm", "main")
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	for i := 0; i < 17; i++ {
		write(fmt.Sprint("file", i), "new")
		git("add", ".")
		git("commit", "-qm", fmt.Sprint(i))
	}
	git("update-ref", "refs/remotes/origin/worker", "HEAD")
	marker := filepath.Join(t.TempDir(), "first-batch")
	r.gitExecutable = gitShim(t, "if [ \"$arg\" = \"patch-id\" ]; then if [ -f '"+marker+"' ]; then kill -KILL $$; fi; : > '"+marker+"'; fi")
	first, err := r.cachedPatchBacklog([]string{"origin/worker"})
	if err == nil || first.Hashed != 16 || first.Batches != 1 {
		t.Fatal("missing completed batch checkpoint", first, err)
	}
	r.gitExecutable = ""
	second, err := r.cachedPatchBacklog([]string{"origin/worker"})
	if err != nil || second.Cached != 16 || second.Hashed != 1 || second.Count != 17 {
		t.Fatal("completed batch rehashed", second, err)
	}
	// A valid JSON file with invalid IDs must not fabricate completion.
	data, err := os.ReadFile(second.CacheFile)
	var cache patchCache
	if err != nil || json.Unmarshal(data, &cache) != nil {
		t.Fatal(err)
	}
	for commit := range cache.Commits {
		cache.Commits[commit] = "not a patch id"
		break
	}
	data, err = json.Marshal(cache)
	if err != nil || os.WriteFile(second.CacheFile, data, 0600) != nil {
		t.Fatal(err)
	}
	if _, err = r.cachedPatchBacklog([]string{"origin/worker"}); err == nil {
		t.Fatal("invalid cache trusted")
	}
}
