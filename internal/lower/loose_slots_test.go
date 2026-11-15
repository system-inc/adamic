package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// looseLowered loads and lowers one program, or returns why it couldn't.
func looseLowered(path string) error {
	program, err := load.Load([]string{path})
	if err != nil {
		return err
	}
	_, err = Lower(context.Background(), program)
	return err
}

// looseFixture writes a loose_slot_cycles pin's committed text into a file of its own.
func looseFixture(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("../oracle/testdata/loose_slot_cycles", name+".ts"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name+".ts")
	if err := os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestLooseSlotStopsUnderProgramRegion holds the stop with a Program region on: its discovery lowering
// neither gathers H nor judges (looseSlotStop), so the lowering after it is what stops g13 and l03.
func TestLooseSlotStopsUnderProgramRegion(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"g13_self_unknown", "l03_two_hops"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			program, err := load.Load([]string{looseFixture(t, name)})
			if err != nil {
				t.Fatal(err)
			}
			_, err = LowerWithOptions(context.Background(), program, Options{ProgramRegion: true})
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.What, looseStopWords) {
				t.Fatalf("got %v, want the loose-slot stop with a Program region on", err)
			}
			t.Logf("stop: %v", err)
		})
	}
}

// TestLooseSlotStopsK17AndK18 holds Lower's two roads (#kaz4pk8) to the loose-slot stop, where the value
// enters the loose slot: k17, a captured unknown whose closure is stored in a declared field, and k18, a
// cycle entering an unknown field by an initializer and closing at a declared field. They were the
// declared arm's reason to be, and the entry arm holds both.
func TestLooseSlotStopsK17AndK18(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"k17_captured_unknown", "k18_entry_then_declared"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := looseLowered(looseFixture(t, name))
			if err == nil || !strings.Contains(err.Error(), looseStopWords) {
				t.Fatalf("got %v, want the loose-slot stop", err)
			}
			t.Logf("stop: %v", err)
		})
	}
}

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

// TestLooseSlotStopCensus measures, when ADAMIC_LOOSE_CENSUS=1, which of the oracle's fixtures naming
// unknown or object the loose-slot stop walls. A program reaches the stop only once everything before the
// cycle finder lowered it, so each one listed is newly walled. ENTRY-ONLY lines are walled programs not
// known to wait on Adamic, and should be none; WAITING lines are looseCensusWaiting's. It's a measurement,
// not a pin.
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
