package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A field read through a union whose object members store the field in different representations is NotYet
// (unionFieldStorageDiffers, #svtsk2x). On main ac79b0b5 Workers' program read YA's a, an unboxed number, as
// the union's boxed string | number, and native crashed in adamic_retain (a sanitizer SEGV) where Node prints
// 3number. Two members sharing a literal tag (XA and YA, both Kind.A) let the switch leave that union
// standing, but the hazard is the storage, so the twins stop too: without the closure, narrowed by an if,
// and a plain XA | YA. The control, whose shared-tag members both store a as a string, lowers and agrees
// with Node on release and sanitized native and on JavaScript. element_key's row is whatever the
// literal-key element read does, measured: it must stop or agree, never crash.
// internal/lower/union_field_storage.mutants.md has the mutant.
func TestUnionFieldStorage(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/union_field_storage"
	const differs = "a field read through a union whose members store .a in different representations"
	stops := map[string]string{
		"workers":       differs,
		"no_capture":    differs,
		"if_narrow":     differs,
		"plain_union":   differs,
		"element_key":   "measure",
		"agree_control": "",
	}
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(stops) {
		t.Fatalf("found %d programs, want %d", len(entries), len(stops))
	}
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		stop, isKnown := stops[name]
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
			if stop != "" && stop != "measure" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.HasPrefix(notYet.What, stop) {
					t.Fatalf("want NotYet starting %q, got %v", stop, err)
				}
				return
			}
			if stop == "measure" && err != nil {
				t.Logf("element_key stops: %v", err)
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
