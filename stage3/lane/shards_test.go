package lane

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// The upstream TypeScript suite is partitioned into manifest units (shards.json): file shards balanced from the
// measured census, exact groups for expensive suites, 16 symlink-watcher cases, and the publicApi byte comparison. Each is a
// TestShard_ test of its own (shards_generated_test.go), so Planner cuts them like any other tests and a red names its
// shard. Every one runs shard.py against the adapted tree product, built ahead (adaptedTree), never a tree it builds.

// shardTeardown is what a shard's limit leaves of its test's deadline: shard.py's teardown after it kills upstream,
// writing its report and removing its view.
const shardTeardown = 5 * time.Second

type counts struct {
	Pass int `json:"pass"`
	Fail int `json:"fail"`
	Skip int `json:"skip"`
}

type unit struct {
	Name    string   `json:"name"`
	Command []string `json:"command"`
	Expect  counts   `json:"expect"`
}

// repositoryRoot is the checkout's root, two directories above stage3/lane.
func repositoryRoot(t testing.TB) string {
	t.Helper()
	// Reads: stage3/*
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// boundedCommand runs program from root through the toolchain, ended when ctx is.
func boundedCommand(ctx context.Context, root, program string, arguments ...string) *exec.Cmd {
	return toolchain.CommandContextIn(ctx, root, program, arguments...)
}

// manifestUnit is the unit named name in the manifest at path.
func manifestUnit(path, name string) (unit, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return unit{}, err
	}
	var manifest struct {
		Version int    `json:"version"`
		Units   []unit `json:"units"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return unit{}, err
	}
	if manifest.Version != 1 || len(manifest.Units) == 0 {
		return unit{}, fmt.Errorf("%s holds no version 1 units", path)
	}
	var found []unit
	for _, row := range manifest.Units {
		if row.Name == name {
			found = append(found, row)
		}
	}
	if len(found) != 1 || len(found[0].Command) == 0 {
		return unit{}, fmt.Errorf("%s holds %d commands for %s, not one", path, len(found), name)
	}
	return found[0], nil
}

// runManifestShard runs the manifest's command for name against the adapted tree product and holds its report to
// the manifest: shard.py itself checks the unit's inputs hash, its exact test identities, baseline paths, failure
// titles and upstream exit, and this checks it passed with the counts the manifest expects.
func runManifestShard(t *testing.T, name string) {
	t.Helper()
	tree := filepath.Join(adaptedTree(t), "tree")
	root := repositoryRoot(t)
	selected, err := manifestUnit(filepath.Join(root, "stage3/lane/shards.json"), name)
	if err != nil {
		t.Fatal(err)
	}
	results := t.TempDir()
	// shard.py's limit is this test's: its deadline, the go test deadline when sooner, less a margin for teardown, so
	// shard.py kills upstream and reports a timeout naming the shard before the context kills shard.py.
	deadline := time.Now().Add(55 * time.Second)
	if testDeadline, ok := t.Deadline(); ok && testDeadline.Before(deadline) {
		deadline = testDeadline
	}
	limit := time.Until(deadline) - shardTeardown
	if limit <= 0 {
		t.Fatalf("%s: the test's deadline leaves no time for the shard", name)
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	arguments := selected.Command[1:]
	// -E and -S keep python off the box's PYTHON* settings and site directories, as the adapted tree's build runs it.
	if selected.Command[0] == "python3" {
		arguments = append([]string{"-E", "-S", "-B"}, arguments...)
	}
	command := boundedCommand(ctx, root, selected.Command[0], arguments...)
	command.Env = append(command.Env, "STAGE3_ADAPTED_TREE="+tree, "STAGE3_RESULTS="+results,
		"STAGE3_SHARD_SECONDS="+strconv.FormatFloat(limit.Seconds(), 'f', 1, 64))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %v\n%s", command, err, output)
	}
	data, err := os.ReadFile(filepath.Join(results, name, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Status string   `json:"status"`
		Counts counts   `json:"counts"`
		Errors []string `json:"errors"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "pass" || report.Counts != selected.Expect {
		t.Fatalf("%s: %s with %+v, where the manifest expects %+v: %v", name, report.Status, report.Counts, selected.Expect, report.Errors)
	}
}
