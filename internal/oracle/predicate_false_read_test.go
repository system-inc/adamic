package oracle

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A checked predicate's false side is checked only where a read observes it (#39wtqdj, Adamic's ruling of
// Nov 11, with Language's condition). tsc's some answers false for an empty array, a lie about the
// reference its false side narrows. no_read is semver.ts's Version.toString: nothing reads this.build on
// the false side, so its false check is dropped, counted unobservable, and an empty build prints as Node
// prints it. Each other program reads the reference where the flow graph carries the false answer: a false
// branch, an else, the code after the true side returns, a closure made on the false path, a second
// narrowing on top (Language's q1), a getter read in a closure, a spread, and the same predicate again. Each
// keeps its false check and stops at the call with the located message where Node prints, on sanitized and
// release native and the JavaScript backend. A getter read on the false side itself, narrowed to undefined,
// stops lowering before the call is reached, on main as here, so getter pins that stop and getter_capture
// reaches the call (internal/lower's TestCheckedPredicateFalseReadDirections holds the plain getter read
// checked). .a refuses the lying some at its declaration, so these are .ts alone.
// Mutant (internal/lower/predicate_false_read.mutants.md): the condition-only arm restored stops no_read at
// its call.
func TestPredicateFalseRead(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, nodeOut, where, checkedOut, source, stop string }{
		{"no_read", "1.2.3\n1.2.3+build.5\n", "32:7", "", "", ""},
		{"false_read", "some\nobject\n", "23:7", "some\n", "readonly string[] | undefined", ""},
		{"else_read", "some 1\nnone object\n", "23:6", "some 1\n", "readonly string[] | undefined", ""},
		{"early_exit", "1\nobject\n0\n", "23:6", "1\n", "readonly string[] | undefined", ""},
		{"capture", "some\nobject\n", "23:7", "some\n", "readonly string[] | undefined", ""},
		{"renarrowed", "0\n", "6:6", "", "readonly string[] | undefined", ""},
		// The stop main 16429a30 and this head's first cut both measured (proof/lower-39w-1-report), at
		// 37:17 before getter.ts's comment grew by a line.
		{"getter", "some\nobject\n", "", "", "", "38:17: stage 0 can't lower a read narrowed to undefined of a property a call may write yet"},
		{"getter_capture", "some\nobject\n", "38:7", "some\n", "readonly string[] | undefined", ""},
		{"spread", "some\n0\n", "25:7", "some\n", "readonly string[] | undefined", ""},
		{"again", "some\nnone\n", "26:7", "some\n", "readonly string[] | undefined", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/predicate_false_read", probe.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(probe.nodeOut)}, node); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			program, err := lowered(t, path)
			if probe.stop != "" {
				var pending *lower.NotYet
				if !errors.As(err, &pending) || !strings.HasSuffix(err.Error(), "/"+probe.name+".ts:"+probe.stop) {
					t.Fatalf("want NotYet ending %q, got %v", probe.stop, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			where, status := "", ""
			for _, site := range program.PredicateChecks.Sites {
				if site.Function == "some" && strings.HasSuffix(site.Where, "/"+probe.name+".ts:"+probe.where) {
					where = site.Where
					for _, direction := range site.Directions {
						if direction.Direction == "false" {
							status = direction.Status
						}
					}
				}
			}
			expected := "checked"
			if probe.source == "" {
				expected = "unobservable"
			}
			if where == "" || status != expected {
				t.Fatalf("some's false direction at %s: %q, want %q, in %+v", probe.where, status, expected, program.PredicateChecks.Sites)
			}
			want := node
			if probe.source != "" {
				message := fmt.Sprintf("adamic: panic: predicate some at %s false branch: source %s, target readonly string[]\n", where, probe.source)
				want = run{stdout: []byte(probe.checkedOut), stderr: []byte(message), exitCode: 70}
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr)
				}
			}
			if want.exitCode == 0 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
