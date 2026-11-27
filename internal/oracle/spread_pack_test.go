package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A call with a spread packs each argument as it's held, and its callee reads each slot as its own
// parameter's type (#p9zm9q5). A number or a boolean packed where a union is read was its bits read
// as a heap pointer: native crashed while JavaScript agreed. Each argument is made what its slots read
// before it's packed now, a spread's elements converted one by one, where those slots read as one
// type; where they read differently the call stops. A reference read as a union is the same pointer,
// so a string's spread runs as it is. Through a function value, the pack fills every target its type
// admits, and each rest tail is copied from it segment by segment, each slot made what the tail holds:
// o10's numbers boxed into show's rest and copied as they are into lead's (#whe5bdk). o18 to o20 run
// the target whose tail converts: numbers with a fixed argument between two spreads, booleans, whose
// boxes are constant, and a string, retained once.

func TestSpreadPackAlikeAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"o01-number-spread-into-union-rest", "o02-number-fixed-into-union-parameter", "o03-string-spread-into-union-rest", "o04-boolean-spread-into-union-rest", "o05-number-after-spread", "o06-method-number-spread", "o07-function-value-number-spread", "o08-function-value-string-spread", "o09-function-value-fixed-beside-spread", "o10-function-value-layout-union-rest-beside-string-spread", "o11-function-value-layout-union-rest", "o12-function-value-two-union-rests-string-spread", "o13-function-value-number-rests-apart", "o14-function-declarations-union-rest-beside-number-rest", "o15-function-value-maybe-number-slot", "o18-function-value-layout-fixed-between-spreads", "o19-function-value-layout-boolean-segment", "o20-function-value-layout-reference-segment"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata", "spread_pack", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node stopped: %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			got, binary := natively(t, program)
			for backend, run := range map[string]run{"sanitized": got, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, run); difference != "" {
					t.Errorf("%s: %s; Node %q, got %q; stderr %q", backend, difference, truth.stdout, run.stdout, run.stderr)
				}
			}
			if leaked := leaks(t, program, binary); leaked != "" {
				t.Fatal(leaked)
			}
		})
	}
}

func TestSpreadPackUnlikeStops(t *testing.T) {
	t.Parallel()
	const union = "union of differently held members"
	for _, row := range []struct{ name, where, what string }{
		// After the empty spread, 7 lands in first or in rest, which read it differently, so there's no one
		// type to make it; it's boxed for rest, which first would read as a number's bits.
		{"o16-number-after-spread-into-two-slot-types", "3:13", "a call with a spread passing a " + union + " where the callee's slots read as a number | undefined and a " + union},
		// lead's own signature reads 7 alike, but its type admits show, whose first the layout reads as a union.
		// After the spread, where 7 lands is known only at run time, and a fixed slot converts nothing.
		{"o17-function-value-number-after-spread-into-layout-fixed-slot", "5:13", "a call with a spread through a function value whose layout reads a " + union + " where the call packs a number"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata", "spread_pack", row.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); truth.exitCode != 0 {
				t.Fatalf("Node stopped: %+v", truth)
			}
			_, err = lowered(t, path)
			var gap *lower.NotYet
			if !errors.As(err, &gap) || gap.What != row.what || !strings.HasSuffix(gap.Where, ":"+row.where) {
				t.Fatalf("got %v; want NotYet at :%s, %q", err, row.where, row.what)
			}
		})
	}
}
