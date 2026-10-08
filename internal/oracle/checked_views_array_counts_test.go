package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func viewRankedArrayCounts(t *testing.T) []string {
	t.Helper()
	var rows []string
	for _, pattern := range []string{"ranked*.a", "reference-array-*.a"} {
		paths, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane2", pattern))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			// This fixture deliberately refuses lowering, so it has no runtime counts.
			if filepath.Base(path) == "ranked3-mutable-array-read.a" {
				continue
			}
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				t.Fatal(err)
			}
			rows = append(rows, counted(t, relative, false, nil, false, false))
		}
	}
	return append(rows, ranked15ArrayCounts(t)...)
}

// Not parallel: updating measured rows writes the shared counts file.
func TestCheckedViewRankedArrayCounts(t *testing.T) {
	rows := viewRankedArrayCounts(t)
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if *updateCounts {
		parts := strings.SplitN(string(data), "\n## Predicate direction counts", 2)
		lines := strings.Split(strings.TrimSuffix(parts[0], "\n"), "\n")
		for _, row := range rows {
			key := strings.Split(row, " | ")[0] + " | "
			found := false
			for i, line := range lines {
				if strings.HasPrefix(line, key) {
					lines[i] = row
					found = true
					break
				}
			}
			if !found {
				lines = append(lines, row)
			}
		}
		updated := strings.Join(lines, "\n") + "\n"
		if len(parts) == 2 {
			updated += "\n## Predicate direction counts" + parts[1]
		}
		if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	for _, row := range rows {
		if !strings.Contains(string(data), row+"\n") {
			t.Errorf("unrecorded ranked array counts: %s", row)
		}
	}
}
