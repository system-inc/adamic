package typeaware

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const testVolumeProfileOverlaysShards = 3

// Each compiler AST identity mutation belongs to exactly one parallel shard.
// Overlays read temporary files outside the repository, so these Go tests must
// run directly rather than through the repository-input product cache.
func TestVolumeProfileOverlays(t *testing.T) {
	started := time.Now()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	changes := []struct{ name, from, to string }{
		{"index-kind", `candidate.Kind.String() == "Kind"+kind`, `kind != ""`},
		{"root-kind", `start == uint64(source.Pos()) && end == uint64(source.End()) && kind == "SourceFile"`, `start == uint64(source.Pos()) && end == uint64(source.End()) && kind != ""`},
		{"index-end", `nodeRange{uint64(candidate.Pos()), uint64(candidate.End())}`, `nodeRange{uint64(candidate.Pos()), 0}`},
	}
	if len(changes) != testVolumeProfileOverlaysShards {
		t.Fatalf("overlay enumeration has %d cases, want %d shards", len(changes), testVolumeProfileOverlaysShards)
	}
	for i, change := range changes {
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			h := &harness{t: t, repository: repository, directory: t.TempDir()}
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
		})
	}
	t.Logf("union: %d compiler AST identity mutations; TestVolumeProfileOverlays (setup): %.6fs", len(changes), time.Since(started).Seconds())
}
