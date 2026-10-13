package flow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFlowPrototypeLinks(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	source, err := os.ReadFile("../oracle/testdata/scout22_prototype_refused/scout22_prototype_links.a")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "main.ts")
	if err = os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	checkAllFlowProgram(t, path)
}
func TestFlowPrototypeLifetimes(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	source, err := os.ReadFile("../oracle/testdata/scout22_prototype_refused/scout22_prototype_lifetimes.a")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "main.ts")
	if err = os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	checkAllFlowProgram(t, path)
}
