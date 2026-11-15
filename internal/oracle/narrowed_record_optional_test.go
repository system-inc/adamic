package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A record's value the checker narrowed present (r["k"] = []) and a call then deleted (clear, which
// TypeScript's narrowing doesn't see) is read under ?. as missing on Node, undefined. records.go wrapped
// that read in a Defined, a TypeError thrown before the ?. tested anything, so a catch took a branch Node
// never takes: caught where Node prints undefined, both exiting 0 (#70kdt66, J-xgt44wh at main 89614cc0).
// Each stop_ program is that read under a ?. in one form, a property, a library call, a push, an element,
// a chain, a template, an if and a try, and a number's toFixed (stop_scalar.ts, whose Unwrap would stop
// where Node reads undefined), and it's NotYet at compile time now, a loud stop where the catchable throw
// was. The same read in parentheses, (r["k"])?.length, panics in recordLiteralRead on
// main, a bug of its own, so it isn't one of them. Each agree_ program is a control held to Node on
// release, sanitized native and JavaScript: a key never narrowed, whose read is a maybe the ?. tests; a
// plain read after the delete, where Node throws too and both catch it; and a narrowed key still present.
func TestNarrowedRecordReadUnderOptional(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/narrowed_record_optional"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	stops, controls := 0, 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".ts" {
			continue
		}
		stop := strings.HasPrefix(entry.Name(), "stop_")
		if stop {
			stops++
		} else {
			controls++
		}
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || !strings.HasSuffix(string(truth.stdout), "\n") {
				t.Fatalf("source Node: exit %d stdout %q stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, absolute)
			if stop {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "a record's value narrowed present and read under ?.") {
					t.Fatalf("want the narrowed record read under ?. NotYet, got %v", err)
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
					t.Fatalf("%s: %s: Node %q, %s %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if stops != 9 || controls != 3 {
		t.Fatalf("%d stop and %d control programs in %s, want 9 and 3", stops, controls, directory)
	}
}
