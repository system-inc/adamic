package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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
