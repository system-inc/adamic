package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/types/program"
)

// The early format pass reads a file before the graph build does (#g3046x5). A file edited between the two
// reads must not keep the early result: the program holds the new bytes, so the result describes a file that
// no longer exists. Here the edit lands for real on disk between the speculation and the build, and the control
// is the same file left alone, without which a speculation that kept nothing would pass.
func TestAnEarlyFormatResultIsDiscardedWhenTheFileChangesBeforeTheBuild(t *testing.T) {
	t.Parallel()
	for _, edited := range []bool{false, true} {
		t.Run(fmt.Sprintf("edited=%t", edited), func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			fileName := filepath.Join(directory, "A.ts")
			if err := os.WriteFile(fileName, []byte("export const a = 1;\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			unchanged := func(_ string, text string, _ *ast.SourceFile) (string, error) { return text, nil }
			speculation := speculateFormatOn([]string{fileName}, unchanged, 1, 1, nil)
			speculation.wait()
			if _, attempted := speculation.attempts[fileName]; !attempted {
				t.Fatal("the speculation attempted nothing, so nothing below is about keeping it")
			}

			if edited {
				if err := os.WriteFile(fileName, []byte("export const a = 2;\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			configuration := `{"compilerOptions":{"strict":true,"module":"esnext","target":"esnext","moduleResolution":"bundler"},"include":["**/*.ts"]}`
			if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(configuration), 0o644); err != nil {
				t.Fatal(err)
			}
			graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
			if err != nil {
				t.Fatal(err)
			}

			_, kept := speculation.keepable(nil, graph)[fileName]
			if kept == edited {
				t.Errorf("edited between the early read and the build: %t, and the early result was kept: %t", edited, kept)
			}
		})
	}
}

// The early pass runs on every core until the walk starts, then narrows to its usual share: the workers past the
// share finish the file they are on and take no more, and the files they leave go to the workers that remain,
// so every file is still attempted. Measured by how many transforms run at once after the narrowing.
func TestANarrowedSpeculationRunsOnFewerWorkersAndStillReachesEveryFile(t *testing.T) {
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
	const workers, narrowed = 8, 2

	// Each transform begun before the narrowing holds its file until the test lets it go, so every worker is
	// mid-file when it lands. Each begun after it holds for a moment, so any two running at once would overlap,
	// and counts how many ran at once.
	started := make(chan struct{}, len(candidates))
	release := make(chan struct{})
	var isNarrowed atomic.Bool
	var runningAfter, mostAfter atomic.Int32
	holding := func(_ string, text string, _ *ast.SourceFile) (string, error) {
		if !isNarrowed.Load() {
			started <- struct{}{}
			<-release
			return text, nil
		}
		now := runningAfter.Add(1)
		defer runningAfter.Add(-1)
		for most := mostAfter.Load(); now > most && !mostAfter.CompareAndSwap(most, now); most = mostAfter.Load() {
		}
		time.Sleep(2 * time.Millisecond)
		return text, nil
	}

	speculation := speculateFormatOn(candidates, holding, 1, workers, nil)
	for range workers {
		<-started
	}
	speculation.narrow(narrowed)
	isNarrowed.Store(true)
	close(release)
	speculation.wait()

	if most := mostAfter.Load(); most == 0 || most > narrowed {
		t.Errorf("after narrowing to %d workers, %d files were formatted at once (0 means none was begun after it)", narrowed, most)
	}
	if len(speculation.attempts) != len(candidates) {
		t.Errorf("narrowed, the speculation kept %d of %d unchanged files: files the narrowed workers left were lost",
			len(speculation.attempts), len(candidates))
	}
}
