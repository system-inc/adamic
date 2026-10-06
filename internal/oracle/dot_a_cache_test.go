package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSharedLoaderInvalidatesOracleEvidence(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	directory := filepath.Join(root, "oracle")
	if err := os.Mkdir(directory, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"node.mjs", "register-dot-a.mjs", "adamic.mjs"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("original"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	before, err := oracleRunnerIdentity(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "register-dot-a.mjs"), []byte("changed loader"), 0644); err != nil {
		t.Fatal(err)
	}
	after, err := oracleRunnerIdentity(root)
	if err != nil {
		t.Fatal(err)
	}
	if before == after {
		t.Fatal("shared loader changed without invalidating oracle evidence")
	}
}
