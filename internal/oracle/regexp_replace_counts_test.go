package oracle

import (
	"os"
	"strings"
	"testing"
)

// Measure only this slice's restored fixtures, avoiding unrelated oracle work.
// Not parallel: can update the shared counts.md file when ADAMIC_UPDATE_COUNTS is set.
func TestRegExpReplacementRestoredCounts(t *testing.T) {
	for _, name := range []string{"effects", "move_effect"} {
		path := "internal/oracle/testdata/regexp_replace/" + name + ".a"
		row := counted(t, path, false, nil, false, false)
		t.Log(row)
		data, err := os.ReadFile(countsPath)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if strings.Contains(text, row+"\n") {
			continue
		}
		if !*updateCounts {
			t.Fatalf("restored fixture counts missing or changed: %s", row)
		}
		if strings.Contains(text, "| "+path+" |") {
			t.Fatal("existing row changed; review it explicitly")
		}
		at := strings.Index(text, "| internal/oracle/testdata/regexp_replace/arguments.a |")
		if at < 0 {
			t.Fatal("regexp fixture boundary missing")
		}
		at += strings.Index(text[at:], "\n") + 1
		if name == "move_effect" {
			if end := strings.Index(text[at:], "| internal/oracle/testdata/regexp_replace/effects.a |"); end == 0 {
				at += strings.Index(text[at:], "\n") + 1
			}
		}
		if err := os.WriteFile(countsPath, []byte(text[:at]+row+"\n"+text[at:]), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
