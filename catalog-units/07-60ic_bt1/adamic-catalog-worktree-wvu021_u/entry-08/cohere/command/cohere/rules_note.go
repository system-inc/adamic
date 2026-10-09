package main

import (
	"fmt"

	release "github.com/system-inc/cohere/internal/release/packaging"
)

// rulesProvenanceNote is the line `-rules` prints to stderr beside a development build's rule list.
//
// The global cohere builds from the checkout's committed tree, so most development builds now name
// a commit with nothing uncommitted in them, and that commit reproduces the list exactly. Telling
// that reader the rules are "whatever was on disk" would be a disclaimer that has stopped being true,
// and a reader who learns one disclaimer is decorative stops reading the ones that are not. A build
// from a dirty tree, or one that names no commit, is the case the disclaimer exists for and keeps it.
// The rule is the same one `-version` applies, so the two surfaces never disagree about one binary.
func rulesProvenanceNote(provenance release.Provenance, ruleCount int) string {
	if provenance.SelfCommit != "" && !provenance.SourceTreeModified {
		return fmt.Sprintf(
			"note: %d rules from a local build of %s, with nothing uncommitted in it, so that commit reproduces them",
			ruleCount, provenance.SelfCommit,
		)
	}
	return fmt.Sprintf(
		"note: %d rules from a local build, so this is whatever was on disk when it was compiled, not necessarily what is committed",
		ruleCount,
	)
}
