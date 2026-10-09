package program_test

import (
	"slices"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/types/program"
)

// TestGraphTimingMeasuresTheBuildItTimes holds the instrument to the build it reports on.
//
// An instrument that is wired in the wrong place still prints plausible numbers, so each count here is
// one the fixture fixes in advance: the loads are exactly the program's files, the tsconfig's include
// pattern lists a directory, the program reads the fixture's own files from disk and asks the disk about
// a package import's candidates. A timing host that the program was not built through would load zero
// files; a timing layer above the cache instead of beneath it would count the cache's answers; either
// fails here.
//
// And timing must change nothing: the timed build holds the same files as an untimed one.
func TestGraphTimingMeasuresTheBuildItTimes(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json":                      minimalConfig,
		"main.ts":                            "import { helper } from './helper';\nimport { packaged } from 'packaged';\nexport const value = helper() + packaged;\n",
		"helper.ts":                          "export function helper(): number { return 1; }\n",
		"node_modules/packaged/package.json": `{"name": "packaged", "types": "index.d.ts"}`,
		"node_modules/packaged/index.d.ts":   "export declare const packaged: number;\n",
	})

	untimed, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building untimed: %v", err)
	}

	timing := &program.GraphTiming{}
	started := time.Now()
	timed, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory, Timing: timing})
	if err != nil {
		t.Fatalf("building timed: %v", err)
	}
	wall := time.Since(started)

	fileNames := func(graph *program.Graph) []string {
		names := []string{}
		for _, sourceFile := range graph.SourceFiles() {
			names = append(names, sourceFile.FileName().AsString())
		}
		slices.Sort(names)
		return names
	}
	if !slices.Equal(fileNames(untimed), fileNames(timed)) {
		t.Fatalf("the timed build holds different files from the untimed one:\nuntimed %v\ntimed   %v", fileNames(untimed), fileNames(timed))
	}

	if timing.Builds != 1 {
		t.Errorf("Builds is %d for one build", timing.Builds)
	}
	if timing.Config <= 0 || timing.Program <= 0 || timing.Verify <= 0 {
		t.Errorf("a wall part read zero: config %v, program %v, verify %v", timing.Config, timing.Program, timing.Verify)
	}
	if parts := timing.Config + timing.Program + timing.Verify; parts > wall {
		t.Errorf("the wall parts add to %v, more than the %v the whole build took", parts, wall)
	}

	if timing.SourceFileLoads != int64(len(timed.SourceFiles())) {
		t.Errorf("the timing host saw %d loads and the program holds %d files, so the program was not built through it",
			timing.SourceFileLoads, len(timed.SourceFiles()))
	}
	if timing.SourceFileSummed <= 0 {
		t.Errorf("%d loads summed to %v", timing.SourceFileLoads, timing.SourceFileSummed)
	}

	if timing.ConfigDisk.Listings.Count == 0 {
		t.Errorf("the tsconfig's include pattern listed no directory: %+v", timing.ConfigDisk)
	}
	// main.ts, helper.ts and the package's declaration are read from disk; the libs come from the binary
	// and the cache above the timing layer answers repeats, so the count is these three and package.json.
	if timing.ProgramDisk.Reads.Count < 3 {
		t.Errorf("the program read %d files from disk, and the fixture has three it must read", timing.ProgramDisk.Reads.Count)
	}
	if timing.ProgramDisk.Existence.Count == 0 {
		t.Errorf("resolving 'packaged' asked the disk nothing: %+v", timing.ProgramDisk)
	}
	if timing.ProgramDisk.Summed() <= 0 {
		t.Errorf("the program's disk calls summed to %v", timing.ProgramDisk.Summed())
	}
}
