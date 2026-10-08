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
	return rows
}

func init() { additionalFixtureCounts = append(additionalFixtureCounts, parserAheadCounts) }

// Only these three rows are updated if inherited fixtures prevent global recording.
// The complete/escaped rows count boundary stops, not successful factory lowering.
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
