package program_test

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/types/program"
)

// A build answered from the run cache's check builds the program a build asking the disk builds, both ways
// (#kdee854). Module resolution decides which files are in the program at all, so a wrong answer here is a wrong
// program, and the case that must not be missed is a module created since the last run: the last build looked for
// it and was told nothing is there.
//
// The tsconfig names only index.ts, which imports './helper'. The first build resolves it to helper/index.ts,
// having looked for helper.ts and found nothing. Then helper.ts is created. The check stats every recorded path
// again, helper.ts among them, so a build answered from it must resolve './helper' to helper.ts, as the disk
// does. A snapshot that kept the recorded "nothing there" instead, built by hand below, resolves it to the old
// file, which is what proves the comparison can fail. And on the untouched tree the answered build matches too,
// with the check answering most of its questions.
func TestABuildAnsweredFromTheCheckBuildsTheProgramTheDiskDoes(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write := func(name string, contents string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("tsconfig.json", `{"compilerOptions":{"strict":true,"noEmit":true,"module":"esnext","moduleResolution":"bundler"},"files":["source/index.ts"]}`)
	write("source/index.ts", "import { a } from './helper';\nexport const b = a;\n")
	write("source/helper/index.ts", "export const a = 1;\n")
	helper := filepath.Join(root, "source", "helper.ts")

	build := func(snapshot *program.StatSnapshot, recorder *program.InputRecorder) ([]string, int64) {
		t.Helper()
		timing := &program.GraphTiming{}
		graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json"), CurrentDirectory: root,
			Inputs: recorder, CheckedStats: snapshot, Timing: timing})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		var ours []string
		for _, sourceFile := range graph.SourceFiles() {
			if name := filepath.FromSlash(sourceFile.FileName().AsString()); strings.HasPrefix(name, root+string(filepath.Separator)) {
				ours = append(ours, name)
			}
		}
		sort.Strings(ours)
		return ours, timing.CheckedAnswers
	}
	record := func() *program.RunCache {
		t.Helper()
		recorder := program.NewInputRecorder()
		build(nil, recorder)
		present, absent, probed := recorder.Inputs()
		if !slices.Contains(absent, helper) {
			t.Fatalf("the build did not record looking for %s, so the case below is not the one this test is about: %v", helper, absent)
		}
		recorded, err := program.RecordRunCache("key", present, nil, absent, probed, []byte("verdict"), 0, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		return recorded
	}

	// The untouched tree.
	recorded := record()
	snapshot := program.NewStatSnapshot()
	if err := recorded.CheckNoting("key", snapshot); err != nil {
		t.Fatalf("the untouched tree missed: %v", err)
	}
	fromDisk, _ := build(nil, nil)
	fromCheck, answered := build(snapshot, nil)
	if !slices.Equal(fromCheck, fromDisk) {
		t.Fatalf("on the untouched tree the build answered from the check holds %v, the disk's %v", fromCheck, fromDisk)
	}
	if answered == 0 {
		t.Fatal("the check answered none of the build's questions, so the agreement above is the disk agreeing with itself")
	}

	// helper.ts created since the last run.
	time.Sleep(10 * time.Millisecond)
	write("source/helper.ts", "export const a = 'text';\n")
	snapshot = program.NewStatSnapshot()
	if err := recorded.CheckNoting("key", snapshot); err == nil {
		t.Fatal("the check replayed over a created module")
	}
	fromDisk, _ = build(nil, nil)
	if !slices.Contains(fromDisk, helper) {
		t.Fatalf("the disk's build does not resolve './helper' to the new %s, so the fixture is wrong: %v", helper, fromDisk)
	}
	fromCheck, answered = build(snapshot, nil)
	if !slices.Equal(fromCheck, fromDisk) {
		t.Errorf("after helper.ts was created the build answered from the check holds %v, the disk's %v", fromCheck, fromDisk)
	}
	if answered == 0 {
		t.Error("the check answered none of the build's questions after the change")
	}

	// The control: answers kept from the recording rather than taken now resolve to the old file.
	stale := program.StaleStatSnapshotForTest(recorded)
	if fromStale, _ := build(stale, nil); slices.Contains(fromStale, helper) {
		t.Errorf("a snapshot that says nothing is at %s still built a program holding it, so the comparison above cannot fail: %v", helper, fromStale)
	}
}
