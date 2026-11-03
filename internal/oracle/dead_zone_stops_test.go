package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Two dead-zone roads stage 0 can't keep yet, each wrong on both backends where Node throws or
// assigns (#aw65psm), stop as not yet, in .a and .ts: a class used from a function that may run
// before the class's declaration (c01; Node throws "Cannot access 'Late' before initialization"),
// and a logical assignment to a module binding checked for its dead zone (g01; Node gives 4 where
// both backends left it undefined). Their controls lower and agree with Node: a class used from a
// function when only declarations come first (c02), and ??= on a function's own local (g02).
// Mutants: classDeadZoneGap returning nil lowers c01; the logical assignment's module-binding stop
// dropped lowers g01.
func TestDeadZoneStopsArePinned(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/dead_zone_stops"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					t.Fatalf("lowered, want %q: Node prints %q", stop, onNode(t, absolute).stdout)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			truth := onNode(t, absolute)
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Fatalf("%s: %s: Node %q, %s %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 8 {
		t.Fatalf("found %d programs, want 8", programs)
	}
}
