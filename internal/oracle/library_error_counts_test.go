package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Refresh this unit and existing Error constructors affected by its ownership,
// without running unrelated fixture families. Preserve their measured rows.
func TestErrorCountsAreRecorded(t *testing.T) {
	content, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string]string{}
	key := func(row string) string { return strings.TrimPrefix(strings.Split(row, " | ")[0], "| ") }
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "| ") && len(strings.Split(line, " | ")) == 7 && strings.Contains(key(line), "/") {
			rows[key(line)] = line
		}
	}
	for _, fixture := range fixtures {
		if !fixture.lowers || uncounted[fixture.path] {
			continue
		}
		source, err := os.ReadFile(filepath.Join(repository, fixture.path))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(fixture.path, "library_error_") && !strings.Contains(string(source), "new Error(") {
			continue
		}
		t.Run(fixture.path, func(t *testing.T) {
			row := counted(t, fixture.path, false, nil, false, false)
			if rows[fixture.path] != row && !*updateCounts {
				t.Errorf("Error counts moved:\nrecorded: %s\nmeasured: %s", rows[fixture.path], row)
			}
			rows[fixture.path] = row
		})
	}
	if t.Failed() || !*updateCounts {
		return
	}
	var table strings.Builder
	table.WriteString(countsHeader)
	for _, fixture := range append(fixtures, slowRegExpFixtures...) {
		if row, ok := rows[fixture.path]; ok {
			table.WriteString(row + "\n")
			delete(rows, fixture.path)
		}
	}
	// Input/file fixtures retain their existing order at the end of the table.
	for _, line := range strings.Split(string(content), "\n") {
		if row, ok := rows[key(line)]; ok {
			table.WriteString(row + "\n")
			delete(rows, key(line))
		}
	}
	if len(rows) != 0 {
		t.Fatalf("unplaced count rows: %v", rows)
	}
	if err := os.WriteFile(countsPath, []byte(table.String()), 0644); err != nil {
		t.Fatal(err)
	}
}
