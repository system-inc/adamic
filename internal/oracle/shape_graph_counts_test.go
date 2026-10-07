package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: this snapshot compares the entire graph fixture family against
// the pre-extraction dependency merge, including the million-object control.
func TestShapeGraphCountSnapshot(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repository, "stage3/shape-conformance/graph-counts-guard.json"))
	if err != nil {
		t.Fatal(err)
	}
	var baseline struct {
		Rows []string `json:"rows"`
	}
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{}
	for _, row := range baseline.Rows {
		expected[strings.TrimSpace(strings.Split(row, "|")[1])] = row
	}
	for _, pattern := range []string{"internal/oracle/testdata/graph_regions*.a", "internal/oracle/testdata/graph_regions/*.a"} {
		paths, err := filepath.Glob(filepath.Join(repository, pattern))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				t.Fatal(err)
			}
			t.Run(relative, func(t *testing.T) {
				row := counted(t, relative, false, nil, false, false)
				t.Log(row)
				if row != expected[relative] {
					t.Fatalf("graph-region counts changed from dependency merge 40ad020d: wanted %s", expected[relative])
				}
			})
		}
	}
}
