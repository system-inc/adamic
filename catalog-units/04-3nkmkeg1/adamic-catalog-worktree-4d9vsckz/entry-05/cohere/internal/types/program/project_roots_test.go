package program

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A file the config's glob lists can be gone, or unreadable, by the time the program reads it. That is
// routine with many writers sharing the tree, and on 2026-10-03 it panicked a build on ahra with a
// message blaming path casing (3,790 named, 3,789 loaded). These pin what Build does instead.

func TestClassifyMissingRootsTellsTheFourCasesApart(t *testing.T) {
	t.Parallel()
	states := map[string]rootState{"gone.ts": rootAbsent, "locked.ts": rootUnreadable, "cased.ts": rootReadable}
	inspect := func(fileName string) rootState { return states[fileName] }

	cases := []struct {
		name    string
		missing []string
		want    rootsVerdict
	}{
		{"nothing missing", nil, rootsComplete},
		{"every missing root is gone, so the tree moved", []string{"gone.ts"}, rootsMoved},
		{"a missing root exists but cannot be read", []string{"locked.ts"}, rootsUnreadable},
		{"a missing root exists and reads, so it did not match", []string{"cased.ts"}, rootsUnmatched},
		// The order is the order of seriousness: one real mismatch is a defect however much else moved,
		// so it is never retried away.
		{"a mismatch outranks a moved tree", []string{"gone.ts", "cased.ts"}, rootsUnmatched},
		{"an unreadable root outranks a moved tree", []string{"gone.ts", "locked.ts"}, rootsUnreadable},
	}
	for _, testCase := range cases {
		if got := classifyMissingRoots(testCase.missing, inspect); got != testCase.want {
			t.Errorf("%s: got verdict %d, want %d", testCase.name, got, testCase.want)
		}
	}
}

// The whole path, with no seam: a root the glob lists and the program cannot read. Before this, Build
// returned a graph and ProjectFiles panicked.
func TestBuildRefusesARootItCannotReadInsteadOfPanicking(t *testing.T) {
	t.Parallel()
	directory := writeMismatchFixture(t, map[string]string{
		"tsconfig.json": mismatchFixtureConfig,
		"Alpha.ts":      "export const alpha = 1;\n",
		"Locked.ts":     "export const locked = 2;\n",
	})
	locked := filepath.Join(directory, "Locked.ts")
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0o644) })
	if file, err := os.Open(locked); err == nil {
		file.Close()
		t.Skip("this user can read a mode-000 file, so the fixture cannot make a root unreadable")
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Build panicked on an unreadable root: %v", recovered)
		}
	}()
	graph, err := Build(Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err == nil {
		t.Fatalf("Build returned a graph of %d project files from a config naming 2, one unreadable",
			len(graph.ProjectFiles()))
	}
	if !strings.Contains(err.Error(), "Locked.ts") || !strings.Contains(err.Error(), "could not be read") {
		t.Fatalf("the error does not name the unreadable file and why: %v", err)
	}
}

// A healthy build returns exactly the files the config named, which is the other direction of the
// guard: a verification that refused everything would pass the test above.
func TestBuildReturnsEveryNamedFileWhenNothingMoved(t *testing.T) {
	t.Parallel()
	directory := writeMismatchFixture(t, map[string]string{
		"tsconfig.json": mismatchFixtureConfig,
		"Alpha.ts":      "export const alpha = 1;\n",
		"Beta.ts":       "export const beta = 2;\n",
	})
	graph, err := Build(Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	if named, got := len(graph.Config.FileNames()), len(graph.ProjectFiles()); named != 2 || got != 2 {
		t.Fatalf("named %d and returned %d project files, want 2 and 2", named, got)
	}
}

// Real concurrency, no seam: files created and deleted while Build runs. Whatever the interleaving, a
// build either returns every file its config named or an error, and never panics.
func TestBuildNeverPanicsWhileTheTreeChangesUnderIt(t *testing.T) {
	t.Parallel()
	directory := writeMismatchFixture(t, map[string]string{
		"tsconfig.json": mismatchFixtureConfig,
		"Alpha.ts":      "export const alpha = 1;\n",
	})
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for index := 0; ; index++ {
			select {
			case <-stop:
				return
			default:
			}
			name := filepath.Join(directory, "Churn"+string(rune('A'+index%8))+".ts")
			os.WriteFile(name, []byte("export const churn = 1;\n"), 0o644)
			os.Remove(name)
		}
	}()
	defer func() {
		close(stop)
		<-done
	}()

	for attempt := 0; attempt < 40; attempt++ {
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("attempt %d: Build panicked while the tree changed: %v", attempt, recovered)
				}
			}()
			graph, err := Build(Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
			if err != nil {
				return
			}
			if named, got := len(graph.Config.FileNames()), len(graph.ProjectFiles()); named != got {
				t.Fatalf("attempt %d: a successful build named %d files and returned %d", attempt, named, got)
			}
		}()
	}
}
