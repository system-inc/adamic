package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCountDigestIgnoresSourceDirectory(t *testing.T) {
	t.Parallel()
	const key = "internal/oracle/testdata/non_null_null.ts"
	source, err := os.ReadFile(filepath.Join(repository, key))
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	digest := func(directory string) string {
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(directory, "witness.ts")
		if err := os.WriteFile(path, source, 0644); err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		runs, _ := counted(t, relative, false, nil, false, false)
		return countDigest(t, key, runs)
	}
	directory := t.TempDir()
	first := digest(filepath.Join(directory, "short"))
	second := digest(filepath.Join(directory, "a-much-longer-source-directory"))
	if first != second {
		t.Fatalf("source directory changed count digest: %s != %s", first, second)
	}
}
