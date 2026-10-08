package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Original declarations are opt-in, as in the semantic oracle. Without them,
// retain the measured rows; with them, bind and remeasure every tuple fixture.
func tupleOriginalCounts(t *testing.T) []string {
	t.Helper()
	const prefix = "| stage3/interface-downcasts/tuples/"
	if os.Getenv("ADAMIC_TUPLE_ORIGINAL_DECLS") == "" {
		data, err := os.ReadFile(countsPath)
		if err != nil {
			t.Fatal(err)
		}
		var rows []string
		for _, row := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(row, prefix) {
				rows = append(rows, row)
			}
		}
		return rows
	}
	root, _ := tupleOriginalInputs(t)
	paths, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/tuples", "*.a"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []string
	absoluteRepository, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range paths {
		name := strings.TrimSuffix(filepath.Base(source), ".a")
		bound, _ := tupleOriginalProgram(t, root, name)
		relative, err := filepath.Rel(absoluteRepository, bound)
		if err != nil {
			t.Fatal(err)
		}
		row := counted(t, relative, false, nil, false, false)
		rows = append(rows, strings.Replace(row, "| "+relative+" |", prefix+name+".a |", 1))
	}
	return rows
}

func TestCheckedViewTupleOriginalCounts(t *testing.T) {
	if os.Getenv("ADAMIC_TUPLE_ORIGINAL_DECLS") == "" {
		t.Skip("original tuple declarations required to measure counts")
	}
	rows := tupleOriginalCounts(t)
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !*updateCounts {
		for _, row := range rows {
			if !strings.Contains(string(data), row+"\n") {
				t.Errorf("unrecorded tuple counts: %s", row)
			}
		}
		return
	}
	parts := strings.SplitN(string(data), "\n## Predicate direction counts", 2)
	var lines []string
	for _, line := range strings.Split(strings.TrimSuffix(parts[0], "\n"), "\n") {
		if !strings.HasPrefix(line, "| stage3/interface-downcasts/tuples/") {
			lines = append(lines, line)
		}
	}
	lines = append(lines, rows...)
	updated := strings.Join(lines, "\n") + "\n"
	if len(parts) == 2 {
		updated += "\n## Predicate direction counts" + parts[1]
	}
	if err := os.WriteFile(countsPath, []byte(updated), 0644); err != nil {
		t.Fatal(err)
	}
}
