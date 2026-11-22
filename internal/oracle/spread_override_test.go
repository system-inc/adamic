package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A field written after a spread whose source may lack it is NotYet (spreadMayLack, #sx8pbw1). On main f7fe5ce5
// Language's three programs, an optional boolean and an optional string on a plain object and an optional string on
// an intersection, each overridden after a spread of a source that lacks the field, made native stop with exit 70
// and "compiler bug: a field the checker proved is there is missing" where Node and the JavaScript backend print
// the field: the spread copies its source's shape and writes the field into the copy by name, and the copy has no
// slot for it. They stop at lowering now, in these words. two_spreads.ts is the road through an earlier spread,
// which adamic/single-spread refuses first. present_source.ts (the field required in the source's type) and
// fresh_literal.ts (the source a fresh literal naming the field) lower and agree with Node on release and
// sanitized native and on JavaScript. A stop row that lowers is run natively and its output reported, so the
// mutant (the stop removed) shows the panic. Language's e04, an optional boolean on an intersection, waits on
// #tq6yzqy, which lowers that intersection.
func TestSpreadOverride(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/spread_override"
	type outcome struct {
		refused bool
		what    string
	}
	outcomes := map[string]outcome{
		"f01_plain_spread_override":        {what: "the field valid written after a spread whose source may lack it"},
		"f02_plain_string_override":        {what: "the field label written after a spread whose source may lack it"},
		"f03_intersection_string_override": {what: "the field label written after a spread whose source may lack it"},
		"two_spreads":                      {refused: true, what: "a spread after the first field"},
		"present_source":                   {},
		"fresh_literal":                    {},
	}
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(outcomes) {
		t.Fatalf("found %d programs, want %d", len(entries), len(outcomes))
	}
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		want, isKnown := outcomes[name]
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			if !isKnown {
				t.Fatalf("%s has no expected outcome", entry.Name())
			}
			absolute, err := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, absolute)
			if want.what != "" {
				if err == nil {
					sanitized, _ := natively(t, program)
					t.Fatalf("want a stop %q, but it lowered: Node %q, native exit %d, stdout %q, stderr %q", want.what, truth.stdout, sanitized.exitCode, sanitized.stdout, sanitized.stderr)
				}
				if want.refused {
					var refused *lower.Refused
					if !errors.As(err, &refused) || refused.What != want.what {
						t.Fatalf("want Refused %q, got %v", want.what, err)
					}
					return
				}
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || notYet.What != want.what {
					t.Fatalf("want NotYet %q, got %v", want.what, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
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
