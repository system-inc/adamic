package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/edit"
)

// A stopped speculation finishes the files its workers are on and begins no more, so every file it did not
// reach is left to formatInParallel, which computes it as it always did, and the run's output is the same
// (#679s763). Never stopped, it keeps every file it leaves unchanged: the control, without which a
// speculation that attempted nothing would pass.
func TestAStoppedSpeculationBeginsNoMoreFiles(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	var candidates []string
	for index := range 40 {
		fileName := filepath.Join(directory, fmt.Sprintf("f%02d.md", index))
		if err := os.WriteFile(fileName, []byte("# Title\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		candidates = append(candidates, fileName)
	}
	workers := speculationWorkers()
	if workers == 0 || workers >= len(candidates) {
		t.Skipf("%d speculation workers for %d files, so a stop could not leave any file unbegun", workers, len(candidates))
	}

	// A formatter that rewrites every third file, so the pass after the walk has real work as well as kept
	// results to take.
	format := func(fileName string, text string, _ *ast.SourceFile) (string, error) {
		if index := slices.Index(candidates, fileName); index%3 == 0 {
			return strings.ToUpper(text), nil
		}
		return text, nil
	}
	// Each transform reports that it began, then holds its file until the test lets it go, so the stop lands
	// while every worker is mid-file.
	started := make(chan struct{}, len(candidates))
	release := make(chan struct{})
	holding := func(fileName string, text string, parsed *ast.SourceFile) (string, error) {
		started <- struct{}{}
		<-release
		return format(fileName, text, parsed)
	}

	stopped := speculateFormat(candidates, holding, 1, nil)
	for range workers {
		<-started
	}
	stopped.stop()
	close(release)
	stopped.wait()
	if len(stopped.attempts) > workers {
		t.Errorf("stopped with %d workers mid-file, the speculation kept %d files", workers, len(stopped.attempts))
	}

	// What the run reports is the same whether speculation was stopped or never ran: the pass after the walk
	// computes every file the stopped speculation did not keep.
	process := func(fileName string) (edit.FileResult, error) {
		return edit.CheckFile(fileName, func(string, string) ([]edit.Proposal, error) { return nil, nil }, format, 1)
	}
	withStop := formatInParallel(candidates, nil, stopped.keepable(nil, nil), process)
	without := formatInParallel(candidates, nil, nil, process)
	for index, fileName := range candidates {
		if withStop[index].done != without[index].done || withStop[index].result.Text != without[index].result.Text ||
			withStop[index].result.Changed != without[index].result.Changed {
			t.Errorf("%s: after a stopped speculation %+v, without one %+v", filepath.Base(fileName), withStop[index], without[index])
		}
	}

	free := speculateFormat(candidates, format, 1, nil)
	free.wait()
	if want := len(candidates) - (len(candidates)+2)/3; len(free.attempts) != want {
		t.Fatalf("never stopped, the speculation kept %d of %d files, want the %d it leaves unchanged", len(free.attempts), len(candidates), want)
	}
}
