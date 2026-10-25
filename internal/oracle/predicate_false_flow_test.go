package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// The false-direction walk's lowering controls (#vdt9f1q, #1761dah). no_escape holds 43 of
// Language's shapes, each reading n.e on the false side of a call through a predicate value, over a
// proven predicate, so no checked predicate escapes and the switch is off: each lowers as on main.
// controls_live makes the switch live with a checked predicate held as a value and never called,
// and reads nothing narrowed on a false side through a value: Language's controls the checker
// narrows nothing by, a true-direction call through a value, and R-T80's direct false forms.
// Mutants: the switch always live stops no_escape at c01's read; the walk counting every
// condition through a value, true sides too, stops controls_live at trueSide's read.
func init() {
	for _, name := range []string{"no_escape.ts", "controls_live.ts"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/predicate_false_flow/" + name, true, false})
	}
}

// Language's attack on the false direction (#vdt9f1q, 49 shapes, tsc 6.0.3), each a program where
// isS, unproven and lying on false, escapes as a value and reaches the shape: the read the checker
// narrowed on the false side is NotYet, naming the call, before any other diagnostic. escape_*
// reach one shape by one road each, the only escape in their program: a field store, a shorthand
// field, a return, an array element, a function expression and an arrow. control_* are the seven
// shapes the checker narrows nothing by, written to read kind, which every Node has; the walk
// leaves them, so each keeps main's diagnostic. Each program runs on Node, and its whole first
// diagnostic is pinned in its .stopped file. Mutants: the walk believing a false step through a
// value returns all 55 attack and escape programs to main's diagnostic (52 to "an indirect call of
// a checked predicate", c30 to its overload twin, c29 and c47 to a refusal), and none lowers; the
// switch seeing only an alias or a call's argument misses the 11 whose escape is a field, a return,
// an element or a default: the four escape_ programs, c19, c20, c25, c27, c36, c47 and c48.
func TestPredicateFalseDirectionReadsAreNotYet(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/predicate_false_flow_stopped"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".ts") {
			continue
		}
		programs++
		path := directory + "/" + name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || len(truth.stdout) == 0 {
				t.Fatalf("Node: %+v", truth)
			}
			_, stopped := lowered(t, absolute)
			var notYet *lower.NotYet
			var refused *lower.Refused
			if !errors.As(stopped, &notYet) && !errors.As(stopped, &refused) {
				t.Fatalf("want a NotYet or a refusal, got %v", stopped)
			}
			expected, err := os.ReadFile(strings.TrimSuffix(absolute, ".ts") + ".stopped")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.ReplaceAll(stopped.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
			walked := strings.Contains(got, "narrowed on the false side of")
			if walked == strings.HasPrefix(name, "control_") {
				t.Fatalf("a %s program, but the walk's NotYet is %v in %q", strings.SplitN(name, "_", 2)[0], walked, got)
			}
		})
	}
	if programs != 62 {
		t.Fatalf("%d programs in %s, want 62: a program added or removed needs this count changed with it", programs, directory)
	}
}

// A direct call is R-T80's, never the walk's, even where the walk is live: Language's direct false
// forms (`!`, a const alias, switch (true), === false, !== true, false ===, a comma), each in a
// program where isS escapes as a value, stop at the call's own false check on both backends, as
// on main. Mutant: the walk taking a direct call stops each at the read instead.
func TestPredicateFalseDirectionLeavesDirectCalls(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"d01_not", "d10_alias_cond", "d11_switch_true", "d12_eq_false", "d14_ne_true", "d15_false_eq", "d17_comma"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/predicate_false_flow/direct", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if difference := disagreement(run{stdout: []byte("function undefined\n")}, truth); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if len(program.PredicateChecks.Sites) != 1 {
				t.Fatalf("predicate check sites %+v", program.PredicateChecks.Sites)
			}
			site := program.PredicateChecks.Sites[0]
			want := run{stderr: []byte(fmt.Sprintf("adamic: panic: predicate isS at %s false branch: source Node, target Statement\n", site.Where)), exitCode: 70}
			sanitized, _ := nativelyUncached(t, program)
			for _, got := range []run{sanitized, onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatal(difference)
				}
			}
		})
	}
}
