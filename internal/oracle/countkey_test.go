package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCountDigestIgnoresSourceDirectory(test *testing.T) {
	test.Parallel()
	const key = "internal/oracle/testdata/non_null_null.ts"
	source, readError := os.ReadFile(filepath.Join(repository, key))
	if readError != nil {
		test.Fatal(readError)
	}
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	digest := func(directory string) string {
		if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		path := filepath.Join(directory, "witness.ts")
		if writeError := os.WriteFile(path, source, 0644); writeError != nil {
			test.Fatal(writeError)
		}
		relative, relativePathError := filepath.Rel(root, path)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		runs, _ := counted(test, relative, false, nil, false, false)
		return countDigest(test, key, runs)
	}
	directory := test.TempDir()
	first := digest(filepath.Join(directory, "short"))
	second := digest(filepath.Join(directory, "a-much-longer-source-directory"))
	if first != second {
		test.Fatalf("source directory changed count digest: %s != %s", first, second)
	}
}
