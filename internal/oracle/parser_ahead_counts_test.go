package oracle

import (
	"os"
	"strings"
	"testing"
)

func parserAheadCounts(t *testing.T) []string {
	t.Helper()
	rows := []string{}
	for _, name := range []string{"already-complete", "complete", "escaped"} {
		rows = append(rows, counted(t, "stage3/parser-ahead/factories/"+name+".a", false, nil, false, false))
	}
	for _, name := range []string{"factory-unset", "factory-use", "own-key-order", "node-array-keys", "node-array-json", "node-array-length", "node-array-unset"} {
		rows = append(rows, counted(t, "stage3/parser-ahead/rulings/"+name+".a", false, nil, false, false))
	}
	rows = append(rows, counted(t, "stage3/parser-ahead/node-array/fields.a", false, nil, false, false))
	rows = append(rows, counted(t, "stage3/parser-ahead/rulings/null-placeholder-pending.a", false, nil, false, false))
	return rows
}

func init() { additionalFixtureCounts = append(additionalFixtureCounts, parserAheadCounts) }

// Update only owned rows when diagnosing a global recording failure.
// Complete and escaped now count executed staged factory programs.
func TestParserAheadCounts(t *testing.T) {
	rows := parserAheadCounts(t)
	b, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	for _, row := range rows {
		key := strings.Split(row, " | ")[0] + " | "
		found := false
		for i, line := range lines {
			if strings.HasPrefix(line, key) {
				found = true
				if *updateCounts {
					lines[i] = row
				} else if line != row {
					t.Fatalf("recorded %s, measured %s", line, row)
				}
			}
		}
		if !found {
			if !*updateCounts {
				t.Fatalf("missing counts %s", row)
			}
			lines = append(lines, row)
		}
	}
	if *updateCounts {
		if err := os.WriteFile(countsPath, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
