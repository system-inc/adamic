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
// as a heap pointer: native crashed while JavaScript agreed. Those calls stop now; a reference read
// as a union is the same pointer, so a string's spread still runs.

func TestSpreadPackAlikeAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"o03-string-spread-into-union-rest", "o05-number-after-spread", "o08-function-value-string-spread", "o09-function-value-fixed-beside-spread", "o12-function-value-two-union-rests-string-spread", "o13-function-value-number-rests-apart", "o14-function-declarations-union-rest-beside-number-rest"} {
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
		{"o01-number-spread-into-union-rest", "4:6", "a call with a spread of elements held as a number where the callee reads a " + union},
		{"o02-number-fixed-into-union-parameter", "4:6", "a call with a spread passing a number where the callee reads a " + union},
		{"o04-boolean-spread-into-union-rest", "4:6", "a call with a spread of elements held as a boolean where the callee reads a " + union},
		{"o06-method-number-spread", "6:36", "a call with a spread of elements held as a number where the callee reads a " + union},
		{"o07-function-value-number-spread", "4:6", "a call with a spread of elements held as a number where the callee reads a " + union},
		// A function value's layout packs for every closure its type admits, so lead's call collects show's union
		// rest from the pack too, though lead's own signature reads its number spread alike (#wzr4atz).
		{"o10-function-value-layout-union-rest-beside-string-spread", "7:9", "a call with a spread through a function value whose layout reads a " + union + " where the call packs a number"},
		{"o11-function-value-layout-union-rest", "6:9", "a call with a spread through a function value whose layout reads a " + union + " where the call packs a number"},
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
