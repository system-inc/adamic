package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// looseSlotWords end lower's loose-slot stop (loose_slots.go's looseStopWords), a NotYet naming the link
// that lowers it.
const looseSlotWords = "a cycle the cycle finder can't follow through unknown or object (#yeep0sx links these slots)"

// looseSlotStops are #kaz4pk8's cycles through a slot typed unknown or object: Language's g roads (#t5q4prw)
// and k roads, Memory's g11 to g13, Language's attack roads (l01 to l08: an upcast at the write and one
// entering H, two loose hops, entry by a library call, a push into an array whose element holds object, a
// generic T), three loose hops (l09) and Lower's k17 (a captured unknown) and k18 (a cycle entering a loose
// slot by an initializer and closing at a declared slot). l03 and l09 are the roads back only through a
// type H holds itself, never through a field of one, which is what the mutant narrowing H to its entries'
// fields loses. Each stops where the value enters the loose slot, at the line and column given. A road
// another wall already stops (other) is held to stopping, whichever stop it is.
var looseSlotStops = []struct {
	name, at string
	other    bool
}{
	{name: "g01_field_unknown", at: ":4:1"},
	{name: "g02_field_object", at: ":4:1"},
	{name: "g06_array_of_unknown", at: ":4:1"},
	{name: "g11_unknown_plain_cycle", at: ":4:1"},
	{name: "g12_object_plain_cycle", at: ":4:1"},
	{name: "g13_self_unknown", at: ":3:1"},
	{name: "k01_generic_t", at: ":3:1"},
	{name: "k03_union_string_object", at: ":3:1"},
	{name: "k04_record_unknown", at: ":3:1"},
	{name: "k06_set_object", at: ":3:1"},
	{name: "k10_index_signature", at: ":3:1"},
	{name: "k11_object_array", at: ":3:1"},
	{name: "k14_default_parameter", at: ":3:5"},
	{name: "k17_captured_unknown", at: ":3:11"},
	{name: "k18_entry_then_declared", at: ":6:"},
	{name: "l01_upcast_value", at: ":5:1"},
	{name: "l02_upcast_entry", at: ":7:1"},
	{name: "l03_two_hops", at: ":5:1"},
	{name: "l04_set_unknown_add", other: true},
	{name: "l05_push_unknown", at: ":4:1"},
	{name: "l06_push_call", other: true},
	{name: "l07_object_assign", other: true},
	{name: "l08_generic_t", at: ":3:5"},
	{name: "l09_three_hops", at: ":7:1"},
}

// looseSlotControls must lower and let go of everything: the controls of #t5q4prw and #kaz4pk8 (g07,
// g09c, g09d, g10, k02, k07, k08, k09, k16), a leaf object written into unknown (c01), a Node handed to
// an unknown parameter that stores it nowhere (c02), a generator whose TNext is unknown collected by new
// Map (c03: only its yield type is stored, as generators/new_map_plain.a's) and Language's u00, which is
// internal/lower's TestAsyncRepeatedBindingsLower row 004 (c04: a catch binding captured by a closure
// pushed into readers, in a loop around an await, every throw a new Error). c04 is the catch narrowing's
// pin (loose_slots.go's looseCaughtRoad): what can be caught can't reach readers, so it lowers and lets go
// of everything, as on main (11 of 11 at J-n1gttw4), and the narrowing's mutant m4 sends it back to the
// stop. region is whether it forms a graph region, as g09c, g10, k02, k07 and k08 did before the stop.
var looseSlotControls = []struct {
	name, want string
	region     bool
}{
	{name: "g07_capture_object", want: "1\n"},
	{name: "g09c_class_field_control", want: "a\n", region: true},
	{name: "g09d_class_no_cycle", want: "a\n"},
	{name: "g10_control_no_record", want: "a\n", region: true},
	{name: "k02_capital_object", want: "a\n", region: true},
	{name: "k07_empty_type", want: "a\n", region: true},
	{name: "k08_nonnullable_unknown", want: "a\n", region: true},
	{name: "k09_capture_unknown", want: "true\n"},
	{name: "k16_honest_unknown_scalar", want: "a\n"},
	{name: "c01_leaf_object", want: "a\n"},
	{name: "c02_unknown_parameter", want: "object a\n"},
	{name: "c03_generator_next_unknown", want: "2\nk0,k1\n"},
	{name: "c04_catch_capture_errors", want: "item0\nitem1\nitem2\n"},
}

// looseSlotSource writes a pin's committed text into a file of its own.
func looseSlotSource(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/loose_slot_cycles", name+".ts"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name+".ts")
	if err := os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestLooseSlotCyclesStop holds each road to its stop, a NotYet. A road that lowers instead is run on the
// counted build, so the failure says what the cycle leaked (g13's 1 allocation and 0 frees without the stop).
func TestLooseSlotCyclesStop(t *testing.T) {
	t.Parallel()
	for _, pin := range looseSlotStops {
		t.Run(pin.name, func(t *testing.T) {
			t.Parallel()
			path := looseSlotSource(t, pin.name)
			program, err := lowered(t, path)
			if err == nil {
				binary := filepath.Join(t.TempDir(), "counted")
				if buildErr := native.Build(native.C(program), binary, native.Options{Count: true}); buildErr != nil {
					t.Fatalf("lowered, want a stop; the counted build failed: %v", buildErr)
				}
				got := execute(t, binary)
				counts := "no counts line"
				if match := leakcheck.CountsLine.FindSubmatch(got.stderr); match != nil {
					counts = "allocations " + string(match[1]) + " frees " + string(match[2])
				}
				t.Fatalf("lowered, want a stop; counted: %s, stdout %q", counts, got.stdout)
			}
			t.Logf("stop: %v", err)
			if pin.other {
				return
			}
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.What, looseSlotWords) || !strings.Contains(notYet.Where, pin.name+".ts"+pin.at) {
				t.Fatalf("got %v, want the loose-slot stop, a NotYet, at %s.ts%s", err, pin.name, pin.at)
			}
		})
	}
}

// TestLooseSlotControlsBalance holds each control to Node and to its want, natively, sanitized with the
// leak check, on the counted build with allocations equal to frees, and on JavaScript (capturedUnionRuns).
func TestLooseSlotControlsBalance(t *testing.T) {
	t.Parallel()
	for _, pin := range looseSlotControls {
		t.Run(pin.name, func(t *testing.T) {
			t.Parallel()
			path := looseSlotSource(t, pin.name)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if pin.region && len(program.GraphTypes) == 0 {
				t.Fatalf("no graph types, so %s no longer forms its region", pin.name)
			}
			capturedUnionRuns(t, path, program, pin.want)
		})
	}
}
