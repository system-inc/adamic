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
	for _, pattern := range []string{"ranked*.a", "ranked*.ts", "reference-array-*.a", "reference-array-*.ts"} {
		paths, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane2", pattern))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			// Named compile-refusal fixtures have no executable runtime to count.
			// Their refusal and Node controls remain in the ranked frontier suites.
			switch strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) + ".a" {
			case "ranked3-mutable-array-read.a",
				"ranked6-jsdoc-parent-cache.a",
				"ranked7-related-assign.a", "ranked10-resolved-arguments-assign.a",
				"ranked14-text-name-fallback.a":
				continue
			}
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				t.Fatal(err)
			}
			rows = append(rows, counted(t, checkedViewFixturePath(relative), false, nil, false, false))
		}
	}
	rows = append(rows, ranked15ArrayCounts(t)...)
	rows = append(rows, originalRankedArrayCounts(t, "16")...)
	rows = append(rows, originalRankedArrayCounts(t, "17")...)
	rows = append(rows, originalRankedArrayCounts(t, "18")...)
	rows = append(rows, originalRankedArrayCounts(t, "19")...)
	rows = append(rows, originalRankedArrayCounts(t, "20")...)
	rows = append(rows, originalRankedArrayCounts(t, "21")...)
	rows = append(rows, originalRankedArrayCounts(t, "22")...)
	return append(rows, arraysBCounts(t)...)
}

// Not parallel: updating measured rows writes the shared counts file.
func TestCheckedViewRankedArrayCounts(t *testing.T) {
	rows := viewRankedArrayCounts(t)
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(checkedViewFixturePath(path))
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
