package lint

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// Diagnostic IDs are observable protocol fields. Each replacement is one
// literal token, scoped to the private rule entry, never an owned checkout file.
func applySemanticEdit(t *testing.T, directory string, d registry.Descriptor) {
	edits := map[string][2]string{
		"no-var":   {"'unexpectedVar'", "'unexpectedVaz'"},
		"no-empty": {"'unexpectedBlock'", "'unexpectedBlocx'"},
		"eqeqeq":   {"'unexpected'", "'unexpectee'"},
	}
	edit, ok := edits[d.Slug]
	if !ok {
		t.Fatalf("no semantic benchmark edit for %s", d.Slug)
	}
	lintChange(t, filepath.Join(directory, "rules", d.Slug, d.Module), edit[0], edit[1])
	t.Logf("semantic token edit %s: %s -> %s", d.Slug, edit[0], edit[1])
}

func semanticComparison(t *testing.T, oracle, binary, directory, path, module string) {
	want := execute(t, "", oracle, "--manifest", path).output
	source := node(t, directory, path, false).output
	javascript := runJavaScript(t, module, path, false).output
	native := execute(t, "", binary, "--manifest", path).output
	if !bytes.Equal(source, native) {
		t.Fatal("semantic edit Node/native outputs differ")
	}
	if !bytes.Equal(source, javascript) {
		t.Fatal("semantic edit Node/JavaScript outputs differ")
	}
	if bytes.Equal(source, want) {
		t.Fatal("semantic edit did not change findings")
	}
	t.Logf("semantic edit Node JavaScript native byte-identical: %d bytes; all disagree with Go: %s", len(source), difference(source, want))
	// A wrong ID must fail certification, even though all port runtimes agree.
	t.Fatal("EXPECTED SEMANTIC EDIT FAILURE: Node JavaScript native agree and disagree with Go")
}
