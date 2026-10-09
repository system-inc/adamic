package main

import (
	"strings"
	"testing"

	release "github.com/system-inc/cohere/internal/release/packaging"
)

// The note has to be true in both directions, like `-version`'s: a clean build of a named commit says
// that commit reproduces the list, and a build from a dirty tree, or naming no commit, keeps the
// disclaimer. Either way the count travels with it.
func TestRulesProvenanceNoteSaysWhetherACommitReproducesTheList(t *testing.T) {
	t.Parallel()

	clean := rulesProvenanceNote(release.Provenance{Version: "dev", SelfCommit: "7b434bb4105b"}, 464)
	if !strings.Contains(clean, "7b434bb4105b") || !strings.Contains(clean, "reproduces them") {
		t.Fatalf("a clean build of a named commit does not name it as reproducing the list: %q", clean)
	}
	if strings.Contains(clean, "whatever was on disk") {
		t.Fatalf("a clean build of a named commit still claims its rules are whatever was on disk: %q", clean)
	}

	dirty := rulesProvenanceNote(release.Provenance{Version: "dev", SelfCommit: "7b434bb4105b", SourceTreeModified: true}, 464)
	if !strings.Contains(dirty, "not necessarily what is committed") {
		t.Fatalf("a build from a dirty tree lost its disclaimer: %q", dirty)
	}

	unnamed := rulesProvenanceNote(release.Provenance{Version: "dev"}, 464)
	if !strings.Contains(unnamed, "not necessarily what is committed") {
		t.Fatalf("a build naming no commit claims something reproduces it: %q", unnamed)
	}

	for _, note := range []string{clean, dirty, unnamed} {
		if !strings.Contains(note, "464 rules") {
			t.Fatalf("the note does not carry the rule count it qualifies: %q", note)
		}
	}
}
