package typeaware

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

const testVolumeProfileOverlaysShards = 3

type volumeProfileOverlayChange struct{ name, from, to string }

func volumeProfileOverlayChanges() []volumeProfileOverlayChange {
	return []volumeProfileOverlayChange{
		{"index-kind", `candidate.Kind.String() == "Kind"+kind`, `kind != ""`},
		{
			"root-kind",
			`start == uint64(source.Pos()) && end == uint64(source.End()) && kind == "SourceFile"`,
			`start == uint64(source.Pos()) && end == uint64(source.End()) && kind != ""`,
		},
		{
			"index-end",
			`nodeRange{uint64(candidate.Pos()), uint64(candidate.End())}`,
			`nodeRange{uint64(candidate.Pos()), 0}`,
		},
	}
}

// Stable mutation names assign each case to exactly one top-level test.
var volumeProfileOverlaySlices = [][]string{
	{"index-kind"},
	{"root-kind"},
	{"index-end"},
}

func TestVolumeProfileOverlaysUnion(test *testing.T) {
	test.Parallel()
	if len(volumeProfileOverlaySlices) != testVolumeProfileOverlaysShards {
		test.Fatalf("got %d slices, want %d shards", len(volumeProfileOverlaySlices), testVolumeProfileOverlaysShards)
	}
	coverage := make(map[string]int)
	for _, change := range volumeProfileOverlayChanges() {
		if _, exists := coverage[change.name]; exists {
			test.Fatalf("duplicate mutation %s", change.name)
		}
		coverage[change.name] = 0
	}
	if len(coverage) == 0 {
		test.Fatal("empty compiler AST identity mutation enumeration")
	}
	for shard, slice := range volumeProfileOverlaySlices {
		if len(slice) == 0 {
			test.Fatalf("empty shard %03d", shard)
		}
		for _, name := range slice {
			if _, exists := coverage[name]; !exists {
				test.Fatalf("shard %03d contains unknown mutation %s", shard, name)
			}
			coverage[name]++
		}
	}
	for name, count := range coverage {
		if count != 1 {
			test.Fatalf("mutation %s covered %d times, want exactly once", name, count)
		}
	}
	test.Logf(
		"union: %d compiler AST identity mutations covered exactly once across %d shards",
		len(coverage),
		len(volumeProfileOverlaySlices),
	)
}

func TestVolumeProfileOverlays_000(test *testing.T) {
	test.Parallel()
	runVolumeProfileOverlayShard(test, 0)
}

func TestVolumeProfileOverlays_001(test *testing.T) {
	test.Parallel()
	runVolumeProfileOverlayShard(test, 1)
}

func TestVolumeProfileOverlays_002(test *testing.T) {
	test.Parallel()
	runVolumeProfileOverlayShard(test, 2)
}

// volumeProfileOverlayBinary is bridge/tsgo/checker's test binary with change planted in facts.go, a product
// (buildcache.GoTestMutated) built ahead by TestProduct_profile_overlay_*, and the directory go test would run it in.
func volumeProfileOverlayBinary(test *testing.T, change volumeProfileOverlayChange) (binary, directory string) {
	test.Helper()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		test.Fatal(err)
	}
	binary = buildcache.GoTestMutated(test, "", "checker-"+change.name+".test", "./bridge/tsgo/checker", nil,
		[]buildcache.Mutation{{File: "bridge/tsgo/checker/facts.go", Before: change.from, After: change.to}})
	return binary, filepath.Join(repository, "bridge/tsgo/checker")
}

func runVolumeProfileOverlayShard(test *testing.T, shard int) {
	test.Helper()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		test.Fatal(err)
	}
	h := &harness{t: test, repository: repository, directory: test.TempDir()}
	for _, name := range volumeProfileOverlaySlices[shard] {
		found := false
		for _, change := range volumeProfileOverlayChanges() {
			if change.name != name {
				continue
			}
			found = true
			binary, directory := volumeProfileOverlayBinary(test, change)
			// h.run starts its own copy of this command, in a process group of its own under typeAwareRunCommand's 90 s
			// context, whose Cancel kills the whole group, the test binary's children included, and reads the output
			// from files. ctx only times the run.
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			test.Cleanup(cancel)
			command := exec.Command(
				binary,
				"-test.run",
				"^TestExactIndexMatchesCompilerNodes$",
				"-test.count=1",
				"-test.timeout=90s",
			)
			command.Dir = directory
			result := h.run(change.name+"-compiler-node-test", command)
			deadline := ctx.Err() == context.DeadlineExceeded
			cancel()
			if deadline {
				test.Fatalf(
					"COOKED: %s exceeded the 90s hard deadline (60s budget); checker test %.6fs",
					change.name,
					result.elapsed.Seconds(),
				)
			}
			if result.elapsed > 60*time.Second {
				test.Fatalf(
					"COOKED: %s exceeded the 60s budget; checker test %.6fs",
					change.name,
					result.elapsed.Seconds(),
				)
			}
			if result.err == nil || !bytes.Contains(result.stdout, []byte("exact index changed compiler node")) {
				test.Fatalf(
					"%s not caught by AST identity: %v %s %s",
					change.name,
					result.err,
					result.stdout,
					result.stderr,
				)
			}
			test.Logf(
				"%s: compiler AST identity oracle catches wrong selector; checker test %.6fs; cooked=false",
				change.name,
				result.elapsed.Seconds(),
			)
		}
		if !found {
			test.Fatalf("shard %03d contains unknown mutation %s", shard, name)
		}
	}
}

func TestProduct_profile_overlay_index_kind(test *testing.T) {
	test.Parallel()
	volumeProfileOverlayBinary(test, volumeProfileOverlayChanges()[0])
}

func TestProduct_profile_overlay_root_kind(test *testing.T) {
	test.Parallel()
	volumeProfileOverlayBinary(test, volumeProfileOverlayChanges()[1])
}

func TestProduct_profile_overlay_index_end(test *testing.T) {
	test.Parallel()
	volumeProfileOverlayBinary(test, volumeProfileOverlayChanges()[2])
}
