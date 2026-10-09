package vendored

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The snapshot is the version it says, and holds every file the loader reads, so a refresh that drops one
// fails here rather than as a design system that quietly builds from less.
func TestTheSnapshotIsWholeAndTheVersionItSays(t *testing.T) {
	t.Parallel()
	root := TailwindPackageRoot()
	for _, name := range []string{"index.css", "theme.css", "preflight.css", "utilities.css", "LICENSE", "package.json"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Errorf("the snapshot is missing %s: %v", name, err)
		}
	}
	contents, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		License string `json:"license"`
	}
	if err := json.Unmarshal(contents, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "tailwindcss" || manifest.Version != TailwindVersion || manifest.License != "MIT" {
		t.Errorf("the snapshot is %s@%s (%s), want tailwindcss@%s (MIT)", manifest.Name, manifest.Version,
			manifest.License, TailwindVersion)
	}
}
