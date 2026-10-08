package oracle

import (
	"os"
	"strings"
	"testing"
)

// Measure only the Buffer/hash unit on Linux. Preserve
// all other recorded rows while ordering them by the live fixture registry.
// Not parallel: -update-counts writes the shared counts.md table.
func TestNodeBufferHost29Counts(t *testing.T) {
	contents, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string]string{}
	for _, line := range strings.Split(string(contents), "\n") {
		if !strings.HasPrefix(line, "| ") || !strings.Contains(strings.Split(line, " | ")[0], "/") {
			continue
		}
		fields := strings.Split(line, " | ")
		key := strings.TrimPrefix(fields[0], "| ")
		if _, exists := rows[key]; exists {
			t.Fatalf("duplicate row: %s", key)
		}
		rows[key] = line
	}
	for _, fixture := range fixtures {
		if !fixture.lowers {
			continue
		}
		if !strings.HasPrefix(fixture.path, "internal/oracle/testdata/node_buffer_") {
			continue
		}

		row := counted(t, fixture.path, false, nil, false, false)
		if !*updateCounts && rows[fixture.path] != row {
			t.Errorf("recorded: %s\nmeasured: %s", rows[fixture.path], row)
		}
		rows[fixture.path] = row
	}
	if !*updateCounts || t.Failed() {
		return
	}
	var table strings.Builder
	table.WriteString(countsHeader)
	appendRow := func(path string) {
		if row, exists := rows[path]; exists {
			table.WriteString(row + "\n")
			delete(rows, path)
		}
	}
	for _, fixture := range fixtures {
		appendRow(fixture.path)
	}
	for _, fixture := range slowRegExpFixtures {
		appendRow(fixture.path)
	}
	for _, fixture := range inputFixtures {
		appendRow(fixture.path)
	}
	for _, fixture := range fsFileFixtures {
		appendRow("internal/oracle/testdata/node_fs_file_" + fixture + ".a")
	}
	if len(rows) != 0 {
		t.Fatalf("unregistered count rows remain: %v", rows)
	}
	if err := os.WriteFile(countsPath, []byte(table.String()), 0644); err != nil {
		t.Fatal(err)
	}
}
