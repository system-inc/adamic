package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Inside a generic body a type parameter stays rigid (docs/rulings.md, the generic body, and the
// invariant-mutable entry, #p0fy047): a mutable slot holding T seen as one holding T's constraint, or
// anything wider, is a widening and refused, and a readonly one stays covariant. c01 to c03 are
// Correctness's: an append of an X into an X[] lowering, a T tsc infers wider than the array's element
// refused, and the Options hiding case with T nested in an array refused. s01 is scanner.ts:2934's
// decimalEscapes = append(decimalEscapes, { pos, end, value }) over core.ts's append overloads. b01 is
// put, whose `const wide: {}[] = to` wrote a 1 into a string[]; b02 and b03 are its Map and field forms,
// and b04 its readonly form, which lowers. r1 to r8 are Language's roads to the same widening inside
// f<T extends {}>(xs: T[]): an explicit type argument to an inner generic, a Set, a mutable tuple, a
// Record, a method seen bivariantly (r5 a method, r5b a property holding a function, each seen as a
// method taking {}, the one target tsc relates bivariantly), a generic class's field in a method, a
// nested closure, and T extends U seen as U. q01 to q03 are sys.ts:304's pollWatchedFileQueue<T extends
// WatchedFileWithIsClosed> handed a queue of a subtype, which reads T out of the queue and writes it
// back: q01 alone, q02 with sys.ts's host and its callbackOnWatchFileStat?.(...), which lowers since a
// function value called through ?. lowers (#036ve6w), and q03 with that call written as a checked one. Each program's stop is pinned in its
// .stop file as its wall, Refused or NotYet, where it stops, and the rule's tag when the stop names one,
// never another rule's words; or "lowers", and a lowering one is held to Node on release, sanitized
// native and JavaScript. internal/lower/generic-body-rigid.mutants.md has the mutants.
func TestGenericBodyKeepsTypeParameterRigid(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/generic_body_rigid"
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
					t.Fatalf("%s lowered, want %q", entry.Name(), stop)
				}
				if got := pinnedStop(err, strings.TrimSuffix(absolute, path)); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %s", pinnedStop(err, strings.TrimSuffix(absolute, path)))
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("%s source Node: exit %d stderr %q", entry.Name(), truth.exitCode, truth.stderr)
			}
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Fatalf("%s %s: %s: Node %q, %s %q", entry.Name(), backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 40 {
		t.Fatalf("found %d programs, want 40", programs)
	}
}

// ruleTag is the tag a stop's fix ends with, (adamic/invariant-mutable), which names the rule.
var ruleTag = regexp.MustCompile(`\(([a-z]+(?:[/-][a-z]+)*)\)$`)

// pinnedStop is a stop as a pin holds it: its wall, where it stops with root taken off, and the rule's
// tag when it names one.
func pinnedStop(err error, root string) string {
	var refused *lower.Refused
	var notYet *lower.NotYet
	switch {
	case errors.As(err, &refused):
		stop := "Refused " + strings.TrimPrefix(refused.Where, root)
		if tag := ruleTag.FindStringSubmatch(refused.Fix); tag != nil {
			stop += " " + tag[1]
		}
		return stop
	case errors.As(err, &notYet):
		return "NotYet " + strings.TrimPrefix(notYet.Where, root)
	}
	return "Other " + strings.ReplaceAll(err.Error(), root, "")
}
