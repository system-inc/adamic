package native

import (
	"os"
	"path/filepath"
	"testing"
)

// Not parallel: verifies cache configuration through process environment.
func TestArtifactInputsAndReuse(t *testing.T) {
	cache, err := newTestBuildCache("../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	t.Setenv("ADAMIC_BUILD_CACHE", "")
	builds := 0
	build := func(directory string) error {
		builds++
		return os.WriteFile(filepath.Join(directory, "binary"), []byte("built"), 0644)
	}
	first, err := cache.Tree("fixture", [][]byte{[]byte("one")}, build)
	if err != nil {
		t.Fatal(err)
	}
	same, err := cache.Tree("fixture", [][]byte{[]byte("one")}, build)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := cache.Tree("fixture", [][]byte{[]byte("two")}, build)
	if err != nil {
		t.Fatal(err)
	}
	if first != same || first == changed || builds != 2 {
		t.Fatalf("artifact inputs/reuse: first=%s same=%s changed=%s builds=%d", first, same, changed, builds)
	}
}
