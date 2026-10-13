package fresh_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFreshPrototypeLinks(t *testing.T) {
	t.Parallel()
	beginFreshUnit(t)
	source, err := os.ReadFile("../oracle/testdata/scout22_prototype_refused/scout22_prototype_links.a")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "main.ts")
	if err = os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	checkFreshProgram(t, path)
}
func TestFreshPrototypeLifetimes(t *testing.T) {
	t.Parallel()
	beginFreshUnit(t)
	source, err := os.ReadFile("../oracle/testdata/scout22_prototype_refused/scout22_prototype_lifetimes.a")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "main.ts")
	if err = os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	checkFreshProgram(t, path)
}
