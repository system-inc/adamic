package program

import (
	"os"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// RecordedRealSignatures is a previous table in which every project source file has a real signature recorded
// for other bytes, so Signatures emits each one, as it does a file edited since its signature was computed.
// For tests about what a computed signature holds: a first run with nothing recorded computes none (#5txm9gg).
func RecordedRealSignatures(g *Graph) map[string]SignatureEntry {
	recorded := map[string]SignatureEntry{}
	for _, sourceFile := range g.ProjectFiles() {
		if sourceFile.IsDeclarationFile || ast.IsJsonSourceFile(sourceFile) {
			continue
		}
		recorded[sourceFile.FileName().AsString()] = SignatureEntry{Version: "recorded for other bytes", Signature: "a real signature, recorded for other bytes"}
	}
	return recorded
}

// SignatureEmits is how many files Signatures has emitted declarations for on this graph.
func (g *Graph) SignatureEmits() int {
	return g.signatureEmits
}

// StaleStatSnapshotForTest is a snapshot that answers each of recorded's inputs as it was recorded rather than as
// the disk is now: absent where the recording found nothing, whatever has been created there since. It is the
// wrong snapshot a check that kept the recording's answers would hand a build, for a test to prove that its
// comparison would catch one.
func StaleStatSnapshotForTest(recorded *RunCache) *StatSnapshot {
	snapshot := NewStatSnapshot()
	for _, input := range recorded.Inputs {
		if !input.Exists {
			snapshot.answers.Store(input.Path, absentAnswer{})
			continue
		}
		information, err := os.Stat(input.Path)
		snapshot.note(input.Path, information, err)
	}
	return snapshot
}
