package typeaware

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const testVolumeProfileOverlaysShards = 3

type volumeProfileOverlayChange struct{ name, from, to string }

func volumeProfileOverlayChanges() []volumeProfileOverlayChange {
	return []volumeProfileOverlayChange{
		{"index-kind", `candidate.Kind.String() == "Kind"+kind`, `kind != ""`},
		{"root-kind", `start == uint64(source.Pos()) && end == uint64(source.End()) && kind == "SourceFile"`, `start == uint64(source.Pos()) && end == uint64(source.End()) && kind != ""`},
		{"index-end", `nodeRange{uint64(candidate.Pos()), uint64(candidate.End())}`, `nodeRange{uint64(candidate.Pos()), 0}`},
	}
}

// Stable mutation names assign each case to exactly one top-level test.
var volumeProfileOverlaySlices = [][]string{
	{"index-kind"},
	{"root-kind"},
	{"index-end"},
}

func TestVolumeProfileOverlaysUnion(t *testing.T) {
	t.Parallel()
	if len(volumeProfileOverlaySlices) != testVolumeProfileOverlaysShards {
		t.Fatalf("got %d slices, want %d shards", len(volumeProfileOverlaySlices), testVolumeProfileOverlaysShards)
	}
	coverage := make(map[string]int)
	for _, change := range volumeProfileOverlayChanges() {
		if _, exists := coverage[change.name]; exists {
			t.Fatalf("duplicate mutation %s", change.name)
		}
		coverage[change.name] = 0
	}
	if len(coverage) == 0 {
		t.Fatal("empty compiler AST identity mutation enumeration")
	}
	for shard, slice := range volumeProfileOverlaySlices {
		if len(slice) == 0 {
			t.Fatalf("empty shard %03d", shard)
		}
		for _, name := range slice {
			if _, exists := coverage[name]; !exists {
				t.Fatalf("shard %03d contains unknown mutation %s", shard, name)
			}
			coverage[name]++
		}
	}
	for name, count := range coverage {
		if count != 1 {
			t.Fatalf("mutation %s covered %d times, want exactly once", name, count)
		}
	}
	t.Logf("union: %d compiler AST identity mutations covered exactly once across %d shards", len(coverage), len(volumeProfileOverlaySlices))
}

func TestVolumeProfileOverlays_000(t *testing.T) {
	t.Parallel()
	runVolumeProfileOverlayShard(t, 0)
}

func TestVolumeProfileOverlays_001(t *testing.T) {
	t.Parallel()
	runVolumeProfileOverlayShard(t, 1)
}

func TestVolumeProfileOverlays_002(t *testing.T) {
	t.Parallel()
	runVolumeProfileOverlayShard(t, 2)
}

// Overlays read temporary files outside the repository, so these Go tests must
// run directly rather than through the repository-input product cache.
func runVolumeProfileOverlayShard(t *testing.T, shard int) {
	t.Helper()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: t.TempDir()}
	for _, name := range volumeProfileOverlaySlices[shard] {
		found := false
		for _, change := range volumeProfileOverlayChanges() {
			if change.name != name {
				continue
			}
			found = true
			overlay := h.overlay(change.name, "bridge/tsgo/checker/facts.go", change.from, change.to)
			result := h.run(change.name+"-compiler-node-test", exec.Command("timeout", "--signal=KILL", "75s", "go", "test", "-overlay", overlay, "./bridge/tsgo/checker", "-run", "^TestExactIndexMatchesCompilerNodes$", "-count=1", "-timeout=75s"))
			// timeout kills the whole command group, including compiler children;
			// Go's timeout alone does not bound dependency compilation.
			if exit, ok := result.err.(*exec.ExitError); ok && (exit.ExitCode() == 124 || exit.ExitCode() == 137) {
				t.Fatalf("COOKED: %s exceeded the 75s hard deadline (60s budget); go test %.6fs", change.name, result.elapsed.Seconds())
			}
			if result.elapsed > 60*time.Second {
				t.Fatalf("COOKED: %s exceeded the 60s budget; go test %.6fs", change.name, result.elapsed.Seconds())
			}
			if result.err == nil || !bytes.Contains(result.stdout, []byte("exact index changed compiler node")) {
				t.Fatalf("%s not caught by AST identity: %v %s %s", change.name, result.err, result.stdout, result.stderr)
			}
			t.Logf("%s: compiler AST identity oracle catches wrong selector; go test %.6fs; cooked=false", change.name, result.elapsed.Seconds())
		}
		if !found {
			t.Fatalf("shard %03d contains unknown mutation %s", shard, name)
		}
	}
}
