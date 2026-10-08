package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Executable controls added by dictionary groups 9 through 11. Reached array-entry
// tuple consumers and eighteen uncredited frontier probes remain outside counts.
var dictionaryIntegrationFixtures = []string{
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-array-unread-empty.a",
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-array-unread-good.a",
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-array-unread-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-number-empty.a",
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-number-good.a",
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-number-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-object-empty.a",
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-object-good.a",
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-object-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-string-empty.a",
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-string-good.a",
	"stage3/interface-downcasts/dictionaries/source/entries-fixed-string-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-array-unread-empty.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-array-unread-good.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-array-unread-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-number-empty.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-number-good.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-number-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-object-empty.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-object-good.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-object-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-string-empty.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-string-good.a",
	"stage3/interface-downcasts/dictionaries/source/entries-producer-string-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/keys-fixed-empty.a",
	"stage3/interface-downcasts/dictionaries/source/keys-fixed-good.a",
	"stage3/interface-downcasts/dictionaries/source/keys-fixed-read-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/keys-producer-empty.a",
	"stage3/interface-downcasts/dictionaries/source/keys-producer-good.a",
	"stage3/interface-downcasts/dictionaries/source/keys-producer-read-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/path-index-good.a",
	"stage3/interface-downcasts/dictionaries/source/path-index-missing.a",
	"stage3/interface-downcasts/dictionaries/source/path-index-negative.a",
	"stage3/interface-downcasts/dictionaries/source/path-index-surrogate.a",
	"stage3/interface-downcasts/dictionaries/source/path-index-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/string-index-good.a",
	"stage3/interface-downcasts/dictionaries/source/string-index-missing.a",
	"stage3/interface-downcasts/dictionaries/source/string-index-negative.a",
	"stage3/interface-downcasts/dictionaries/source/string-index-surrogate.a",
	"stage3/interface-downcasts/dictionaries/source/string-index-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/values-check-wrong-array.a",
	"stage3/interface-downcasts/dictionaries/source/values-check-wrong-number.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-array-empty.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-array-good.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-array-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-number-empty.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-number-good.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-number-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-object-empty.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-object-good.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-object-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-string-empty.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-string-good.a",
	"stage3/interface-downcasts/dictionaries/source/values-fixed-string-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-array-empty.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-array-good.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-array-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-number-empty.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-number-good.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-number-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-object-empty.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-object-good.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-object-wrong.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-string-empty.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-string-good.a",
	"stage3/interface-downcasts/dictionaries/source/values-producer-string-wrong.a",
	"stage3/interface-downcasts/dictionaries/recheck/values.a",
	"stage3/interface-downcasts/dictionaries/recheck/entries.a",
}

func init() {
	for _, path := range dictionaryIntegrationFixtures {
		name := filepath.Base(path)
		checked := (strings.HasSuffix(name, "-wrong.a") && !strings.Contains(name, "-unread-")) || strings.HasPrefix(name, "values-check-wrong-")
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, checked})
	}
}

// Not parallel: the explicit update writes measured rows in the shared table.
func TestCheckedViewDictionaryIntegrationCounts(t *testing.T) {
	var rows []string
	for _, path := range dictionaryIntegrationFixtures {
		rows = append(rows, counted(t, path, false, nil, false, false))
	}
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if *updateCounts {
		lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
		for _, row := range rows {
			key := strings.Split(row, " | ")[0] + " | "
			replaced := false
			for i, line := range lines {
				if strings.HasPrefix(line, key) {
					lines[i] = row
					replaced = true
					break
				}
			}
			if !replaced {
				insertion := len(lines)
				for i, line := range lines {
					if line == "## Predicate direction counts" {
						insertion = i - 1
						break
					}
				}
				lines = append(lines[:insertion], append([]string{row}, lines[insertion:]...)...)
			}
		}
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	for _, row := range rows {
		if !strings.Contains(string(data), row+"\n") {
			t.Errorf("unrecorded dictionary counts: %s", row)
		}
	}
}
