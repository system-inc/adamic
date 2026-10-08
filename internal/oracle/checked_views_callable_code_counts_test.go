package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Refresh only the new code group's measured rows when inherited fixtures fail.
func TestCheckedViewCallableCodeCounts(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane5/code/*/*/*.a"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("fixtures: %v, %v", paths, err)
	}
	domains, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane5/code/set-domains/*.a"))
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, domains...)
	path := filepath.Join(repository, "internal/oracle/counts.md")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(contents), "\n"), "\n")
	for _, fixture := range paths {
		relative, err := filepath.Rel(repository, fixture)
		if err != nil {
			t.Fatal(err)
		}
		row := counted(t, relative, false, nil, false, false)
		if !*updateCounts {
			if !strings.Contains(string(contents), row+"\n") {
				t.Errorf("unrecorded callable code counts: %s", row)
			}
			continue
		}
		key := strings.Split(row, " | ")[0] + " | "
		replaced := false
		for i, line := range lines {
			if strings.HasPrefix(line, key) {
				lines[i], replaced = row, true
				break
			}
		}
		if !replaced {
			lines = append(lines, row)
		}
	}
	if *updateCounts {
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
