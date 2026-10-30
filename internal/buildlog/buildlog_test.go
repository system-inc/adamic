package buildlog

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// A phase's line is buildLog's five fields: phase, its name, no key, off, and the seconds since it started to three
// places. An unset log writes nothing, and one that can't be opened fails nothing. Mutants: the key left out (four
// fields); the seconds taken from now rather than since started.
//
// Not parallel: t.Setenv changes ADAMIC_BUILD_LOG for the process.
func TestPhaseWritesOneBuildLogLine(t *testing.T) {
	log := filepath.Join(t.TempDir(), "build.log")
	t.Setenv("ADAMIC_BUILD_LOG", log)
	Phase("lower", time.Now().Add(-1500*time.Millisecond))
	t.Setenv("ADAMIC_BUILD_LOG", "")
	Phase("emit", time.Now())
	// A log under a file is a path no one can open.
	t.Setenv("ADAMIC_BUILD_LOG", filepath.Join(log, "beneath"))
	Phase("emit", time.Now())
	written, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^phase lower - off 1\.5[0-9]{2}\n$`).Match(written) {
		t.Fatalf("wrote %q, want one line: phase lower - off 1.5xx", written)
	}
}
