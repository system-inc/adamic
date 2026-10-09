package lint

import (
	"bytes"
	"fmt"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/stage1/cohere/lint/shards"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestCompilerAndStage1Agree(t *testing.T) {
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Skip("set ADAMIC_TYPESCRIPT_SOURCE to pinned v6.0.3")
	}
	patterns := []string{"*.ts", "*.a"}
	rows := corpusfiles.Upstream(t, source, compilerCommit, []string{"src/compiler"}, patterns)
	rows = append(rows, compilerStage1Sources(t, repository)...)
	// The old walk also included TestMain's generated dispatch. Preserve that
	// coverage as an explicit generated input, independent of stray worktree files.
	generatedRegistry, err := filepath.Abs(".generated/registry.ts")
	if err != nil {
		t.Fatal(err)
	}
	rows = append(rows, generatedRegistry)
	t.Logf("generated input retained from old walk: %s (count 1)", generatedRegistry)

	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("compiler and stage1: %d files", len(rows))
	path := manifest(t, rows)
	want := execute(t, "", goOracle(t), "--manifest", path)
	binary := buildPort(t, directory, true)
	module := emittedJavaScript(t, directory)
	prepareRegistry(t, directory)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	count := min(runtime.NumCPU(), 8)
	assignments := compilerShardAssignments(t, rows, count)
	for _, side := range []struct {
		name    string
		command string
		args    []string
	}{
		{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts")}},
		{"emitted JavaScript", "node", []string{"--disable-warning=ExperimentalWarning", runner, module}},
		{"sanitized native", binary, nil},
	} {
		launcher, timings := compilerShardLauncher(t, side.command, side.args, rows, assignments)
		started := time.Now()
		got, err := shards.Run(launcher, path, count, false)
		if err != nil {
			t.Fatalf("%s: %s", side.name, compilerCasePath(rows, err.Error()))
		}
		for index := 0; index < count; index++ {
			elapsed, err := os.ReadFile(filepath.Join(timings, fmt.Sprintf("%d-%d.time", index, count)))
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%s shard %d/%d: %s ms", side.name, index, count, bytes.TrimSpace(elapsed))
		}
		t.Logf("%s: %d shards in %s", side.name, count, time.Since(started))
		if diff := difference(got, want.output); diff != "" {
			t.Fatalf("%s: %s", side.name, compilerCasePath(rows, diff))
		}
	}
	t.Logf("Go, Node, emitted JavaScript, native identical: %d bytes", len(want.output))
}
