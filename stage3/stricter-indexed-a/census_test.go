package indexedwitness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAssignedLedgerIsCovered(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("ledger.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		ID     string `json:"id"`
		File   string `json:"file"`
		Option string `json:"option"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 27 {
		t.Fatalf("assigned census changed: %d", len(rows))
	}
	seen := map[string]string{}
	paths, err := filepath.Glob("D*.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var fixture witness
		if err := json.Unmarshal(data, &fixture); err != nil {
			t.Fatal(err)
		}
		if fixture.Source == "" || fixture.Read == "" || fixture.Receiver == "" {
			t.Fatalf("incomplete witness %s", path)
		}
		for _, id := range fixture.IDs {
			if previous := seen[id]; previous != "" {
				t.Fatalf("%s duplicated by %s and %s", id, previous, path)
			}
			seen[id] = path
		}
	}
	for _, row := range rows {
		if row.File != "src/compiler/checker.ts" || row.Option != "noUncheckedIndexedAccess" {
			t.Fatalf("outside assigned territory: %+v", row)
		}
		if seen[row.ID] == "" {
			t.Errorf("missing assigned row %s", row.ID)
		}
		delete(seen, row.ID)
	}
	if len(seen) != 0 {
		t.Fatalf("extra rows: %v", seen)
	}
}
