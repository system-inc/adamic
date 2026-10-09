package release

import (
	"strings"
	"testing"
)

// The local-build note has to be true in both directions. A committed-tree build names a commit and
// has nothing uncommitted in it, so it is reproducible and must say so; telling that reader its rules
// are "whatever was on disk" is a disclaimer that has stopped being true. A dirty-tree build is the
// case the disclaimer exists for, and must keep it.
func TestTheLocalBuildNoteSaysWhetherACommitReproducesIt(t *testing.T) {
	t.Parallel()

	committed := Provenance{Version: "dev", SelfCommit: "7b434bb4105bb12698ec04d4ea5b932d37cc77a5"}.String()
	if !strings.Contains(committed, "so that commit reproduces it") {
		t.Fatalf("a clean build of a named commit does not say that commit reproduces it:\n%s", committed)
	}
	if strings.Contains(committed, "whatever was on disk") {
		t.Fatalf("a clean build of a named commit still claims its rules are whatever was on disk:\n%s", committed)
	}

	dirty := Provenance{Version: "dev", SelfCommit: "7b434bb4105bb12698ec04d4ea5b932d37cc77a5", SourceTreeModified: true}.String()
	if !strings.Contains(dirty, "whatever was on disk") {
		t.Fatalf("a build from a dirty tree lost its disclaimer:\n%s", dirty)
	}

	// No commit at all cannot be reproduced from one, however clean the tree was.
	unnamed := Provenance{Version: "dev"}.String()
	if !strings.Contains(unnamed, "whatever was on disk") {
		t.Fatalf("a build naming no commit claims something reproduces it:\n%s", unnamed)
	}
}
