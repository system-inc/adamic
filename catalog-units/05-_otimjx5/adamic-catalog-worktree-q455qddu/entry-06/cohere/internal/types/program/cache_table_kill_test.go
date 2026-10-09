package program

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// TestAWriterKilledBeforeItsRenameLeavesTheOldTable stops a writer after it has written a temporary and
// before it renames it into place, kills it there, and reads the table: every file whole, the old one or the
// new one, never half of either. A run that returns to its caller early writes its table after the caller
// has moved on, where nothing stops a kill (#zqsdzbq, early return).
//
// The writer replaces two files, a run and the findings, and is killed at the first rename and, separately,
// at the second. Before the first, both files are the old ones. Between them, the run is the new one and the
// findings the old: the files are not atomic with each other, and need not be, since each carries its own
// proof (#45ekc65).
//
// The writer says where it stopped: at the held rename it writes the temporary's name to a marker, whole, by
// rename, and then blocks, and the test kills it only once that marker exists. Nothing is timed. The test used to
// kill 200ms after the first temporary appeared, which under load could land after the writer had already renamed
// the run's temporary away (#dnmvs98).
func TestAWriterKilledBeforeItsRenameLeavesTheOldTable(t *testing.T) {
	t.Parallel()
	identity := CacheTableIdentity{SelfCommit: "kill-test"}
	sections := CacheTableSections{Runs: []string{"--no-fix"}, Findings: true}
	tableSaying := func(text string) *CacheTable {
		table := NewCacheTable()
		table.Runs["--no-fix"] = &RunCache{Version: runCacheVersion, Key: text, Output: []byte(text)}
		table.Findings = &LintCache{Version: lintCacheVersion, Entries: []LintCacheEntry{{Path: "/" + text + ".ts"}}}
		return table
	}
	if directory := os.Getenv("COHERE_TEST_WRITE_TABLE"); directory != "" {
		// The writer: the new table, held just before its nth rename until it is killed.
		stopAt, _ := strconv.Atoi(os.Getenv("COHERE_TEST_STOP_AT_RENAME"))
		marker := os.Getenv("COHERE_TEST_HELD_MARKER")
		renames := 0
		beforeCacheTableRename = func(temporaryName string) {
			renames++
			if renames != stopAt {
				return
			}
			// Written beside the marker and renamed onto it, so the test never reads half a name.
			if err := os.WriteFile(marker+".partial", []byte(temporaryName), 0o644); err == nil {
				os.Rename(marker+".partial", marker)
			}
			time.Sleep(time.Minute)
		}
		WriteCacheTable(directory, tableSaying("new"), identity, sections, PathAnchor{})
		return
	}

	for _, stopAt := range []int{1, 2} {
		t.Run(fmt.Sprintf("killed at rename %d", stopAt), func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			// Outside the table's directory, which the write after the kill cleans.
			marker := filepath.Join(t.TempDir(), "held")
			if err := WriteCacheTable(directory, tableSaying("old"), identity, sections, PathAnchor{}); err != nil {
				t.Fatal(err)
			}

			writer := exec.Command(os.Args[0], "-test.run=^TestAWriterKilledBeforeItsRenameLeavesTheOldTable$")
			writer.Env = append(os.Environ(), "COHERE_TEST_WRITE_TABLE="+directory, fmt.Sprintf("COHERE_TEST_STOP_AT_RENAME=%d", stopAt),
				"COHERE_TEST_HELD_MARKER="+marker)
			if err := writer.Start(); err != nil {
				t.Fatal(err)
			}
			// Killed once the writer says it is held at the rename, naming the temporary it holds. The deadline
			// only ends a writer that never got there; it never decides when the kill lands.
			deadline := time.Now().Add(time.Minute)
			temporary := ""
			for temporary == "" && time.Now().Before(deadline) {
				if held, err := os.ReadFile(marker); err == nil {
					temporary = string(held)
				}
				time.Sleep(10 * time.Millisecond)
			}
			if temporary == "" {
				writer.Process.Kill()
				writer.Wait()
				t.Fatal("the writer never reached its held rename, so the kill tests nothing")
			}
			writer.Process.Signal(syscall.SIGKILL)
			writer.Wait()

			read, err := ReadCacheTable(directory, identity, sections, PathAnchor{})
			if err != nil {
				t.Fatalf("the table is unreadable after the writer died before a rename: %v", err)
			}
			wantRun, wantFindings := "old", "/old.ts"
			if stopAt == 2 {
				wantRun = "new"
			}
			if run := read.Runs["--no-fix"]; run == nil || string(run.Output) != wantRun {
				t.Errorf("the run after the kill is %v, want the %s one", run, wantRun)
			}
			if read.Findings == nil || len(read.Findings.Entries) != 1 || read.Findings.Entries[0].Path != wantFindings {
				t.Errorf("the findings after the kill are %v, want the old ones", read.Findings)
			}
			if _, err := os.Stat(temporary); err != nil {
				t.Errorf("the temporary is gone, so the writer was not stopped where this test meant: %v", err)
			}

			// And the next run rebuilds cleanly over what the killed one left, its temporary included.
			if err := WriteCacheTable(directory, tableSaying("rebuilt"), identity, sections, PathAnchor{}); err != nil {
				t.Fatalf("the write after the kill failed: %v", err)
			}
			rebuilt, err := ReadCacheTable(directory, identity, sections, PathAnchor{})
			if err != nil || string(rebuilt.Runs["--no-fix"].Output) != "rebuilt" || rebuilt.Findings.Entries[0].Path != "/rebuilt.ts" {
				t.Errorf("the table after the rebuild is not the rebuilt one: %v", err)
			}
		})
	}
}
