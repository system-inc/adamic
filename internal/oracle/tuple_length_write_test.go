package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A fixed tuple's length written with the length it has, xs.length = 2 on a [number, number], stores
// nothing (tupleLengthWrite, #87dwc5c): a tuple is an object of fields "0", "1" with no length field, and
// on main ae125455 the write was a SetProperty of "length" that panicked natively ("a field the checker
// proved is there is missing", exit 70) where Node and the JavaScript backend print 2. same,
// parenthesized and the conformance program lower and agree with Node on release and sanitized native and
// on JavaScript. The rest stop with tupleLengthWrite's words, each condition failing closed: a value
// that's the length's literal only through a name, a tuple with an optional element, whose length is
// 1 | 2, and a tuple held in an object's field. internal/lower/tuple_length_write.mutants.md has the
// mutant.
func TestTupleLengthWrite(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/tuple_length_write"
	stops := map[string]string{
		"same":             "",
		"parenthesized":    "",
		"named_value":      "writing a tuple's length with a value that isn't its own length, written as a literal",
		"optional_element": "writing the length of a tuple with optional or rest elements",
		"through_field":    "writing the length of a tuple that isn't a plain local",
	}
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{
		"internal/conformance/testdata/refusals/tuple_length_control_fix.a",
		"internal/conformance/testdata/refusals/tuple_length_control_fix.ts",
	}
	for _, entry := range entries {
		paths = append(paths, filepath.Join(directory, entry.Name()))
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if len(paths) != 12 {
		t.Fatalf("found %d programs, want 12", len(paths))
	}
	for _, path := range paths {
		name := filepath.Base(path)
		// The conformance program, #87dwc5c's own, lowers.
		stop, isKnown := "", true
		if filepath.Dir(path) == directory {
			stop, isKnown = stops[strings.TrimSuffix(name, filepath.Ext(name))]
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if !isKnown {
				t.Fatalf("%s has no expected outcome", path)
			}
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, absolute)
			if stop != "" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || notYet.What != stop {
					t.Fatalf("want NotYet %q, got %v", stop, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			if string(truth.stdout) != "2\n" {
				t.Fatalf("source Node printed %q, want 2", truth.stdout)
			}
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Errorf("%s: %s: Node %q, got %q, stderr %q", backend.name, difference, truth.stdout, backend.got.stdout, backend.got.stderr)
				}
			}
		})
	}
}
