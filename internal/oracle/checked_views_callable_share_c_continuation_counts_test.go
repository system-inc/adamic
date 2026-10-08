package oracle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewCallableShareCContinuationCounts(t *testing.T) {
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	manifest, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-c/continuation-candidates.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []callableShareCRow
	if err = json.Unmarshal(manifest, &rows); err != nil {
		t.Fatal(err)
	}
	maps := callableShareCMapRows(t)
	mapRanks := map[int]bool{}
	for _, row := range maps {
		mapRanks[row.Rank] = true
	}
	rows = append(rows, maps...)
	for _, row := range rows {
		variants := []string{"good", "wrong-element"}
		if mapRanks[row.Rank] {
			variants = []string{"good", "wrong-map"}
		}
		for _, variant := range variants {
			p := fmt.Sprintf("stage3/interface-downcasts/lane5/share-c/families/rank-%d/%s.a", row.Rank, variant)
			measured := counted(t, checkedViewFixturePath(p), false, nil, false, false)
			if !*updateCounts {
				if !strings.Contains(string(data), measured+"\n") {
					t.Errorf("unrecorded counts: %s", measured)
				}
				continue
			}
			key := strings.Split(measured, " | ")[0] + " | "
			found := false
			for i, line := range lines {
				if strings.HasPrefix(line, key) {
					lines[i] = measured
					found = true
					break
				}
			}
			if !found {
				lines = append(lines, measured)
			}
		}
	}
	if *updateCounts {
		if err = os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
