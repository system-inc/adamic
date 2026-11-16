//go:build loosecensus

// The loose-slot census is a measurement, not a pin, so it builds only under its tag and never skips in the
// lane, whose skip census refuses a skip it can't classify: go test -tags loosecensus -run
// '^TestLooseSlotStopCensus$' -v ./internal/lower/.

package lower

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// looseCensusText is a fixture that names a loose type, the only ones the stop can wall.
var looseCensusText = regexp.MustCompile(`\b(unknown|object)\b`)

// looseCensusWaiting are the oracle programs the stop walls that balance natively only because they break
// their own cycle by hand before the end (arrays.a pops itself, after and outer, probes.a pops a). Their
// cycles are real, so they wait on Adamic: the link (H handed to reachLinks and graphLinks) or a change
// to the fixtures.
var looseCensusWaiting = map[string]bool{
	"to_primitive/arrays.a": true,
	"to_primitive/probes.a": true,
}

// Not parallel: a measurement that lowers the oracle corpus one program at a time, read as one log.
// TestLooseSlotStopCensus measures which of the oracle's fixtures naming
// unknown or object the loose-slot stop walls. A program reaches the stop only once everything before the
// cycle finder lowered it, so each one listed is newly walled. ENTRY-ONLY lines are walled programs not
// known to wait on Adamic, and should be none; WAITING lines are looseCensusWaiting's. It's a measurement,
// not a pin.
func TestLooseSlotStopCensus(t *testing.T) {
	root, err := filepath.Abs("../oracle/testdata")
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{}
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || strings.Contains(path, "loose_slot_cycles") {
			return err
		}
		if extension := filepath.Ext(path); extension != ".a" && extension != ".ts" {
			return nil
		}
		text, err := os.ReadFile(path)
		if err == nil && looseCensusText.Match(text) {
			paths = append(paths, path)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, path := range paths {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		err = looseLowered(path)
		switch {
		case err == nil:
			counts["lowered"]++
		case !strings.Contains(err.Error(), looseStopWords):
			counts["stopped otherwise"]++
		case looseCensusWaiting[filepath.ToSlash(relative)]:
			counts["waiting"]++
			t.Logf("WAITING %s: %v", relative, err)
		default:
			counts["entry"]++
			t.Logf("ENTRY-ONLY %s: %v", relative, err)
		}
	}
	t.Logf("CENSUS fixtures %d lowered %d entry-only %d waiting %d stopped-otherwise %d", len(paths), counts["lowered"], counts["entry"], counts["waiting"], counts["stopped otherwise"])
}
