package lower

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// looseArmStops lowers one program with both arms of the loose-slot stop observed, and returns each
// arm's stop, or nil, as the last judgment heard it. Callers run one program at a time, so what's heard
// is that program's.
func looseArmStops(t *testing.T, path string) (map[string]error, error) {
	t.Helper()
	var mutex sync.Mutex
	stops := map[string]error{}
	observe := func(_, arm string, err error) {
		mutex.Lock()
		defer mutex.Unlock()
		stops[arm] = err
	}
	looseArmsObserved.Store(&observe)
	defer looseArmsObserved.Store(nil)
	program, err := load.Load([]string{path})
	if err != nil {
		return nil, err
	}
	_, err = Lower(context.Background(), program)
	mutex.Lock()
	defer mutex.Unlock()
	heard := map[string]error{}
	for arm, stop := range stops {
		heard[arm] = stop
	}
	return heard, err
}

// TestLooseDeclaredArmAloneHoldsK17AndK18 says which arm holds Lower's two roads (#kaz4pk8): k17, a
// captured unknown whose closure is stored in a declared field, and k18, a cycle entering an unknown field
// by an initializer and closing at a declared field. The entry arm stops each where the value enters the
// loose slot; the declared arm alone stops each at the declared write. Not parallel: the observer is the
// package's.
func TestLooseDeclaredArmAloneHoldsK17AndK18(t *testing.T) {
	for _, name := range []string{"k17_captured_unknown", "k18_entry_then_declared"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("../oracle/testdata/loose_slot_cycles", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), name+".ts")
			if err := os.WriteFile(path, source, 0o644); err != nil {
				t.Fatal(err)
			}
			stops, err := looseArmStops(t, path)
			if err == nil {
				t.Fatal("lowered, want the loose-slot stop")
			}
			for _, arm := range []string{"entry", "declared"} {
				if stops[arm] == nil || !strings.Contains(stops[arm].Error(), looseStopWords) {
					t.Errorf("the %s arm doesn't stop %s: %v", arm, name, stops[arm])
					continue
				}
				t.Logf("%s arm: %v", arm, stops[arm])
			}
		})
	}
}

// looseCensusText is a fixture that names a loose type, the only ones the stop can wall.
var looseCensusText = regexp.MustCompile(`\b(unknown|object)\b`)

// TestLooseSlotStopCensus measures, when ADAMIC_LOOSE_CENSUS=1, which of the oracle's fixtures naming
// unknown or object each arm of the loose-slot stop walls. A program reaches the stop only once everything
// before the cycle finder lowered it, so each one listed is newly walled. DECLARED-ONLY lines are the ones
// the declared arm alone walls, its own cost. It's a measurement, not a pin.
func TestLooseSlotStopCensus(t *testing.T) {
	if os.Getenv("ADAMIC_LOOSE_CENSUS") != "1" {
		t.Skip("set ADAMIC_LOOSE_CENSUS=1 to measure")
	}
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
		relative, _ := filepath.Rel(root, path)
		stops, err := looseArmStops(t, path)
		entry, declared := stops["entry"] != nil, stops["declared"] != nil
		switch {
		case stops == nil:
			counts["not loaded"]++
		case entry && declared:
			counts["both arms"]++
			t.Logf("BOTH %s: %v", relative, stops["entry"])
		case entry:
			counts["entry arm only"]++
			t.Logf("ENTRY-ONLY %s: %v", relative, stops["entry"])
		case declared:
			counts["declared arm only"]++
			t.Logf("DECLARED-ONLY %s: %v", relative, stops["declared"])
		case err == nil:
			counts["lowered"]++
		case len(stops) == 0:
			counts["stopped before the finder"]++
		default:
			counts["stopped after, otherwise"]++
		}
	}
	t.Logf("CENSUS fixtures %d lowered %d entry-only %d declared-only %d both %d stopped-before-finder %d stopped-after-otherwise %d not-loaded %d", len(paths), counts["lowered"], counts["entry arm only"], counts["declared arm only"], counts["both arms"], counts["stopped before the finder"], counts["stopped after, otherwise"], counts["not loaded"])
}
