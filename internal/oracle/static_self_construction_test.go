package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A class constructing itself in its own static initializer lowers when the construction, its
// parameter defaults, instance field initializers, constructor body and everything they reach,
// reads no static of that class initialized later in source order (#1451npc, Adamic's ruling, the
// future-read entry of docs/rulings.md). Each agrees with Node natively, sanitized and on JavaScript,
// tsc's real Version.zero (tsc_version_zero.ts) among them: 0.0.0-0, 1.2.3-beta.1+build.5, 2.0.0,
// then the Debug Failure its constructor throws for a negative major. The stops are pinned in
// internal/lower's TestStaticSelfConstructionStops, TestStaticWalkReachAdmissions and
// TestPrepareStackTraceStaysStopped. Mutants: internal/lower/static-self-construction.mutants.md.
func TestStaticSelfConstructionAgreesWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/static_self_construction"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") && !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, directory+"/"+entry.Name())
		})
	}
	if programs != 2 {
		t.Fatalf("%d programs in %s, want 2:a program added or removed needs this count changed with it", programs, directory)
	}
}
