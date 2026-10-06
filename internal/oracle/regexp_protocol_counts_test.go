package oracle

import (
	"flag"
	"os"
	"strings"
	"testing"
)

var updateProtocolCounts = flag.Bool("update-regexp-protocol-counts", false, "update only this unit's regexp fixture rows")

// Updating a filtered TestCountsAreRecorded run would drop unrelated rows.
// Keep the shared table intact while refreshing just these changed fixtures.
func TestRegExpProtocolCounts(t *testing.T) {
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, path := range []string{
		"internal/oracle/testdata/regexp_matchall_nonglobal.a",
		"internal/oracle/testdata/regexp_replaceall_nonglobal.a",
		"internal/oracle/testdata/regexp_protocol.a",
		"internal/oracle/testdata/regexp_indices.a",
		"internal/oracle/testdata/regexp_replacement_callback.a",
		"internal/oracle/testdata/regexp_string.a",
		"internal/oracle/testdata/regexp_errors.a",
	} {
		row := counted(t, path, false, nil, false, false)
		t.Log(row)
		if strings.Contains(text, row+"\n") {
			continue
		}
		if !*updateProtocolCounts {
			t.Errorf("fixture counts changed: %s", row)
			continue
		}
		prefix := "| " + path + " |"
		at := strings.Index(text, prefix)
		if at < 0 {
			// New normal fixtures precede the input fixtures in the shared table.
			at := strings.Index(text, "| internal/oracle/testdata/regexp_errors.a |")
			if at < 0 {
				at = strings.Index(text, "| internal/oracle/testdata/read_files.a |")
			}
			if at < 0 {
				t.Fatal("input fixture boundary missing from counts table")
			}
			text = text[:at] + row + "\n" + text[at:]
		} else {
			end := at + strings.Index(text[at:], "\n")
			text = text[:at] + row + text[end:]
		}
	}
	if *updateProtocolCounts && !t.Failed() {
		if err := os.WriteFile(countsPath, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
